package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Plan string

const (
	PlanBasic      Plan = "BASIC"
	PlanStarter    Plan = "STARTER"
	PlanGrowth     Plan = "GROWTH"
	PlanEnterprise Plan = "ENTERPRISE"
	PlanOnPrem     Plan = "ON_PREM"
)

type TenantStatus string

const (
	TenantPending          TenantStatus = "PENDING"
	TenantProvisioning     TenantStatus = "PROVISIONING"
	TenantAwaitingBranding TenantStatus = "AWAITING_BRANDING"
	TenantActive           TenantStatus = "ACTIVE"
	TenantFailed           TenantStatus = "FAILED"
	TenantSuspended        TenantStatus = "SUSPENDED"
	TenantDeleted          TenantStatus = "DELETED"
)

type Tenant struct {
	ID           string         `gorm:"type:char(36);primaryKey"`
	Name         string         `gorm:"type:varchar(255);not null"`
	Domain       *string        `gorm:"type:varchar(255);uniqueIndex:ux_tenants_domain"`
	Slug         *string        `gorm:"type:varchar(50);uniqueIndex:ux_tenants_slug"`
	Plan         *Plan          `gorm:"type:enum('BASIC','STARTER','GROWTH','ENTERPRISE','ON_PREM')"`
	Status       TenantStatus   `gorm:"type:enum('PENDING','PROVISIONING','AWAITING_BRANDING','ACTIVE','FAILED','SUSPENDED','DELETED');not null;default:'PENDING'"`
	CreatedBy    *string        `gorm:"type:char(36)"`
	FailedReason *string        `gorm:"type:text"`
	CreatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	if t.Status == "" {
		t.Status = TenantPending
	}
	return nil
}

func (Tenant) TableName() string {
	return "tenants"
}
