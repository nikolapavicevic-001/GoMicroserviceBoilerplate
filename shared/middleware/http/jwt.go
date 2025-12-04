package http

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/yourorg/boilerplate/shared/auth"
)

// JWTMiddleware creates a JWT authentication middleware for Bearer tokens
func JWTMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract token from Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "missing authorization header",
				})
			}

			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "invalid authorization header format",
				})
			}

			tokenString := parts[1]

			// Validate token
			claims, err := auth.ValidateToken(tokenString, secret)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"message": "invalid token",
				})
			}

			// Set user context
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)

			return next(c)
		}
	}
}

// GetUserID extracts user ID from context (set by JWT middleware)
func GetUserID(c echo.Context) (string, bool) {
	userID, ok := c.Get("user_id").(string)
	return userID, ok && userID != ""
}

// GetEmail extracts email from context (set by JWT middleware)
func GetEmail(c echo.Context) (string, bool) {
	email, ok := c.Get("email").(string)
	return email, ok && email != ""
}

