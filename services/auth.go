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

func (s *authService) Login(ctx context.Context, input LoginInput) (AuthToken, UserRead, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return AuthToken{}, UserRead{}, domain.ErrInvalidCredentials
	}
	if !utils.CheckPassword(user.Password, input.Password) {
		return AuthToken{}, UserRead{}, domain.ErrInvalidCredentials
	}
	token, exp, err := utils.SignJWT(user.ID, user.TenantID, user.Email, string(user.Role), 24*time.Hour)
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
	output := make([]UserRead, 0, len(list))
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

func (s *authService) CreateUser(ctx context.Context, tenantID string, actor UserRead, input CreateUserInput) (UserRead, string, error) {
	if actor.Role != models.RoleAdmin {
		return UserRead{}, "", domain.ErrForbidden
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	exists, err := s.users.EmailExists(ctx, tenantID, email)
	if err != nil {
		return UserRead{}, "", err
	}
	if exists {
		return UserRead{}, "", domain.ErrEmailInUse
	}

	temp := utils.GenerateTempPassword()

	hashed, err := utils.HashPassword(temp)
	if err != nil {
		return UserRead{}, "", err
	}

	user := &models.User{
		TenantID:            tenantID,
		ID:                  uuid.NewString(),
		Email:               email,
		Name:                strings.TrimSpace(input.Name),
		Role:                input.Role,
		Password:            hashed,
		ForcePasswordChange: true,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return UserRead{}, "", err
	}
	return toRead(user), temp, nil
}

func (s *authService) ChangePassword(ctx context.Context, tenantID string, actor UserRead, input ChangePasswordInput) error {
	u, err := s.users.FindByID(ctx, tenantID, actor.ID)
	if err != nil {
		return ErrUserNotFound
	}

	if !utils.CheckPassword(u.Password, input.CurrentPassword) || input.CurrentPassword == input.NewPassword {
		return ErrInvalidCredentials
	}
	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	return s.users.UpdatePassword(ctx, tenantID, actor.ID, hashedPassword, true)
}

func toRead(u *models.User) UserRead {
	return UserRead{
		ID: u.ID, TenantID: u.TenantID, Name: u.Name, Email: u.Email,
		Role: u.Role, MustChangePassword: u.ForcePasswordChange,
	}
}
