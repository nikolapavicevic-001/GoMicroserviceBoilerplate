# Job Orchestrator Service

A Go microservice that orchestrates Flink data processing jobs with scheduling capabilities. The orchestrator handles all job submission to Flink via REST API.

## Architecture

This service follows hexagonal architecture (ports and adapters):

```
services/job-orchestrator/
├── cmd/server/              # Application entry point
├── domain/                  # Domain layer (business logic)
│   ├── entity/             # Domain entities (Job, JobExecution)
│   ├── port/
│   │   ├── input/          # Input ports (use cases)
│   │   └── output/         # Output ports (repositories, clients)
│   └── service/            # Domain services
├── adapters/
│   ├── input/http/         # HTTP handlers (Chi router)
│   └── output/
│       ├── flink/          # Flink REST API client
│       └── repository/     # PostgreSQL repository
├── infrastructure/
│   ├── config/             # Configuration management
│   ├── router/             # Router setup
│   └── scheduler/          # Cron scheduler
└── sql/migrations/         # Database migrations
```

## Features

- **Job Management**: Create, read, update, delete job definitions
- **Job Execution**: Start jobs immediately, stop running jobs
- **Scheduling**: Cron-based scheduling for periodic jobs
- **JAR Management**: Downloads JAR from data-processor service and uploads to Flink
- **Execution Tracking**: Tracks job execution history and status
- **REST API**: Full REST API for job management

## API Endpoints

### Job Management

- `POST /api/v1/jobs` - Create job definition
- `GET /api/v1/jobs` - List all jobs
- `GET /api/v1/jobs/:id` - Get job details
- `PUT /api/v1/jobs/:id` - Update job
- `DELETE /api/v1/jobs/:id` - Delete job
- `POST /api/v1/jobs/:id/start` - Start job immediately
- `POST /api/v1/jobs/:id/stop` - Stop running job

### Scheduling

- `POST /api/v1/jobs/:id/schedule` - Set schedule (cron expression)
- `DELETE /api/v1/jobs/:id/schedule` - Remove schedule
- `GET /api/v1/jobs/:id/executions` - Get job execution history

### Health

- `GET /health` - Health check

## Configuration

Environment variables:

- `PORT` - HTTP server port (default: `8080`)
- `FLINK_JOBMANAGER_URL` - Flink JobManager REST API URL (default: `http://flink-jobmanager:8081`)
- `DATABASE_URL` - PostgreSQL connection string
- `DATA_PROCESSOR_URL` - Data processor service URL for JAR download (default: `http://data-processor:8080`)

## Database Schema

The service uses PostgreSQL to store job definitions and execution history:

- `jobs` - Job definitions with configuration and schedules
- `job_executions` - Execution history with status tracking

See `sql/migrations/001_create_jobs.sql` for the full schema.

## Job Submission Flow

1. Orchestrator downloads JAR from data-processor service (HTTP GET)
2. Orchestrator uploads JAR to Flink via REST API
3. Flink returns JAR ID
4. Orchestrator submits job using JAR ID and configuration
5. Orchestrator tracks job execution status
6. Orchestrator stores execution history in database

## Development

### Build

```bash
cd services/job-orchestrator
go build ./cmd/server
```

### Run

```bash
./job-orchestrator
```

### Run with Docker

```bash
make rebuild-orchestrator
```

### Test

```bash
go test ./...
```

## Example: Create and Start a Job

```bash
# Create a job
curl -X POST http://localhost:8085/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "data-processor-job",
    "jarURL": "http://data-processor:8080/data-processor.jar",
    "entryClass": "com.microserviceboilerplate.DataProcessor",
    "config": {
      "source": "socket",
      "sink": "postgres"
    }
  }'

# Start the job
curl -X POST http://localhost:8085/api/v1/jobs/{job-id}/start

# Set a schedule (runs every minute)
curl -X POST http://localhost:8085/api/v1/jobs/{job-id}/schedule \
  -H "Content-Type: application/json" \
  -d '{"cron": "0 * * * * *"}'
```

