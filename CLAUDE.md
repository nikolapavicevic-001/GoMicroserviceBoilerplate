# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Go Microservices Boilerplate

## Project Overview

A production-ready monorepo boilerplate for building event-driven Go microservices with modern observability, authentication, and hypermedia-driven web UIs.

**Key Characteristics:**
- **Language**: Go 1.22+
- **Architecture**: Event-driven microservices with clean architecture
- **Communication**: REST API Gateway (external) + gRPC (internal services) + Kafka (async events)
- **Storage**: MongoDB with repository pattern (easily swappable to PostgreSQL/MySQL)
- **Deployment**: Docker Compose (local) + Kubernetes (production)
- **Observability**: Zerolog (logging) + Prometheus (metrics) + Jaeger (tracing)
- **Authentication**: JWT tokens + OAuth2 (Google, GitHub, etc.)
- **Frontend**: HTMX + Alpine.js for hypermedia-driven UIs

**Example Services Included:**
- `user-service`: Full backend microservice demonstrating all patterns
- `web-app`: HTMX + Alpine.js application with login and dashboard

---

## Architecture Overview

### High-Level System Design

```
┌─────────────────┐
│  Web Browsers   │ (HTML/HTMX requests)
└────────┬────────┘
         │
    ┌────▼────────┐
    │  Web App    │ (Port 3000)
    │  Service    │ HTMX + Alpine.js
    └────────┬────┘
             │ HTTP/gRPC
             │
┌────────────┼────────────┐
│  External  │  Internal  │
│  Clients   │  Requests  │
└─────┬──────┴──────┬─────┘
      │ HTTP/REST   │
      ▼             ▼
┌──────────────────────┐
│   API Gateway        │ (Port 8080)
│   - Echo Framework   │
│   - Auth Middleware  │
│   - Rate Limiting    │
│   - Swagger Docs     │
└──────────┬───────────┘
           │ gRPC (internal)
           │
    ┌──────┴──────┬──────────┬──────────┐
    ▼             ▼          ▼          ▼
┌─────────┐  ┌─────────┐ ┌─────────┐ ┌─────────┐
│  User   │  │ Order   │ │ Product │ │ Payment │
│ Service │  │ Service │ │ Service │ │ Service │
│:50051   │  │:50052   │ │:50053   │ │:50054   │
└────┬────┘  └────┬────┘ └────┬────┘ └────┬────┘
     │            │           │           │
     └────────────┴───────────┴───────────┘
                  │
           ┌──────▼──────────┐
           │  Kafka Cluster  │
           │  - user.events  │
           │  - order.events │
           │  - payment.*    │
           └──────┬──────────┘
                  │
       ┌──────────┴──────────┐
       ▼                     ▼
┌─────────────┐      ┌─────────────┐
│  MongoDB    │      │  MongoDB    │
│  (Users DB) │      │  (Orders DB)│
└─────────────┘      └─────────────┘

Observability Stack:
┌────────────┬────────────┬────────────┐
│  Zerolog   │ Prometheus │   Jaeger   │
│  (Logs)    │ (Metrics)  │  (Traces)  │
└────────────┴────────────┴────────────┘
```

### Communication Patterns

#### 1. External → API Gateway (REST/HTTP)
- Clients send HTTP requests to API Gateway
- Gateway handles: authentication, rate limiting, request validation, Swagger docs
- Echo framework with middleware for observability

#### 2. Browser → Web App (HTMX)
- Server-side rendered HTML with Go templates
- HTMX for dynamic partial updates without full page reloads
- Alpine.js for client-side interactivity and state
- Session-based authentication with cookies

#### 3. API Gateway → Services (gRPC)
- Gateway translates REST to gRPC calls
- Protobuf definitions in `shared/proto/`
- Type safety, performance, streaming support
- gRPC interceptors for logging, auth, tracing

#### 4. Service → Service (Kafka Events)
- Asynchronous event-driven communication
- Example: Order Service publishes "OrderCreated" → Payment Service consumes
- Event schemas in `shared/events/`
- At-least-once delivery, idempotent consumers

#### 5. Service → Database (Repository Pattern)
- Each service owns its database (database-per-service)
- Repository interface abstracts storage implementation
- Easy to swap MongoDB for PostgreSQL/MySQL

### Design Principles

1. **Loose Coupling**: Services communicate via well-defined interfaces (gRPC, Kafka)
2. **Single Responsibility**: Each service handles one business domain
3. **Database Per Service**: No shared databases between services
4. **Event-Driven**: Async communication for eventual consistency
5. **API Gateway Pattern**: Single entry point for external clients
6. **Hypermedia-Driven**: Web UIs use HTMX for progressive enhancement

---

## Project Structure

```
/home/nikola/personal/GoMicroserviceBoilerplate/
├── services/                           # All microservices
│   ├── api-gateway/                   # HTTP REST API Gateway
│   │   ├── cmd/
│   │   │   └── main.go                # Entry point
│   │   ├── internal/
│   │   │   ├── handler/               # HTTP handlers (Echo)
│   │   │   ├── middleware/            # Auth, logging, metrics
│   │   │   ├── grpc/                  # gRPC client connections
│   │   │   └── router/                # Route definitions
│   │   ├── config/
│   │   ├── docs/                      # Swagger/OpenAPI
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   ├── user-service/                  # EXAMPLE: User management service
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── domain/                # Business entities
│   │   │   │   └── user.go
│   │   │   ├── repository/            # Data access layer
│   │   │   │   ├── repository.go      # Interface
│   │   │   │   └── mongodb/
│   │   │   │       └── user_repo.go   # MongoDB implementation
│   │   │   ├── service/               # Business logic
│   │   │   │   └── user_service.go
│   │   │   ├── grpc/                  # gRPC server
│   │   │   │   └── user_handler.go
│   │   │   └── events/                # Kafka producers/consumers
│   │   │       ├── producer.go
│   │   │       └── consumer.go
│   │   ├── config/
│   │   ├── migrations/                # Database migrations
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   ├── web-app/                       # EXAMPLE: HTMX + Alpine.js Web App
│   │   ├── cmd/
│   │   │   └── main.go
│   │   ├── internal/
│   │   │   ├── handler/               # HTTP handlers
│   │   │   │   ├── auth.go            # Login, OAuth2 callbacks
│   │   │   │   ├── dashboard.go       # Dashboard pages
│   │   │   │   └── partials.go        # HTMX partial responses
│   │   │   ├── service/               # Business logic
│   │   │   ├── middleware/            # Session, auth, CSRF
│   │   │   └── session/               # Session management
│   │   ├── templates/                 # Go html/template files
│   │   │   ├── layouts/
│   │   │   │   └── base.html          # Base layout
│   │   │   ├── pages/
│   │   │   │   ├── login.html         # Login page
│   │   │   │   └── dashboard.html     # Dashboard
│   │   │   └── partials/              # HTMX partials
│   │   │       └── user_card.html
│   │   ├── static/                    # Static assets
│   │   │   ├── css/
│   │   │   ├── js/
│   │   │   │   └── alpine-plugins.js
│   │   │   └── images/
│   │   ├── config/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── [order-service, product-service, payment-service...]
│
├── shared/                            # Shared code across services
│   ├── proto/                         # Protobuf definitions
│   │   ├── user/
│   │   │   └── v1/
│   │   │       └── user.proto
│   │   ├── order/
│   │   │   └── v1/
│   │   │       └── order.proto
│   │   ├── common/
│   │   │   └── v1/
│   │   │       ├── pagination.proto
│   │   │       └── error.proto
│   │   └── gen/                       # Generated Go code
│   │       ├── user/
│   │       └── order/
│   │
│   ├── events/                        # Event schemas
│   │   ├── event.go                   # Base Event struct
│   │   ├── user_events.go
│   │   └── order_events.go
│   │
│   ├── logger/                        # Zerolog setup
│   │   └── logger.go
│   │
│   ├── config/                        # Shared config utilities
│   │   └── config.go
│   │
│   ├── errors/                        # Error handling
│   │   ├── errors.go                  # Custom error types
│   │   └── codes.go                   # Error codes
│   │
│   ├── middleware/                    # Shared middleware
│   │   ├── grpc/                      # gRPC interceptors
│   │   │   ├── logging.go
│   │   │   ├── recovery.go
│   │   │   ├── tracing.go             # Jaeger
│   │   │   └── metrics.go             # Prometheus
│   │   └── http/                      # HTTP middleware (Echo)
│   │       ├── logging.go
│   │       ├── cors.go
│   │       ├── tracing.go
│   │       └── metrics.go
│   │
│   ├── auth/                          # Authentication utilities
│   │   ├── jwt.go                     # JWT generation/validation
│   │   └── oauth2.go                  # OAuth2 providers
│   │
│   ├── database/                      # Database utilities
│   │   └── mongodb/
│   │       └── client.go
│   │
│   ├── kafka/                         # Kafka utilities
│   │   ├── producer.go
│   │   ├── consumer.go
│   │   └── config.go
│   │
│   └── testutil/                      # Testing utilities
│       ├── containers/                # Testcontainers
│       │   ├── mongodb.go
│       │   └── kafka.go
│       └── fixtures/                  # Test data
│
├── deployments/                       # Deployment configurations
│   ├── docker-compose/
│   │   ├── docker-compose.yml         # Full stack
│   │   ├── docker-compose.dev.yml     # Dev overrides (hot reload)
│   │   └── docker-compose.deps.yml    # Just infrastructure
│   │
│   └── kubernetes/
│       ├── base/                      # Base K8s configs
│       │   ├── api-gateway/
│       │   ├── user-service/
│       │   └── web-app/
│       ├── overlays/                  # Environment-specific
│       │   ├── dev/
│       │   ├── staging/
│       │   └── prod/
│       └── infrastructure/            # Infrastructure services
│           ├── mongodb/
│           ├── kafka/
│           ├── prometheus/
│           └── jaeger/
│
├── scripts/                           # Build and automation
│   ├── proto-gen.sh                   # Generate protobuf code
│   ├── db-migrate.sh                  # Run migrations
│   ├── test-integration.sh            # Integration tests
│   └── lint.sh                        # Linting
│
├── docs/                              # Additional documentation
│   ├── api/                           # API documentation
│   ├── architecture/                  # Architecture diagrams
│   └── guides/                        # How-to guides
│
├── Makefile                           # Common commands
├── go.work                            # Go workspace file
├── go.work.sum
├── .gitignore
├── .env.example                       # Example environment variables
├── README.md                          # Public-facing README
└── CLAUDE.md                          # This file
```

### Directory Responsibilities

- **services/**: Each subdirectory is an independently deployable microservice with its own go.mod
- **shared/**: Common code used across services via Go workspace. Never contains business logic.
- **deployments/**: Infrastructure as code (docker-compose for local, K8s for prod)
- **scripts/**: Automation for common tasks

---

## Common Commands (Makefile Reference)

### Service Management
```bash
make up                  # Start all services via Docker Compose
make down                # Stop all services
make restart             # Restart all services
make logs                # View logs from all services
make logs-gateway        # View API Gateway logs
make logs-user           # View user-service logs
make logs-web            # View web-app logs
make ps                  # Show status of all services
```

### Dependencies (Infrastructure Only)
```bash
make deps-up             # Start MongoDB, Kafka, Zookeeper only (fast!)
make deps-down           # Stop infrastructure
make deps-restart        # Restart infrastructure
```

### Development
```bash
make proto-gen           # Generate Go code from .proto files
make deps-install        # Install Go dependencies for all services
make lint                # Run golangci-lint on all services
make fmt                 # Format all Go code
make tidy                # Run go mod tidy on all services
```

### Testing
```bash
make test-unit           # Run unit tests for all services
make test-integration    # Run integration tests (uses testcontainers)
make test-e2e            # Run end-to-end tests
make test-coverage       # Generate coverage report (HTML)
make test-watch          # Run tests in watch mode
```

### Database
```bash
make db-migrate          # Run migrations for all services
make db-migrate-user     # Run migrations for user-service only
make db-seed             # Seed databases with test data
make db-shell-user       # Open MongoDB shell for user DB
make db-reset            # Drop all databases and re-migrate
```

### Building
```bash
make build               # Build all services
make build-user          # Build user-service binary
make build-web           # Build web-app binary
make docker-build        # Build all Docker images
make docker-push         # Push images to registry
```

### Cleanup
```bash
make clean               # Remove built binaries
make clean-all           # Remove binaries, Docker volumes, images
make reset               # Nuclear option: clean everything and reset
```

### Utilities
```bash
make setup               # Initial setup (copy .env.example to .env)
make new-service NAME=x  # Scaffold new service
make health              # Check health of all services
make docs                # Generate/serve API documentation (Swagger)
```

---

## Development Workflow

### Initial Setup (First Time)

1. **Prerequisites**
   ```bash
   # Required
   - Go 1.22+ installed
   - Docker & Docker Compose installed
   - Make installed

   # Optional but recommended
   - kubectl (for K8s deployment)
   - grpcurl (for testing gRPC endpoints)
   - kafkacat/kcat (for Kafka debugging)
   ```

2. **Clone and Setup**
   ```bash
   cd /home/nikola/personal/GoMicroserviceBoilerplate

   # Copy environment template
   make setup
   # This creates .env from .env.example

   # Start dependencies (MongoDB, Kafka)
   make deps-up

   # Generate protobuf code
   make proto-gen

   # Install all service dependencies
   make deps-install
   ```

3. **Verify Setup**
   ```bash
   make test-unit        # Run all unit tests
   make health           # Check services can connect to deps
   ```

### Daily Development Scenarios

#### Scenario 1: Working on a Single Service

```bash
# Start dependencies only (MongoDB, Kafka) - fast startup
make deps-up

# Work on user-service
cd services/user-service

# Run the service locally
go run cmd/main.go
# Service runs on port defined in .env (USER_SERVICE_GRPC_PORT=50051)

# In another terminal, run tests with watch
make test-watch

# Make changes, tests re-run automatically
```

#### Scenario 2: Working on Web App (HTMX)

```bash
# Start dependencies
make deps-up

# Start backend services web app depends on
cd services/user-service && go run cmd/main.go &

# Run web app with hot reload
cd services/web-app
air  # or go run cmd/main.go

# Visit http://localhost:3000
# Edit templates/ files - changes reflected immediately
```

#### Scenario 3: Testing Full System Integration

```bash
# Start all services via Docker Compose
make up

# View logs from all services
make logs

# View logs from specific service
make logs-web

# Make a request to API Gateway
curl http://localhost:8080/api/v1/users

# Test web app
open http://localhost:3000

# Tear down when done
make down
```

#### Scenario 4: Working with Protobuf Changes

```bash
# 1. Edit .proto file
vim shared/proto/user/v1/user.proto

# 2. Regenerate Go code
make proto-gen

# 3. Restart affected services
make restart-user
make restart-gateway
```

#### Scenario 5: Adding a New Service

```bash
# Use scaffolding script
make new-service NAME=notification-service

# This creates:
# - services/notification-service/ with standard structure
# - Dockerfile
# - Makefile
# - Basic config files
# - Updates docker-compose.yml
# - Updates go.work
```

### Testing Strategy

#### Unit Tests
```bash
# Test specific service
cd services/user-service
go test ./internal/service/... -v

# Test all services
make test-unit

# With coverage
make test-coverage
open coverage.html
```

#### Integration Tests
```bash
# Uses testcontainers - spins up real MongoDB and Kafka
cd services/user-service
go test -tags=integration ./internal/repository/... -v

# Full integration test suite (all services)
make test-integration
```

#### E2E Tests
```bash
# Starts full stack, runs end-to-end tests
make test-e2e
```

### Debugging Tools

#### Debug gRPC Calls
```bash
# List all gRPC methods
grpcurl -plaintext localhost:50051 list

# Describe a service
grpcurl -plaintext localhost:50051 describe user.v1.UserService

# Call a gRPC method
grpcurl -plaintext -d '{"id": "123"}' \
  localhost:50051 user.v1.UserService/GetUser
```

#### Debug Kafka Messages
```bash
# Consume from topic
kafkacat -b localhost:9092 -t user.events -C

# Produce test message
echo '{"event":"test"}' | \
  kafkacat -b localhost:9092 -t user.events -P

# List topics
kafkacat -b localhost:9092 -L
```

#### Debug Database
```bash
# Connect to MongoDB
make db-shell-user
# Opens mongo shell connected to user-service database

# Or directly
mongosh mongodb://localhost:27017/users_db
```

---

## Key Conventions

### Error Handling

#### Service Layer
```go
// Use custom error types from shared/errors
import "GoMicroserviceBoilerplate/shared/errors"

func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, errors.ErrNotFound) {
            return nil, errors.NewNotFoundError("user", id)
        }
        return nil, errors.Wrap(err, "failed to get user")
    }
    return user, nil
}
```

#### gRPC Layer (services/user-service/internal/grpc/user_handler.go)
```go
// Convert domain errors to gRPC status codes
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    user, err := h.service.GetUser(ctx, req.Id)
    if err != nil {
        return nil, toGRPCError(err)
    }
    return &pb.GetUserResponse{User: toProtoUser(user)}, nil
}

func toGRPCError(err error) error {
    switch {
    case errors.IsNotFound(err):
        return status.Error(codes.NotFound, err.Error())
    case errors.IsValidation(err):
        return status.Error(codes.InvalidArgument, err.Error())
    case errors.IsAlreadyExists(err):
        return status.Error(codes.AlreadyExists, err.Error())
    default:
        return status.Error(codes.Internal, "internal server error")
    }
}
```

#### API Gateway Layer (services/api-gateway/internal/handler/user.go)
```go
// Convert gRPC errors to HTTP status codes
import "github.com/labstack/echo/v4"

func (h *UserHandler) GetUser(c echo.Context) error {
    id := c.Param("id")

    user, err := h.userClient.GetUser(c.Request().Context(), &pb.GetUserRequest{Id: id})
    if err != nil {
        return toHTTPError(c, err)
    }

    return c.JSON(http.StatusOK, user)
}

func toHTTPError(c echo.Context, err error) error {
    st, ok := status.FromError(err)
    if !ok {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "internal server error",
        })
    }

    switch st.Code() {
    case codes.NotFound:
        return c.JSON(http.StatusNotFound, map[string]string{"error": st.Message()})
    case codes.InvalidArgument:
        return c.JSON(http.StatusBadRequest, map[string]string{"error": st.Message()})
    case codes.AlreadyExists:
        return c.JSON(http.StatusConflict, map[string]string{"error": st.Message()})
    case codes.Unauthenticated:
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": st.Message()})
    case codes.PermissionDenied:
        return c.JSON(http.StatusForbidden, map[string]string{"error": st.Message()})
    default:
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "internal server error",
        })
    }
}
```

### Logging (Zerolog)

#### Standard Log Fields
```go
// shared/logger/logger.go - Zerolog setup
import (
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func New(level, format string) zerolog.Logger {
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

    logLevel, _ := zerolog.ParseLevel(level)
    zerolog.SetGlobalLevel(logLevel)

    if format == "console" {
        return log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
    }
    return zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func WithContext(ctx context.Context) zerolog.Logger {
    logger := log.Ctx(ctx)
    if logger.GetLevel() == zerolog.Disabled {
        return log.Logger
    }
    return *logger
}
```

#### Service-Level Logging
```go
// services/user-service/internal/service/user_service.go
import "GoMicroserviceBoilerplate/shared/logger"

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    log := logger.WithContext(ctx)

    log.Info().
        Str("email", req.Email).
        Str("action", "create_user").
        Msg("creating user")

    user, err := s.repo.Create(ctx, req)
    if err != nil {
        log.Error().
            Err(err).
            Str("email", req.Email).
            Str("action", "create_user").
            Msg("failed to create user")
        return nil, err
    }

    log.Info().
        Str("user_id", user.ID).
        Str("email", user.Email).
        Str("action", "create_user").
        Msg("user created successfully")

    return user, nil
}
```

#### Log Levels
- **Debug**: Detailed diagnostic information (disabled in production)
- **Info**: Significant application events (user created, order placed)
- **Warn**: Potentially harmful situations (retry logic triggered, deprecated API used)
- **Error**: Error events that require attention

### HTTP Framework (Echo)

#### Echo Setup with Middleware (services/api-gateway/cmd/main.go)
```go
import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    echoSwagger "github.com/swaggo/echo-swagger"
    "GoMicroserviceBoilerplate/shared/middleware/http"
)

func main() {
    e := echo.New()

    // Global middleware
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())
    e.Use(httpmw.LoggingMiddleware(logger))      // Zerolog
    e.Use(httpmw.TracingMiddleware())            // Jaeger
    e.Use(httpmw.MetricsMiddleware())            // Prometheus

    // Swagger documentation
    e.GET("/swagger/*", echoSwagger.WrapHandler)

    // API routes
    api := e.Group("/api/v1")
    api.Use(httpmw.AuthMiddleware(jwtValidator)) // JWT validation

    // User routes
    users := api.Group("/users")
    users.POST("", userHandler.CreateUser)
    users.GET("/:id", userHandler.GetUser)
    users.PUT("/:id", userHandler.UpdateUser)
    users.DELETE("/:id", userHandler.DeleteUser)

    e.Logger.Fatal(e.Start(":8080"))
}
```

#### Echo Handler Example
```go
// services/api-gateway/internal/handler/user.go

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User data"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(c echo.Context) error {
    var req CreateUserRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, ErrorResponse{Message: "invalid request"})
    }

    if err := c.Validate(&req); err != nil {
        return c.JSON(http.StatusBadRequest, ErrorResponse{Message: err.Error()})
    }

    // Call user-service via gRPC
    resp, err := h.userClient.CreateUser(c.Request().Context(), &pb.CreateUserRequest{
        Email: req.Email,
        Name:  req.Name,
    })
    if err != nil {
        return toHTTPError(c, err)
    }

    return c.JSON(http.StatusCreated, toUserResponse(resp.User))
}
```

### gRPC Interceptors

#### Server-Side Interceptors (services/user-service/cmd/main.go)
```go
import (
    "google.golang.org/grpc"
    grpcmw "GoMicroserviceBoilerplate/shared/middleware/grpc"
)

func main() {
    server := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            grpcmw.LoggingInterceptor(logger),    // Log all requests
            grpcmw.RecoveryInterceptor(),         // Recover from panics
            grpcmw.TracingInterceptor(),          // Jaeger tracing
            grpcmw.MetricsInterceptor(),          // Prometheus metrics
            grpcmw.ValidationInterceptor(),       // Validate requests
        ),
    )

    pb.RegisterUserServiceServer(server, userHandler)

    lis, _ := net.Listen("tcp", ":50051")
    server.Serve(lis)
}
```

#### Client-Side Interceptors (services/api-gateway/internal/grpc/client.go)
```go
func NewUserServiceClient(addr string) (pb.UserServiceClient, error) {
    conn, err := grpc.Dial(
        addr,
        grpc.WithInsecure(),
        grpc.WithUnaryInterceptor(
            grpcmw.ClientLoggingInterceptor(logger),
            grpcmw.ClientTracingInterceptor(),
        ),
    )
    if err != nil {
        return nil, err
    }

    return pb.NewUserServiceClient(conn), nil
}
```

### Event Schemas (Kafka)

#### Base Event Structure (shared/events/event.go)
```go
package events

import (
    "encoding/json"
    "time"

    "github.com/google/uuid"
)

// Event is the base structure for all events
type Event struct {
    ID          string    `json:"id"`           // UUID
    Type        string    `json:"type"`         // "user.created"
    AggregateID string    `json:"aggregate_id"` // Entity ID
    Timestamp   time.Time `json:"timestamp"`
    Version     int       `json:"version"`      // Schema version
    Data        []byte    `json:"data"`         // JSON payload
}

func NewEvent(eventType, aggregateID string, data interface{}) (*Event, error) {
    dataBytes, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }

    return &Event{
        ID:          uuid.New().String(),
        Type:        eventType,
        AggregateID: aggregateID,
        Timestamp:   time.Now().UTC(),
        Version:     1,
        Data:        dataBytes,
    }, nil
}
```

#### User Events (shared/events/user_events.go)
```go
package events

import "time"

const (
    UserCreated = "user.created"
    UserUpdated = "user.updated"
    UserDeleted = "user.deleted"
)

type UserCreatedData struct {
    UserID    string    `json:"user_id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}

func NewUserCreatedEvent(userID, email, name string) (*Event, error) {
    data := UserCreatedData{
        UserID:    userID,
        Email:     email,
        Name:      name,
        CreatedAt: time.Now().UTC(),
    }
    return NewEvent(UserCreated, userID, data)
}
```

#### Publishing Events (services/user-service/internal/service/user_service.go)
```go
func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    log := logger.WithContext(ctx)

    // 1. Persist to database
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        return nil, err
    }

    // 2. Publish event (best-effort)
    event, _ := events.NewUserCreatedEvent(user.ID, user.Email, user.Name)
    if err := s.eventProducer.Publish(ctx, "user.events", event); err != nil {
        log.Error().Err(err).Msg("failed to publish user created event")
        // Don't fail the request - event publishing is best-effort
    }

    return user, nil
}
```

#### Consuming Events (services/order-service/internal/events/consumer.go)
```go
func (h *UserEventHandler) HandleUserCreated(ctx context.Context, event *events.Event) error {
    log := logger.WithContext(ctx)

    // 1. Check for duplicate event (idempotency)
    if h.isDuplicate(ctx, event.ID) {
        log.Info().Str("event_id", event.ID).Msg("duplicate event, skipping")
        return nil
    }

    // 2. Unmarshal event data
    var data events.UserCreatedData
    if err := json.Unmarshal(event.Data, &data); err != nil {
        log.Error().Err(err).Msg("failed to unmarshal event")
        return err
    }

    // 3. Process event
    log.Info().
        Str("user_id", data.UserID).
        Str("email", data.Email).
        Msg("processing user created event")

    // Create user record in order service database
    if err := h.userRepo.Create(ctx, &User{
        ID:    data.UserID,
        Email: data.Email,
        Name:  data.Name,
    }); err != nil {
        log.Error().Err(err).Msg("failed to create user")
        return err
    }

    // 4. Mark as processed
    return h.markProcessed(ctx, event.ID)
}
```

### Repository Pattern

#### Interface Definition (services/user-service/internal/repository/repository.go)
```go
package repository

import (
    "context"
    "GoMicroserviceBoilerplate/services/user-service/internal/domain"
)

type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    FindByID(ctx context.Context, id string) (*domain.User, error)
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter *ListFilter) ([]*domain.User, int64, error)
}

type ListFilter struct {
    Limit  int
    Offset int
    Search string
}
```

#### MongoDB Implementation (services/user-service/internal/repository/mongodb/user_repo.go)
```go
package mongodb

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "GoMicroserviceBoilerplate/services/user-service/internal/domain"
    "GoMicroserviceBoilerplate/services/user-service/internal/repository"
    "GoMicroserviceBoilerplate/shared/errors"
)

type mongoUserRepository struct {
    collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) repository.UserRepository {
    return &mongoUserRepository{
        collection: db.Collection("users"),
    }
}

func (r *mongoUserRepository) Create(ctx context.Context, user *domain.User) error {
    _, err := r.collection.InsertOne(ctx, user)
    if err != nil {
        if mongo.IsDuplicateKeyError(err) {
            return errors.ErrAlreadyExists
        }
        return errors.Wrap(err, "failed to insert user")
    }
    return nil
}

func (r *mongoUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
    var user domain.User
    err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
    if err == mongo.ErrNoDocuments {
        return nil, errors.ErrNotFound
    }
    if err != nil {
        return nil, errors.Wrap(err, "failed to find user")
    }
    return &user, nil
}

func (r *mongoUserRepository) Update(ctx context.Context, user *domain.User) error {
    result, err := r.collection.UpdateOne(
        ctx,
        bson.M{"_id": user.ID},
        bson.M{"$set": user},
    )
    if err != nil {
        return errors.Wrap(err, "failed to update user")
    }
    if result.MatchedCount == 0 {
        return errors.ErrNotFound
    }
    return nil
}

func (r *mongoUserRepository) Delete(ctx context.Context, id string) error {
    result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        return errors.Wrap(err, "failed to delete user")
    }
    if result.DeletedCount == 0 {
        return errors.ErrNotFound
    }
    return nil
}

// EnsureIndexes creates required indexes
func (r *mongoUserRepository) EnsureIndexes(ctx context.Context) error {
    indexes := []mongo.IndexModel{
        {
            Keys:    bson.D{{Key: "email", Value: 1}},
            Options: options.Index().SetUnique(true),
        },
        {
            Keys: bson.D{{Key: "created_at", Value: -1}},
        },
    }

    _, err := r.collection.Indexes().CreateMany(ctx, indexes)
    return err
}
```

#### Swapping to PostgreSQL (Example)
```go
// services/user-service/internal/repository/postgres/user_repo.go
package postgres

import (
    "context"
    "database/sql"
    "GoMicroserviceBoilerplate/services/user-service/internal/domain"
    "GoMicroserviceBoilerplate/services/user-service/internal/repository"
    "GoMicroserviceBoilerplate/shared/errors"
)

type postgresUserRepository struct {
    db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
    return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
    var user domain.User
    err := r.db.QueryRowContext(ctx,
        "SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1",
        id,
    ).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)

    if err == sql.ErrNoRows {
        return nil, errors.ErrNotFound
    }
    if err != nil {
        return nil, errors.Wrap(err, "failed to find user")
    }
    return &user, nil
}

// Only change main.go dependency injection:
// FROM: repo := mongodb.NewUserRepository(mongoClient.Database(cfg.MongoDB))
// TO:   repo := postgres.NewUserRepository(pgClient)
```

### Configuration Management

#### Environment Variable Naming Convention

Format: `{SERVICE}_{COMPONENT}_{PROPERTY}`

Examples:
- `USER_SERVICE_GRPC_PORT` - User service gRPC port
- `GATEWAY_HTTP_PORT` - Gateway HTTP port
- `USER_SERVICE_MONGO_URI` - User service MongoDB connection
- `WEB_APP_SESSION_SECRET` - Web app session secret

#### Service Config Struct (services/user-service/config/config.go)
```go
package config

import "github.com/caarlos0/env/v10"

type Config struct {
    // Server
    GRPCPort int `env:"USER_SERVICE_GRPC_PORT" envDefault:"50051"`

    // Database
    MongoURI      string `env:"USER_SERVICE_MONGO_URI" envDefault:"mongodb://localhost:27017"`
    MongoDB       string `env:"USER_SERVICE_MONGO_DB" envDefault:"users_db"`
    MongoTimeout  int    `env:"USER_SERVICE_MONGO_TIMEOUT" envDefault:"10"`

    // Kafka
    KafkaBrokers string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
    KafkaGroupID string `env:"USER_SERVICE_KAFKA_GROUP" envDefault:"user-service"`

    // Logging
    LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
    LogFormat string `env:"LOG_FORMAT" envDefault:"json"`

    // Observability
    JaegerEndpoint   string `env:"JAEGER_ENDPOINT" envDefault:"http://localhost:14268/api/traces"`
    MetricsPort      int    `env:"USER_SERVICE_METRICS_PORT" envDefault:"9090"`

    // Environment
    Environment string `env:"ENVIRONMENT" envDefault:"development"`
}

func Load() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, err
    }

    // Validate
    if err := cfg.Validate(); err != nil {
        return nil, err
    }

    return cfg, nil
}

func (c *Config) Validate() error {
    if c.GRPCPort < 1024 || c.GRPCPort > 65535 {
        return fmt.Errorf("invalid GRPC port: %d", c.GRPCPort)
    }
    // ... more validation
    return nil
}
```

#### Example .env File
```bash
# .env (for local development - gitignored)

# ====================================
# Global Settings
# ====================================
ENVIRONMENT=development
LOG_LEVEL=debug
LOG_FORMAT=console

# ====================================
# Infrastructure
# ====================================
MONGO_URI=mongodb://localhost:27017
KAFKA_BROKERS=localhost:9092

# Observability
JAEGER_ENDPOINT=http://localhost:14268/api/traces
PROMETHEUS_ENDPOINT=http://localhost:9090

# ====================================
# API Gateway
# ====================================
GATEWAY_HTTP_PORT=8080
GATEWAY_TIMEOUT=30s
GATEWAY_USER_SERVICE_ADDR=localhost:50051
GATEWAY_ORDER_SERVICE_ADDR=localhost:50052

# JWT Configuration
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY=24h

# OAuth2 Configuration
OAUTH2_GOOGLE_CLIENT_ID=your-google-client-id
OAUTH2_GOOGLE_CLIENT_SECRET=your-google-client-secret
OAUTH2_GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback

OAUTH2_GITHUB_CLIENT_ID=your-github-client-id
OAUTH2_GITHUB_CLIENT_SECRET=your-github-client-secret
OAUTH2_GITHUB_REDIRECT_URL=http://localhost:8080/auth/github/callback

# ====================================
# User Service
# ====================================
USER_SERVICE_GRPC_PORT=50051
USER_SERVICE_MONGO_URI=${MONGO_URI}
USER_SERVICE_MONGO_DB=users_db
USER_SERVICE_KAFKA_BROKERS=${KAFKA_BROKERS}
USER_SERVICE_METRICS_PORT=9091

# ====================================
# Web App
# ====================================
WEB_APP_HTTP_PORT=3000
WEB_APP_SESSION_SECRET=change-this-in-production
WEB_APP_COOKIE_SECURE=false  # true in production
WEB_APP_USER_SERVICE_ADDR=localhost:50051
```

### Authentication

#### JWT Token Generation (shared/auth/jwt.go)
```go
package auth

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    jwt.RegisteredClaims
}

func GenerateToken(userID, email, secret string, expiry time.Duration) (string, error) {
    claims := Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "boilerplate-api",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }

    return nil, jwt.ErrSignatureInvalid
}
```

#### JWT Middleware (services/api-gateway/internal/middleware/auth.go)
```go
func AuthMiddleware(secret string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // Extract token from Authorization header
            authHeader := c.Request().Header.Get("Authorization")
            if authHeader == "" {
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "error": "missing authorization header",
                })
            }

            // Extract Bearer token
            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "error": "invalid authorization header format",
                })
            }

            // Validate token
            claims, err := auth.ValidateToken(parts[1], secret)
            if err != nil {
                return c.JSON(http.StatusUnauthorized, map[string]string{
                    "error": "invalid token",
                })
            }

            // Set user context
            c.Set("user_id", claims.UserID)
            c.Set("email", claims.Email)

            return next(c)
        }
    }
}
```

#### OAuth2 Integration (shared/auth/oauth2.go)
```go
package auth

import (
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "golang.org/x/oauth2/github"
)

type OAuth2Provider struct {
    Config *oauth2.Config
}

func NewGoogleProvider(clientID, clientSecret, redirectURL string) *OAuth2Provider {
    return &OAuth2Provider{
        Config: &oauth2.Config{
            ClientID:     clientID,
            ClientSecret: clientSecret,
            RedirectURL:  redirectURL,
            Scopes: []string{
                "https://www.googleapis.com/auth/userinfo.email",
                "https://www.googleapis.com/auth/userinfo.profile",
            },
            Endpoint: google.Endpoint,
        },
    }
}

func NewGitHubProvider(clientID, clientSecret, redirectURL string) *OAuth2Provider {
    return &OAuth2Provider{
        Config: &oauth2.Config{
            ClientID:     clientID,
            ClientSecret: clientSecret,
            RedirectURL:  redirectURL,
            Scopes:       []string{"user:email"},
            Endpoint:     github.Endpoint,
        },
    }
}

func (p *OAuth2Provider) GetAuthURL(state string) string {
    return p.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (p *OAuth2Provider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
    return p.Config.Exchange(ctx, code)
}
```

#### OAuth2 Handler (services/api-gateway/internal/handler/auth.go)
```go
func (h *AuthHandler) GoogleLogin(c echo.Context) error {
    state := generateRandomState()

    // Store state in session for validation
    sess, _ := session.Get("auth", c)
    sess.Values["oauth_state"] = state
    sess.Save(c.Request(), c.Response())

    url := h.googleProvider.GetAuthURL(state)
    return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c echo.Context) error {
    // Validate state
    sess, _ := session.Get("auth", c)
    savedState := sess.Values["oauth_state"]
    if c.QueryParam("state") != savedState {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "invalid state parameter",
        })
    }

    // Exchange code for token
    code := c.QueryParam("code")
    token, err := h.googleProvider.ExchangeCode(c.Request().Context(), code)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "failed to exchange code",
        })
    }

    // Get user info from Google
    userInfo, err := h.getUserInfoFromGoogle(token)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "failed to get user info",
        })
    }

    // Create or update user in database (via user-service)
    user, err := h.createOrUpdateUser(c.Request().Context(), userInfo)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "failed to create user",
        })
    }

    // Generate JWT token
    jwtToken, err := auth.GenerateToken(user.ID, user.Email, h.jwtSecret, 24*time.Hour)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "failed to generate token",
        })
    }

    return c.JSON(http.StatusOK, map[string]string{
        "token": jwtToken,
    })
}
```

### Web App (HTMX + Alpine.js)

#### Template Structure (services/web-app/templates/)

**Base Layout** (templates/layouts/base.html):
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}Microservices App{{end}}</title>

    <!-- HTMX -->
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>

    <!-- Alpine.js -->
    <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>

    <!-- TailwindCSS (or your CSS framework) -->
    <link href="/static/css/styles.css" rel="stylesheet">

    {{block "head" .}}{{end}}
</head>
<body>
    {{block "nav" .}}
    <nav class="navbar">
        <a href="/">Home</a>
        {{if .User}}
            <a href="/dashboard">Dashboard</a>
            <a href="/logout">Logout</a>
        {{else}}
            <a href="/login">Login</a>
        {{end}}
    </nav>
    {{end}}

    <main>
        {{block "content" .}}{{end}}
    </main>

    {{block "scripts" .}}{{end}}
</body>
</html>
```

**Login Page** (templates/pages/login.html):
```html
{{template "layouts/base.html" .}}

{{define "title"}}Login{{end}}

{{define "content"}}
<div class="login-container">
    <h1>Login</h1>

    <!-- Standard Login Form -->
    <form hx-post="/auth/login"
          hx-target="#error-message"
          hx-swap="innerHTML">

        <input type="email" name="email" placeholder="Email" required>
        <input type="password" name="password" placeholder="Password" required>

        <button type="submit">Login</button>
    </form>

    <div id="error-message" class="error"></div>

    <!-- OAuth2 Login Buttons -->
    <div class="oauth-buttons">
        <a href="/auth/google" class="btn-google">
            Sign in with Google
        </a>
        <a href="/auth/github" class="btn-github">
            Sign in with GitHub
        </a>
    </div>
</div>
{{end}}
```

**Dashboard Page** (templates/pages/dashboard.html):
```html
{{template "layouts/base.html" .}}

{{define "title"}}Dashboard{{end}}

{{define "content"}}
<div class="dashboard" x-data="dashboard()">
    <h1>Welcome, {{.User.Name}}!</h1>

    <!-- User Stats with HTMX polling -->
    <div hx-get="/api/stats"
         hx-trigger="every 5s"
         hx-swap="innerHTML">
        {{template "partials/stats.html" .Stats}}
    </div>

    <!-- User List with Alpine.js state -->
    <div class="user-list">
        <h2>Users</h2>

        <!-- Search box with HTMX -->
        <input type="search"
               name="q"
               placeholder="Search users..."
               hx-get="/api/users/search"
               hx-trigger="keyup changed delay:300ms"
               hx-target="#user-results"
               hx-swap="innerHTML">

        <div id="user-results">
            {{template "partials/user_list.html" .Users}}
        </div>
    </div>

    <!-- Infinite Scroll -->
    <div hx-get="/api/posts?page=2"
         hx-trigger="revealed"
         hx-swap="afterend">
        {{template "partials/post_list.html" .Posts}}
    </div>
</div>
{{end}}

{{define "scripts"}}
<script>
function dashboard() {
    return {
        notifications: [],

        init() {
            // Alpine.js initialization
            this.fetchNotifications();
        },

        fetchNotifications() {
            fetch('/api/notifications')
                .then(r => r.json())
                .then(data => this.notifications = data);
        }
    }
}
</script>
{{end}}
```

**Partial Template** (templates/partials/user_card.html):
```html
<div class="user-card" id="user-{{.ID}}">
    <img src="{{.Avatar}}" alt="{{.Name}}">
    <h3>{{.Name}}</h3>
    <p>{{.Email}}</p>

    <button hx-delete="/api/users/{{.ID}}"
            hx-confirm="Are you sure?"
            hx-target="#user-{{.ID}}"
            hx-swap="outerHTML swap:1s">
        Delete
    </button>
</div>
```

#### Handler with HTMX Support (services/web-app/internal/handler/dashboard.go)
```go
package handler

import (
    "net/http"
    "github.com/labstack/echo/v4"
)

type DashboardHandler struct {
    userService UserService
}

func (h *DashboardHandler) Dashboard(c echo.Context) error {
    // Get user from session
    sess, _ := session.Get("session", c)
    userID := sess.Values["user_id"].(string)

    user, _ := h.userService.GetUser(c.Request().Context(), userID)
    users, _ := h.userService.ListUsers(c.Request().Context())

    return c.Render(http.StatusOK, "pages/dashboard.html", map[string]interface{}{
        "User":  user,
        "Users": users,
    })
}

// HTMX partial response
func (h *DashboardHandler) SearchUsers(c echo.Context) error {
    query := c.QueryParam("q")

    users, _ := h.userService.SearchUsers(c.Request().Context(), query)

    // Detect HTMX request
    if c.Request().Header.Get("HX-Request") == "true" {
        // Return partial template only
        return c.Render(http.StatusOK, "partials/user_list.html", users)
    }

    // Return full page for non-HTMX requests
    return c.Render(http.StatusOK, "pages/search.html", map[string]interface{}{
        "Users": users,
        "Query": query,
    })
}

// HTMX delete with swap
func (h *DashboardHandler) DeleteUser(c echo.Context) error {
    userID := c.Param("id")

    if err := h.userService.DeleteUser(c.Request().Context(), userID); err != nil {
        return c.String(http.StatusInternalServerError, "Failed to delete user")
    }

    // Return empty response (HTMX will remove the element with hx-swap)
    return c.NoContent(http.StatusOK)
}
```

#### Session Middleware (services/web-app/internal/middleware/session.go)
```go
package middleware

import (
    "github.com/gorilla/sessions"
    "github.com/labstack/echo/v4"
)

var store *sessions.CookieStore

func InitSessionStore(secret string) {
    store = sessions.NewCookieStore([]byte(secret))
    store.Options = &sessions.Options{
        Path:     "/",
        MaxAge:   86400 * 7, // 7 days
        HttpOnly: true,
        Secure:   false, // true in production
        SameSite: http.SameSiteLaxMode,
    }
}

func SessionMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            return next(c)
        }
    }
}

func AuthRequired() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            sess, _ := session.Get("session", c)

            if sess.Values["user_id"] == nil {
                return c.Redirect(http.StatusTemporaryRedirect, "/login")
            }

            return next(c)
        }
    }
}
```

---

## Testing Approach

### Test Pyramid

```
       /\
      /E2E\         <- Few (Full system, browser tests)
     /------\
    /  INT  \       <- More (Service + DB + Kafka)
   /----------\
  /   UNIT     \    <- Most (Pure business logic)
 /--------------\
```

### Unit Tests

**What to Test:**
- Business logic in service layer
- Domain entity methods
- Utility functions
- Validation logic

**Example**: Service Layer Test (services/user-service/internal/service/user_service_test.go)
```go
package service_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock repository
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.User), args.Error(1)
}

func TestUserService_GetUser(t *testing.T) {
    // Arrange
    mockRepo := new(MockUserRepository)
    mockProducer := new(MockEventProducer)
    service := NewUserService(mockRepo, mockProducer)

    expectedUser := &domain.User{
        ID:    "123",
        Email: "test@example.com",
        Name:  "Test User",
    }

    mockRepo.On("FindByID", mock.Anything, "123").Return(expectedUser, nil)

    // Act
    user, err := service.GetUser(context.Background(), "123")

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedUser.Email, user.Email)
    mockRepo.AssertExpectations(t)
}

func TestUserService_GetUser_NotFound(t *testing.T) {
    mockRepo := new(MockUserRepository)
    mockProducer := new(MockEventProducer)
    service := NewUserService(mockRepo, mockProducer)

    mockRepo.On("FindByID", mock.Anything, "999").Return(nil, errors.ErrNotFound)

    user, err := service.GetUser(context.Background(), "999")

    assert.Error(t, err)
    assert.Nil(t, user)
    assert.True(t, errors.IsNotFound(err))
}
```

**Running Unit Tests:**
```bash
cd services/user-service
go test ./internal/service/... -v
```

### Integration Tests

**What to Test:**
- Repository layer with real database
- Event producer/consumer with real Kafka
- gRPC handler with real service

**Example**: Repository Integration Test (services/user-service/internal/repository/mongodb/user_repo_integration_test.go)
```go
//go:build integration
package mongodb_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/mongodb"
)

func TestUserRepository_Integration(t *testing.T) {
    // Start MongoDB container
    ctx := context.Background()
    mongoContainer, err := mongodb.RunContainer(ctx,
        testcontainers.WithImage("mongo:7"),
    )
    require.NoError(t, err)
    defer mongoContainer.Terminate(ctx)

    // Get connection string
    connStr, err := mongoContainer.ConnectionString(ctx)
    require.NoError(t, err)

    // Create repository
    client, _ := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
    db := client.Database("test_db")
    repo := NewUserRepository(db)

    // Test Create
    user := &domain.User{
        ID:    "123",
        Email: "test@example.com",
        Name:  "Test User",
    }
    err = repo.Create(ctx, user)
    assert.NoError(t, err)

    // Test FindByID
    found, err := repo.FindByID(ctx, "123")
    assert.NoError(t, err)
    assert.Equal(t, user.Email, found.Email)

    // Test Update
    found.Name = "Updated Name"
    err = repo.Update(ctx, found)
    assert.NoError(t, err)

    updated, _ := repo.FindByID(ctx, "123")
    assert.Equal(t, "Updated Name", updated.Name)

    // Test Delete
    err = repo.Delete(ctx, "123")
    assert.NoError(t, err)

    // Verify deleted
    _, err = repo.FindByID(ctx, "123")
    assert.ErrorIs(t, err, errors.ErrNotFound)
}
```

**Running Integration Tests:**
```bash
# With testcontainers (slower but isolated)
cd services/user-service
go test -tags=integration ./internal/repository/... -v

# Or use make target
make test-integration
```

### E2E Tests

**What to Test:**
- Full user journeys through API Gateway
- Multi-service workflows
- HTMX web app interactions

**Example**: E2E Test (tests/e2e/user_flow_test.go)
```go
//go:build e2e
package e2e_test

import (
    "testing"
    "net/http"
    "github.com/stretchr/testify/assert"
)

func TestUserCreationFlow(t *testing.T) {
    // Assumes all services are running (via docker-compose)
    gatewayURL := "http://localhost:8080"

    // 1. Create user via API Gateway
    resp := createUser(t, gatewayURL, &CreateUserRequest{
        Email: "test@example.com",
        Name:  "Test User",
    })
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    var user UserResponse
    json.NewDecoder(resp.Body).Decode(&user)
    assert.NotEmpty(t, user.ID)

    // 2. Get user
    resp = getUser(t, gatewayURL, user.ID)
    assert.Equal(t, http.StatusOK, resp.StatusCode)

    // 3. Verify event was published (wait for eventual consistency)
    time.Sleep(2 * time.Second)

    // 4. Check side effect in another service
    // (e.g., order-service created user record)
    orderServiceUser := getOrderServiceUser(t, user.ID)
    assert.Equal(t, user.Email, orderServiceUser.Email)
}
```

**Running E2E Tests:**
```bash
# Start full stack
make up

# Run E2E tests
make test-e2e

# Cleanup
make down
```

### When to Mock vs Use Real Dependencies

**Mock:**
- External services (payment gateways, email services)
- Other microservices (in unit tests)
- Time-dependent logic (use clock interface)

**Use Real (via Testcontainers):**
- Database (MongoDB, PostgreSQL)
- Kafka
- Redis

---

## Observability

### Zerolog (Structured Logging)

**Setup** (shared/logger/logger.go):
```go
package logger

import (
    "context"
    "os"

    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

func New(level, format string) zerolog.Logger {
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

    logLevel, _ := zerolog.ParseLevel(level)
    zerolog.SetGlobalLevel(logLevel)

    if format == "console" {
        return log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
    }

    return zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
}

func WithContext(ctx context.Context) *zerolog.Logger {
    logger := zerolog.Ctx(ctx)
    if logger.GetLevel() == zerolog.Disabled {
        l := log.Logger
        return &l
    }
    return logger
}

func AddTraceID(ctx context.Context, traceID string) context.Context {
    logger := log.With().Str("trace_id", traceID).Logger()
    return logger.WithContext(ctx)
}
```

**Usage**:
```go
log := logger.WithContext(ctx)

log.Info().
    Str("user_id", userID).
    Str("action", "create_user").
    Dur("duration_ms", time.Since(start)).
    Msg("user created successfully")

log.Error().
    Err(err).
    Str("user_id", userID).
    Msg("failed to create user")
```

### Prometheus (Metrics)

**Metrics Middleware** (shared/middleware/http/metrics.go):
```go
package http

import (
    "strconv"
    "time"

    "github.com/labstack/echo/v4"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)

func MetricsMiddleware() echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            start := time.Now()

            err := next(c)

            duration := time.Since(start).Seconds()
            status := c.Response().Status

            httpRequestsTotal.WithLabelValues(
                c.Request().Method,
                c.Path(),
                strconv.Itoa(status),
            ).Inc()

            httpRequestDuration.WithLabelValues(
                c.Request().Method,
                c.Path(),
            ).Observe(duration)

            return err
        }
    }
}
```

**Custom Business Metrics**:
```go
// services/user-service/internal/service/metrics.go
var (
    usersCreated = promauto.NewCounter(prometheus.CounterOpts{
        Name: "users_created_total",
        Help: "Total number of users created",
    })

    activeUsers = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "active_users",
        Help: "Number of currently active users",
    })
)

func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        return nil, err
    }

    usersCreated.Inc()
    activeUsers.Inc()

    return user, nil
}
```

**Metrics Endpoint**:
```go
// Expose /metrics endpoint
import "github.com/prometheus/client_golang/prometheus/promhttp"

e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
```

### Jaeger (Distributed Tracing)

**Tracing Setup** (shared/middleware/grpc/tracing.go):
```go
package grpc

import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    tracesdk "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
    "google.golang.org/grpc"
)

func InitTracer(serviceName, jaegerEndpoint string) (*tracesdk.TracerProvider, error) {
    exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
    if err != nil {
        return nil, err
    }

    tp := tracesdk.NewTracerProvider(
        tracesdk.WithBatcher(exporter),
        tracesdk.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(serviceName),
        )),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}

func TracingInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        tracer := otel.Tracer("grpc-server")
        ctx, span := tracer.Start(ctx, info.FullMethod)
        defer span.End()

        return handler(ctx, req)
    }
}
```

**Request Flow with Tracing**:
```
Request Trace (trace_id: abc123)
│
├─ API Gateway (span: gateway-span-001) [45ms]
│  └─ HTTP POST /api/v1/users
│
├─ User Service (span: user-service-span-002) [40ms]
│  ├─ gRPC CreateUser [35ms]
│  ├─ MongoDB Insert (span: mongodb-span-003) [10ms]
│  └─ Kafka Publish (span: kafka-span-004) [5ms]
│
└─ Order Service (span: order-service-span-005) [15ms]
   └─ Kafka Consume UserCreated
```

---

## Deployment

### Local Development (Docker Compose)

**Full Stack**:
```bash
make up
# Starts: API Gateway, User Service, Web App, MongoDB, Kafka, Jaeger, Prometheus
```

**Dependencies Only**:
```bash
make deps-up
# Starts: MongoDB, Kafka, Jaeger, Prometheus
# Run services locally with `go run cmd/main.go`
```

### Kubernetes

**Deploy to K8s**:
```bash
# Apply base configuration
kubectl apply -k deployments/kubernetes/base/

# Apply production overlay
kubectl apply -k deployments/kubernetes/overlays/prod/

# Check status
kubectl get pods -n production
kubectl get services -n production

# View logs
kubectl logs -f deployment/user-service -n production

# Scale service
kubectl scale deployment/user-service --replicas=10 -n production
```

---

## Example Services Guide

### user-service (Backend Microservice Example)

**Purpose**: Demonstrates complete backend microservice patterns

**Key Files**:
- [services/user-service/cmd/main.go](services/user-service/cmd/main.go) - Entry point
- [services/user-service/internal/domain/user.go](services/user-service/internal/domain/user.go) - Domain entity
- [services/user-service/internal/repository/mongodb/user_repo.go](services/user-service/internal/repository/mongodb/user_repo.go) - MongoDB repository
- [services/user-service/internal/service/user_service.go](services/user-service/internal/service/user_service.go) - Business logic
- [services/user-service/internal/grpc/user_handler.go](services/user-service/internal/grpc/user_handler.go) - gRPC handlers
- [services/user-service/internal/events/producer.go](services/user-service/internal/events/producer.go) - Kafka producer

**Patterns Demonstrated**:
- Clean architecture (domain, repository, service, API)
- Repository pattern with MongoDB
- gRPC service implementation
- Kafka event publishing
- Zerolog structured logging
- Prometheus metrics
- Jaeger tracing

### web-app (HTMX Web Application Example)

**Purpose**: Demonstrates hypermedia-driven UI with HTMX + Alpine.js

**Key Files**:
- [services/web-app/cmd/main.go](services/web-app/cmd/main.go) - Entry point
- [services/web-app/internal/handler/auth.go](services/web-app/internal/handler/auth.go) - Login, OAuth2
- [services/web-app/internal/handler/dashboard.go](services/web-app/internal/handler/dashboard.go) - Dashboard with HTMX
- [services/web-app/templates/layouts/base.html](services/web-app/templates/layouts/base.html) - Base layout
- [services/web-app/templates/pages/login.html](services/web-app/templates/pages/login.html) - Login page
- [services/web-app/templates/pages/dashboard.html](services/web-app/templates/pages/dashboard.html) - Dashboard
- [services/web-app/templates/partials/](services/web-app/templates/partials/) - HTMX partials

**Patterns Demonstrated**:
- Server-side rendering with Go templates
- HTMX for dynamic partial updates
- Alpine.js for client-side interactivity
- Session-based authentication
- JWT + OAuth2 integration
- CSRF protection
- Progressive enhancement

---

## Resources

### Documentation
- [Go Documentation](https://go.dev/doc/)
- [Echo Framework](https://echo.labstack.com/guide/)
- [gRPC Go Guide](https://grpc.io/docs/languages/go/)
- [MongoDB Go Driver](https://www.mongodb.com/docs/drivers/go/current/)
- [Kafka Go Client (Sarama)](https://github.com/IBM/sarama)
- [Zerolog](https://github.com/rs/zerolog)
- [Prometheus Go Client](https://github.com/prometheus/client_golang)
- [Jaeger OpenTelemetry](https://www.jaegertracing.io/)
- [HTMX Documentation](https://htmx.org/docs/)
- [Alpine.js Documentation](https://alpinejs.dev/)
- [Go html/template](https://pkg.go.dev/html/template)

### Tools
- [grpcurl](https://github.com/fullstorydev/grpcurl) - cURL for gRPC
- [kafkacat/kcat](https://github.com/edenhill/kcat) - Kafka debugging
- [k9s](https://k9scli.io/) - Kubernetes TUI
- [mongosh](https://www.mongodb.com/docs/mongodb-shell/) - MongoDB shell

### Best Practices
- [Three Dots Labs - Repository Pattern](https://threedots.tech/post/repository-pattern-in-go/)
- [Go Microservice with Clean Architecture](https://medium.com/@jfeng45/go-micro-service-with-clean-architecture-design-principle-118d2eeef1e6)
- [gRPC Gateway Guide](https://www.koyeb.com/tutorials/build-a-grpc-api-using-go-and-grpc-gateway)
- [HTMX Essays](https://htmx.org/essays/)
