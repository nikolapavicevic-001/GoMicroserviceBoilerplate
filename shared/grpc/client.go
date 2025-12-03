package grpc

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	grpcMiddleware "github.com/yourorg/boilerplate/shared/middleware/grpc"
)

// ClientConfig holds configuration for gRPC client connections
type ClientConfig struct {
	// MaxRecvMsgSize is the maximum message size in bytes the client can receive
	MaxRecvMsgSize int
	// MaxSendMsgSize is the maximum message size in bytes the client can send
	MaxSendMsgSize int
	// EnableTracing enables distributed tracing
	EnableTracing bool
}

// DefaultClientConfig returns a ClientConfig with sensible defaults
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		MaxRecvMsgSize: 10 * 1024 * 1024, // 10MB
		MaxSendMsgSize: 10 * 1024 * 1024, // 10MB
		EnableTracing:  true,
	}
}

// NewClientConn creates a new gRPC client connection with default configuration
func NewClientConn(addr string) (*grpc.ClientConn, error) {
	return NewClientConnWithConfig(addr, DefaultClientConfig())
}

// NewClientConnWithConfig creates a new gRPC client connection with custom configuration
func NewClientConnWithConfig(addr string, cfg ClientConfig) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(cfg.MaxRecvMsgSize),
			grpc.MaxCallSendMsgSize(cfg.MaxSendMsgSize),
		),
	}

	// Add tracing interceptor if enabled
	if cfg.EnableTracing {
		opts = append(opts, grpc.WithUnaryInterceptor(grpcMiddleware.ClientTracingInterceptor()))
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", addr, err)
	}

	return conn, nil
}

