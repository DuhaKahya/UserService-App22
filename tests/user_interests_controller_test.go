package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"group1-userservice/app/config"
	controller "group1-userservice/app/controllers"
	"group1-userservice/app/models"
	"group1-userservice/app/repository"
	"group1-userservice/app/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupInterestsTestRouter sets up a Postgres test database and initializes the controller
func setupInterestsTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := openTestDB(t)
	config.DB = db

	if err := config.DB.AutoMigrate(
		&models.User{},
		&models.Interest{},
		&models.UserInterest{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	config.SeedInterests()

	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)

	interestsRepo := repository.NewUserInterestsRepository()
	interestsService := service.NewUserInterestsService(interestsRepo)

	interestsController := controller.NewUserInterestsController(interestsService, userService)

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		if sub := c.GetHeader("X-TEST-USER-SUB"); sub != "" {
			// Must match middleware.GetUserID key ("user_id" / "sub" depending on your middleware)
			c.Set("user_id", sub)
			c.Set("sub", sub)
		}
		c.Next()
	})

	router.GET("/users/me/interests", interestsController.GetForMe)
	router.PUT("/users/me/interests", interestsController.UpdateForMe)

	return router, db
}

// createInterestsTestUser inserts a user used in interests tests
func createInterestsTestUser(t *testing.T, db *gorm.DB, email, keycloakID string) *models.User {
	t.Helper()

	// Make values unique to avoid unique constraint collisions across tests
	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	if email == "" {
		email = "test_" + suffix + "@example.com"
	} else {
		email = email + "_" + suffix
	}
	if keycloakID == "" {
		keycloakID = "kc-" + suffix
	} else {
		keycloakID = keycloakID + "-" + suffix
	}

	user := models.User{
		Email:      email,
		KeycloakID: keycloakID,
		FirstName:  "Test",
		LastName:   "User",
		Password:   "hash",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return &user
}

// Tests for GET /users/me/interests

func TestGetInterests_Unauthorized_NoSub(t *testing.T) {
	router, _ := setupInterestsTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me/interests", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetInterests_UserNotFound(t *testing.T) {
	router, _ := setupInterestsTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me/interests", nil)
	req.Header.Set("X-TEST-USER-SUB", "unknown-sub")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetInterests_Success_ReturnsDefaults(t *testing.T) {
	router, db := setupInterestsTestRouter(t)

	user := createInterestsTestUser(t, db, "interests@example.com", "kc-123")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me/interests", nil)
	req.Header.Set("X-TEST-USER-SUB", user.KeycloakID)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	// New response contains email + list of interests
	assert.Contains(t, body, `"email":"`+user.Email+`"`)

	// Defaults: value false for all (since user has no rows yet)
	assert.Contains(t, body, `"key":"ICT"`)
	assert.Contains(t, body, `"value":false`)
	assert.Contains(t, body, `"key":"Media en Communicatie"`)

}

// Tests for PUT /users/me/interests

func TestUpdateInterests_Unauthorized_NoSub(t *testing.T) {
	router, _ := setupInterestsTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/users/me/interests", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateInterests_UserNotFound(t *testing.T) {
	router, _ := setupInterestsTestRouter(t)

	body := []byte(`{"interests":[{"id":1,"value":true}]}`)
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPut, "/users/me/interests", bytes.NewBuffer(body))
	req.Header.Set("X-TEST-USER-SUB", "unknown-user")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateInterests_InvalidJSON_Returns400(t *testing.T) {
	router, db := setupInterestsTestRouter(t)

	user := createInterestsTestUser(t, db, "brokenjson@example.com", "kc-json")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/users/me/interests", bytes.NewBufferString("{bad json"))

	req.Header.Set("X-TEST-USER-SUB", user.KeycloakID)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateInterests_Success(t *testing.T) {
	router, db := setupInterestsTestRouter(t)

	user := createInterestsTestUser(t, db, "update@example.com", "kc-update")

	ictID := getInterestIDByKey(t, db, "ICT")
	mediaID := getInterestIDByKey(t, db, "Media en Communicatie")

	body := map[string]interface{}{
		"interests": []map[string]interface{}{
			{"id": ictID, "value": false},
			{"id": mediaID, "value": true},
		},
	}

	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/users/me/interests", bytes.NewBuffer(jsonBody))

	req.Header.Set("X-TEST-USER-SUB", user.KeycloakID)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := w.Body.String()

	// Response should include email and interests array
	assert.Contains(t, resp, `"email":"`+user.Email+`"`)

	// Check returned items contain keys and correct values
	assert.Contains(t, resp, `"key":"ICT"`)
	assert.Contains(t, resp, `"key":"Media en Communicatie"`)

	assert.Contains(t, resp, `"value":false`)
	assert.Contains(t, resp, `"value":true`)
}
