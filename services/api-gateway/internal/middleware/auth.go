package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/yourorg/boilerplate/shared/auth"
)

// JWTConfig holds JWT middleware configuration
type JWTConfig struct {
	Secret string
}

// JWTMiddleware creates a JWT authentication middleware
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

// OptionalJWTMiddleware validates JWT if present but doesn't require it
func OptionalJWTMiddleware(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return next(c)
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				claims, err := auth.ValidateToken(parts[1], secret)
				if err == nil {
					c.Set("user_id", claims.UserID)
					c.Set("email", claims.Email)
				}
			}

			return next(c)
		}
	}
}
