package service

import (
	"context"
	"fmt"

	natsclient "github.com/microserviceboilerplate/web/adapters/output/nats"
)

// DeviceService handles device-related operations
type DeviceService struct {
	deviceClient *natsclient.DeviceClient
}

// NewDeviceService creates a new device service
func NewDeviceService(deviceClient *natsclient.DeviceClient) *DeviceService {
	return &DeviceService{
		deviceClient: deviceClient,
	}
}

// ListDevices retrieves a list of devices
func (s *DeviceService) ListDevices(ctx context.Context, page, pageSize int32) (*natsclient.ListDevicesResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	resp, err := s.deviceClient.ListDevices(ctx, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	return resp, nil
}

// GetDevice retrieves a device by ID
func (s *DeviceService) GetDevice(ctx context.Context, id string) (*natsclient.Device, error) {
	if id == "" {
		return nil, fmt.Errorf("device ID is required")
	}

	device, err := s.deviceClient.GetDevice(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	return device, nil
}

