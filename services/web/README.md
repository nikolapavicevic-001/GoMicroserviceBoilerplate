# Web Service - Hexagonal Architecture

This web service is built using **Hexagonal Architecture** (also known as Ports and Adapters) with the **Chi** router.

## Architecture Overview

The application is organized into clear layers:

```
services/web/
├── domain/                    # Core business logic (no dependencies on external frameworks)
│   ├── entity/               # Domain entities
│   ├── port/
│   │   ├── input/           # Input ports (use case interfaces)
│   │   └── output/           # Output ports (repository interfaces)
│   └── service/              # Use case implementations
├── adapters/                 # Adapters that connect to external world
│   ├── input/http/          # HTTP handlers (input adapters)
│   └── output/               # External service adapters
│       ├── keycloak/        # Keycloak authentication client
│       ├── session/         # Session storage adapter
│       └── template/        # Template rendering adapter
├── infrastructure/           # Infrastructure concerns
│   ├── config/              # Configuration management
│   └── router/              # Router setup and middleware
└── main.go                   # Dependency injection and wiring

```

## Key Principles

### 1. **Domain Layer** (Business Logic)
- **Entities**: Core business objects (`User`)
- **Ports**: Interfaces that define contracts
  - **Input Ports**: Use case interfaces (what the application can do)
  - **Output Ports**: Repository interfaces (what the application needs from outside)
- **Services**: Implementation of use cases (business logic)

### 2. **Adapters Layer** (External Interfaces)
- **Input Adapters**: HTTP handlers that receive requests and translate them to use cases
- **Output Adapters**: Implementations of repositories (Keycloak client, session store, template renderer)

### 3. **Infrastructure Layer**
- Configuration management
- Router setup (Chi)
- Middleware configuration

## Benefits

1. **Testability**: Easy to mock interfaces for testing
2. **Flexibility**: Can swap implementations (e.g., change from Keycloak to another auth provider)
3. **Independence**: Domain logic is independent of frameworks
4. **Maintainability**: Clear separation of concerns
5. **Scalability**: Easy to add new features without affecting existing code

## Usage

### Running the Service

```bash
# Build
go build -o web-service .

# Run
./web-service

# Or with Docker
make rebuild-web
```

### Adding New Features

1. **Add a new use case**:
   - Define interface in `domain/port/input/`
   - Implement in `domain/service/`

2. **Add a new HTTP endpoint**:
   - Create handler in `adapters/input/http/`
   - Register routes in `infrastructure/router/router.go`

3. **Add a new external service**:
   - Define interface in `domain/port/output/`
   - Implement adapter in `adapters/output/`

## Dependencies

- **Chi Router**: Lightweight HTTP router
- **Gorilla Sessions**: Session management
- **OAuth2**: Authentication with Keycloak

## Configuration

Environment variables:
- `PORT`: Server port (default: 8080)
- `KEYCLOAK_URL`: Keycloak server URL
- `KEYCLOAK_REALM`: Keycloak realm name
- `KEYCLOAK_CLIENT_ID`: OAuth client ID
- `KEYCLOAK_CLIENT_SECRET`: OAuth client secret
- `TEMPLATES_PATH`: Path to HTML templates (default: `templates/*.html`)

