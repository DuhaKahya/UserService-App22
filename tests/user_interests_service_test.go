package tests

import (
	"testing"

	"group1-userservice/app/config"
	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"
	"group1-userservice/app/repository"
	"group1-userservice/app/service"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupUserInterestsServiceTest(t *testing.T) (interfaces.UserInterestsService, *gorm.DB, *repository.UserInterestsRepository) {
	t.Helper()

	db := openTestDB(t)
	config.DB = db

	_ = config.DB.Migrator().DropTable(&models.UserInterest{}, &models.Interest{})

	if err := config.DB.AutoMigrate(
		&models.Interest{},
		&models.UserInterest{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	config.SeedInterests()

	repo := repository.NewUserInterestsRepository()
	svc := service.NewUserInterestsService(repo)

	return svc, db, repo
}

func TestUserInterests_GetForUser_ReturnsDefaultsWhenNoRows(t *testing.T) {
	svc, _, _ := setupUserInterestsServiceTest(t)

	email := "user55@example.com"

	result, err := svc.GetForUser(email)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result), 12)

	foundICT := false
	foundMedia := false

	for _, item := range result {
		if item.Key == "ICT" {
			foundICT = true
			assert.False(t, item.Value)
			assert.NotZero(t, item.ID)
		}
		if item.Key == "Media en Communicatie" {
			foundMedia = true
			assert.False(t, item.Value)
			assert.NotZero(t, item.ID)
		}
	}

	assert.True(t, foundICT)
	assert.True(t, foundMedia)
}

func TestUserInterests_GetForUser_ReturnsExistingValues(t *testing.T) {
	svc, db, _ := setupUserInterestsServiceTest(t)

	ictID := getInterestIDByKey(t, db, "ICT")
	mediaID := getInterestIDByKey(t, db, "Media en Communicatie")

	email := "user10@example.com"

	assert.NoError(t, db.Create(&models.UserInterest{
		UserEmail:  email,
		InterestID: ictID,
		Value:      true,
	}).Error)

	assert.NoError(t, db.Create(&models.UserInterest{
		UserEmail:  email,
		InterestID: mediaID,
		Value:      false,
	}).Error)

	result, err := svc.GetForUser(email)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result), 12)

	var ictVal *bool
	var mediaVal *bool

	for _, item := range result {
		if item.Key == "ICT" {
			v := item.Value
			ictVal = &v
		}
		if item.Key == "Media en Communicatie" {
			v := item.Value
			mediaVal = &v
		}
	}

	if assert.NotNil(t, ictVal) {
		assert.True(t, *ictVal)
	}
	if assert.NotNil(t, mediaVal) {
		assert.False(t, *mediaVal)
	}
}

func TestUserInterests_UpdateForUser_UpdatesAndReturnsMergedList(t *testing.T) {
	svc, db, _ := setupUserInterestsServiceTest(t)

	ictID := getInterestIDByKey(t, db, "ICT")
	eduID := getInterestIDByKey(t, db, "Onderwijs, Cultuur en Wetenschap")
	mediaID := getInterestIDByKey(t, db, "Media en Communicatie")

	input := interfaces.UserInterestsUpdateInput{
		Interests: []interfaces.UserInterestItemInput{
			{ID: ictID, Value: true},
			{ID: eduID, Value: true},
			{ID: mediaID, Value: false},
		},
	}

	email := "user99@example.com"

	result, err := svc.UpdateForUser(email, input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result), 12)

	var ictVal *bool
	var eduVal *bool
	var mediaVal *bool

	for _, item := range result {
		if item.ID == ictID {
			v := item.Value
			ictVal = &v
			assert.Equal(t, "ICT", item.Key)
		}
		if item.ID == eduID {
			v := item.Value
			eduVal = &v
			assert.Equal(t, "Onderwijs, Cultuur en Wetenschap", item.Key)
		}
		if item.ID == mediaID {
			v := item.Value
			mediaVal = &v
			assert.Equal(t, "Media en Communicatie", item.Key)
		}
	}

	if assert.NotNil(t, ictVal) {
		assert.True(t, *ictVal)
	}
	if assert.NotNil(t, eduVal) {
		assert.True(t, *eduVal)
	}
	if assert.NotNil(t, mediaVal) {
		assert.False(t, *mediaVal)
	}

	var savedICT models.UserInterest
	err = db.First(&savedICT, "user_email = ? AND interest_id = ?", email, ictID).Error
	assert.NoError(t, err)
	assert.True(t, savedICT.Value)

	var savedEdu models.UserInterest
	err = db.First(&savedEdu, "user_email = ? AND interest_id = ?", email, eduID).Error
	assert.NoError(t, err)
	assert.True(t, savedEdu.Value)

	var savedMedia models.UserInterest
	err = db.First(&savedMedia, "user_email = ? AND interest_id = ?", email, mediaID).Error
	assert.NoError(t, err)
	assert.False(t, savedMedia.Value)
}
