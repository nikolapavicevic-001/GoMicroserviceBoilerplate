package grpc

import (
	"google.golang.org/grpc"
	pb "github.com/microserviceboilerplate/device/proto/device"
)

// RegisterDeviceService registers the DeviceService with the gRPC server
func RegisterDeviceService(server *grpc.Server, handler *DeviceHandler) {
	pb.RegisterDeviceServiceServer(server, handler)
}

