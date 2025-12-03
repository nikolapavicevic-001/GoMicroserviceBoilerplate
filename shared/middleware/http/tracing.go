package http

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const httpTracerName = "http-server"

// TracingMiddleware returns an Echo middleware that adds OpenTelemetry tracing
func TracingMiddleware() echo.MiddlewareFunc {
	tracer := otel.Tracer(httpTracerName)
	propagator := otel.GetTextMapPropagator()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			ctx := req.Context()

			// Extract trace context from incoming request headers
			ctx = propagator.Extract(ctx, propagation.HeaderCarrier(req.Header))

			// Start a new span
			spanName := c.Path()
			if spanName == "" {
				spanName = req.URL.Path
			}

			ctx, span := tracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", req.Method),
					attribute.String("http.url", req.URL.String()),
					attribute.String("http.host", req.Host),
					attribute.String("http.user_agent", req.UserAgent()),
					attribute.String("http.scheme", c.Scheme()),
					attribute.String("net.peer.ip", c.RealIP()),
				),
			)
			defer span.End()

			// Update request context
			c.SetRequest(req.WithContext(ctx))

			// Process request
			err := next(c)

			// Record response status
			status := c.Response().Status
			span.SetAttributes(attribute.Int("http.status_code", status))

			// Set span status based on HTTP status code
			if status >= 400 {
				if err != nil {
					span.RecordError(err)
				}
				span.SetStatus(codes.Error, "")
			} else {
				span.SetStatus(codes.Ok, "")
			}

			return err
		}
	}
}

// TracingMiddlewareWithConfig returns a tracing middleware with custom configuration
type TracingConfig struct {
	// TracerName is the name of the tracer
	TracerName string

	// SkipPaths is a list of paths to skip tracing
	SkipPaths []string

	// SpanNameFormatter is a function to format span names
	SpanNameFormatter func(c echo.Context) string
}

// TracingMiddlewareWithConfig returns a tracing middleware with custom configuration
func TracingMiddlewareWithConfig(config TracingConfig) echo.MiddlewareFunc {
	tracerName := config.TracerName
	if tracerName == "" {
		tracerName = httpTracerName
	}

	tracer := otel.Tracer(tracerName)
	propagator := otel.GetTextMapPropagator()

	skipMap := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipMap[path] = true
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip tracing for certain paths
			if skipMap[c.Path()] {
				return next(c)
			}

			req := c.Request()
			ctx := req.Context()

			// Extract trace context from incoming request headers
			ctx = propagator.Extract(ctx, propagation.HeaderCarrier(req.Header))

			// Determine span name
			var spanName string
			if config.SpanNameFormatter != nil {
				spanName = config.SpanNameFormatter(c)
			} else {
				spanName = c.Path()
				if spanName == "" {
					spanName = req.URL.Path
				}
			}

			ctx, span := tracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", req.Method),
					attribute.String("http.url", req.URL.String()),
					attribute.String("http.host", req.Host),
					attribute.String("http.user_agent", req.UserAgent()),
					attribute.String("http.scheme", c.Scheme()),
					attribute.String("net.peer.ip", c.RealIP()),
				),
			)
			defer span.End()

			// Update request context
			c.SetRequest(req.WithContext(ctx))

			// Process request
			err := next(c)

			// Record response status
			status := c.Response().Status
			span.SetAttributes(attribute.Int("http.status_code", status))

			// Set span status based on HTTP status code
			if status >= 400 {
				if err != nil {
					span.RecordError(err)
				}
				span.SetStatus(codes.Error, "")
			} else {
				span.SetStatus(codes.Ok, "")
			}

			return err
		}
	}
}

// InjectTraceContext injects the trace context into outgoing HTTP headers
func InjectTraceContext(c echo.Context) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(c.Request().Context(), propagation.HeaderCarrier(c.Request().Header))
}

// GetTraceID returns the trace ID from the context
func GetTraceID(c echo.Context) string {
	span := trace.SpanFromContext(c.Request().Context())
	if span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpanID returns the span ID from the context
func GetSpanID(c echo.Context) string {
	span := trace.SpanFromContext(c.Request().Context())
	if span.SpanContext().IsValid() {
		return span.SpanContext().SpanID().String()
	}
	return ""
}

