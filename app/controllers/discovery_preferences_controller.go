package controller

import (
	"net/http"

	"group1-userservice/app/interfaces"
	"group1-userservice/app/middleware"

	"github.com/gin-gonic/gin"
)

type DiscoveryPreferencesController struct {
	PrefsService interfaces.DiscoveryPreferencesService
	UserService  interfaces.UserService
}

// DiscoveryPreferencesResponse is the API response format
type DiscoveryPreferencesResponse struct {
	Email    string `json:"email"`
	RadiusKm int    `json:"radius_km"`
}

func NewDiscoveryPreferencesController(
	prefs interfaces.DiscoveryPreferencesService,
	userService interfaces.UserService,
) *DiscoveryPreferencesController {
	return &DiscoveryPreferencesController{
		PrefsService: prefs,
		UserService:  userService,
	}
}

// @Summary Get discovery preferences for current user
// @Description Returns the discovery preferences (km radius) for the authenticated user. If none exist, defaults are returned.
// @Tags DiscoveryPreferences
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Success 200 {object} controller.DiscoveryPreferencesResponse
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/me/discovery-preferences [get]
func (dc *DiscoveryPreferencesController) GetForMe(c *gin.Context) {
	sub, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := dc.UserService.GetByKeycloakID(sub)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	prefs, err := dc.PrefsService.GetForEmail(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get discovery preferences"})
		return
	}

	c.JSON(http.StatusOK, DiscoveryPreferencesResponse{
		Email:    user.Email,
		RadiusKm: prefs.RadiusKm,
	})
}

// @Summary Update discovery preferences for current user
// @Description Updates the discovery preferences (km radius) for the authenticated user.
// @Tags DiscoveryPreferences
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param request body interfaces.DiscoveryPreferencesInput true "Discovery preferences input"
// @Success 200 {object} controller.DiscoveryPreferencesResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/me/discovery-preferences [put]
func (dc *DiscoveryPreferencesController) UpdateForMe(c *gin.Context) {
	sub, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := dc.UserService.GetByKeycloakID(sub)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var input interfaces.DiscoveryPreferencesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	updated, err := dc.PrefsService.UpdateForEmail(user.Email, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DiscoveryPreferencesResponse{
		Email:    user.Email,
		RadiusKm: updated.RadiusKm,
	})
}

// @Summary Get discovery preferences by email (internal)
// @Description Internal endpoint - requires X-Service-Token. Returns discovery preferences for a user by email.
// @Tags DiscoveryPreferences
// @Produce json
// @Param X-Service-Token header string true "Service token"
// @Param email path string true "User email"
// @Success 200 {object} controller.DiscoveryPreferencesResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /internal/users/{email}/discovery-preferences [get]
func (dc *DiscoveryPreferencesController) GetByEmailInternal(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing email"})
		return
	}

	// Optioneel: check of user bestaat (zoals je al deed)
	if _, err := dc.UserService.GetByEmail(email); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	prefs, err := dc.PrefsService.GetForEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get discovery preferences"})
		return
	}

	c.JSON(http.StatusOK, DiscoveryPreferencesResponse{
		Email:    email,
		RadiusKm: prefs.RadiusKm,
	})
}
