#!/bin/bash

# Integration Test Script
# Usage: ./scripts/test-integration.sh [service]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SERVICE="${1:-all}"

echo -e "${GREEN}Integration Test Runner${NC}"
echo "=================================="
echo "Service: $SERVICE"
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running${NC}"
    echo "Integration tests require Docker for testcontainers"
    exit 1
fi

# Function to run integration tests for a service
run_integration_tests() {
    local service=$1
    local service_dir="services/$service"
    
    if [ ! -d "$service_dir" ]; then
        echo -e "${RED}Error: Service '$service' not found${NC}"
        return 1
    fi
    
    echo -e "${YELLOW}Running integration tests for $service...${NC}"
    
    # Check for integration test files
    if ! find "$service_dir" -name "*_integration_test.go" -o -name "*_test.go" | xargs grep -l "//go:build integration" > /dev/null 2>&1; then
        echo -e "${YELLOW}No integration tests found for $service, skipping...${NC}"
        return 0
    fi
    
    # Run tests with integration tag
    cd "$service_dir"
    go test -tags=integration -v -race -timeout 5m ./...
    cd - > /dev/null
    
    echo -e "${GREEN}✓ Integration tests passed for $service${NC}"
    echo ""
}

# Run shared package integration tests
run_shared_integration_tests() {
    echo -e "${YELLOW}Running integration tests for shared packages...${NC}"
    
    cd shared
    if find . -name "*_integration_test.go" > /dev/null 2>&1; then
        go test -tags=integration -v -race -timeout 5m ./...
    else
        echo "No integration tests found for shared packages"
    fi
    cd - > /dev/null
    
    echo ""
}

# Main execution
case $SERVICE in
    all)
        run_shared_integration_tests
        for dir in services/*/; do
            service=$(basename "$dir")
            run_integration_tests "$service"
        done
        ;;
    shared)
        run_shared_integration_tests
        ;;
    *)
        run_integration_tests "$SERVICE"
        ;;
esac

echo -e "${GREEN}✓ All integration tests completed!${NC}"

