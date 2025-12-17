package domain

import (
	"context"
)

// DeviceRepository defines the interface for device persistence
type DeviceRepository interface {
	// GetByID retrieves a device by its ID
	GetByID(ctx context.Context, id string) (*Device, error)
	
	// Create creates a new device
	Create(ctx context.Context, device *Device) (*Device, error)
	
	// Update updates an existing device
	Update(ctx context.Context, device *Device) (*Device, error)
	
	// Delete removes a device by ID
	Delete(ctx context.Context, id string) error
	
	// List retrieves devices with pagination
	List(ctx context.Context, limit, offset int32) ([]*Device, error)
	
	// Count returns the total number of devices
	Count(ctx context.Context) (int32, error)
}

