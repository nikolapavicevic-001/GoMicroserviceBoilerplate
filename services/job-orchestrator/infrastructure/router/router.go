package router

import (
	"github.com/go-chi/chi/v5"
	httphandler "github.com/microserviceboilerplate/job-orchestrator/adapters/input/http"
	"github.com/nikolapavicevic-001/CommonGo/httpx"
	"github.com/nikolapavicevic-001/CommonGo/logger"
	"github.com/rs/zerolog"
)

// Router wraps chi router and provides route registration
type Router struct {
	chi *chi.Mux
}

// NewRouter creates a new router instance using CommonGo's httpx
func NewRouter() *Router {
	log := logger.New("info", "job-orchestrator")
	return NewRouterWithLogger(log)
}

// NewRouterWithLogger creates a router with a custom logger
func NewRouterWithLogger(log zerolog.Logger) *Router {
	r := httpx.NewRouter(
		httpx.WithMiddleware(httpx.RequestLogger(log)),
	)

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

