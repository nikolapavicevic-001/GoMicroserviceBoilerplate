package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/microserviceboilerplate/job-orchestrator/domain/entity"
	"github.com/microserviceboilerplate/job-orchestrator/domain/port/input"
	"github.com/microserviceboilerplate/job-orchestrator/domain/port/output"
)

// JobService implements the JobUseCase interface
type JobService struct {
	repo        output.JobRepository
	flinkClient output.FlinkClient
	dataProcessorURL string
	httpClient  *http.Client
}

// NewJobService creates a new job service
func NewJobService(
	repo output.JobRepository,
	flinkClient output.FlinkClient,
	dataProcessorURL string,
) input.JobUseCase {
	return &JobService{
		repo:             repo,
		flinkClient:      flinkClient,
		dataProcessorURL: dataProcessorURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateJob creates a new job definition
func (s *JobService) CreateJob(ctx context.Context, req input.CreateJobRequest) (*entity.Job, error) {
	// Check if job with same name already exists
	existing, err := s.repo.GetByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, entity.ErrJobAlreadyExists
	}

	job := &entity.Job{
		ID:          uuid.New().String(),
		Name:        req.Name,
		JarURL:      req.JarURL,
		EntryClass:  req.EntryClass,
		Config:      req.Config,
		ScheduleCron: req.ScheduleCron,
		Enabled:     true,
	}

	if job.EntryClass == "" {
		job.EntryClass = "com.microserviceboilerplate.DataProcessor"
	}

	if err := job.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return job, nil
}

// GetJob retrieves a job by ID
func (s *JobService) GetJob(ctx context.Context, id string) (*entity.Job, error) {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, entity.ErrJobNotFound
	}
	return job, nil
}

// ListJobs lists all jobs
func (s *JobService) ListJobs(ctx context.Context) ([]*entity.Job, error) {
	return s.repo.List(ctx)
}

// UpdateJob updates an existing job
func (s *JobService) UpdateJob(ctx context.Context, id string, req input.UpdateJobRequest) (*entity.Job, error) {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, entity.ErrJobNotFound
	}

	if req.Name != nil {
		// Check if new name conflicts with existing job
		if *req.Name != job.Name {
			existing, err := s.repo.GetByName(ctx, *req.Name)
			if err == nil && existing != nil {
				return nil, entity.ErrJobAlreadyExists
			}
		}
		job.Name = *req.Name
	}
	if req.JarURL != nil {
		job.JarURL = *req.JarURL
	}
	if req.EntryClass != nil {
		job.EntryClass = *req.EntryClass
	}
	if req.Config != nil {
		job.Config = *req.Config
	}
	if req.ScheduleCron != nil {
		job.ScheduleCron = req.ScheduleCron
	}
	if req.Enabled != nil {
		job.Enabled = *req.Enabled
	}

	if err := job.Validate(); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to update job: %w", err)
	}

	return job, nil
}

// DeleteJob deletes a job
func (s *JobService) DeleteJob(ctx context.Context, id string) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return entity.ErrJobNotFound
	}

	return s.repo.Delete(ctx, id)
}

// StartJob starts a job immediately
func (s *JobService) StartJob(ctx context.Context, id string) (*entity.JobExecution, error) {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, entity.ErrJobNotFound
	}

	if !job.Enabled {
		return nil, fmt.Errorf("job is disabled")
	}

	// Create execution record
	execution := &entity.JobExecution{
		ID:        uuid.New().String(),
		JobID:     job.ID,
		Status:    entity.JobStatusSubmitted,
		StartedAt: timePtr(time.Now()),
	}

	if err := s.repo.CreateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to create execution: %w", err)
	}

	// Download JAR from data-processor service
	jarURL := job.JarURL
	if jarURL == "" || jarURL == "data-processor.jar" {
		jarURL = fmt.Sprintf("%s/data-processor.jar", s.dataProcessorURL)
	}

	jarData, jarName, err := s.downloadJAR(ctx, jarURL)
	if err != nil {
		execution.Status = entity.JobStatusFailed
		errorMsg := fmt.Sprintf("failed to download JAR: %v", err)
		execution.ErrorMessage = &errorMsg
		s.repo.UpdateExecution(ctx, execution)
		return nil, fmt.Errorf("failed to download JAR: %w", err)
	}

	// Upload JAR to Flink
	jarID, err := s.flinkClient.UploadJAR(ctx, jarData, jarName)
	if err != nil {
		execution.Status = entity.JobStatusFailed
		errorMsg := fmt.Sprintf("failed to upload JAR to Flink: %v", err)
		execution.ErrorMessage = &errorMsg
		s.repo.UpdateExecution(ctx, execution)
		return nil, fmt.Errorf("failed to upload JAR to Flink: %w", err)
	}

	execution.FlinkJarID = &jarID

	// Submit job to Flink
	flinkJobID, err := s.flinkClient.SubmitJob(ctx, jarID, job.EntryClass, job.Config)
	if err != nil {
		execution.Status = entity.JobStatusFailed
		errorMsg := fmt.Sprintf("failed to submit job to Flink: %v", err)
		execution.ErrorMessage = &errorMsg
		s.repo.UpdateExecution(ctx, execution)
		return nil, fmt.Errorf("failed to submit job to Flink: %w", err)
	}

	execution.FlinkJobID = &flinkJobID
	execution.Status = entity.JobStatusRunning

	if err := s.repo.UpdateExecution(ctx, execution); err != nil {
		return nil, fmt.Errorf("failed to update execution: %w", err)
	}

	return execution, nil
}

// StopJob stops a running job
func (s *JobService) StopJob(ctx context.Context, id string) error {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return entity.ErrJobNotFound
	}

	// Find the latest running execution
	executions, err := s.repo.GetExecutionsByJobID(ctx, job.ID)
	if err != nil {
		return fmt.Errorf("failed to get executions: %w", err)
	}

	var runningExecution *entity.JobExecution
	for _, exec := range executions {
		if exec.IsRunning() {
			runningExecution = exec
			break
		}
	}

	if runningExecution == nil || runningExecution.FlinkJobID == nil {
		return fmt.Errorf("no running job found")
	}

	// Cancel job in Flink
	if err := s.flinkClient.CancelJob(ctx, *runningExecution.FlinkJobID); err != nil {
		return fmt.Errorf("failed to cancel job in Flink: %w", err)
	}

	// Update execution status
	now := time.Now()
	runningExecution.Status = entity.JobStatusCancelled
	runningExecution.FinishedAt = &now

	return s.repo.UpdateExecution(ctx, runningExecution)
}

// GetJobExecutions retrieves execution history for a job
func (s *JobService) GetJobExecutions(ctx context.Context, jobID string) ([]*entity.JobExecution, error) {
	_, err := s.repo.GetByID(ctx, jobID)
	if err != nil {
		return nil, entity.ErrJobNotFound
	}

	return s.repo.GetExecutionsByJobID(ctx, jobID)
}

// SetSchedule sets a cron schedule for a job
func (s *JobService) SetSchedule(ctx context.Context, id string, cronExpr string) error {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return entity.ErrJobNotFound
	}

	job.ScheduleCron = &cronExpr
	return s.repo.Update(ctx, job)
}

// RemoveSchedule removes the schedule from a job
func (s *JobService) RemoveSchedule(ctx context.Context, id string) error {
	job, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return entity.ErrJobNotFound
	}

	job.ScheduleCron = nil
	return s.repo.Update(ctx, job)
}

// downloadJAR downloads a JAR file from the given URL
func (s *JobService) downloadJAR(ctx context.Context, url string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download JAR: status %d", resp.StatusCode)
	}

	jarData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// Extract filename from URL or use default
	jarName := "data-processor.jar"
	if resp.Header.Get("Content-Disposition") != "" {
		// Try to extract filename from Content-Disposition header
		// For now, use default
	}

	return jarData, jarName, nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}

