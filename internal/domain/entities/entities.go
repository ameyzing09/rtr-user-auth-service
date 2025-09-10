package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Role represents user roles in the system
type Role string

const (
	RoleAdmin       Role = "ADMIN"
	RoleHR          Role = "HR"
	RoleInterviewer Role = "INTERVIEWER"
	RoleCandidate   Role = "CANDIDATE"
)

// IsValid checks if the role is valid
func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleHR, RoleInterviewer, RoleCandidate:
		return true
	default:
		return false
	}
}

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID        uuid.UUID      `json:"id" gorm:"type:char(36);primary_key" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name      string         `json:"name" gorm:"not null;unique" validate:"required,min=2,max=100" example:"Acme Corp"`
	Domain    string         `json:"domain" gorm:"not null;unique" validate:"required,hostname" example:"acme.com"`
	IsActive  bool           `json:"is_active" gorm:"default:true" example:"true"`
	CreatedAt time.Time      `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time      `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	Users []User `json:"-" gorm:"foreignKey:TenantID"`
}

// BeforeCreate generates UUID before creating tenant
func (t *Tenant) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// User represents a user in the system
type User struct {
	ID         uuid.UUID      `json:"id" gorm:"type:char(36);primary_key" example:"123e4567-e89b-12d3-a456-426614174001"`
	TenantID   uuid.UUID      `json:"tenant_id" gorm:"type:char(36);not null;index:idx_tenant_email,unique" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email      string         `json:"email" gorm:"not null;index:idx_tenant_email,unique" validate:"required,email" example:"user@acme.com"`
	Password   string         `json:"-" gorm:"not null" validate:"required,min=8"`
	FirstName  string         `json:"first_name" gorm:"not null" validate:"required,min=2,max=50" example:"John"`
	LastName   string         `json:"last_name" gorm:"not null" validate:"required,min=2,max=50" example:"Doe"`
	Role       Role           `json:"role" gorm:"not null" validate:"required" example:"CANDIDATE"`
	IsActive   bool           `json:"is_active" gorm:"default:true" example:"true"`
	CreatedAt  time.Time      `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt  time.Time      `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	Tenant Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// BeforeCreate generates UUID before creating user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// RefreshToken represents refresh tokens for JWT authentication
type RefreshToken struct {
	ID        uuid.UUID      `json:"id" gorm:"type:char(36);primary_key" example:"123e4567-e89b-12d3-a456-426614174002"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:char(36);not null;index" example:"123e4567-e89b-12d3-a456-426614174001"`
	TenantID  uuid.UUID      `json:"tenant_id" gorm:"type:char(36);not null;index" example:"123e4567-e89b-12d3-a456-426614174000"`
	Token     string         `json:"token" gorm:"not null;unique"`
	ExpiresAt time.Time      `json:"expires_at" example:"2023-01-08T00:00:00Z"`
	IsRevoked bool           `json:"is_revoked" gorm:"default:false" example:"false"`
	CreatedAt time.Time      `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt time.Time      `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Relations
	User   User   `json:"-" gorm:"foreignKey:UserID"`
	Tenant Tenant `json:"-" gorm:"foreignKey:TenantID"`
}

// BeforeCreate generates UUID before creating refresh token
func (rt *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	return nil
}

// IsExpired checks if the refresh token is expired
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid checks if the refresh token is valid (not expired and not revoked)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsExpired() && !rt.IsRevoked
}