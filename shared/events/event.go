package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event is the base structure for all events
type Event struct {
	ID          string    `json:"id"`           // UUID
	Type        string    `json:"type"`         // "user.created"
	AggregateID string    `json:"aggregate_id"` // Entity ID
	Timestamp   time.Time `json:"timestamp"`
	Version     int       `json:"version"` // Schema version
	Data        json.RawMessage `json:"data"`   // JSON payload
}

// NewEvent creates a new event
func NewEvent(eventType, aggregateID string, data interface{}) (*Event, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &Event{
		ID:          uuid.New().String(),
		Type:        eventType,
		AggregateID: aggregateID,
		Timestamp:   time.Now().UTC(),
		Version:     1,
		Data:        dataBytes,
	}, nil
}

// UnmarshalData unmarshals the event data into the provided struct
func (e *Event) UnmarshalData(v interface{}) error {
	return json.Unmarshal(e.Data, v)
}
