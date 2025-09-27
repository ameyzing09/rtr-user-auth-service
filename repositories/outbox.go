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
