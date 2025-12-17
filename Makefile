.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Variables
DOCKER_COMPOSE := docker-compose
COMPOSE_FILE := infrastructure/docker-compose.yml
DEVICE_PROTO_DIR := services/device/proto/device
DEVICE_PROTO_SRC := $(DEVICE_PROTO_DIR)/device.proto

# Docker Compose Commands
.PHONY: up
up: ## Start all development containers
	@echo "Starting development containers..."
	cd infrastructure && $(DOCKER_COMPOSE) up -d --build
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo ""
	@echo "Services started! Access points:"
	@echo "  - Web App:        http://localhost:8080"
	@echo "  - Keycloak:       http://localhost:8090 (admin/admin)"
	@echo "  - Prometheus:     http://localhost:9090"
	@echo "  - Grafana:        http://localhost:3000 (admin/admin)"
	@echo "  - MinIO Console:  http://localhost:9001 (minioadmin/minioadmin)"
	@echo "  - Flink UI:       http://localhost:8081"
	@echo "  - Job Orchestrator: http://localhost:8085"
	@echo "  - Data Processor:   http://localhost:8084"

.PHONY: down
down: ## Stop all development containers
	@echo "Stopping development containers..."
	cd infrastructure && $(DOCKER_COMPOSE) down

.PHONY: restart
restart: ## Restart all development containers
	@echo "Restarting development containers..."
	cd infrastructure && $(DOCKER_COMPOSE) restart

.PHONY: logs
logs: ## Show logs from all containers
	cd infrastructure && $(DOCKER_COMPOSE) logs -f

.PHONY: logs-device
logs-device: ## Show logs from Device service
	cd infrastructure && $(DOCKER_COMPOSE) logs -f device-service

.PHONY: logs-web
logs-web: ## Show logs from Web service
	cd infrastructure && $(DOCKER_COMPOSE) logs -f web-service

.PHONY: rebuild-web
rebuild-web: ## Rebuild and restart Web service
	@echo "Rebuilding Web service..."
	cd infrastructure && $(DOCKER_COMPOSE) up -d --build --force-recreate web-service
	@echo "Web service rebuilt and restarted!"

.PHONY: restart-web
restart-web: ## Restart Web service (without rebuild)
	@echo "Restarting Web service..."
	cd infrastructure && $(DOCKER_COMPOSE) restart web-service
	@echo "Web service restarted!"

.PHONY: rebuild-device
rebuild-device: ## Rebuild and restart Device service
	@echo "Rebuilding Device service..."
	cd infrastructure && $(DOCKER_COMPOSE) up -d --build --force-recreate device-service
	@echo "Device service rebuilt and restarted!"

.PHONY: restart-device
restart-device: ## Restart Device service (without rebuild)
	@echo "Restarting Device service..."
	cd infrastructure && $(DOCKER_COMPOSE) restart device-service
	@echo "Device service restarted!"

.PHONY: rebuild-orchestrator
rebuild-orchestrator: ## Rebuild and restart Job Orchestrator service
	@echo "Rebuilding Job Orchestrator service..."
	cd infrastructure && $(DOCKER_COMPOSE) up -d --build --force-recreate job-orchestrator
	@echo "Job Orchestrator rebuilt and restarted!"

.PHONY: restart-orchestrator
restart-orchestrator: ## Restart Job Orchestrator service (without rebuild)
	@echo "Restarting Job Orchestrator service..."
	cd infrastructure && $(DOCKER_COMPOSE) restart job-orchestrator
	@echo "Job Orchestrator restarted!"

.PHONY: rebuild-data-processor
rebuild-data-processor: ## Rebuild and restart Data Processor service
	@echo "Rebuilding Data Processor service..."
	cd infrastructure && $(DOCKER_COMPOSE) up -d --build --force-recreate data-processor
	@echo "Data Processor rebuilt and restarted!"

.PHONY: restart-data-processor
restart-data-processor: ## Restart Data Processor service (without rebuild)
	@echo "Restarting Data Processor service..."
	cd infrastructure && $(DOCKER_COMPOSE) restart data-processor
	@echo "Data Processor restarted!"

.PHONY: logs-orchestrator
logs-orchestrator: ## Show logs from Job Orchestrator service
	cd infrastructure && $(DOCKER_COMPOSE) logs -f job-orchestrator

.PHONY: logs-data-processor
logs-data-processor: ## Show logs from Data Processor service
	cd infrastructure && $(DOCKER_COMPOSE) logs -f data-processor

.PHONY: ps
ps: ## Show status of all containers
	cd infrastructure && $(DOCKER_COMPOSE) ps

.PHONY: clean-containers
clean-containers: ## Stop and remove all containers, networks, and volumes
	@echo "Cleaning up containers, networks, and volumes..."
	cd infrastructure && $(DOCKER_COMPOSE) down -v --remove-orphans

# Protobuf Generation
.PHONY: proto
proto: ## Generate protobuf files for device service
	@echo "Generating protobuf files..."
	@cd services/device && $(MAKE) proto-generate
	@echo "Protobuf files generated successfully!"

.PHONY: proto-clean
proto-clean: ## Clean generated protobuf files
	@echo "Cleaning generated protobuf files..."
	@rm -rf $(DEVICE_PROTO_DIR)/*.pb.go
	@echo "Protobuf files cleaned!"

# Database Commands
.PHONY: migrate
migrate: ## Run database migrations
	@echo "Running database migrations..."
	@./scripts/migrate-db.sh

.PHONY: db-reset
db-reset: ## Reset database (WARNING: This will delete all data)
	@echo "Resetting database..."
	@read -p "Are you sure you want to reset the database? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		cd infrastructure && $(DOCKER_COMPOSE) down -v postgres; \
		cd infrastructure && $(DOCKER_COMPOSE) up -d postgres; \
		sleep 5; \
		$(MAKE) migrate; \
	fi

# Build Commands
.PHONY: build-device
build-device: ## Build Device service Docker image
	@echo "Building Device service Docker image..."
	cd infrastructure && $(DOCKER_COMPOSE) build device-service

.PHONY: build-web
build-web: ## Build Web service Docker image
	@echo "Building Web service Docker image..."
	cd infrastructure && $(DOCKER_COMPOSE) build web-service

.PHONY: build-orchestrator
build-orchestrator: ## Build Job Orchestrator service Docker image
	@echo "Building Job Orchestrator service Docker image..."
	cd infrastructure && $(DOCKER_COMPOSE) build job-orchestrator

.PHONY: build
build: build-device build-web build-orchestrator ## Build all Go service Docker images
	@echo "All Go services built successfully!"

# Health Checks
.PHONY: health
health: ## Check health of all services
	@./scripts/health-check.sh

.PHONY: health-device
health-device: ## Check Device service health
	@echo "Checking Device service health..."
	@curl -f http://localhost:8082/health || echo "Device service is not responding"

# Development Utilities
.PHONY: install-deps
install-deps: ## Install development dependencies
	@echo "Installing development dependencies..."
	@command -v protoc >/dev/null 2>&1 || { echo "Error: protoc is not installed. Please install Protocol Buffers compiler."; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "Error: docker is not installed. Please install Docker."; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "Error: docker-compose is not installed. Please install Docker Compose."; exit 1; }
	@echo "Installing Go dependencies..."
	@cd services/device && go mod download
	@cd services/web && go mod download
	@cd services/job-orchestrator && go mod download
	@echo "Dependencies installed!"

# Grouped Rebuild Targets
.PHONY: rebuild-go
rebuild-go: rebuild-device rebuild-web rebuild-orchestrator ## Rebuild all Go services (device, web, job-orchestrator)
	@echo "All Go services rebuilt!"

.PHONY: rebuild-backend
rebuild-backend: rebuild-device rebuild-orchestrator rebuild-data-processor ## Rebuild backend services (device, job-orchestrator, data-processor)
	@echo "All backend services rebuilt!"

.PHONY: rebuild-services
rebuild-services: rebuild-device rebuild-web rebuild-orchestrator rebuild-data-processor ## Rebuild all custom application services
	@echo "All custom services rebuilt!"

.PHONY: rebuild-deps
rebuild-deps: ## Rebuild infrastructure dependencies (postgres, nats, keycloak, etc.)
	@echo "Rebuilding infrastructure services..."
	cd infrastructure && $(DOCKER_COMPOSE) build postgres postgres-keycloak nats keycloak minio prometheus grafana flink-jobmanager flink-taskmanager envoy
	@echo "Infrastructure services rebuilt!"

.PHONY: rebuild-all
rebuild-all: rebuild-services rebuild-deps ## Rebuild everything (all services and dependencies)
	@echo "Everything rebuilt!"

# Development Mode Commands (Hot Reload with Air)
.PHONY: dev
dev: ## Start development environment with hot reload (uses docker-compose.dev.yml)
	@echo "Starting development environment with hot reload..."
	cd infrastructure && $(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml up -d --build
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Running database migrations..."
	@$(MAKE) migrate
	@echo ""
	@echo "Development environment started with hot reload!"
	@echo "Services will automatically rebuild when source files change."
	@echo ""
	@echo "Access points:"
	@echo "  - Web App:        http://localhost:8080"
	@echo "  - Keycloak:       http://localhost:8090 (admin/admin)"
	@echo "  - Prometheus:     http://localhost:9090"
	@echo "  - Grafana:        http://localhost:3000 (admin/admin)"
	@echo "  - MinIO Console:  http://localhost:9001 (minioadmin/minioadmin)"
	@echo "  - Flink UI:       http://localhost:8081"
	@echo "  - Job Orchestrator: http://localhost:8085"
	@echo "  - Data Processor:   http://localhost:8084"
	@echo ""
	@echo "Use 'make dev-logs' to view logs, 'make dev-down' to stop."

.PHONY: dev-up
dev-up: ## Start development environment (hot reload mode)
	cd infrastructure && $(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml up -d

.PHONY: dev-down
dev-down: ## Stop development environment (hot reload mode)
	@echo "Stopping development environment..."
	cd infrastructure && $(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml down

.PHONY: dev-logs
dev-logs: ## Show logs from development environment
	cd infrastructure && $(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml logs -f

.PHONY: dev-ps
dev-ps: ## Show status of development containers
	cd infrastructure && $(DOCKER_COMPOSE) -f docker-compose.yml -f docker-compose.dev.yml ps

# Production/Standard Development Commands
.PHONY: prod-dev
prod-dev: build up migrate ## Start development environment without hot reload (build Docker images, start containers, run migrations)
	@echo "Development environment ready!"

# Cleanup
.PHONY: clean
clean: proto-clean ## Clean generated files and build artifacts
	@echo "Cleaning build artifacts..."
	@cd services/device && $(MAKE) clean || true
	@rm -f services/web/web-service
	@echo "Cleanup complete!"

.PHONY: clean-all
clean-all: clean-containers clean ## Clean everything (containers, volumes, and generated files)
	@echo "Full cleanup complete!"

# Testing (placeholder for future test commands)
.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@cd services/device && go test ./... || true
	@cd services/web && go test ./... || true
	@cd services/job-orchestrator && go test ./... || true

.PHONY: test-device
test-device: ## Run Device service tests
	@echo "Running Device service tests..."
	@cd services/device && go test ./...

.PHONY: test-web
test-web: ## Run Web service tests
	@echo "Running Web service tests..."
	@cd services/web && go test ./...

