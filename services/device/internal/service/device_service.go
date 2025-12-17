package service

import (
	"context"
	"fmt"

	"github.com/microserviceboilerplate/device/internal/domain"
)

// DeviceService implements business logic for device operations
type DeviceService struct {
	repo    domain.DeviceRepository
	events  domain.EventPublisher
}

// NewDeviceService creates a new device service
func NewDeviceService(repo domain.DeviceRepository, events domain.EventPublisher) *DeviceService {
	return &DeviceService{
		repo:   repo,
		events: events,
	}
}

// GetDevice retrieves a device by ID
func (s *DeviceService) GetDevice(ctx context.Context, id string) (*domain.Device, error) {
	if id == "" {
		return nil, fmt.Errorf("device ID is required")
	}

	device, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	return device, nil
}

// CreateDevice creates a new device
func (s *DeviceService) CreateDevice(ctx context.Context, name, deviceType, status string) (*domain.Device, error) {
	// Validation
	if name == "" {
		return nil, fmt.Errorf("device name is required")
	}
	if deviceType == "" {
		return nil, fmt.Errorf("device type is required")
	}
	if status == "" {
		status = domain.DeviceStatusActive
	}
	if !domain.IsValidStatus(status) {
		return nil, fmt.Errorf("invalid device status: %s", status)
	}

	device := &domain.Device{
		Name:   name,
		Type:   deviceType,
		Status: status,
	}

	created, err := s.repo.Create(ctx, device)
	if err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	// Publish event
	if err := s.events.PublishDeviceCreated(ctx, created); err != nil {
		// Log error but don't fail the operation
		// In production, consider using an outbox pattern
	}

	return created, nil
}

// UpdateDevice updates an existing device
func (s *DeviceService) UpdateDevice(ctx context.Context, id, name, deviceType, status string) (*domain.Device, error) {
	if id == "" {
		return nil, fmt.Errorf("device ID is required")
	}

	// Get existing device
	device, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	// Update fields if provided
	if name != "" {
		device.Name = name
	}
	if deviceType != "" {
		device.Type = deviceType
	}
	if status != "" {
		if !domain.IsValidStatus(status) {
			return nil, fmt.Errorf("invalid device status: %s", status)
		}
		device.Status = status
	}

	updated, err := s.repo.Update(ctx, device)
	if err != nil {
		return nil, fmt.Errorf("failed to update device: %w", err)
	}

	// Publish event
	if err := s.events.PublishDeviceUpdated(ctx, updated); err != nil {
		// Log error but don't fail the operation
	}

	return updated, nil
}

// DeleteDevice removes a device
func (s *DeviceService) DeleteDevice(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("device ID is required")
	}

	// Verify device exists
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	// Publish event
	if err := s.events.PublishDeviceDeleted(ctx, id); err != nil {
		// Log error but don't fail the operation
	}

	return nil
}

// ListDevices retrieves devices with pagination
func (s *DeviceService) ListDevices(ctx context.Context, page, pageSize int32) ([]*domain.Device, int32, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	devices, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list devices: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count devices: %w", err)
	}

	return devices, total, nil
}

