package config

import (
	"fmt"
	"time"

	"github.com/yourorg/boilerplate/shared/config"
)

// Config holds the configuration for api-gateway
type Config struct {
	config.BaseConfig

	// Server
	HTTPPort int           `env:"GATEWAY_HTTP_PORT" envDefault:"8080"`
	Timeout  time.Duration `env:"GATEWAY_TIMEOUT" envDefault:"30s"`

	// Services
	UserServiceAddr string `env:"GATEWAY_USER_SERVICE_ADDR" envDefault:"localhost:50051"`

	// JWT
	JWTSecret       string        `env:"JWT_SECRET" envDefault:"your-secret-key-change-in-production"`
	JWTExpiry       time.Duration `env:"JWT_EXPIRY" envDefault:"24h"`
	JWTRefreshExpiry time.Duration `env:"JWT_REFRESH_EXPIRY" envDefault:"168h"`

	// OAuth2 - Google
	OAuth2GoogleClientID     string `env:"OAUTH2_GOOGLE_CLIENT_ID"`
	OAuth2GoogleClientSecret string `env:"OAUTH2_GOOGLE_CLIENT_SECRET"`
	OAuth2GoogleRedirectURL  string `env:"OAUTH2_GOOGLE_REDIRECT_URL" envDefault:"http://localhost:8080/auth/google/callback"`

	// OAuth2 - GitHub
	OAuth2GitHubClientID     string `env:"OAUTH2_GITHUB_CLIENT_ID"`
	OAuth2GitHubClientSecret string `env:"OAUTH2_GITHUB_CLIENT_SECRET"`
	OAuth2GitHubRedirectURL  string `env:"OAUTH2_GITHUB_REDIRECT_URL" envDefault:"http://localhost:8080/auth/github/callback"`

	// CORS
	CORSAllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:3000,http://localhost:8080"`
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	if err := config.Load(cfg); err != nil {
		return nil, err
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if err := cfg.validateGateway(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validateGateway() error {
	if c.HTTPPort < 1024 || c.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", c.HTTPPort)
	}

	if c.UserServiceAddr == "" {
		return fmt.Errorf("USER_SERVICE_ADDR is required")
	}

	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	return nil
}
