package repository

import (
	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"

	"gorm.io/gorm"

	"github.com/google/uuid"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) interfaces.UserRepository {
	return &userRepository{db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByEmail(email string) (models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return user, err
}

func (r *userRepository) FindAll() ([]models.User, error) {
	var users []models.User
	err := r.db.Find(&users).Error
	return users, err
}

func (r *userRepository) ExistsByEmail(email string) bool {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	return err == nil
}

func (r *userRepository) FindByKeycloakID(sub string) (models.User, error) {
	var user models.User
	err := r.db.Where("keycloak_id = ?", sub).First(&user).Error
	return user, err
}

func (r *userRepository) FindPublicInfoByFirstLast(firstName, lastName string) (*models.UserPublicInfo, error) {
	var out models.UserPublicInfo

	err := r.db.
		Table("users").
		Select("first_name, last_name, email").
		Where("first_name = ? AND last_name = ?", firstName, lastName).
		First(&out).Error

	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *userRepository) FindByFirstLastInsensitive(first, last string) (models.User, error) {
	var user models.User

	err := r.db.
		Where("LOWER(first_name) = LOWER(?) AND LOWER(last_name) = LOWER(?)", first, last).
		First(&user).
		Error

	return user, err
}

func (r *userRepository) UpdatePasswordHashByEmail(email string, passwordHash string) error {
	return r.db.Model(&models.User{}).
		Where("email = ?", email).
		Update("password", passwordHash).Error
}

func (r *userRepository) UpdateFieldsByEmail(email string, fields map[string]any) (models.User, error) {
	if err := r.db.Model(&models.User{}).
		Where("email = ?", email).
		Updates(fields).Error; err != nil {
		return models.User{}, err
	}

	return r.FindByEmail(email)
}

func (r *userRepository) UpdateProfilePhotoURLByKeycloakID(keycloakID, url string) error {
	tx := r.db.Model(&models.User{}).
		Where("keycloak_id = ?", keycloakID).
		Update("profile_photo_url", url)

	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *userRepository) GetByID(id uuid.UUID) (models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	return user, err
}
