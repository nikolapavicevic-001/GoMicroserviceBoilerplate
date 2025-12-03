package events

import "time"

const (
	// UserCreated event type
	UserCreated = "user.created"
	// UserUpdated event type
	UserUpdated = "user.updated"
	// UserDeleted event type
	UserDeleted = "user.deleted"
)

// UserCreatedData represents the data for a user created event
type UserCreatedData struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// NewUserCreatedEvent creates a new user created event
func NewUserCreatedEvent(userID, email, name string) (*Event, error) {
	data := UserCreatedData{
		UserID:    userID,
		Email:     email,
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
	return NewEvent(UserCreated, userID, data)
}

// UserUpdatedData represents the data for a user updated event
type UserUpdatedData struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewUserUpdatedEvent creates a new user updated event
func NewUserUpdatedEvent(userID, email, name string) (*Event, error) {
	data := UserUpdatedData{
		UserID:    userID,
		Email:     email,
		Name:      name,
		UpdatedAt: time.Now().UTC(),
	}
	return NewEvent(UserUpdated, userID, data)
}

// UserDeletedData represents the data for a user deleted event
type UserDeletedData struct {
	UserID    string    `json:"user_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// NewUserDeletedEvent creates a new user deleted event
func NewUserDeletedEvent(userID string) (*Event, error) {
	data := UserDeletedData{
		UserID:    userID,
		DeletedAt: time.Now().UTC(),
	}
	return NewEvent(UserDeleted, userID, data)
}
