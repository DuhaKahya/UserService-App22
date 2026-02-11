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

// setupRegisterTestRouter creates a Postgres test database and initializes the RegisterController
func setupRegisterTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
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
		t.Fatalf("failed to migrate tables: %v", err)
	}

	config.SeedInterests()

	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)
	registerController := controller.NewRegisterController(userService)

	router := gin.Default()
	router.POST("/users/register", registerController.Handle)

	return router, db
}

// setupRegisterSuccessRouter uses the fakeUserService to avoid Keycloak calls
func setupRegisterSuccessRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	fakeSvc := &fakeUserService{}
	registerController := controller.NewRegisterController(fakeSvc)

	router := gin.Default()
	router.POST("/users/register", registerController.Handle)

	return router
}

// Tests for POST /users/register

func TestRegister_InvalidJSON_Returns400(t *testing.T) {
	router, _ := setupRegisterTestRouter(t)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegister_EmailAlreadyExists_Returns409(t *testing.T) {
	router, db := setupRegisterTestRouter(t)

	existing := models.User{
		Email:      "test@example.com",
		Password:   hashPassword(t, "secret"),
		FirstName:  "Existing",
		LastName:   "User",
		KeycloakID: "kc-register-controller-exists",
	}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatalf("failed to create existing user: %v", err)
	}

	body := []byte(`{
		"email": "test@example.com",
		"password": "secret123",
		"first_name": "John",
		"last_name": "Doe"
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	bodyStr := w.Body.String()
	assert.Contains(t, bodyStr, "Email already in use")
}

func TestRegister_Success_Returns201(t *testing.T) {
	router := setupRegisterSuccessRouter(t)

	body := []byte(`{
		"email": "newuser@example.com",
		"password": "Mypassword123",
		"first_name": "Alice",
		"last_name": "Smith"
	}`)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var returnedUser models.User
	err := json.Unmarshal(w.Body.Bytes(), &returnedUser)
	assert.NoError(t, err)

	assert.Equal(t, "", returnedUser.Password)
	assert.Equal(t, "newuser@example.com", returnedUser.Email)
	assert.Equal(t, "Alice", returnedUser.FirstName)
	assert.Equal(t, "Smith", returnedUser.LastName)
}
