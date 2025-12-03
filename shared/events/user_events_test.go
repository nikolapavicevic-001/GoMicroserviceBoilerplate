package events

import (
	"encoding/json"
	"testing"
)

func TestNewUserCreatedEvent(t *testing.T) {
	event, err := NewUserCreatedEvent("user-123", "test@example.com", "Test User")
	if err != nil {
		t.Fatalf("NewUserCreatedEvent() error = %v", err)
	}

	if event.Type != UserCreated {
		t.Errorf("NewUserCreatedEvent() Type = %v, want %v", event.Type, UserCreated)
	}
	if event.AggregateID != "user-123" {
		t.Errorf("NewUserCreatedEvent() AggregateID = %v, want user-123", event.AggregateID)
	}

	var data UserCreatedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		t.Fatalf("Failed to unmarshal event data: %v", err)
	}

	if data.UserID != "user-123" {
		t.Errorf("UserCreatedData.UserID = %v, want user-123", data.UserID)
	}
	if data.Email != "test@example.com" {
		t.Errorf("UserCreatedData.Email = %v, want test@example.com", data.Email)
	}
	if data.Name != "Test User" {
		t.Errorf("UserCreatedData.Name = %v, want Test User", data.Name)
	}
	if data.CreatedAt.IsZero() {
		t.Error("UserCreatedData.CreatedAt should not be zero")
	}
}

func TestNewUserUpdatedEvent(t *testing.T) {
	event, err := NewUserUpdatedEvent("user-123", "updated@example.com", "Updated Name")
	if err != nil {
		t.Fatalf("NewUserUpdatedEvent() error = %v", err)
	}

	if event.Type != UserUpdated {
		t.Errorf("NewUserUpdatedEvent() Type = %v, want %v", event.Type, UserUpdated)
	}
	if event.AggregateID != "user-123" {
		t.Errorf("NewUserUpdatedEvent() AggregateID = %v, want user-123", event.AggregateID)
	}

	var data UserUpdatedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		t.Fatalf("Failed to unmarshal event data: %v", err)
	}

	if data.UserID != "user-123" {
		t.Errorf("UserUpdatedData.UserID = %v, want user-123", data.UserID)
	}
	if data.Email != "updated@example.com" {
		t.Errorf("UserUpdatedData.Email = %v, want updated@example.com", data.Email)
	}
	if data.Name != "Updated Name" {
		t.Errorf("UserUpdatedData.Name = %v, want Updated Name", data.Name)
	}
	if data.UpdatedAt.IsZero() {
		t.Error("UserUpdatedData.UpdatedAt should not be zero")
	}
}

func TestNewUserDeletedEvent(t *testing.T) {
	event, err := NewUserDeletedEvent("user-123")
	if err != nil {
		t.Fatalf("NewUserDeletedEvent() error = %v", err)
	}

	if event.Type != UserDeleted {
		t.Errorf("NewUserDeletedEvent() Type = %v, want %v", event.Type, UserDeleted)
	}
	if event.AggregateID != "user-123" {
		t.Errorf("NewUserDeletedEvent() AggregateID = %v, want user-123", event.AggregateID)
	}

	var data UserDeletedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		t.Fatalf("Failed to unmarshal event data: %v", err)
	}

	if data.UserID != "user-123" {
		t.Errorf("UserDeletedData.UserID = %v, want user-123", data.UserID)
	}
	if data.DeletedAt.IsZero() {
		t.Error("UserDeletedData.DeletedAt should not be zero")
	}
}

func TestEventConstants(t *testing.T) {
	if UserCreated != "user.created" {
		t.Errorf("UserCreated = %v, want user.created", UserCreated)
	}
	if UserUpdated != "user.updated" {
		t.Errorf("UserUpdated = %v, want user.updated", UserUpdated)
	}
	if UserDeleted != "user.deleted" {
		t.Errorf("UserDeleted = %v, want user.deleted", UserDeleted)
	}
}

