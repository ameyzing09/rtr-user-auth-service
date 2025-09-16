package services

import (
	"context"
	"rtr-user-auth-service/domain"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/utils"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type authService struct {
	db      *gorm.DB
	users   UserRepository
	tenants TenantRepository
}

func NewAuthService(db *gorm.DB, u UserRepository, t TenantRepository) *authService {
	return &authService{db: db, users: u, tenants: t}
}

func (s *authService) Register(ctx context.Context, actor UserRead, input RegisterInput) (UserRead, error) {
	switch actor.Role {
	case models.RoleAdmin:
	case models.RoleHR:
		if input.Role != models.RoleCandidate {
			return UserRead{}, domain.ErrForbidden
		}
	default:
		return UserRead{}, domain.ErrForbidden
	}
	if actor.TenantID != input.TenantID {
		return UserRead{}, domain.ErrForbidden
	}
	exists, err := s.tenants.Exists(ctx, input.TenantID)
	if err != nil {
		return UserRead{}, err
	}

	if !exists {
		return UserRead{}, domain.ErrTenantNotFound
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	already, err := s.users.EmailExists(ctx, input.TenantID, email)
	if err != nil {
		return UserRead{}, err
	}
	if already {
		return UserRead{}, domain.ErrEmailInUse
	}

	hash, err := utils.HashPassword(input.Password)
	if err != nil {
		return UserRead{}, err
	}

	user := &models.User{
		ID:       uuid.NewString(),
		TenantID: input.TenantID,
		Name:     strings.TrimSpace(input.Name),
		Email:    email,
		Password: hash,
		Role:     input.Role,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return UserRead{}, err
	}

	return UserRead{
		ID:       user.ID,
		TenantID: user.TenantID,
		Name:     user.Name,
		Email:    user.Email,
		Role:     user.Role,
	}, nil
}

func (s *authService) Login(ctx context.Context, input LoginInput) (AuthToken, UserRead, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return AuthToken{}, UserRead{}, domain.ErrInvalidCredentials
	}
	if !utils.CheckPassword(user.Password, input.Password) {
		return AuthToken{}, UserRead{}, domain.ErrInvalidCredentials
	}
	token, exp, err := utils.SignJWT(user.ID, user.TenantID, string(user.Role), 24*time.Hour)
	if err != nil {
		return AuthToken{}, UserRead{}, err
	}
	return AuthToken{Token: token, ExpiresAt: exp}, UserRead{
		ID:       user.ID,
		TenantID: user.TenantID,
		Name:     user.Name,
		Email:    user.Email,
		Role:     user.Role,
	}, nil
}

func (s *authService) GetMe(ctx context.Context, userID, tenantID string) (UserRead, error) {
	user, err := s.users.FindByID(ctx, userID, tenantID)
	if err != nil {
		return UserRead{}, err
	}
	return UserRead{
		ID:       user.ID,
		TenantID: user.TenantID,
		Name:     user.Name,
		Email:    user.Email,
		Role:     user.Role,
	}, nil
}

func (s *authService) ListUsers(ctx context.Context, tenantID string) ([]UserRead, error) {
	list, err := s.users.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	output := make([]UserRead, len(list))
	for _, user := range list {
		output = append(output, UserRead{
			ID:       user.ID,
			TenantID: user.TenantID,
			Name:     user.Name,
			Email:    user.Email,
			Role:     user.Role,
		})
	}
	return output, nil
}
