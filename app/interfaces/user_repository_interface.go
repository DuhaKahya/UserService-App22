package interfaces

import (
	"group1-userservice/app/models"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (models.User, error)
	FindAll() ([]models.User, error)
	ExistsByEmail(email string) bool
	FindByKeycloakID(sub string) (models.User, error)
	FindPublicInfoByFirstLast(firstName, lastName string) (*models.UserPublicInfo, error)
	FindByFirstLastInsensitive(first, last string) (models.User, error)
	UpdatePasswordHashByEmail(email string, passwordHash string) error
	UpdateFieldsByEmail(email string, fields map[string]any) (models.User, error)
	UpdateProfilePhotoURLByKeycloakID(keycloakID, url string) error
	GetByID(id uuid.UUID) (models.User, error)
}
