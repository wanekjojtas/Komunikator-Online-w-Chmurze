package config

import "os"

// GetSecretKey retrieves the JWT secret key from an environment variable or defaults to "secret"
func GetSecretKey() string {
    key := os.Getenv("JWT_SECRET")
    if key == "" {
        key = "secret" // Default value for development
    }
    return key
}
