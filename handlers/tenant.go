package handlers

import (
	"net/http"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils/httpx"

	"github.com/gin-gonic/gin"
)

type TenantHandler struct {
	tenantService services.TenantService
}

func NewTenantHandler(tenantService services.TenantService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

// requireSuperAdmin checks if the authenticated user has SUPERADMIN role
func (h *TenantHandler) requireSuperAdmin(c *gin.Context) bool {
	actor, exists := c.Get("actor")
	if !exists || actor.(services.UserRead).Role != models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "SUPERADMIN access required"})
		return false
	}
	return true
}

// OnboardTenant creates a new tenant with an admin user (SUPERADMIN only)
// POST /admin/tenants/onboard
func (h *TenantHandler) OnboardTenant(c *gin.Context) {
	if !h.requireSuperAdmin(c) {
		return
	}

	var req TenantOnboardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	result, err := h.tenantService.Onboard(c, services.TenantOnboardRequest{
		Name:       req.Name,
		Domain:     req.Domain,
		AdminName:  req.AdminName,
		AdminEmail: req.AdminEmail,
		Plan:       req.Plan,
	})
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	response := TenantOnboardResponse{
		Tenant: TenantResponse{
			ID:     result.TenantID,
			Name:   req.Name,
			Domain: result.Domain,
		},
		AdminUserID:  result.AdminUserID,
		TempPassword: result.TempPassword,
	}

	c.JSON(http.StatusCreated, response)
}

// GetTenant retrieves a tenant by ID (SUPERADMIN only)
// GET /admin/tenants/:id
func (h *TenantHandler) GetTenant(c *gin.Context) {
	if !h.requireSuperAdmin(c) {
		return
	}

	tenantID := c.Param("id")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant ID is required"})
		return
	}

	tenant, err := h.tenantService.GetTenant(c, tenantID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	response := TenantResponse{
		ID:     tenant.ID,
		Name:   tenant.Name,
		Domain: tenant.Domain,
	}

	c.JSON(http.StatusOK, response)
}

// GetTenantByDomain retrieves a tenant by domain (SUPERADMIN only)
// GET /admin/tenants/domain/:domain
func (h *TenantHandler) GetTenantByDomain(c *gin.Context) {
	if !h.requireSuperAdmin(c) {
		return
	}

	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain is required"})
		return
	}

	tenant, err := h.tenantService.GetTenantByDomain(c, domain)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	response := TenantResponse{
		ID:     tenant.ID,
		Name:   tenant.Name,
		Domain: tenant.Domain,
	}

	c.JSON(http.StatusOK, response)
}
