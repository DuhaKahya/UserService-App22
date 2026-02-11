package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	controller "group1-userservice/app/controllers"
	"group1-userservice/app/interfaces"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// fakePasswordResetService avoids real DB / Keycloak calls
type fakePasswordResetService struct {
	requestResetFn func(email string) (string, error)
	resetFn        func(token string, newPassword string) error
}

func (f *fakePasswordResetService) RequestReset(email string) (string, error) {
	if f.requestResetFn != nil {
		return f.requestResetFn(email)
	}
	return "dummy-token", nil
}

func (f *fakePasswordResetService) ResetPassword(token string, newPassword string) error {
	if f.resetFn != nil {
		return f.resetFn(token, newPassword)
	}
	return nil
}

func setupPasswordResetRouter(t *testing.T, svc interfaces.PasswordResetService) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	pc := controller.NewPasswordResetController(svc, "")
	r := gin.Default()
	r.POST("/auth/forgot-password", pc.Forgot)
	r.POST("/auth/reset-password", pc.Reset)
	return r
}

func TestForgotPassword_InvalidJSON_Returns400(t *testing.T) {
	router := setupPasswordResetRouter(t, &fakePasswordResetService{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestForgotPassword_Success_Returns200(t *testing.T) {
	router := setupPasswordResetRouter(t, &fakePasswordResetService{
		requestResetFn: func(email string) (string, error) {
			return "test-token", nil
		},
	})

	body := []byte(`{"email":"test@example.com"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp, "message")
	assert.NotContains(t, resp, "token")
}

func TestResetPassword_MissingFields_Returns400(t *testing.T) {
	router := setupPasswordResetRouter(t, &fakePasswordResetService{})

	body := []byte(`{"token":"","new_password":""}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestResetPassword_Success_Returns200(t *testing.T) {
	router := setupPasswordResetRouter(t, &fakePasswordResetService{
		resetFn: func(token, pw string) error {
			return nil
		},
	})

	body := []byte(`{"token":"abc","new_password":"Welkom1234"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "password updated")
}
