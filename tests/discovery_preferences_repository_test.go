package tests

import (
	"testing"

	"group1-userservice/app/models"
	"group1-userservice/app/repository"

	"github.com/stretchr/testify/assert"
)

func TestDiscoveryPrefsRepo_Upsert_And_GetByEmail(t *testing.T) {
	db := openTestDB(t)

	err := db.AutoMigrate(&models.DiscoveryPreferences{})
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	repo := repository.NewDiscoveryPreferencesRepository(db)

	email := "repo-test@example.com"

	// Upsert first time (insert)
	p1, err := repo.Upsert(&models.DiscoveryPreferences{
		Email:    email,
		RadiusKm: 30,
	})
	assert.NoError(t, err)
	assert.Equal(t, email, p1.Email)
	assert.Equal(t, 30, p1.RadiusKm)

	// Upsert again (update existing row)
	p2, err := repo.Upsert(&models.DiscoveryPreferences{
		Email:    email,
		RadiusKm: 80,
	})
	assert.NoError(t, err)
	assert.Equal(t, email, p2.Email)
	assert.Equal(t, 80, p2.RadiusKm)

	// GetByEmail should return updated value
	got, err := repo.GetByEmail(email)
	assert.NoError(t, err)
	assert.Equal(t, email, got.Email)
	assert.Equal(t, 80, got.RadiusKm)
}
