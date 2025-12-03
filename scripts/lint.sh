#!/bin/bash

# Linting Script
# Usage: ./scripts/lint.sh [service|all]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SERVICE="${1:-all}"

echo -e "${GREEN}Go Linting Script${NC}"
echo "=================================="
echo "Target: $SERVICE"
echo ""

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${RED}Error: golangci-lint is not installed${NC}"
    echo "Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    echo "Or see: https://golangci-lint.run/usage/install/"
    exit 1
fi

echo "golangci-lint version: $(golangci-lint --version)"
echo ""

# Function to lint a directory
lint_dir() {
    local dir=$1
    local name=$2
    
    if [ ! -d "$dir" ]; then
        echo -e "${RED}Error: Directory '$dir' not found${NC}"
        return 1
    fi
    
    echo -e "${YELLOW}Linting $name...${NC}"
    
    cd "$dir"
    golangci-lint run --timeout 5m ./...
    cd - > /dev/null
    
    echo -e "${GREEN}✓ $name passed linting${NC}"
    echo ""
}

# Main execution
case $SERVICE in
    all)
        lint_dir "shared" "shared packages"
        for dir in services/*/; do
            service=$(basename "$dir")
            lint_dir "$dir" "$service"
        done
        ;;
    shared)
        lint_dir "shared" "shared packages"
        ;;
    *)
        lint_dir "services/$SERVICE" "$SERVICE"
        ;;
esac

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}✓ All linting checks passed!${NC}"
echo -e "${GREEN}================================${NC}"

