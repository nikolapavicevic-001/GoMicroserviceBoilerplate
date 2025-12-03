package session

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yourorg/boilerplate/shared/auth"
)

const (
	CookieName = "auth_token"
)

// Store manages JWT-based authentication via cookies
type Store struct {
	jwtSecret string
	maxAge    time.Duration
	secure    bool
}

// NewStore creates a new JWT cookie store
func NewStore(jwtSecret string, maxAge time.Duration) *Store {
	return &Store{
		jwtSecret: jwtSecret,
		maxAge:    maxAge,
		secure:    false, // Set to true in production with HTTPS
	}
}

// SetUser creates a JWT token and sets it as an httpOnly cookie
func (s *Store) SetUser(c echo.Context, userID, email, name string) error {
	// Generate JWT token
	token, err := auth.GenerateToken(userID, email, s.jwtSecret, s.maxAge)
	if err != nil {
		return err
	}

	// Set cookie
	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(s.maxAge.Seconds()),
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	// Also store name in a separate non-httpOnly cookie for display purposes
	nameCookie := &http.Cookie{
		Name:     "user_name",
		Value:    name,
		Path:     "/",
		MaxAge:   int(s.maxAge.Seconds()),
		HttpOnly: false,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(nameCookie)

	return nil
}

// GetUserID retrieves user ID from JWT cookie
func (s *Store) GetUserID(c echo.Context) (string, bool) {
	cookie, err := c.Cookie(CookieName)
	if err != nil {
		return "", false
	}

	claims, err := auth.ValidateToken(cookie.Value, s.jwtSecret)
	if err != nil {
		return "", false
	}

	return claims.UserID, true
}

// GetUser retrieves all user info from JWT cookie
func (s *Store) GetUser(c echo.Context) (userID, email, name string, ok bool) {
	cookie, err := c.Cookie(CookieName)
	if err != nil {
		return "", "", "", false
	}

	claims, err := auth.ValidateToken(cookie.Value, s.jwtSecret)
	if err != nil {
		return "", "", "", false
	}

	// Get name from the separate cookie
	nameCookie, err := c.Cookie("user_name")
	if err != nil {
		name = "" // Name is optional
	} else {
		name = nameCookie.Value
	}

	return claims.UserID, claims.Email, name, true
}

// Clear clears the authentication cookies
func (s *Store) Clear(c echo.Context) error {
	// Clear auth token
	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	// Clear name cookie
	nameCookie := &http.Cookie{
		Name:     "user_name",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: false,
		Secure:   s.secure,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(nameCookie)

	return nil
}

// IsAuthenticated checks if the user has a valid JWT token
func (s *Store) IsAuthenticated(c echo.Context) bool {
	_, ok := s.GetUserID(c)
	return ok
}
