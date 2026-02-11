package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"group1-userservice/app/config"
	controller "group1-userservice/app/controllers"
	"group1-userservice/app/models"
	"group1-userservice/app/repository"
	"group1-userservice/app/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupAuthTestRouter configures a Postgres test database and initializes the login controller
func setupAuthTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	// Open Postgres test DB
	db := openTestDB(t)

	// Override global DB
	config.DB = db

	// Migrate required tables
	if err := config.DB.AutoMigrate(
		&models.User{},
		&models.NotificationSettings{},
		&models.Interest{},
		&models.UserInterest{},
	); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	// Optional: seed interests so IDs/keys exist in test DB
	config.SeedInterests()

	// Initialize repository and service
	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)
	loginController := controller.NewLoginController(userService)

	// Create Gin router with authentication routes
	router := gin.Default()
	router.POST("/auth/login", loginController.Handle)
	router.POST("/auth/refresh", loginController.Refresh)

	return router, db
}

// Tests for POST /auth/login

// Invalid JSON should return 400
func TestLogin_InvalidJSON_Returns400(t *testing.T) {
	router, _ := setupAuthTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Login with non-existing email should return 401
func TestLogin_UserNotFound_Returns401(t *testing.T) {
	router, _ := setupAuthTestRouter(t)

	body := []byte(`{"email":"missing@example.com","password":"whatever"}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Login with incorrect password should return 401
func TestLogin_WrongPassword_Returns401(t *testing.T) {
	router, db := setupAuthTestRouter(t)

	const email = "test@example.com"
	const correctPassword = "correct-password"
	createTestUser(t, db, email, correctPassword)

	body := []byte(`{"email":"` + email + `","password":"wrong-password"}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Tests for POST /auth/refresh

// Missing refresh_token should return 400
func TestRefresh_MissingToken_Returns400(t *testing.T) {
	router, _ := setupAuthTestRouter(t)

	body := []byte(`{}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
