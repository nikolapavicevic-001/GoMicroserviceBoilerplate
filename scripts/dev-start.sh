#!/bin/bash

set -e

echo "Starting microservice boilerplate for local development..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running. Please start Docker and try again."
    exit 1
fi

# Navigate to infrastructure directory
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT/infrastructure"

# Start services
echo "Starting Docker Compose services..."
docker-compose up -d --build

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 15

# Run wait script if available
if [ -f "$PROJECT_ROOT/scripts/wait-for-services.sh" ]; then
    chmod +x "$PROJECT_ROOT/scripts/wait-for-services.sh"
    "$PROJECT_ROOT/scripts/wait-for-services.sh"
fi

# Check service health
echo "Checking service health..."
docker-compose ps

echo ""
echo "Services started successfully!"
echo ""
echo "Access points:"
echo "  - Web App:        http://localhost:8080"
echo "  - Keycloak:       http://localhost:8090 (admin/admin)"
echo "  - Prometheus:     http://localhost:9090"
echo "  - Grafana:        http://localhost:3000 (admin/admin)"
echo "  - MinIO Console:  http://localhost:9001 (minioadmin/minioadmin)"
echo "  - Flink UI:       http://localhost:8081"
echo ""
echo "Next steps:"
echo "  1. Wait for Keycloak to fully start (60-90 seconds)"
echo "  2. Realm should be automatically imported"
echo "  3. Access web app at http://localhost:8080 and register/login"
echo ""
echo "To stop services: docker-compose down"

