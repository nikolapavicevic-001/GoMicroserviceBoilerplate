# Troubleshooting Guide

## Common Issues

### "No healthy upstream" on port 8080 (Envoy)

This means Envoy can't reach the backend services. Check:

1. **Services are running:**
```bash
docker-compose ps
```

2. **Web service is healthy:**
```bash
curl http://localhost:8081/health
```

3. **API service is healthy:**
```bash
curl http://localhost:8082/health
```

4. **Keycloak is ready:**
```bash
curl http://localhost:8090/health/ready
```

### Keycloak not accessible on port 8090

1. **Check Keycloak logs:**
```bash
docker logs keycloak
```

2. **Verify database connection:**
```bash
docker exec postgres-keycloak pg_isready -U keycloak
```

3. **Wait for Keycloak to fully start** (can take 60-90 seconds):
```bash
docker logs -f keycloak
# Wait for "Listening on" message
```

### Realm not automatically imported

The realm should import automatically on first startup with `--import-realm` flag.

If it doesn't:

1. **Check if realm already exists:**
   - Access http://localhost:8090
   - Login with admin/admin
   - Check if "microservice" realm exists

2. **Manually import:**
   - Go to Keycloak admin console
   - Click "Add realm"
   - Click "Import" and select `auth/keycloak/realm-config.json`

3. **Force re-import:**
```bash
# Stop Keycloak
docker-compose stop keycloak

# Remove Keycloak database volume (WARNING: deletes all data)
docker volume rm microserviceboilerplate_keycloak-data

# Restart
docker-compose up -d keycloak
```

## Service Startup Order

Services start in this order:
1. PostgreSQL databases (with health checks)
2. Keycloak (waits for postgres-keycloak to be healthy)
3. Web service (waits for Keycloak to be healthy)
4. API service (waits for postgres and nats)
5. Envoy (waits for all services)

## Health Check Endpoints

- **Keycloak:** http://localhost:8090/health/ready
- **Web Service:** http://localhost:8081/health
- **API Service:** http://localhost:8082/health
- **Envoy:** http://localhost:8080 (admin interface on :9901)

## Rebuilding Services

If you make code changes:

```bash
cd infrastructure
docker-compose up -d --build
```

## Viewing Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f keycloak
docker-compose logs -f web-service
docker-compose logs -f envoy
```

## Resetting Everything

```bash
cd infrastructure
docker-compose down -v  # Removes volumes too
docker-compose up -d --build
```

