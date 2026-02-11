package tests

import (
	"testing"

	"group1-userservice/app/interfaces"
	"group1-userservice/app/keycloak"
	"group1-userservice/app/models"
	"group1-userservice/app/service"

	"github.com/stretchr/testify/assert"
)

func TestPasswordResetService_ResetPassword_Success(t *testing.T) {
	orig := keycloak.ResetPasswordByEmail
	defer func() { keycloak.ResetPasswordByEmail = orig }()

	keycloak.ResetPasswordByEmail = func(email, pw string) error {
		return nil
	}

	repo := &fakeResetRepo{
		row: &models.PasswordResetToken{
			ID:    1,
			Email: "test@example.com",
		},
	}

	var _ interfaces.PasswordResetRepository = repo

	svc := service.NewPasswordResetService(repo, &fakeUserService{})

	err := svc.ResetPassword("raw-token", "Welkom1234")
	assert.NoError(t, err)
}
