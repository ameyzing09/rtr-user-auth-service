package services

import (
	"context"
	"rtr-user-auth-service/models"
)

var _ TenantSettingService = (*tenantSettingService)(nil)

type tenantSettingService struct {
	tenantSettings TenantSettingRepository
}

func NewTenantSettingService(repo TenantSettingRepository) *tenantSettingService {
	return &tenantSettingService{tenantSettings: repo}
}

func (s *tenantSettingService) GetConfiguration(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	if tenantID == "" {
		return nil, ErrInvalidInput
	}

	row, err := s.tenantSettings.Get(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return configFromModel(row), nil
}

func (s *tenantSettingService) UpdateConfiguration(ctx context.Context, tenantID string, config map[string]interface{}) (map[string]interface{}, error) {
	if tenantID == "" {
		return nil, ErrInvalidInput
	}

	ts := &models.TenantSetting{
		TenantID: tenantID,
		Config:   toJSONMap(config),
	}
	if err := s.tenantSettings.PutReplace(ctx, ts); err != nil {
		return nil, err
	}
	return configFromModel(ts), nil
}

func (s *tenantSettingService) GetConfigurationValue(ctx context.Context, tenantID, key string) (interface{}, error) {
	if tenantID == "" || key == "" {
		return nil, ErrInvalidInput
	}

	config, err := s.GetConfiguration(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	value, exists := config[key]
	if !exists {
		return nil, nil // Return nil for non-existent keys
	}
	return value, nil
}

func (s *tenantSettingService) SetConfigurationValue(ctx context.Context, tenantID, key string, value interface{}) error {
	if tenantID == "" || key == "" {
		return ErrInvalidInput
	}

	// Get current configuration
	config, err := s.GetConfiguration(ctx, tenantID)
	if err != nil {
		// If no config exists, start with empty map
		config = make(map[string]interface{})
	}

	// Update the specific key
	config[key] = value

	// Save the updated configuration
	_, err = s.UpdateConfiguration(ctx, tenantID, config)
	return err
}

func (s *tenantSettingService) RemoveConfigurationValue(ctx context.Context, tenantID, key string) error {
	if tenantID == "" || key == "" {
		return ErrInvalidInput
	}

	// Get current configuration
	config, err := s.GetConfiguration(ctx, tenantID)
	if err != nil {
		return err
	}

	// Remove the key if it exists
	delete(config, key)

	// Save the updated configuration
	_, err = s.UpdateConfiguration(ctx, tenantID, config)
	return err
}

func (s *tenantSettingService) ResetConfiguration(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return ErrInvalidInput
	}

	// Reset to empty configuration
	_, err := s.UpdateConfiguration(ctx, tenantID, make(map[string]interface{}))
	return err
}

func configFromModel(setting *models.TenantSetting) map[string]interface{} {
	if setting == nil {
		return map[string]interface{}{}
	}
	return cloneJSONMap(setting.Config)
}

func cloneJSONMap(cfg models.JSONMap) map[string]interface{} {
	if cfg == nil {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{}, len(cfg))
	for k, v := range cfg {
		out[k] = v
	}
	return out
}

func toJSONMap(cfg map[string]interface{}) models.JSONMap {
	if cfg == nil {
		return models.JSONMap{}
	}
	out := make(models.JSONMap, len(cfg))
	for k, v := range cfg {
		out[k] = v
	}
	return out
}
