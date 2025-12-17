package service

import (
	"context"
	"fmt"

	natsclient "github.com/microserviceboilerplate/web/adapters/output/nats"
)

// JobService handles job-related operations
type JobService struct {
	jobClient *natsclient.JobClient
}

// NewJobService creates a new job service
func NewJobService(jobClient *natsclient.JobClient) *JobService {
	return &JobService{
		jobClient: jobClient,
	}
}

// ListJobs retrieves a list of jobs
func (s *JobService) ListJobs(ctx context.Context) (*natsclient.ListJobsResponse, error) {
	resp, err := s.jobClient.ListJobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	return resp, nil
}

// GetJob retrieves a job by ID
func (s *JobService) GetJob(ctx context.Context, id string) (*natsclient.Job, error) {
	if id == "" {
		return nil, fmt.Errorf("job ID is required")
	}

	job, err := s.jobClient.GetJob(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return job, nil
}

// GetJobExecutions retrieves execution history for a job
func (s *JobService) GetJobExecutions(ctx context.Context, jobID string, limit int) (*natsclient.JobExecutionsResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	resp, err := s.jobClient.GetJobExecutions(ctx, jobID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get job executions: %w", err)
	}

	return resp, nil
}

