package config

import "os"

// GetSecretKey retrieves the JWT secret key for access tokens from an environment variable or defaults to "access_secret" for development
func GetSecretKey() string {
    key := os.Getenv("JWT_ACCESS_SECRET")
    if key == "" {
        key = "access_secret" // Default value for development
    }
    return key
}

// GetRefreshSecretKey retrieves the JWT secret key for refresh tokens from an environment variable or defaults to "refresh_secret" for development
func GetRefreshSecretKey() string {
    key := os.Getenv("JWT_REFRESH_SECRET")
    if key == "" {
        key = "refresh_secret" // Default value for development
    }
    return key
}
