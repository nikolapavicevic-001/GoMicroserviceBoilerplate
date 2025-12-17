package nats

import (
	"context"
	"fmt"
	"time"
)

// Job represents a job from the job orchestrator
type Job struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Status      string                 `json:"status"`
	FlinkJobID  string                 `json:"flink_job_id"`
	Enabled     bool                   `json:"enabled"`
	ScheduleCron *string               `json:"schedule_cron"`
	CreatedAt   int64                  `json:"created_at"`
	UpdatedAt   int64                  `json:"updated_at"`
}

// JobExecution represents a job execution
type JobExecution struct {
	ID          string  `json:"id"`
	JobID       string  `json:"job_id"`
	Status      string  `json:"status"`
	FlinkJobID  *string `json:"flink_job_id"`
	StartedAt   *int64  `json:"started_at"`
	CompletedAt *int64  `json:"completed_at"`
	ErrorMessage *string `json:"error_message"`
	CreatedAt   int64   `json:"created_at"`
}

// ListJobsResponse represents the response from listing jobs
type ListJobsResponse struct {
	Jobs  []*Job `json:"jobs"`
	Total int    `json:"total"`
}

// JobExecutionsResponse represents the response from getting job executions
type JobExecutionsResponse struct {
	Executions []*JobExecution `json:"executions"`
	Total      int             `json:"total"`
}

// JobClient handles job orchestrator communication via NATS
type JobClient struct {
	client *Client
}

// NewJobClient creates a new job client
func NewJobClient(client *Client) *JobClient {
	return &JobClient{client: client}
}

// ListJobs requests a list of jobs
func (c *JobClient) ListJobs(ctx context.Context) (*ListJobsResponse, error) {
	req := map[string]interface{}{}

	var resp ListJobsResponse
	if err := c.client.RequestJSON(ctx, "request.job.list", req, &resp, 10*time.Second); err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	return &resp, nil
}

// GetJob requests a single job by ID
func (c *JobClient) GetJob(ctx context.Context, id string) (*Job, error) {
	req := map[string]interface{}{
		"id": id,
	}

	var job Job
	if err := c.client.RequestJSON(ctx, "request.job.get", req, &job, 10*time.Second); err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return &job, nil
}

// GetJobExecutions requests execution history for a job
func (c *JobClient) GetJobExecutions(ctx context.Context, jobID string, limit int) (*JobExecutionsResponse, error) {
	req := map[string]interface{}{
		"job_id": jobID,
		"limit":  limit,
	}

	var resp JobExecutionsResponse
	if err := c.client.RequestJSON(ctx, "request.job.executions", req, &resp, 10*time.Second); err != nil {
		return nil, fmt.Errorf("failed to get job executions: %w", err)
	}

	return &resp, nil
}

