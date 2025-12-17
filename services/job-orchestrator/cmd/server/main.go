package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	httphandler "github.com/microserviceboilerplate/job-orchestrator/adapters/input/http"
	"github.com/microserviceboilerplate/job-orchestrator/adapters/output/flink"
	"github.com/microserviceboilerplate/job-orchestrator/adapters/output/repository"
	"github.com/microserviceboilerplate/job-orchestrator/domain/service"
	"github.com/microserviceboilerplate/job-orchestrator/infrastructure/config"
	"github.com/microserviceboilerplate/job-orchestrator/infrastructure/router"
	"github.com/microserviceboilerplate/job-orchestrator/infrastructure/scheduler"
	natsadapter "github.com/microserviceboilerplate/job-orchestrator/internal/adapter/nats"
	commoncfg "github.com/nikolapavicevic-001/CommonGo/config"
	"github.com/nikolapavicevic-001/CommonGo/logger"
	natsx "github.com/nikolapavicevic-001/CommonGo/nats"
	"github.com/nikolapavicevic-001/CommonGo/postgres"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger using CommonGo
	log := logger.New(commoncfg.GetEnv("LOG_LEVEL", "info"), "job-orchestrator")

	// Load configuration
	cfg := config.Load()

	// Database connection using CommonGo
	pgCfg := postgres.DefaultConfig(cfg.DatabaseURL)
	pool, err := postgres.Open(ctx, pgCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer pool.Close()
	log.Info().Msg("Connected to PostgreSQL")

	// Run database migrations
	if err := runMigrations(ctx, pool); err != nil {
		log.Fatal().Err(err).Msg("Failed to run migrations")
	}
	log.Info().Msg("Database migrations completed")

	// NATS connection using CommonGo
	natsURL := commoncfg.GetEnv("NATS_URL", "nats://nats:4222")
	natsCfg := natsx.DefaultConfig(natsURL, "job-orchestrator")
	nc, err := natsx.Connect(natsCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to NATS")
	}
	defer nc.Close()
	log.Info().Msg("Connected to NATS")

	// Initialize repositories
	jobRepo := repository.NewPostgresJobRepository(pool)
	flinkClient := flink.NewFlinkRESTClient(cfg.FlinkJobManagerURL)

	// Initialize service
	jobService := service.NewJobService(jobRepo, flinkClient, cfg.DataProcessorURL)

	// Register NATS request-reply handlers
	natsHandler := natsadapter.NewHandler(nc, jobService)
	if err := natsHandler.RegisterHandlers(); err != nil {
		log.Warn().Err(err).Msg("Failed to register NATS handlers")
	} else {
		log.Info().Msg("NATS request-reply handlers registered")
	}

	// Initialize HTTP handlers
	jobHandler := httphandler.NewJobHandler(jobService)
	healthHandler := httphandler.NewHealthHandler()

	// Initialize router
	r := router.NewRouter()
	r.RegisterHandlers(jobHandler, healthHandler)

	// Initialize and start cron scheduler
	cronScheduler := scheduler.NewCronScheduler(jobService, jobRepo)
	if err := cronScheduler.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to start scheduler")
	}
	defer cronScheduler.Stop()

	// Start HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r.GetChiRouter(),
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Info().Msg("Shutting down server...")
		cancel()
		cronScheduler.Stop()
		if err := server.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Server shutdown error")
		}
	}()

	log.Info().Str("port", cfg.Port).Msg("Starting job orchestrator")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}

// runMigrations runs database migrations
func runMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Migration SQL - inline for simplicity
	migrationSQL := `
-- Create jobs table
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    jar_url VARCHAR(500) NOT NULL,
    entry_class VARCHAR(255) DEFAULT 'com.microserviceboilerplate.DataProcessor',
    config JSONB DEFAULT '{}',
    schedule_cron VARCHAR(100),
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create job_executions table
CREATE TABLE IF NOT EXISTS job_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID REFERENCES jobs(id) ON DELETE CASCADE,
    flink_job_id VARCHAR(255),
    flink_jar_id VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_job_executions_job_id ON job_executions(job_id);
CREATE INDEX IF NOT EXISTS idx_job_executions_status ON job_executions(status);
CREATE INDEX IF NOT EXISTS idx_job_executions_flink_job_id ON job_executions(flink_job_id);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_jobs_updated_at ON jobs;
CREATE TRIGGER update_jobs_updated_at BEFORE UPDATE ON jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
`

	_, err := pool.Exec(ctx, migrationSQL)
	return err
}
