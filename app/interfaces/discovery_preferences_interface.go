package interfaces

import "group1-userservice/app/models"

type DiscoveryPreferencesInput struct {
	RadiusKm int `json:"radius_km"`
}

type DiscoveryPreferencesService interface {
	GetForEmail(email string) (*models.DiscoveryPreferences, error)
	UpdateForEmail(email string, input DiscoveryPreferencesInput) (*models.DiscoveryPreferences, error)
}
