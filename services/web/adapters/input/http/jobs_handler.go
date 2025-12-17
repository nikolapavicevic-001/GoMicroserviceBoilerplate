package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/microserviceboilerplate/web/domain/port/input"
	"github.com/microserviceboilerplate/web/domain/port/output"
)

// JobsHandler handles jobs page requests
type JobsHandler struct {
	authUseCase  input.AuthUseCase
	templateRepo output.TemplateRepository
	sessionRepo  output.SessionRepository
}

// NewJobsHandler creates a new jobs handler
func NewJobsHandler(
	authUseCase input.AuthUseCase,
	templateRepo output.TemplateRepository,
	sessionRepo output.SessionRepository,
) *JobsHandler {
	return &JobsHandler{
		authUseCase:  authUseCase,
		templateRepo: templateRepo,
		sessionRepo:  sessionRepo,
	}
}

// RegisterRoutes registers jobs routes
func (h *JobsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/jobs", h.Jobs)
}

// Jobs renders the jobs page
func (h *JobsHandler) Jobs(w http.ResponseWriter, r *http.Request) {
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
		"Title":       "Jobs",
		"CurrentPage": "jobs",
		"User":        user,
	}

	html, err := h.templateRepo.RenderString(ctx, "jobs.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

