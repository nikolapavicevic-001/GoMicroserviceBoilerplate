package flink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/microserviceboilerplate/job-orchestrator/domain/entity"
	"github.com/microserviceboilerplate/job-orchestrator/domain/port/output"
)

// FlinkRESTClient implements the FlinkClient interface using Flink REST API
type FlinkRESTClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewFlinkRESTClient creates a new Flink REST API client
func NewFlinkRESTClient(baseURL string) output.FlinkClient {
	return &FlinkRESTClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// UploadJAR uploads a JAR file to Flink
func (c *FlinkRESTClient) UploadJAR(ctx context.Context, jarData []byte, jarName string) (string, error) {
	url := fmt.Sprintf("%s/v1/jars/upload", c.baseURL)
	
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("jarfile", jarName)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	
	if _, err := part.Write(jarData); err != nil {
		return "", fmt.Errorf("failed to write JAR data: %w", err)
	}
	
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close multipart writer: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to upload JAR: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to upload JAR: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	
	var result struct {
		Filename string `json:"filename"`
		Status   string `json:"status"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Extract JAR ID from filename (format: jar_<id>_<name>)
	jarID := result.Filename
	return jarID, nil
}

// SubmitJob submits a job to Flink
func (c *FlinkRESTClient) SubmitJob(ctx context.Context, jarID string, entryClass string, programArgs map[string]interface{}) (string, error) {
	url := fmt.Sprintf("%s/v1/jars/%s/run", c.baseURL, jarID)
	
	requestBody := map[string]interface{}{
		"entryClass": entryClass,
	}
	
	// Flink expects programArgs as a string
	// Convert map to space-separated key=value pairs
	if programArgs != nil && len(programArgs) > 0 {
		var args []string
		for k, v := range programArgs {
			args = append(args, fmt.Sprintf("--%s=%v", k, v))
		}
		if len(args) > 0 {
			// Join args with spaces
			argsStr := ""
			for i, arg := range args {
				if i > 0 {
					argsStr += " "
				}
				argsStr += arg
			}
			requestBody["programArgs"] = argsStr
		}
	}
	
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to submit job: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusAccepted {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to submit job: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	
	var result struct {
		JobID string `json:"jobid"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	return result.JobID, nil
}

// GetJobStatus retrieves the status of a Flink job
func (c *FlinkRESTClient) GetJobStatus(ctx context.Context, flinkJobID string) (entity.JobStatus, error) {
	url := fmt.Sprintf("%s/v1/jobs/%s", c.baseURL, flinkJobID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get job status: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get job status: status %d", resp.StatusCode)
	}
	
	var result struct {
		State string `json:"state"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Map Flink states to our JobStatus
	switch result.State {
	case "CREATED", "INITIALIZING":
		return entity.JobStatusSubmitted, nil
	case "RUNNING":
		return entity.JobStatusRunning, nil
	case "FINISHED":
		return entity.JobStatusFinished, nil
	case "FAILED":
		return entity.JobStatusFailed, nil
	case "CANCELED", "CANCELLING":
		return entity.JobStatusCancelled, nil
	default:
		return entity.JobStatusSubmitted, nil
	}
}

// CancelJob cancels a running Flink job
func (c *FlinkRESTClient) CancelJob(ctx context.Context, flinkJobID string) error {
	url := fmt.Sprintf("%s/v1/jobs/%s", c.baseURL, flinkJobID)
	
	req, err := http.NewRequestWithContext(ctx, "PATCH", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to cancel job: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusAccepted {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to cancel job: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	
	return nil
}

// ListJobs lists all running jobs in Flink
func (c *FlinkRESTClient) ListJobs(ctx context.Context) ([]output.FlinkJobInfo, error) {
	url := fmt.Sprintf("%s/v1/jobs", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list jobs: status %d", resp.StatusCode)
	}
	
	var result struct {
		Jobs []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Name   string `json:"name"`
		} `json:"jobs"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	jobs := make([]output.FlinkJobInfo, len(result.Jobs))
	for i, job := range result.Jobs {
		jobs[i] = output.FlinkJobInfo{
			JobID:  job.ID,
			Status: job.Status,
			Name:   job.Name,
		}
	}
	
	return jobs, nil
}

// DeleteJAR deletes a JAR from Flink
func (c *FlinkRESTClient) DeleteJAR(ctx context.Context, jarID string) error {
	url := fmt.Sprintf("%s/v1/jars/%s", c.baseURL, jarID)
	
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete JAR: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete JAR: status %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	
	return nil
}

