package output

import (
	"context"
	"github.com/microserviceboilerplate/web/domain/entity"
)

// SessionRepository defines the session repository interface
type SessionRepository interface {
	// Save saves user session data
	Save(ctx context.Context, sessionID string, user *entity.User) error
	
	// Get retrieves user session data
	Get(ctx context.Context, sessionID string) (*entity.User, error)
	
	// Delete removes a session
	Delete(ctx context.Context, sessionID string) error
	
	// Exists checks if a session exists
	Exists(ctx context.Context, sessionID string) (bool, error)
}

