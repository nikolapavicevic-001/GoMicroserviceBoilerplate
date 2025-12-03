package http

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	httpRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
		[]string{"method", "path"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8), // 100B to 10GB
		},
		[]string{"method", "path"},
	)
)

// MetricsMiddleware returns an Echo middleware that collects Prometheus metrics
func MetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			path := c.Path()
			if path == "" {
				path = req.URL.Path
			}
			method := req.Method

			// Increment in-flight requests
			httpRequestsInFlight.WithLabelValues(method, path).Inc()
			defer httpRequestsInFlight.WithLabelValues(method, path).Dec()

			start := time.Now()

			// Process request
			err := next(c)

			// Calculate duration
			duration := time.Since(start).Seconds()

			// Get status code
			status := strconv.Itoa(c.Response().Status)

			// Record metrics
			httpRequestsTotal.WithLabelValues(method, path, status).Inc()
			httpRequestDuration.WithLabelValues(method, path).Observe(duration)
			httpResponseSize.WithLabelValues(method, path).Observe(float64(c.Response().Size))

			return err
		}
	}
}

// MetricsConfig holds the configuration for metrics middleware
type MetricsConfig struct {
	// SkipPaths is a list of paths to skip metrics collection
	SkipPaths []string

	// Subsystem is the Prometheus metrics subsystem
	Subsystem string
}

// MetricsMiddlewareWithConfig returns a metrics middleware with custom configuration
func MetricsMiddlewareWithConfig(config MetricsConfig) echo.MiddlewareFunc {
	skipMap := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipMap[path] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			path := c.Path()
			if path == "" {
				path = req.URL.Path
			}

			// Skip metrics for certain paths
			if skipMap[path] {
				return next(c)
			}

			method := req.Method

			// Increment in-flight requests
			httpRequestsInFlight.WithLabelValues(method, path).Inc()
			defer httpRequestsInFlight.WithLabelValues(method, path).Dec()

			start := time.Now()

			// Process request
			err := next(c)

			// Calculate duration
			duration := time.Since(start).Seconds()

			// Get status code
			status := strconv.Itoa(c.Response().Status)

			// Record metrics
			httpRequestsTotal.WithLabelValues(method, path, status).Inc()
			httpRequestDuration.WithLabelValues(method, path).Observe(duration)
			httpResponseSize.WithLabelValues(method, path).Observe(float64(c.Response().Size))

			return err
		}
	}
}

// MetricsHandler returns an Echo handler for the /metrics endpoint
func MetricsHandler() echo.HandlerFunc {
	h := promhttp.Handler()
	return func(c echo.Context) error {
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

// RegisterMetricsEndpoint registers the /metrics endpoint on the Echo instance
func RegisterMetricsEndpoint(e *echo.Echo) {
	e.GET("/metrics", MetricsHandler())
}

