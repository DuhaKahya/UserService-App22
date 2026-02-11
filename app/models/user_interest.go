package models

type UserInterest struct {
	UserEmail  string `gorm:"primaryKey;not null;index"`
	InterestID uint   `gorm:"primaryKey;not null"`
	Value      bool   `json:"value" gorm:"not null;default:false"`

	Interest Interest `json:"interest" gorm:"foreignKey:InterestID;references:ID"`
}
