.PHONY: help setup up down restart logs deps-up deps-down proto-gen build test lint dev dev-user dev-gateway dev-web

# Default target
help:
	@echo "Go Microservices Boilerplate - Available Commands:"
	@echo ""
	@echo "Setup & Configuration:"
	@echo "  make setup              - Initial setup (copy .env.example to .env)"
	@echo ""
	@echo "Dependencies (Infrastructure Only):"
	@echo "  make deps-up            - Start MongoDB"
	@echo "  make deps-down          - Stop infrastructure"
	@echo "  make deps-restart       - Restart infrastructure"
	@echo ""
	@echo "Development (Hot Reload):"
	@echo "  make dev                - Start deps + all services with Air"
	@echo ""
	@echo "Development (Other):"
	@echo "  make proto-gen          - Generate Go code from .proto files"
	@echo "  make deps-install       - Install Go dependencies for all services"
	@echo "  make lint               - Run golangci-lint on all services"
	@echo "  make fmt                - Format all Go code"
	@echo "  make tidy               - Run go mod tidy on all services"
	@echo ""
	@echo "Testing:"
	@echo "  make test-unit          - Run unit tests for all services"
	@echo "  make test-integration   - Run integration tests"
	@echo "  make test-e2e           - Run end-to-end tests"
	@echo "  make test-coverage      - Generate coverage report"
	@echo ""
	@echo "Building:"
	@echo "  make build              - Build all services"
	@echo "  make build-user         - Build user-service"
	@echo "  make build-gateway      - Build api-gateway"
	@echo "  make build-web          - Build web-app"
	@echo "  make docker-build       - Build all Docker images"
	@echo ""
	@echo "Database:"
	@echo "  make db-shell-user      - Open MongoDB shell for user DB"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean              - Remove built binaries"
	@echo "  make clean-all          - Remove binaries, volumes, images"

# Setup
setup:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env from .env.example"; \
		echo "Please update .env with your actual configuration"; \
	else \
		echo "✓ .env already exists"; \
	fi

# Dependencies (Infrastructure Only)
deps-up:
	docker-compose -f deployments/docker-compose/docker-compose.deps.yml up -d
	@echo "Infrastructure started (MongoDB)"

deps-down:
	docker-compose -f deployments/docker-compose/docker-compose.deps.yml down
	@echo "Infrastructure stopped"

deps-restart:
	docker-compose -f deployments/docker-compose/docker-compose.deps.yml restart
	@echo "Infrastructure restarted"

# Development - Hot Reload with Air
# Prerequisites: go install github.com/air-verse/air@latest

dev: deps-up
	@echo "Starting all services with hot reload..."
	@echo "Press Ctrl+C to stop all services"
	@echo ""
	@echo "Services:"
	@echo "  User Service:  localhost:50051 (gRPC)"
	@echo "  API Gateway:   localhost:8080"
	@echo "  Web App:       localhost:3000 (proxied via API Gateway)"
	@echo ""
	@trap 'kill 0' SIGINT; \
	(cd services/user-service && air) & \
	(cd services/api-gateway && air) & \
	(cd services/web-app && air) & \
	wait

# Development - Code Generation
proto-gen:
	@echo "Generating protobuf code..."
	./scripts/proto-gen.sh
	@echo "✓ Protobuf code generated"

deps-install:
	@echo "Installing dependencies for all services..."
	cd shared && go mod download
	cd services/api-gateway && go mod download
	cd services/user-service && go mod download
	cd services/web-app && go mod download
	@echo "✓ Dependencies installed"

lint:
	@echo "Running linters..."
	golangci-lint run ./shared/...
	golangci-lint run ./services/api-gateway/...
	golangci-lint run ./services/user-service/...
	golangci-lint run ./services/web-app/...

fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	@echo "✓ Code formatted"

tidy:
	@echo "Tidying modules..."
	cd shared && go mod tidy
	cd services/api-gateway && go mod tidy
	cd services/user-service && go mod tidy
	cd services/web-app && go mod tidy
	@echo "✓ Modules tidied"

# Testing
test-unit:
	@echo "Running unit tests..."
	go test ./shared/... -v
	go test ./services/api-gateway/... -v
	go test ./services/user-service/internal/service/... -v
	go test ./services/web-app/... -v

test-integration:
	@echo "Running integration tests..."
	go test -tags=integration ./services/user-service/internal/repository/... -v

test-e2e:
	@echo "Running end-to-end tests..."
	./scripts/test-e2e.sh

test-coverage:
	@echo "Generating coverage report..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# Building
build:
	@echo "Building all services..."
	cd services/api-gateway && go build -o ../../bin/api-gateway cmd/main.go
	cd services/user-service && go build -o ../../bin/user-service cmd/main.go
	cd services/web-app && go build -o ../../bin/web-app cmd/main.go
	@echo "✓ All services built"

build-user:
	cd services/user-service && go build -o ../../bin/user-service cmd/main.go

build-gateway:
	cd services/api-gateway && go build -o ../../bin/api-gateway cmd/main.go

build-web:
	cd services/web-app && go build -o ../../bin/web-app cmd/main.go

docker-build:
	@echo "Building Docker images..."
	docker build -t boilerplate/api-gateway -f services/api-gateway/Dockerfile .
	docker build -t boilerplate/user-service -f services/user-service/Dockerfile .
	docker build -t boilerplate/web-app -f services/web-app/Dockerfile .
	@echo "✓ Docker images built"

# Database
db-shell-user:
	docker exec -it mongodb mongosh users_db

# Cleanup
clean:
	@echo "Cleaning built binaries..."
	rm -rf bin/
	@echo "✓ Binaries removed"

clean-all: down
	@echo "Cleaning everything..."
	rm -rf bin/
	docker-compose -f deployments/docker-compose/docker-compose.yml down -v
	docker-compose -f deployments/docker-compose/docker-compose.deps.yml down -v
	@echo "✓ Everything cleaned"

# New service scaffolding
new-service:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make new-service NAME=my-service"; \
		exit 1; \
	fi
	./scripts/new-service.sh $(NAME)
