package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
)

var (
	keycloakClient *gocloak.GoCloak
	realm          string
)

// InitKeycloak initializes the Keycloak client and loads configuration
func InitKeycloak() {
	keycloakURL := os.Getenv("KEYCLOAK_URL")
	realm = os.Getenv("KEYCLOAK_REALM")

	// Both URL and realm are required to validate tokens
	if keycloakURL == "" || realm == "" {
		log.Fatal("KEYCLOAK_URL and KEYCLOAK_REALM must be set")
	}

	// Create Keycloak client
	keycloakClient = gocloak.NewClient(keycloakURL)
}

// AuthMiddleware validates JWT access tokens issued by Keycloak
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Read Authorization header
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			ctx.Abort()
			return
		}

		// Remove "Bearer prefix
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			ctx.Abort()
			return
		}

		// Decode and validate the JWT access token
		_, claims, err := keycloakClient.DecodeAccessToken(
			context.Background(),
			token,
			realm,
		)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			ctx.Abort()
			return
		}

		// Extract the user ID from the "sub" (subject) claim
		sub, _ := (*claims)["sub"].(string)

		// Store user ID in Gin context for later handlers
		ctx.Set("user_id", sub)

		ctx.Next()
	}
}

// GetUserID retrieves the Keycloak user ID from the Gin context
func GetUserID(ctx *gin.Context) (string, bool) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}
