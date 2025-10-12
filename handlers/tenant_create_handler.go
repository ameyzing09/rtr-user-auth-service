package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
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
		IsTrial:    req.IsTrial,
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
	// Parse pagination parameters
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := c.Query("size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	result, err := h.service.ListTenants(c.Request.Context(), page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	items := make([]TenantListItem, 0, len(result.Tenants))
	for _, tenant := range result.Tenants {
		items = append(items, mapTenantDTOToListItem(tenant))
	}

	response := TenantListPaginatedResponse{
		Tenants:  items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}

	c.JSON(http.StatusOK, response)
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

	response := mapTenantDTOToGetResponse(tenant)
	c.JSON(http.StatusOK, response)
}

func (h *TenantCreateHandler) Update(c *gin.Context) {
	actor := ActorFromContext(c)

	tenantID := strings.TrimSpace(c.Param("id"))
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": errcodes.ErrCodeValidation, "message": "tenant id is required"})
		return
	}

	var req TenantUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	// Convert string pointers to model types
	var planPtr *models.Plan
	if req.Plan != nil {
		plan := models.Plan(*req.Plan)
		planPtr = &plan
	}

	var statusPtr *models.TenantStatus
	if req.Status != nil {
		status := models.TenantStatus(*req.Status)
		statusPtr = &status
	}

	serviceReq := services.UpdateTenantReq{
		Name:   req.Name,
		Domain: req.Domain,
		Plan:   planPtr,
		Status: statusPtr,
	}

	tenant, err := h.service.UpdateTenant(c.Request.Context(), tenantID, serviceReq, actor.ID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := mapTenantDTOToGetResponse(tenant)
	c.JSON(http.StatusOK, response)
}

func (h *TenantCreateHandler) Delete(c *gin.Context) {
	actor := ActorFromContext(c)

	tenantID := strings.TrimSpace(c.Param("id"))
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": errcodes.ErrCodeValidation, "message": "tenant id is required"})
		return
	}

	if err := h.service.DeleteTenant(c.Request.Context(), tenantID, actor.ID); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
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

func mapTenantDTOToListItem(tenant services.TenantDTO) TenantListItem {
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

func mapTenantDTOToGetResponse(tenant services.TenantDTO) TenantGetResponse {
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

// Archive endpoints
func (h *TenantCreateHandler) ListArchived(c *gin.Context) {
	_ = ActorFromContext(c)

	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	result, err := h.service.ListArchivedTenants(c.Request.Context(), page, pageSize)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := make([]map[string]interface{}, len(result.Archives))
	for i, archive := range result.Archives {
		response[i] = map[string]interface{}{
			"id":            archive.ID,
			"name":          archive.Name,
			"domain":        archive.Domain,
			"slug":          archive.Slug,
			"plan":          planToStringPointer(archive.Plan),
			"status":        string(archive.Status),
			"created_by":    archive.CreatedBy,
			"created_at":    archive.CreatedAt.UTC().Format(time.RFC3339),
			"updated_at":    archive.UpdatedAt.UTC().Format(time.RFC3339),
			"failed_reason": archive.FailedReason,
			"deleted_by":    archive.DeletedBy,
			"deleted_at":    archive.DeletedAt.UTC().Format(time.RFC3339),
			"delete_reason": archive.DeleteReason,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"archives":  response,
		"total":     result.Total,
		"page":      result.Page,
		"page_size": result.PageSize,
	})
}

func (h *TenantCreateHandler) GetArchived(c *gin.Context) {
	_ = ActorFromContext(c)

	tenantID := strings.TrimSpace(c.Param("id"))
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": errcodes.ErrCodeValidation, "message": "tenant id is required"})
		return
	}

	archive, err := h.service.GetArchivedTenant(c.Request.Context(), tenantID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	response := map[string]interface{}{
		"id":            archive.ID,
		"name":          archive.Name,
		"domain":        archive.Domain,
		"slug":          archive.Slug,
		"plan":          planToStringPointer(archive.Plan),
		"status":        string(archive.Status),
		"created_by":    archive.CreatedBy,
		"created_at":    archive.CreatedAt.UTC().Format(time.RFC3339),
		"updated_at":    archive.UpdatedAt.UTC().Format(time.RFC3339),
		"failed_reason": archive.FailedReason,
		"deleted_by":    archive.DeletedBy,
		"deleted_at":    archive.DeletedAt.UTC().Format(time.RFC3339),
		"delete_reason": archive.DeleteReason,
	}

	c.JSON(http.StatusOK, response)
}
