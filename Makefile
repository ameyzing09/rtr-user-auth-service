# Variables
BINARY_NAME=recrutr-auth-service
DOCKER_IMAGE=recrutr-auth-service
DOCKER_TAG=latest
DOCKER_COMPOSE_FILE=docker-compose.yml

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Build paths
BUILD_DIR=./bin
SOURCE_DIR=./cmd/server
COVERAGE_DIR=./coverage

.PHONY: help build clean test test-coverage run dev docker-build docker-run docker-compose-up docker-compose-down fmt lint mod-tidy mod-download install-tools

# Default target
all: clean fmt lint test build

# Show help
help: ## Display this help screen
	@echo "Recrutr Auth Service - Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Build the binary
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(SOURCE_DIR)
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for Linux (useful for Docker)
build-linux: ## Build the application for Linux
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(SOURCE_DIR)
	@echo "Linux build completed: $(BUILD_DIR)/$(BINARY_NAME)-linux"

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@echo "Clean completed"

# Run tests
test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	@$(GOTEST) -v -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	@$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

# Run the application
run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

# Run in development mode with hot reload (requires air)
dev: ## Run in development mode with hot reload
	@echo "Running in development mode..."
	@if which air > /dev/null; then \
		air; \
	else \
		echo "air is not installed. Install it with: go install github.com/air-verse/air@latest"; \
		echo "Falling back to regular run..."; \
		$(MAKE) run; \
	fi

# Format code
fmt: ## Format Go code
	@echo "Formatting code..."
	@$(GOFMT) -w .

# Lint code (requires golangci-lint)
lint: ## Lint Go code
	@echo "Linting code..."
	@if which golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Install it from https://golangci-lint.run/usage/install/"; \
	fi

# Tidy dependencies
mod-tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	@$(GOMOD) tidy

# Download dependencies
mod-download: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GOMOD) download

# Install development tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@$(GOCMD) install github.com/air-verse/air@latest
	@$(GOCMD) install github.com/swaggo/swag/cmd/swag@latest
	@echo "Development tools installed"

# Generate Swagger documentation
swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@if which swag > /dev/null; then \
		swag init -g cmd/server/main.go -o ./docs; \
		echo "Swagger documentation generated in ./docs"; \
	else \
		echo "swag is not installed. Install it with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-run: docker-build ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker Compose commands
docker-compose-up: ## Start services with Docker Compose
	@echo "Starting services with Docker Compose..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) up --build -d
	@echo "Services started. Check logs with: make docker-compose-logs"

docker-compose-down: ## Stop services with Docker Compose
	@echo "Stopping services with Docker Compose..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down
	@echo "Services stopped"

docker-compose-logs: ## Show Docker Compose logs
	@docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

docker-compose-restart: ## Restart services with Docker Compose
	@$(MAKE) docker-compose-down
	@$(MAKE) docker-compose-up

# Database commands
db-migrate: ## Run database migrations (requires running service)
	@echo "Database migrations are run automatically when the service starts"

db-reset: ## Reset database (WARNING: This will delete all data)
	@echo "Resetting database..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down -v
	@docker-compose -f $(DOCKER_COMPOSE_FILE) up --build -d mysql
	@echo "Database reset completed"

# Setup development environment
setup-dev: mod-download install-tools ## Setup development environment
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Development environment setup completed"
	@echo "Please edit .env file with your configuration"

# Production build
build-prod: clean fmt lint test build-linux ## Build for production
	@echo "Production build completed"

# Health check
health-check: ## Check if the service is healthy
	@echo "Checking service health..."
	@curl -f http://localhost:8080/health || (echo "Service is not healthy" && exit 1)
	@echo "Service is healthy"