package service

import (
	"errors"
	"regexp"

	"group1-userservice/app/interfaces"
	"group1-userservice/app/keycloak"
	"group1-userservice/app/models"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
)

type userService struct {
	repo interfaces.UserRepository
}

func NewUserService(repo interfaces.UserRepository) interfaces.UserService {
	return &userService{repo}
}

func (s *userService) Register(user *models.User) error {
	// Basic field validation
	if user.Email == "" {
		return errors.New("Email is required")
	}
	if user.FirstName == "" {
		return errors.New("First name is required")
	}
	if user.LastName == "" {
		return errors.New("Last name is required")
	}

	plainPassword := user.Password

	if s.repo.ExistsByEmail(user.Email) {
		return errors.New("email already exists")
	}

	if len(user.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	matched, _ := regexp.MatchString(`[0-9]`, user.Password)
	if !matched {
		return errors.New("password must contain at least one number")
	}

	kcID, err := keycloak.CreateUserInKeycloakWithPassword(*user, plainPassword)
	if err != nil {
		if errors.Is(err, keycloak.ErrEmailAlreadyExists) {
			return errors.New("email already exists")
		}
		return err
	}

	hashed, err := HashPassword(user.Password)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = hashed
	user.KeycloakID = kcID

	if err := s.repo.Create(user); err != nil {
		return err
	}

	return nil
}

func (s *userService) GetByEmail(email string) (models.User, error) {
	return s.repo.FindByEmail(email)
}

func (s *userService) CheckPassword(hash string, raw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw)) == nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (s *userService) GetByKeycloakID(sub string) (models.User, error) {
	return s.repo.FindByKeycloakID(sub)
}

func (s *userService) GetPublicInfoByFirstLast(first, last string) (*models.UserPublicInfo, error) {
	user, err := s.repo.FindByFirstLastInsensitive(first, last)
	if err != nil {
		return nil, err
	}

	return &models.UserPublicInfo{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}, nil
}

func (s *userService) UpdatePasswordByEmail(email string, newPlainPassword string) error {
	if len(newPlainPassword) < 6 {
		return errors.New("password must be at least 6 characters long")
	}
	matched, _ := regexp.MatchString(`[0-9]`, newPlainPassword)
	if !matched {
		return errors.New("password must contain at least one number")
	}

	hashed, err := HashPassword(newPlainPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	return s.repo.UpdatePasswordHashByEmail(email, hashed)
}

func (s *userService) UpdateByEmail(email string, input *models.UserUpdateInput) (models.User, error) {
	fields := map[string]any{}

	if input.FirstName != "" {
		fields["first_name"] = input.FirstName
	}
	if input.LastName != "" {
		fields["last_name"] = input.LastName
	}
	if input.PhoneNumber != "" {
		fields["phone_number"] = input.PhoneNumber
	}
	if input.PhoneNumberVisible != nil {
		fields["phone_number_visible"] = *input.PhoneNumberVisible
	}

	if input.Country != "" {
		fields["country"] = input.Country
	}
	if input.JobFunction != "" {
		fields["job_function"] = input.JobFunction
	}
	if input.Sector != "" {
		fields["sector"] = input.Sector
	}
	if input.Biography != "" {
		fields["biography"] = input.Biography
	}
	if input.ProfilePhotoURL != "" {
		fields["profile_photo_url"] = input.ProfilePhotoURL
	}

	return s.repo.UpdateFieldsByEmail(email, fields)
}

func (s *userService) UpdateProfilePhotoURLByKeycloakID(keycloakID, url string) error {
	return s.repo.UpdateProfilePhotoURLByKeycloakID(keycloakID, url)
}

func IsProfileComplete(u models.User) bool {
	return u.FirstName != "" &&
		u.LastName != "" &&
		u.PhoneNumber != "" &&
		u.Country != "" &&
		u.JobFunction != "" &&
		u.Sector != "" &&
		u.Biography != "" &&
		u.ProfilePhotoURL != ""
}

func (s *userService) GetByID(id uuid.UUID) (models.User, error) {
	return s.repo.GetByID(id)
}
