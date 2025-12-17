package service

import (
	"context"
	"fmt"

	"github.com/microserviceboilerplate/web/domain/port/input"
	"github.com/microserviceboilerplate/web/domain/port/output"
)

// AuthService implements the AuthUseCase interface
type AuthService struct {
	authRepo    output.AuthRepository
	sessionRepo output.SessionRepository
}

// NewAuthService creates a new authentication service
func NewAuthService(
	authRepo output.AuthRepository,
	sessionRepo output.SessionRepository,
) input.AuthUseCase {
	return &AuthService{
		authRepo:    authRepo,
		sessionRepo: sessionRepo,
	}
}

// Login authenticates a user with email and password
func (s *AuthService) Login(ctx context.Context, req input.LoginRequest) (*input.UserResponse, error) {
	// Authenticate with Keycloak
	user, err := s.authRepo.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	// Convert entity to response
	response := &input.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Name:     user.Name,
	}

	return response, nil
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req input.RegisterRequest) (*input.UserResponse, error) {
	// Create user in Keycloak
	err := s.authRepo.CreateUser(ctx, req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Auto-login after registration
	user, err := s.authRepo.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	// Convert entity to response
	response := &input.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Name:     user.Name,
	}

	return response, nil
}

// Logout logs out the current user
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	return s.sessionRepo.Delete(ctx, sessionID)
}

// GetUser retrieves user information from session
func (s *AuthService) GetUser(ctx context.Context, sessionID string) (*input.UserResponse, error) {
	user, err := s.sessionRepo.Get(ctx, sessionID)
	if err != nil {
		return nil, input.ErrUserNotFound
	}

	return &input.UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		Name:     user.Name,
	}, nil
}

// ValidateSession checks if a session is valid
func (s *AuthService) ValidateSession(ctx context.Context, sessionID string) (bool, error) {
	exists, err := s.sessionRepo.Exists(ctx, sessionID)
	if err != nil {
		return false, err
	}
	return exists, nil
}
