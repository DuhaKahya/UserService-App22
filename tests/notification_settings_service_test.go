package tests

import (
	"testing"

	"group1-userservice/app/config"
	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"
	"group1-userservice/app/repository"
	"group1-userservice/app/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// NotificationSettings service test setup
func setupNotificationSettingsServiceTest(t *testing.T) (interfaces.NotificationSettingsService, *gorm.DB, *repository.NotificationSettingsRepository) {
	t.Helper()

	db := openTestDB(t)
	config.DB = db

	if err := config.DB.AutoMigrate(&models.NotificationSettings{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := repository.NewNotificationSettingsRepository()
	svc := service.NewNotificationSettingsService(repo)

	return svc, db, repo
}

// GetByEmail

func TestNotificationSettings_GetByEmail_ReturnsExisting(t *testing.T) {
	svc, db, _ := setupNotificationSettingsServiceTest(t)
	wipeNotificationSettings(t, db)

	email := uniqueEmail("notif")

	existing := &models.NotificationSettings{
		UserEmail:     email,
		LikeEmail:     false,
		LikePush:      false,
		FavoriteEmail: true,
		FavoritePush:  false,
		ChatEmail:     true,
		ChatPush:      false,
		SystemEmail:   true,
		SystemPush:    false,
		ExpoPushToken: "expo-existing",
	}
	if err := db.Create(existing).Error; err != nil {
		t.Fatalf("failed to create existing settings: %v", err)
	}

	result, err := svc.GetByEmail(email)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, email, result.UserEmail)
	assert.False(t, result.LikeEmail)
	assert.False(t, result.LikePush)
	assert.True(t, result.FavoriteEmail)
	assert.False(t, result.FavoritePush)
	assert.Equal(t, "expo-existing", result.ExpoPushToken)
}

func TestNotificationSettings_GetByEmail_CreatesDefault(t *testing.T) {
	svc, db, _ := setupNotificationSettingsServiceTest(t)
	wipeNotificationSettings(t, db)

	email := "new_" + uuid.New().String() + "@example.com"

	result, err := svc.GetByEmail(email)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, email, result.UserEmail)

	// defaults (zoals in service)
	assert.True(t, result.LikeEmail)
	assert.True(t, result.LikePush)
	assert.True(t, result.FavoriteEmail)
	assert.True(t, result.FavoritePush)
	assert.True(t, result.ChatEmail)
	assert.True(t, result.ChatPush)
	assert.True(t, result.SystemEmail)

	// service default is false
	assert.False(t, result.SystemPush)

	// expo token default (zero value)
	assert.Equal(t, "", result.ExpoPushToken)

	// check dat defaults ook echt persisted zijn
	var dbResult models.NotificationSettings
	err2 := db.First(&dbResult, "user_email = ?", email).Error
	assert.NoError(t, err2)
	assert.Equal(t, email, dbResult.UserEmail)

	// extra: persisted value is also false
	assert.False(t, dbResult.SystemPush)
}

// UpdateForEmail

func TestNotificationSettings_UpdateForEmail(t *testing.T) {
	svc, db, _ := setupNotificationSettingsServiceTest(t)
	wipeNotificationSettings(t, db)

	email := "update_" + uuid.New().String() + "@example.com"

	input := interfaces.NotificationSettingsInput{
		LikeEmail:     false,
		LikePush:      true,
		FavoriteEmail: false,
		FavoritePush:  true,
		ChatEmail:     true,
		ChatPush:      false,
		SystemEmail:   false,
		SystemPush:    true,
		ExpoPushToken: "expo-token-123",
	}

	result, err := svc.UpdateForEmail(email, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, email, result.UserEmail)

	assert.False(t, result.LikeEmail)
	assert.True(t, result.LikePush)
	assert.False(t, result.FavoriteEmail)
	assert.True(t, result.FavoritePush)
	assert.True(t, result.ChatEmail)
	assert.False(t, result.ChatPush)
	assert.False(t, result.SystemEmail)
	assert.True(t, result.SystemPush)
	assert.Equal(t, "expo-token-123", result.ExpoPushToken)

	var saved models.NotificationSettings
	err2 := db.First(&saved, "user_email = ?", email).Error
	assert.NoError(t, err2)

	assert.Equal(t, result.UserEmail, saved.UserEmail)
	assert.Equal(t, result.LikeEmail, saved.LikeEmail)
	assert.Equal(t, result.LikePush, saved.LikePush)
	assert.Equal(t, result.ExpoPushToken, saved.ExpoPushToken)
}
