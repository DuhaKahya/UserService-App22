package config

import (
	"log"

	"group1-userservice/app/models"
)

func SeedBadges() {
	badges := []models.Badge{
		{
			Key:         "profile_photo_uploaded",
			Name:        "Profile picture",
			Description: "User uploaded a profile photo",
		},
		{
			Key:         "profile_complete",
			Name:        "Complete Profile",
			Description: "User completed all required profile information",
		},
		{
			Key:         "watch_50_videos",
			Name:        "Watched 50 Videos",
			Description: "User has watched 50 videos",
		},
		{
			Key:         "share_10_videos",
			Name:        "Shared 10 Videos",
			Description: "User has shared 10 videos",
		},
		{
			Key:         "like_25_videos",
			Name:        "Liked 25 Videos",
			Description: "User has liked 25 videos",
		},
	}

	for _, b := range badges {
		var existing models.Badge
		err := DB.Where("key = ?", b.Key).First(&existing).Error
		if err == nil {
			// already exists, skip
			continue
		}

		if err := DB.Create(&b).Error; err != nil {
			log.Fatalf("failed to seed badge %s: %v", b.Key, err)
		}
	}
}
