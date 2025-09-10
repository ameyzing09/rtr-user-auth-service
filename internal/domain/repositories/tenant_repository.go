package repositories

import (
	"context"
	"errors"

	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// tenantRepository implements TenantRepository interface
type tenantRepository struct {
	db *gorm.DB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *entities.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *tenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Tenant, error) {
	var tenant entities.Tenant
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) GetByDomain(ctx context.Context, domain string) (*entities.Tenant, error) {
	var tenant entities.Tenant
	err := r.db.WithContext(ctx).Where("domain = ?", domain).First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) Update(ctx context.Context, tenant *entities.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

func (r *tenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.Tenant{}, "id = ?", id).Error
}

func (r *tenantRepository) List(ctx context.Context, limit, offset int) ([]*entities.Tenant, int64, error) {
	var tenants []*entities.Tenant
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&entities.Tenant{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get tenants with pagination
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Find(&tenants).Error

	return tenants, total, err
}