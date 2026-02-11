package main

import (
	"log"
	"os"
	"time"

	"group1-userservice/app/config"
	controller "group1-userservice/app/controllers"
	"group1-userservice/app/middleware"
	"group1-userservice/app/repository"
	"group1-userservice/app/service"
	"group1-userservice/app/storage"
	_ "group1-userservice/docs"

	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env (optional in Docker)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found — using environment variables")
	}

	// Init Keycloak middleware
	middleware.InitKeycloak()

	// Connect DB
	config.ConnectDatabase()

	// Seed Interests
	config.SeedInterests()

	// Seed Badges
	config.SeedBadges()

	// Services
	userRepo := repository.NewUserRepository(config.DB)
	userService := service.NewUserService(userRepo)

	notifRepo := repository.NewNotificationSettingsRepository()
	notifService := service.NewNotificationSettingsService(notifRepo)

	interestsRepo := repository.NewUserInterestsRepository()
	interestsService := service.NewUserInterestsService(interestsRepo)

	userBadgeRepo := repository.NewUserBadgeRepository(config.DB)
	userBadgeService := service.NewUserBadgeService(userBadgeRepo)

	// Controllers
	registerController := controller.NewRegisterController(userService)
	loginController := controller.NewLoginController(userService)
	userController := controller.NewUserController(userService, userBadgeService)
	notifController := controller.NewNotificationSettingsController(notifService, userService)
	interestsController := controller.NewUserInterestsController(interestsService, userService)
	prefsRepo := repository.NewDiscoveryPreferencesRepository(config.DB)
	prefsService := service.NewDiscoveryPreferencesService(prefsRepo)
	prefsController := controller.NewDiscoveryPreferencesController(prefsService, userService)
	badgeController := controller.NewBadgeController(userBadgeService, userService)

	resetRepo := repository.NewPasswordResetRepository(config.DB)
	resetService := service.NewPasswordResetService(resetRepo, userService)
	notificationURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	resetController := controller.NewPasswordResetController(resetService, notificationURL)

	s3, err := storage.NewS3()
	if err != nil {
		log.Fatalf("failed to init s3: %v", err)
	}

	for i := 1; i <= 10; i++ {
		if err := s3.EnsureBucket(); err == nil {
			log.Println("S3 bucket ready")
			break
		}
		log.Printf("S3 EnsureBucket failed (attempt %d/10): %v", i, err)
		time.Sleep(2 * time.Second)
	}

	// Router
	router := gin.Default()

	// Prometheus metrics
	router.Use(middleware.PrometheusMiddleware())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes
	router.POST("/users/register", registerController.Handle)
	router.POST("/auth/login", loginController.Handle)
	router.POST("/auth/refresh", loginController.Refresh)

	router.POST("/auth/forgot-password", resetController.Forgot)
	router.POST("/auth/reset-password", resetController.Reset)

	// public info by first and last name
	usersProtected := router.Group("/users")
	usersProtected.Use(middleware.AuthMiddleware())

	usersProtected.GET("/:firstname/:lastname", userController.GetByFirstLast)
	usersProtected.GET("/id/:id/badges", userController.GetBadgesByUserID)

	// User routes
	router.GET("/users/keycloak/:sub", userController.GetByKeycloakSub)

	// Protected
	protected := router.Group("/users/me")
	protected.Use(middleware.AuthMiddleware())
	protected.PUT("/notification-settings", notifController.UpdateForMe)
	protected.PUT("", userController.UpdateMe)

	protected.GET("/interests", interestsController.GetForMe)
	protected.PUT("/interests", interestsController.UpdateForMe)

	protected.GET("/discovery-preferences", prefsController.GetForMe)
	protected.PUT("/discovery-preferences", prefsController.UpdateForMe)

	protected.GET("/profile-photo/url", userController.PresignProfilePhotoGet(s3))
	protected.POST("/profile-photo", userController.UploadProfilePhoto(s3))

	protected.GET("/badges", userController.GetMyBadges)

	// Internal service-to-service endpoints
	internal := router.Group("/internal")
	internal.Use(middleware.ServiceAuthMiddleware())
	internal.GET("/users/:email", userController.GetByEmail)
	internal.GET("/users/:email/notification-settings", notifController.GetByEmailInternal)
	internal.GET("/users/:email/interests", interestsController.GetForUserInternal)
	internal.GET("/users/:email/discovery-preferences", prefsController.GetByEmailInternal)
	internal.POST("/badges/award", badgeController.Award)

	// Port
	port := os.Getenv("APP_PORT")

	for _, r := range router.Routes() {
		log.Printf("ROUTE: %s %s", r.Method, r.Path)
	}

	go func() {
		log.Println("Starting metrics server on :9090")
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())

		if err := http.ListenAndServe(":9090", mux); err != nil {
			log.Fatalf("metrics server failed: %v", err)
		}
	}()

	log.Println("User Service running on port", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
