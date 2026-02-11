package tests

import (
	"testing"

	"group1-userservice/app/config"
	"group1-userservice/app/models"
	"group1-userservice/app/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// NotificationSettingsRepository test setup
func setupNotificationSettingsRepositoryTest(t *testing.T) (*repository.NotificationSettingsRepository, *gorm.DB) {
	t.Helper()

	db := openTestDB(t)
	config.DB = db

	if err := config.DB.AutoMigrate(&models.NotificationSettings{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := repository.NewNotificationSettingsRepository()
	return repo, db
}

func uniqueEmail(prefix string) string {
	return prefix + "_" + uuid.New().String() + "@example.com"
}

// GetByEmail

func TestNotificationSettingsRepository_GetByEmail_Found(t *testing.T) {
	repo, db := setupNotificationSettingsRepositoryTest(t)
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

	result, err := repo.GetByEmail(email)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, email, result.UserEmail)
	assert.False(t, result.LikeEmail)
	assert.False(t, result.LikePush)
	assert.True(t, result.FavoriteEmail)
	assert.False(t, result.FavoritePush)
	assert.Equal(t, "expo-existing", result.ExpoPushToken)
}

func TestNotificationSettingsRepository_GetByEmail_NotFound(t *testing.T) {
	repo, db := setupNotificationSettingsRepositoryTest(t)
	wipeNotificationSettings(t, db)

	result, err := repo.GetByEmail(uniqueEmail("missing"))

	assert.Error(t, err)
	assert.Nil(t, result)
}

// Upsert (by email)

func TestNotificationSettingsRepository_Upsert_CreateNew(t *testing.T) {
	repo, db := setupNotificationSettingsRepositoryTest(t)
	wipeNotificationSettings(t, db)

	email := uniqueEmail("new")

	settings := &models.NotificationSettings{
		UserEmail:     email,
		LikeEmail:     true,
		LikePush:      false,
		FavoriteEmail: true,
		FavoritePush:  false,
		ChatEmail:     true,
		ChatPush:      true,
		SystemEmail:   false,
		SystemPush:    true,
		ExpoPushToken: "expo-new",
	}

	result, err := repo.Upsert(settings)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotZero(t, result.ID)
	assert.Equal(t, email, result.UserEmail)

	var saved models.NotificationSettings
	if err := db.First(&saved, "user_email = ?", email).Error; err != nil {
		t.Fatalf("failed to load saved settings: %v", err)
	}

	assert.Equal(t, result.ID, saved.ID)
	assert.Equal(t, "expo-new", saved.ExpoPushToken)
	assert.True(t, saved.LikeEmail)
	assert.False(t, saved.LikePush)
}

func TestNotificationSettingsRepository_Upsert_UpdateExisting(t *testing.T) {
	repo, db := setupNotificationSettingsRepositoryTest(t)
	wipeNotificationSettings(t, db)

	email := uniqueEmail("update")

	existing := &models.NotificationSettings{
		UserEmail:     email,
		LikeEmail:     true,
		LikePush:      true,
		FavoriteEmail: false,
		FavoritePush:  false,
		ChatEmail:     false,
		ChatPush:      false,
		SystemEmail:   true,
		SystemPush:    true,
		ExpoPushToken: "expo-old",
	}
	if err := db.Create(existing).Error; err != nil {
		t.Fatalf("failed to create existing settings: %v", err)
	}

	update := &models.NotificationSettings{
		UserEmail:     email,
		LikeEmail:     false,
		LikePush:      false,
		FavoriteEmail: true,
		FavoritePush:  true,
		ChatEmail:     true,
		ChatPush:      true,
		SystemEmail:   false,
		SystemPush:    false,
		ExpoPushToken: "expo-updated",
	}

	result, err := repo.Upsert(update)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, email, result.UserEmail)

	var saved models.NotificationSettings
	if err := db.First(&saved, "user_email = ?", email).Error; err != nil {
		t.Fatalf("failed to load saved settings: %v", err)
	}

	assert.Equal(t, "expo-updated", saved.ExpoPushToken)
	assert.False(t, saved.LikeEmail)
	assert.False(t, saved.LikePush)
	assert.True(t, saved.FavoriteEmail)
	assert.True(t, saved.FavoritePush)
	assert.True(t, saved.ChatEmail)
	assert.True(t, saved.ChatPush)
	assert.False(t, saved.SystemEmail)
	assert.False(t, saved.SystemPush)
}
