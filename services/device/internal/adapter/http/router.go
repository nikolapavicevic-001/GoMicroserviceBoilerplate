package http

import (
	"github.com/go-chi/chi/v5"
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
	log := logger.New("info", "device-service")
	return NewRouterWithLogger(log)
}

// NewRouterWithLogger creates a router with a custom logger
func NewRouterWithLogger(log zerolog.Logger) *Router {
	r := httpx.NewRouter(
		httpx.WithMiddleware(httpx.RequestLogger(log)),
	)

	return &Router{chi: r}
}

// RegisterHealthHandler registers the health check handler
func (r *Router) RegisterHealthHandler(handler *HealthHandler) {
	handler.RegisterRoutes(r.chi)
}

// GetChiRouter returns the underlying chi router
func (r *Router) GetChiRouter() *chi.Mux {
	return r.chi
}

