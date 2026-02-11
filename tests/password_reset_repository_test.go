package tests

import (
	"testing"
	"time"

	"group1-userservice/app/config"
	"group1-userservice/app/interfaces"
	"group1-userservice/app/models"
	"group1-userservice/app/repository"

	"github.com/stretchr/testify/assert"
)

// setupPasswordResetRepo creates a test DB and migrates the reset token table
func setupPasswordResetRepo(t *testing.T) interfaces.PasswordResetRepository {
	t.Helper()

	db := openTestDB(t)
	config.DB = db

	if err := config.DB.AutoMigrate(&models.PasswordResetToken{}); err != nil {
		t.Fatalf("failed to migrate tables: %v", err)
	}

	return repository.NewPasswordResetRepository(config.DB)
}

func TestPasswordResetRepo_CreateAndFindValid(t *testing.T) {
	repo := setupPasswordResetRepo(t)

	err := repo.Create(
		"test@example.com",
		"hash-123",
		time.Now().Add(30*time.Minute),
	)
	assert.NoError(t, err)

	row, err := repo.FindValidByTokenHash("hash-123")
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", row.Email)
	assert.False(t, row.Used)
}

func TestPasswordResetRepo_ExpiredToken_ReturnsError(t *testing.T) {
	repo := setupPasswordResetRepo(t)

	_ = repo.Create(
		"test@example.com",
		"expired-hash",
		time.Now().Add(-1*time.Minute),
	)

	row, err := repo.FindValidByTokenHash("expired-hash")
	assert.Error(t, err)
	assert.Nil(t, row)
}

func TestPasswordResetRepo_MarkUsed_InvalidatesToken(t *testing.T) {
	repo := setupPasswordResetRepo(t)

	_ = repo.Create(
		"test@example.com",
		"used-hash",
		time.Now().Add(30*time.Minute),
	)

	row, _ := repo.FindValidByTokenHash("used-hash")
	assert.NoError(t, repo.MarkUsed(row.ID))

	row2, err := repo.FindValidByTokenHash("used-hash")
	assert.Error(t, err)
	assert.Nil(t, row2)
}
