package handlers

import (
	"time"

	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"
)

type RegisterRequest struct {
	Name     string      `json:"name" binding:"required,min=2"`
	Email    string      `json:"email" binding:"required,email"`
	Password string      `json:"password" binding:"required,min=6"`
	Role     models.Role `json:"role" binding:"required,oneof=admin hr interviewer candidate"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	Token            string            `json:"Token"`
	ExpiresAt        time.Time         `json:"ExpiresAt"`
	User             services.UserRead `json:"User"`
	PlatformBranding *PlatformBranding `json:"PlatformBranding,omitempty"`
	TenantBranding   *TenantBranding   `json:"TenantBranding,omitempty"`
}

type PlatformBranding struct {
	Name         string            `json:"name"`
	LogoURL      string            `json:"logo_url"`
	PrimaryColor string            `json:"primary_color"`
	AccentColor  string            `json:"accent_color"`
	NavbarTitle  string            `json:"navbar_title"`
	SidebarTitle string            `json:"sidebar_title"`
	SidebarLinks []PlatformNavItem `json:"sidebar_links"`
}

type TenantBranding struct {
	Name         string `json:"name"`
	LogoURL      string `json:"logo_url"`
	PrimaryColor string `json:"primary_color"`
	AccentColor  string `json:"accent_color"`
	NavbarTitle  string `json:"navbar_title"`
}

type PlatformNavItem struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Path  string `json:"path"`
}

type CreateUserRequest struct {
	Email string      `json:"email" binding:"required,email"`
	Name  string      `json:"name" binding:"required,min=2"`
	Role  models.Role `json:"role" binding:"required,oneof=admin hr interviewer candidate"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=6"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type SuperadminChangePasswordRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	TenantID string `json:"tenant_id" binding:"required"`
}

type SuperadminChangePasswordResponse struct {
	TemporaryPassword string `json:"temporary_password"`
}

type TenantCreateRequest struct {
	Name       string  `json:"name" binding:"required,min=2"`
	Domain     *string `json:"domain" binding:"omitempty,min=3"`
	AdminName  string  `json:"admin_name" binding:"required,min=2"`
	AdminEmail string  `json:"admin_email" binding:"required,email"`
	Plan       *string `json:"plan" binding:"omitempty,oneof=BASIC STARTER GROWTH ENTERPRISE ON_PREM"`
	IsTrial    bool    `json:"is_trial"`
}

type TenantSummary struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Domain *string `json:"domain,omitempty"`
	Slug   *string `json:"slug,omitempty"`
}

type TenantCreateResponse struct {
	Tenant       TenantSummary `json:"tenant"`
	TempPassword string        `json:"temp_password"`
	Status       string        `json:"status"`
}

type TenantGetResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Domain       *string `json:"domain,omitempty"`
	Slug         *string `json:"slug,omitempty"`
	Plan         *string `json:"plan,omitempty"`
	Status       string  `json:"status"`
	CreatedBy    *string `json:"created_by,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	FailedReason *string `json:"failed_reason,omitempty"`
}

type TenantListItem struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Domain       *string `json:"domain,omitempty"`
	Slug         *string `json:"slug,omitempty"`
	Plan         *string `json:"plan,omitempty"`
	Status       string  `json:"status"`
	CreatedBy    *string `json:"created_by,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	FailedReason *string `json:"failed_reason,omitempty"`
}

type TenantListResponse struct {
	Tenants []TenantListItem `json:"tenants"`
}

type TenantStatusResponse struct {
	Status string   `json:"status"`
	Steps  []string `json:"steps,omitempty"`
}

type TenantUpdateRequest struct {
	Name   *string `json:"name,omitempty" binding:"omitempty,min=2"`
	Domain *string `json:"domain,omitempty" binding:"omitempty,min=3"`
	Plan   *string `json:"plan,omitempty" binding:"omitempty,oneof=BASIC STARTER GROWTH ENTERPRISE ON_PREM"`
	Status *string `json:"status,omitempty" binding:"omitempty,oneof=PENDING PROVISIONING AWAITING_BRANDING ACTIVE FAILED SUSPENDED DELETED"`
}

type TenantListPaginatedResponse struct {
	Tenants  []TenantListItem `json:"tenants"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

type SubscriptionResponse struct {
	ID            uint64  `json:"id"`
	TenantID      string  `json:"tenant_id"`
	Plan          string  `json:"plan"`
	BillingCycle  string  `json:"billing_cycle"`
	Status        string  `json:"status"`
	DerivedStatus string  `json:"derived_status"`
	Currency      string  `json:"currency"`
	AmountCents   uint32  `json:"amount_cents"`
	PeriodStart   *string `json:"period_start,omitempty"`
	PeriodEnd     *string `json:"period_end,omitempty"`
	TrialEndsAt   *string `json:"trial_ends_at,omitempty"`
	NextRenewalAt *string `json:"next_renewal_at,omitempty"`
	CanceledAt    *string `json:"canceled_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type SubscriptionActivateRequest struct {
	BillingCycle string `json:"billing_cycle" binding:"required,oneof=MONTHLY ANNUAL"`
	AmountCents  uint32 `json:"amount_cents,omitempty"`
}
