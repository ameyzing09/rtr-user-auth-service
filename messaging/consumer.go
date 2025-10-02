package messaging

import (
	"context"
	"encoding/json"
	"fmt"
)

// MessageConsumer defines the interface for consuming messages from a message broker
type MessageConsumer interface {
	// Subscribe registers a handler for messages on a specific topic
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	// Start begins consuming messages
	Start(ctx context.Context) error
	// Close closes the consumer connection
	Close() error
}

// MessageHandler processes incoming messages
type MessageHandler func(ctx context.Context, message []byte) error

// EventHandler processes specific event types
type EventHandler interface {
	// HandleEvent processes an event based on its type
	HandleEvent(ctx context.Context, envelope EventEnvelope) error
}

// TenantEventHandler handles tenant-related events
type TenantEventHandler struct {
	handlers map[string]MessageHandler
	logger   Logger
}

// NewTenantEventHandler creates a new tenant event handler
func NewTenantEventHandler(logger Logger) *TenantEventHandler {
	return &TenantEventHandler{
		handlers: make(map[string]MessageHandler),
		logger:   logger,
	}
}

// RegisterHandler registers a handler for a specific event type
func (h *TenantEventHandler) RegisterHandler(eventType string, handler MessageHandler) {
	h.handlers[eventType] = handler
}

// HandleEvent routes events to the appropriate handler
func (h *TenantEventHandler) HandleEvent(ctx context.Context, envelope EventEnvelope) error {
	handler, exists := h.handlers[envelope.Type]
	if !exists {
		h.logger.Info("No handler registered for event type",
			"eventType", envelope.Type,
			"aggregateType", envelope.AggregateType,
		)
		return nil // Not an error, just skip
	}

	// Call the registered handler with the raw payload
	if err := handler(ctx, envelope.Payload); err != nil {
		return fmt.Errorf("handler failed for event type %s: %w", envelope.Type, err)
	}

	return nil
}

// ProcessMessage is a convenience method to unmarshal and handle messages
func (h *TenantEventHandler) ProcessMessage(ctx context.Context, message []byte) error {
	var envelope EventEnvelope
	if err := json.Unmarshal(message, &envelope); err != nil {
		return fmt.Errorf("failed to unmarshal event envelope: %w", err)
	}

	h.logger.Info("Processing event",
		"type", envelope.Type,
		"aggregateType", envelope.AggregateType,
		"aggregateId", envelope.AggregateID,
	)

	return h.HandleEvent(ctx, envelope)
}

// MockConsumer is a simple mock consumer for testing
type MockConsumer struct {
	subscriptions map[string]MessageHandler
	logger        Logger
}

// NewMockConsumer creates a new mock consumer
func NewMockConsumer(logger Logger) *MockConsumer {
	return &MockConsumer{
		subscriptions: make(map[string]MessageHandler),
		logger:        logger,
	}
}

// Subscribe registers a handler
func (c *MockConsumer) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	c.subscriptions[topic] = handler
	c.logger.Info("Subscribed to topic", "topic", topic)
	return nil
}

// Start is a no-op for mock consumer
func (c *MockConsumer) Start(ctx context.Context) error {
	c.logger.Info("Mock consumer started")
	return nil
}

// Close is a no-op for mock consumer
func (c *MockConsumer) Close() error {
	c.logger.Info("Mock consumer closed")
	return nil
}

// SimulateMessage simulates receiving a message (for testing)
func (c *MockConsumer) SimulateMessage(ctx context.Context, topic string, message []byte) error {
	handler, exists := c.subscriptions[topic]
	if !exists {
		return fmt.Errorf("no handler for topic: %s", topic)
	}
	return handler(ctx, message)
}
