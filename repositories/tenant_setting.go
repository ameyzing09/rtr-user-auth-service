package repositories

import (
	"context"
	"rtr-user-auth-service/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormTenantSettingRepo struct {
	db *gorm.DB
}

func NewGormTenantSettingRepo(db *gorm.DB) *GormTenantSettingRepo {
	return &GormTenantSettingRepo{db: db}
}

func (r *GormTenantSettingRepo) Get(ctx context.Context, tenantID string) (*models.TenantSetting, error) {
	var settings models.TenantSetting
	if err := r.db.WithContext(ctx).First(&settings, "tenant_id = ?", tenantID).Error; err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *GormTenantSettingRepo) PutReplace(ctx context.Context, ts *models.TenantSetting) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"config", "updated_at"}),
	}).Create(ts).Error
}
