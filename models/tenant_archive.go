package models

import (
	"time"
)

// TenantArchive represents an archived tenant record
type TenantArchive struct {
	ID           string       `gorm:"type:char(36);primaryKey"`
	Name         string       `gorm:"type:varchar(255);not null"`
	Domain       *string      `gorm:"type:varchar(255)"`
	Slug         *string      `gorm:"type:varchar(50)"`
	Plan         *Plan        `gorm:"type:enum('BASIC','STARTER','GROWTH','ENTERPRISE','ON_PREM')"`
	Status       TenantStatus `gorm:"type:enum('PENDING','PROVISIONING','AWAITING_BRANDING','ACTIVE','FAILED','SUSPENDED','DELETED');not null"`
	CreatedBy    *string      `gorm:"type:char(36)"`
	FailedReason *string      `gorm:"type:text"`
	CreatedAt    time.Time    `gorm:"not null"`
	UpdatedAt    time.Time    `gorm:"not null"`

	// Archive-specific fields
	DeletedBy    string    `gorm:"type:char(36);not null"`
	DeletedAt    time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeleteReason *string   `gorm:"type:text"`
}

func (TenantArchive) TableName() string {
	return "tenant_archives"
}

// FromTenant creates a TenantArchive from a Tenant
func (ta *TenantArchive) FromTenant(tenant *Tenant, deletedBy string, reason *string) {
	ta.ID = tenant.ID
	ta.Name = tenant.Name
	ta.Domain = tenant.Domain
	ta.Slug = tenant.Slug
	ta.Plan = tenant.Plan
	ta.Status = tenant.Status
	ta.CreatedBy = tenant.CreatedBy
	ta.FailedReason = tenant.FailedReason
	ta.CreatedAt = tenant.CreatedAt
	ta.UpdatedAt = tenant.UpdatedAt
	ta.DeletedBy = deletedBy
	ta.DeletedAt = time.Now()
	ta.DeleteReason = reason
}
