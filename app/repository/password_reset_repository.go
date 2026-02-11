package repository

import (
	"time"

	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"

	"gorm.io/gorm"
)

type PasswordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) interfaces.PasswordResetRepository {
	return &PasswordResetRepository{db: db}
}

func (r *PasswordResetRepository) Create(email string, tokenHash string, expiresAt time.Time) error {
	row := &models.PasswordResetToken{
		Email:     email,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		Used:      false,
	}
	return r.db.Create(row).Error
}

func (r *PasswordResetRepository) FindValidByTokenHash(tokenHash string) (*models.PasswordResetToken, error) {
	var row models.PasswordResetToken
	err := r.db.
		Where("token_hash = ? AND used = false AND expires_at > NOW()", tokenHash).
		First(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *PasswordResetRepository) MarkUsed(id uint) error {
	return r.db.Model(&models.PasswordResetToken{}).
		Where("id = ?", id).
		Update("used", true).Error
}
