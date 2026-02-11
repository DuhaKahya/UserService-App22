package models

import (
	"time"

	"github.com/google/uuid"
)

// Badge is a master table with all possible badges.
type Badge struct {
	ID          uint   `json:"-" gorm:"primaryKey"`
	Key         string `json:"key" gorm:"uniqueIndex;size:100"` // e.g. "profile_photo_uploaded"
	Name        string `json:"name" gorm:"size:200"`            // e.g. "Profile picture"
	Description string `json:"description" gorm:"size:500"`
}

// UserBadge links a user to a specific badge.
type UserBadge struct {
	ID       uint      `json:"id" gorm:"primaryKey"`
	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	BadgeKey string    `json:"badge_key" gorm:"size:100;index"`
	EarnedAt time.Time `json:"earned_at" gorm:"autoCreateTime"`

	// Join UserBadge.BadgeKey -> Badge.Key
	Badge Badge `json:"badge" gorm:"foreignKey:BadgeKey;references:Key"`
}
