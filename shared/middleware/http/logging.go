package http

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// LoggingMiddleware returns an Echo middleware that logs HTTP requests using Zerolog
func LoggingMiddleware(logger *zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Get request info
			req := c.Request()
			res := c.Response()

			// Process request
			err := next(c)

			// Calculate duration
			duration := time.Since(start)

			// Determine log level based on status code
			var logEvent *zerolog.Event
			status := res.Status
			switch {
			case status >= 500:
				logEvent = logger.Error()
			case status >= 400:
				logEvent = logger.Warn()
			default:
				logEvent = logger.Info()
			}

			// Add error if present
			if err != nil {
				logEvent = logEvent.Err(err)
			}

			// Log the request
			logEvent.
				Str("protocol", "http").
				Str("method", req.Method).
				Str("uri", req.RequestURI).
				Str("path", c.Path()).
				Int("status", status).
				Int64("size", res.Size).
				Dur("duration", duration).
				Str("ip", c.RealIP()).
				Str("user_agent", req.UserAgent()).
				Str("request_id", res.Header().Get(echo.HeaderXRequestID)).
				Msg("HTTP request completed")

			return err
		}
	}
}

// LoggingMiddlewareWithConfig returns an Echo middleware with custom configuration
type LoggingConfig struct {
	// Logger is the zerolog logger instance
	Logger *zerolog.Logger

	// SkipPaths is a list of paths to skip logging
	SkipPaths []string

	// IncludeHeaders determines whether to log request headers
	IncludeHeaders bool
}

// LoggingMiddlewareWithConfig returns a logging middleware with custom configuration
func LoggingMiddlewareWithConfig(config LoggingConfig) echo.MiddlewareFunc {
	skipMap := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipMap[path] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip logging for certain paths
			if skipMap[c.Path()] {
				return next(c)
			}

			start := time.Now()

			// Get request info
			req := c.Request()
			res := c.Response()

			// Process request
			err := next(c)

			// Calculate duration
			duration := time.Since(start)

			// Determine log level based on status code
			var logEvent *zerolog.Event
			status := res.Status
			switch {
			case status >= 500:
				logEvent = config.Logger.Error()
			case status >= 400:
				logEvent = config.Logger.Warn()
			default:
				logEvent = config.Logger.Info()
			}

			// Add error if present
			if err != nil {
				logEvent = logEvent.Err(err)
			}

			// Log the request
			logEvent = logEvent.
				Str("protocol", "http").
				Str("method", req.Method).
				Str("uri", req.RequestURI).
				Str("path", c.Path()).
				Int("status", status).
				Int64("size", res.Size).
				Dur("duration", duration).
				Str("ip", c.RealIP()).
				Str("user_agent", req.UserAgent()).
				Str("request_id", res.Header().Get(echo.HeaderXRequestID))

			// Include headers if configured
			if config.IncludeHeaders {
				headers := make(map[string]string)
				for k, v := range req.Header {
					if len(v) > 0 {
						// Skip sensitive headers
						if k == "Authorization" || k == "Cookie" {
							headers[k] = "[REDACTED]"
						} else {
							headers[k] = v[0]
						}
					}
				}
				logEvent = logEvent.Interface("headers", headers)
			}

			logEvent.Msg("HTTP request completed")

			return err
		}
	}
}

