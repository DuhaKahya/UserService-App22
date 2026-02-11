package controller

import (
	"group1-userservice/app/interfaces"
	"group1-userservice/app/keycloak"

	"github.com/gin-gonic/gin"

	"group1-userservice/app/metrics"
	"time"
)

type LoginController struct {
	UserService interfaces.UserService
}

func NewLoginController(us interfaces.UserService) *LoginController {
	return &LoginController{
		UserService: us,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// @Summary User login
// @Description Authenticate a user and return a Keycloak access token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} keycloak.TokenResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (lc *LoginController) Handle(c *gin.Context) {
	start := time.Now()
	defer func() {
		metrics.UserRequestDuration.Observe(time.Since(start).Seconds())
	}()

	metrics.UserRequestsTotal.Inc()

	var req LoginRequest

	if err := c.BindJSON(&req); err != nil {
		metrics.UserRequestOutcomesTotal.WithLabelValues("bad_request").Inc()
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	// Check user exists in your database
	user, err := lc.UserService.GetByEmail(req.Email)
	if err != nil {
		metrics.UserRequestOutcomesTotal.WithLabelValues("unauthorized").Inc()
		c.JSON(401, gin.H{"error": "Invalid email or password"})
		return
	}

	// Verify password using bcrypt
	if !lc.UserService.CheckPassword(user.Password, req.Password) {
		metrics.UserRequestOutcomesTotal.WithLabelValues("unauthorized").Inc()
		c.JSON(401, gin.H{"error": "Invalid email or password"})
		return
	}

	// Ask Keycloak for an access token (auth)
	token, err := keycloak.GetAccessToken(req.Email, req.Password)
	if err != nil {
		metrics.UserRequestOutcomesTotal.WithLabelValues("unauthorized").Inc()
		c.JSON(401, gin.H{"error": "Authentication failed"})
		return
	}

	metrics.UserRequestOutcomesTotal.WithLabelValues("success").Inc()
	c.JSON(200, token)
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// @Summary Refresh access token
// @Description Use refresh token to get a new access token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} keycloak.TokenResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (lc *LoginController) Refresh(c *gin.Context) {
	var req RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(400, gin.H{"error": "refresh_token is required"})
		return
	}

	token, err := keycloak.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	c.JSON(200, token)
}
