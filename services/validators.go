package services

import (
	"fmt"
	"strings"

	"rtr-user-auth-service/models"
	"rtr-user-auth-service/utils"
)

// ValidatedTenantInput contains validated and normalized tenant onboarding data
type ValidatedTenantInput struct {
	Name       string
	Slug       string
	AdminName  string
	AdminEmail string
	Domain     *string
	Plan       *models.Plan
}

// ValidateTenantOnboardInput validates and normalizes tenant onboarding input
func ValidateTenantOnboardInput(req TenantOnboardAsyncRequest) (*ValidatedTenantInput, error) {
	// Validate and normalize tenant name
	normalizedName := strings.TrimSpace(req.Name)
	if normalizedName == "" {
		return nil, fmt.Errorf("tenant name is required: %w", ErrInvalidInput)
	}

	// Validate and normalize admin name
	adminName := strings.TrimSpace(req.AdminName)
	if adminName == "" {
		return nil, fmt.Errorf("admin name is required: %w", ErrInvalidInput)
	}

	// Validate and normalize admin email
	if strings.TrimSpace(req.AdminEmail) == "" {
		return nil, fmt.Errorf("admin email is required: %w", ErrInvalidInput)
	}

	adminEmail, err := utils.NormalizeEmail(req.AdminEmail)
	if err != nil {
		return nil, fmt.Errorf("invalid admin email format: %w", ErrInvalidInput)
	}

	// Validate and normalize domain if provided
	var domainPtr *string
	if req.Domain != nil {
		normalizedDomain, err := utils.NormalizeDomain(*req.Domain)
		if err != nil {
			return nil, fmt.Errorf("invalid domain format: %w", ErrInvalidInput)
		}
		domainPtr = &normalizedDomain
	}

	// Generate and validate slug
	slug, err := utils.NormalizeSlug(normalizedName)
	if err != nil {
		return nil, fmt.Errorf("failed to generate slug from name: %w", ErrInvalidInput)
	}

	// Validate and normalize plan
	planPtr, err := normalizePlan(req.Plan)
	if err != nil {
		return nil, fmt.Errorf("invalid plan: %w", err)
	}

	return &ValidatedTenantInput{
		Name:       normalizedName,
		Slug:       slug,
		AdminName:  adminName,
		AdminEmail: adminEmail,
		Domain:     domainPtr,
		Plan:       planPtr,
	}, nil
}

// ValidateUserInput validates user creation input
func ValidateUserInput(input CreateUserInput) error {
	if strings.TrimSpace(input.Email) == "" {
		return fmt.Errorf("email is required: %w", ErrInvalidInput)
	}

	if _, err := utils.NormalizeEmail(input.Email); err != nil {
		return fmt.Errorf("invalid email format: %w", ErrInvalidInput)
	}

	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("name is required: %w", ErrInvalidInput)
	}

	if input.Role == "" {
		return fmt.Errorf("role is required: %w", ErrInvalidInput)
	}

	return nil
}

// ValidatePasswordInput validates password change input
func ValidatePasswordInput(input ChangePasswordInput) error {
	if strings.TrimSpace(input.CurrentPassword) == "" {
		return fmt.Errorf("current password is required: %w", ErrInvalidInput)
	}

	if strings.TrimSpace(input.NewPassword) == "" {
		return fmt.Errorf("new password is required: %w", ErrInvalidInput)
	}

	if len(input.NewPassword) < 6 {
		return fmt.Errorf("new password must be at least 6 characters: %w", ErrInvalidInput)
	}

	if input.CurrentPassword == input.NewPassword {
		return fmt.Errorf("new password must be different from current password: %w", ErrInvalidInput)
	}

	return nil
}
