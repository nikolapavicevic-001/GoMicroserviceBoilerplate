package entity

// User represents a user entity in the domain
type User struct {
	ID               string
	Email            string
	Username        string
	Name            string
	IsAuthenticated bool
	RefreshToken    string
}

// NewUser creates a new user entity
func NewUser(id, email, username, name string) *User {
	return &User{
		ID:        id,
		Email:     email,
		Username:  username,
		Name:      name,
	}
}

