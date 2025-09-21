package repositories

import (
	"context"
	"rtr-user-auth-service/models"

	"gorm.io/gorm"
)

type TenantRepository interface {
	Exists(ctx context.Context, tenantID string) (bool, error)
	FindByDomain(ctx context.Context, domain string) (*models.Tenant, error)
	FindByID(ctx context.Context, tenantID string) (*models.Tenant, error)
}

type GormTenantRepo struct {
	db *gorm.DB
}

func NewGormTenantRepo(db *gorm.DB) *GormTenantRepo {
	return &GormTenantRepo{db: db}
}

func (r *GormTenantRepo) Exists(ctx context.Context, tenantID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Count(&count).
		Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *GormTenantRepo) FindByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	var t models.Tenant
	if err := r.db.WithContext(ctx).
		Where("domain = ?", domain).
		First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *GormTenantRepo) FindByID(ctx context.Context, tenantID string) (*models.Tenant, error) {
	var t models.Tenant
	if err := r.db.WithContext(ctx).
		Where("id = ?", tenantID).
		First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}
