package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// LoggerMiddleware creates a request logging middleware
func LoggerMiddleware(logger *zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			req := c.Request()
			res := c.Response()

			err := next(c)

			duration := time.Since(start)

			logger.Info().
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Str("remote_ip", c.RealIP()).
				Int("status", res.Status).
				Dur("duration_ms", duration).
				Str("user_agent", req.UserAgent()).
				Msg("http request")

			return err
		}
	}
}
