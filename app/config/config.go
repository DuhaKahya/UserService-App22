package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"group1-userservice/app/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// ConnectDatabase initializes the PostgreSQL connection and runs migrations
func ConnectDatabase() {
	host := getEnv("DB_HOST", "localhost")
	user := getEnv("DB_USER", "admin")
	password := getEnv("DB_PASSWORD", "admin")
	dbname := getEnv("DB_NAME", "userservice")
	port := getEnv("DB_PORT", "5431")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	var database *gorm.DB
	var err error
	maxAttempts := 10
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Database connection failed (attempt %d/%d): %v. Retrying in 2s...", attempt, maxAttempts, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Database connection failed after %d attempts: %v", maxAttempts, err)
	}

	DB = database
	log.Println("Connected to PostgreSQL")

	err = DB.AutoMigrate(
		&models.User{},
		&models.NotificationSettings{},
		&models.Interest{},
		&models.UserInterest{},
		&models.DiscoveryPreferences{},
		&models.PasswordResetToken{},
		&models.Badge{},
		&models.UserBadge{},
	)

	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
}

// getEnv returns an environment variable or a fallback value if not set
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
