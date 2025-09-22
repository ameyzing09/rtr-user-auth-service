package handlers

import "rtr-user-auth-service/models"

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

type CreateUserRequest struct {
	Email string      `json:"email" binding:"required,email"`
	Name  string      `json:"name" binding:"required,min=2"`
	Role  models.Role `json:"role" binding:"required,oneof=admin hr interviewer candidate"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=6"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

type TenantOnboardRequest struct {
	Name       string `json:"name" binding:"required,min=2"`
	Domain     string `json:"domain" binding:"required,min=2"`
	AdminName  string `json:"admin_name" binding:"required,min=2"`
	AdminEmail string `json:"admin_email" binding:"required,email"`
	Plan       string `json:"plan" binding:"required"`
}

type TenantResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

type TenantOnboardResponse struct {
	Tenant       TenantResponse `json:"tenant"`
	AdminUserID  string         `json:"admin_user_id"`
	TempPassword string         `json:"temp_password"`
}
