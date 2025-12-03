package grpc

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor returns a gRPC unary server interceptor that logs requests
func LoggingInterceptor(logger *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			} else {
				code = codes.Unknown
			}
		}

		// Log the request
		logEvent := logger.Info()
		if err != nil {
			logEvent = logger.Error().Err(err)
		}

		logEvent.
			Str("protocol", "grpc").
			Str("method", info.FullMethod).
			Str("code", code.String()).
			Dur("duration", duration).
			Msg("gRPC request completed")

		return resp, err
	}
}

// StreamLoggingInterceptor returns a gRPC stream server interceptor that logs requests
func StreamLoggingInterceptor(logger *zerolog.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()

		// Call the handler
		err := handler(srv, ss)

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			} else {
				code = codes.Unknown
			}
		}

		// Log the request
		logEvent := logger.Info()
		if err != nil {
			logEvent = logger.Error().Err(err)
		}

		logEvent.
			Str("protocol", "grpc-stream").
			Str("method", info.FullMethod).
			Str("code", code.String()).
			Dur("duration", duration).
			Msg("gRPC stream completed")

		return err
	}
}

// ClientLoggingInterceptor returns a gRPC unary client interceptor that logs requests
func ClientLoggingInterceptor(logger *zerolog.Logger) grpc.UnaryClientInterceptor {
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
		duration := time.Since(start)

		// Get status code
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			} else {
				code = codes.Unknown
			}
		}

		// Log the request
		logEvent := logger.Debug()
		if err != nil {
			logEvent = logger.Error().Err(err)
		}

		logEvent.
			Str("protocol", "grpc-client").
			Str("method", method).
			Str("target", cc.Target()).
			Str("code", code.String()).
			Dur("duration", duration).
			Msg("gRPC client call completed")

		return err
	}
}

