package models

type TenantSetting struct {
	TenantID  string  `gorm:"type:char(36);primaryKey"`
	Config    JSONMap `gorm:"type:json:not null"`
	CreatedAt int64   `gorm:"autoCreateTime:milli"`
	UpdatedAt int64   `gorm:"autoUpdateTime:milli"`
}

func (TenantSetting) TableName() string {
	return "tenants_settings"
}
