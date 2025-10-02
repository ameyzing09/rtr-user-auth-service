package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rtr-user-auth-service/config"
	"rtr-user-auth-service/consumers"
	"rtr-user-auth-service/internal/db"
	"rtr-user-auth-service/messaging"
	"rtr-user-auth-service/repositories"
	"rtr-user-auth-service/services"
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
	tenantRepo := repositories.NewGormTenantRepo(database)
	userRepo := repositories.NewGormUserRepo(database)

	// Initialize services
	provisioningSvc := services.NewTenantProvisioningService(
		tenantRepo,
		userRepo,
		logger,
	)

	// Initialize event consumers
	tenantEventConsumer := consumers.NewTenantEventConsumer(provisioningSvc, logger)

	// Create tenant event handler
	tenantHandler := messaging.NewTenantEventHandler(logger)

	// Register event handlers
	tenantHandler.RegisterHandler("tenant.created", tenantEventConsumer.HandleTenantCreated)
	tenantHandler.RegisterHandler("tenant.provisioned", tenantEventConsumer.HandleTenantProvisioned)

	// Initialize message consumer (using MockConsumer for now)
	// In production, replace with actual Kafka/RabbitMQ consumer
	messageConsumer := messaging.NewMockConsumer(logger)
	// For production Kafka:
	// messageConsumer := messaging.NewKafkaConsumer(cfg.Kafka, logger)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Subscribe to tenant topic
	if err := messageConsumer.Subscribe(ctx, "tenant", tenantHandler.ProcessMessage); err != nil {
		log.Fatalf("Failed to subscribe to tenant topic: %v", err)
	}

	// Start consumer
	if err := messageConsumer.Start(ctx); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	logger.Info("🚀 Event consumer started successfully")
	logger.Info("Listening for events...", "topics", []string{"tenant"})

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Received shutdown signal, stopping consumer...")

	// Graceful shutdown
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Close consumer
	if err := messageConsumer.Close(); err != nil {
		logger.Error("Failed to close consumer", "error", err)
	}

	select {
	case <-shutdownCtx.Done():
		logger.Warn("Shutdown timeout exceeded, forcing exit")
	default:
		logger.Info("✅ Event consumer stopped gracefully")
	}
}
