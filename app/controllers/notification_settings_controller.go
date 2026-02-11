package controller

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"group1-userservice/app/interfaces"
	"group1-userservice/app/middleware"
)

type NotificationSettingsController struct {
	Service     interfaces.NotificationSettingsService
	UserService interfaces.UserService
}

func NewNotificationSettingsController(
	s interfaces.NotificationSettingsService,
	us interfaces.UserService,
) *NotificationSettingsController {
	return &NotificationSettingsController{
		Service:     s,
		UserService: us,
	}
}

// @Summary Get notification settings by email (internal)
// @Description Internal endpoint for services to get notification settings by email
// @Tags Notifications
// @Produce json
// @Param X-Service-Token header string true "Service token"
// @Param email path string true "User email"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /internal/users/{email}/notification-settings [get]
func (nc *NotificationSettingsController) GetByEmailInternal(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing email"})
		return
	}

	// Verify user exists
	_, err := nc.UserService.GetByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	settings, err := nc.Service.GetByEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": gin.H{
			"like":         gin.H{"email": settings.LikeEmail, "push": settings.LikePush},
			"favorite":     gin.H{"email": settings.FavoriteEmail, "push": settings.FavoritePush},
			"chat_message": gin.H{"email": settings.ChatEmail, "push": settings.ChatPush},
			"system_alert": gin.H{"email": settings.SystemEmail, "push": settings.SystemPush},
		},
		"expo_push_token": settings.ExpoPushToken,
	})
}

// @Summary Update notification settings
// @Description Update notification settings for the authenticated user.
// System alerts cannot be modified and must not be included in the request body.
// @Tags Notifications
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param body body interfaces.NotificationSettingsPatchInput true "Updated settings"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/me/notification-settings [put]
func (nc *NotificationSettingsController) UpdateForMe(c *gin.Context) {
	sub, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := nc.UserService.GetByKeycloakID(sub)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Read raw body once
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	// Strict decode: typos / unknown fields / type errors => 400
	var patch interfaces.NotificationSettingsPatchInput
	dec := json.NewDecoder(bytes.NewReader(bodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input", "details": err.Error()})
		return
	}

	// Reject system fields if present
	if patch.SystemEmail != nil || patch.SystemPush != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "system_alert settings cannot be modified"})
		return
	}

	// Load current settings
	current, err := nc.Service.GetByEmail(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load settings"})
		return
	}

	// Merge only provided fields
	if patch.LikeEmail != nil {
		current.LikeEmail = *patch.LikeEmail
	}
	if patch.LikePush != nil {
		current.LikePush = *patch.LikePush
	}
	if patch.FavoriteEmail != nil {
		current.FavoriteEmail = *patch.FavoriteEmail
	}
	if patch.FavoritePush != nil {
		current.FavoritePush = *patch.FavoritePush
	}
	if patch.ChatEmail != nil {
		current.ChatEmail = *patch.ChatEmail
	}
	if patch.ChatPush != nil {
		current.ChatPush = *patch.ChatPush
	}

	if patch.ExpoPushToken != nil {
		current.ExpoPushToken = strings.TrimSpace(*patch.ExpoPushToken)
	}

	updated, err := nc.Service.Upsert(current)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": gin.H{
			"like":         gin.H{"email": updated.LikeEmail, "push": updated.LikePush},
			"favorite":     gin.H{"email": updated.FavoriteEmail, "push": updated.FavoritePush},
			"chat_message": gin.H{"email": updated.ChatEmail, "push": updated.ChatPush},
		},
		"expo_push_token": updated.ExpoPushToken,
	})
}
