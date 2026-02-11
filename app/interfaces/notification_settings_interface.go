package interfaces

import "group1-userservice/app/models"

type NotificationSettingsService interface {
	GetByEmail(email string) (*models.NotificationSettings, error)
	UpdateForEmail(email string, input NotificationSettingsInput) (*models.NotificationSettings, error)
	Upsert(settings *models.NotificationSettings) (*models.NotificationSettings, error)
}

type NotificationSettingsInput struct {
	LikeEmail     bool   `json:"like_email"`
	LikePush      bool   `json:"like_push"`
	FavoriteEmail bool   `json:"favorite_email"`
	FavoritePush  bool   `json:"favorite_push"`
	ChatEmail     bool   `json:"chat_email"`
	ChatPush      bool   `json:"chat_push"`
	SystemEmail   bool   `json:"system_email"`
	SystemPush    bool   `json:"system_push"`
	ExpoPushToken string `json:"expo_push_token"`
}

type NotificationSettingsPatchInput struct {
	LikeEmail     *bool `json:"like_email,omitempty"`
	LikePush      *bool `json:"like_push,omitempty"`
	FavoriteEmail *bool `json:"favorite_email,omitempty"`
	FavoritePush  *bool `json:"favorite_push,omitempty"`
	ChatEmail     *bool `json:"chat_email,omitempty"`
	ChatPush      *bool `json:"chat_push,omitempty"`

	SystemEmail *bool `json:"system_email,omitempty"`
	SystemPush  *bool `json:"system_push,omitempty"`

	ExpoPushToken *string `json:"expo_push_token,omitempty"`
}
