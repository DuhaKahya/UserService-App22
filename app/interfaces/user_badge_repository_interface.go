package interfaces

import (
	"group1-userservice/app/models"

	"github.com/google/uuid"
)

type UserBadgeRepository interface {
	CreateIfNotExists(userID uuid.UUID, badgeKey string) (models.UserBadge, bool, error)
	FindByUserID(userID uuid.UUID) ([]models.UserBadge, error)
}
