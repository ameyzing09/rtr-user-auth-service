package services

import (
	"context"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/utils"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TenantOnboardRequest struct {
	Name       string
	Domain     string
	AdminName  string
	AdminEmail string
	Plan       string
}

type TenantOnboardResult struct {
	TenantID     string
	Domain       string
	AdminUserID  string
	TempPassword string // don't store this, just return it once
}

var _ TenantService = (*tenantService)(nil) // suggested by claude sonnet 4 for early detection of interface implementation errors

type tenantService struct {
	db             *gorm.DB
	tenants        TenantRepository
	users          UserRepository
	tenantSettings TenantSettingRepository
}

func NewTenantService(db *gorm.DB, tr TenantRepository, ur UserRepository, tsr TenantSettingRepository) *tenantService {
	return &tenantService{
		db:             db,
		tenants:        tr,
		users:          ur,
		tenantSettings: tsr,
	}
}

func (s *tenantService) Onboard(ctx context.Context, req TenantOnboardRequest) (TenantOnboardResult, error) {
	// Validate input
	if err := s.validateOnboardRequest(req); err != nil {
		return TenantOnboardResult{}, err
	}

	// Normalize inputs
	domain := strings.ToLower(strings.TrimSpace(req.Domain))
	adminEmail := strings.ToLower(strings.TrimSpace(req.AdminEmail))

	// Check if tenant already exists
	exists, err := s.tenants.Exists(ctx, domain)
	if err != nil {
		return TenantOnboardResult{}, err
	}
	if exists {
		return TenantOnboardResult{}, ErrTenantAlreadyExists
	}

	// Perform onboarding in a transaction
	var result TenantOnboardResult
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		tenantID := uuid.NewString()

		// Create tenant
		tenant := &models.Tenant{
			ID:     tenantID,
			Name:   strings.TrimSpace(req.Name),
			Domain: domain,
		}
		if err := tx.Create(tenant).Error; err != nil {
			return err
		}

		// Generate credentials for admin user
		adminUserID := uuid.NewString()
		tempPassword, err := utils.GenerateTempPassword()
		if err != nil {
			return err
		}

		hashedPassword, err := utils.HashPassword(tempPassword)
		if err != nil {
			return err
		}

		// Create admin user
		adminUser := &models.User{
			ID:                  adminUserID,
			TenantID:            tenantID,
			Email:               adminEmail,
			Name:                strings.TrimSpace(req.AdminName),
			Role:                models.RoleAdmin,
			Password:            hashedPassword,
			ForcePasswordChange: true,
		}
		if err := s.users.Create(ctx, adminUser); err != nil {
			return err
		}

		// Initialize tenant settings
		tenantSetting := &models.TenantSetting{
			TenantID: tenantID,
			Config:   models.JSONMap{"plan": req.Plan},
		}
		if err := s.tenantSettings.PutReplace(ctx, tenantSetting); err != nil {
			return err
		}

		// Set result for return
		result = TenantOnboardResult{
			TenantID:     tenantID,
			Domain:       domain,
			AdminUserID:  adminUserID,
			TempPassword: tempPassword,
		}
		return nil
	})

	if err != nil {
		return TenantOnboardResult{}, err
	}

	return result, nil
}

func (s *tenantService) GetTenant(ctx context.Context, tenantID string) (*models.Tenant, error) {
	if tenantID == "" {
		return nil, ErrInvalidInput
	}
	return s.tenants.FindByID(ctx, tenantID)
}

func (s *tenantService) GetTenantByDomain(ctx context.Context, domain string) (*models.Tenant, error) {
	if domain == "" {
		return nil, ErrInvalidInput
	}
	normalizedDomain := strings.ToLower(strings.TrimSpace(domain))
	return s.tenants.FindByDomain(ctx, normalizedDomain)
}

func (s *tenantService) validateOnboardRequest(req TenantOnboardRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(req.Domain) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(req.AdminName) == "" {
		return ErrInvalidInput
	}
	if strings.TrimSpace(req.AdminEmail) == "" {
		return ErrInvalidInput
	}
	return nil
}
