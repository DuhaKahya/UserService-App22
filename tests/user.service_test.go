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

// Setup helper
func setupUserServiceTest(t *testing.T) (interfaces.UserService, *gorm.DB) {
	t.Helper()

	db := openTestDB(t)
	config.DB = db

	if err := config.DB.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)

	return userService, db
}

// Register() validation tests

func TestRegister_EmailRequired(t *testing.T) {
	userService, _ := setupUserServiceTest(t)

	user := &models.User{
		Email:     "",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "abc123",
	}

	err := userService.Register(user)

	assert.Error(t, err)
	assert.Equal(t, "Email is required", err.Error())
}

func TestRegister_FirstNameRequired(t *testing.T) {
	userService, _ := setupUserServiceTest(t)

	user := &models.User{
		Email:     "test@example.com",
		FirstName: "",
		LastName:  "Doe",
		Password:  "abc123",
	}

	err := userService.Register(user)

	assert.Error(t, err)
	assert.Equal(t, "First name is required", err.Error())
}

func TestRegister_LastNameRequired(t *testing.T) {
	userService, _ := setupUserServiceTest(t)

	user := &models.User{
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "",
		Password:  "abc123",
	}

	err := userService.Register(user)

	assert.Error(t, err)
	assert.Equal(t, "Last name is required", err.Error())
}

func TestRegister_PasswordTooShort(t *testing.T) {
	userService, _ := setupUserServiceTest(t)

	user := &models.User{
		Email:     "short@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "a1b2",
	}

	err := userService.Register(user)

	assert.Error(t, err)
	assert.Equal(t, "password must be at least 6 characters long", err.Error())
}

func TestRegister_PasswordMissingNumber(t *testing.T) {
	userService, _ := setupUserServiceTest(t)

	user := &models.User{
		Email:     "pass@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "abcdef",
	}

	err := userService.Register(user)

	assert.Error(t, err)
	assert.Equal(t, "password must contain at least one number", err.Error())
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	userService, db := setupUserServiceTest(t)

	existing := &models.User{
		Email:      "exists@example.com",
		FirstName:  "Existing",
		LastName:   "User",
		Password:   "hash",
		KeycloakID: "kc-user-service-exists",
	}

	if err := db.Create(existing).Error; err != nil {
		t.Fatalf("failed to create existing user: %v", err)
	}

	user := &models.User{
		Email:     "exists@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Password:  "abc123",
	}

	err := userService.Register(user)

	assert.Error(t, err)
	assert.Equal(t, "email already exists", err.Error())
}

// GetByEmail / GetByKeycloakID

func TestGetByEmail_Found(t *testing.T) {
	userService, db := setupUserServiceTest(t)

	db.Create(&models.User{
		Email:      "getemail@example.com",
		Password:   "hash",
		KeycloakID: "kc-1",
		FirstName:  "Test",
		LastName:   "User",
	})

	u, err := userService.GetByEmail("getemail@example.com")

	assert.NoError(t, err)
	assert.Equal(t, "getemail@example.com", u.Email)
}

func TestGetByEmail_NotFound(t *testing.T) {
	userService, _ := setupUserServiceTest(t)

	_, err := userService.GetByEmail("unknown@example.com")
	assert.Error(t, err)
}

func TestGetByKeycloakID_Found(t *testing.T) {
	userService, db := setupUserServiceTest(t)

	db.Create(&models.User{
		Email:      "kcsearch@example.com",
		Password:   "hash",
		KeycloakID: "kc-sub-123",
		FirstName:  "Test",
		LastName:   "User",
	})

	u, err := userService.GetByKeycloakID("kc-sub-123")

	assert.NoError(t, err)
	assert.Equal(t, "kc-sub-123", u.KeycloakID)
}

func TestGetByKeycloakID_NotFound(t *testing.T) {
	userService, _ := setupUserServiceTest(t)

	_, err := userService.GetByKeycloakID("missing")
	assert.Error(t, err)
}

// HashPassword / CheckPassword

func TestHashPassword_ProducesValidHash(t *testing.T) {
	raw := "mysecret123"

	hashed, err := service.HashPassword(raw)

	assert.NoError(t, err)
	assert.NotEqual(t, raw, hashed)
	assert.NotEmpty(t, hashed)
}

func TestCheckPassword_CorrectPassword(t *testing.T) {
	userService, _ := setupUserServiceTest(t)

	raw := "mysecret123"
	hashed, _ := service.HashPassword(raw)

	match := userService.CheckPassword(hashed, raw)
	assert.True(t, match)
}

func TestCheckPassword_InvalidPassword(t *testing.T) {
	userService, _ := setupUserServiceTest(t)

	hashed, _ := service.HashPassword("correct")

	match := userService.CheckPassword(hashed, "wrong")
	assert.False(t, match)
}

func TestGetPublicInfoByFirstLast_CaseInsensitive(t *testing.T) {
	userService, db := setupUserServiceTest(t)

	db.Create(&models.User{
		Email:      "public@example.com",
		Password:   "hash",
		KeycloakID: "kc-public",
		FirstName:  "John",
		LastName:   "Doe",
	})

	info, err := userService.GetPublicInfoByFirstLast("john", "doe")
	assert.NoError(t, err)
	assert.Equal(t, "public@example.com", info.Email)
	assert.Equal(t, "John", info.FirstName)
	assert.Equal(t, "Doe", info.LastName)
}

func TestUpdateByEmail_UpdatesAllowedFields(t *testing.T) {
	userService, db := setupUserServiceTest(t)

	err := db.Create(&models.User{
		Email:      "update@example.com",
		Password:   "hash",
		KeycloakID: "kc-update",
		FirstName:  "Old",
		LastName:   "Name",
		Country:    "NL",
	}).Error
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	input := &models.UserUpdateInput{
		FirstName: "NewFirst",
		Biography: "New bio",
	}

	updated, err := userService.UpdateByEmail("update@example.com", input)
	assert.NoError(t, err)

	// Check returned
	assert.Equal(t, "update@example.com", updated.Email)
	assert.Equal(t, "NewFirst", updated.FirstName)
	assert.Equal(t, "Name", updated.LastName)
	assert.Equal(t, "New bio", updated.Biography)

	// Check persisted
	u2, err := userService.GetByEmail("update@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "NewFirst", u2.FirstName)
	assert.Equal(t, "Name", u2.LastName)
	assert.Equal(t, "New bio", u2.Biography)
	assert.Equal(t, "NL", u2.Country)
}

func TestUpdateByEmail_UpdatesPhoneNumberVisibleTrue(t *testing.T) {
	userService, db := setupUserServiceTest(t)

	err := db.Create(&models.User{
		Email:      "updatebool@example.com",
		Password:   "hash",
		KeycloakID: "kc-updatebool",
		FirstName:  "Old",
		LastName:   "Name",
	}).Error
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	v := true
	input := &models.UserUpdateInput{
		PhoneNumberVisible: &v,
	}

	updated, err := userService.UpdateByEmail("updatebool@example.com", input)
	assert.NoError(t, err)

	// returned struct
	assert.True(t, updated.PhoneNumberVisible)

	// persisted
	u2, err := userService.GetByEmail("updatebool@example.com")
	assert.NoError(t, err)
	assert.True(t, u2.PhoneNumberVisible)
}

func TestUpdateByEmail_DoesNotChangeRestrictedFields(t *testing.T) {
	userService, db := setupUserServiceTest(t)

	err := db.Create(&models.User{
		Email:      "restricted@example.com",
		Password:   "original-hash",
		KeycloakID: "kc-original",
		FirstName:  "A",
		LastName:   "B",
	}).Error
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	input := &models.UserUpdateInput{
		FirstName: "Changed",
	}

	updated, err := userService.UpdateByEmail("restricted@example.com", input)
	assert.NoError(t, err)

	assert.Equal(t, "restricted@example.com", updated.Email)
	assert.Equal(t, "Changed", updated.FirstName)

	// Persist check
	u2, err := userService.GetByEmail("restricted@example.com")
	assert.NoError(t, err)

	assert.Equal(t, "restricted@example.com", u2.Email)
	assert.Equal(t, "kc-original", u2.KeycloakID)
	assert.Equal(t, "original-hash", u2.Password)
	assert.Equal(t, "Changed", u2.FirstName)
}

func TestUpdateByEmail_UserNotFound(t *testing.T) {
	userService, _ := setupUserServiceTest(t)

	_, err := userService.UpdateByEmail("missing@example.com", &models.UserUpdateInput{FirstName: "X"})
	assert.Error(t, err)
}

func TestUserService_UpdateProfilePhotoURLByKeycloakID_UpdatesDB(t *testing.T) {
	userService, db := setupUserServiceTest(t)

	// seed user
	err := db.Create(&models.User{
		Email:      "svcphoto@example.com",
		Password:   "hash",
		KeycloakID: "kc-svc-photo-1",
		FirstName:  "Test",
		LastName:   "User",
	}).Error
	assert.NoError(t, err)

	// act
	wantURL := "http://localhost:9000/user-media/users/kc-svc-photo-1/profile.jpg"
	err = userService.UpdateProfilePhotoURLByKeycloakID("kc-svc-photo-1", wantURL)
	assert.NoError(t, err)

	// assert via service read (of direct DB)
	u, err := userService.GetByKeycloakID("kc-svc-photo-1")
	assert.NoError(t, err)
	assert.Equal(t, wantURL, u.ProfilePhotoURL)
}
