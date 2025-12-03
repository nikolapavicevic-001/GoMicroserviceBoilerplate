package events

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewEvent(t *testing.T) {
	data := map[string]string{"key": "value"}

	event, err := NewEvent("test.event", "agg-123", data)
	if err != nil {
		t.Fatalf("NewEvent() error = %v", err)
	}

	if event.ID == "" {
		t.Error("NewEvent() ID should not be empty")
	}
	if event.Type != "test.event" {
		t.Errorf("NewEvent() Type = %v, want test.event", event.Type)
	}
	if event.AggregateID != "agg-123" {
		t.Errorf("NewEvent() AggregateID = %v, want agg-123", event.AggregateID)
	}
	if event.Version != 1 {
		t.Errorf("NewEvent() Version = %v, want 1", event.Version)
	}
	if event.Timestamp.IsZero() {
		t.Error("NewEvent() Timestamp should not be zero")
	}
	if len(event.Data) == 0 {
		t.Error("NewEvent() Data should not be empty")
	}

	// Verify data can be unmarshaled
	var decoded map[string]string
	if err := json.Unmarshal(event.Data, &decoded); err != nil {
		t.Errorf("NewEvent() Data unmarshaling error = %v", err)
	}
	if decoded["key"] != "value" {
		t.Errorf("NewEvent() Data[key] = %v, want value", decoded["key"])
	}
}

func TestNewEvent_ComplexData(t *testing.T) {
	type ComplexData struct {
		Name    string    `json:"name"`
		Count   int       `json:"count"`
		Active  bool      `json:"active"`
		Created time.Time `json:"created"`
	}

	data := ComplexData{
		Name:    "test",
		Count:   42,
		Active:  true,
		Created: time.Now().UTC(),
	}

	event, err := NewEvent("complex.event", "agg-456", data)
	if err != nil {
		t.Fatalf("NewEvent() error = %v", err)
	}

	var decoded ComplexData
	if err := json.Unmarshal(event.Data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal data: %v", err)
	}

	if decoded.Name != data.Name {
		t.Errorf("Decoded Name = %v, want %v", decoded.Name, data.Name)
	}
	if decoded.Count != data.Count {
		t.Errorf("Decoded Count = %v, want %v", decoded.Count, data.Count)
	}
	if decoded.Active != data.Active {
		t.Errorf("Decoded Active = %v, want %v", decoded.Active, data.Active)
	}
}

func TestNewEvent_NilData(t *testing.T) {
	event, err := NewEvent("nil.event", "agg-789", nil)
	if err != nil {
		t.Fatalf("NewEvent() error = %v", err)
	}

	if string(event.Data) != "null" {
		t.Errorf("NewEvent() Data = %v, want null", string(event.Data))
	}
}

func TestEventUniqueness(t *testing.T) {
	event1, _ := NewEvent("test", "agg", nil)
	event2, _ := NewEvent("test", "agg", nil)

	if event1.ID == event2.ID {
		t.Error("Two events should have different IDs")
	}
}

