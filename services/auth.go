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
	tenantID := strings.TrimSpace(input.TenantID)

	var user *models.User
	var err error

	// Handle SuperAdmin login (without tenant context) vs regular tenant user login
	if tenantID == "" {
		// No tenant context - only allow for SuperAdmin login
		// Query across all tenants, but verify user is SuperAdmin after retrieval
		user, err = s.findSuperAdminByEmail(ctx, email)
		if err != nil {
			return AuthToken{}, UserRead{}, domain.ErrInvalidCredentials
		}
	} else {
		// Tenant context provided - enforce tenant isolation for regular users
		user, err = s.users.FindByEmail(ctx, tenantID, email)
		if err != nil {
			return AuthToken{}, UserRead{}, domain.ErrInvalidCredentials
		}
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

	// Get permissions for this role
	permissions := models.GetRolePermissions(user.Role)

	token, exp, err := utils.SignJWT(user.ID, user.TenantID, user.Email, string(user.Role), permissions, 24*time.Hour)
	if err != nil {
		return AuthToken{}, UserRead{}, err
	}
	return AuthToken{Token: token, ExpiresAt: exp}, toRead(user), nil
}

func (s *authService) GetMe(ctx context.Context, userID, tenantID string) (UserRead, error) {
	user, err := s.users.FindByID(ctx, tenantID, userID)
	if err != nil {
		return UserRead{}, err
	}

	return toRead(user), nil
}

func (s *authService) ListUsers(ctx context.Context, tenantID string) ([]UserRead, error) {
	list, err := s.users.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	output := make([]UserRead, 0, len(list))
	for _, user := range list {
		output = append(output, toRead(&user))
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

	return s.users.UpdatePassword(ctx, tenantID, actor.ID, hashedPassword, utils.BoolPtr(false))
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
	if err := s.users.UpdatePassword(ctx, input.TenantID, input.UserID, hashedPassword, utils.BoolPtr(false)); err != nil {
		return "", err
	}

	return tempPassword, nil
}


// AdminListUsers lists all users, optionally filtered by tenant, role, and search term
// Used by superadmins to manage users across the platform
func (s *authService) AdminListUsers(ctx context.Context, tenantID *string, role *string, search *string, page, limit int) ([]UserRead, int, error) {
	// For now, we'll fetch all users if no tenant_id is provided, or users from a specific tenant
	var users []models.User
	var total int64

	query := s.db.WithContext(ctx)

	// Apply tenant filter if provided
	if tenantID != nil && *tenantID != "" {
		query = query.Where("tenant_id = ?", *tenantID)
	}

	// Apply role filter if provided
	if role != nil && *role != "" {
		query = query.Where("role = ?", *role)
	}

	// Apply search filter if provided (search by name or email)
	if search != nil && *search != "" {
		searchTerm := "%" + *search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", searchTerm, searchTerm)
	}

	// Get total count
	if err := query.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated results
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	// Convert to UserRead
	result := make([]UserRead, 0, len(users))
	for _, user := range users {
		result = append(result, toRead(&user))
	}

	return result, int(total), nil
}

// AdminGetUser gets a specific user by ID (across all tenants)
// Used by superadmins to view user details
func (s *authService) AdminGetUser(ctx context.Context, userID string) (UserRead, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return UserRead{}, domain.ErrUserNotFound
		}
		return UserRead{}, err
	}

	return toRead(&user), nil
}

// AdminResetPassword resets a user's password and optionally forces them to change it on next login
// If newPassword is nil, a temporary password is generated
// If forceChange is true, the user's must_change_password flag is set to true
func (s *authService) AdminResetPassword(ctx context.Context, userID string, newPassword *string, forceChange bool) (string, error) {
	// Get the user
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", domain.ErrUserNotFound
		}
		return "", err
	}

	// Determine the password to use
	var passwordToHash string
	if newPassword != nil && *newPassword != "" {
		passwordToHash = *newPassword
	} else {
		// Generate a temporary password
		var err error
		passwordToHash, err = utils.GenerateTempPassword()
		if err != nil {
			return "", err
		}
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(passwordToHash)
	if err != nil {
		return "", err
	}

	// Update password - set force_password_reset based on forceChange parameter
	// If forceChange is true, set ForcePasswordReset to true; if false, set to false
	if err := s.users.UpdatePassword(ctx, user.TenantID, userID, hashedPassword, utils.BoolPtr(forceChange)); err != nil {
		return "", err
	}

	return passwordToHash, nil
}

// findSuperAdminByEmail searches for a user by email across all tenants
// and returns the user ONLY if they have the SuperAdmin role.
// This is used for admin login where no tenant context is available.
func (s *authService) findSuperAdminByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.db.WithContext(ctx).
		Where("email = ? AND role = ?", email, models.RoleSuperAdmin).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func toRead(u *models.User) UserRead {
	permissions := models.GetRolePermissions(u.Role)
	return UserRead{
		ID: u.ID, TenantID: u.TenantID, Name: u.Name, Email: u.Email,
		Role: u.Role, Permissions: permissions, ForcePasswordReset: u.ForcePasswordReset,
		CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt, LastLogin: nil,
	}
}
