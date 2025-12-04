package auth

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/yourorg/boilerplate/shared/auth"
)

const (
	CookieName = "auth_token"
)

// SetJWTCookie sets a JWT token as an httpOnly cookie
func SetJWTCookie(c echo.Context, userID, email, secret string, expiry time.Duration) error {
	// Generate JWT token
	token, err := auth.GenerateToken(userID, email, secret, expiry)
	if err != nil {
		return err
	}

	// Set cookie
	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(expiry.Seconds()),
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)

	return nil
}

// ClearJWTCookie clears the JWT cookie
func ClearJWTCookie(c echo.Context) {
	cookie := &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(cookie)
}

// GetJWTCookie extracts and validates JWT from cookie
func GetJWTCookie(c echo.Context, secret string) (*auth.Claims, error) {
	cookie, err := c.Cookie(CookieName)
	if err != nil {
		return nil, err
	}

	claims, err := auth.ValidateToken(cookie.Value, secret)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

