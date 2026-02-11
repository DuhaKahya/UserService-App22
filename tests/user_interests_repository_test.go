package tests

import (
	"testing"

	"group1-userservice/app/config"
	"group1-userservice/app/models"
	"group1-userservice/app/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupUserInterestsRepositoryTest(t *testing.T) (*repository.UserInterestsRepository, *gorm.DB) {
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

	repo := repository.NewUserInterestsRepository()
	return repo, db
}

func TestUserInterestsRepository_AnyInterestsExist_FalseWhenEmpty(t *testing.T) {
	repo, _ := setupUserInterestsRepositoryTest(t)

	exists, err := repo.AnyInterestsExist()
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserInterestsRepository_CreateInterests_ThenAnyInterestsExist_True(t *testing.T) {
	repo, _ := setupUserInterestsRepositoryTest(t)

	err := repo.CreateInterests([]string{
		"ICT",
		"Onderwijs, Cultuur en Wetenschap",
	})
	assert.NoError(t, err)

	exists, err := repo.AnyInterestsExist()
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestUserInterestsRepository_ListAllInterests_ReturnsOrdered(t *testing.T) {
	repo, db := setupUserInterestsRepositoryTest(t)
	seedInterests(t, db)

	list, err := repo.ListAllInterests()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 3)

	keys := map[string]bool{}
	for _, it := range list {
		keys[it.Key] = true
	}

	assert.True(t, keys["ICT"])
	assert.True(t, keys["Onderwijs, Cultuur en Wetenschap"])
	assert.True(t, keys["Interesse om te Investeren"])
}

func TestUserInterestsRepository_GetUserInterests_EmptyWhenNoRows(t *testing.T) {
	repo, db := setupUserInterestsRepositoryTest(t)
	_ = seedInterests(t, db)

	rows, err := repo.GetUserInterests("missing@example.com")
	assert.NoError(t, err)
	assert.Len(t, rows, 0)
}

func TestUserInterestsRepository_GetUserInterests_ReturnsRowsForUser(t *testing.T) {
	repo, db := setupUserInterestsRepositoryTest(t)
	ints := seedInterests(t, db)

	email := "user10@example.com"

	err := db.Create(&models.UserInterest{
		UserEmail:  email,
		InterestID: ints[0].ID,
		Value:      true,
	}).Error
	assert.NoError(t, err)

	err = db.Create(&models.UserInterest{
		UserEmail:  email,
		InterestID: ints[2].ID,
		Value:      false,
	}).Error
	assert.NoError(t, err)

	rows, err := repo.GetUserInterests(email)
	assert.NoError(t, err)
	assert.Len(t, rows, 2)

	got := map[uint]bool{}
	for _, r := range rows {
		got[r.InterestID] = r.Value
	}

	assert.True(t, got[ints[0].ID])
	assert.False(t, got[ints[2].ID])
}

func TestUserInterestsRepository_UpsertUserInterest_CreateNew(t *testing.T) {
	repo, db := setupUserInterestsRepositoryTest(t)
	ints := seedInterests(t, db)

	email := "user42@example.com"

	row := &models.UserInterest{
		UserEmail:  email,
		InterestID: ints[1].ID,
		Value:      true,
	}

	err := repo.UpsertUserInterest(row)
	assert.NoError(t, err)

	var saved models.UserInterest
	err = db.First(&saved, "user_email = ? AND interest_id = ?", email, ints[1].ID).Error
	assert.NoError(t, err)
	assert.True(t, saved.Value)
}

func TestUserInterestsRepository_UpsertUserInterest_UpdateExisting(t *testing.T) {
	repo, db := setupUserInterestsRepositoryTest(t)
	ints := seedInterests(t, db)

	email := "user50@example.com"

	initial := &models.UserInterest{
		UserEmail:  email,
		InterestID: ints[0].ID,
		Value:      true,
	}
	assert.NoError(t, db.Create(initial).Error)

	initial.Value = false
	err := repo.UpsertUserInterest(initial)
	assert.NoError(t, err)

	var saved models.UserInterest
	err = db.First(&saved, "user_email = ? AND interest_id = ?", email, ints[0].ID).Error
	assert.NoError(t, err)
	assert.False(t, saved.Value)
}
