package fixtures

import (
	"time"
)

// TestUser represents a user fixture for testing
type TestUser struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	AvatarURL    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Users returns a slice of test users
func Users() []TestUser {
	now := time.Now().UTC()
	return []TestUser{
		{
			ID:           "user-001",
			Email:        "john.doe@example.com",
			Name:         "John Doe",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // password: "password123"
			AvatarURL:    "https://example.com/avatars/john.png",
			CreatedAt:    now.Add(-30 * 24 * time.Hour),
			UpdatedAt:    now.Add(-7 * 24 * time.Hour),
		},
		{
			ID:           "user-002",
			Email:        "jane.smith@example.com",
			Name:         "Jane Smith",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // password: "password123"
			AvatarURL:    "https://example.com/avatars/jane.png",
			CreatedAt:    now.Add(-14 * 24 * time.Hour),
			UpdatedAt:    now.Add(-3 * 24 * time.Hour),
		},
		{
			ID:           "user-003",
			Email:        "bob.wilson@example.com",
			Name:         "Bob Wilson",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // password: "password123"
			AvatarURL:    "",
			CreatedAt:    now.Add(-7 * 24 * time.Hour),
			UpdatedAt:    now.Add(-24 * time.Hour),
		},
	}
}

// AdminUser returns an admin user fixture
func AdminUser() TestUser {
	now := time.Now().UTC()
	return TestUser{
		ID:           "admin-001",
		Email:        "admin@example.com",
		Name:         "System Admin",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // password: "password123"
		AvatarURL:    "https://example.com/avatars/admin.png",
		CreatedAt:    now.Add(-365 * 24 * time.Hour),
		UpdatedAt:    now.Add(-24 * time.Hour),
	}
}

// NewUserFixture returns a fixture for creating a new user
func NewUserFixture() TestUser {
	return TestUser{
		Email: "new.user@example.com",
		Name:  "New User",
	}
}

// InvalidUserFixtures returns a slice of invalid user fixtures for testing validation
func InvalidUserFixtures() []struct {
	Name  string
	User  TestUser
	Error string
} {
	return []struct {
		Name  string
		User  TestUser
		Error string
	}{
		{
			Name: "empty email",
			User: TestUser{
				Email: "",
				Name:  "Test User",
			},
			Error: "email is required",
		},
		{
			Name: "invalid email format",
			User: TestUser{
				Email: "not-an-email",
				Name:  "Test User",
			},
			Error: "invalid email format",
		},
		{
			Name: "empty name",
			User: TestUser{
				Email: "test@example.com",
				Name:  "",
			},
			Error: "name is required",
		},
		{
			Name: "name too long",
			User: TestUser{
				Email: "test@example.com",
				Name:  "This is a very long name that exceeds the maximum allowed length for a user name field in the system which should be limited to a reasonable size",
			},
			Error: "name too long",
		},
	}
}

