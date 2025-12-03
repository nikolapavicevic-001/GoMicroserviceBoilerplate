package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	pb "github.com/yourorg/boilerplate/shared/proto/gen/user/v1"
	"github.com/yourorg/boilerplate/services/user-service/config"
	grpcHandler "github.com/yourorg/boilerplate/services/user-service/internal/grpc"
	"github.com/yourorg/boilerplate/services/user-service/internal/repository/mongodb"
	"github.com/yourorg/boilerplate/services/user-service/internal/service"
	mongoClient "github.com/yourorg/boilerplate/shared/database/mongodb"
	"github.com/yourorg/boilerplate/shared/kafka"
	"github.com/yourorg/boilerplate/shared/logger"
	grpcMiddleware "github.com/yourorg/boilerplate/shared/middleware/grpc"
	"github.com/yourorg/boilerplate/shared/tracing"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.NewWithService(cfg.LogLevel, cfg.LogFormat, "user-service")

	log.Info().
		Str("service", "user-service").
		Str("environment", cfg.Environment).
		Int("grpc_port", cfg.GRPCPort).
		Msg("Starting user service")

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize tracing
	tp, err := tracing.InitTracer(ctx, tracing.Config{
		ServiceName:    "user-service",
		JaegerEndpoint: cfg.OTLPEndpoint,
		Environment:    cfg.Environment,
		Enabled:        cfg.TracingEnabled,
	})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize tracing, continuing without it")
	} else if tp != nil {
		defer tracing.Shutdown(ctx, tp)
		log.Info().Str("endpoint", cfg.OTLPEndpoint).Msg("Tracing initialized")
	} else {
		log.Info().Msg("Tracing disabled")
	}

	// Connect to MongoDB
	mongoConfig := mongoClient.NewConfig(cfg.MongoURI, cfg.MongoDB)
	client, err := mongoClient.Connect(ctx, mongoConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer client.Disconnect(ctx)

	log.Info().Msg("Connected to MongoDB")

	// Initialize repository
	db := client.Database(cfg.MongoDB)
	userRepo := mongodb.NewUserRepository(db)

	// Ensure indexes
	if err := userRepo.EnsureIndexes(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to create indexes")
	}

	// Initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(cfg.GetKafkaBrokers(), &log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer kafkaProducer.Close()

	log.Info().Msg("Kafka producer initialized")

	// Initialize service
	userService := service.NewUserService(userRepo, kafkaProducer)

	// Initialize gRPC handler
	userHandler := grpcHandler.NewUserHandler(userService)

	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcMiddleware.TracingInterceptor(),
			grpcMiddleware.LoggingInterceptor(&log),
			grpcMiddleware.RecoveryInterceptor(&log),
		),
	)

	// Register services
	pb.RegisterUserServiceServer(grpcServer, userHandler)

	// Register health check
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("user.v1.UserService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection for grpcurl
	reflection.Register(grpcServer)

	// Start gRPC server
	addr := fmt.Sprintf(":%d", cfg.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen")
	}

	// Start server in goroutine
	go func() {
		log.Info().
			Str("address", addr).
			Msg("gRPC server listening")

		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("Failed to serve gRPC")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down user service...")

	// Graceful shutdown
	grpcServer.GracefulStop()

	log.Info().Msg("User service stopped")
}
