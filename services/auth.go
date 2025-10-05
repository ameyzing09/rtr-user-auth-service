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

var _ AuthService = (*authService)(nil) // suggested by claude sonnet 4 for early detection of interface implementation errors

type authService struct {
	db      *gorm.DB
	users   UserRepository
	tenants TenantRepository
	subs    SubscriptionService
}

func NewAuthService(db *gorm.DB, u UserRepository, t TenantRepository, s SubscriptionService) *authService {
	return &authService{db: db, users: u, tenants: t, subs: s}
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

	// Check subscription status for non-superadmin users
	if user.Role != models.RoleSuperAdmin {
		sub, err := s.subs.GetSubscription(ctx, user.TenantID)
		if err != nil {
			return AuthToken{}, UserRead{}, err
		}

		now := time.Now().UTC()
		effectiveStatus := EffectiveStatus(sub, now)
		if effectiveStatus == models.SubSuspended || effectiveStatus == models.SubCanceled {
			return AuthToken{}, UserRead{}, domain.ErrSubscriptionInactive
		}
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
	user, err := s.users.FindByID(ctx, tenantID, userID)
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

	tempPassword, tempPasswordError := utils.GenerateTempPassword()
	if tempPasswordError != nil {
		return UserRead{}, "", tempPasswordError
	}

	hashed, err := utils.HashPassword(tempPassword)
	if err != nil {
		return UserRead{}, "", err
	}

	user := &models.User{
		TenantID:           tenantID,
		ID:                 uuid.NewString(),
		Email:              email,
		Name:               strings.TrimSpace(input.Name),
		Role:               input.Role,
		Password:           hashed,
		ForcePasswordReset: true,
	}

	if err := s.users.Create(ctx, user); err != nil {
		return UserRead{}, "", err
	}
	return toRead(user), tempPassword, nil
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

func (s *authService) SuperadminChangePassword(ctx context.Context, actor UserRead, input SuperadminChangePasswordInput) (string, error) {
	// Only superadmin can use this method
	if actor.Role != models.RoleSuperAdmin {
		return "", domain.ErrForbidden
	}

	// Find the target user
	user, err := s.users.FindByID(ctx, input.TenantID, input.UserID)
	if err != nil {
		return "", ErrUserNotFound
	}

	// Only allow changing password if force_password_reset is true
	if !user.ForcePasswordReset {
		return "", domain.ErrForbidden
	}

	// Generate a new temporary password
	tempPassword, err := utils.GenerateTempPassword()
	if err != nil {
		return "", err
	}

	// Hash the new password
	hashedPassword, err := utils.HashPassword(tempPassword)
	if err != nil {
		return "", err
	}

	// Update password and clear the force_password_reset flag
	if err := s.users.UpdatePassword(ctx, input.TenantID, input.UserID, hashedPassword, true); err != nil {
		return "", err
	}

	return tempPassword, nil
}

func toRead(u *models.User) UserRead {
	return UserRead{
		ID: u.ID, TenantID: u.TenantID, Name: u.Name, Email: u.Email,
		Role: u.Role, MustChangePassword: u.ForcePasswordReset,
	}
}
