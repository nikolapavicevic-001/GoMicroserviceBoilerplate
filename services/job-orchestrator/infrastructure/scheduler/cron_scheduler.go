package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/microserviceboilerplate/job-orchestrator/domain/entity"
	"github.com/microserviceboilerplate/job-orchestrator/domain/port/input"
	"github.com/microserviceboilerplate/job-orchestrator/domain/port/output"
)

// CronScheduler manages scheduled job execution
type CronScheduler struct {
	cron      *cron.Cron
	jobUseCase input.JobUseCase
	jobRepo   output.JobRepository
	ctx       context.Context
}

// NewCronScheduler creates a new cron scheduler
func NewCronScheduler(jobUseCase input.JobUseCase, jobRepo output.JobRepository) *CronScheduler {
	return &CronScheduler{
		cron:      cron.New(cron.WithSeconds()),
		jobUseCase: jobUseCase,
		jobRepo:   jobRepo,
	}
}

// Start starts the scheduler
func (s *CronScheduler) Start(ctx context.Context) error {
	s.ctx = ctx

	// Load all jobs with schedules
	jobs, err := s.jobRepo.List(ctx)
	if err != nil {
		return err
	}

	// Schedule all enabled jobs with cron expressions
	for _, job := range jobs {
		if job.Enabled && job.ScheduleCron != nil {
			s.scheduleJob(job)
		}
	}

	s.cron.Start()
	log.Println("Cron scheduler started")

	// Watch for new jobs or schedule changes
	go s.watchJobs(ctx)

	return nil
}

// Stop stops the scheduler
func (s *CronScheduler) Stop() {
	s.cron.Stop()
	log.Println("Cron scheduler stopped")
}

// scheduleJob schedules a single job
func (s *CronScheduler) scheduleJob(job *entity.Job) {
	if job.ScheduleCron == nil {
		return
	}

	jobID := job.ID
	jobName := job.Name

	_, err := s.cron.AddFunc(*job.ScheduleCron, func() {
		log.Printf("Triggering scheduled job: %s (ID: %s)", jobName, jobID)
		
		// Re-fetch job to get latest configuration
		latestJob, err := s.jobRepo.GetByID(s.ctx, jobID)
		if err != nil {
			log.Printf("Failed to get job %s: %v", jobID, err)
			return
		}

		if !latestJob.Enabled {
			log.Printf("Job %s is disabled, skipping", jobID)
			return
		}

		// Start the job
		_, err = s.jobUseCase.StartJob(s.ctx, jobID)
		if err != nil {
			log.Printf("Failed to start scheduled job %s: %v", jobID, err)
		} else {
			log.Printf("Successfully started scheduled job: %s", jobID)
		}
	})

	if err != nil {
		log.Printf("Failed to schedule job %s: %v", jobID, err)
	} else {
		log.Printf("Scheduled job %s with cron: %s", jobID, *job.ScheduleCron)
	}
}

// watchJobs periodically checks for new or updated jobs
func (s *CronScheduler) watchJobs(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Reload jobs and reschedule if needed
			// This is a simple implementation - in production you might want
			// to track which jobs are already scheduled to avoid duplicates
			jobs, err := s.jobRepo.List(ctx)
			if err != nil {
				log.Printf("Failed to reload jobs: %v", err)
				continue
			}

			// For simplicity, we'll just log - in production you'd want
			// to track scheduled jobs and update them
			log.Printf("Watching %d jobs for schedule changes", len(jobs))
		}
	}
}

