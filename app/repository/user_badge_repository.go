package repository

import (
	"errors"
	"time"

	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userBadgeRepository struct {
	db *gorm.DB
}

func NewUserBadgeRepository(db *gorm.DB) interfaces.UserBadgeRepository {
	return &userBadgeRepository{db: db}
}

func (r *userBadgeRepository) CreateIfNotExists(userID uuid.UUID, badgeKey string) (models.UserBadge, bool, error) {
	var existing models.UserBadge

	err := r.db.Where("user_id = ? AND badge_key = ?", userID, badgeKey).First(&existing).Error
	if err == nil {
		return existing, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.UserBadge{}, false, err
	}

	newBadge := models.UserBadge{
		UserID:   userID,
		BadgeKey: badgeKey,
		EarnedAt: time.Now(),
	}

	if err := r.db.Create(&newBadge).Error; err != nil {
		return models.UserBadge{}, false, err
	}

	return newBadge, true, nil
}

func (r *userBadgeRepository) FindByUserID(userID uuid.UUID) ([]models.UserBadge, error) {
	var badges []models.UserBadge

	err := r.db.
		Preload("Badge").
		Where("user_id = ?", userID).
		Order("earned_at ASC").
		Find(&badges).Error

	return badges, err
}
