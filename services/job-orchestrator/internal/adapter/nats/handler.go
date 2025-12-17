package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/microserviceboilerplate/job-orchestrator/domain/port/input"
)

// Handler handles NATS request-reply messages for job operations
type Handler struct {
	jobUseCase input.JobUseCase
	nc         *nats.Conn
}

// NewHandler creates a new NATS handler
func NewHandler(nc *nats.Conn, jobUseCase input.JobUseCase) *Handler {
	return &Handler{
		jobUseCase: jobUseCase,
		nc:         nc,
	}
}

// RegisterHandlers registers all NATS request handlers
func (h *Handler) RegisterHandlers() error {
	// List jobs handler
	_, err := h.nc.Subscribe("request.job.list", h.handleListJobs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to request.job.list: %w", err)
	}

	// Get job handler
	_, err = h.nc.Subscribe("request.job.get", h.handleGetJob)
	if err != nil {
		return fmt.Errorf("failed to subscribe to request.job.get: %w", err)
	}

	// Get job executions handler
	_, err = h.nc.Subscribe("request.job.executions", h.handleGetJobExecutions)
	if err != nil {
		return fmt.Errorf("failed to subscribe to request.job.executions: %w", err)
	}

	return nil
}

// handleListJobs handles job list requests
func (h *Handler) handleListJobs(msg *nats.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jobs, err := h.jobUseCase.ListJobs(ctx)
	if err != nil {
		h.sendError(msg.Reply, "failed to list jobs", err)
		return
	}

	// Convert jobs to JSON-serializable format
	jobList := make([]map[string]interface{}, len(jobs))
	for i, job := range jobs {
		// Get latest execution to determine status
		executions, _ := h.jobUseCase.GetJobExecutions(ctx, job.ID)
		status := "UNKNOWN"
		flinkJobID := ""
		if len(executions) > 0 {
			latestExec := executions[0] // First one is latest (ordered by created_at DESC)
			status = string(latestExec.Status)
			if latestExec.FlinkJobID != nil {
				flinkJobID = *latestExec.FlinkJobID
			}
		}

		jobList[i] = map[string]interface{}{
			"id":            job.ID,
			"name":          job.Name,
			"status":        status,
			"flink_job_id":  flinkJobID,
			"enabled":       job.Enabled,
			"schedule_cron": job.ScheduleCron,
			"created_at":    job.CreatedAt.Unix(),
			"updated_at":    job.UpdatedAt.Unix(),
		}
	}

	response := map[string]interface{}{
		"jobs": jobList,
		"total": len(jobList),
	}

	data, err := json.Marshal(response)
	if err != nil {
		h.sendError(msg.Reply, "failed to marshal response", err)
		return
	}

	if msg.Reply != "" {
		msg.Respond(data)
	}
}

// handleGetJob handles single job requests
func (h *Handler) handleGetJob(msg *nats.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req struct {
		ID string `json:"id"`
	}

	if err := json.Unmarshal(msg.Data, &req); err != nil {
		h.sendError(msg.Reply, "invalid request", err)
		return
	}

	if req.ID == "" {
		h.sendError(msg.Reply, "job ID is required", fmt.Errorf("empty ID"))
		return
	}

	job, err := h.jobUseCase.GetJob(ctx, req.ID)
	if err != nil {
		h.sendError(msg.Reply, "failed to get job", err)
		return
	}

	executions, _ := h.jobUseCase.GetJobExecutions(ctx, job.ID)
	status := "UNKNOWN"
	flinkJobID := ""
	if len(executions) > 0 {
		latestExec := executions[0] // First one is latest
		status = string(latestExec.Status)
		if latestExec.FlinkJobID != nil {
			flinkJobID = *latestExec.FlinkJobID
		}
	}

	response := map[string]interface{}{
		"id":            job.ID,
		"name":          job.Name,
		"jar_url":       job.JarURL,
		"entry_class":   job.EntryClass,
		"config":        job.Config,
		"status":        status,
		"flink_job_id":  flinkJobID,
		"enabled":       job.Enabled,
		"schedule_cron": job.ScheduleCron,
		"created_at":    job.CreatedAt.Unix(),
		"updated_at":    job.UpdatedAt.Unix(),
	}

	data, err := json.Marshal(response)
	if err != nil {
		h.sendError(msg.Reply, "failed to marshal response", err)
		return
	}

	if msg.Reply != "" {
		msg.Respond(data)
	}
}

// handleGetJobExecutions handles job execution history requests
func (h *Handler) handleGetJobExecutions(msg *nats.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var req struct {
		JobID string `json:"job_id"`
		Limit int    `json:"limit"`
	}

	if len(msg.Data) > 0 {
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			h.sendError(msg.Reply, "invalid request", err)
			return
		}
	}

	if req.JobID == "" {
		h.sendError(msg.Reply, "job ID is required", fmt.Errorf("empty job_id"))
		return
	}

	executions, err := h.jobUseCase.GetJobExecutions(ctx, req.JobID)
	if err != nil {
		h.sendError(msg.Reply, "failed to get executions", err)
		return
	}

	// Apply limit if specified
	if req.Limit > 0 && len(executions) > req.Limit {
		executions = executions[:req.Limit]
	}
	if err != nil {
		h.sendError(msg.Reply, "failed to get executions", err)
		return
	}

	execList := make([]map[string]interface{}, len(executions))
	for i, exec := range executions {
		execMap := map[string]interface{}{
			"id":         exec.ID,
			"job_id":     exec.JobID,
			"status":     string(exec.Status),
			"created_at": exec.CreatedAt.Unix(),
		}

		if exec.StartedAt != nil {
			execMap["started_at"] = exec.StartedAt.Unix()
		}
		if exec.FinishedAt != nil {
			execMap["completed_at"] = exec.FinishedAt.Unix()
		}
		if exec.FlinkJobID != nil {
			execMap["flink_job_id"] = *exec.FlinkJobID
		}
		if exec.ErrorMessage != nil {
			execMap["error_message"] = *exec.ErrorMessage
		}

		execList[i] = execMap
	}

	response := map[string]interface{}{
		"executions": execList,
		"total":      len(execList),
	}

	data, err := json.Marshal(response)
	if err != nil {
		h.sendError(msg.Reply, "failed to marshal response", err)
		return
	}

	if msg.Reply != "" {
		msg.Respond(data)
	}
}


// sendError sends an error response
func (h *Handler) sendError(reply string, message string, err error) {
	if reply == "" {
		return
	}

	errorResp := map[string]interface{}{
		"error":   message,
		"details": err.Error(),
	}

	data, _ := json.Marshal(errorResp)
	h.nc.Publish(reply, data)
}

