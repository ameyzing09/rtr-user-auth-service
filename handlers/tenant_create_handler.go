package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"rtr-user-auth-service/domain"
	errcodes "rtr-user-auth-service/errors"
	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils"
	"rtr-user-auth-service/utils/httpx"

	"github.com/gin-gonic/gin"
)

type TenantCreateHandler struct {
	service services.TenantService
}

func NewTenantCreateHandler(service services.TenantService) *TenantCreateHandler {
	return &TenantCreateHandler{service: service}
}

func (h *TenantCreateHandler) Create(c *gin.Context) {
	actor := ActorFromContext(c)

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": errcodes.ErrCodeValidation, "message": "unable to read request body"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	var req TenantCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	canonicalBody, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": errcodes.ErrCodeInternal, "message": "failed to marshal request"})
		return
	}

	idempotencyKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if idempotencyKey == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": errcodes.ErrCodeValidation, "message": "Idempotency-Key header is required"})
		return
	}

	keyHash := utils.HashKey(idempotencyKey)
	requestHash := utils.HashRequest(canonicalBody)

	serviceReq := services.TenantOnboardAsyncRequest{
		Name:       req.Name,
		Domain:     req.Domain,
		AdminName:  req.AdminName,
		AdminEmail: req.AdminEmail,
		Plan:       PlanPointer(req.Plan),
	}

	result, cached, err := h.service.OnboardTenantAsync(c.Request.Context(), actor, serviceReq, keyHash, requestHash)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := TenantCreateResponse{
		Tenant: TenantSummary{
			ID:     result.TenantID,
			Name:   result.Name,
			Domain: result.Domain,
			Slug:   result.Slug,
		},
		TempPassword: result.TempPassword,
		Status:       string(result.Status),
	}

	status := http.StatusCreated
	if cached {
		status = http.StatusOK
	}

	c.JSON(status, response)
}

func (h *TenantCreateHandler) List(c *gin.Context) {
	actor := ActorFromContext(c)

	tenants, err := h.service.ListTenants(c.Request.Context(), actor)
	if err != nil {
		h.handleError(c, err)
		return
	}

	items := make([]TenantListItem, 0, len(tenants))
	for _, tenant := range tenants {
		items = append(items, mapTenantToListItem(&tenant))
	}

	c.JSON(http.StatusOK, TenantListResponse{Tenants: items})
}

func (h *TenantCreateHandler) Get(c *gin.Context) {
	_ = ActorFromContext(c)

	tenantID := strings.TrimSpace(c.Param("id"))
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": errcodes.ErrCodeValidation, "message": "tenant id is required"})
		return
	}

	tenant, err := h.service.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := mapTenantToGetResponse(tenant)
	c.JSON(http.StatusOK, response)
}

func (h *TenantCreateHandler) Status(c *gin.Context) {
	_ = ActorFromContext(c)

	tenantID := strings.TrimSpace(c.Param("id"))
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": errcodes.ErrCodeValidation, "message": "tenant id is required"})
		return
	}

	statusView, err := h.service.GetTenantStatus(c.Request.Context(), tenantID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := TenantStatusResponse{
		Status: string(statusView.Status),
		Steps:  statusView.Steps,
	}
	c.JSON(http.StatusOK, response)
}

func (h *TenantCreateHandler) Retry(c *gin.Context) {
	actor := ActorFromContext(c)

	tenantID := strings.TrimSpace(c.Param("id"))
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": errcodes.ErrCodeValidation, "message": "tenant id is required"})
		return
	}

	if err := h.service.RetryProvisioning(c.Request.Context(), actor, tenantID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusAccepted)
}

func (h *TenantCreateHandler) handleError(c *gin.Context, err error) {
	var suggestionProvider interface {
		Suggestions() []string
	}

	if errors.As(err, &suggestionProvider) {
		status, code := utils.ResolveHTTPError(domain.ErrTenantSlugTaken)
		c.JSON(status, gin.H{
			"code":        code,
			"message":     err.Error(),
			"suggestions": suggestionProvider.Suggestions(),
		})
		return
	}

	status, code := utils.ResolveHTTPError(err)
	c.JSON(status, gin.H{
		"code":    code,
		"message": err.Error(),
	})
}

func mapTenantToListItem(tenant *models.Tenant) TenantListItem {
	return TenantListItem{
		ID:           tenant.ID,
		Name:         tenant.Name,
		Domain:       StringPointer(tenant.Domain),
		Slug:         StringPointer(tenant.Slug),
		Plan:         planToStringPointer(tenant.Plan),
		Status:       string(tenant.Status),
		CreatedBy:    tenant.CreatedBy,
		CreatedAt:    tenant.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    tenant.UpdatedAt.UTC().Format(time.RFC3339),
		FailedReason: tenant.FailedReason,
	}
}

func mapTenantToGetResponse(tenant *models.Tenant) TenantGetResponse {
	resp := TenantGetResponse{
		ID:           tenant.ID,
		Name:         tenant.Name,
		Domain:       StringPointer(tenant.Domain),
		Slug:         StringPointer(tenant.Slug),
		Plan:         planToStringPointer(tenant.Plan),
		Status:       string(tenant.Status),
		CreatedAt:    tenant.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:    tenant.UpdatedAt.UTC().Format(time.RFC3339),
		FailedReason: tenant.FailedReason,
	}
	if tenant.CreatedBy != nil {
		resp.CreatedBy = tenant.CreatedBy
	}
	return resp
}

func planToStringPointer(plan *models.Plan) *string {
	if plan == nil {
		return nil
	}
	str := string(*plan)
	return &str
}
