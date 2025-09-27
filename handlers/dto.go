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
}

type PlatformBranding struct {
	Name         string `json:"name"`
	LogoURL      string `json:"logo_url"`
	PrimaryColor string `json:"primary_color"`
	AccentColor  string `json:"accent_color"`
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

type TenantCreateRequest struct {
	Name       string  `json:"name" binding:"required,min=2"`
	Domain     *string `json:"domain" binding:"omitempty,min=3"`
	AdminName  string  `json:"admin_name" binding:"required,min=2"`
	AdminEmail string  `json:"admin_email" binding:"required,email"`
	Plan       *string `json:"plan" binding:"omitempty,oneof=BASIC STARTER GROWTH ENTERPRISE ON_PREM"`
}

type TenantSummary struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Domain *string `json:"domain,omitempty"`
	Slug   *string `json:"slug,omitempty"`
}

type TenantCreateResponse struct {
	Tenant       TenantSummary `json:"tenant"`
	AdminUserID  string        `json:"admin_user_id"`
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

type TenantStatusResponse struct {
	Status string   `json:"status"`
	Steps  []string `json:"steps,omitempty"`
}
