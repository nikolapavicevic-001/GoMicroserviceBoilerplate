#!/bin/bash

# New Service Scaffolding Script
# Usage: ./scripts/new-service.sh <service-name>

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check for service name argument
if [ -z "$1" ]; then
    echo -e "${RED}Error: Service name is required${NC}"
    echo "Usage: ./scripts/new-service.sh <service-name>"
    echo "Example: ./scripts/new-service.sh order-service"
    exit 1
fi

SERVICE_NAME="$1"
SERVICE_DIR="services/$SERVICE_NAME"

# Check if service already exists
if [ -d "$SERVICE_DIR" ]; then
    echo -e "${RED}Error: Service '$SERVICE_NAME' already exists${NC}"
    exit 1
fi

echo -e "${GREEN}Creating new service: $SERVICE_NAME${NC}"
echo "=================================="

# Create directory structure
echo -e "${YELLOW}Creating directory structure...${NC}"
mkdir -p "$SERVICE_DIR"/{cmd,config,internal/{domain,repository/mongodb,service,grpc,events},migrations}

# Create go.mod
echo -e "${YELLOW}Creating go.mod...${NC}"
cat > "$SERVICE_DIR/go.mod" << EOF
module github.com/yourorg/boilerplate/services/$SERVICE_NAME

go 1.22

require (
	github.com/yourorg/boilerplate/shared v0.0.0
	google.golang.org/grpc v1.68.0
)

replace github.com/yourorg/boilerplate/shared => ../../shared
EOF

# Create config/config.go
echo -e "${YELLOW}Creating config...${NC}"
cat > "$SERVICE_DIR/config/config.go" << 'EOF'
package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v10"
)

// Config holds the configuration for the service
type Config struct {
	// Server
	GRPCPort int `env:"${SERVICE_NAME_UPPER}_GRPC_PORT" envDefault:"50052"`

	// Database
	MongoURI  string `env:"${SERVICE_NAME_UPPER}_MONGO_URI" envDefault:"mongodb://localhost:27017"`
	MongoDB   string `env:"${SERVICE_NAME_UPPER}_MONGO_DB" envDefault:"${SERVICE_NAME_SNAKE}_db"`

	// Kafka
	KafkaBrokers string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`

	// Logging
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat string `env:"LOG_FORMAT" envDefault:"json"`

	// Environment
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return cfg, nil
}

// GetKafkaBrokers returns Kafka brokers as a slice
func (c *Config) GetKafkaBrokers() []string {
	return strings.Split(c.KafkaBrokers, ",")
}
EOF

# Replace placeholders in config
SERVICE_NAME_UPPER=$(echo "$SERVICE_NAME" | tr '[:lower:]-' '[:upper:]_')
SERVICE_NAME_SNAKE=$(echo "$SERVICE_NAME" | tr '-' '_')
sed -i "s/\${SERVICE_NAME_UPPER}/${SERVICE_NAME_UPPER}/g" "$SERVICE_DIR/config/config.go"
sed -i "s/\${SERVICE_NAME_SNAKE}/${SERVICE_NAME_SNAKE}/g" "$SERVICE_DIR/config/config.go"

# Create cmd/main.go
echo -e "${YELLOW}Creating main.go...${NC}"
cat > "$SERVICE_DIR/cmd/main.go" << EOF
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/yourorg/boilerplate/services/$SERVICE_NAME/config"
	"github.com/yourorg/boilerplate/shared/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel, cfg.LogFormat)

	log.Info().
		Str("service", "$SERVICE_NAME").
		Str("environment", cfg.Environment).
		Int("grpc_port", cfg.GRPCPort).
		Msg("Starting service")

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: Initialize dependencies (MongoDB, Kafka, etc.)
	_ = ctx

	// Create gRPC server
	grpcServer := grpc.NewServer()

	// TODO: Register your service handlers here
	// pb.Register<Service>Server(grpcServer, handler)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	// Register reflection for grpcurl
	reflection.Register(grpcServer)

	// Start gRPC server
	addr := fmt.Sprintf(":%d", cfg.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen")
	}

	// Start server in goroutine
	go func() {
		log.Info().
			Str("address", addr).
			Msg("gRPC server listening")

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("Failed to serve gRPC")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down service...")

	// Graceful shutdown
	grpcServer.GracefulStop()

	log.Info().Msg("Service stopped")
}
EOF

# Create Dockerfile
echo -e "${YELLOW}Creating Dockerfile...${NC}"
cat > "$SERVICE_DIR/Dockerfile" << EOF
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy shared module
COPY shared/ ./shared/

# Copy service
COPY services/$SERVICE_NAME/ ./services/$SERVICE_NAME/

# Build
WORKDIR /app/services/$SERVICE_NAME
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/services/$SERVICE_NAME/main .

EXPOSE 50052

CMD ["./main"]
EOF

# Create .gitkeep files for empty directories
touch "$SERVICE_DIR/internal/domain/.gitkeep"
touch "$SERVICE_DIR/internal/repository/mongodb/.gitkeep"
touch "$SERVICE_DIR/internal/service/.gitkeep"
touch "$SERVICE_DIR/internal/grpc/.gitkeep"
touch "$SERVICE_DIR/internal/events/.gitkeep"
touch "$SERVICE_DIR/migrations/.gitkeep"

# Update go.work
echo -e "${YELLOW}Updating go.work...${NC}"
if ! grep -q "$SERVICE_DIR" go.work; then
    sed -i "/use (/a\\	./$SERVICE_DIR" go.work
    echo "Added $SERVICE_DIR to go.work"
fi

echo ""
echo -e "${GREEN}✓ Service '$SERVICE_NAME' created successfully!${NC}"
echo ""
echo "Next steps:"
echo "  1. Define your protobuf schema in shared/proto/$SERVICE_NAME/v1/"
echo "  2. Run 'make proto-gen' to generate Go code"
echo "  3. Implement domain entities in internal/domain/"
echo "  4. Implement repository in internal/repository/"
echo "  5. Implement service logic in internal/service/"
echo "  6. Implement gRPC handlers in internal/grpc/"
echo "  7. Update cmd/main.go to wire everything together"
echo ""
echo "To run the service:"
echo "  cd $SERVICE_DIR && go run cmd/main.go"

