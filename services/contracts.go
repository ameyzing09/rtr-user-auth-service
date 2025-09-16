package services

import (
	"context"
	"rtr-user-auth-service/models"
	"time"
)

type AuthToken struct {
	Token     string
	ExpiresAt time.Time
}

type UserRead struct {
	ID       string
	TenantID string
	Name     string
	Email    string
	Role     models.Role
}

type RegisterInput struct {
	TenantID string
	Name     string
	Email    string
	Password string
	Role     models.Role
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthService interface {
	Register(ctx context.Context, actor UserRead, input RegisterInput) (UserRead, error)
	Login(ctx context.Context, input LoginInput) (AuthToken, UserRead, error)
	GetMe(ctx context.Context, userID, tenantID string) (UserRead, error)
	ListUsers(ctx context.Context, tenantID string) ([]UserRead, error)
}

type UserRepository interface {
	EmailExists(ctx context.Context, tenantID, email string) (bool, error)
	Create(ctx context.Context, u *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, tenantID, userID string) (*models.User, error)
	ListByTenant(ctx context.Context, tenantID string) ([]models.User, error)
}

type TenantRepository interface {
	Exists(ctx context.Context, tenantID string) (bool, error)
}
