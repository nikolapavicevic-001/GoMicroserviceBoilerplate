package grpc

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/yourorg/boilerplate/shared/proto/gen/user/v1"
	grpcMiddleware "github.com/yourorg/boilerplate/shared/middleware/grpc"
)

// NewUserServiceClient creates a new gRPC client for user service
func NewUserServiceClient(addr string) (pb.UserServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
			grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
		),
		// Add tracing interceptor for distributed tracing
		grpc.WithUnaryInterceptor(grpcMiddleware.ClientTracingInterceptor()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial user service: %w", err)
	}

	client := pb.NewUserServiceClient(conn)
	return client, conn, nil
}
