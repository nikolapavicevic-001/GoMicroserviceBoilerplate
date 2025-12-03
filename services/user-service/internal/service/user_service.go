package service

import (
	"context"

	"github.com/yourorg/boilerplate/services/user-service/internal/domain"
	"github.com/yourorg/boilerplate/services/user-service/internal/repository"
	"github.com/yourorg/boilerplate/shared/errors"
	"github.com/yourorg/boilerplate/shared/events"
	"github.com/yourorg/boilerplate/shared/kafka"
	"github.com/yourorg/boilerplate/shared/logger"
)

// UserService handles user business logic
type UserService struct {
	repo     repository.UserRepository
	producer *kafka.Producer
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository, producer *kafka.Producer) *UserService {
	return &UserService{
		repo:     repo,
		producer: producer,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(ctx context.Context, email, name, password string) (*domain.User, error) {
	log := logger.WithContext(ctx)

	// Check if user already exists
	existing, err := s.repo.FindByEmail(ctx, email)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewAlreadyExistsError("user", "email", email)
	}

	// Create new user
	user, err := domain.NewUser(email, name, password)
	if err != nil {
		log.Error().Err(err).Msg("failed to create user domain object")
		return nil, errors.Wrap(err, "failed to create user")
	}

	// Save to database
	if err := s.repo.Create(ctx, user); err != nil {
		log.Error().Err(err).Str("email", email).Msg("failed to save user")
		return nil, err
	}

	log.Info().
		Str("user_id", user.ID).
		Str("email", user.Email).
		Msg("user created successfully")

	// Publish user created event
	event, _ := events.NewUserCreatedEvent(user.ID, user.Email, user.Name)
	if err := s.producer.Publish(ctx, "user.events", event); err != nil {
		log.Error().Err(err).Msg("failed to publish user created event")
		// Don't fail the request - event publishing is best-effort
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, id, name, avatarURL string) (*domain.User, error) {
	log := logger.WithContext(ctx)

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Update(name, avatarURL)

	if err := s.repo.Update(ctx, user); err != nil {
		log.Error().Err(err).Str("user_id", id).Msg("failed to update user")
		return nil, err
	}

	log.Info().
		Str("user_id", user.ID).
		Msg("user updated successfully")

	// Publish user updated event
	event, _ := events.NewUserUpdatedEvent(user.ID, user.Email, user.Name)
	if err := s.producer.Publish(ctx, "user.events", event); err != nil {
		log.Error().Err(err).Msg("failed to publish user updated event")
	}

	return user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	log := logger.WithContext(ctx)

	if err := s.repo.Delete(ctx, id); err != nil {
		log.Error().Err(err).Str("user_id", id).Msg("failed to delete user")
		return err
	}

	log.Info().
		Str("user_id", id).
		Msg("user deleted successfully")

	// Publish user deleted event
	event, _ := events.NewUserDeletedEvent(id)
	if err := s.producer.Publish(ctx, "user.events", event); err != nil {
		log.Error().Err(err).Msg("failed to publish user deleted event")
	}

	return nil
}

// ListUsers lists users with pagination
func (s *UserService) ListUsers(ctx context.Context, limit, offset int, search string) ([]*domain.User, int64, error) {
	return s.repo.List(ctx, limit, offset, search)
}
