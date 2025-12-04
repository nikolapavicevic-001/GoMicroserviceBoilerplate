package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	gatewayAuth "github.com/yourorg/boilerplate/services/api-gateway/internal/auth"
)

// JWTCookieMiddleware creates a JWT authentication middleware that reads from cookies
func JWTCookieMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Try to get JWT from cookie
			claims, err := gatewayAuth.GetJWTCookie(c, secret)
			if err != nil {
				// If no valid cookie, redirect to login for web requests, return 401 for API requests
				if c.Request().Header.Get("Accept") == "application/json" || c.Request().URL.Path[:4] == "/api" {
					return c.JSON(http.StatusUnauthorized, map[string]string{
						"message": "authentication required",
					})
				}
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			// Set user context (same as existing JWT middleware)
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)

			return next(c)
		}
	}
}

