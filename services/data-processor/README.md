# Data Processor - Hexagonal Architecture

This Flink job processes streaming events using a **Hexagonal Architecture** (Ports and Adapters pattern), making it easy to add new event sources, processors, and sinks.

## Architecture Overview

The application is organized into clear layers:

```
services/data-processor/
├── domain/                    # Core business logic
│   ├── model/                # Domain entities (Event)
│   ├── port/
│   │   ├── input/           # Input ports (EventSource)
│   │   └── output/          # Output ports (EventSink, EventProcessor)
│   └── service/              # Use case orchestration (EventProcessingService)
├── adapters/                 # Adapters that connect to external world
│   ├── input/               # Source adapters (SocketSource, KafkaSource, NatsSource)
│   ├── output/              # Sink adapters (PostgresSink, S3Sink, ClickHouseSink)
│   └── processor/           # Processing adapters (JsonEventParser)
├── infrastructure/          # Infrastructure concerns
│   ├── config/              # Configuration management (FlinkConfig)
│   └── job/                 # Job orchestration (DataProcessorJob)
└── DataProcessor.java       # Entry point
```

## Key Components

### Domain Layer

- **Event Model**: Core event entity with validation
- **EventSource Port**: Interface for reading events (from NATS, Kafka, Socket, etc.)
- **EventSink Port**: Interface for writing events (to PostgreSQL, S3, ClickHouse, etc.)
- **EventProcessor Port**: Interface for transforming events (JSON parsing, routing, etc.)
- **EventProcessingService**: Orchestrates the pipeline

### Adapters Layer

**Source Adapters:**
- `SocketSource`: Reads from socket (for testing)
- `KafkaSourceAdapter`: Reads from Kafka (placeholder - needs implementation)
- `NatsSource`: Reads from NATS (placeholder - needs implementation)

**Sink Adapters:**
- `PostgresSink`: Writes to PostgreSQL with batch processing
- `S3Sink`: Writes to S3/MinIO with rolling file policy
- `ClickHouseSink`: Placeholder for future ClickHouse support

**Processor Adapters:**
- `JsonEventParser`: Parses JSON strings into Event objects using Jackson

### Infrastructure Layer

- **FlinkConfig**: Loads configuration from environment variables
- **DataProcessorJob**: Sets up Flink environment and wires components together

## Building

```bash
mvn clean package
```

The JAR will be created in `target/data-processor-1.0-SNAPSHOT.jar`.

## Configuration

Set environment variables:

### Source Configuration
- `SOURCE_TYPE`: Source type - `socket`, `kafka`, or `nats` (default: `socket`)
- `KAFKA_BOOTSTRAP_SERVERS`: Kafka bootstrap servers (default: `localhost:9092`)
- `KAFKA_TOPIC`: Kafka topic name (default: `events`)
- `NATS_URL`: NATS server URL (default: `nats://nats:4222`)
- `NATS_SUBJECT`: NATS subject name (default: `events`)

### PostgreSQL Configuration
- `POSTGRES_URL`: PostgreSQL connection URL (default: `jdbc:postgresql://postgres:5432/postgres`)
- `POSTGRES_USER`: PostgreSQL username (default: `postgres`)
- `POSTGRES_PASSWORD`: PostgreSQL password (default: `postgres`)

### S3/MinIO Configuration
- `S3_ENDPOINT`: S3 endpoint URL (default: `http://minio:9000`)
- `S3_ACCESS_KEY`: S3 access key (default: `minioadmin`)
- `S3_SECRET_KEY`: S3 secret key (default: `minioadmin`)
- `S3_BUCKET`: S3 bucket name (default: `raw`)

### Flink Configuration
- `CHECKPOINT_INTERVAL_MS`: Checkpoint interval in milliseconds (default: `60000`)

## Running

### Local Development (Socket Source)

1. Start a socket server to send test data:
```bash
nc -l 9999
```

2. Run the Flink job:
```bash
java -jar target/data-processor-1.0-SNAPSHOT.jar
```

3. Send JSON events to the socket:
```json
{"user_id": "123", "event_type": "user.created", "timestamp": 1699123456}
```

### Production (Kafka/NATS)

Set `SOURCE_TYPE=kafka` or `SOURCE_TYPE=nats` and configure the respective connection details.

## Event Format

Events should be JSON with the following fields:

```json
{
  "id": "optional-event-id",
  "user_id": "user-123",
  "event_type": "user.created",
  "payload": "{\"email\":\"user@example.com\"}",
  "timestamp": 1699123456
}
```

- `id`: Optional event ID (auto-generated if not provided)
- `user_id`: Required user identifier
- `event_type`: Required event type
- `payload`: Required event payload (can be JSON string or any string)
- `timestamp`: Optional Unix timestamp (seconds or milliseconds) or ISO-8601 string

## Extending the Architecture

### Adding a New Source

1. Implement `EventSource` interface:
```java
public class MyCustomSource implements EventSource {
    @Override
    public DataStream<String> createSource(StreamExecutionEnvironment env) {
        // Your implementation
    }
}
```

2. Add it to `DataProcessorJob.createSource()` method.

### Adding a New Sink

1. Implement `EventSink` interface:
```java
public class MyCustomSink implements EventSink {
    @Override
    public void addSink(DataStream<Event> stream, StreamExecutionEnvironment env) {
        // Your implementation
    }
}
```

2. Add it to `DataProcessorJob.createSinks()` method.

### Adding a New Processor

1. Implement `EventProcessor` interface:
```java
public class MyCustomProcessor implements EventProcessor {
    @Override
    public DataStream<Event> process(DataStream<String> rawStream) {
        // Your implementation
    }
}
```

2. Use it in `DataProcessorJob.createProcessor()` method.

## Benefits

1. **Testability**: Each component can be unit tested independently
2. **Extensibility**: Easy to add new sources, sinks, or processors
3. **Maintainability**: Clear separation of concerns
4. **Flexibility**: Swap implementations without changing core logic
5. **Type Safety**: Strong typing with domain models

## Database Schema

The PostgreSQL sink expects the following table:

```sql
CREATE TABLE events (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL
);
```

## Future Enhancements

- [ ] Implement Kafka source connector
- [ ] Implement NATS source connector
- [ ] Implement ClickHouse sink
- [ ] Add dead letter queue for failed events
- [ ] Add metrics and monitoring
- [ ] Add event routing based on event type
- [ ] Add unit tests
- [ ] Add integration tests
