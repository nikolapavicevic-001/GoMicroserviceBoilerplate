# Go Microservices Boilerplate

A production-ready monorepo boilerplate for building event-driven Go microservices with modern observability, authentication, and hypermedia-driven web UIs.

## Features

- **Event-Driven Architecture**: Kafka-based async communication
- **gRPC & REST**: gRPC for internal services, REST for external clients
- **MongoDB**: Repository pattern for easy database swapping
- **Authentication**: JWT tokens + OAuth2 (Google, GitHub)
- **Observability**: Zerolog, Prometheus, Jaeger
- **HTMX + Alpine.js**: Modern hypermedia-driven web UIs
- **Docker & Kubernetes**: Multiple deployment options
- **Go Workspace**: Monorepo with shared packages

## Project Structure

```
├── services/              # Microservices
│   ├── api-gateway/      # REST API Gateway (Echo)
│   ├── user-service/     # Example gRPC service
│   └── web-app/          # HTMX + Alpine.js web app
├── shared/               # Shared packages
│   ├── auth/             # JWT & OAuth2
│   ├── config/           # Configuration
│   ├── database/         # MongoDB client
│   ├── errors/           # Error handling
│   ├── events/           # Event schemas
│   ├── kafka/            # Kafka producer/consumer
│   ├── logger/           # Zerolog setup
│   └── proto/            # Protobuf definitions
├── deployments/          # Deployment configs
│   ├── docker-compose/   # Local development
│   └── kubernetes/       # K8s manifests
├── scripts/              # Build scripts
├── Makefile              # Common commands
└── CLAUDE.md             # AI-friendly documentation

```

## Quick Start

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make
- Protocol Buffers compiler (protoc)

### Setup

```bash
# Clone the repository
git clone <your-repo-url>
cd GoMicroserviceBoilerplate

# Copy environment variables
make setup

# Start infrastructure (MongoDB, Kafka, Jaeger, Prometheus)
make deps-up

# Generate protobuf code
make proto-gen

# Install dependencies
make deps-install

# Run tests
make test-unit

# Start all services
make up
```

### Accessing Services

- **API Gateway**: http://localhost:8080
- **Web App**: http://localhost:3000
- **Prometheus**: http://localhost:9090
- **Jaeger UI**: http://localhost:16686
- **Swagger Docs**: http://localhost:8080/swagger/index.html

## Development Workflow

### Working on a Single Service

```bash
# Start only dependencies
make deps-up

# Run a service locally
cd services/user-service
go run cmd/main.go
```

### Working on Web App

```bash
make deps-up
cd services/web-app
go run cmd/main.go
# Visit http://localhost:3000
```

### Full System

```bash
# Start everything
make up

# View logs
make logs
make logs-user
make logs-web

# Stop everything
make down
```

## Common Commands

See [`Makefile`](./Makefile) for all available commands:

```bash
make up               # Start all services
make down             # Stop all services
make deps-up          # Start infrastructure only
make proto-gen        # Generate protobuf code
make test-unit        # Run unit tests
make test-integration # Run integration tests
make lint             # Run linters
make docker-build     # Build Docker images
```

## Architecture

### Communication Patterns

1. **External → API Gateway (REST)**: Clients use HTTP/REST
2. **Browser → Web App (HTMX)**: Server-side rendered with dynamic updates
3. **Gateway → Services (gRPC)**: Type-safe internal communication
4. **Service → Service (Kafka)**: Event-driven async messaging
5. **Service → Database (Repository)**: Abstracted data access

### Example Services

- **user-service**: Complete backend microservice example
  - gRPC server
  - MongoDB repository pattern
  - Kafka event publishing
  - Full observability (logs, metrics, traces)

- **web-app**: HTMX + Alpine.js web application
  - Server-side rendering with Go templates
  - Login page (JWT + OAuth2)
  - Dashboard with dynamic content
  - Progressive enhancement

## Configuration

All services use environment variables. See [`.env.example`](./.env.example) for all available options.

### Environment Variables Format

```
{SERVICE}_{COMPONENT}_{PROPERTY}
```

Examples:
- `USER_SERVICE_GRPC_PORT=50051`
- `GATEWAY_HTTP_PORT=8080`
- `KAFKA_BROKERS=localhost:9092`

## Testing

```bash
# Unit tests (mocked dependencies)
make test-unit

# Integration tests (real MongoDB, Kafka via testcontainers)
make test-integration

# E2E tests (full system)
make test-e2e

# Coverage report
make test-coverage
```

## Deployment

### Docker Compose (Local)

```bash
make up              # Full stack
make deps-up         # Infrastructure only
```

### Kubernetes

```bash
# Deploy to Kubernetes
kubectl apply -k deployments/kubernetes/base/

# Production overlay
kubectl apply -k deployments/kubernetes/overlays/prod/
```

## Adding a New Service

```bash
make new-service NAME=notification-service
```

This creates:
- Service directory with standard structure
- Dockerfile
- go.mod
- Updates go.work and docker-compose.yml

## Documentation

- **[CLAUDE.md](./CLAUDE.md)**: Comprehensive documentation for AI assistants and developers
- **[Architecture Docs](./docs/architecture/)**: Architecture diagrams and decisions
- **[API Docs](http://localhost:8080/swagger/)**: Swagger/OpenAPI documentation

## Technology Stack

- **Language**: Go 1.22+
- **HTTP Framework**: Echo
- **gRPC**: Protocol Buffers
- **Database**: MongoDB (swappable)
- **Message Queue**: Kafka
- **Logging**: Zerolog
- **Metrics**: Prometheus
- **Tracing**: Jaeger
- **Frontend**: HTMX + Alpine.js
- **Auth**: JWT + OAuth2

## License

MIT License - see [LICENSE](./LICENSE) file for details

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## Support

For issues and questions:
- Create an issue on GitHub
- Check [CLAUDE.md](./CLAUDE.md) for detailed documentation
- Review example services in `services/user-service` and `services/web-app`
