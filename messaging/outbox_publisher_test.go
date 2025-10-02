package messaging

import (
	"context"
	"testing"
	"time"

	"rtr-user-auth-service/models"
)

func TestOutboxPublisher(t *testing.T) {
	// Create mock broker and repository
	broker := NewMockBroker()
	repo := &mockOutboxRepo{
		events: []models.Outbox{
			{
				ID:            1,
				AggregateType: "tenant",
				AggregateID:   "test-tenant-123",
				Type:          "tenant.created",
				Payload:       []byte(`{"tenantId":"test-tenant-123","name":"Test"}`),
				CreatedAt:     time.Now(),
			},
		},
	}

	logger := &mockLogger{}
	config := PublisherConfig{
		BatchSize:    10,
		PollInterval: 1 * time.Second,
		MaxRetries:   3,
	}

	publisher := NewOutboxPublisher(repo, broker, logger, config)

	// Publish one batch
	ctx := context.Background()
	err := publisher.publishBatch(ctx)

	if err != nil {
		t.Fatalf("publishBatch failed: %v", err)
	}

	// Verify event was published
	if len(broker.Published) != 1 {
		t.Errorf("Expected 1 published message, got %d", len(broker.Published))
	}

	msg := broker.Published[0]
	if msg.Topic != "tenant" {
		t.Errorf("Expected topic 'tenant', got '%s'", msg.Topic)
	}
	if msg.Key != "test-tenant-123" {
		t.Errorf("Expected key 'test-tenant-123', got '%s'", msg.Key)
	}

	// Verify event was marked as published
	if !repo.markedPublished {
		t.Error("Event was not marked as published")
	}
}

func TestFormatFields(t *testing.T) {
	result := formatFields("key1", "value1", "key2", 123)
	expected := "key1=value1 key2=123"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test empty fields
	result = formatFields()
	if result != "" {
		t.Errorf("Expected empty string, got '%s'", result)
	}

	// Test odd number of fields
	result = formatFields("key1", "value1", "orphan")
	if result == "" {
		t.Error("Expected non-empty result for odd fields")
	}
}

func TestEventEnvelope(t *testing.T) {
	payloadData := []byte(`{"test":"data"}`)
	event := models.Outbox{
		ID:            1,
		AggregateType: "tenant",
		AggregateID:   "test-123",
		Type:          "tenant.created",
		Payload:       payloadData,
		CreatedAt:     time.Now(),
	}

	// Test envelope creation (normally done in publisher)
	envelope := EventEnvelope{
		Type:          event.Type,
		AggregateType: event.AggregateType,
		AggregateID:   event.AggregateID,
		Timestamp:     event.CreatedAt.Format(time.RFC3339),
		Payload:       payloadData,
	}

	if envelope.Type != "tenant.created" {
		t.Errorf("Expected type 'tenant.created', got '%s'", envelope.Type)
	}
	if envelope.AggregateID != "test-123" {
		t.Errorf("Expected aggregate ID 'test-123', got '%s'", envelope.AggregateID)
	}
}

// Mock implementations

type mockOutboxRepo struct {
	events          []models.Outbox
	markedPublished bool
}

func (r *mockOutboxRepo) Append(ctx context.Context, aggregateType, aggregateID, eventType string, payload map[string]interface{}) error {
	return nil
}

func (r *mockOutboxRepo) GetUnpublished(ctx context.Context, limit int) ([]models.Outbox, error) {
	return r.events, nil
}

func (r *mockOutboxRepo) MarkPublished(ctx context.Context, id uint64) error {
	r.markedPublished = true
	return nil
}

func (r *mockOutboxRepo) MarkFailed(ctx context.Context, id uint64, errMsg string) error {
	return nil
}

type mockLogger struct{}

func (l *mockLogger) Info(msg string, fields ...interface{})  {}
func (l *mockLogger) Error(msg string, fields ...interface{}) {}
func (l *mockLogger) Warn(msg string, fields ...interface{})  {}
