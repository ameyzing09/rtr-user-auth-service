package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tenant struct {
	ID       string    `gorm:"type:char(36);primaryKey"`
	Name     string    `gorm:"type:varchar(255);uniqueIndex:ux_tenants_name,priority:1, not null"`
	CreateAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdateAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
}

func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	return nil
}

func (Tenant) TableName() string {
	return "tenants"
}
