#!/bin/bash

set -e

echo "Health check for microservice boilerplate..."

# Check Docker services
if command -v docker &> /dev/null && docker info > /dev/null 2>&1; then
    echo "Checking Docker services..."
    docker-compose -f infrastructure/docker-compose.yml ps
fi

# Check Kubernetes services
if command -v kubectl &> /dev/null && kubectl cluster-info &> /dev/null 2>&1; then
    echo ""
    echo "Checking Kubernetes services..."
    kubectl get pods --all-namespaces
    echo ""
    echo "Service endpoints:"
    kubectl get svc --all-namespaces
fi

echo ""
echo "Health check completed!"

