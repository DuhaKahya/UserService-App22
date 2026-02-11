package tests

import (
	"testing"

	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"
	"group1-userservice/app/repository"
	"group1-userservice/app/service"

	"github.com/stretchr/testify/assert"
)

func TestDiscoveryPrefsService_GetForEmail_ReturnsDefaults_WhenMissing(t *testing.T) {
	db := openTestDB(t)

	err := db.AutoMigrate(&models.DiscoveryPreferences{})
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	repo := repository.NewDiscoveryPreferencesRepository(db)
	svc := service.NewDiscoveryPreferencesService(repo)

	email := "missing@example.com"

	prefs, err := svc.GetForEmail(email)
	assert.NoError(t, err)
	assert.Equal(t, email, prefs.Email)
	assert.Equal(t, 50, prefs.RadiusKm) // default
}

func TestDiscoveryPrefsService_UpdateForEmail_Valid_Upserts(t *testing.T) {
	db := openTestDB(t)

	err := db.AutoMigrate(&models.DiscoveryPreferences{})
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	repo := repository.NewDiscoveryPreferencesRepository(db)
	svc := service.NewDiscoveryPreferencesService(repo)

	email := "update@example.com"

	updated, err := svc.UpdateForEmail(email, interfaces.DiscoveryPreferencesInput{RadiusKm: 120})
	assert.NoError(t, err)
	assert.Equal(t, email, updated.Email)
	assert.Equal(t, 120, updated.RadiusKm)

	// Check persisted in repo
	got, err := repo.GetByEmail(email)
	assert.NoError(t, err)
	assert.Equal(t, email, got.Email)
	assert.Equal(t, 120, got.RadiusKm)
}

func TestDiscoveryPrefsService_UpdateForEmail_InvalidRange_ReturnsError(t *testing.T) {
	db := openTestDB(t)

	err := db.AutoMigrate(&models.DiscoveryPreferences{})
	if err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	repo := repository.NewDiscoveryPreferencesRepository(db)
	svc := service.NewDiscoveryPreferencesService(repo)

	email := "range@example.com"

	_, err = svc.UpdateForEmail(email, interfaces.DiscoveryPreferencesInput{RadiusKm: 0})
	assert.Error(t, err)

	_, err = svc.UpdateForEmail(email, interfaces.DiscoveryPreferencesInput{RadiusKm: 9999})
	assert.Error(t, err)
}
