package logger

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// New creates a new zerolog logger with the specified level and format
func New(level, format string) zerolog.Logger {
	return NewWithService(level, format, "")
}

// NewWithService creates a new zerolog logger with a service name prefix
func NewWithService(level, format, service string) zerolog.Logger {
	return NewWithServiceAndOutput(level, format, service, os.Stdout)
}

// NewWithOutput creates a new zerolog logger with custom output writer
func NewWithOutput(level, format string, w io.Writer) zerolog.Logger {
	return NewWithServiceAndOutput(level, format, "", w)
}

// NewWithServiceAndOutput creates a new zerolog logger with service name and custom output
func NewWithServiceAndOutput(level, format, service string, w io.Writer) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	var logger zerolog.Logger

	if format == "console" {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: w}).
			Level(logLevel).
			With().
			Timestamp().
			Logger()
	} else {
		logger = zerolog.New(w).
			Level(logLevel).
			With().
			Timestamp().
			Caller().
			Logger()
	}

	// Add service name if provided
	if service != "" {
		logger = logger.With().Str("service", service).Logger()
	}

	return logger
}

// WithContext extracts the logger from context or returns the global logger
func WithContext(ctx context.Context) *zerolog.Logger {
	logger := zerolog.Ctx(ctx)
	if logger.GetLevel() == zerolog.Disabled {
		l := log.Logger
		return &l
	}
	return logger
}

// AddToContext adds the logger to the context
func AddToContext(ctx context.Context, logger *zerolog.Logger) context.Context {
	return logger.WithContext(ctx)
}

// AddTraceID adds a trace ID to the context logger
func AddTraceID(ctx context.Context, traceID string) context.Context {
	logger := log.With().Str("trace_id", traceID).Logger()
	return logger.WithContext(ctx)
}

// AddRequestID adds a request ID to the context logger
func AddRequestID(ctx context.Context, requestID string) context.Context {
	logger := log.With().Str("request_id", requestID).Logger()
	return logger.WithContext(ctx)
}
