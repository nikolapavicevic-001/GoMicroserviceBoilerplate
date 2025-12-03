package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/yourorg/boilerplate/services/api-gateway/config"
	_ "github.com/yourorg/boilerplate/services/api-gateway/docs" // Swagger docs
	"github.com/yourorg/boilerplate/services/api-gateway/internal/grpc"
	"github.com/yourorg/boilerplate/services/api-gateway/internal/handler"
	mw "github.com/yourorg/boilerplate/services/api-gateway/internal/middleware"
	"github.com/yourorg/boilerplate/shared/auth"
	"github.com/yourorg/boilerplate/shared/logger"
)

// @title Go Microservices API
// @version 1.0
// @description API Gateway for Go Microservices Boilerplate
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel, cfg.LogFormat)

	log.Info().
		Str("service", "api-gateway").
		Str("environment", cfg.Environment).
		Int("http_port", cfg.HTTPPort).
		Msg("Starting API Gateway")

	// Create gRPC clients
	userClient, userConn, err := grpc.NewUserServiceClient(cfg.UserServiceAddr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create user service client")
	}
	defer userConn.Close()

	log.Info().Str("address", cfg.UserServiceAddr).Msg("Connected to user service")

	// Initialize OAuth2 providers
	var googleProvider, githubProvider *auth.OAuth2Provider

	if cfg.OAuth2GoogleClientID != "" {
		googleProvider = auth.NewGoogleProvider(
			cfg.OAuth2GoogleClientID,
			cfg.OAuth2GoogleClientSecret,
			cfg.OAuth2GoogleRedirectURL,
		)
		log.Info().Msg("Google OAuth2 provider initialized")
	}

	if cfg.OAuth2GitHubClientID != "" {
		githubProvider = auth.NewGitHubProvider(
			cfg.OAuth2GitHubClientID,
			cfg.OAuth2GitHubClientSecret,
			cfg.OAuth2GitHubRedirectURL,
		)
		log.Info().Msg("GitHub OAuth2 provider initialized")
	}

	// Create Echo server
	e := echo.New()
	e.HideBanner = true

	// Global middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
	}))
	e.Use(mw.LoggerMiddleware(&log))

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	// Initialize handlers
	userHandler := handler.NewUserHandler(userClient, cfg.JWTSecret, cfg.JWTExpiry)
	authHandler := handler.NewAuthHandler(
		userClient,
		cfg.JWTSecret,
		cfg.JWTExpiry,
		googleProvider,
		githubProvider,
	)

	// Public routes
	api := e.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.POST("/login", userHandler.Login)

	if googleProvider != nil {
		auth.GET("/google", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
	}

	if githubProvider != nil {
		auth.GET("/github", authHandler.GitHubLogin)
		auth.GET("/github/callback", authHandler.GitHubCallback)
	}

	// User routes (some public, some protected)
	users := api.Group("/users")
	users.POST("", userHandler.CreateUser) // Public - user registration

	// Protected routes
	protected := api.Group("")
	protected.Use(mw.JWTMiddleware(cfg.JWTSecret))

	protected.GET("/users", userHandler.ListUsers)
	protected.GET("/users/:id", userHandler.GetUser)
	protected.PUT("/users/:id", userHandler.UpdateUser)
	protected.DELETE("/users/:id", userHandler.DeleteUser)

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Start server
	go func() {
		addr := fmt.Sprintf(":%d", cfg.HTTPPort)
		log.Info().Str("address", addr).Msg("HTTP server listening")

		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down API Gateway...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to shutdown server")
	}

	log.Info().Msg("API Gateway stopped")
}
