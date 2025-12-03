#!/bin/bash

# Database Migration Script
# Usage: ./scripts/db-migrate.sh [service] [up|down|status]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
SERVICE="${1:-all}"
ACTION="${2:-up}"

# MongoDB connection settings (from environment or defaults)
MONGO_URI="${MONGO_URI:-mongodb://localhost:27017}"

echo -e "${GREEN}Database Migration Script${NC}"
echo "=================================="
echo "Service: $SERVICE"
echo "Action: $ACTION"
echo "MongoDB URI: $MONGO_URI"
echo ""

# Function to run migrations for a service
run_migration() {
    local service=$1
    local action=$2
    local db_name=$3
    
    echo -e "${YELLOW}Running $action migrations for $service (database: $db_name)...${NC}"
    
    migration_dir="services/$service/migrations"
    
    if [ ! -d "$migration_dir" ]; then
        echo -e "${YELLOW}No migrations directory found for $service, skipping...${NC}"
        return 0
    fi
    
    case $action in
        up)
            # Apply all pending migrations
            for migration in $(ls -1 "$migration_dir"/*.js 2>/dev/null | sort); do
                echo "Applying migration: $(basename $migration)"
                mongosh "$MONGO_URI/$db_name" --quiet "$migration"
            done
            ;;
        down)
            # Rollback last migration (implement based on your needs)
            echo "Rollback not implemented - please handle manually"
            ;;
        status)
            # Show migration status
            echo "Checking migration status for $db_name..."
            mongosh "$MONGO_URI/$db_name" --quiet --eval "db.migrations.find().pretty()"
            ;;
        *)
            echo -e "${RED}Unknown action: $action${NC}"
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}✓ $action completed for $service${NC}"
    echo ""
}

# Function to ensure indexes for a service
ensure_indexes() {
    local service=$1
    local db_name=$2
    
    echo -e "${YELLOW}Ensuring indexes for $service...${NC}"
    
    case $service in
        user-service)
            mongosh "$MONGO_URI/$db_name" --quiet --eval '
                db.users.createIndex({ "email": 1 }, { unique: true });
                db.users.createIndex({ "created_at": -1 });
                print("Indexes created for users collection");
            '
            ;;
        *)
            echo "No index definitions for $service"
            ;;
    esac
}

# Main execution
case $SERVICE in
    all)
        run_migration "user-service" "$ACTION" "users_db"
        ensure_indexes "user-service" "users_db"
        ;;
    user-service)
        run_migration "user-service" "$ACTION" "users_db"
        ensure_indexes "user-service" "users_db"
        ;;
    *)
        echo -e "${RED}Unknown service: $SERVICE${NC}"
        echo "Available services: user-service, all"
        exit 1
        ;;
esac

echo -e "${GREEN}✓ All migrations completed!${NC}"

