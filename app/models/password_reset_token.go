package models

import "time"

type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"index;not null"`
	TokenHash string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"index;not null"`
	Used      bool      `gorm:"not null;default:false"`
	CreatedAt time.Time
}
