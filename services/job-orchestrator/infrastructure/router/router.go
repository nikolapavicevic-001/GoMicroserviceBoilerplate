package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httphandler "github.com/microserviceboilerplate/job-orchestrator/adapters/input/http"
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
	jobHandler *httphandler.JobHandler,
	healthHandler *httphandler.HealthHandler,
) {
	r.chi.Route("/", func(r chi.Router) {
		healthHandler.RegisterRoutes(r)
		jobHandler.RegisterRoutes(r)
	})
}

// GetChiRouter returns the underlying chi router
func (r *Router) GetChiRouter() *chi.Mux {
	return r.chi
}

