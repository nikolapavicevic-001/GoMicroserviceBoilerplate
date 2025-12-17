package main

import (
	"context"
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
	"github.com/nikolapavicevic-001/CommonGo/logger"
	natsx "github.com/nikolapavicevic-001/CommonGo/nats"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger
	log := logger.New("info", "web-service")

	// Load configuration
	cfg := config.Load()

	// NATS connection using CommonGo
	natsCfg := natsx.DefaultConfig(cfg.NATSURL, "web-service")
	nc, err := natsx.Connect(natsCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to NATS")
	}
	defer nc.Close()
	log.Info().Msg("Connected to NATS")

	// Initialize NATS clients
	natsClient := natsclient.NewClient(nc)
	deviceClient := natsclient.NewDeviceClient(natsClient)
	jobClient := natsclient.NewJobClient(natsClient)

	// Initialize repositories (output adapters)
	authRepo := keycloak.NewKeycloakClient()
	sessionRepo := session.NewCookieSessionStore()
	templateRepo, err := template.NewHTMLTemplateRenderer(cfg.TemplatesPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize template renderer")
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

		log.Info().Msg("Shutting down server...")
		cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Server shutdown error")
		}
	}()

	log.Info().Str("port", cfg.Port).Msg("Starting web service")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
