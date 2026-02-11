package models

import "github.com/google/uuid"

type User struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	KeycloakID         string    `json:"keycloak_id" gorm:"uniqueIndex"`
	Email              string    `json:"email"`
	Password           string    `json:"password"`
	FirstName          string    `json:"first_name"`
	LastName           string    `json:"last_name"`
	PhoneNumber        string    `json:"phone_number"`
	PhoneNumberVisible bool      `json:"phone_number_visible" gorm:"default:false"`
	Country            string    `json:"country"`
	JobFunction        string    `json:"job_function"`
	Sector             string    `json:"sector"`
	Biography          string    `json:"biography"`
	IsBlocked          bool      `json:"is_blocked"`
	ProfilePhotoURL    string    `json:"profile_photo_url" gorm:"default:''"`
}

type UserUpdateInput struct {
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	PhoneNumber        string `json:"phone_number"`
	PhoneNumberVisible *bool  `json:"phone_number_visible"`
	Country            string `json:"country"`
	JobFunction        string `json:"job_function"`
	Sector             string `json:"sector"`
	Biography          string `json:"biography"`
	ProfilePhotoURL    string `json:"profile_photo_url"`
}
