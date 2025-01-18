package middleware

import (
	"log"
	"net/http"
	"server/util"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// Check for token in query parameters for WebSocket routes
		if strings.HasPrefix(c.Request.URL.Path, "/ws/") {
			token = c.Query("token")
			log.Printf("WebSocket request detected. Token from query: %s", token)
		}

		// Fallback to Authorization header for regular HTTP requests
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			token = strings.TrimPrefix(authHeader, "Bearer ")
			log.Printf("Token from Authorization header: %s", token)
		}

		if token == "" {
			log.Println("Token missing")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// Validate the token
		claims, err := util.ValidateJWT(token)
		if err != nil {
			log.Printf("Invalid token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Log validated claims for debugging
		log.Printf("Token validated. UserID: %s, Username: %s", claims.ID, claims.Username)

		// Set user information in the context
		c.Set("userID", claims.ID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
