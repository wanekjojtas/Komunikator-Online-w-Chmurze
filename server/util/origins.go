package util

import (
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetAllowedOrigins() []string {
    if gin.Mode() == gin.ReleaseMode {
        origins := os.Getenv("FRONTEND_ORIGIN")
        if origins == "" {
            log.Println("FRONTEND_ORIGIN is not set, using default localhost")
            return []string{"http://localhost:3000"}
        }
        log.Printf("CORS Allowed Origins: %v", origins)
        return strings.Split(origins, ",")
    }
    return []string{"http://localhost:3000"}
}
