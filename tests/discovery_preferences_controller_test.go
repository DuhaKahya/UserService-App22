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

// setupDiscoveryPrefsTestRouter sets up a Postgres test database and initializes the controller
func setupDiscoveryPrefsTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := openTestDB(t)
	config.DB = db

	if err := config.DB.AutoMigrate(
		&models.User{},
		&models.DiscoveryPreferences{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// Real services
	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)

	prefsRepo := repository.NewDiscoveryPreferencesRepository(config.DB)
	prefsService := service.NewDiscoveryPreferencesService(prefsRepo)

	prefsController := controller.NewDiscoveryPreferencesController(prefsService, userService)

	router := gin.Default()

	// Test-only middleware stub: mimics AuthMiddleware by setting "sub"/"user_id" in context
	router.Use(func(c *gin.Context) {
		if sub := c.GetHeader("X-TEST-USER-SUB"); sub != "" {
			// Keep both keys to match whatever middleware.GetUserID(c) reads
			c.Set("user_id", sub)
			c.Set("sub", sub)
		}
		c.Next()
	})

	router.GET("/users/me/discovery-preferences", prefsController.GetForMe)
	router.PUT("/users/me/discovery-preferences", prefsController.UpdateForMe)

	// Internal endpoint (no ServiceAuthMiddleware in tests)
	router.GET("/internal/users/:email/discovery-preferences", prefsController.GetByEmailInternal)

	return router, db
}

// createDiscoveryPrefsTestUser inserts a user used in discovery preferences tests
func createDiscoveryPrefsTestUser(t *testing.T, db *gorm.DB, email, keycloakID string) *models.User {
	t.Helper()

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

// Tests for GET /users/me/discovery-preferences

func TestGetDiscoveryPrefs_Unauthorized_NoSub(t *testing.T) {
	router, _ := setupDiscoveryPrefsTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me/discovery-preferences", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetDiscoveryPrefs_UserNotFound(t *testing.T) {
	router, _ := setupDiscoveryPrefsTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me/discovery-preferences", nil)
	req.Header.Set("X-TEST-USER-SUB", "unknown-sub")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetDiscoveryPrefs_Success_ReturnsDefaults(t *testing.T) {
	router, db := setupDiscoveryPrefsTestRouter(t)

	user := createDiscoveryPrefsTestUser(t, db, "prefs@example.com", "kc-prefs-123")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/me/discovery-preferences", nil)
	req.Header.Set("X-TEST-USER-SUB", user.KeycloakID)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()

	// Response should contain email and default radius_km = 50
	assert.Contains(t, body, `"email":"prefs@example.com"`)
	assert.Contains(t, body, `"radius_km":50`)
}

// Tests for PUT /users/me/discovery-preferences

func TestUpdateDiscoveryPrefs_Unauthorized_NoSub(t *testing.T) {
	router, _ := setupDiscoveryPrefsTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/users/me/discovery-preferences", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateDiscoveryPrefs_UserNotFound(t *testing.T) {
	router, _ := setupDiscoveryPrefsTestRouter(t)

	body := []byte(`{"radius_km":25}`)
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPut, "/users/me/discovery-preferences", bytes.NewBuffer(body))
	req.Header.Set("X-TEST-USER-SUB", "unknown-user")
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateDiscoveryPrefs_InvalidJSON_Returns400(t *testing.T) {
	router, db := setupDiscoveryPrefsTestRouter(t)

	user := createDiscoveryPrefsTestUser(t, db, "brokenjson@example.com", "kc-json")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/users/me/discovery-preferences", bytes.NewBufferString("{bad json"))

	req.Header.Set("X-TEST-USER-SUB", user.KeycloakID)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateDiscoveryPrefs_InvalidRange_Returns400(t *testing.T) {
	router, db := setupDiscoveryPrefsTestRouter(t)

	user := createDiscoveryPrefsTestUser(t, db, "range@example.com", "kc-range")

	body := []byte(`{"radius_km":0}`) // invalid
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPut, "/users/me/discovery-preferences", bytes.NewBuffer(body))
	req.Header.Set("X-TEST-USER-SUB", user.KeycloakID)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateDiscoveryPrefs_Success(t *testing.T) {
	router, db := setupDiscoveryPrefsTestRouter(t)

	user := createDiscoveryPrefsTestUser(t, db, "updateprefs@example.com", "kc-updateprefs")

	body := []byte(`{"radius_km":25}`)
	w := httptest.NewRecorder()

	req, _ := http.NewRequest(http.MethodPut, "/users/me/discovery-preferences", bytes.NewBuffer(body))
	req.Header.Set("X-TEST-USER-SUB", user.KeycloakID)
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := w.Body.String()

	// Response should contain email and updated radius_km
	assert.Contains(t, resp, `"email":"updateprefs@example.com"`)
	assert.Contains(t, resp, `"radius_km":25`)
}

// Tests for GET /internal/users/:email/discovery-preferences

func TestDiscoveryPrefs_InternalGet_UserNotFound_Returns404(t *testing.T) {
	router, _ := setupDiscoveryPrefsTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/internal/users/missing@example.com/discovery-preferences", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDiscoveryPrefs_InternalGet_UserExists_Returns200_Defaults(t *testing.T) {
	router, db := setupDiscoveryPrefsTestRouter(t)

	// Internal endpoint checks user exists (if controller still does UserService.GetByEmail)
	_ = createDiscoveryPrefsTestUser(t, db, "internalprefs@example.com", "kc-internal")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/internal/users/internalprefs@example.com/discovery-preferences", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := w.Body.String()

	// Response should contain email and default radius_km = 50
	assert.Contains(t, resp, `"email":"internalprefs@example.com"`)
	assert.Contains(t, resp, `"radius_km":50`)
}
