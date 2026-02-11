package tests

import (
	"testing"

	"group1-userservice/app/models"
	"group1-userservice/app/service"
)

func TestIsProfileComplete(t *testing.T) {
	u := models.User{
		FirstName:       "John",
		LastName:        "Doe",
		Email:           "john@example.com",
		PhoneNumber:     "123",
		Country:         "NL",
		JobFunction:     "Developer",
		Sector:          "IT",
		Biography:       "Hello, I am John",
		ProfilePhotoURL: "http://example.com/p.jpg",
	}

	if !service.IsProfileComplete(u) {
		t.Error("expected profile to be complete")
	}

	// clear a required field
	u.PhoneNumber = ""

	if service.IsProfileComplete(u) {
		t.Error("profile should NOT be complete")
	}
}
