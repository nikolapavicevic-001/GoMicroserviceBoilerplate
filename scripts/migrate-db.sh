#!/bin/bash

set -e

DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-postgres}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

echo "Running database migrations..."

# Prefer running migrations inside the postgres container when psql isn't available locally.
if command -v psql >/dev/null 2>&1; then
	# Check if running in Docker or locally
	if [ -f /.dockerenv ] || [ -n "$KUBERNETES_SERVICE_HOST" ]; then
		PGHOST="$DB_HOST"
	else
		PGHOST="localhost"
	fi

	export PGPASSWORD="$DB_PASSWORD"
	psql -h "$PGHOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f infrastructure/init-scripts/01-init.sql
else
	if ! command -v docker >/dev/null 2>&1; then
		echo "Error: psql not found and docker not available. Install psql or run migrations from inside the postgres container."
		exit 1
	fi

	echo "psql not found locally; running migrations inside Docker container 'postgres'..."
	# Feed the SQL file via stdin so we don't depend on container filesystem paths.
	docker exec -i postgres psql -U "$DB_USER" -d "$DB_NAME" < infrastructure/init-scripts/01-init.sql
fi

echo "Database migrations completed!"

