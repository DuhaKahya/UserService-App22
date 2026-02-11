package service

import (
	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"
	"group1-userservice/app/repository"
)

type notificationSettingsService struct {
	repo *repository.NotificationSettingsRepository
}

func NewNotificationSettingsService(repo *repository.NotificationSettingsRepository) interfaces.NotificationSettingsService {
	return &notificationSettingsService{repo: repo}
}

func (s *notificationSettingsService) GetByEmail(email string) (*models.NotificationSettings, error) {
	settings, err := s.repo.GetByEmail(email)
	if err == nil {
		return settings, nil
	}

	// defaults
	defaults := &models.NotificationSettings{
		UserEmail:     email,
		LikeEmail:     true,
		LikePush:      true,
		FavoriteEmail: true,
		FavoritePush:  true,
		ChatEmail:     true,
		ChatPush:      true,
		SystemEmail:   true,
		SystemPush:    false,
	}

	return s.repo.Upsert(defaults)
}

func (s *notificationSettingsService) UpdateForEmail(email string, input interfaces.NotificationSettingsInput) (*models.NotificationSettings, error) {
	settings := &models.NotificationSettings{
		UserEmail:     email,
		LikeEmail:     input.LikeEmail,
		LikePush:      input.LikePush,
		FavoriteEmail: input.FavoriteEmail,
		FavoritePush:  input.FavoritePush,
		ChatEmail:     input.ChatEmail,
		ChatPush:      input.ChatPush,
		SystemEmail:   input.SystemEmail,
		SystemPush:    input.SystemPush,
		ExpoPushToken: input.ExpoPushToken,
	}

	return s.repo.Upsert(settings)
}

func (s *notificationSettingsService) Upsert(settings *models.NotificationSettings) (*models.NotificationSettings, error) {
	return s.repo.Upsert(settings)
}
