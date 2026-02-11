package tests

import (
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
)

// End to end test for GET /users/:email
func TestEndToEnd_GetUserByEmail_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// test database
	db := openTestDB(t)
	config.DB = db

	// migrate
	if err := config.DB.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	// seed one user
	user := models.User{
		Email:      "e2e@example.com",
		FirstName:  "End",
		LastName:   "ToEnd",
		Password:   "hashed-pass",
		KeycloakID: "kc-e2e-1",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create seed user: %v", err)
	}

	// real service stack
	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)
	fakeBadges := &fakeBadgeService{}
	userController := controller.NewUserController(userService, fakeBadges)

	// real HTTP router + route
	router := gin.Default()
	router.GET("/users/:email", userController.GetByEmail)

	// perform HTTP request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/users/e2e@example.com", nil)

	router.ServeHTTP(w, req)

	// assertions on full response
	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, `"email":"e2e@example.com"`)
	assert.NotContains(t, body, `"password":"hashed-pass"`)
}
