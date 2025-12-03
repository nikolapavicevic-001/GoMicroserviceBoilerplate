package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/yourorg/boilerplate/services/web-app/internal/session"
)

// RequireAuth middleware ensures user is authenticated
func RequireAuth(sessionStore *session.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := sessionStore.GetUserID(c)
			if !ok || userID == "" {
				// For HTMX requests, return 401 which htmx can handle
				if c.Request().Header.Get("HX-Request") == "true" {
					c.Response().Header().Set("HX-Redirect", "/login")
					return c.NoContent(http.StatusUnauthorized)
				}
				// For regular requests, redirect to login
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			// Store user ID in context for handlers
			c.Set("user_id", userID)
			return next(c)
		}
	}
}

// RedirectIfAuthenticated middleware redirects to dashboard if already logged in
func RedirectIfAuthenticated(sessionStore *session.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := sessionStore.GetUserID(c)
			if ok && userID != "" {
				return c.Redirect(http.StatusSeeOther, "/dashboard")
			}
			return next(c)
		}
	}
}
