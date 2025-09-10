package repositories

import (
	"context"
	"errors"

	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("email = ? AND tenant_id = ?", email, tenantID).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, tenantID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&entities.User{}, "id = ? AND tenant_id = ?", userID, tenantID).Error
}

func (r *userRepository) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Get total count for the tenant
	if err := r.db.WithContext(ctx).
		Model(&entities.User{}).
		Where("tenant_id = ?", tenantID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get users with pagination
	err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("tenant_id = ?", tenantID).
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}

func (r *userRepository) ListByRole(ctx context.Context, tenantID uuid.UUID, role entities.Role, limit, offset int) ([]*entities.User, int64, error) {
	var users []*entities.User
	var total int64

	// Get total count for the tenant and role
	if err := r.db.WithContext(ctx).
		Model(&entities.User{}).
		Where("tenant_id = ? AND role = ?", tenantID, role).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get users with pagination
	err := r.db.WithContext(ctx).
		Preload("Tenant").
		Where("tenant_id = ? AND role = ?", tenantID, role).
		Limit(limit).
		Offset(offset).
		Find(&users).Error

	return users, total, err
}