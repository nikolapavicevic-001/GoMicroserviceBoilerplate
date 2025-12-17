# Development Guide

This guide explains how to work with the microservice boilerplate during development.

## Development Workflows

### 1. Hot Reload Mode (Recommended for Active Development)

Use hot reload mode when actively coding. Changes to Go source files will automatically trigger rebuilds and restarts.

#### Starting Hot Reload Mode

```bash
make dev
```

This command:
- Builds development Docker images (with Air installed)
- Mounts source code as volumes
- Starts all services with hot reload enabled
- Services automatically rebuild when you save files

#### Managing Hot Reload Environment

```bash
make dev-up      # Start services (if already built)
make dev-down    # Stop services
make dev-logs    # View logs from all services
make dev-ps      # Show container status
```

#### Viewing Service Logs

```bash
# All services
make dev-logs

# Individual services (still works)
make logs-web
make logs-device
make logs-orchestrator
```

#### How It Works

- **Air** watches for file changes in Go services
- When a `.go` file is modified, Air automatically:
  1. Stops the running process
  2. Rebuilds the binary
  3. Restarts the service
- No Docker rebuild needed for code changes!

#### Supported Services

Hot reload is enabled for:
- `web-service` - Web application
- `device-service` - Device management service
- `job-orchestrator` - Job orchestration service

Other services (PostgreSQL, NATS, Keycloak, etc.) run normally without hot reload.

### 2. Standard Development Mode

Use standard mode when you need to rebuild Docker images or test the production build process.

```bash
make prod-dev    # Build images and start (same as old 'make dev')
make up          # Start containers
make down        # Stop containers
```

In this mode:
- Code changes require manual rebuilds
- Docker images are built from scratch
- Better for testing production-like builds

### 3. Rebuilding Services

#### Grouped Rebuilds (Quick Commands)

Instead of rebuilding services one by one, use grouped commands:

```bash
make rebuild-go        # Rebuild all Go services (device, web, orchestrator)
make rebuild-backend   # Rebuild backend services (device, orchestrator, data-processor)
make rebuild-services  # Rebuild all custom application services
make rebuild-deps      # Rebuild infrastructure (postgres, nats, keycloak, etc.)
make rebuild-all       # Rebuild everything
```

#### Individual Service Rebuilds

```bash
make rebuild-web           # Rebuild web service
make rebuild-device        # Rebuild device service
make rebuild-orchestrator  # Rebuild job orchestrator
make rebuild-data-processor # Rebuild data processor
```

#### Restart Without Rebuild

```bash
make restart-web
make restart-device
make restart-orchestrator
```

## When to Use Which Workflow

### Use Hot Reload (`make dev`) when:
- ✅ Actively writing code
- ✅ Making frequent changes
- ✅ Testing new features
- ✅ Debugging
- ✅ Want fastest iteration cycle

### Use Standard Mode (`make prod-dev` or `make up`) when:
- ✅ Testing production builds
- ✅ Need to rebuild Docker images
- ✅ Testing deployment process
- ✅ Code changes are infrequent
- ✅ Performance testing

## Quick Reference

### Starting Development

```bash
# Hot reload (recommended)
make dev

# Standard mode
make prod-dev
```

### Rebuilding Services

```bash
# All Go services at once
make rebuild-go

# All backend services
make rebuild-backend

# Everything
make rebuild-all
```

### Managing Services

```bash
# Hot reload mode
make dev-up / make dev-down / make dev-logs

# Standard mode
make up / make down / make logs
```

### Viewing Logs

```bash
# All services
make dev-logs      # Hot reload mode
make logs          # Standard mode

# Individual services
make logs-web
make logs-device
make logs-orchestrator
```

## Troubleshooting

### Hot Reload Not Working

1. **Check if Air is running:**
   ```bash
   make dev-ps
   docker exec -it web-service ps aux | grep air
   ```

2. **Check Air logs:**
   ```bash
   make dev-logs | grep -i air
   ```

3. **Rebuild dev images:**
   ```bash
   make dev-down
   cd infrastructure
   docker-compose -f docker-compose.yml -f docker-compose.dev.yml build --no-cache web-service
   make dev-up
   ```

### Build Errors in Hot Reload

- Check the `build-errors.log` file in the service directory
- Verify Go syntax is correct
- Ensure dependencies are installed (`make install-deps`)

### Port Conflicts

If ports are already in use:
```bash
# Stop all containers
make dev-down
# or
make down

# Check what's using the port
lsof -i :8080
```

## Configuration

### Air Configuration

Air configuration files:
- `services/web/.air.toml`
- `services/device/.air.toml`
- `services/job-orchestrator/.air.toml`

You can customize:
- Build commands
- Watched file extensions
- Excluded directories
- Build delays

### Docker Compose Overrides

- `infrastructure/docker-compose.yml` - Base configuration
- `infrastructure/docker-compose.dev.yml` - Development overrides

The dev override:
- Mounts source code as volumes
- Sets Air as the command
- Configures development environment variables

## Tips

1. **First time setup:** Run `make install-deps` to install required tools
2. **Database migrations:** Run `make migrate` after starting services
3. **Clean start:** Use `make clean-all` to remove everything and start fresh
4. **Multiple terminals:** Use one terminal for logs (`make dev-logs`) and another for commands
5. **Service dependencies:** Some services depend on others - wait for dependencies to start before testing

## Additional Resources

- [Air Documentation](https://github.com/cosmtrek/air)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Makefile Help](../Makefile) - Run `make help` for all available commands

