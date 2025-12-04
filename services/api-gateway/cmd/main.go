package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/yourorg/boilerplate/services/api-gateway/docs" // Swagger docs
	"github.com/yourorg/boilerplate/services/api-gateway/internal/grpc"
	"github.com/yourorg/boilerplate/services/api-gateway/internal/handler"
	gatewayMiddleware "github.com/yourorg/boilerplate/services/api-gateway/internal/middleware"
	sharedConfig "github.com/yourorg/boilerplate/shared/config"
	"github.com/yourorg/boilerplate/shared/logger"
	httpMiddleware "github.com/yourorg/boilerplate/shared/middleware/http"
)

// @title Go Microservices API
// @version 1.0
// @description API Gateway
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load configuration
	cfg := &sharedConfig.BaseConfig{}
	if err := sharedConfig.Load(cfg); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.NewWithService(cfg.LogLevel, cfg.LogFormat, "api-gateway")

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

	// Create Echo server
	e := echo.New()
	e.HideBanner = true

	// Global middleware
	e.Use(middleware.Recover())

	// Parse CORS allowed origins
	origins := []string{"http://localhost:3000", "http://localhost:8080"}
	if cfg.CORSAllowedOrigins != "" {
		// Split by comma and trim spaces
		parts := strings.Split(cfg.CORSAllowedOrigins, ",")
		origins = make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: origins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{echo.HeaderContentType, echo.HeaderAuthorization},
	}))
	e.Use(httpMiddleware.LoggingMiddleware(&log))

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	// Setup reverse proxy to web-app for non-API routes
	webAppURL, err := url.Parse(cfg.WebAppAddr)
	if err != nil {
		log.Fatal().Err(err).Str("url", cfg.WebAppAddr).Msg("Failed to parse web-app URL")
	}

	// Initialize handler with web-app URL for proxying
	h := handler.NewHandler(
		userClient,
		cfg.JWTSecret,
		cfg.GetJWTExpiry(),
		webAppURL,
	)

	// Public routes
	api := e.Group("/api/v1")

	// Auth routes (public - JSON API)
	authGroup := api.Group("/auth")
	authGroup.POST("/register", h.Register)
	authGroup.POST("/login", h.Login)

	// Protected routes (JSON API - Bearer token)
	protected := api.Group("")
	protected.Use(httpMiddleware.JWTMiddleware(cfg.JWTSecret))

	protected.GET("/users", h.ListUsers)
	protected.GET("/users/:id", h.GetUser)
	protected.PUT("/users/:id", h.UpdateUser)
	protected.DELETE("/users/:id", h.DeleteUser)

	// Swagger documentation
	e.GET("/swagger", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Public auth routes (HTML forms)
	e.GET("/login", h.ShowLoginPage)
	e.POST("/login", h.Login)
	e.GET("/signup", h.ShowSignupPage)
	e.POST("/signup", h.Register)
	e.GET("/logout", h.Logout)

	// Setup reverse proxy for web-app static content and dashboard
	proxy := httputil.NewSingleHostReverseProxy(webAppURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error().Err(err).Str("path", r.URL.Path).Msg("Proxy error")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Web application unavailable"))
	}

	// Proxy handler for web-app routes (protected by JWT cookie)
	proxyHandler := func(c echo.Context) error {
		proxy.ServeHTTP(c.Response(), c.Request())
		return nil
	}

	// Protected web-app routes (require JWT cookie)
	protectedWebApp := e.Group("")
	protectedWebApp.Use(gatewayMiddleware.JWTCookieMiddleware(cfg.JWTSecret))
	protectedWebApp.GET("/dashboard", proxyHandler)
	protectedWebApp.GET("/dashboard/*", proxyHandler)
	protectedWebApp.GET("/static/*", proxyHandler)

	// Root redirect to login
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/login")
	})

	log.Info().Str("webapp_url", cfg.WebAppAddr).Msg("Reverse proxy configured for web-app")

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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("Failed to shutdown server")
	}

	log.Info().Msg("API Gateway stopped")
}
