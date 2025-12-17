package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/microserviceboilerplate/web/domain/port/output"
)

// IndexHandler handles index/root requests
type IndexHandler struct {
	sessionRepo output.SessionRepository
}

// NewIndexHandler creates a new index handler
func NewIndexHandler(sessionRepo output.SessionRepository) *IndexHandler {
	return &IndexHandler{
		sessionRepo: sessionRepo,
	}
}

// RegisterRoutes registers index routes
func (h *IndexHandler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.Index)
}

// Index handles root requests
func (h *IndexHandler) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, "request", r)
	
	exists, err := h.sessionRepo.Exists(ctx, "")
	if err == nil && exists {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	
	http.Redirect(w, r, "/auth/login", http.StatusFound)
}

