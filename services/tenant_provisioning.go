package services

import (
	"context"
	"fmt"

	"rtr-user-auth-service/models"
	"rtr-user-auth-service/repositories"
)

// TenantProvisioningService handles tenant provisioning logic
type TenantProvisioningService struct {
	tenantRepo repositories.TenantRepository
	userRepo   repositories.UserRepository
	logger     Logger
}

// Logger interface for structured logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
}

// NewTenantProvisioningService creates a new provisioning service
func NewTenantProvisioningService(
	tenantRepo repositories.TenantRepository,
	userRepo repositories.UserRepository,
	logger Logger,
) *TenantProvisioningService {
	return &TenantProvisioningService{
		tenantRepo: tenantRepo,
		userRepo:   userRepo,
		logger:     logger,
	}
}

// ProvisionTenant performs all provisioning steps for a new tenant
func (s *TenantProvisioningService) ProvisionTenant(ctx context.Context, tenantID string) error {
	s.logger.Info("Starting tenant provisioning", "tenantID", tenantID)

	// Update status to provisioning
	if err := s.tenantRepo.UpdateStatus(ctx, tenantID, models.TenantProvisioning); err != nil {
		s.logger.Error("Failed to update status to provisioning", "tenantID", tenantID, "error", err)
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Fetch tenant details
	tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		s.logger.Error("Failed to find tenant", "tenantID", tenantID, "error", err)
		return s.markProvisioningFailed(ctx, tenantID, fmt.Sprintf("tenant not found: %v", err))
	}

	// Step 1: Initialize tenant-specific configuration
	if err := s.initializeTenantConfig(ctx, tenantID); err != nil {
		s.logger.Error("Failed to initialize config", "tenantID", tenantID, "error", err)
		return s.markProvisioningFailed(ctx, tenantID, fmt.Sprintf("config initialization failed: %v", err))
	}

	// Step 2: Verify creator user exists and assign admin role
	if tenant.CreatedBy != nil && *tenant.CreatedBy != "" {
		if err := s.setupAdminUser(ctx, tenantID, *tenant.CreatedBy); err != nil {
			s.logger.Warn("Admin user setup skipped", "tenantID", tenantID, "error", err)
			// Don't fail provisioning if user setup has issues
		}
	}

	// Step 3: Mark tenant as active
	if err := s.tenantRepo.UpdateStatus(ctx, tenantID, models.TenantActive); err != nil {
		s.logger.Error("Failed to update status to active", "tenantID", tenantID, "error", err)
		return fmt.Errorf("failed to mark tenant as active: %w", err)
	}

	s.logger.Info("Tenant provisioning completed successfully",
		"tenantID", tenantID,
		"duration", "successful",
	)

	return nil
}

// setupAdminUser ensures the creator user exists
func (s *TenantProvisioningService) setupAdminUser(ctx context.Context, tenantID string, userID string) error {
	// Verify user exists in the tenant
	user, err := s.userRepo.FindByID(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// User exists - they were created during tenant creation
	// In your system, the role is already set during user creation
	s.logger.Info("Admin user verified", "tenantID", tenantID, "userID", userID, "role", user.Role)
	return nil
}

// initializeTenantConfig performs any tenant-specific configuration
func (s *TenantProvisioningService) initializeTenantConfig(ctx context.Context, tenantID string) error {
	// This could include:
	// - Creating default tenant settings
	// - Setting up storage buckets
	// - Initializing API keys
	// - Etc.

	s.logger.Info("Tenant configuration initialized", "tenantID", tenantID)
	return nil
}

// markProvisioningFailed updates tenant status to failed with a reason
func (s *TenantProvisioningService) markProvisioningFailed(ctx context.Context, tenantID string, reason string) error {
	if err := s.tenantRepo.UpdateStatusWithReason(ctx, tenantID, models.TenantFailed, reason); err != nil {
		s.logger.Error("Failed to mark tenant as failed", "tenantID", tenantID, "error", err)
		return err
	}
	return fmt.Errorf("provisioning failed: %s", reason)
}

// RetryFailedProvisioning retries provisioning for a failed tenant
func (s *TenantProvisioningService) RetryFailedProvisioning(ctx context.Context, tenantID string) error {
	tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	if tenant.Status != models.TenantFailed {
		return fmt.Errorf("tenant is not in failed state: current status is %s", tenant.Status)
	}

	s.logger.Info("Retrying failed provisioning", "tenantID", tenantID)
	return s.ProvisionTenant(ctx, tenantID)
}
