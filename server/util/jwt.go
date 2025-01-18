package util

import (
	"errors"
	"time"

	"server/config"

	"github.com/golang-jwt/jwt/v4"
)

type MyJWTClaims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT for a given user ID and username
func GenerateJWT(userID, username string) (string, error) {
	claims := MyJWTClaims{
		ID:       userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetSecretKey()))
}

// ValidateJWT validates a JWT and returns the claims
func ValidateJWT(tokenString string) (*MyJWTClaims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &MyJWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetSecretKey()), nil
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