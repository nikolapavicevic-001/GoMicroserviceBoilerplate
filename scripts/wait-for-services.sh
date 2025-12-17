#!/bin/bash

set -e

echo "Waiting for services to be ready..."

# Wait for PostgreSQL
echo "Waiting for PostgreSQL..."
until docker exec postgres pg_isready -U postgres > /dev/null 2>&1; do
  sleep 2
done
echo "PostgreSQL is ready"

# Wait for Keycloak database
echo "Waiting for Keycloak database..."
until docker exec postgres-keycloak pg_isready -U keycloak > /dev/null 2>&1; do
  sleep 2
done
echo "Keycloak database is ready"

# Wait for Keycloak
echo "Waiting for Keycloak..."
until curl -f http://localhost:8090/health/ready > /dev/null 2>&1; do
  sleep 5
done
echo "Keycloak is ready"

# Wait for Web Service
echo "Waiting for Web Service..."
until curl -f http://localhost:8081/health > /dev/null 2>&1; do
  sleep 2
done
echo "Web Service is ready"

# Wait for API Service
echo "Waiting for API Service..."
until curl -f http://localhost:8082/health > /dev/null 2>&1; do
  sleep 2
done
echo "API Service is ready"

echo "All services are ready!"

