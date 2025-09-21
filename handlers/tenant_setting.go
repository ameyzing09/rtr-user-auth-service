package handlers

import (
	"net/http"
	"rtr-user-auth-service/middleware"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils/httpx"

	"github.com/gin-gonic/gin"
)

type TenantSettingHandler struct {
	tenantSettingService services.TenantSettingService
}

func NewTenantSettingHandler(tss services.TenantSettingService) *TenantSettingHandler {
	return &TenantSettingHandler{tenantSettingService: tss}
}

type tenantSettingsPayload struct {
	Config map[string]interface{} `json:"config" binding:"required"`
}

func (h *TenantSettingHandler) Get(c *gin.Context) {
	tenantID := c.GetString(middleware.CtxTenantIDKey)
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant not resolved"})
		return
	}

	cfg, err := h.tenantSettingService.Get(c, tenantID)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	if cfg == nil {
		cfg = map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

func (h *TenantSettingHandler) Put(c *gin.Context) {
	tenantID := c.GetString(middleware.CtxTenantIDKey)
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenant not resolved"})
		return
	}

	var payload tenantSettingsPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	cfg, err := h.tenantSettingService.PutReplace(c, tenantID, payload.Config)
	if err != nil {
		httpx.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": cfg})
}
