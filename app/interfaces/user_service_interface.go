package interfaces

import (
	"group1-userservice/app/models"

	"github.com/google/uuid"
)

type UserService interface {
	Register(user *models.User) error
	GetByEmail(email string) (models.User, error)
	CheckPassword(hash string, raw string) bool
	GetByKeycloakID(sub string) (models.User, error)
	GetPublicInfoByFirstLast(first, last string) (*models.UserPublicInfo, error)
	UpdatePasswordByEmail(email string, newPlainPassword string) error
	UpdateByEmail(email string, input *models.UserUpdateInput) (models.User, error)
	UpdateProfilePhotoURLByKeycloakID(keycloakID, url string) error
	GetByID(id uuid.UUID) (models.User, error)
}
