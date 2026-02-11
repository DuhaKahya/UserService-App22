package tests

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"group1-userservice/app/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Converts int → string for building URLs
func intToString(v int) string {
	return strconv.Itoa(v)
}

// Hashes a plaintext password for use in tests
func hashPassword(t *testing.T, plain string) string {
	t.Helper()

	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("error hashing password: %v", err)
	}
	return string(hashed)
}

// Inserts a user with a hashed password into the test database
func createTestUser(t *testing.T, db *gorm.DB, email, plainPassword string) {
	t.Helper()

	user := models.User{
		KeycloakID: "test-kc-" + uuid.NewString(),

		Email:    email,
		Password: hashPassword(t, plainPassword),

		FirstName: "Test",
		LastName:  "User",
	}

	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
}

func truncateIfExists(db *gorm.DB, table string) {
	// Truncate only if table exists to avoid "relation does not exist" noise
	db.Exec(fmt.Sprintf(`
DO $$
BEGIN
	IF to_regclass('public.%s') IS NOT NULL THEN
		EXECUTE 'TRUNCATE TABLE %s RESTART IDENTITY CASCADE';
	END IF;
END $$;`, table, table))
}

// Open a Postgres test database using env variables or sensible defaults
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	host := getEnvOrDefault("DB_HOST", "localhost")
	user := getEnvOrDefault("DB_USER", "admin")
	password := getEnvOrDefault("DB_PASSWORD", "admin")
	dbname := getEnvOrDefault("DB_NAME", "userservice")
	port := getEnvOrDefault("DB_PORT", "5431")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test postgres db: %v", err)
	}

	// Wipe all tables before each test (only if they exist)
	truncateIfExists(db, "users")
	truncateIfExists(db, "notification_settings")
	truncateIfExists(db, "user_interests")
	truncateIfExists(db, "interests")
	truncateIfExists(db, "discovery_preferences")
	truncateIfExists(db, "password_reset_tokens")

	return db
}

func getEnvOrDefault(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getInterestIDByKey(t *testing.T, db *gorm.DB, key string) uint {
	t.Helper()

	var interest models.Interest
	if err := db.Where("key = ?", key).First(&interest).Error; err != nil {
		t.Fatalf("failed to find interest %s: %v", key, err)
	}

	return interest.ID
}

func seedInterests(t *testing.T, db *gorm.DB) []models.Interest {
	t.Helper()

	keys := []string{
		"Gezondheidszorg en Welzijn",
		"Handel en Dienstverlening",
		"ICT",
		"Justitie, Veiligheid en Openbaar Bestuur",
		"Milieu en Agrarische Sector",
		"Media en Communicatie",
		"Onderwijs, Cultuur en Wetenschap",
		"Techniek, Productie en Bouw",
		"Toerisme, Recreatie en Horeca",
		"Transport en Logistiek",
		"Behoefte aan Investering",
		"Interesse om te Investeren",
	}

	out := make([]models.Interest, 0, len(keys))

	for _, k := range keys {
		i := models.Interest{Key: k}
		if err := db.FirstOrCreate(&i, models.Interest{Key: k}).Error; err != nil {
			t.Fatalf("failed to seed interest %s: %v", k, err)
		}
		out = append(out, i)
	}
	return out
}

func wipeNotificationSettings(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.Exec("DELETE FROM notification_settings").Error; err != nil {
		t.Fatalf("failed to wipe notification_settings: %v", err)
	}
}
