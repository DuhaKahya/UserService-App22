package service

import (
	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"

	"github.com/google/uuid"
)

// Define badge keys as constants so they can reused safely.
const BadgeKeyProfilePhotoUploaded = "profile_photo_uploaded"
const BadgeKeyProfileComplete = "profile_complete"
const BadgeKeyWatch50 = "watch_50_videos"
const BadgeKeyShare10 = "share_10_videos"
const BadgeKeyLike25 = "like_25_videos"

type userBadgeService struct {
	repo interfaces.UserBadgeRepository
}

func NewUserBadgeService(repo interfaces.UserBadgeRepository) interfaces.UserBadgeService {
	return &userBadgeService{repo: repo}
}

func (s *userBadgeService) AwardBadge(userID uuid.UUID, badgeKey string) (bool, error) {
	_, created, err := s.repo.CreateIfNotExists(userID, badgeKey)
	return created, err
}

func (s *userBadgeService) GetBadgesForUser(userID uuid.UUID) ([]models.UserBadge, error) {
	return s.repo.FindByUserID(userID)
}
