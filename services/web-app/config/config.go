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

	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	return nil
}
