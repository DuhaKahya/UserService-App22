package tests

import (
	"errors"
	"time"

	"group1-userservice/app/models"

	"github.com/google/uuid"
)

// fakeUserService avoids real Keycloak calls in tests
type fakeUserService struct{}

func (f *fakeUserService) Register(user *models.User) error {
	if user.Email == "" {
		return errors.New("Email is required")
	}
	user.Password = ""
	if user.KeycloakID == "" {
		user.KeycloakID = "kc-fake-register-success"
	}
	return nil
}

func (f *fakeUserService) GetByEmail(email string) (models.User, error) {
	return models.User{}, errors.New("not implemented")
}

func (f *fakeUserService) CheckPassword(hash string, raw string) bool {
	return true
}

func (f *fakeUserService) GetByKeycloakID(sub string) (models.User, error) {
	return models.User{}, errors.New("not implemented")
}

func (f *fakeUserService) GetPublicInfoByFirstLast(firstName, lastName string) (*models.UserPublicInfo, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeUserService) UpdatePasswordByEmail(email string, newPlainPassword string) error {
	return nil
}

// fakeResetRepo is a minimal in-memory reset repository for service tests
type fakeResetRepo struct {
	row *models.PasswordResetToken
}

func (f *fakeResetRepo) Create(email, hash string, exp time.Time) error {
	return nil
}

func (f *fakeResetRepo) FindValidByTokenHash(hash string) (*models.PasswordResetToken, error) {
	if f.row == nil {
		return nil, errors.New("not found")
	}
	return f.row, nil
}

func (f *fakeResetRepo) MarkUsed(id uint) error {
	return nil
}

func (f *fakeUserService) UpdateByEmail(email string, input *models.UserUpdateInput) (models.User, error) {
	u := models.User{Email: email}

	if input != nil {
		if input.FirstName != "" {
			u.FirstName = input.FirstName
		}
		if input.LastName != "" {
			u.LastName = input.LastName
		}
		if input.PhoneNumberVisible != nil {
			u.PhoneNumberVisible = *input.PhoneNumberVisible
		}
	}

	return u, nil
}

func (f *fakeUserService) UpdateProfilePhotoURLByKeycloakID(keycloakID, url string) error {
	return nil
}

func (f *fakeUserService) GetByID(id uuid.UUID) (models.User, error) {
	return models.User{}, nil
}

// --- Fake BadgeService for controller tests ---
type fakeBadgeService struct{}

func (f *fakeBadgeService) AwardBadge(userID uuid.UUID, badgeKey string) (bool, error) {
	return false, nil
}

func (f *fakeBadgeService) GetBadgesForUser(userID uuid.UUID) ([]models.UserBadge, error) {
	return []models.UserBadge{}, nil
}
