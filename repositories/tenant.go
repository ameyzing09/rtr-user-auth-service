package repositories

import (
	"context"
	"rtr-user-auth-service/models"

	"gorm.io/gorm"
)

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
		Where("id=?", tenantID).
		Count(&count).
		Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
