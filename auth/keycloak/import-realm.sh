#!/bin/bash

# Wait for Keycloak to be ready
echo "Waiting for Keycloak to be ready..."
until curl -f http://localhost:8080/health/ready 2>/dev/null; do
  echo "Keycloak is not ready yet, waiting..."
  sleep 5
done

echo "Keycloak is ready, importing realm..."

# Get admin token
TOKEN=$(curl -s -X POST http://localhost:8080/realms/master/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=admin" \
  -d "password=admin" \
  -d "grant_type=password" \
  -d "client_id=admin-cli" | jq -r '.access_token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "Failed to get admin token"
  exit 1
fi

# Import realm
REALM_FILE="/opt/keycloak/data/import/realm-config.json"
if [ -f "$REALM_FILE" ]; then
  curl -s -X POST http://localhost:8080/admin/realms \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d @"$REALM_FILE"
  
  if [ $? -eq 0 ]; then
    echo "Realm imported successfully"
  else
    echo "Realm may already exist or import failed"
  fi
else
  echo "Realm file not found at $REALM_FILE"
fi

