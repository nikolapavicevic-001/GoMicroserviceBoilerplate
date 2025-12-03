package repository

import (
	"context"

	"github.com/yourorg/boilerplate/services/user-service/internal/domain"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int, search string) ([]*domain.User, int64, error)
	EnsureIndexes(ctx context.Context) error
}
