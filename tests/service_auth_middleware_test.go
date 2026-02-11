package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"group1-userservice/app/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Test router for middleware
func setupAuthMiddlewareTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.Default()
	r.Use(middleware.AuthMiddleware())

	// This handler only runs if middleware succeeds
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	return r
}

// AuthMiddleware tests

func TestAuthMiddleware_MissingAuthorizationHeader_Returns401(t *testing.T) {
	router := setupAuthMiddlewareTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing authorization header")
}

func TestAuthMiddleware_InvalidAuthorizationFormat_Returns401(t *testing.T) {
	router := setupAuthMiddlewareTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)

	// wrong format → missing "Bearer "
	req.Header.Set("Authorization", "Token abc123")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid authorization format")
}

// GetUserID tests

func TestGetUserID_ReturnsIDAndTrue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Set("user_id", "kc-sub-123")

	id, ok := middleware.GetUserID(c)

	assert.True(t, ok)
	assert.Equal(t, "kc-sub-123", id)
}

func TestGetUserID_NotSet_ReturnsFalse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	id, ok := middleware.GetUserID(c)

	assert.False(t, ok)
	assert.Equal(t, "", id)
}
