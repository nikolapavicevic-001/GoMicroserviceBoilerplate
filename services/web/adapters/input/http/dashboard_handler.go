package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/microserviceboilerplate/web/domain/port/input"
	"github.com/microserviceboilerplate/web/domain/port/output"
)

// DashboardHandler handles dashboard HTTP requests
type DashboardHandler struct {
	authUseCase  input.AuthUseCase
	templateRepo output.TemplateRepository
	sessionRepo output.SessionRepository
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(
	authUseCase input.AuthUseCase,
	templateRepo output.TemplateRepository,
	sessionRepo output.SessionRepository,
) *DashboardHandler {
	return &DashboardHandler{
		authUseCase:  authUseCase,
		templateRepo: templateRepo,
		sessionRepo:  sessionRepo,
	}
}

// RegisterRoutes registers dashboard routes
func (h *DashboardHandler) RegisterRoutes(r chi.Router) {
	r.Get("/dashboard", h.Dashboard)
}

// Dashboard renders the dashboard page
func (h *DashboardHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	// Check if user is authenticated
	exists, err := h.authUseCase.ValidateSession(ctx, "")
	if err != nil || !exists {
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}

	// Get user from session
	user, err := h.authUseCase.GetUser(ctx, "")
	if err != nil {
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}

	data := map[string]interface{}{
		"Title":       "Dashboard",
		"CurrentPage": "dashboard",
		"User":        user,
	}

	html, err := h.templateRepo.RenderString(ctx, "dashboard.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

