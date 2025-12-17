package input

import (
	"context"
	"errors"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrSessionSaveFailed = errors.New("failed to save session")
)

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string
	Password string
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// AuthUseCase defines the authentication use cases
type AuthUseCase interface {
	// Login authenticates a user with email and password
	Login(ctx context.Context, req LoginRequest) (*UserResponse, error)
	
	// Register creates a new user account
	Register(ctx context.Context, req RegisterRequest) (*UserResponse, error)
	
	// Logout logs out the current user
	Logout(ctx context.Context, sessionID string) error
	
	// GetUser retrieves user information from session
	GetUser(ctx context.Context, sessionID string) (*UserResponse, error)
	
	// ValidateSession checks if a session is valid
	ValidateSession(ctx context.Context, sessionID string) (bool, error)
}

// UserResponse represents a user response
type UserResponse struct {
	ID        string
	Email     string
	Username  string
	Name      string
}

