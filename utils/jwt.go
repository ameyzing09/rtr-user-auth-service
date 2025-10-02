package utils

import (
	"fmt"
	"time"

	"rtr-user-auth-service/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   string `json:"uid"`
	TenantID string `json:"tid"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func SignJWT(userID, tenantID, email, role string, ttl time.Duration) (string, time.Time, error) {
	cfg := config.Get()
	if cfg == nil {
		return "", time.Time{}, fmt.Errorf("config not initialized")
	}

	secret := cfg.JWT.Secret
	exp := time.Now().Add(ttl)
	claims := &Claims{
		UserID:   userID,
		TenantID: tenantID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign JWT: %w", err)
	}
	return signed, exp, nil
}
