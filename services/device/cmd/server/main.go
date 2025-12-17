package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	grpcadapter "github.com/microserviceboilerplate/device/internal/adapter/grpc"
	httpadapter "github.com/microserviceboilerplate/device/internal/adapter/http"
	natsadapter "github.com/microserviceboilerplate/device/internal/adapter/nats"
	eventadapter "github.com/microserviceboilerplate/device/internal/adapter/event"
	repositoryadapter "github.com/microserviceboilerplate/device/internal/adapter/repository"
	"github.com/microserviceboilerplate/device/internal/service"
	"github.com/nikolapavicevic-001/CommonGo/config"
	"github.com/nikolapavicevic-001/CommonGo/logger"
	natsx "github.com/nikolapavicevic-001/CommonGo/nats"
	"github.com/nikolapavicevic-001/CommonGo/postgres"
)

const (
	grpcPort = ":9090"
	httpPort = ":8080"
)

func main() {
	ctx := context.Background()

	// Initialize logger using CommonGo
	log := logger.New(config.GetEnv("LOG_LEVEL", "info"), "device-service")

	// Database connection using CommonGo
	dbURL := config.GetEnv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable")
	pgCfg := postgres.DefaultConfig(dbURL)
	pool, err := postgres.Open(ctx, pgCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer pool.Close()
	log.Info().Msg("Connected to PostgreSQL")

	// NATS connection using CommonGo
	natsURL := config.GetEnv("NATS_URL", "nats://nats:4222")
	natsCfg := natsx.DefaultConfig(natsURL, "device-service")
	nc, err := natsx.Connect(natsCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to NATS")
	}
	defer nc.Close()

	// JetStream
	js, err := nc.JetStream()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get JetStream")
	}

	// Ensure stream exists for device events
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "DEVICE_EVENTS",
		Subjects: []string{"events.device.*"},
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		log.Warn().Err(err).Msg("Failed to create stream")
	} else {
		log.Info().Msg("NATS JetStream stream ensured")
	}

	// Initialize adapters (hexagonal architecture)
	repo := repositoryadapter.NewPostgresRepository(pool)
	eventPub := eventadapter.NewNATSPublisher(js)
	deviceService := service.NewDeviceService(repo, eventPub)

	// Register NATS request-reply handlers
	natsHandler := natsadapter.NewHandler(nc, deviceService)
	if err := natsHandler.RegisterHandlers(); err != nil {
		log.Warn().Err(err).Msg("Failed to register NATS handlers")
	} else {
		log.Info().Msg("NATS request-reply handlers registered")
	}

	// gRPC server
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen")
	}

	grpcServer := grpc.NewServer()

	// Register health service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register device service
	handler := grpcadapter.NewDeviceHandler(deviceService)
	grpcadapter.RegisterDeviceService(grpcServer, handler)

	// HTTP health check endpoint using chi router
	go func() {
		httpRouter := httpadapter.NewRouter()
		healthHandler := httpadapter.NewHealthHandler()
		httpRouter.RegisterHealthHandler(healthHandler)

		log.Info().Str("port", httpPort).Msg("HTTP health check server listening")
		if err := http.ListenAndServe(httpPort, httpRouter.GetChiRouter()); err != nil {
			log.Error().Err(err).Msg("Health check server error")
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info().Str("port", grpcPort).Msg("gRPC server listening")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("Failed to serve")
		}
	}()

	<-sigChan
	log.Info().Msg("Shutting down gracefully...")

	// Graceful shutdown
	healthServer.Shutdown()
	grpcServer.GracefulStop()
	log.Info().Msg("Server stopped")
}

