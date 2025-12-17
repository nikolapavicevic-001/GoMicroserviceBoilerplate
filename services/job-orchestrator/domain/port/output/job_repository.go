package output

import (
	"context"
	"github.com/microserviceboilerplate/job-orchestrator/domain/entity"
)

// JobRepository defines the interface for job persistence
type JobRepository interface {
	// Create creates a new job
	Create(ctx context.Context, job *entity.Job) error
	
	// GetByID retrieves a job by ID
	GetByID(ctx context.Context, id string) (*entity.Job, error)
	
	// GetByName retrieves a job by name
	GetByName(ctx context.Context, name string) (*entity.Job, error)
	
	// List retrieves all jobs
	List(ctx context.Context) ([]*entity.Job, error)
	
	// Update updates an existing job
	Update(ctx context.Context, job *entity.Job) error
	
	// Delete deletes a job by ID
	Delete(ctx context.Context, id string) error
	
	// CreateExecution creates a new job execution record
	CreateExecution(ctx context.Context, execution *entity.JobExecution) error
	
	// GetExecutionsByJobID retrieves all executions for a job
	GetExecutionsByJobID(ctx context.Context, jobID string) ([]*entity.JobExecution, error)
	
	// UpdateExecution updates a job execution
	UpdateExecution(ctx context.Context, execution *entity.JobExecution) error
	
	// GetExecutionByFlinkJobID retrieves an execution by Flink job ID
	GetExecutionByFlinkJobID(ctx context.Context, flinkJobID string) (*entity.JobExecution, error)
}

