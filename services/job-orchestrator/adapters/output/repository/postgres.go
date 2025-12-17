package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/microserviceboilerplate/job-orchestrator/domain/entity"
	"github.com/microserviceboilerplate/job-orchestrator/domain/port/output"
)

// PostgresJobRepository implements the JobRepository interface using PostgreSQL
type PostgresJobRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresJobRepository creates a new PostgreSQL job repository
func NewPostgresJobRepository(pool *pgxpool.Pool) output.JobRepository {
	return &PostgresJobRepository{
		pool: pool,
	}
}

// Create creates a new job
func (r *PostgresJobRepository) Create(ctx context.Context, job *entity.Job) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	now := time.Now()
	job.CreatedAt = now
	job.UpdatedAt = now

	configJSON, err := json.Marshal(job.Config)
	if err != nil {
		return err
	}

	var scheduleCron sql.NullString
	if job.ScheduleCron != nil {
		scheduleCron = sql.NullString{String: *job.ScheduleCron, Valid: true}
	}

	query := `
		INSERT INTO jobs (id, name, jar_url, entry_class, config, schedule_cron, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.pool.Exec(ctx, query,
		job.ID, job.Name, job.JarURL, job.EntryClass, configJSON,
		scheduleCron, job.Enabled, job.CreatedAt, job.UpdatedAt)

	return err
}

// GetByID retrieves a job by ID
func (r *PostgresJobRepository) GetByID(ctx context.Context, id string) (*entity.Job, error) {
	query := `
		SELECT id, name, jar_url, entry_class, config, schedule_cron, enabled, created_at, updated_at
		FROM jobs
		WHERE id = $1
	`

	var job entity.Job
	var configJSON []byte
	var scheduleCron sql.NullString

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&job.ID, &job.Name, &job.JarURL, &job.EntryClass, &configJSON,
		&scheduleCron, &job.Enabled, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configJSON, &job.Config); err != nil {
		return nil, err
	}

	if scheduleCron.Valid {
		job.ScheduleCron = &scheduleCron.String
	}

	return &job, nil
}

// GetByName retrieves a job by name
func (r *PostgresJobRepository) GetByName(ctx context.Context, name string) (*entity.Job, error) {
	query := `
		SELECT id, name, jar_url, entry_class, config, schedule_cron, enabled, created_at, updated_at
		FROM jobs
		WHERE name = $1
	`

	var job entity.Job
	var configJSON []byte
	var scheduleCron sql.NullString

	err := r.pool.QueryRow(ctx, query, name).Scan(
		&job.ID, &job.Name, &job.JarURL, &job.EntryClass, &configJSON,
		&scheduleCron, &job.Enabled, &job.CreatedAt, &job.UpdatedAt)

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(configJSON, &job.Config); err != nil {
		return nil, err
	}

	if scheduleCron.Valid {
		job.ScheduleCron = &scheduleCron.String
	}

	return &job, nil
}

// List retrieves all jobs
func (r *PostgresJobRepository) List(ctx context.Context) ([]*entity.Job, error) {
	query := `
		SELECT id, name, jar_url, entry_class, config, schedule_cron, enabled, created_at, updated_at
		FROM jobs
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*entity.Job
	for rows.Next() {
		var job entity.Job
		var configJSON []byte
		var scheduleCron sql.NullString

		if err := rows.Scan(
			&job.ID, &job.Name, &job.JarURL, &job.EntryClass, &configJSON,
			&scheduleCron, &job.Enabled, &job.CreatedAt, &job.UpdatedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(configJSON, &job.Config); err != nil {
			return nil, err
		}

		if scheduleCron.Valid {
			job.ScheduleCron = &scheduleCron.String
		}

		jobs = append(jobs, &job)
	}

	return jobs, rows.Err()
}

// Update updates an existing job
func (r *PostgresJobRepository) Update(ctx context.Context, job *entity.Job) error {
	job.UpdatedAt = time.Now()

	configJSON, err := json.Marshal(job.Config)
	if err != nil {
		return err
	}

	var scheduleCron sql.NullString
	if job.ScheduleCron != nil {
		scheduleCron = sql.NullString{String: *job.ScheduleCron, Valid: true}
	}

	query := `
		UPDATE jobs
		SET name = $2, jar_url = $3, entry_class = $4, config = $5, schedule_cron = $6, enabled = $7, updated_at = $8
		WHERE id = $1
	`

	_, err = r.pool.Exec(ctx, query,
		job.ID, job.Name, job.JarURL, job.EntryClass, configJSON,
		scheduleCron, job.Enabled, job.UpdatedAt)

	return err
}

// Delete deletes a job by ID
func (r *PostgresJobRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM jobs WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// CreateExecution creates a new job execution record
func (r *PostgresJobRepository) CreateExecution(ctx context.Context, execution *entity.JobExecution) error {
	if execution.ID == "" {
		execution.ID = uuid.New().String()
	}
	execution.CreatedAt = time.Now()

	var flinkJobID, flinkJarID sql.NullString
	if execution.FlinkJobID != nil {
		flinkJobID = sql.NullString{String: *execution.FlinkJobID, Valid: true}
	}
	if execution.FlinkJarID != nil {
		flinkJarID = sql.NullString{String: *execution.FlinkJarID, Valid: true}
	}

	var startedAt sql.NullTime
	if execution.StartedAt != nil {
		startedAt = sql.NullTime{Time: *execution.StartedAt, Valid: true}
	}

	var errorMessage sql.NullString
	if execution.ErrorMessage != nil {
		errorMessage = sql.NullString{String: *execution.ErrorMessage, Valid: true}
	}

	query := `
		INSERT INTO job_executions (id, job_id, flink_job_id, flink_jar_id, status, started_at, error_message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		execution.ID, execution.JobID, flinkJobID, flinkJarID,
		execution.Status, startedAt, errorMessage, execution.CreatedAt)

	return err
}

// GetExecutionsByJobID retrieves all executions for a job
func (r *PostgresJobRepository) GetExecutionsByJobID(ctx context.Context, jobID string) ([]*entity.JobExecution, error) {
	query := `
		SELECT id, job_id, flink_job_id, flink_jar_id, status, started_at, finished_at, error_message, created_at
		FROM job_executions
		WHERE job_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executions []*entity.JobExecution
	for rows.Next() {
		execution := &entity.JobExecution{}
		var flinkJobID, flinkJarID sql.NullString
		var startedAt, finishedAt sql.NullTime
		var errorMessage sql.NullString

		if err := rows.Scan(
			&execution.ID, &execution.JobID, &flinkJobID, &flinkJarID,
			&execution.Status, &startedAt, &finishedAt, &errorMessage, &execution.CreatedAt); err != nil {
			return nil, err
		}

		if flinkJobID.Valid {
			execution.FlinkJobID = &flinkJobID.String
		}
		if flinkJarID.Valid {
			execution.FlinkJarID = &flinkJarID.String
		}
		if startedAt.Valid {
			execution.StartedAt = &startedAt.Time
		}
		if finishedAt.Valid {
			execution.FinishedAt = &finishedAt.Time
		}
		if errorMessage.Valid {
			execution.ErrorMessage = &errorMessage.String
		}

		executions = append(executions, execution)
	}

	return executions, rows.Err()
}

// UpdateExecution updates a job execution
func (r *PostgresJobRepository) UpdateExecution(ctx context.Context, execution *entity.JobExecution) error {
	var flinkJobID, flinkJarID sql.NullString
	if execution.FlinkJobID != nil {
		flinkJobID = sql.NullString{String: *execution.FlinkJobID, Valid: true}
	}
	if execution.FlinkJarID != nil {
		flinkJarID = sql.NullString{String: *execution.FlinkJarID, Valid: true}
	}

	var startedAt, finishedAt sql.NullTime
	if execution.StartedAt != nil {
		startedAt = sql.NullTime{Time: *execution.StartedAt, Valid: true}
	}
	if execution.FinishedAt != nil {
		finishedAt = sql.NullTime{Time: *execution.FinishedAt, Valid: true}
	}

	var errorMessage sql.NullString
	if execution.ErrorMessage != nil {
		errorMessage = sql.NullString{String: *execution.ErrorMessage, Valid: true}
	}

	query := `
		UPDATE job_executions
		SET flink_job_id = $2, flink_jar_id = $3, status = $4, started_at = $5, finished_at = $6, error_message = $7
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		execution.ID, flinkJobID, flinkJarID, execution.Status,
		startedAt, finishedAt, errorMessage)

	return err
}

// GetExecutionByFlinkJobID retrieves an execution by Flink job ID
func (r *PostgresJobRepository) GetExecutionByFlinkJobID(ctx context.Context, flinkJobID string) (*entity.JobExecution, error) {
	query := `
		SELECT id, job_id, flink_job_id, flink_jar_id, status, started_at, finished_at, error_message, created_at
		FROM job_executions
		WHERE flink_job_id = $1
	`

	execution := &entity.JobExecution{}
	var flinkJobIDVal, flinkJarID sql.NullString
	var startedAt, finishedAt sql.NullTime
	var errorMessage sql.NullString

	err := r.pool.QueryRow(ctx, query, flinkJobID).Scan(
		&execution.ID, &execution.JobID, &flinkJobIDVal, &flinkJarID,
		&execution.Status, &startedAt, &finishedAt, &errorMessage, &execution.CreatedAt)

	if err != nil {
		return nil, err
	}

	if flinkJobIDVal.Valid {
		execution.FlinkJobID = &flinkJobIDVal.String
	}
	if flinkJarID.Valid {
		execution.FlinkJarID = &flinkJarID.String
	}
	if startedAt.Valid {
		execution.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		execution.FinishedAt = &finishedAt.Time
	}
	if errorMessage.Valid {
		execution.ErrorMessage = &errorMessage.String
	}

	return execution, nil
}

