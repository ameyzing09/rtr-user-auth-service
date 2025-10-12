package repositories

import (
	"context"
	"rtr-user-auth-service/models"

	"gorm.io/gorm"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *models.Tenant) error
	Update(ctx context.Context, tenant *models.Tenant) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*models.Tenant, error)
	FindByDomain(ctx context.Context, domain string) (*models.Tenant, error)
	FindBySlug(ctx context.Context, slug string) (*models.Tenant, error)
	ListAll(ctx context.Context) ([]models.Tenant, error)
	ListPaginated(ctx context.Context, page, pageSize int) ([]models.Tenant, int, error)
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

func (r *GormTenantRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Tenant{}).Error
}

func (r *GormTenantRepo) ListAll(ctx context.Context) ([]models.Tenant, error) {
	var tenants []models.Tenant
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (r *GormTenantRepo) ListPaginated(ctx context.Context, page, pageSize int) ([]models.Tenant, int, error) {
	var tenants []models.Tenant
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.Tenant{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get paginated results
	if err := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tenants).Error; err != nil {
		return nil, 0, err
	}

	return tenants, int(total), nil
}
