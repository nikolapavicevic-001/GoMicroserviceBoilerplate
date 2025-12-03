package grpc

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// Server metrics
	grpcServerRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_server_requests_total",
			Help: "Total number of gRPC requests received",
		},
		[]string{"method", "code"},
	)

	grpcServerRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_server_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	grpcServerRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_server_requests_in_flight",
			Help: "Number of gRPC requests currently being processed",
		},
		[]string{"method"},
	)

	// Client metrics
	grpcClientRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_client_requests_total",
			Help: "Total number of gRPC client requests sent",
		},
		[]string{"method", "code"},
	)

	grpcClientRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_client_request_duration_seconds",
			Help:    "gRPC client request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

// MetricsInterceptor returns a gRPC unary server interceptor that collects Prometheus metrics
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		method := info.FullMethod

		// Increment in-flight requests
		grpcServerRequestsInFlight.WithLabelValues(method).Inc()
		defer grpcServerRequestsInFlight.WithLabelValues(method).Dec()

		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get status code
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			} else {
				code = codes.Unknown
			}
		}

		// Record metrics
		grpcServerRequestsTotal.WithLabelValues(method, code.String()).Inc()
		grpcServerRequestDuration.WithLabelValues(method).Observe(duration)

		return resp, err
	}
}

// StreamMetricsInterceptor returns a gRPC stream server interceptor that collects Prometheus metrics
func StreamMetricsInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		method := info.FullMethod

		// Increment in-flight requests
		grpcServerRequestsInFlight.WithLabelValues(method).Inc()
		defer grpcServerRequestsInFlight.WithLabelValues(method).Dec()

		start := time.Now()

		// Call the handler
		err := handler(srv, ss)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get status code
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			} else {
				code = codes.Unknown
			}
		}

		// Record metrics
		grpcServerRequestsTotal.WithLabelValues(method, code.String()).Inc()
		grpcServerRequestDuration.WithLabelValues(method).Observe(duration)

		return err
	}
}

// ClientMetricsInterceptor returns a gRPC unary client interceptor that collects Prometheus metrics
func ClientMetricsInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		start := time.Now()

		// Call the invoker
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get status code
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			} else {
				code = codes.Unknown
			}
		}

		// Record metrics
		grpcClientRequestsTotal.WithLabelValues(method, code.String()).Inc()
		grpcClientRequestDuration.WithLabelValues(method).Observe(duration)

		return err
	}
}

