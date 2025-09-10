package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/entities"
	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/repositories"
	"github.com/ameyzing09/rtr-user-auth-service/internal/utils"
	"github.com/google/uuid"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidRole       = errors.New("invalid role")
	ErrTenantNotFound    = errors.New("tenant not found")
)

// UserService handles user operations
type UserService interface {
	Create(ctx context.Context, req *CreateUserRequest) (*entities.User, error)
	GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*entities.User, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entities.User, error)
	Update(ctx context.Context, tenantID, userID uuid.UUID, req *UpdateUserRequest) (*entities.User, error)
	Delete(ctx context.Context, tenantID, userID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, req *ListUsersRequest) (*ListUsersResponse, error)
	ListByRole(ctx context.Context, tenantID uuid.UUID, role entities.Role, req *ListUsersRequest) (*ListUsersResponse, error)
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	TenantID  uuid.UUID     `json:"tenant_id" validate:"required"`
	Email     string        `json:"email" validate:"required,email"`
	Password  string        `json:"password" validate:"required,min=8"`
	FirstName string        `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string        `json:"last_name" validate:"required,min=2,max=50"`
	Role      entities.Role `json:"role" validate:"required"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Email     *string        `json:"email,omitempty" validate:"omitempty,email"`
	FirstName *string        `json:"first_name,omitempty" validate:"omitempty,min=2,max=50"`
	LastName  *string        `json:"last_name,omitempty" validate:"omitempty,min=2,max=50"`
	Role      *entities.Role `json:"role,omitempty"`
	IsActive  *bool          `json:"is_active,omitempty"`
}

// ListUsersRequest represents the request to list users
type ListUsersRequest struct {
	Page  int `json:"page" validate:"min=1"`
	Limit int `json:"limit" validate:"min=1,max=100"`
}

// ListUsersResponse represents the response from listing users
type ListUsersResponse struct {
	Users      []*entities.User `json:"users"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

// userService implements UserService interface
type userService struct {
	userRepo   repositories.UserRepository
	tenantRepo repositories.TenantRepository
}

// NewUserService creates a new user service
func NewUserService(
	userRepo repositories.UserRepository,
	tenantRepo repositories.TenantRepository,
) UserService {
	return &userService{
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}
}

func (s *userService) Create(ctx context.Context, req *CreateUserRequest) (*entities.User, error) {
	// Validate role
	if !req.Role.IsValid() {
		return nil, ErrInvalidRole
	}

	// Check if tenant exists and is active
	tenant, err := s.tenantRepo.GetByID(ctx, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	if tenant == nil {
		return nil, ErrTenantNotFound
	}
	if !tenant.IsActive {
		return nil, ErrTenantInactive
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.TenantID, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &entities.User{
		TenantID:  req.TenantID,
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      req.Role,
		IsActive:  true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Clear password from response
	user.Password = ""
	return user, nil
}

func (s *userService) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Clear password from response
	user.Password = ""
	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entities.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Clear password from response
	user.Password = ""
	return user, nil
}

func (s *userService) Update(ctx context.Context, tenantID, userID uuid.UUID, req *UpdateUserRequest) (*entities.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Update fields if provided
	if req.Email != nil {
		// Check if new email already exists
		if *req.Email != user.Email {
			existingUser, err := s.userRepo.GetByEmail(ctx, tenantID, *req.Email)
			if err != nil {
				return nil, fmt.Errorf("failed to check existing email: %w", err)
			}
			if existingUser != nil {
				return nil, ErrUserAlreadyExists
			}
		}
		user.Email = *req.Email
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}

	if req.LastName != nil {
		user.LastName = *req.LastName
	}

	if req.Role != nil {
		if !req.Role.IsValid() {
			return nil, ErrInvalidRole
		}
		user.Role = *req.Role
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	// Save user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Clear password from response
	user.Password = ""
	return user, nil
}

func (s *userService) Delete(ctx context.Context, tenantID, userID uuid.UUID) error {
	// Check if user exists
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, tenantID, userID)
}

func (s *userService) List(ctx context.Context, tenantID uuid.UUID, req *ListUsersRequest) (*ListUsersResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	offset := (req.Page - 1) * req.Limit

	users, total, err := s.userRepo.List(ctx, tenantID, req.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Clear passwords from response
	for _, user := range users {
		user.Password = ""
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &ListUsersResponse{
		Users:      users,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *userService) ListByRole(ctx context.Context, tenantID uuid.UUID, role entities.Role, req *ListUsersRequest) (*ListUsersResponse, error) {
	// Validate role
	if !role.IsValid() {
		return nil, ErrInvalidRole
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	offset := (req.Page - 1) * req.Limit

	users, total, err := s.userRepo.ListByRole(ctx, tenantID, role, req.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users by role: %w", err)
	}

	// Clear passwords from response
	for _, user := range users {
		user.Password = ""
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &ListUsersResponse{
		Users:      users,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}