package util

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func GetAllowedOrigins() []string {
	if gin.Mode() == gin.ReleaseMode {
		origin := os.Getenv("FRONTEND_ORIGIN")
		if origin == "" {
			log.Fatalf("FRONTEND_ORIGIN environment variable is not set")
		}
		log.Printf("CORS Allowed Origins: %v", origin) // Log allowed origins
		return []string{origin}
	}
	return []string{"http://localhost:3000"}
}
