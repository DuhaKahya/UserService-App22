package tests

import (
	"bytes"
	"encoding/json"
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

// setupUserTestRouter configures a Postgres test DB and initializes the UserController
func setupUserTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, *controller.UserController) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := openTestDB(t)
	config.DB = db

	if err := config.DB.AutoMigrate(
		&models.User{},
		&models.NotificationSettings{},
		&models.Interest{},
		&models.UserInterest{},
		&models.UserBadge{},
		&models.Badge{},
	); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}

	config.SeedInterests()
	config.SeedBadges()

	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)

	userBadgeRepo := repository.NewUserBadgeRepository(config.DB)
	userBadgeService := service.NewUserBadgeService(userBadgeRepo)

	userController := controller.NewUserController(userService, userBadgeService)

	router := gin.Default()

	// public info route
	router.GET("/users/:firstname/:lastname", userController.GetByFirstLast)

	// keycloak route
	router.GET("/users/keycloak/:sub", userController.GetByKeycloakSub)

	// internal email route
	router.GET("/internal/users/:email", userController.GetByEmail)

	router.PUT("/users/me", func(c *gin.Context) {
		c.Set("user_id", "kc-me-sub-1")
		userController.UpdateMe(c)
	})

	return router, db, userController
}

// createUserTestUser inserts a user into the test database
func createUserTestUser(t *testing.T, db *gorm.DB, email, keycloakID, password string) *models.User {
	t.Helper()

	user := models.User{
		Email:      email,
		KeycloakID: keycloakID,
		Password:   password,
		FirstName:  "Test",
		LastName:   "User",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return &user
}

// Tests for GET /internal/users/:email

func TestGetUserByEmail_EmptyEmail_Returns404(t *testing.T) {
	router, _, _ := setupUserTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/internal/users/", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserByEmail_NotFound_Returns404(t *testing.T) {
	router, _, _ := setupUserTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/internal/users/unknown@example.com", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserByEmail_Success_Returns200(t *testing.T) {
	router, db, _ := setupUserTestRouter(t)

	createUserTestUser(t, db, "emailtest@example.com", "kc-sub-email-1", "hashed-pass")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/internal/users/emailtest@example.com", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"email":"emailtest@example.com"`)
	assert.NotContains(t, body, `"password":"hashed-pass"`)
}

// Tests for GET /users/keycloak/:sub

func TestGetUserByKeycloakSub_EmptySub_Returns400(t *testing.T) {
	_, _, userController := setupUserTestRouter(t)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// No sub path param set -> GetByKeycloakSub will see empty string
	c.Params = gin.Params{}

	userController.GetByKeycloakSub(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUserByKeycloakSub_NotFound_Returns404(t *testing.T) {
	router, _, _ := setupUserTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/keycloak/unknown-sub", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserByKeycloakSub_Success_Returns200(t *testing.T) {
	router, db, _ := setupUserTestRouter(t)

	createUserTestUser(t, db, "kc@example.com", "kc-controller-sub-123", "hashed-pass")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/keycloak/kc-controller-sub-123", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"email":"kc@example.com"`)
	assert.NotContains(t, body, `"password":"hashed-pass"`)
}

// Tests for GET /users/:firstname/:lastname (public info)

func TestGetUserByFirstLast_NotFound_Returns404(t *testing.T) {
	router, _, _ := setupUserTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/Nope/Unknown", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetUserByFirstLast_Success_ReturnsOnlyEmailAndNames(t *testing.T) {
	router, db, _ := setupUserTestRouter(t)

	createUserTestUser(t, db, "john.doe@example.com", "kc-firstlast-1", "hashed-pass")
	db.Model(&models.User{}).Where("email = ?", "john.doe@example.com").
		Updates(map[string]any{"first_name": "John", "last_name": "Doe"})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/john/doe", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, `"email":"john.doe@example.com"`)
	assert.Contains(t, body, `"first_name":"John"`)
	assert.Contains(t, body, `"last_name":"Doe"`)

	assert.NotContains(t, body, `"password"`)
	assert.NotContains(t, body, `"keycloak_id"`)
}

func TestUpdateMe_InvalidBody_Returns400(t *testing.T) {
	router, _, _ := setupUserTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewBufferString("{invalid-json"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateMe_Success_Returns200AndUpdatesUser(t *testing.T) {
	router, db, _ := setupUserTestRouter(t)

	createUserTestUser(t, db, "me@example.com", "kc-me-sub-1", "hashed-pass")

	payload := map[string]any{
		"first_name": "NewName",
		"biography":  "New bio",
	}
	b, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, `"email":"me@example.com"`)
	assert.Contains(t, body, `"first_name":"NewName"`)
	assert.Contains(t, body, `"biography":"New bio"`)
	assert.NotContains(t, body, `"password":"hashed-pass"`)

	var fromDB models.User
	err := db.First(&fromDB, "email = ?", "me@example.com").Error
	assert.NoError(t, err)
	assert.Equal(t, "NewName", fromDB.FirstName)
	assert.Equal(t, "New bio", fromDB.Biography)
}

func TestUpdateMe_CanSetPhoneNumberVisibleTrue(t *testing.T) {
	router, db, _ := setupUserTestRouter(t)

	createUserTestUser(t, db, "me@example.com", "kc-me-sub-1", "hashed-pass")

	payload := map[string]any{
		"phone_number_visible": true,
	}
	b, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Response check
	body := w.Body.String()
	assert.Contains(t, body, `"phone_number_visible":true`)

	// DB check
	var fromDB models.User
	err := db.First(&fromDB, "email = ?", "me@example.com").Error
	assert.NoError(t, err)
	assert.True(t, fromDB.PhoneNumberVisible)
}

func TestUpdateMe_CanTogglePhoneNumberVisibleTrueThenFalse(t *testing.T) {
	router, db, _ := setupUserTestRouter(t)
	createUserTestUser(t, db, "me@example.com", "kc-me-sub-1", "hashed-pass")

	// set true
	{
		payload := map[string]any{"phone_number_visible": true}
		b, _ := json.Marshal(payload)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// set false
	{
		payload := map[string]any{"phone_number_visible": false}
		b, _ := json.Marshal(payload)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	var fromDB models.User
	err := db.First(&fromDB, "email = ?", "me@example.com").Error
	assert.NoError(t, err)
	assert.False(t, fromDB.PhoneNumberVisible)
}

func TestUpdateMe_DoesNotChangePasswordOrKeycloakID(t *testing.T) {
	router, db, _ := setupUserTestRouter(t)

	createUserTestUser(t, db, "me@example.com", "kc-me-sub-1", "keep-hash")

	payload := map[string]any{
		"password":    "hacked",
		"keycloak_id": "kc-hacked",
		"first_name":  "Ok",
	}
	b, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var fromDB models.User
	err := db.First(&fromDB, "email = ?", "me@example.com").Error
	assert.NoError(t, err)
	assert.Equal(t, "keep-hash", fromDB.Password)
	assert.Equal(t, "kc-me-sub-1", fromDB.KeycloakID)
	assert.Equal(t, "Ok", fromDB.FirstName)
}
