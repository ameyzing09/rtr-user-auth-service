package messaging

import (
	"context"
	"encoding/json"
	"fmt"
)

// MessageBroker defines the interface for publishing messages to a message broker
type MessageBroker interface {
	// Publish sends a message to the specified topic/queue
	Publish(ctx context.Context, topic string, key string, data []byte) error
	// Close closes the connection to the message broker
	Close() error
}

// EventEnvelope wraps event data with metadata for consistent message format
type EventEnvelope struct {
	Type          string          `json:"type"`
	AggregateType string          `json:"aggregateType"`
	AggregateID   string          `json:"aggregateId"`
	Timestamp     string          `json:"timestamp"`
	Payload       json.RawMessage `json:"payload"`
}

// MockBroker is a simple in-memory broker for development/testing
type MockBroker struct {
	Published []MockMessage
}

// MockMessage stores published messages for testing
type MockMessage struct {
	Topic string
	Key   string
	Data  []byte
}

// NewMockBroker creates a new mock broker instance
func NewMockBroker() *MockBroker {
	return &MockBroker{
		Published: make([]MockMessage, 0),
	}
}

// Publish stores the message in memory
func (b *MockBroker) Publish(ctx context.Context, topic string, key string, data []byte) error {
	b.Published = append(b.Published, MockMessage{
		Topic: topic,
		Key:   key,
		Data:  data,
	})
	return nil
}

// Close is a no-op for mock broker
func (b *MockBroker) Close() error {
	return nil
}

// GetMessages returns all published messages (for testing)
func (b *MockBroker) GetMessages() []MockMessage {
	return b.Published
}

// LogBroker logs messages instead of publishing (useful for development)
type LogBroker struct {
	logger Logger
}

// Logger defines a simple logging interface
type Logger interface {
	Info(msg string, fields ...interface{})
}

// NewLogBroker creates a broker that logs messages instead of publishing
func NewLogBroker(logger Logger) *LogBroker {
	return &LogBroker{logger: logger}
}

// Publish logs the message instead of publishing
func (b *LogBroker) Publish(ctx context.Context, topic string, key string, data []byte) error {
	b.logger.Info("Publishing message (log-only mode)",
		"topic", topic,
		"key", key,
		"data", string(data),
	)
	return nil
}

// Close is a no-op for log broker
func (b *LogBroker) Close() error {
	return nil
}

// BrokerError wraps broker-related errors
type BrokerError struct {
	Op  string
	Err error
}

func (e *BrokerError) Error() string {
	return fmt.Sprintf("broker error during %s: %v", e.Op, e.Err)
}

func (e *BrokerError) Unwrap() error {
	return e.Err
}
