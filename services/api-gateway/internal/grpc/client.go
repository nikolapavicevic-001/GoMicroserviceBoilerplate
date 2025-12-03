package grpc

import (
	pb "github.com/yourorg/boilerplate/shared/proto/gen/user/v1"
	sharedGrpc "github.com/yourorg/boilerplate/shared/grpc"
	"google.golang.org/grpc"
)

// NewUserServiceClient creates a new gRPC client for user service
func NewUserServiceClient(addr string) (pb.UserServiceClient, *grpc.ClientConn, error) {
	conn, err := sharedGrpc.NewClientConn(addr)
	if err != nil {
		return nil, nil, err
	}
	return pb.NewUserServiceClient(conn), conn, nil
}
