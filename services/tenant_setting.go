package services

import (
	"context"
	"rtr-user-auth-service/models"
)

type TenantSettingRepository interface {
	GetSettings(ctx context.Context, tenantID string) (*models.TenantSetting, error)
	UpsertReplace(ctx context.Context, ts *models.TenantSetting) error
}

var _ TenantSettingService = (*tenantSettingService)(nil)

type tenantSettingService struct {
	tenantSettings TenantSettingRepository
}

func NewTenantSettingService(repo TenantSettingRepository) *tenantSettingService {
	return &tenantSettingService{tenantSettings: repo}
}

func (s *tenantSettingService) Get(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	row, err := s.tenantSettings.GetSettings(ctx, tenantID)
	if err != nil {
		return map[string]interface{}{}, err
	}
	if row == nil || row.Config == nil {
		return map[string]interface{}{}, nil
	}
	return map[string]interface{}(row.Config), nil
}

func (s *tenantSettingService) PutReplace(ctx context.Context, tenantID string, cfg map[string]interface{}) (map[string]interface{}, error) {
	ts := &models.TenantSetting{
		TenantID: tenantID,
		Config:   models.JSONMap(cfg),
	}
	if err := s.tenantSettings.UpsertReplace(ctx, ts); err != nil {
		return nil, err
	}
	return map[string]interface{}(ts.Config), nil
}
