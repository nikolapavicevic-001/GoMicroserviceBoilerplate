// Package testutil provides utilities for testing across all services
package testutil

import (
	"context"
	"os"
	"testing"
	"time"
)

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
}

// SkipIfCI skips the test if running in CI environment
func SkipIfCI(t *testing.T) {
	t.Helper()
	if os.Getenv("CI") != "" {
		t.Skip("skipping test in CI environment")
	}
}

// SkipIfNoDocker skips the test if Docker is not available
func SkipIfNoDocker(t *testing.T) {
	t.Helper()
	if os.Getenv("SKIP_DOCKER_TESTS") != "" {
		t.Skip("skipping test - Docker tests disabled")
	}
}

// TestContext returns a context with timeout for tests
func TestContext(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// TestContextWithTimeout returns a context with custom timeout for tests
func TestContextWithTimeout(t *testing.T, timeout time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), timeout)
}

// MustGetEnv gets an environment variable or fails the test
func MustGetEnv(t *testing.T, key string) string {
	t.Helper()
	value := os.Getenv(key)
	if value == "" {
		t.Fatalf("required environment variable %s is not set", key)
	}
	return value
}

// GetEnvOrDefault gets an environment variable or returns a default value
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupEnv sets environment variables for testing and returns a cleanup function
func SetupEnv(t *testing.T, env map[string]string) func() {
	t.Helper()

	// Store original values
	original := make(map[string]string)
	for key := range env {
		original[key] = os.Getenv(key)
	}

	// Set new values
	for key, value := range env {
		os.Setenv(key, value)
	}

	// Return cleanup function
	return func() {
		for key, value := range original {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}
}

// RetryWithBackoff retries a function with exponential backoff
func RetryWithBackoff(t *testing.T, attempts int, initialDelay time.Duration, fn func() error) error {
	t.Helper()

	var err error
	delay := initialDelay

	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		if i < attempts-1 {
			t.Logf("Attempt %d failed: %v. Retrying in %v...", i+1, err, delay)
			time.Sleep(delay)
			delay *= 2
		}
	}

	return err
}

// WaitFor waits for a condition to be true or times out
func WaitFor(t *testing.T, timeout time.Duration, interval time.Duration, condition func() bool) bool {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}

// AssertEventually asserts that a condition becomes true within timeout
func AssertEventually(t *testing.T, timeout time.Duration, interval time.Duration, condition func() bool, msgAndArgs ...interface{}) {
	t.Helper()

	if !WaitFor(t, timeout, interval, condition) {
		if len(msgAndArgs) > 0 {
			t.Fatalf("Condition never became true: %v", msgAndArgs...)
		} else {
			t.Fatal("Condition never became true within timeout")
		}
	}
}

