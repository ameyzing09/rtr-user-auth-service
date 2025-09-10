package repositories

import (
	"context"

	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/entities"
	"github.com/google/uuid"
)

// TenantRepository defines the interface for tenant operations
type TenantRepository interface {
	Create(ctx context.Context, tenant *entities.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Tenant, error)
	GetByDomain(ctx context.Context, domain string) (*entities.Tenant, error)
	Update(ctx context.Context, tenant *entities.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*entities.Tenant, int64, error)
}

// UserRepository defines the interface for user operations
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*entities.User, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, tenantID, userID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entities.User, int64, error)
	ListByRole(ctx context.Context, tenantID uuid.UUID, role entities.Role, limit, offset int) ([]*entities.User, int64, error)
}

// RefreshTokenRepository defines the interface for refresh token operations
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entities.RefreshToken) error
	GetByToken(ctx context.Context, tenantID uuid.UUID, token string) (*entities.RefreshToken, error)
	GetByUserID(ctx context.Context, tenantID, userID uuid.UUID) ([]*entities.RefreshToken, error)
	Update(ctx context.Context, token *entities.RefreshToken) error
	RevokeByUserID(ctx context.Context, tenantID, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}