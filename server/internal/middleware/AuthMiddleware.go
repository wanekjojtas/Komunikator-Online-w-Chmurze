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

		// Return error if no token is provided
		if token == "" {
			log.Println("Token missing")
			sendUnauthorizedResponse(c, "Token is required")
			return
		}

		// Validate the token (assumes it's an access token, hence `false`)
		claims, err := util.ValidateToken(token, false)
		if err != nil {
			log.Printf("Token validation error: %v", err)

			// Handle specific error cases
			if strings.Contains(err.Error(), "expired") {
				sendUnauthorizedResponse(c, "Token expired")
			} else {
				sendUnauthorizedResponse(c, "Invalid token")
			}
			return
		}

		// Log validated claims for debugging
		log.Printf("Token validated successfully. UserID: %s, Username: %s", claims.ID, claims.Username)

		// Store user information in the context for later use
		c.Set("userID", claims.ID)
		c.Set("username", claims.Username)

		// Proceed to the next handler
		c.Next()
	}
}

// sendUnauthorizedResponse sends a 401 response and closes WebSocket connections properly
func sendUnauthorizedResponse(c *gin.Context, message string) {
	response := gin.H{"error": message}

	// Handle WebSocket requests with a JSON response and abort the context
	if strings.HasPrefix(c.Request.URL.Path, "/ws/") {
		c.JSON(http.StatusUnauthorized, response)
		c.Abort()
		return
	}

	// Handle standard HTTP requests with a JSON response
	c.JSON(http.StatusUnauthorized, response)
	c.Abort()
}
