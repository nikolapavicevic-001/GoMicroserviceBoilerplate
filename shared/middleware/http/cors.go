package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CORSConfig holds the configuration for CORS middleware
type CORSConfig struct {
	// AllowOrigins is a list of origins that may access the resource
	AllowOrigins []string

	// AllowMethods is a list of methods allowed when accessing the resource
	AllowMethods []string

	// AllowHeaders is a list of headers that can be used when making the actual request
	AllowHeaders []string

	// ExposeHeaders is a list of headers that browsers are allowed to access
	ExposeHeaders []string

	// AllowCredentials indicates whether the request can include user credentials
	AllowCredentials bool

	// MaxAge indicates how long the results of a preflight request can be cached
	MaxAge int
}

// DefaultCORSConfig returns a CORS configuration with sensible defaults
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderAccept,
			"Accept-Language",
			echo.HeaderContentType,
			"Content-Language",
			echo.HeaderAuthorization,
			echo.HeaderOrigin,
			echo.HeaderXRequestedWith,
			echo.HeaderXRequestID,
		},
		ExposeHeaders: []string{
			echo.HeaderXRequestID,
		},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// ProductionCORSConfig returns a CORS configuration suitable for production
func ProductionCORSConfig(allowedOrigins []string) CORSConfig {
	return CORSConfig{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowHeaders: []string{
			echo.HeaderAccept,
			"Accept-Language",
			echo.HeaderContentType,
			"Content-Language",
			echo.HeaderAuthorization,
			echo.HeaderOrigin,
		},
		ExposeHeaders:    []string{},
		AllowCredentials: true,
		MaxAge:           3600, // 1 hour
	}
}

// CORSMiddleware returns an Echo CORS middleware with the specified configuration
func CORSMiddleware(config CORSConfig) echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     config.AllowOrigins,
		AllowMethods:     config.AllowMethods,
		AllowHeaders:     config.AllowHeaders,
		ExposeHeaders:    config.ExposeHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	})
}

// CORSMiddlewareWithDefaults returns an Echo CORS middleware with default configuration
func CORSMiddlewareWithDefaults() echo.MiddlewareFunc {
	return CORSMiddleware(DefaultCORSConfig())
}

