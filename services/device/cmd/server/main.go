package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
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
)

const (
	grpcPort = ":9090"
	httpPort = ":8080"
)

func main() {
	ctx := context.Background()

	// Database connection
	dbURL := getEnv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable")
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Verify database connection
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// NATS connection
	natsURL := getEnv("NATS_URL", "nats://nats:4222")
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// JetStream
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to get JetStream: %v", err)
	}

	// Ensure stream exists for device events
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "DEVICE_EVENTS",
		Subjects: []string{"events.device.*"},
	})
	if err != nil && err != nats.ErrStreamNameAlreadyInUse {
		log.Printf("Warning: Failed to create stream: %v", err)
	} else {
		log.Println("NATS JetStream stream ensured")
	}

	// Initialize adapters (hexagonal architecture)
	repo := repositoryadapter.NewPostgresRepository(pool)
	eventPub := eventadapter.NewNATSPublisher(js)
	deviceService := service.NewDeviceService(repo, eventPub)

	// Register NATS request-reply handlers
	natsHandler := natsadapter.NewHandler(nc, deviceService)
	if err := natsHandler.RegisterHandlers(); err != nil {
		log.Printf("Warning: Failed to register NATS handlers: %v", err)
	} else {
		log.Println("NATS request-reply handlers registered")
	}

	// gRPC server
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
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

		log.Printf("HTTP health check server listening on %s", httpPort)
		if err := http.ListenAndServe(httpPort, httpRouter.GetChiRouter()); err != nil {
			log.Printf("Health check server error: %v", err)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("gRPC server listening on %s", grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	<-sigChan
	log.Println("Shutting down gracefully...")

	// Graceful shutdown
	healthServer.Shutdown()
	grpcServer.GracefulStop()
	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

