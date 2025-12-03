package config

import (
	"fmt"
	"time"

	"github.com/yourorg/boilerplate/shared/config"
)

// Config holds the configuration for web-app
type Config struct {
	config.BaseConfig

	// Server
	HTTPPort int           `env:"WEBAPP_HTTP_PORT" envDefault:"3000"`
	Timeout  time.Duration `env:"WEBAPP_TIMEOUT" envDefault:"30s"`

	// Services
	UserServiceAddr string `env:"WEBAPP_USER_SERVICE_ADDR" envDefault:"localhost:50051"`
	GatewayAddr     string `env:"WEBAPP_GATEWAY_ADDR" envDefault:"http://localhost:8080"`

	// Session
	SessionSecret string        `env:"SESSION_SECRET" envDefault:"change-this-secret-in-production-minimum-32-chars"`
	SessionMaxAge time.Duration `env:"SESSION_MAX_AGE" envDefault:"24h"`

	// JWT
	JWTSecret string        `env:"JWT_SECRET" envDefault:"your-secret-key-change-in-production"`
	JWTExpiry time.Duration `env:"JWT_EXPIRY" envDefault:"24h"`

	// OAuth2 - Google
	OAuth2GoogleClientID     string `env:"OAUTH2_GOOGLE_CLIENT_ID"`
	OAuth2GoogleClientSecret string `env:"OAUTH2_GOOGLE_CLIENT_SECRET"`
	OAuth2GoogleRedirectURL  string `env:"OAUTH2_GOOGLE_REDIRECT_URL" envDefault:"http://localhost:3000/auth/google/callback"`

	// OAuth2 - GitHub
	OAuth2GitHubClientID     string `env:"OAUTH2_GITHUB_CLIENT_ID"`
	OAuth2GitHubClientSecret string `env:"OAUTH2_GITHUB_CLIENT_SECRET"`
	OAuth2GitHubRedirectURL  string `env:"OAUTH2_GITHUB_REDIRECT_URL" envDefault:"http://localhost:3000/auth/github/callback"`

	// CSRF
	CSRFSecret string `env:"CSRF_SECRET" envDefault:"change-this-csrf-secret-in-production"`
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

	if err := cfg.validateWebApp(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validateWebApp() error {
	if c.HTTPPort < 1024 || c.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", c.HTTPPort)
	}

	if c.UserServiceAddr == "" {
		return fmt.Errorf("USER_SERVICE_ADDR is required")
	}

	if len(c.SessionSecret) < 32 {
		return fmt.Errorf("SESSION_SECRET must be at least 32 characters")
	}

	return nil
}
