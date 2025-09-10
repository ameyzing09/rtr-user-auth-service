package utils

import (
	"errors"
	"time"

	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/entities"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents JWT claims with tenant and role information
type JWTClaims struct {
	UserID   uuid.UUID      `json:"user_id"`
	TenantID uuid.UUID      `json:"tenant_id"`
	Email    string         `json:"email"`
	Role     entities.Role  `json:"role"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	secret           []byte
	accessExpiry     time.Duration
	refreshExpiry    time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secret string, accessExpiry, refreshExpiry time.Duration) *JWTService {
	return &JWTService{
		secret:           []byte(secret),
		accessExpiry:     accessExpiry,
		refreshExpiry:    refreshExpiry,
	}
}

// GenerateAccessToken generates a new access token
func (j *JWTService) GenerateAccessToken(user *entities.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "recrutr-auth-service",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// GenerateRefreshToken generates a new refresh token
func (j *JWTService) GenerateRefreshToken(user *entities.User) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "recrutr-auth-service",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetTokenExpiry returns the access token expiry duration
func (j *JWTService) GetTokenExpiry() time.Duration {
	return j.accessExpiry
}

// GetRefreshTokenExpiry returns the refresh token expiry duration
func (j *JWTService) GetRefreshTokenExpiry() time.Duration {
	return j.refreshExpiry
}