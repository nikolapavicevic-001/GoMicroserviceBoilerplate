package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/yourorg/boilerplate/services/web-app/internal/handler"
	"github.com/yourorg/boilerplate/services/web-app/templating"
	sharedConfig "github.com/yourorg/boilerplate/shared/config"
	sharedGrpc "github.com/yourorg/boilerplate/shared/grpc"
	"github.com/yourorg/boilerplate/shared/logger"
	httpMiddleware "github.com/yourorg/boilerplate/shared/middleware/http"
	pb "github.com/yourorg/boilerplate/shared/proto/gen/user/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func main() {
	// Load configuration
	cfg := &sharedConfig.BaseConfig{}
	if err := sharedConfig.Load(cfg); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.NewWithService(cfg.LogLevel, cfg.LogFormat, "web-app")
	log.Info().Msg("Starting web-app service")

	// Template FuncMap
	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"formatTime": func(t interface{}) string {
			switch v := t.(type) {
			case *timestamppb.Timestamp:
				if v == nil {
					return "N/A"
				}
				return v.AsTime().Format("Jan 02, 2006 15:04")
			case time.Time:
				return v.Format("Jan 02, 2006 15:04")
			default:
				return fmt.Sprintf("%v", t)
			}
		},
	}

	// Load **all templates** automatically
	loaded, err := templating.LoadTemplates(funcMap)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load templates")
	}

	// Initialize Echo
	e := echo.New()
	e.Renderer = templating.NewRenderer(loaded)
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(httpMiddleware.LoggingMiddleware(&log))

	// Static files
	e.Static("/static", "static")

	// Connect to user service via gRPC
	log.Info().Str("address", cfg.UserServiceAddr).Msg("Connecting to user service")
	userConn, err := sharedGrpc.NewClientConn(cfg.UserServiceAddr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to user service")
	}
	defer userConn.Close()
	userClient := pb.NewUserServiceClient(userConn)

	// Dashboard handler
	dashboardHandler := handler.NewDashboardHandler(cfg.JWTSecret, userClient)

	// Routes (web pages)
	e.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", map[string]interface{}{
			"Title": "Login",
		})
	})
	e.GET("/signup", func(c echo.Context) error {
		return c.Render(http.StatusOK, "signup.html", map[string]interface{}{
			"Title": "Sign Up",
		})
	})
	e.GET("/dashboard", dashboardHandler.ShowDashboard)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": "web-app",
		})
	})

	// Start server in background
	go func() {
		addr := fmt.Sprintf(":%d", cfg.WebAppHTTPPort)
		log.Info().Str("address", addr).Msg("Starting HTTP server")
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Wait for shutdown signal (graceful shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}
