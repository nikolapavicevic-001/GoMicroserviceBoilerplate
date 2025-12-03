package fixtures

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TestEvent represents an event fixture for testing
type TestEvent struct {
	ID          string
	Type        string
	AggregateID string
	Timestamp   time.Time
	Version     int
	Data        []byte
}

// UserCreatedEventData represents the data for a user created event
type UserCreatedEventData struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// UserUpdatedEventData represents the data for a user updated event
type UserUpdatedEventData struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserDeletedEventData represents the data for a user deleted event
type UserDeletedEventData struct {
	UserID    string    `json:"user_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// UserCreatedEvent returns a user created event fixture
func UserCreatedEvent() TestEvent {
	now := time.Now().UTC()
	data := UserCreatedEventData{
		UserID:    "user-001",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: now,
	}

	dataBytes, _ := json.Marshal(data)

	return TestEvent{
		ID:          uuid.New().String(),
		Type:        "user.created",
		AggregateID: "user-001",
		Timestamp:   now,
		Version:     1,
		Data:        dataBytes,
	}
}

// UserUpdatedEvent returns a user updated event fixture
func UserUpdatedEvent() TestEvent {
	now := time.Now().UTC()
	data := UserUpdatedEventData{
		UserID:    "user-001",
		Email:     "updated@example.com",
		Name:      "Updated User",
		UpdatedAt: now,
	}

	dataBytes, _ := json.Marshal(data)

	return TestEvent{
		ID:          uuid.New().String(),
		Type:        "user.updated",
		AggregateID: "user-001",
		Timestamp:   now,
		Version:     1,
		Data:        dataBytes,
	}
}

// UserDeletedEvent returns a user deleted event fixture
func UserDeletedEvent() TestEvent {
	now := time.Now().UTC()
	data := UserDeletedEventData{
		UserID:    "user-001",
		DeletedAt: now,
	}

	dataBytes, _ := json.Marshal(data)

	return TestEvent{
		ID:          uuid.New().String(),
		Type:        "user.deleted",
		AggregateID: "user-001",
		Timestamp:   now,
		Version:     1,
		Data:        dataBytes,
	}
}

// Events returns a slice of various test events
func Events() []TestEvent {
	return []TestEvent{
		UserCreatedEvent(),
		UserUpdatedEvent(),
		UserDeletedEvent(),
	}
}

// InvalidEvents returns events that should fail validation
func InvalidEvents() []struct {
	Name  string
	Event TestEvent
	Error string
} {
	return []struct {
		Name  string
		Event TestEvent
		Error string
	}{
		{
			Name: "empty event type",
			Event: TestEvent{
				ID:          uuid.New().String(),
				Type:        "",
				AggregateID: "user-001",
				Timestamp:   time.Now(),
				Version:     1,
				Data:        []byte("{}"),
			},
			Error: "event type is required",
		},
		{
			Name: "empty aggregate ID",
			Event: TestEvent{
				ID:          uuid.New().String(),
				Type:        "user.created",
				AggregateID: "",
				Timestamp:   time.Now(),
				Version:     1,
				Data:        []byte("{}"),
			},
			Error: "aggregate ID is required",
		},
		{
			Name: "invalid version",
			Event: TestEvent{
				ID:          uuid.New().String(),
				Type:        "user.created",
				AggregateID: "user-001",
				Timestamp:   time.Now(),
				Version:     0,
				Data:        []byte("{}"),
			},
			Error: "version must be positive",
		},
		{
			Name: "empty data",
			Event: TestEvent{
				ID:          uuid.New().String(),
				Type:        "user.created",
				AggregateID: "user-001",
				Timestamp:   time.Now(),
				Version:     1,
				Data:        nil,
			},
			Error: "event data is required",
		},
	}
}

