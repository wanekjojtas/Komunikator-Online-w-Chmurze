package util

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"server/config"

	"github.com/golang-jwt/jwt/v4"
)

type MyJWTClaims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// RateLimiter to track token refresh attempts
type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string]time.Time
	interval time.Duration
}

// NewRateLimiter initializes a new rate limiter
func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{
		attempts: make(map[string]time.Time),
		interval: interval,
	}
}

// Allow checks if a token refresh is allowed
func (rl *RateLimiter) Allow(userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	lastAttempt, exists := rl.attempts[userID]
	if !exists || time.Since(lastAttempt) > rl.interval {
		rl.attempts[userID] = time.Now()
		log.Printf("RateLimiter: Refresh allowed for user %s", userID)
		return true
	}

	log.Printf("RateLimiter: Refresh denied for user %s. Retry after %v", userID, rl.interval-time.Since(lastAttempt))
	return false
}

// Global instance of RateLimiter for refresh tokens
var RefreshRateLimiter = NewRateLimiter(1 * time.Minute) // Limit to 1 refresh per minute

// GenerateAccessToken generates a short-lived access token
func GenerateAccessToken(userID, username string) (string, error) {
	claims := MyJWTClaims{
		ID:       userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // 15 minutes expiry
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetSecretKey())) // Use Access Secret Key
}

// GenerateRefreshToken generates a long-lived refresh token
func GenerateRefreshToken(userID, username string) (string, error) {
	claims := MyJWTClaims{
		ID:       userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days expiry
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetRefreshSecretKey())) // Use Refresh Secret Key
}

// ValidateToken validates a JWT and returns the claims
func ValidateToken(tokenString string, isRefreshToken bool) (*MyJWTClaims, error) {
	var secretKey string
	if isRefreshToken {
		secretKey = config.GetRefreshSecretKey() // Use refresh secret key
	} else {
		secretKey = config.GetSecretKey() // Use access secret key
	}

	parsedToken, err := jwt.ParseWithClaims(tokenString, &MyJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsedToken.Claims.(*MyJWTClaims)
	if !ok || !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token is expired")
	}

	return claims, nil
}

// RefreshToken validates the refresh token and generates a new access token
func RefreshToken(refreshToken string) (string, error) {
	claims, err := ValidateToken(refreshToken, true) // Validate refresh token
	if err != nil {
		return "", err
	}

	userID := claims.ID

	// Check rate limiter
	if !RefreshRateLimiter.Allow(userID) {
		return "", errors.New("rate limit exceeded. Try again later")
	}

	// Generate new access token
	return GenerateAccessToken(userID, claims.Username)
}

// RefreshTokenHandler handles token refresh requests
func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Header.Get("Authorization")
	if refreshToken == "" {
		http.Error(w, "No refresh token provided", http.StatusUnauthorized)
		return
	}

	// Remove the "Bearer " prefix if present
	if len(refreshToken) > 7 && refreshToken[:7] == "Bearer " {
		refreshToken = refreshToken[7:]
	}

	newAccessToken, err := RefreshToken(refreshToken)
	if err != nil {
		if err.Error() == "rate limit exceeded. Try again later" {
			http.Error(w, err.Error(), http.StatusTooManyRequests)
		} else {
			http.Error(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		}
		return
	}

	// Respond with the new access token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"accessToken": newAccessToken})
}
