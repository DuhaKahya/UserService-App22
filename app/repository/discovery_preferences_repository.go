package repository

import (
	"group1-userservice/app/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DiscoveryPreferencesRepository struct{ db *gorm.DB }

func NewDiscoveryPreferencesRepository(db *gorm.DB) *DiscoveryPreferencesRepository {
	return &DiscoveryPreferencesRepository{db: db}
}

func (r *DiscoveryPreferencesRepository) GetByEmail(email string) (*models.DiscoveryPreferences, error) {
	var p models.DiscoveryPreferences
	if err := r.db.Where("email = ?", email).First(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *DiscoveryPreferencesRepository) Upsert(p *models.DiscoveryPreferences) (*models.DiscoveryPreferences, error) {
	err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "email"}},
		DoUpdates: clause.AssignmentColumns([]string{"radius_km"}),
	}).Create(p).Error

	if err != nil {
		return nil, err
	}
	return p, nil
}
