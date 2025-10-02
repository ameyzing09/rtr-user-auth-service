package consumers

import (
	"context"
	"encoding/json"
	"fmt"

	"rtr-user-auth-service/services"
)

// TenantCreatedEvent represents the tenant.created event payload
type TenantCreatedEvent struct {
	V             int    `json:"v"`
	TenantID      string `json:"tenantId"`
	Name          string `json:"name"`
	Domain        string `json:"domain,omitempty"`
	Plan          string `json:"plan"`
	CreatorUserID string `json:"creatorUserId"`
	CreatedAt     string `json:"createdAt"`
}

// TenantProvisionedEvent represents the tenant.provisioned event payload
type TenantProvisionedEvent struct {
	V        int    `json:"v"`
	TenantID string `json:"tenantId"`
	Status   string `json:"status"` // "active" or "failed"
	Reason   string `json:"reason,omitempty"`
}

// TenantEventConsumer handles tenant-related events
type TenantEventConsumer struct {
	provisioningSvc *services.TenantProvisioningService
	logger          services.Logger
}

// NewTenantEventConsumer creates a new tenant event consumer
func NewTenantEventConsumer(
	provisioningSvc *services.TenantProvisioningService,
	logger services.Logger,
) *TenantEventConsumer {
	return &TenantEventConsumer{
		provisioningSvc: provisioningSvc,
		logger:          logger,
	}
}

// HandleTenantCreated processes tenant.created events and triggers provisioning
func (c *TenantEventConsumer) HandleTenantCreated(ctx context.Context, message []byte) error {
	var event TenantCreatedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		c.logger.Error("Failed to unmarshal tenant.created event", "error", err)
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	c.logger.Info("Received tenant.created event",
		"tenantID", event.TenantID,
		"name", event.Name,
	)

	// Trigger provisioning
	if err := c.provisioningSvc.ProvisionTenant(ctx, event.TenantID); err != nil {
		c.logger.Error("Tenant provisioning failed",
			"tenantID", event.TenantID,
			"error", err,
		)
		return fmt.Errorf("provisioning failed: %w", err)
	}

	c.logger.Info("Tenant provisioning completed",
		"tenantID", event.TenantID,
	)

	return nil
}

// HandleTenantProvisioned processes tenant.provisioned events (if using external provisioner)
func (c *TenantEventConsumer) HandleTenantProvisioned(ctx context.Context, message []byte) error {
	var event TenantProvisionedEvent
	if err := json.Unmarshal(message, &event); err != nil {
		c.logger.Error("Failed to unmarshal tenant.provisioned event", "error", err)
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	c.logger.Info("Received tenant.provisioned event",
		"tenantID", event.TenantID,
		"status", event.Status,
	)

	// This handler would be used if you have a separate provisioning service
	// For now, we handle provisioning synchronously after tenant.created
	c.logger.Info("Tenant provisioned event processed",
		"tenantID", event.TenantID,
	)

	return nil
}
