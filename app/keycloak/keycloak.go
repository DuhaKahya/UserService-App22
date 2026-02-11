package keycloak

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"group1-userservice/app/models"
)

// Returned when Keycloak reports a duplicate email (HTTP 409 Conflict)
var ErrEmailAlreadyExists = errors.New("email already exists in keycloak")

// TokenResponse matches the JSON structure returned by Keycloak's token endpoint
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
}

// GetAccessToken exchanges user credentials for an access token using OAuth2 password grant
func GetAccessToken(email, password string) (*TokenResponse, error) {
	// Token endpoint for the configured realm
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("KEYCLOAK_REALM"),
	)

	// Form-encoded OAuth2 password grant payload
	data := fmt.Sprintf(
		"client_id=%s&client_secret=%s&grant_type=password&username=%s&password=%s",
		os.Getenv("KEYCLOAK_CLIENT_ID"),
		os.Getenv("KEYCLOAK_CLIENT_SECRET"),
		email, password,
	)

	// Create HTTP POST request with a body stream (io.Reader) from the form data
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	// Prevent resource leaks by always closing the response body
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to get access token, status: %d", resp.StatusCode)
	}

	// Decode JSON response body into the TokenResponse struct (stream decoding)
	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

// CreateUserInKeycloakWithPassword creates a user via Keycloak Admin API and sets an initial password
func CreateUserInKeycloakWithPassword(user models.User, plainPassword string) (string, error) {
	// Step 1: Get an admin access token from the master realm (needed for admin endpoints)
	adminTokenURL := fmt.Sprintf(
		"%s/realms/master/protocol/openid-connect/token",
		os.Getenv("KEYCLOAK_URL"),
	)

	// Admin login using password grant against the built-in "admin-cli" client
	data := fmt.Sprintf(
		"client_id=admin-cli&grant_type=password&username=%s&password=%s",
		os.Getenv("KEYCLOAK_ADMIN_USER"),
		os.Getenv("KEYCLOAK_ADMIN_PASS"),
	)

	req, _ := http.NewRequest("POST", adminTokenURL, bytes.NewBuffer([]byte(data)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get admin token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to get admin token, status: %d", resp.StatusCode)
	}

	// Parse admin access token from JSON response
	var adminToken TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&adminToken); err != nil {
		return "", fmt.Errorf("failed to decode admin token response: %w", err)
	}

	// Step 2: Create the user in the configured realm using the Admin API
	createURL := fmt.Sprintf(
		"%s/admin/realms/%s/users",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("KEYCLOAK_REALM"),
	)

	// Payload for the "create user" call (Keycloak expects JSON)
	newUser := map[string]interface{}{
		"email":     user.Email,
		"username":  user.Email,
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"enabled":   true,
	}

	// Convert Go map to JSON bytes for the request body
	body, _ := json.Marshal(newUser)

	req, _ = http.NewRequest("POST", createURL, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken.AccessToken)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call keycloak create user: %w", err)
	}
	defer resp.Body.Close()

	// 409 Conflict with duplicate email in Keycloak
	if resp.StatusCode == http.StatusConflict {
		return "", ErrEmailAlreadyExists
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to create user in keycloak, status: %d", resp.StatusCode)
	}

	// Step 3: Query the user to retrieve the generated Keycloak user ID (UUID)
	findURL := fmt.Sprintf(
		"%s/admin/realms/%s/users?email=%s",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("KEYCLOAK_REALM"),
		user.Email,
	)

	getReq, _ := http.NewRequest("GET", findURL, nil)
	getReq.Header.Set("Authorization", "Bearer "+adminToken.AccessToken)

	findResp, err := http.DefaultClient.Do(getReq)
	if err != nil {
		return "", fmt.Errorf("failed to query created user: %w", err)
	}
	defer findResp.Body.Close()

	if findResp.StatusCode < 200 || findResp.StatusCode >= 300 {
		return "", fmt.Errorf("failed to query created user, status: %d", findResp.StatusCode)
	}

	// Decode JSON array of users returned by Keycloak search
	var found []map[string]interface{}
	if err := json.NewDecoder(findResp.Body).Decode(&found); err != nil {
		return "", fmt.Errorf("failed to decode user search response: %w", err)
	}

	// Keycloak search returned no user, something went wrong after creation
	if len(found) == 0 {
		return "", fmt.Errorf("user not found after creation")
	}

	// Extract the id field from the first result
	idRaw, ok := found[0]["id"]
	if !ok {
		return "", fmt.Errorf("user id missing in keycloak response")
	}

	// Type assertion: JSON decoding uses interface so we assert it's a string UUID
	kcID, ok := idRaw.(string)
	if !ok {
		return "", fmt.Errorf("invalid keycloak id type")
	}

	// Step 4: Set initial password using the reset-password admin endpoint
	passURL := fmt.Sprintf(
		"%s/admin/realms/%s/users/%s/reset-password",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("KEYCLOAK_REALM"),
		kcID,
	)

	passPayload := map[string]interface{}{
		"type":      "password",
		"value":     plainPassword,
		"temporary": false,
	}

	// Convert password payload to JSON
	passJSON, _ := json.Marshal(passPayload)

	reqPass, _ := http.NewRequest("PUT", passURL, bytes.NewBuffer(passJSON))
	reqPass.Header.Set("Content-Type", "application/json")
	reqPass.Header.Set("Authorization", "Bearer "+adminToken.AccessToken)

	passResp, err := http.DefaultClient.Do(reqPass)
	if err != nil {
		return "", fmt.Errorf("failed to set password in keycloak: %w", err)
	}
	defer passResp.Body.Close()

	// Keycloak returns 204 No Content when password update succeeds
	if passResp.StatusCode != http.StatusNoContent {
		return "", fmt.Errorf("failed to set password, status: %d", passResp.StatusCode)
	}

	return kcID, nil
}

// ResetKeycloakPasswordByEmail finds a user by email and updates their password via Admin API
func ResetKeycloakPasswordByEmail(email, newPassword string) error {
	// Get admin access token (same flow as in user creation)
	adminTokenURL := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", os.Getenv("KEYCLOAK_URL"))

	data := fmt.Sprintf(
		"client_id=admin-cli&grant_type=password&username=%s&password=%s",
		os.Getenv("KEYCLOAK_ADMIN_USER"),
		os.Getenv("KEYCLOAK_ADMIN_PASS"),
	)

	req, _ := http.NewRequest("POST", adminTokenURL, bytes.NewBuffer([]byte(data)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get admin token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to get admin token, status: %d", resp.StatusCode)
	}

	// Decode admin token from JSON response
	var adminToken TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&adminToken); err != nil {
		return fmt.Errorf("failed to decode admin token response: %w", err)
	}

	// Search for the user by email to get their Keycloak user ID
	findURL := fmt.Sprintf(
		"%s/admin/realms/%s/users?email=%s",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("KEYCLOAK_REALM"),
		email,
	)

	getReq, _ := http.NewRequest("GET", findURL, nil)
	getReq.Header.Set("Authorization", "Bearer "+adminToken.AccessToken)

	findResp, err := http.DefaultClient.Do(getReq)
	if err != nil {
		return fmt.Errorf("failed to query user by email: %w", err)
	}
	defer findResp.Body.Close()

	if findResp.StatusCode < 200 || findResp.StatusCode >= 300 {
		return fmt.Errorf("failed to query user, status: %d", findResp.StatusCode)
	}

	// Decode search results (Keycloak returns an array)
	var found []map[string]interface{}
	if err := json.NewDecoder(findResp.Body).Decode(&found); err != nil {
		return fmt.Errorf("failed to decode user search response: %w", err)
	}
	if len(found) == 0 {
		return fmt.Errorf("user not found in keycloak")
	}

	// Extract and type-assert the Keycloak user ID (UUID string)
	idRaw, ok := found[0]["id"]
	if !ok {
		return fmt.Errorf("user id missing in keycloak response")
	}
	kcID, ok := idRaw.(string)
	if !ok {
		return fmt.Errorf("invalid keycloak id type")
	}

	// Build reset-password endpoint and send new password payload
	passURL := fmt.Sprintf("%s/admin/realms/%s/users/%s/reset-password",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("KEYCLOAK_REALM"),
		kcID,
	)

	passPayload := map[string]interface{}{
		"type":      "password",
		"value":     newPassword,
		"temporary": false,
	}
	passJSON, _ := json.Marshal(passPayload)

	reqPass, _ := http.NewRequest("PUT", passURL, bytes.NewBuffer(passJSON))
	reqPass.Header.Set("Content-Type", "application/json")
	reqPass.Header.Set("Authorization", "Bearer "+adminToken.AccessToken)

	passResp, err := http.DefaultClient.Do(reqPass)
	if err != nil {
		return fmt.Errorf("failed to set password in keycloak: %w", err)
	}
	defer passResp.Body.Close()

	// 204 No Content indicates success
	if passResp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to set password, status: %d", passResp.StatusCode)
	}

	return nil
}

// Function variable so tests can replace it with a mock implementation
var ResetPasswordByEmail = ResetKeycloakPasswordByEmail
