package tests

import (
	"testing"

	"group1-userservice/app/config"
	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"
	"group1-userservice/app/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setup helper for UserRepository tests
func setupUserRepositoryTest(t *testing.T) (interfaces.UserRepository, *gorm.DB) {
	t.Helper()

	db := openTestDB(t)
	config.DB = db

	if err := config.DB.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	repo := repository.NewUserRepository(config.DB)
	return repo, db
}

// Create

func TestUserRepository_Create(t *testing.T) {
	repo, db := setupUserRepositoryTest(t)

	user := &models.User{
		Email:      "create@example.com",
		Password:   "hash",
		KeycloakID: "kc-create",
		FirstName:  "Test",
		LastName:   "User",
	}

	err := repo.Create(user)
	assert.NoError(t, err)

	var saved models.User
	db.First(&saved, "email = ?", "create@example.com")
	assert.Equal(t, "kc-create", saved.KeycloakID)
}

// FindByEmail

func TestUserRepository_FindByEmail_Found(t *testing.T) {
	repo, db := setupUserRepositoryTest(t)

	db.Create(&models.User{
		Email:      "find@example.com",
		Password:   "hash",
		KeycloakID: "kc-find",
		FirstName:  "Test",
		LastName:   "User",
	})

	user, err := repo.FindByEmail("find@example.com")

	assert.NoError(t, err)
	assert.Equal(t, "kc-find", user.KeycloakID)
}

func TestUserRepository_FindByEmail_NotFound(t *testing.T) {
	repo, _ := setupUserRepositoryTest(t)

	_, err := repo.FindByEmail("unknown@example.com")

	assert.Error(t, err)
}

// ExistsByEmail

func TestUserRepository_ExistsByEmail_True(t *testing.T) {
	repo, db := setupUserRepositoryTest(t)

	db.Create(&models.User{
		Email:      "exists@example.com",
		KeycloakID: "kc-repository-exists-email-true",
		FirstName:  "Test",
		LastName:   "User",
	})

	exists := repo.ExistsByEmail("exists@example.com")
	assert.True(t, exists)
}

func TestUserRepository_ExistsByEmail_False(t *testing.T) {
	repo, _ := setupUserRepositoryTest(t)

	exists := repo.ExistsByEmail("doesnotexist@example.com")
	assert.False(t, exists)
}

// FindByKeycloakID

func TestUserRepository_FindByKeycloakID_Found(t *testing.T) {
	repo, db := setupUserRepositoryTest(t)

	db.Create(&models.User{
		Email:      "kc@example.com",
		Password:   "hash",
		KeycloakID: "kc-test",
		FirstName:  "Test",
		LastName:   "User",
	})

	user, err := repo.FindByKeycloakID("kc-test")

	assert.NoError(t, err)
	assert.Equal(t, "kc@example.com", user.Email)
}

func TestUserRepository_FindByKeycloakID_NotFound(t *testing.T) {
	repo, _ := setupUserRepositoryTest(t)

	_, err := repo.FindByKeycloakID("missing-kc")

	assert.Error(t, err)
}

// FindAll

func TestUserRepository_FindAll(t *testing.T) {
	repo, db := setupUserRepositoryTest(t)

	db.Create(&models.User{
		Email:      "u1@example.com",
		KeycloakID: "kc-repo-findall-1",
		FirstName:  "Test",
		LastName:   "User",
	})
	db.Create(&models.User{
		Email:      "u2@example.com",
		KeycloakID: "kc-repo-findall-2",
		FirstName:  "Test",
		LastName:   "User",
	})

	users, err := repo.FindAll()
	assert.NoError(t, err)

	emailSet := make(map[string]bool)
	for _, u := range users {
		emailSet[u.Email] = true
	}

	assert.True(t, emailSet["u1@example.com"])
	assert.True(t, emailSet["u2@example.com"])
}

func TestUserRepository_FindByFirstLastInsensitive_Found(t *testing.T) {
	repo, db := setupUserRepositoryTest(t)

	db.Exec(`DELETE FROM users`)

	db.Create(&models.User{
		Email:      "case@example.com",
		Password:   "hash",
		KeycloakID: "kc-case",
		FirstName:  "CaseyUnique",
		LastName:   "TestUnique",
	})

	user, err := repo.FindByFirstLastInsensitive("caseyunique", "testunique")

	assert.NoError(t, err)
	assert.Equal(t, "case@example.com", user.Email)
}

func TestUserRepository_UpdateFieldsByEmail_UpdatesAndReturnsUser(t *testing.T) {
	repo, db := setupUserRepositoryTest(t)

	db.Create(&models.User{
		Email:      "upd@example.com",
		Password:   "hash",
		KeycloakID: "kc-upd",
		FirstName:  "Old",
		LastName:   "Name",
		Country:    "NL",
	})

	fields := map[string]any{
		"first_name": "New",
		"biography":  "Hello",
	}

	updated, err := repo.UpdateFieldsByEmail("upd@example.com", fields)
	assert.NoError(t, err)

	assert.Equal(t, "upd@example.com", updated.Email)
	assert.Equal(t, "New", updated.FirstName)
	assert.Equal(t, "Name", updated.LastName)
	assert.Equal(t, "Hello", updated.Biography)
	assert.Equal(t, "NL", updated.Country)

	var fromDB models.User
	err = db.First(&fromDB, "email = ?", "upd@example.com").Error
	assert.NoError(t, err)
	assert.Equal(t, "New", fromDB.FirstName)
	assert.Equal(t, "Hello", fromDB.Biography)
	assert.Equal(t, "NL", fromDB.Country)
}

func TestUserRepository_UpdateFieldsByEmail_EmptyFields_ReturnsExistingUser(t *testing.T) {
	repo, db := setupUserRepositoryTest(t)

	db.Create(&models.User{
		Email:      "emptyfields@example.com",
		Password:   "hash",
		KeycloakID: "kc-empty",
		FirstName:  "A",
		LastName:   "B",
	})

	updated, err := repo.UpdateFieldsByEmail("emptyfields@example.com", map[string]any{})
	assert.NoError(t, err)
	assert.Equal(t, "emptyfields@example.com", updated.Email)
	assert.Equal(t, "A", updated.FirstName)
	assert.Equal(t, "B", updated.LastName)
}

func TestUserRepository_UpdateFieldsByEmail_UserNotFound(t *testing.T) {
	repo, _ := setupUserRepositoryTest(t)

	_, err := repo.UpdateFieldsByEmail("missing@example.com", map[string]any{"first_name": "X"})
	assert.Error(t, err)
}

func TestUserRepository_UpdatePasswordHashByEmail_UpdatesPassword(t *testing.T) {
	repo, db := setupUserRepositoryTest(t)

	db.Create(&models.User{
		Email:      "pwupdate@example.com",
		Password:   "oldhash",
		KeycloakID: "kc-pw",
		FirstName:  "Test",
		LastName:   "User",
	})

	err := repo.UpdatePasswordHashByEmail("pwupdate@example.com", "newhash")
	assert.NoError(t, err)

	var fromDB models.User
	err = db.First(&fromDB, "email = ?", "pwupdate@example.com").Error
	assert.NoError(t, err)
	assert.Equal(t, "newhash", fromDB.Password)
}

func TestUserRepository_UpdateProfilePhotoURLByKeycloakID_UpdatesField(t *testing.T) {
	repo, db := setupUserRepositoryTest(t)

	// seed user
	err := db.Create(&models.User{
		Email:      "photo@example.com",
		Password:   "hash",
		KeycloakID: "kc-photo-1",
		FirstName:  "Test",
		LastName:   "User",
	}).Error
	assert.NoError(t, err)

	// act
	wantURL := "http://localhost:9000/user-media/users/kc-photo-1/profile.jpg"
	err = repo.UpdateProfilePhotoURLByKeycloakID("kc-photo-1", wantURL)
	assert.NoError(t, err)

	// assert persisted
	var fromDB models.User
	err = db.First(&fromDB, "keycloak_id = ?", "kc-photo-1").Error
	assert.NoError(t, err)
	assert.Equal(t, wantURL, fromDB.ProfilePhotoURL)
}

func TestUserRepository_UpdateProfilePhotoURLByKeycloakID_UserNotFound_ReturnsError(t *testing.T) {
	repo, _ := setupUserRepositoryTest(t)
	err := repo.UpdateProfilePhotoURLByKeycloakID("missing", "http://x")
	assert.Error(t, err)
}
