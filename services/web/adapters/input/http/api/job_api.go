package api

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	natsclient "github.com/microserviceboilerplate/web/adapters/output/nats"
	"github.com/microserviceboilerplate/web/domain/port/input"
	jobservice "github.com/microserviceboilerplate/web/domain/service"
)

// JobAPIHandler handles job API requests
type JobAPIHandler struct {
	authUseCase input.AuthUseCase
	jobService  *jobservice.JobService
}

// NewJobAPIHandler creates a new job API handler
func NewJobAPIHandler(authUseCase input.AuthUseCase, jobService *jobservice.JobService) *JobAPIHandler {
	return &JobAPIHandler{
		authUseCase: authUseCase,
		jobService:  jobService,
	}
}

// RegisterRoutes registers job API routes
func (h *JobAPIHandler) RegisterRoutes(r chi.Router) {
	r.Get("/jobs/table", h.ListJobs)
	r.Post("/jobs/{id}/start", h.StartJob)
	r.Post("/jobs/{id}/stop", h.StopJob)
}

// ListJobs returns jobs table HTML
func (h *JobAPIHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	// Check authentication
	exists, _ := h.authUseCase.ValidateSession(ctx, "")
	if !exists {
		errorHTML := h.renderErrorHTML("Unauthorized. Please log in.")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errorHTML)
		return
	}

	// Call service - returns *natsclient.ListJobsResponse
	resp, err := h.jobService.ListJobs(ctx)
	if err != nil {
		// Return error as HTML fragment for HTMX
		errorMsg := "Failed to load jobs"
		if err != nil {
			errorMsg += ": " + err.Error()
		}
		errorHTML := h.renderErrorHTML(errorMsg)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, errorHTML)
		return
	}

	// Render jobs table
	tableHTML := h.renderJobsTableHTML(resp)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, tableHTML)
}

func (h *JobAPIHandler) renderJobsTableHTML(resp *natsclient.ListJobsResponse) string {
	if resp == nil || len(resp.Jobs) == 0 {
		return `<table><thead><tr><th>Name</th><th>Status</th><th>Flink Job ID</th><th>Created At</th><th>Actions</th></tr></thead><tbody><tr><td colspan="5" class="empty-state">No jobs found yet. Create a job in the orchestrator to see it listed here.</td></tr></tbody></table>`
	}

	htmlStr := `<table><thead><tr><th>Name</th><th>Status</th><th>Flink Job ID</th><th>Created At</th><th>Actions</th></tr></thead><tbody>`
	for _, job := range resp.Jobs {
		createdAt := time.Unix(job.CreatedAt, 0).Format("2006-01-02 15:04:05")
		// Escape all user-generated content to prevent XSS
		jobID := html.EscapeString(job.ID)
		jobName := html.EscapeString(job.Name)
		jobStatus := html.EscapeString(job.Status)
		flinkJobID := html.EscapeString(job.FlinkJobID)

		// Sanitize status for CSS class (only allow alphanumeric and hyphens)
		statusClass := "status-" + strings.ToLower(strings.ReplaceAll(jobStatus, " ", "-"))
		// Remove any potentially dangerous characters
		statusClass = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
				return r
			}
			return -1
		}, statusClass)

		htmlStr += fmt.Sprintf(`<tr><td>%s</td><td><span class="status-badge %s">%s</span></td><td>%s</td><td>%s</td><td><button class="action-btn btn-start" hx-post="/jobs/%s/start" hx-target="#jobs-table" hx-swap="innerHTML" hx-trigger="click">Start</button><button class="action-btn btn-stop" hx-post="/jobs/%s/stop" hx-target="#jobs-table" hx-swap="innerHTML" hx-trigger="click">Stop</button></td></tr>`,
			jobName, statusClass, jobStatus, flinkJobID, html.EscapeString(createdAt), jobID, jobID)
	}
	htmlStr += `</tbody></table>`
	return htmlStr
}

func (h *JobAPIHandler) renderErrorHTML(message string) string {
	return fmt.Sprintf(`<div class="empty-state" style="padding: 40px; text-align: center; color: #ef4444;">%s</div>`, html.EscapeString(message))
}

// StartJob starts a job
func (h *JobAPIHandler) StartJob(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	// Check authentication
	exists, _ := h.authUseCase.ValidateSession(ctx, "")
	if !exists {
		errorHTML := h.renderErrorHTML("Unauthorized. Please log in.")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errorHTML)
		return
	}

	w.Header().Set("HX-Trigger", "jobsUpdated")
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// StopJob stops a job
func (h *JobAPIHandler) StopJob(w http.ResponseWriter, r *http.Request) {
	ctx := context.WithValue(r.Context(), "request", r)
	ctx = context.WithValue(ctx, "response", w)

	// Check authentication
	exists, _ := h.authUseCase.ValidateSession(ctx, "")
	if !exists {
		errorHTML := h.renderErrorHTML("Unauthorized. Please log in.")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, errorHTML)
		return
	}

	w.Header().Set("HX-Trigger", "jobsUpdated")
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
