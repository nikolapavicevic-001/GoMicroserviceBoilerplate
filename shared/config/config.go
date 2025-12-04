package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v10"
)

type BaseConfig struct {
	// Logging
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat string `env:"LOG_FORMAT" envDefault:"console"`

	// Environment
	Environment string `env:"ENVIRONMENT" envDefault:"development"`

	// JWT Authentication (shared across services)
	JWTSecret      string        `env:"JWT_SECRET" envDefault:"your-secret-key-change-in-production-minimum-32-characters"`
	JWTExpiry      string        `env:"JWT_EXPIRY" envDefault:"24h"`
	JWTRefreshExpiry time.Duration `env:"JWT_REFRESH_EXPIRY" envDefault:"168h"`

	// API Gateway
	HTTPPort          int           `env:"GATEWAY_HTTP_PORT" envDefault:"8080"`
	GatewayTimeout    time.Duration `env:"GATEWAY_TIMEOUT" envDefault:"30s"`
	UserServiceAddr   string        `env:"USER_SERVICE_ADDR" envDefault:"localhost:50051"`
	WebAppAddr        string        `env:"WEBAPP_ADDR" envDefault:"http://localhost:3000"`
	CORSAllowedOrigins string       `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:3000,http://localhost:8080"`

	// Web App
	WebAppHTTPPort int           `env:"WEBAPP_HTTP_PORT" envDefault:"3000"`
	WebAppTimeout  time.Duration `env:"WEBAPP_TIMEOUT" envDefault:"30s"`

	// User Service
	UserServicePort     int    `env:"USER_SERVICE_PORT" envDefault:"50051"`
	MongoURI     string `env:"MONGO_URI" envDefault:"mongodb://localhost:27017"`
	MongoDB      string `env:"MONGO_DB" envDefault:"users_db"`
	MongoTimeout int    `env:"MONGO_TIMEOUT" envDefault:"10"`
}

// Load loads configuration from environment variables
func Load(cfg interface{}) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return nil
}

// IsDevelopment returns true if running in development mode
func (c *BaseConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *BaseConfig) IsProduction() bool {
	return c.Environment == "production"
}


// GetJWTExpiry parses the JWT expiry string to a duration
func (c *BaseConfig) GetJWTExpiry() time.Duration {
	d, err := time.ParseDuration(c.JWTExpiry)
	if err != nil {
		return 24 * time.Hour // default
	}
	return d
}
