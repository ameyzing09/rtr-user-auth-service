package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/entities"
	"github.com/ameyzing09/rtr-user-auth-service/internal/domain/repositories"
	"github.com/google/uuid"
)

var (
	ErrTenantAlreadyExists = errors.New("tenant already exists")
	ErrDomainAlreadyExists = errors.New("domain already exists")
)

// TenantService handles tenant operations
type TenantService interface {
	Create(ctx context.Context, req *CreateTenantRequest) (*entities.Tenant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Tenant, error)
	GetByDomain(ctx context.Context, domain string) (*entities.Tenant, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateTenantRequest) (*entities.Tenant, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, req *ListTenantsRequest) (*ListTenantsResponse, error)
}

// CreateTenantRequest represents the request to create a tenant
type CreateTenantRequest struct {
	Name   string `json:"name" validate:"required,min=2,max=100"`
	Domain string `json:"domain" validate:"required,hostname"`
}

// UpdateTenantRequest represents the request to update a tenant
type UpdateTenantRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Domain   *string `json:"domain,omitempty" validate:"omitempty,hostname"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// ListTenantsRequest represents the request to list tenants
type ListTenantsRequest struct {
	Page  int `json:"page" validate:"min=1"`
	Limit int `json:"limit" validate:"min=1,max=100"`
}

// ListTenantsResponse represents the response from listing tenants
type ListTenantsResponse struct {
	Tenants    []*entities.Tenant `json:"tenants"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}

// tenantService implements TenantService interface
type tenantService struct {
	tenantRepo repositories.TenantRepository
}

// NewTenantService creates a new tenant service
func NewTenantService(tenantRepo repositories.TenantRepository) TenantService {
	return &tenantService{
		tenantRepo: tenantRepo,
	}
}

func (s *tenantService) Create(ctx context.Context, req *CreateTenantRequest) (*entities.Tenant, error) {
	// Check if tenant with domain already exists
	existingTenant, err := s.tenantRepo.GetByDomain(ctx, req.Domain)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing tenant: %w", err)
	}
	if existingTenant != nil {
		return nil, ErrDomainAlreadyExists
	}

	// Create tenant
	tenant := &entities.Tenant{
		Name:     req.Name,
		Domain:   req.Domain,
		IsActive: true,
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return tenant, nil
}

func (s *tenantService) GetByID(ctx context.Context, id uuid.UUID) (*entities.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	if tenant == nil {
		return nil, ErrTenantNotFound
	}

	return tenant, nil
}

func (s *tenantService) GetByDomain(ctx context.Context, domain string) (*entities.Tenant, error) {
	tenant, err := s.tenantRepo.GetByDomain(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	if tenant == nil {
		return nil, ErrTenantNotFound
	}

	return tenant, nil
}

func (s *tenantService) Update(ctx context.Context, id uuid.UUID, req *UpdateTenantRequest) (*entities.Tenant, error) {
	// Get existing tenant
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	if tenant == nil {
		return nil, ErrTenantNotFound
	}

	// Update fields if provided
	if req.Name != nil {
		tenant.Name = *req.Name
	}

	if req.Domain != nil {
		// Check if new domain already exists
		if *req.Domain != tenant.Domain {
			existingTenant, err := s.tenantRepo.GetByDomain(ctx, *req.Domain)
			if err != nil {
				return nil, fmt.Errorf("failed to check existing domain: %w", err)
			}
			if existingTenant != nil {
				return nil, ErrDomainAlreadyExists
			}
		}
		tenant.Domain = *req.Domain
	}

	if req.IsActive != nil {
		tenant.IsActive = *req.IsActive
	}

	// Save tenant
	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

func (s *tenantService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if tenant exists
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}
	if tenant == nil {
		return ErrTenantNotFound
	}

	return s.tenantRepo.Delete(ctx, id)
}

func (s *tenantService) List(ctx context.Context, req *ListTenantsRequest) (*ListTenantsResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	offset := (req.Page - 1) * req.Limit

	tenants, total, err := s.tenantRepo.List(ctx, req.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	totalPages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	return &ListTenantsResponse{
		Tenants:    tenants,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}