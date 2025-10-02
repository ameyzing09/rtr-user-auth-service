package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rtr-user-auth-service/config"
	"rtr-user-auth-service/internal/db"
	"rtr-user-auth-service/messaging"
	"rtr-user-auth-service/repositories"
	"rtr-user-auth-service/utils"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize logger
	utilsLogger := utils.NewLogger(utils.LogLevelInfo)
	logger := messaging.NewLoggerAdapter(utilsLogger)

	// Initialize repositories
	outboxRepo := repositories.NewGormOutboxRepo(database)

	// Initialize message broker (using LogBroker for now)
	// In production, replace with Kafka, RabbitMQ, or other broker
	broker := messaging.NewLogBroker(logger)
	// For production Kafka:
	// broker := messaging.NewKafkaBroker(cfg.Kafka)

	// Configure and create publisher
	publisherConfig := messaging.DefaultPublisherConfig()
	publisher := messaging.NewOutboxPublisher(outboxRepo, broker, logger, publisherConfig)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start publisher in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := publisher.Start(ctx); err != nil && err != context.Canceled {
			errChan <- err
		}
	}()

	logger.Info("🚀 Outbox publisher started successfully",
		"batchSize", publisherConfig.BatchSize,
		"pollInterval", publisherConfig.PollInterval,
	)

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		logger.Info("Received shutdown signal, stopping publisher...")
	case err := <-errChan:
		logger.Error("Publisher error", "error", err)
	}

	// Graceful shutdown
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), publisherConfig.ShutdownTimeout)
	defer shutdownCancel()

	// Give time for in-flight operations to complete
	time.Sleep(2 * time.Second)

	// Close broker connection
	if err := broker.Close(); err != nil {
		logger.Error("Failed to close broker", "error", err)
	}

	select {
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout exceeded, forcing exit")
	default:
		logger.Info("✅ Outbox publisher stopped gracefully")
	}
}
