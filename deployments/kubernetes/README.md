# Kubernetes Deployment

This directory contains Kubernetes manifests and Kustomize configurations for deploying the boilerplate application.

## Structure

```
kubernetes/
├── base/                    # Base Kubernetes manifests
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secret.yaml
│   ├── mongodb.yaml
│   ├── kafka.yaml
│   ├── jaeger.yaml
│   ├── prometheus.yaml
│   ├── user-service.yaml
│   ├── api-gateway.yaml
│   ├── web-app.yaml
│   └── kustomization.yaml
└── overlays/               # Environment-specific overlays
    ├── dev/
    │   └── kustomization.yaml
    ├── staging/
    │   └── kustomization.yaml
    └── prod/
        └── kustomization.yaml
```

## Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured
- kustomize (built into kubectl 1.14+)
- Container registry access for pushing images

## Quick Start

### 1. Build and Push Docker Images

```bash
# Build all services
make docker-build

# Tag for your registry
docker tag yourorg/user-service:latest your-registry.com/user-service:latest
docker tag yourorg/api-gateway:latest your-registry.com/api-gateway:latest
docker tag yourorg/web-app:latest your-registry.com/web-app:latest

# Push to registry
docker push your-registry.com/user-service:latest
docker push your-registry.com/api-gateway:latest
docker push your-registry.com/web-app:latest
```

### 2. Update Image Names

Update the image names in the kustomization files to match your registry:

```bash
# Edit overlays/dev/kustomization.yaml, staging/kustomization.yaml, prod/kustomization.yaml
# Change yourorg/* to your-registry.com/*
```

### 3. Create Secrets

Before deploying, create the secrets with actual values:

```bash
kubectl create namespace boilerplate-dev

kubectl create secret generic app-secrets \
  --from-literal=JWT_SECRET='your-actual-jwt-secret-minimum-32-characters' \
  --from-literal=SESSION_SECRET='your-actual-session-secret-minimum-32-chars' \
  --from-literal=CSRF_SECRET='your-actual-csrf-secret' \
  --from-literal=OAUTH2_GOOGLE_CLIENT_ID='your-google-client-id' \
  --from-literal=OAUTH2_GOOGLE_CLIENT_SECRET='your-google-client-secret' \
  --from-literal=OAUTH2_GITHUB_CLIENT_ID='your-github-client-id' \
  --from-literal=OAUTH2_GITHUB_CLIENT_SECRET='your-github-client-secret' \
  -n boilerplate-dev
```

### 4. Deploy

#### Development

```bash
kubectl apply -k deployments/kubernetes/overlays/dev
```

#### Staging

```bash
kubectl apply -k deployments/kubernetes/overlays/staging
```

#### Production

```bash
kubectl apply -k deployments/kubernetes/overlays/prod
```

## Verify Deployment

```bash
# Check all resources
kubectl get all -n boilerplate-dev

# Check pods
kubectl get pods -n boilerplate-dev

# Check services
kubectl get svc -n boilerplate-dev

# View logs
kubectl logs -f deployment/user-service -n boilerplate-dev
kubectl logs -f deployment/api-gateway -n boilerplate-dev
kubectl logs -f deployment/web-app -n boilerplate-dev
```

## Access Services

### Using Port Forwarding

```bash
# API Gateway
kubectl port-forward svc/api-gateway 8080:8080 -n boilerplate-dev

# Web App
kubectl port-forward svc/web-app 3000:3000 -n boilerplate-dev

# Jaeger UI
kubectl port-forward svc/jaeger 16686:16686 -n boilerplate-dev

# Prometheus
kubectl port-forward svc/prometheus 9090:9090 -n boilerplate-dev
```

### Using LoadBalancer (if available)

```bash
# Get external IPs
kubectl get svc -n boilerplate-dev

# Access services at the external IPs
```

## Configuration

### Environment-Specific Settings

Each overlay (dev/staging/prod) has different configurations:

**Development:**
- 1 replica per service
- Debug logging
- Console log format
- Lower resource limits

**Staging:**
- 2 replicas per service
- Info logging
- JSON log format
- Medium resource limits

**Production:**
- 3 replicas per service
- Info logging
- JSON log format
- Higher resource limits
- Larger storage volumes

### Updating Configuration

Edit the ConfigMap in `base/configmap.yaml` or override in overlay kustomization files:

```yaml
configMapGenerator:
  - name: app-config
    behavior: merge
    literals:
      - LOG_LEVEL=debug
      - NEW_CONFIG_KEY=value
```

### Updating Secrets

```bash
# Update secret
kubectl create secret generic app-secrets \
  --from-literal=JWT_SECRET='new-secret' \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart deployments to pick up new secrets
kubectl rollout restart deployment/user-service -n boilerplate-dev
kubectl rollout restart deployment/api-gateway -n boilerplate-dev
kubectl rollout restart deployment/web-app -n boilerplate-dev
```

## Scaling

```bash
# Manual scaling
kubectl scale deployment/user-service --replicas=5 -n boilerplate-dev

# Horizontal Pod Autoscaler (HPA)
kubectl autoscale deployment/user-service \
  --cpu-percent=70 \
  --min=2 \
  --max=10 \
  -n boilerplate-dev
```

## Monitoring

### Health Checks

All services have liveness and readiness probes configured.

```bash
# Check probe status
kubectl describe pod <pod-name> -n boilerplate-dev
```

### Metrics

Prometheus scrapes metrics from all services on their metrics ports:
- user-service: 9091
- api-gateway: 9090
- web-app: 9092

Access Prometheus UI:
```bash
kubectl port-forward svc/prometheus 9090:9090 -n boilerplate-dev
# Open http://localhost:9090
```

### Tracing

Jaeger collects distributed traces from all services.

Access Jaeger UI:
```bash
kubectl port-forward svc/jaeger 16686:16686 -n boilerplate-dev
# Open http://localhost:16686
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n boilerplate-dev
kubectl describe pod <pod-name> -n boilerplate-dev
```

### View Logs

```bash
# Recent logs
kubectl logs <pod-name> -n boilerplate-dev

# Follow logs
kubectl logs -f <pod-name> -n boilerplate-dev

# Previous container logs (if crashed)
kubectl logs <pod-name> --previous -n boilerplate-dev
```

### Exec into Container

```bash
kubectl exec -it <pod-name> -n boilerplate-dev -- /bin/sh
```

### Check Events

```bash
kubectl get events -n boilerplate-dev --sort-by='.lastTimestamp'
```

### Common Issues

1. **ImagePullBackOff**: Check image name and registry credentials
2. **CrashLoopBackOff**: Check logs with `kubectl logs`
3. **Pending Pods**: Check resource availability with `kubectl describe pod`
4. **Service Not Accessible**: Check service and endpoints with `kubectl get svc,endpoints`

## Cleanup

```bash
# Delete development environment
kubectl delete -k deployments/kubernetes/overlays/dev

# Or delete namespace (removes everything)
kubectl delete namespace boilerplate-dev
```

## Production Considerations

1. **Secrets Management**: Use external secrets manager (e.g., Sealed Secrets, External Secrets Operator)
2. **Ingress**: Set up Ingress controller for external access
3. **TLS**: Configure TLS certificates (Let's Encrypt, cert-manager)
4. **Storage**: Use persistent storage class appropriate for your cloud provider
5. **Monitoring**: Set up alerts and dashboards
6. **Backup**: Configure backup strategy for MongoDB
7. **Network Policies**: Implement network policies for security
8. **Resource Quotas**: Set namespace resource quotas
9. **RBAC**: Configure proper role-based access control
10. **GitOps**: Consider using ArgoCD or Flux for GitOps workflow
