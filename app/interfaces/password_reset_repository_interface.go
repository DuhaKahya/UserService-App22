package interfaces

import (
	"group1-userservice/app/models"
	"time"
)

type PasswordResetRepository interface {
	Create(email string, tokenHash string, expiresAt time.Time) error
	FindValidByTokenHash(tokenHash string) (*models.PasswordResetToken, error)
	MarkUsed(id uint) error
}
