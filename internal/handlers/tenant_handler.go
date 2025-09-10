package handlers

import (
	"net/http"
	"strconv"

	"github.com/ameyzing09/rtr-user-auth-service/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TenantHandler handles tenant-related endpoints
type TenantHandler struct {
	tenantService services.TenantService
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService services.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// CreateTenantRequest represents the request to create a tenant
type CreateTenantRequest struct {
	Name   string `json:"name" binding:"required,min=2,max=100"`
	Domain string `json:"domain" binding:"required"`
}

// UpdateTenantRequest represents the request to update a tenant
type UpdateTenantRequest struct {
	Name     *string `json:"name,omitempty"`
	Domain   *string `json:"domain,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// CreateTenant godoc
// @Summary Create a new tenant
// @Description Create a new tenant (Admin only)
// @Tags tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTenantRequest true "Tenant details"
// @Success 201 {object} entities.Tenant
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tenants [post]
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	serviceReq := &services.CreateTenantRequest{
		Name:   req.Name,
		Domain: req.Domain,
	}

	tenant, err := h.tenantService.Create(c.Request.Context(), serviceReq)
	if err != nil {
		switch err {
		case services.ErrDomainAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{
				"error":   "domain_already_exists",
				"message": "Domain already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to create tenant",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, tenant)
}

// GetTenant godoc
// @Summary Get tenant by ID
// @Description Get tenant details by ID
// @Tags tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantId path string true "Tenant ID"
// @Success 200 {object} entities.Tenant
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tenants/{tenantId} [get]
func (h *TenantHandler) GetTenant(c *gin.Context) {
	tenantIDParam := c.Param("tenantId")
	tenantID, err := uuid.Parse(tenantIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid tenant ID format",
		})
		return
	}

	tenant, err := h.tenantService.GetByID(c.Request.Context(), tenantID)
	if err != nil {
		switch err {
		case services.ErrTenantNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "tenant_not_found",
				"message": "Tenant not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to get tenant",
			})
		}
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// GetTenantByDomain godoc
// @Summary Get tenant by domain
// @Description Get tenant details by domain
// @Tags tenants
// @Accept json
// @Produce json
// @Param domain query string true "Domain name"
// @Success 200 {object} entities.Tenant
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tenants/by-domain [get]
func (h *TenantHandler) GetTenantByDomain(c *gin.Context) {
	domain := c.Query("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Domain parameter is required",
		})
		return
	}

	tenant, err := h.tenantService.GetByDomain(c.Request.Context(), domain)
	if err != nil {
		switch err {
		case services.ErrTenantNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "tenant_not_found",
				"message": "Tenant not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to get tenant",
			})
		}
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// UpdateTenant godoc
// @Summary Update tenant
// @Description Update tenant details (Admin only)
// @Tags tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantId path string true "Tenant ID"
// @Param request body UpdateTenantRequest true "Updated tenant details"
// @Success 200 {object} entities.Tenant
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tenants/{tenantId} [put]
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	tenantIDParam := c.Param("tenantId")
	tenantID, err := uuid.Parse(tenantIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid tenant ID format",
		})
		return
	}

	var req UpdateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	serviceReq := &services.UpdateTenantRequest{
		Name:     req.Name,
		Domain:   req.Domain,
		IsActive: req.IsActive,
	}

	tenant, err := h.tenantService.Update(c.Request.Context(), tenantID, serviceReq)
	if err != nil {
		switch err {
		case services.ErrTenantNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "tenant_not_found",
				"message": "Tenant not found",
			})
		case services.ErrDomainAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{
				"error":   "domain_already_exists",
				"message": "Domain already exists",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to update tenant",
			})
		}
		return
	}

	c.JSON(http.StatusOK, tenant)
}

// DeleteTenant godoc
// @Summary Delete tenant
// @Description Delete tenant (Admin only)
// @Tags tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tenantId path string true "Tenant ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tenants/{tenantId} [delete]
func (h *TenantHandler) DeleteTenant(c *gin.Context) {
	tenantIDParam := c.Param("tenantId")
	tenantID, err := uuid.Parse(tenantIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid tenant ID format",
		})
		return
	}

	err = h.tenantService.Delete(c.Request.Context(), tenantID)
	if err != nil {
		switch err {
		case services.ErrTenantNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "tenant_not_found",
				"message": "Tenant not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to delete tenant",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant deleted successfully",
	})
}

// ListTenants godoc
// @Summary List tenants
// @Description List all tenants with pagination (Admin only)
// @Tags tenants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} services.ListTenantsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/tenants [get]
func (h *TenantHandler) ListTenants(c *gin.Context) {
	// Parse query parameters
	page := 1
	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	req := &services.ListTenantsRequest{
		Page:  page,
		Limit: limit,
	}

	response, err := h.tenantService.List(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to list tenants",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}