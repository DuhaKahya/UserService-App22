package service

import (
	"errors"

	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"
	"group1-userservice/app/repository"
)

type discoveryPreferencesService struct {
	repo *repository.DiscoveryPreferencesRepository
}

func NewDiscoveryPreferencesService(repo *repository.DiscoveryPreferencesRepository) interfaces.DiscoveryPreferencesService {
	return &discoveryPreferencesService{repo: repo}
}

func (s *discoveryPreferencesService) GetForEmail(email string) (*models.DiscoveryPreferences, error) {
	prefs, err := s.repo.GetByEmail(email)
	if err == nil {
		return prefs, nil
	}

	defaults := &models.DiscoveryPreferences{
		Email:    email,
		RadiusKm: 50,
	}
	return s.repo.Upsert(defaults)
}

func (s *discoveryPreferencesService) UpdateForEmail(email string, input interfaces.DiscoveryPreferencesInput) (*models.DiscoveryPreferences, error) {
	if input.RadiusKm < 1 || input.RadiusKm > 500 {
		return nil, errors.New("radius_km must be between 1 and 500")
	}

	prefs := &models.DiscoveryPreferences{
		Email:    email,
		RadiusKm: input.RadiusKm,
	}
	return s.repo.Upsert(prefs)
}
