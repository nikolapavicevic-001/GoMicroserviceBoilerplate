package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

// RegisterHealthHandler registers the health check handler
func (r *Router) RegisterHealthHandler(handler *HealthHandler) {
	handler.RegisterRoutes(r.chi)
}

// GetChiRouter returns the underlying chi router
func (r *Router) GetChiRouter() *chi.Mux {
	return r.chi
}

