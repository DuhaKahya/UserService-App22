package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"group1-userservice/app/interfaces"
	"group1-userservice/app/keycloak"
)

type passwordResetService struct {
	resetRepo interfaces.PasswordResetRepository
	userSvc   interfaces.UserService
}

func NewPasswordResetService(r interfaces.PasswordResetRepository, userSvc interfaces.UserService) interfaces.PasswordResetService {
	return &passwordResetService{resetRepo: r, userSvc: userSvc}
}

func (s *passwordResetService) RequestReset(email string) (string, error) {
	_, _ = s.userSvc.GetByEmail(email)

	rawToken, err := generateToken(32)
	if err != nil {
		return "", err
	}

	tokenHash := hashToken(rawToken)
	expires := time.Now().Add(30 * time.Minute)

	if err := s.resetRepo.Create(email, tokenHash, expires); err != nil {
		return "", err
	}

	return rawToken, nil
}

func (s *passwordResetService) ResetPassword(rawToken string, newPassword string) error {
	if rawToken == "" {
		return errors.New("token is required")
	}

	tokenHash := hashToken(rawToken)
	row, err := s.resetRepo.FindValidByTokenHash(tokenHash)
	if err != nil {
		return errors.New("invalid or expired token")
	}

	// Keycloak reset
	if err := keycloak.ResetPasswordByEmail(row.Email, newPassword); err != nil {
		return err
	}

	// Postgres update (bcrypt)
	if err := s.userSvc.UpdatePasswordByEmail(row.Email, newPassword); err != nil {
		return err
	}

	// mark token used
	if err := s.resetRepo.MarkUsed(row.ID); err != nil {
		return err
	}

	return nil
}

func generateToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
