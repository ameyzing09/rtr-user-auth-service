package eventbus

import (
	"context"
	"time"

	"rtr-user-auth-service/repositories"
)

type EventBus interface {
	PublishTenantCreated(ctx context.Context, payload TenantCreatedV1) error
}

type TenantCreatedV1 struct {
	V             int    `json:"v"`
	TenantID      string `json:"tenantId"`
	Name          string `json:"name"`
	Domain        string `json:"domain,omitempty"`
	Plan          string `json:"plan"`
	CreatorUserID string `json:"creatorUserId"`
	CreatedAt     string `json:"createdAt"`
}

type outboxBus struct {
	outbox repositories.OutboxRepository
}

func NewOutboxBus(repo repositories.OutboxRepository) EventBus {
	return &outboxBus{outbox: repo}
}

func (b *outboxBus) PublishTenantCreated(ctx context.Context, payload TenantCreatedV1) error {
	data := map[string]interface{}{
		"v":             payload.V,
		"tenantId":      payload.TenantID,
		"name":          payload.Name,
		"plan":          payload.Plan,
		"creatorUserId": payload.CreatorUserID,
		"createdAt":     payload.CreatedAt,
	}

	if payload.Domain != "" {
		data["domain"] = payload.Domain
	}
	if payload.CreatedAt == "" {
		data["createdAt"] = time.Now().UTC().Format(time.RFC3339)
	}

	return b.outbox.Append(ctx, "tenant", payload.TenantID, "tenant.created", data)
}
