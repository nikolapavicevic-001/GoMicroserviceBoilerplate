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
	return NewWithOutput(level, format, os.Stdout)
}

// NewWithOutput creates a new zerolog logger with custom output writer
func NewWithOutput(level, format string, w io.Writer) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	if format == "console" {
		return zerolog.New(zerolog.ConsoleWriter{Out: w}).
			Level(logLevel).
			With().
			Timestamp().
			Logger()
	}

	return zerolog.New(w).
		Level(logLevel).
		With().
		Timestamp().
		Caller().
		Logger()
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
