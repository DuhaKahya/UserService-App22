package interfaces

import (
	"group1-userservice/app/models"

	"github.com/google/uuid"
)

type UserBadgeService interface {
	AwardBadge(userID uuid.UUID, badgeKey string) (bool, error)
	GetBadgesForUser(userID uuid.UUID) ([]models.UserBadge, error)
}
