package keycloak

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// RefreshAccessToken exchanges a refresh token for a new access token using OAuth2
func RefreshAccessToken(refreshToken string) (*TokenResponse, error) {
	// Token endpoint for the configured realm
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token",
		os.Getenv("KEYCLOAK_URL"),
		os.Getenv("KEYCLOAK_REALM"),
	)

	// OAuth2 refresh_token grant payload (form-encoded)
	data := fmt.Sprintf(
		"client_id=%s&client_secret=%s&grant_type=refresh_token&refresh_token=%s",
		os.Getenv("KEYCLOAK_CLIENT_ID"),
		os.Getenv("KEYCLOAK_CLIENT_SECRET"),
		refreshToken,
	)

	// Create HTTP POST request with the refresh token as body
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Keycloak returns 200 OK on successful token refresh
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to refresh token, status: %d", resp.StatusCode)
	}

	// Decode the new access and refresh tokens from the JSON response
	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}

	// Return token as a pointer to avoid copying the struct
	return &token, nil
}
