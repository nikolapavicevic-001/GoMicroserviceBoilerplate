package grpc

import (
	"context"
	"runtime/debug"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryInterceptor returns a gRPC unary server interceptor that recovers from panics
func RecoveryInterceptor(logger *zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic
				logger.Error().
					Interface("panic", r).
					Str("method", info.FullMethod).
					Str("stack", string(debug.Stack())).
					Msg("gRPC handler panic recovered")

				// Return internal error
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// StreamRecoveryInterceptor returns a gRPC stream server interceptor that recovers from panics
func StreamRecoveryInterceptor(logger *zerolog.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic
				logger.Error().
					Interface("panic", r).
					Str("method", info.FullMethod).
					Str("stack", string(debug.Stack())).
					Msg("gRPC stream handler panic recovered")

				// Return internal error
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(srv, ss)
	}
}

