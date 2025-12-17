package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	
	"github.com/microserviceboilerplate/device/internal/domain"
	"github.com/microserviceboilerplate/device/internal/service"
	pb "github.com/microserviceboilerplate/device/proto/device"
)

// DeviceHandler implements the gRPC DeviceService
type DeviceHandler struct {
	pb.UnimplementedDeviceServiceServer
	service *service.DeviceService
}

// NewDeviceHandler creates a new gRPC device handler
func NewDeviceHandler(svc *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		service: svc,
	}
}

// GetDevice retrieves a device by ID
func (h *DeviceHandler) GetDevice(ctx context.Context, req *pb.GetDeviceRequest) (*pb.Device, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "device ID is required")
	}

	device, err := h.service.GetDevice(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return toProtoDevice(device), nil
}

// CreateDevice creates a new device
func (h *DeviceHandler) CreateDevice(ctx context.Context, req *pb.CreateDeviceRequest) (*pb.Device, error) {
	device, err := h.service.CreateDevice(ctx, req.Name, req.Type, req.Status)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return toProtoDevice(device), nil
}

// UpdateDevice updates an existing device
func (h *DeviceHandler) UpdateDevice(ctx context.Context, req *pb.UpdateDeviceRequest) (*pb.Device, error) {
	device, err := h.service.UpdateDevice(ctx, req.Id, req.Name, req.Type, req.Status)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return toProtoDevice(device), nil
}

// DeleteDevice removes a device
func (h *DeviceHandler) DeleteDevice(ctx context.Context, req *pb.DeleteDeviceRequest) (*pb.DeleteDeviceResponse, error) {
	err := h.service.DeleteDevice(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &pb.DeleteDeviceResponse{Success: true}, nil
}

// ListDevices retrieves devices with pagination
func (h *DeviceHandler) ListDevices(ctx context.Context, req *pb.ListDevicesRequest) (*pb.ListDevicesResponse, error) {
	devices, total, err := h.service.ListDevices(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoDevices := make([]*pb.Device, len(devices))
	for i, d := range devices {
		protoDevices[i] = toProtoDevice(d)
	}

	return &pb.ListDevicesResponse{
		Devices:  protoDevices,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// toProtoDevice converts domain device to protobuf message
func toProtoDevice(d *domain.Device) *pb.Device {
	return &pb.Device{
		Id:        d.ID,
		Name:      d.Name,
		Type:      d.Type,
		Status:    d.Status,
		CreatedAt: d.CreatedAt.Unix(),
		UpdatedAt: d.UpdatedAt.Unix(),
	}
}

