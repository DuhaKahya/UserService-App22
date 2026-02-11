package models

type Interest struct {
	ID  uint   `json:"id" gorm:"primaryKey"`
	Key string `json:"key" gorm:"uniqueIndex;not null"`
}
