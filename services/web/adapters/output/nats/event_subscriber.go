package nats

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/nats-io/nats.go"
)

// DeviceEvent represents a device event
type DeviceEvent struct {
	EventType string `json:"event_type"`
	DeviceID  string `json:"device_id"`
	Name      string `json:"name,omitempty"`
	Type      string `json:"type,omitempty"`
	Status    string `json:"status,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// EventHandler handles incoming events
type EventHandler interface {
	OnDeviceCreated(ctx context.Context, event *DeviceEvent)
	OnDeviceUpdated(ctx context.Context, event *DeviceEvent)
	OnDeviceDeleted(ctx context.Context, event *DeviceEvent)
}

// EventSubscriber subscribes to NATS events
type EventSubscriber struct {
	nc      *nats.Conn
	handler EventHandler
	mu      sync.RWMutex
	subs    []*nats.Subscription
}

// NewEventSubscriber creates a new event subscriber
func NewEventSubscriber(nc *nats.Conn, handler EventHandler) *EventSubscriber {
	return &EventSubscriber{
		nc:      nc,
		handler: handler,
		subs:    make([]*nats.Subscription, 0),
	}
}

// Subscribe subscribes to device events
func (s *EventSubscriber) Subscribe(ctx context.Context) error {
	// Subscribe to device events
	sub, err := s.nc.Subscribe("events.device.*", func(msg *nats.Msg) {
		var event DeviceEvent
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Failed to unmarshal device event: %v", err)
			return
		}

		// Route to appropriate handler
		switch event.EventType {
		case "device.created":
			s.handler.OnDeviceCreated(ctx, &event)
		case "device.updated":
			s.handler.OnDeviceUpdated(ctx, &event)
		case "device.deleted":
			s.handler.OnDeviceDeleted(ctx, &event)
		default:
			log.Printf("Unknown event type: %s", event.EventType)
		}
	})
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.subs = append(s.subs, sub)
	s.mu.Unlock()

	log.Println("Subscribed to device events")
	return nil
}

// Unsubscribe unsubscribes from all events
func (s *EventSubscriber) Unsubscribe() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sub := range s.subs {
		if err := sub.Unsubscribe(); err != nil {
			return err
		}
	}
	s.subs = nil
	return nil
}

