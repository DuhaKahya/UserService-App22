package models

type DiscoveryPreferences struct {
	Email    string `json:"email" gorm:"type:varchar(255);not null;primaryKey"`
	RadiusKm int    `json:"radius_km" gorm:"not null;default:50"`
}
