package repositories

import (
	"context"
	"rtr-user-auth-service/models"

	"gorm.io/gorm"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *models.Tenant) error
	Update(ctx context.Context, tenant *models.Tenant) error
	UpdateStatus(ctx context.Context, tenantID string, status models.TenantStatus) error
	UpdateStatusWithReason(ctx context.Context, tenantID string, status models.TenantStatus, reason string) error
	FindByID(ctx context.Context, id string) (*models.Tenant, error)
	FindByDomain(ctx context.Context, domain string) (*models.Tenant, error)
	FindBySlug(ctx context.Context, slug string) (*models.Tenant, error)
	ListAll(ctx context.Context) ([]models.Tenant, error)
}

type GormTenantRepo struct {
	db *gorm.DB
}

func NewGormTenantRepo(db *gorm.DB) *GormTenantRepo {
	return &GormTenantRepo{db: db}
}

func (r *GormTenantRepo) Create(ctx context.Context, tenant *models.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *GormTenantRepo) Update(ctx context.Context, tenant *models.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

func (r *GormTenantRepo) FindByID(ctx context.Context, id string) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *GormTenantRepo) FindByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := r.db.WithContext(ctx).Where("domain = ?", domain).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *GormTenantRepo) FindBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	var tenant models.Tenant
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *GormTenantRepo) ListAll(ctx context.Context) ([]models.Tenant, error) {
	var tenants []models.Tenant
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

// UpdateStatus updates only the status field of a tenant
func (r *GormTenantRepo) UpdateStatus(ctx context.Context, tenantID string, status models.TenantStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Update("status", status).Error
}

// UpdateStatusWithReason updates the status and failed_reason fields
func (r *GormTenantRepo) UpdateStatusWithReason(ctx context.Context, tenantID string, status models.TenantStatus, reason string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if reason != "" {
		updates["failed_reason"] = reason
	}
	return r.db.WithContext(ctx).
		Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Updates(updates).Error
}
