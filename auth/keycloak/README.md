# Keycloak Configuration

This directory contains Keycloak realm configuration for the microservice boilerplate.

## Setup

1. Start Keycloak (via Docker Compose or Kubernetes)
2. Access Keycloak admin console at http://localhost:8090
3. Login with admin/admin
4. Import the realm configuration:
   - Go to "Add realm"
   - Click "Import" and select `realm-config.json`

## OAuth2 Client Configuration

The `web-client` is configured with:
- Client ID: `web-client`
- Client Secret: `web-client-secret`
- Redirect URIs: Configured for local and Kubernetes environments
- Flow: Authorization Code Flow (standard OAuth2)

## Custom UI

The authentication UI is built in the Go web service (`services/web/`) using HTMX.
This service communicates with Keycloak via:
- OAuth2/OIDC endpoints for login flow
- REST API for user registration

No Keycloak theme customization is needed.

