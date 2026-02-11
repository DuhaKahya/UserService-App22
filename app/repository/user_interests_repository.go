package repository

import (
	"group1-userservice/app/config"
	"group1-userservice/app/models"

	"gorm.io/gorm"
)

type UserInterestsRepository struct {
	db *gorm.DB
}

func NewUserInterestsRepository() *UserInterestsRepository {
	return &UserInterestsRepository{db: config.DB}
}

func (r *UserInterestsRepository) AnyInterestsExist() (bool, error) {
	var count int64
	if err := r.db.Model(&models.Interest{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *UserInterestsRepository) CreateInterests(keys []string) error {
	for _, key := range keys {
		interest := models.Interest{Key: key}
		if err := r.db.FirstOrCreate(&interest, models.Interest{Key: key}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *UserInterestsRepository) ListAllInterests() ([]models.Interest, error) {
	var interests []models.Interest
	if err := r.db.Order("id asc").Find(&interests).Error; err != nil {
		return nil, err
	}
	return interests, nil
}

func (r *UserInterestsRepository) GetUserInterests(email string) ([]models.UserInterest, error) {
	var rows []models.UserInterest
	if err := r.db.
		Where("user_email = ?", email).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *UserInterestsRepository) UpsertUserInterest(row *models.UserInterest) error {
	return r.db.Save(row).Error
}
