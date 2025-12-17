package nats

import (
	"context"
	"fmt"
	"time"
)

// Device represents a device from the device service
type Device struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// ListDevicesResponse represents the response from listing devices
type ListDevicesResponse struct {
	Devices  []*Device `json:"devices"`
	Total    int32     `json:"total"`
	Page     int32     `json:"page"`
	PageSize int32     `json:"page_size"`
}

// DeviceClient handles device service communication via NATS
type DeviceClient struct {
	client *Client
}

// NewDeviceClient creates a new device client
func NewDeviceClient(client *Client) *DeviceClient {
	return &DeviceClient{client: client}
}

// ListDevices requests a list of devices
func (c *DeviceClient) ListDevices(ctx context.Context, page, pageSize int32) (*ListDevicesResponse, error) {
	req := map[string]interface{}{
		"page":      page,
		"page_size": pageSize,
	}

	var resp ListDevicesResponse
	if err := c.client.RequestJSON(ctx, "request.device.list", req, &resp, 10*time.Second); err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	return &resp, nil
}

// GetDevice requests a single device by ID
func (c *DeviceClient) GetDevice(ctx context.Context, id string) (*Device, error) {
	req := map[string]interface{}{
		"id": id,
	}

	var device Device
	if err := c.client.RequestJSON(ctx, "request.device.get", req, &device, 10*time.Second); err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	return &device, nil
}

