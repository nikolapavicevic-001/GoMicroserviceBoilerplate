#!/bin/bash

# End-to-End Test Script
# Usage: ./scripts/test-e2e.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}End-to-End Test Runner${NC}"
echo "=================================="
echo ""

# Configuration
COMPOSE_FILE="deployments/docker-compose/docker-compose.yml"
TIMEOUT=120
API_GATEWAY_URL="http://localhost:8080"
WEB_APP_URL="http://localhost:3000"

# Cleanup function
cleanup() {
    echo -e "${YELLOW}Cleaning up...${NC}"
    docker-compose -f "$COMPOSE_FILE" down -v --remove-orphans 2>/dev/null || true
}

# Set trap for cleanup on exit
trap cleanup EXIT

# Function to wait for a service to be healthy
wait_for_service() {
    local url=$1
    local name=$2
    local max_attempts=$3
    local attempt=1
    
    echo -e "${YELLOW}Waiting for $name to be ready...${NC}"
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$url/health" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ $name is ready${NC}"
            return 0
        fi
        echo "Attempt $attempt/$max_attempts - $name not ready yet..."
        sleep 2
        ((attempt++))
    done
    
    echo -e "${RED}✗ $name failed to become ready${NC}"
    return 1
}

# Start services
echo -e "${YELLOW}Starting all services...${NC}"
docker-compose -f "$COMPOSE_FILE" up -d --build

# Wait for services to be healthy
echo ""
echo -e "${YELLOW}Waiting for services to be healthy...${NC}"
sleep 10

wait_for_service "$API_GATEWAY_URL" "API Gateway" 30
wait_for_service "$WEB_APP_URL" "Web App" 30

echo ""
echo -e "${GREEN}All services are ready!${NC}"
echo ""

# Run E2E tests
echo -e "${YELLOW}Running E2E tests...${NC}"
echo ""

# Test 1: Health check endpoints
echo "Test 1: Health check endpoints"
echo "  Testing API Gateway health..."
if curl -s -f "$API_GATEWAY_URL/health" > /dev/null; then
    echo -e "  ${GREEN}✓ API Gateway health check passed${NC}"
else
    echo -e "  ${RED}✗ API Gateway health check failed${NC}"
    exit 1
fi

echo "  Testing Web App health..."
if curl -s -f "$WEB_APP_URL/health" > /dev/null; then
    echo -e "  ${GREEN}✓ Web App health check passed${NC}"
else
    echo -e "  ${RED}✗ Web App health check failed${NC}"
    exit 1
fi
echo ""

# Test 2: User creation flow
echo "Test 2: User creation flow"
echo "  Creating test user..."
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_GATEWAY_URL/api/v1/users" \
    -H "Content-Type: application/json" \
    -d '{"email":"e2e-test@example.com","name":"E2E Test User","password":"testpass123"}')

HTTP_CODE=$(echo "$RESPONSE" | tail -1)
BODY=$(echo "$RESPONSE" | head -n -1)

if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "409" ]; then
    echo -e "  ${GREEN}✓ User creation endpoint works (HTTP $HTTP_CODE)${NC}"
else
    echo -e "  ${RED}✗ User creation failed (HTTP $HTTP_CODE)${NC}"
    echo "  Response: $BODY"
    exit 1
fi
echo ""

# Test 3: User listing (requires auth in production)
echo "Test 3: API endpoints accessibility"
echo "  Testing user listing endpoint..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$API_GATEWAY_URL/api/v1/users")
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "401" ]; then
    echo -e "  ${GREEN}✓ User listing endpoint accessible (HTTP $HTTP_CODE)${NC}"
else
    echo -e "  ${RED}✗ User listing endpoint failed (HTTP $HTTP_CODE)${NC}"
    exit 1
fi
echo ""

# Test 4: Web App pages
echo "Test 4: Web App pages"
echo "  Testing login page..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$WEB_APP_URL/login")
if [ "$HTTP_CODE" = "200" ]; then
    echo -e "  ${GREEN}✓ Login page loads (HTTP $HTTP_CODE)${NC}"
else
    echo -e "  ${RED}✗ Login page failed (HTTP $HTTP_CODE)${NC}"
    exit 1
fi
echo ""

# Test 5: Static assets
echo "Test 5: Static assets"
echo "  Testing CSS file..."
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$WEB_APP_URL/static/css/styles.css")
if [ "$HTTP_CODE" = "200" ]; then
    echo -e "  ${GREEN}✓ Static CSS loads (HTTP $HTTP_CODE)${NC}"
else
    echo -e "  ${YELLOW}⚠ Static CSS not found (HTTP $HTTP_CODE) - may be expected${NC}"
fi
echo ""

# Run Go E2E tests if they exist
if [ -d "tests/e2e" ]; then
    echo -e "${YELLOW}Running Go E2E tests...${NC}"
    go test -tags=e2e -v -timeout 5m ./tests/e2e/...
fi

echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}✓ All E2E tests passed!${NC}"
echo -e "${GREEN}================================${NC}"

