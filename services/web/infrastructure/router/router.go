package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httphandler "github.com/microserviceboilerplate/web/adapters/input/http"
	apihandler "github.com/microserviceboilerplate/web/adapters/input/http/api"
)

// Router wraps chi router and provides route registration
type Router struct {
	chi *chi.Mux
}

// NewRouter creates a new router instance
func NewRouter() *Router {
	r := chi.NewRouter()
	
	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	
	return &Router{chi: r}
}

// RegisterHandlers registers all HTTP handlers
func (r *Router) RegisterHandlers(
	authHandler *httphandler.AuthHandler,
	dashboardHandler *httphandler.DashboardHandler,
	jobsHandler *httphandler.JobsHandler,
	healthHandler *httphandler.HealthHandler,
	indexHandler *httphandler.IndexHandler,
	deviceAPIHandler *apihandler.DeviceAPIHandler,
	jobAPIHandler *apihandler.JobAPIHandler,
) {
	r.chi.Route("/", func(r chi.Router) {
		indexHandler.RegisterRoutes(r)
		healthHandler.RegisterRoutes(r)
		authHandler.RegisterRoutes(r)
		dashboardHandler.RegisterRoutes(r)
		jobsHandler.RegisterRoutes(r)
		deviceAPIHandler.RegisterRoutes(r)
		jobAPIHandler.RegisterRoutes(r)
	})
}

// GetChiRouter returns the underlying chi router
func (r *Router) GetChiRouter() *chi.Mux {
	return r.chi
}
