package util

import (
	"log"
	"os"
	"strings"
)

func GetAllowedOrigins() []string {
    allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
    if allowedOrigins == "" {
        log.Println("ALLOWED_ORIGINS is not set, using default localhost")
        return []string{"http://localhost:3000"}
    }
    log.Printf("CORS Allowed Origins: %s", allowedOrigins)
    return strings.Split(allowedOrigins, ",")
}

