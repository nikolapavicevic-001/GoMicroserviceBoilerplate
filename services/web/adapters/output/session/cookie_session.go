package session

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/microserviceboilerplate/web/domain/entity"
	"github.com/microserviceboilerplate/web/domain/port/output"
)

// CookieSessionStore implements the SessionRepository interface using cookies
type CookieSessionStore struct {
	store *sessions.CookieStore
}

// NewCookieSessionStore creates a new cookie-based session store
func NewCookieSessionStore() output.SessionRepository {
	authKey := generateRandomKey(32)
	encryptionKey := generateRandomKey(32)
	store := sessions.NewCookieStore(authKey, encryptionKey)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24, // 24 hours
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}

	return &CookieSessionStore{
		store: store,
	}
}

// Save saves user session data
func (s *CookieSessionStore) Save(ctx context.Context, sessionID string, user *entity.User) error {
	// Extract request and response writer from context
	r, ok := ctx.Value("request").(*http.Request)
	if !ok {
		return fmt.Errorf("request not found in context")
	}
	w, ok := ctx.Value("response").(http.ResponseWriter)
	if !ok {
		return fmt.Errorf("response writer not found in context")
	}

	sess, _ := s.store.Get(r, "session")
	sess.Values["authenticated"] = user.IsAuthenticated
	sess.Values["user_id"] = user.ID
	sess.Values["user_email"] = user.Email
	sess.Values["user_username"] = user.Username
	if user.Name != "" {
		sess.Values["user_name"] = user.Name
	}
	if user.RefreshToken != "" {
		sess.Values["refresh_token"] = user.RefreshToken
	}

	return sess.Save(r, w)
}

// Get retrieves user session data
func (s *CookieSessionStore) Get(ctx context.Context, sessionID string) (*entity.User, error) {
	r, ok := ctx.Value("request").(*http.Request)
	if !ok {
		return nil, fmt.Errorf("request not found in context")
	}

	sess, _ := s.store.Get(r, "session")
	if sess.Values["authenticated"] != true {
		return nil, fmt.Errorf("session not authenticated")
	}

	userID, _ := sess.Values["user_id"].(string)
	userEmail, _ := sess.Values["user_email"].(string)
	userUsername, _ := sess.Values["user_username"].(string)
	userName, _ := sess.Values["user_name"].(string)
	refreshToken, _ := sess.Values["refresh_token"].(string)

	if userID == "" || userEmail == "" {
		return nil, fmt.Errorf("invalid session data")
	}

	user := entity.NewUser(userID, userEmail, userUsername, userName)
	user.IsAuthenticated = true
	user.RefreshToken = refreshToken

	return user, nil
}

// Delete removes a session
func (s *CookieSessionStore) Delete(ctx context.Context, sessionID string) error {
	r, ok := ctx.Value("request").(*http.Request)
	if !ok {
		return fmt.Errorf("request not found in context")
	}
	w, ok := ctx.Value("response").(http.ResponseWriter)
	if !ok {
		return fmt.Errorf("response writer not found in context")
	}

	sess, _ := s.store.Get(r, "session")
	sess.Values = make(map[interface{}]interface{})
	sess.Options.MaxAge = -1
	return sess.Save(r, w)
}

// Exists checks if a session exists
func (s *CookieSessionStore) Exists(ctx context.Context, sessionID string) (bool, error) {
	r, ok := ctx.Value("request").(*http.Request)
	if !ok {
		return false, fmt.Errorf("request not found in context")
	}

	sess, _ := s.store.Get(r, "session")
	return sess.Values["authenticated"] == true, nil
}

func generateRandomKey(length int) []byte {
	key := make([]byte, length)
	rand.Read(key)
	return key
}

