package controller

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"group1-userservice/app/interfaces"

	"github.com/gin-gonic/gin"

	"group1-userservice/app/metrics"
)

type PasswordResetController struct {
	Service         interfaces.PasswordResetService
	NotificationURL string
}

func NewPasswordResetController(
	s interfaces.PasswordResetService,
	notificationURL string,
) *PasswordResetController {
	return &PasswordResetController{
		Service:         s,
		NotificationURL: notificationURL,
	}
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

func (pc *PasswordResetController) sendPasswordResetAlert(email string, token string) {
	if pc.NotificationURL == "" {
		log.Println("[notification] NOTIFICATION_SERVICE_URL is empty, skipping")
		return
	}

	metrics.NotificationCallsTotal.WithLabelValues("attempt").Inc()

	payload := map[string]interface{}{
		"email":     email,
		"message":   "A password reset was requested for your account. If this was you, please follow the reset instructions.",
		"title":     "Password reset requested",
		"type":      "system_alert",
		"token":     token,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[notification] marshal error: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", pc.NotificationURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("[notification] request error for %s: %v\n", pc.NotificationURL, err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	// Service-to-service authentication
	serviceToken := os.Getenv("NOTIFICATION_SERVICE_TOKEN")
	if serviceToken != "" {
		req.Header.Set("X-Service-Token", serviceToken)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		metrics.NotificationCallsTotal.WithLabelValues("error").Inc()
		log.Printf("[notification] http error calling %s: %v\n", pc.NotificationURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		metrics.NotificationCallsTotal.WithLabelValues("success").Inc()
	} else {
		metrics.NotificationCallsTotal.WithLabelValues("failed").Inc()
	}

	log.Printf("[notification] called %s -> status %d\n", pc.NotificationURL, resp.StatusCode)
}

// Forgot
// @Summary Request a password reset
// @Description If the email exists, a reset message will be sent.
// Always returns 200 to prevent user enumeration.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body controller.ForgotPasswordRequest true "Forgot password request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /auth/forgot-password [post]
func (pc *PasswordResetController) Forgot(c *gin.Context) {
	start := time.Now()
	defer func() {
		metrics.UserRequestDuration.Observe(time.Since(start).Seconds())
	}()

	metrics.UserRequestsTotal.Inc()

	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		metrics.UserRequestOutcomesTotal.WithLabelValues("bad_request").Inc()
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	metrics.PasswordResetRequestsTotal.Inc()

	// Request password reset token from the service layer
	token, _ := pc.Service.RequestReset(req.Email)

	// Send notification asynchronously only when a token is created
	if token != "" {
		go pc.sendPasswordResetAlert(req.Email, token)
	}

	metrics.UserRequestOutcomesTotal.WithLabelValues("success").Inc()

	c.JSON(http.StatusOK, gin.H{
		"message": "if the email exists, a reset link will be sent",
	})
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// Reset
// @Summary Reset password using a token
// @Description Resets the user's password using the provided token and new password.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body controller.ResetPasswordRequest true "Reset password request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/reset-password [post]
func (pc *PasswordResetController) Reset(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token and new_password are required"})
		return
	}

	if err := pc.Service.ResetPassword(req.Token, req.NewPassword); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}
