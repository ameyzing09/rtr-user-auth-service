package repositories

import (
	"context"
	"rtr-user-auth-service/models"

	"gorm.io/gorm"
)

type TenantArchiveRepository interface {
	Create(ctx context.Context, archive *models.TenantArchive) error
	FindByID(ctx context.Context, id string) (*models.TenantArchive, error)
	FindByOriginalTenantID(ctx context.Context, tenantID string) (*models.TenantArchive, error)
	ListPaginated(ctx context.Context, page, pageSize int) ([]models.TenantArchive, int, error)
}

type GormTenantArchiveRepo struct {
	db *gorm.DB
}

func NewGormTenantArchiveRepo(db *gorm.DB) *GormTenantArchiveRepo {
	return &GormTenantArchiveRepo{db: db}
}

func (r *GormTenantArchiveRepo) Create(ctx context.Context, archive *models.TenantArchive) error {
	return r.db.WithContext(ctx).Create(archive).Error
}

func (r *GormTenantArchiveRepo) FindByID(ctx context.Context, id string) (*models.TenantArchive, error) {
	var archive models.TenantArchive
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&archive).Error; err != nil {
		return nil, err
	}
	return &archive, nil
}

func (r *GormTenantArchiveRepo) FindByOriginalTenantID(ctx context.Context, tenantID string) (*models.TenantArchive, error) {
	var archive models.TenantArchive
	if err := r.db.WithContext(ctx).Where("id = ?", tenantID).First(&archive).Error; err != nil {
		return nil, err
	}
	return &archive, nil
}

func (r *GormTenantArchiveRepo) ListPaginated(ctx context.Context, page, pageSize int) ([]models.TenantArchive, int, error) {
	var archives []models.TenantArchive
	var total int64

	// Get total count
	if err := r.db.WithContext(ctx).Model(&models.TenantArchive{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get paginated results
	if err := r.db.WithContext(ctx).Order("deleted_at DESC").Offset(offset).Limit(pageSize).Find(&archives).Error; err != nil {
		return nil, 0, err
	}

	return archives, int(total), nil
}
