package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"rtr-user-auth-service/models"
	"rtr-user-auth-service/repositories"
)

// OutboxPublisher reads unpublished events from the outbox and publishes them
type OutboxPublisher struct {
	repo   repositories.OutboxRepository
	broker MessageBroker
	logger Logger
	config PublisherConfig
}

// PublisherConfig contains configuration for the outbox publisher
type PublisherConfig struct {
	BatchSize       int           // Number of events to process per batch
	PollInterval    time.Duration // Time between polling for new events
	MaxRetries      int           // Maximum retries for failed publishes
	ShutdownTimeout time.Duration // Timeout for graceful shutdown
}

// DefaultPublisherConfig returns sensible defaults
func DefaultPublisherConfig() PublisherConfig {
	return PublisherConfig{
		BatchSize:       100,
		PollInterval:    5 * time.Second,
		MaxRetries:      3,
		ShutdownTimeout: 30 * time.Second,
	}
}

// NewOutboxPublisher creates a new outbox publisher instance
func NewOutboxPublisher(
	repo repositories.OutboxRepository,
	broker MessageBroker,
	logger Logger,
	config PublisherConfig,
) *OutboxPublisher {
	return &OutboxPublisher{
		repo:   repo,
		broker: broker,
		logger: logger,
		config: config,
	}
}

// Start begins the publishing loop
func (p *OutboxPublisher) Start(ctx context.Context) error {
	p.logger.Info("Outbox publisher started",
		"batchSize", p.config.BatchSize,
		"pollInterval", p.config.PollInterval,
	)

	ticker := time.NewTicker(p.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("Outbox publisher shutting down...")
			return ctx.Err()
		case <-ticker.C:
			if err := p.publishBatch(ctx); err != nil {
				p.logger.Info("Error publishing batch", "error", err)
			}
		}
	}
}

// publishBatch processes a batch of unpublished events
func (p *OutboxPublisher) publishBatch(ctx context.Context) error {
	events, err := p.repo.GetUnpublished(ctx, p.config.BatchSize)
	if err != nil {
		return fmt.Errorf("failed to fetch unpublished events: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	p.logger.Info("Publishing batch", "count", len(events))

	successCount := 0
	failureCount := 0

	for _, event := range events {
		if err := p.publishEvent(ctx, event); err != nil {
			p.logger.Info("Failed to publish event",
				"eventID", event.ID,
				"type", event.Type,
				"error", err,
			)
			// Mark as failed (or skip marking to retry later)
			_ = p.repo.MarkFailed(ctx, event.ID, err.Error())
			failureCount++
			continue
		}

		// Mark as published
		if err := p.repo.MarkPublished(ctx, event.ID); err != nil {
			p.logger.Info("Failed to mark event as published",
				"eventID", event.ID,
				"error", err,
			)
			failureCount++
			continue
		}

		successCount++
	}

	p.logger.Info("Batch published",
		"success", successCount,
		"failures", failureCount,
		"total", len(events),
	)

	return nil
}

// publishEvent publishes a single event to the message broker
func (p *OutboxPublisher) publishEvent(ctx context.Context, event models.Outbox) error {
	// Create event envelope
	envelope := EventEnvelope{
		Type:          event.Type,
		AggregateType: event.AggregateType,
		AggregateID:   event.AggregateID,
		Timestamp:     event.CreatedAt.Format(time.RFC3339),
		Payload:       json.RawMessage(event.Payload),
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("failed to marshal event envelope: %w", err)
	}

	// Publish to broker (topic = aggregate type, key = aggregate ID for partitioning)
	topic := event.AggregateType
	key := event.AggregateID

	if err := p.broker.Publish(ctx, topic, key, data); err != nil {
		return fmt.Errorf("failed to publish to broker: %w", err)
	}

	p.logger.Info("Event published",
		"eventID", event.ID,
		"type", event.Type,
		"aggregateID", event.AggregateID,
	)

	return nil
}
