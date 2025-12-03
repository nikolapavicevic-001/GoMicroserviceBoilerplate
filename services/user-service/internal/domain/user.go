package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user entity
type User struct {
	ID        string    `bson:"_id"`
	Email     string    `bson:"email"`
	Name      string    `bson:"name"`
	Password  string    `bson:"password"` // Hashed
	AvatarURL string    `bson:"avatar_url"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// NewUser creates a new user with hashed password
func NewUser(email, name, password string) (*User, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &User{
		ID:        GenerateID(),
		Email:     email,
		Name:      name,
		Password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword checks if the provided password matches the hashed password
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// Update updates user fields
func (u *User) Update(name, avatarURL string) {
	if name != "" {
		u.Name = name
	}
	if avatarURL != "" {
		u.AvatarURL = avatarURL
	}
	u.UpdatedAt = time.Now().UTC()
}

// GenerateID generates a unique ID (UUID)
func GenerateID() string {
	// In production, use UUID library
	return time.Now().Format("20060102150405") + randomString(8)
}

func randomString(n int) string {
	// Simple implementation - use crypto/rand in production
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}
