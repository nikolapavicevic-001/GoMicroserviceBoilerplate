package input

import (
	"context"
	"github.com/microserviceboilerplate/job-orchestrator/domain/entity"
)

// JobUseCase defines the job management use cases
type JobUseCase interface {
	// CreateJob creates a new job definition
	CreateJob(ctx context.Context, req CreateJobRequest) (*entity.Job, error)
	
	// GetJob retrieves a job by ID
	GetJob(ctx context.Context, id string) (*entity.Job, error)
	
	// ListJobs lists all jobs
	ListJobs(ctx context.Context) ([]*entity.Job, error)
	
	// UpdateJob updates an existing job
	UpdateJob(ctx context.Context, id string, req UpdateJobRequest) (*entity.Job, error)
	
	// DeleteJob deletes a job
	DeleteJob(ctx context.Context, id string) error
	
	// StartJob starts a job immediately
	StartJob(ctx context.Context, id string) (*entity.JobExecution, error)
	
	// StopJob stops a running job
	StopJob(ctx context.Context, id string) error
	
	// GetJobExecutions retrieves execution history for a job
	GetJobExecutions(ctx context.Context, jobID string) ([]*entity.JobExecution, error)
	
	// SetSchedule sets a cron schedule for a job
	SetSchedule(ctx context.Context, id string, cronExpr string) error
	
	// RemoveSchedule removes the schedule from a job
	RemoveSchedule(ctx context.Context, id string) error
}

// CreateJobRequest represents a request to create a job
type CreateJobRequest struct {
	Name        string
	JarURL      string
	EntryClass  string
	Config      map[string]interface{}
	ScheduleCron *string
}

// UpdateJobRequest represents a request to update a job
type UpdateJobRequest struct {
	Name        *string
	JarURL      *string
	EntryClass  *string
	Config      *map[string]interface{}
	ScheduleCron *string
	Enabled     *bool
}

