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

// setupNotificationTestRouter configures a Postgres test DB and initializes the notification settings controller
func setupNotificationTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, *controller.NotificationSettingsController) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := openTestDB(t)
	config.DB = db

	if err := config.DB.AutoMigrate(
		&models.User{},
		&models.NotificationSettings{},
		&models.Interest{},
		&models.UserInterest{},
	); err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}

	config.SeedInterests()

	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)

	notifRepo := repository.NewNotificationSettingsRepository()
	notifService := service.NewNotificationSettingsService(notifRepo)

	notifController := controller.NewNotificationSettingsController(notifService, userService)

	router := gin.Default()

	router.GET("/internal/users/email/:email/notification-settings", notifController.GetByEmailInternal)

	router.PUT("/users/me/notification-settings", notifController.UpdateForMe)

	return router, db, notifController
}

// createNotificationTestUser creates a user and notification settings saved by EMAIL
func createNotificationTestUser(t *testing.T, db *gorm.DB, email, keycloakID string) *models.User {
	t.Helper()

	user := models.User{
		Email:      email,
		KeycloakID: keycloakID,
		Password:   "hashed-password",
		FirstName:  "Test",
		LastName:   "User",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	settings := models.NotificationSettings{
		UserEmail:     email,
		LikeEmail:     true,
		LikePush:      true,
		FavoriteEmail: false,
		FavoritePush:  false,
		ChatEmail:     true,
		ChatPush:      false,
		SystemEmail:   true,
		SystemPush:    true,
		ExpoPushToken: "ExponentPushToken[TEST123]",
	}
	if err := db.Create(&settings).Error; err != nil {
		t.Fatalf("failed to create test notification settings: %v", err)
	}

	return &user
}

// Tests for GET /internal/users/email/:email/notification-settings

func TestGetNotificationByEmail_MissingEmail_Returns400(t *testing.T) {
	// Gin will 404 if route param is missing (can't call .../email//notification-settings)
	router, _, _ := setupNotificationTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/internal/users/email//notification-settings", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetNotificationByEmail_UserNotFound_Returns404(t *testing.T) {
	router, _, _ := setupNotificationTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/internal/users/email/unknown@example.com/notification-settings", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetNotificationByEmail_Success_Returns200(t *testing.T) {
	router, db, _ := setupNotificationTestRouter(t)

	user := createNotificationTestUser(t, db, "notif@example.com", "keycloak-sub-123")

	url := "/internal/users/email/" + user.Email + "/notification-settings"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, url, nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"expo_push_token":"ExponentPushToken[TEST123]"`)
}

// Tests for PUT /users/me/notification-settings

func TestUpdateNotificationSettings_Unauthorized_Returns401(t *testing.T) {
	router, _, _ := setupNotificationTestRouter(t)

	body := []byte(`{
		"like_email": true,
		"like_push": true,
		"favorite_email": true,
		"favorite_push": true,
		"chat_email": true,
		"chat_push": true,
		"system_email": true,
		"system_push": true,
		"expo_push_token": "ExponentPushToken[NEW]"
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/users/me/notification-settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func createContextWithUserID(
	t *testing.T,
	notifController *controller.NotificationSettingsController,
	db *gorm.DB,
	keycloakID string,
	body []byte,
) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodPut, "/users/me/notification-settings", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	c.Set("user_id", keycloakID)
	c.Set("sub", keycloakID)

	return c, w
}

func TestUpdateNotificationSettings_InvalidJSON_Returns400(t *testing.T) {
	_, db, notifController := setupNotificationTestRouter(t)

	createNotificationTestUser(t, db, "notif2@example.com", "keycloak-sub-456")

	body := []byte(`{invalid`)

	c, w := createContextWithUserID(t, notifController, db, "keycloak-sub-456", body)

	notifController.UpdateForMe(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
