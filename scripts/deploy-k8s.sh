#!/bin/bash

set -e

echo "Deploying microservice boilerplate to Kubernetes..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed. Please install kubectl and try again."
    exit 1
fi

# Check if cluster is accessible
if ! kubectl cluster-info &> /dev/null; then
    echo "Error: Kubernetes cluster is not accessible. Please configure kubectl and try again."
    exit 1
fi

BASE_DIR="$(dirname "$0")/.."

# Create namespaces
echo "Creating namespaces..."
kubectl apply -f "$BASE_DIR/infrastructure/kubernetes/namespaces/"

# Deploy data services
echo "Deploying data services..."
kubectl apply -f "$BASE_DIR/infrastructure/kubernetes/data/"

# Deploy messaging
echo "Deploying messaging services..."
kubectl apply -f "$BASE_DIR/infrastructure/kubernetes/messaging/"

# Deploy monitoring
echo "Deploying monitoring..."
kubectl apply -f "$BASE_DIR/infrastructure/kubernetes/monitoring/"

# Deploy auth
echo "Deploying authentication..."
kubectl apply -f "$BASE_DIR/infrastructure/kubernetes/auth/"

# Wait for Keycloak to be ready
echo "Waiting for Keycloak to be ready..."
kubectl wait --for=condition=ready pod -l app=keycloak -n auth --timeout=300s

# Deploy gateway
echo "Deploying gateway..."
kubectl apply -f "$BASE_DIR/infrastructure/kubernetes/gateway/"

# Deploy services
echo "Deploying application services..."
kubectl apply -f "$BASE_DIR/services/web/kubernetes/"
kubectl apply -f "$BASE_DIR/services/api/kubernetes/"

echo ""
echo "Deployment completed!"
echo ""
echo "Next steps:"
echo "  1. Initialize Keycloak realm (import auth/keycloak/realm-config.json)"
echo "  2. Check service status: kubectl get pods --all-namespaces"
echo "  3. Get service URLs: kubectl get svc --all-namespaces"
echo ""

