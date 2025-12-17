package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/microserviceboilerplate/job-orchestrator/domain/port/input"
)

// JobHandler handles job management HTTP requests
type JobHandler struct {
	jobUseCase input.JobUseCase
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobUseCase input.JobUseCase) *JobHandler {
	return &JobHandler{
		jobUseCase: jobUseCase,
	}
}

// RegisterRoutes registers job management routes
func (h *JobHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/jobs", func(r chi.Router) {
		r.Post("/", h.CreateJob)
		r.Get("/", h.ListJobs)
		r.Get("/{id}", h.GetJob)
		r.Put("/{id}", h.UpdateJob)
		r.Delete("/{id}", h.DeleteJob)
		r.Post("/{id}/start", h.StartJob)
		r.Post("/{id}/stop", h.StopJob)
		r.Get("/{id}/executions", h.GetJobExecutions)
		r.Post("/{id}/schedule", h.SetSchedule)
		r.Delete("/{id}/schedule", h.RemoveSchedule)
	})
}

// CreateJob handles job creation
func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req input.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job, err := h.jobUseCase.CreateJob(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(job)
}

// ListJobs handles listing all jobs
func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.jobUseCase.ListJobs(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

// GetJob handles getting a job by ID
func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	job, err := h.jobUseCase.GetJob(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// UpdateJob handles updating a job
func (h *JobHandler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req input.UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job, err := h.jobUseCase.UpdateJob(r.Context(), id, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

// DeleteJob handles deleting a job
func (h *JobHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.jobUseCase.DeleteJob(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// StartJob handles starting a job
func (h *JobHandler) StartJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	execution, err := h.jobUseCase.StartJob(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(execution)
}

// StopJob handles stopping a job
func (h *JobHandler) StopJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.jobUseCase.StopJob(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetJobExecutions handles getting job execution history
func (h *JobHandler) GetJobExecutions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	executions, err := h.jobUseCase.GetJobExecutions(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(executions)
}

// SetSchedule handles setting a schedule for a job
func (h *JobHandler) SetSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		Cron string `json:"cron"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.jobUseCase.SetSchedule(r.Context(), id, req.Cron); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// RemoveSchedule handles removing a schedule from a job
func (h *JobHandler) RemoveSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.jobUseCase.RemoveSchedule(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

