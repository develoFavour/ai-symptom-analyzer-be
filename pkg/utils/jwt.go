package utils

import (
	"errors"
	"strings"
	"time"

	"ai-symptom-checker/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"` // patient | doctor | admin
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT for a given user
func GenerateToken(userID uuid.UUID, email, role string, isRefresh bool) (string, error) {
	var expiry time.Time
	var secret string

	if isRefresh {
		expiry = time.Now().Add(config.App.JWTRefreshExpiry)
		secret = config.App.JWTRefreshSecret
	} else {
		expiry = time.Now().Add(config.App.JWTExpiry)
		secret = config.App.JWTSecret
	}

	claims := Claims{
		UserID: userID,
		Role:   role,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken parses and validates an access JWT string
func ValidateToken(tokenString string) (*Claims, error) {
	return validate(tokenString, config.App.JWTSecret)
}

// ValidateRefreshToken parses and validates a refresh JWT string
func ValidateRefreshToken(tokenString string) (*Claims, error) {
	return validate(tokenString, config.App.JWTRefreshSecret)
}

func validate(tokenString, secret string) (*Claims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
