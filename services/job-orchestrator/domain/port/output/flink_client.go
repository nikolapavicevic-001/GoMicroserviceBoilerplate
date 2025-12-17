package output

import (
	"context"
	"github.com/microserviceboilerplate/job-orchestrator/domain/entity"
)

// FlinkClient defines the interface for interacting with Flink REST API
type FlinkClient interface {
	// UploadJAR uploads a JAR file to Flink and returns the JAR ID
	UploadJAR(ctx context.Context, jarData []byte, jarName string) (string, error)
	
	// SubmitJob submits a job to Flink using a JAR ID
	SubmitJob(ctx context.Context, jarID string, entryClass string, programArgs map[string]interface{}) (string, error)
	
	// GetJobStatus retrieves the status of a Flink job
	GetJobStatus(ctx context.Context, flinkJobID string) (entity.JobStatus, error)
	
	// CancelJob cancels a running Flink job
	CancelJob(ctx context.Context, flinkJobID string) error
	
	// ListJobs lists all running jobs in Flink
	ListJobs(ctx context.Context) ([]FlinkJobInfo, error)
	
	// DeleteJAR deletes a JAR from Flink
	DeleteJAR(ctx context.Context, jarID string) error
}

// FlinkJobInfo represents information about a Flink job
type FlinkJobInfo struct {
	JobID  string
	Status string
	Name   string
}

