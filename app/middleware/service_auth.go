package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// ServiceAuthMiddleware protects internal service-to-service endpoints
func ServiceAuthMiddleware() gin.HandlerFunc {
	// Token shared between trusted services
	requiredToken := os.Getenv("USER_SERVICE_TOKEN")

	return func(c *gin.Context) {
		// Read custom service token header
		provided := c.GetHeader("X-Service-Token")

		// Reject request if token is missing or does not match
		if requiredToken == "" || provided == "" || provided != requiredToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid service token",
			})
			return
		}

		// Token is valid, continue request
		c.Next()
	}
}
