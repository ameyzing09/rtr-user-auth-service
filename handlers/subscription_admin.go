package handlers

import (
	"net/http"
	"strings"
	"time"

	"rtr-user-auth-service/models"
	"rtr-user-auth-service/services"
	"rtr-user-auth-service/utils/httpx"

	"github.com/gin-gonic/gin"
)

type SubscriptionAdminHandler struct {
	subscriptionSvc services.SubscriptionService
}

func NewSubscriptionAdminHandler(subscriptionSvc services.SubscriptionService) *SubscriptionAdminHandler {
	return &SubscriptionAdminHandler{
		subscriptionSvc: subscriptionSvc,
	}
}

func (h *SubscriptionAdminHandler) Get(c *gin.Context) {
	tenantID := strings.TrimSpace(c.Param("id"))
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "validation_error", "message": "tenant id is required"})
		return
	}

	sub, err := h.subscriptionSvc.GetSubscription(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "failed to get subscription"})
		return
	}

	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": "not_found", "message": "subscription not found"})
		return
	}

	now := time.Now().UTC()
	effectiveStatus := services.EffectiveStatus(sub, now)

	response := mapSubscriptionToResponse(sub, effectiveStatus)
	c.JSON(http.StatusOK, response)
}

func (h *SubscriptionAdminHandler) Activate(c *gin.Context) {
	actor := ActorFromContext(c)

	tenantID := strings.TrimSpace(c.Param("id"))
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "validation_error", "message": "tenant id is required"})
		return
	}

	var req SubscriptionActivateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.HandleBindingError(c, err)
		return
	}

	billingCycle := models.BillingCycle(req.BillingCycle)
	if err := h.subscriptionSvc.ActivateSubscription(c.Request.Context(), tenantID, billingCycle, req.AmountCents, actor.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "failed to activate subscription"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *SubscriptionAdminHandler) Suspend(c *gin.Context) {
	actor := ActorFromContext(c)

	tenantID := strings.TrimSpace(c.Param("id"))
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "validation_error", "message": "tenant id is required"})
		return
	}

	if err := h.subscriptionSvc.SuspendSubscription(c.Request.Context(), tenantID, actor.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "failed to suspend subscription"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *SubscriptionAdminHandler) Resume(c *gin.Context) {
	actor := ActorFromContext(c)

	tenantID := strings.TrimSpace(c.Param("id"))
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "validation_error", "message": "tenant id is required"})
		return
	}

	if err := h.subscriptionSvc.ResumeSubscription(c.Request.Context(), tenantID, actor.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "failed to resume subscription"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *SubscriptionAdminHandler) Cancel(c *gin.Context) {
	actor := ActorFromContext(c)

	tenantID := strings.TrimSpace(c.Param("id"))
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "validation_error", "message": "tenant id is required"})
		return
	}

	if err := h.subscriptionSvc.CancelSubscription(c.Request.Context(), tenantID, actor.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "internal_error", "message": "failed to cancel subscription"})
		return
	}

	c.Status(http.StatusOK)
}

func mapSubscriptionToResponse(sub *models.Subscription, effectiveStatus models.SubscriptionStatus) SubscriptionResponse {
	resp := SubscriptionResponse{
		ID:            sub.ID,
		TenantID:      sub.TenantID,
		Plan:          string(sub.Plan),
		BillingCycle:  string(sub.BillingCycle),
		Status:        string(sub.Status),
		DerivedStatus: string(effectiveStatus),
		Currency:      sub.Currency,
		AmountCents:   sub.AmountCents,
		CreatedAt:     sub.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:     sub.UpdatedAt.UTC().Format(time.RFC3339),
	}

	if sub.PeriodStart != nil {
		periodStart := sub.PeriodStart.UTC().Format(time.RFC3339)
		resp.PeriodStart = &periodStart
	}

	if sub.PeriodEnd != nil {
		periodEnd := sub.PeriodEnd.UTC().Format(time.RFC3339)
		resp.PeriodEnd = &periodEnd
	}

	if sub.TrialEndsAt != nil {
		trialEndsAt := sub.TrialEndsAt.UTC().Format(time.RFC3339)
		resp.TrialEndsAt = &trialEndsAt
	}

	if sub.NextRenewalAt != nil {
		nextRenewalAt := sub.NextRenewalAt.UTC().Format(time.RFC3339)
		resp.NextRenewalAt = &nextRenewalAt
	}

	if sub.CanceledAt != nil {
		canceledAt := sub.CanceledAt.UTC().Format(time.RFC3339)
		resp.CanceledAt = &canceledAt
	}

	return resp
}
