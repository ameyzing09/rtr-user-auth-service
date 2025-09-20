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
	ID                 string
	TenantID           string
	Name               string
	Email              string
	Role               models.Role
	MustChangePassword bool
}

type LoginInput struct {
	Email    string
	Password string
}

type CreateUserInput struct {
	Email string
	Name  string
	Role  models.Role
}

type ChangePasswordInput struct {
	CurrentPassword string
	NewPassword     string
}

type CreateTenantInput struct {
	Name   string
	Domain string
	Email  string
}

type AuthService interface {
	Login(ctx context.Context, input LoginInput) (AuthToken, UserRead, error)
	GetMe(ctx context.Context, userID, tenantID string) (UserRead, error)
	ListUsers(ctx context.Context, tenantID string) ([]UserRead, error)
	CreateUser(ctx context.Context, tenantID string, actor UserRead, input CreateUserInput) (UserRead, string, error)
	ChangePassword(ctx context.Context, tenantID string, actor UserRead, input ChangePasswordInput) error
}

type UserRepository interface {
	EmailExists(ctx context.Context, tenantID, email string) (bool, error)
	Create(ctx context.Context, u *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, tenantID, userID string) (*models.User, error)
	ListByTenant(ctx context.Context, tenantID string) ([]models.User, error)
	UpdatePassword(ctx context.Context, tenantID, userID, hashedPassword string, clearForce bool) error
}

type TenantRepository interface {
	Exists(ctx context.Context, tenantID string) (bool, error)
	FindByDomain(ctx context.Context, domain string) (*models.Tenant, error)
	FindByID(ctx context.Context, tenantID string) (*models.Tenant, error)
}

type TenantSettingService interface {
	Get(ctx context.Context, tenantID string) (map[string]interface{}, error)
	PutReplace(ctx context.Context, tenantID string, cfg map[string]interface{}) (map[string]interface{}, error)
}
