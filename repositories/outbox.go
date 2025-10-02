package repositories

import (
	"context"
	"encoding/json"
	"rtr-user-auth-service/models"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type OutboxRepository interface {
	Append(ctx context.Context, aggregateType, aggregateID, eventType string, payload map[string]interface{}) error
	GetUnpublished(ctx context.Context, limit int) ([]models.Outbox, error)
	MarkPublished(ctx context.Context, id uint64) error
	MarkFailed(ctx context.Context, id uint64, errMsg string) error
}

type GormOutboxRepo struct {
	db *gorm.DB
}

func NewGormOutboxRepo(db *gorm.DB) *GormOutboxRepo {
	return &GormOutboxRepo{db: db}
}

func (r *GormOutboxRepo) Append(ctx context.Context, aggregateType, aggregateID, eventType string, payload map[string]interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	record := &models.Outbox{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Type:          eventType,
		Payload:       datatypes.JSON(body),
	}

	return r.db.WithContext(ctx).Create(record).Error
}

// GetUnpublished retrieves unpublished events from the outbox table
func (r *GormOutboxRepo) GetUnpublished(ctx context.Context, limit int) ([]models.Outbox, error) {
	var events []models.Outbox
	err := r.db.WithContext(ctx).
		Where("published_at IS NULL").
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

// MarkPublished marks an event as published with the current timestamp
func (r *GormOutboxRepo) MarkPublished(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&models.Outbox{}).
		Where("id = ?", id).
		Update("published_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}

// MarkFailed marks an event as failed (for future retry logic if needed)
func (r *GormOutboxRepo) MarkFailed(ctx context.Context, id uint64, errMsg string) error {
	// For now, we'll just mark it as published to avoid infinite retries
	// In production, you might want a separate failed_at column and retry logic
	return r.MarkPublished(ctx, id)
}
