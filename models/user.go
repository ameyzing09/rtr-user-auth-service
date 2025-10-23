package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                  string         `gorm:"type:char(36);primaryKey"`
	TenantID            string         `gorm:"type:char(36);not null;index:idx_users_tenants,priority:1"`
	Name                string         `gorm:"type:varchar(150);not null"`
	Email               string         `gorm:"type:varchar(190);not null;uniqueIndex:ux_users_tenant_email,priority:2"`
	Password            string         `gorm:"type:char(60);not null" json:"-"`
	Role                Role           `gorm:"type:ENUM('SUPERADMIN','ADMIN','HR','INTERVIEWER','VIEWER','CANDIDATE');not null;default:'CANDIDATE'"`
	IsOwner             bool           `gorm:"not null;default:false"`
	IsActive            bool           `gorm:"not null;default:true"`
	ForcePasswordReset  bool           `gorm:"column:force_password_reset;not null;default:false"`
	CreatedAt           time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt           time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	DeletedAt           gorm.DeletedAt `gorm:"index"`

	Tenant Tenant `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:TenantID;references:ID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	return nil
}

func (User) TableName() string { return "users" }
