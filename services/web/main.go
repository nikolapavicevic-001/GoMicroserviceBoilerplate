package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httphandler "github.com/microserviceboilerplate/web/adapters/input/http"
	apihandler "github.com/microserviceboilerplate/web/adapters/input/http/api"
	"github.com/microserviceboilerplate/web/adapters/output/keycloak"
	natsclient "github.com/microserviceboilerplate/web/adapters/output/nats"
	"github.com/microserviceboilerplate/web/adapters/output/session"
	"github.com/microserviceboilerplate/web/adapters/output/template"
	"github.com/microserviceboilerplate/web/domain/service"
	"github.com/microserviceboilerplate/web/infrastructure/config"
	"github.com/microserviceboilerplate/web/infrastructure/router"
	"github.com/nats-io/nats.go"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg := config.Load()

	// NATS connection
	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()
	log.Println("Connected to NATS")

	// Initialize NATS clients
	natsClient := natsclient.NewClient(nc)
	deviceClient := natsclient.NewDeviceClient(natsClient)
	jobClient := natsclient.NewJobClient(natsClient)

	// Initialize repositories (output adapters)
	authRepo := keycloak.NewKeycloakClient()
	sessionRepo := session.NewCookieSessionStore()
	templateRepo, err := template.NewHTMLTemplateRenderer(cfg.TemplatesPath)
	if err != nil {
		log.Fatalf("Failed to initialize template renderer: %v", err)
	}

	// Initialize use cases (domain services)
	authUseCase := service.NewAuthService(authRepo, sessionRepo)
	deviceService := service.NewDeviceService(deviceClient)
	jobService := service.NewJobService(jobClient)

	// Initialize HTTP handlers (input adapters)
	authHandler := httphandler.NewAuthHandler(authUseCase, templateRepo, sessionRepo)
	dashboardHandler := httphandler.NewDashboardHandler(authUseCase, templateRepo, sessionRepo)
	jobsHandler := httphandler.NewJobsHandler(authUseCase, templateRepo, sessionRepo)
	healthHandler := httphandler.NewHealthHandler()
	indexHandler := httphandler.NewIndexHandler(sessionRepo)

	// Initialize API handlers
	deviceAPIHandler := apihandler.NewDeviceAPIHandler(authUseCase, deviceService, templateRepo)
	jobAPIHandler := apihandler.NewJobAPIHandler(authUseCase, jobService)

	// Initialize router
	r := router.NewRouter()
	r.RegisterHandlers(authHandler, dashboardHandler, jobsHandler, healthHandler, indexHandler, deviceAPIHandler, jobAPIHandler)

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

		log.Println("Shutting down server...")
		cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	log.Printf("Starting web service on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
