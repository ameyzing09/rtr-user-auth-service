package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/entities"
	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/repositories"
	"github.com/ameyzing09/rtr-user-auth-service/internal/utils"
	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user is inactive")
	ErrTenantInactive     = errors.New("tenant is inactive")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrTokenExpired        = errors.New("token expired")
)

// AuthService handles authentication operations
type AuthService interface {
	Login(ctx context.Context, tenantID uuid.UUID, email, password string) (*LoginResponse, error)
	RefreshToken(ctx context.Context, tenantID uuid.UUID, refreshToken string) (*LoginResponse, error)
	Logout(ctx context.Context, tenantID, userID uuid.UUID) error
	ValidateToken(ctx context.Context, tokenString string) (*utils.JWTClaims, error)
}

// LoginResponse represents the response from login operation
type LoginResponse struct {
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
	TokenType    string           `json:"token_type"`
	ExpiresIn    int64            `json:"expires_in"`
	User         *entities.User   `json:"user"`
}

// authService implements AuthService interface
type authService struct {
	userRepo         repositories.UserRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	tenantRepo       repositories.TenantRepository
	jwtService       *utils.JWTService
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo repositories.UserRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	tenantRepo repositories.TenantRepository,
	jwtService *utils.JWTService,
) AuthService {
	return &authService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		tenantRepo:       tenantRepo,
		jwtService:       jwtService,
	}
}

func (s *authService) Login(ctx context.Context, tenantID uuid.UUID, email, password string) (*LoginResponse, error) {
	// Verify tenant exists and is active
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	if tenant == nil {
		return nil, ErrInvalidCredentials
	}
	if !tenant.IsActive {
		return nil, ErrTenantInactive
	}

	// Get user by email within the tenant
	user, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Verify password
	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// Generate access token
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshTokenString, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in database
	refreshToken := &entities.RefreshToken{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTokenExpiry()),
		IsRevoked: false,
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Clear password from response
	user.Password = ""

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtService.GetTokenExpiry().Seconds()),
		User:         user,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, tenantID uuid.UUID, refreshTokenString string) (*LoginResponse, error) {
	// Get refresh token from database
	refreshToken, err := s.refreshTokenRepo.GetByToken(ctx, tenantID, refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	if refreshToken == nil {
		return nil, ErrInvalidRefreshToken
	}

	// Check if token is valid
	if !refreshToken.IsValid() {
		return nil, ErrTokenExpired
	}

	// Verify JWT token
	claims, err := s.jwtService.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, tenantID, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrUserInactive
	}

	// Generate new access token
	accessToken, err := s.jwtService.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate new refresh token
	newRefreshTokenString, err := s.jwtService.GenerateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Revoke old refresh token
	refreshToken.IsRevoked = true
	if err := s.refreshTokenRepo.Update(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	// Store new refresh token
	newRefreshToken := &entities.RefreshToken{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Token:     newRefreshTokenString,
		ExpiresAt: time.Now().Add(s.jwtService.GetRefreshTokenExpiry()),
		IsRevoked: false,
	}

	if err := s.refreshTokenRepo.Create(ctx, newRefreshToken); err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	// Clear password from response
	user.Password = ""

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.jwtService.GetTokenExpiry().Seconds()),
		User:         user,
	}, nil
}

func (s *authService) Logout(ctx context.Context, tenantID, userID uuid.UUID) error {
	// Revoke all refresh tokens for the user
	return s.refreshTokenRepo.RevokeByUserID(ctx, tenantID, userID)
}

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (*utils.JWTClaims, error) {
	return s.jwtService.ValidateToken(tokenString)
}