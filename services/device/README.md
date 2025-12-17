# Device Service

Go-based gRPC microservice implementing hexagonal architecture with:
- **sqlc** for type-safe PostgreSQL queries
- **NATS JetStream** for event publishing
- **gRPC** API for device CRUD operations

## Architecture

The service follows hexagonal architecture (ports and adapters):

```
cmd/server/           # Application entry point
internal/
  domain/             # Domain layer (entities and interfaces)
  adapter/
    grpc/             # gRPC adapter (primary port)
    repository/       # PostgreSQL adapter (secondary port)
    event/            # NATS adapter (secondary port)
  service/            # Application service layer (business logic)
sql/
  migrations/         # Database migrations
  queries/            # sqlc queries
proto/device/         # Protocol buffer definitions
```

## Building

1. **Generate sqlc code:**
```bash
make sqlc-generate
```

2. **Generate proto code:**
```bash
make proto-generate
```

3. **Build:**
```bash
make build
```

Or use Docker:
```bash
docker build -t device-service .
```

## Running

```bash
make run
```

Or with environment variables:
```bash
DATABASE_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" \
NATS_URL="nats://localhost:4222" \
./bin/device-service
```

## Environment Variables

- `DATABASE_URL`: PostgreSQL connection string
- `NATS_URL`: NATS server URL
- `HTTP_PORT`: HTTP health check port (default: 8080)

## API

The service exposes gRPC endpoints:

- `GetDevice(GetDeviceRequest) returns (Device)` - Get device by ID
- `CreateDevice(CreateDeviceRequest) returns (Device)` - Create new device
- `UpdateDevice(UpdateDeviceRequest) returns (Device)` - Update device
- `DeleteDevice(DeleteDeviceRequest) returns (DeleteDeviceResponse)` - Delete device
- `ListDevices(ListDevicesRequest) returns (ListDevicesResponse)` - List devices with pagination

## Events

The service publishes events to NATS JetStream:

- `events.device.created` - Published when a device is created
- `events.device.updated` - Published when a device is updated
- `events.device.deleted` - Published when a device is deleted

## Database Schema

The `devices` table has the following structure:

- `id` (UUID, primary key)
- `name` (VARCHAR(255))
- `type` (VARCHAR(100))
- `status` (VARCHAR(50), default: 'active')
- `created_at` (TIMESTAMPTZ)
- `updated_at` (TIMESTAMPTZ)

