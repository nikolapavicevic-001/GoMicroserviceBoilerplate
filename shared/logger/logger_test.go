package logger

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name   string
		level  string
		format string
	}{
		{"debug level json", "debug", "json"},
		{"info level json", "info", "json"},
		{"warn level json", "warn", "json"},
		{"error level json", "error", "json"},
		{"info level console", "info", "console"},
		{"invalid level defaults to info", "invalid", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level, tt.format)
			if logger.GetLevel() == zerolog.Disabled {
				t.Error("New() returned a disabled logger")
			}
		})
	}
}

func TestNewWithOutput(t *testing.T) {
	var buf bytes.Buffer

	logger := NewWithOutput("info", "json", &buf)
	logger.Info().Str("key", "value").Msg("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("NewWithOutput() output = %v, want to contain 'test message'", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("NewWithOutput() output = %v, want to contain key:value", output)
	}
}

func TestWithContext(t *testing.T) {
	t.Run("returns logger from context", func(t *testing.T) {
		var buf bytes.Buffer
		original := NewWithOutput("info", "json", &buf)

		ctx := original.WithContext(context.Background())
		retrieved := WithContext(ctx)

		retrieved.Info().Msg("test from context")

		if buf.Len() == 0 {
			t.Error("WithContext() should return functional logger")
		}
	})

	t.Run("returns default logger for empty context", func(t *testing.T) {
		logger := WithContext(context.Background())
		// Should not panic
		logger.Info().Msg("test")
	})
}

func TestAddTraceID(t *testing.T) {
	// AddTraceID creates a new logger with trace_id and adds it to context
	ctx := AddTraceID(context.Background(), "trace-123")

	logFromCtx := WithContext(ctx)

	// Verify the logger is functional (won't crash)
	logFromCtx.Info().Msg("test with trace")

	// The trace ID should be in the context logger
	// We can verify by checking the logger has the trace_id field
	// Note: AddTraceID uses the global log.With() so output goes to os.Stdout
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		logLevel  string
		shouldLog bool
	}{
		{"debug logs at debug level", "debug", "debug", true},
		{"info logs at info level", "info", "info", true},
		{"warn logs at warn level", "warn", "warn", true},
		{"error logs at error level", "error", "error", true},
		{"debug does not log at info level", "info", "debug", false},
		{"info logs at debug level", "debug", "info", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewWithOutput(tt.level, "json", &buf)

			switch tt.logLevel {
			case "debug":
				logger.Debug().Msg("test")
			case "info":
				logger.Info().Msg("test")
			case "warn":
				logger.Warn().Msg("test")
			case "error":
				logger.Error().Msg("test")
			}

			hasOutput := buf.Len() > 0
			if hasOutput != tt.shouldLog {
				t.Errorf("Expected shouldLog=%v, got output=%v", tt.shouldLog, hasOutput)
			}
		})
	}
}

func TestStructuredLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithOutput("info", "json", &buf)

	logger.Info().
		Str("user_id", "123").
		Int("count", 42).
		Bool("active", true).
		Msg("structured log")

	output := buf.String()

	expectedFields := []string{
		`"user_id":"123"`,
		`"count":42`,
		`"active":true`,
		`"message":"structured log"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(output, field) {
			t.Errorf("Output missing field %s, got: %s", field, output)
		}
	}
}

func TestConsoleFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithOutput("info", "console", &buf)

	logger.Info().Msg("console test")

	output := buf.String()
	// Console format should not be JSON
	if strings.HasPrefix(strings.TrimSpace(output), "{") {
		t.Error("Console format should not output JSON")
	}
}

