package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/microserviceboilerplate/device/internal/domain"
)

// NATSPublisher implements EventPublisher using NATS JetStream
type NATSPublisher struct {
	js nats.JetStreamContext
}

// NewNATSPublisher creates a new NATS event publisher
func NewNATSPublisher(js nats.JetStreamContext) domain.EventPublisher {
	return &NATSPublisher{js: js}
}

// PublishDeviceCreated publishes a device.created event
func (p *NATSPublisher) PublishDeviceCreated(ctx context.Context, device *domain.Device) error {
	event := map[string]interface{}{
		"event_type": "device.created",
		"device_id":  device.ID,
		"name":       device.Name,
		"type":       device.Type,
		"status":     device.Status,
		"timestamp":  time.Now().Unix(),
	}

	return p.publishEvent(ctx, "device.created", event)
}

// PublishDeviceUpdated publishes a device.updated event
func (p *NATSPublisher) PublishDeviceUpdated(ctx context.Context, device *domain.Device) error {
	event := map[string]interface{}{
		"event_type": "device.updated",
		"device_id":  device.ID,
		"name":       device.Name,
		"type":       device.Type,
		"status":     device.Status,
		"timestamp":  time.Now().Unix(),
	}

	return p.publishEvent(ctx, "device.updated", event)
}

// PublishDeviceDeleted publishes a device.deleted event
func (p *NATSPublisher) PublishDeviceDeleted(ctx context.Context, deviceID string) error {
	event := map[string]interface{}{
		"event_type": "device.deleted",
		"device_id":  deviceID,
		"timestamp":  time.Now().Unix(),
	}

	return p.publishEvent(ctx, "device.deleted", event)
}

// publishEvent is a helper to publish events to NATS
func (p *NATSPublisher) publishEvent(ctx context.Context, subject string, event map[string]interface{}) error {
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	_, err = p.js.Publish(fmt.Sprintf("events.%s", subject), eventData)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

