# Microservice Boilerplate

A complete cloud-native microservices boilerplate with Envoy + Keycloak gateway, Go services with gRPC, NATS event streaming, Flink data processing, PostgreSQL storage, and HTMX web app.

## Architecture

```
External Clients
    ↓
Envoy Gateway (JWT validation, gRPC routing)
    ↓
Keycloak (OAuth2/OIDC Authentication)
    ↓
Go Services (gRPC)
    ↓
NATS (Event Broker)
    ↓
Apache Flink (Streaming + Batch Processing)
    ↓
PostgreSQL + MinIO (Data Lake)
```

## Components

- **Envoy Gateway**: API gateway with JWT validation and gRPC routing
- **Keycloak**: Identity and access management (OAuth2/OIDC)
- **Go Web Service**: HTMX-based web application with custom authentication UI
- **Go gRPC Services**: Microservices with gRPC APIs
- **NATS**: Lightweight event streaming and messaging
- **Apache Flink**: Unified streaming and batch data processing
- **PostgreSQL**: Main database for transactional data
- **MinIO**: S3-compatible data lake for historical data
- **Prometheus + Grafana**: Monitoring and observability

## Features

- ✅ Custom HTMX UI for authentication (not Keycloak default theme)
- ✅ OAuth2/OIDC with JWT tokens
- ✅ gRPC service communication
- ✅ Event-driven architecture with NATS
- ✅ Real-time and batch data processing with Flink
- ✅ Cloud-native (Kubernetes-ready)
- ✅ Easy ClickHouse integration (documented)

## Quick Start

### Local Development (Docker Compose)

1. **Start all services:**
```bash
cd infrastructure
docker-compose up -d
```

2. **Initialize Keycloak:**
   - Access Keycloak admin console: http://localhost:8090
   - Login with `admin/admin`
   - Import realm: Go to "Add realm" → Import `auth/keycloak/realm-config.json`

3. **Access services:**
   - Web App: http://localhost:8080
   - Keycloak: http://localhost:8090
   - Envoy Gateway: http://localhost:8080
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000 (admin/admin)
   - MinIO Console: http://localhost:9001 (minioadmin/minioadmin)
   - Flink UI: http://localhost:8081

### Kubernetes Deployment

1. **Create namespaces:**
```bash
kubectl apply -f infrastructure/kubernetes/namespaces/
```

2. **Deploy infrastructure:**
```bash
kubectl apply -f infrastructure/kubernetes/data/
kubectl apply -f infrastructure/kubernetes/messaging/
kubectl apply -f infrastructure/kubernetes/monitoring/
```

3. **Deploy auth and gateway:**
```bash
kubectl apply -f infrastructure/kubernetes/auth/
kubectl apply -f infrastructure/kubernetes/gateway/
```

4. **Deploy services:**
```bash
kubectl apply -f services/web/kubernetes/
kubectl apply -f services/api/kubernetes/
```

5. **Initialize Keycloak** (same as local setup)

## Project Structure

```
microserviceboilerplate/
├── infrastructure/
│   ├── docker-compose.yml          # Local development
│   ├── kubernetes/                 # K8s manifests
│   ├── prometheus/                 # Prometheus config
│   └── grafana/                    # Grafana config
├── services/
│   ├── web/                        # Go + HTMX web app
│   ├── api/                        # Go gRPC service
│   └── data-processor/             # Flink job
├── gateway/
│   └── envoy/                      # Envoy configuration
├── auth/
│   └── keycloak/                   # Keycloak realm config
├── proto/                          # gRPC proto definitions
└── scripts/                        # Deployment scripts
```

## Development

### Web Service (Go + HTMX)

```bash
cd services/web
go mod download
go run main.go
```

### API Service (gRPC)

```bash
cd services/api
make proto  # Generate proto code
go run main.go
```

### Flink Job

```bash
cd services/data-processor
mvn clean package
# Submit to Flink cluster
```

## Authentication Flow

1. User visits web app → redirected to `/auth/login`
2. Custom HTMX login form → submits to Go web service
3. Web service authenticates with Keycloak (OAuth2 password grant or authorization code)
4. Keycloak returns JWT token
5. Web service stores JWT in HTTP-only cookie
6. Subsequent requests include JWT → Envoy validates with Keycloak
7. Envoy routes to gRPC services with validated JWT

## Adding ClickHouse

The architecture is designed to easily add ClickHouse for analytical queries.

### Step 1: Add ClickHouse to Infrastructure

**Docker Compose** (`infrastructure/docker-compose.yml`):
```yaml
clickhouse:
  image: clickhouse/clickhouse-server:latest
  ports:
    - "8123:8123"
    - "9000:9000"
  volumes:
    - clickhouse-data:/var/lib/clickhouse
  networks:
    - microservice-network
```

**Kubernetes** (`infrastructure/kubernetes/data/clickhouse-deployment.yaml`):
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: clickhouse
  namespace: data
# ... (see ClickHouse K8s operator docs)
```

### Step 2: Add ClickHouse Sink to Flink

Update `services/data-processor/pom.xml`:
```xml
<dependency>
    <groupId>com.clickhouse</groupId>
    <artifactId>clickhouse-jdbc</artifactId>
    <version>0.6.0</version>
</dependency>
```

Update `DataProcessor.java` to add ClickHouse sink:
```java
// Add ClickHouse sink similar to PostgreSQL sink
parsedEvents.addSink(JdbcSink.sink(
    "INSERT INTO events (user_id, event_type, payload, created_at) VALUES",
    // ... ClickHouse connection options
));
```

### Step 3: Query ClickHouse from Services

Add ClickHouse client to Go services:
```go
import "github.com/ClickHouse/clickhouse-go/v2"

conn, err := clickhouse.Open(&clickhouse.Options{
    Addr: []string{"clickhouse:9000"},
    Auth: clickhouse.Auth{
        Database: "default",
        Username: "default",
        Password: "",
    },
})
```

### Benefits

- **Hot Storage**: PostgreSQL for transactional queries
- **Analytical Storage**: ClickHouse for fast aggregations
- **Data Lake**: MinIO for long-term storage
- **Easy Migration**: Architecture supports all three

## Configuration

### Environment Variables

**Web Service:**
- `KEYCLOAK_URL`: Keycloak server URL
- `KEYCLOAK_REALM`: Keycloak realm name
- `KEYCLOAK_CLIENT_ID`: OAuth2 client ID
- `KEYCLOAK_CLIENT_SECRET`: OAuth2 client secret
- `PORT`: HTTP server port

**API Service:**
- `DATABASE_URL`: PostgreSQL connection string
- `NATS_URL`: NATS server URL

**Flink Job:**
- `POSTGRES_URL`: PostgreSQL connection string
- `S3_ENDPOINT`: MinIO endpoint
- `S3_ACCESS_KEY`: MinIO access key
- `S3_SECRET_KEY`: MinIO secret key

## Monitoring

- **Prometheus**: Metrics collection at http://localhost:9090
- **Grafana**: Dashboards at http://localhost:3000
- **Flink UI**: Job monitoring at http://localhost:8081

## Security Notes

- JWT tokens stored in HTTP-only cookies
- Envoy validates all JWT tokens with Keycloak
- gRPC services receive validated JWT in headers
- Use HTTPS in production (set `Secure: true` in session options)

## Troubleshooting

### Keycloak not accessible
- Check Keycloak is running: `docker ps | grep keycloak`
- Verify realm is imported
- Check OAuth2 client configuration

### Envoy JWT validation fails
- Verify Keycloak URL in Envoy config
- Check JWT issuer matches Keycloak realm
- Ensure Keycloak is accessible from Envoy

### Services can't connect to NATS
- Verify NATS is running
- Check NATS URL in service configuration
- Ensure services are on same network

## License

MIT

