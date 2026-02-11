package models

type NotificationSettings struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	UserEmail string `json:"email" gorm:"uniqueIndex"`

	LikeEmail     bool   `json:"like_email"`
	LikePush      bool   `json:"like_push"`
	FavoriteEmail bool   `json:"favorite_email"`
	FavoritePush  bool   `json:"favorite_push"`
	ChatEmail     bool   `json:"chat_email"`
	ChatPush      bool   `json:"chat_push"`
	SystemEmail   bool   `json:"system_email"`
	SystemPush    bool   `json:"system_push"`
	ExpoPushToken string `json:"expo_push_token" gorm:"size:255"`
}
