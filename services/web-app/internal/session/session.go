package session

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const (
	SessionName     = "webapp_session"
	SessionKeyUser  = "user_id"
	SessionKeyEmail = "email"
	SessionKeyName  = "name"
)

// Store wraps gorilla sessions store
type Store struct {
	store *sessions.CookieStore
}

// NewStore creates a new session store
func NewStore(secret string, maxAge time.Duration) *Store {
	store := sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	return &Store{
		store: store,
	}
}

// Get retrieves a session
func (s *Store) Get(c echo.Context) (*sessions.Session, error) {
	return s.store.Get(c.Request(), SessionName)
}

// Save saves a session
func (s *Store) Save(c echo.Context, session *sessions.Session) error {
	return session.Save(c.Request(), c.Response())
}

// SetUser sets user information in session
func (s *Store) SetUser(c echo.Context, userID, email, name string) error {
	session, err := s.Get(c)
	if err != nil {
		return err
	}

	session.Values[SessionKeyUser] = userID
	session.Values[SessionKeyEmail] = email
	session.Values[SessionKeyName] = name

	return s.Save(c, session)
}

// GetUserID retrieves user ID from session
func (s *Store) GetUserID(c echo.Context) (string, bool) {
	session, err := s.Get(c)
	if err != nil {
		return "", false
	}

	userID, ok := session.Values[SessionKeyUser].(string)
	return userID, ok
}

// GetUser retrieves all user info from session
func (s *Store) GetUser(c echo.Context) (userID, email, name string, ok bool) {
	session, err := s.Get(c)
	if err != nil {
		return "", "", "", false
	}

	userID, ok1 := session.Values[SessionKeyUser].(string)
	email, ok2 := session.Values[SessionKeyEmail].(string)
	name, ok3 := session.Values[SessionKeyName].(string)

	return userID, email, name, ok1 && ok2 && ok3
}

// Clear clears the session
func (s *Store) Clear(c echo.Context) error {
	session, err := s.Get(c)
	if err != nil {
		return err
	}

	session.Options.MaxAge = -1
	return s.Save(c, session)
}
