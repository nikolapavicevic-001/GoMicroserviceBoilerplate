package domain

import "context"

// EventPublisher defines the interface for publishing domain events
type EventPublisher interface {
	// PublishDeviceCreated publishes a device.created event
	PublishDeviceCreated(ctx context.Context, device *Device) error
	
	// PublishDeviceUpdated publishes a device.updated event
	PublishDeviceUpdated(ctx context.Context, device *Device) error
	
	// PublishDeviceDeleted publishes a device.deleted event
	PublishDeviceDeleted(ctx context.Context, deviceID string) error
}

