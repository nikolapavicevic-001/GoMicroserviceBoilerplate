package http

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

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
				Dur("duration", duration)

			return err
		}
	}
}
