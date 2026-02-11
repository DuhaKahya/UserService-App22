package repository

import (
	"group1-userservice/app/config"
	"group1-userservice/app/models"
)

type NotificationSettingsRepository struct{}

func NewNotificationSettingsRepository() *NotificationSettingsRepository {
	return &NotificationSettingsRepository{}
}

func (r *NotificationSettingsRepository) GetByEmail(email string) (*models.NotificationSettings, error) {
	var s models.NotificationSettings
	err := config.DB.Where("user_email = ?", email).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *NotificationSettingsRepository) Upsert(s *models.NotificationSettings) (*models.NotificationSettings, error) {
	var existing models.NotificationSettings

	err := config.DB.Where("user_email = ?", s.UserEmail).First(&existing).Error
	if err == nil {
		existing.LikeEmail = s.LikeEmail
		existing.LikePush = s.LikePush
		existing.FavoriteEmail = s.FavoriteEmail
		existing.FavoritePush = s.FavoritePush
		existing.ChatEmail = s.ChatEmail
		existing.ChatPush = s.ChatPush
		existing.SystemEmail = s.SystemEmail
		existing.SystemPush = s.SystemPush
		existing.ExpoPushToken = s.ExpoPushToken

		config.DB.Save(&existing)
		return &existing, nil
	}

	config.DB.Create(s)
	return s, nil
}
