package output

import (
	"context"
	"github.com/microserviceboilerplate/web/domain/entity"
)

// AuthRepository defines the authentication repository interface
type AuthRepository interface {
	// Authenticate authenticates a user with email and password
	Authenticate(ctx context.Context, email, password string) (*entity.User, error)
	
	// CreateUser creates a new user
	CreateUser(ctx context.Context, email, password, firstName, lastName string) error
	
	// GetUserInfo retrieves user information using an access token
	GetUserInfo(ctx context.Context, accessToken string) (*entity.User, error)
	
	// GetAdminToken retrieves an admin token for Keycloak operations
	GetAdminToken(ctx context.Context) (string, error)
}

