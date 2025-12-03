package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v10"
)

// BaseConfig contains common configuration shared across all services
type BaseConfig struct {
	// Logging
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat string `env:"LOG_FORMAT" envDefault:"console"`

	// Environment
	Environment string `env:"ENVIRONMENT" envDefault:"development"`

	// Kafka (shared broker)
	KafkaBrokers string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`

	// Observability
	TracingEnabled bool   `env:"TRACING_ENABLED" envDefault:"true"`
	OTLPEndpoint   string `env:"OTLP_ENDPOINT" envDefault:"localhost:4318"`
	MetricsPort    int    `env:"METRICS_PORT" envDefault:"9090"`
}

// Load loads configuration from environment variables
func Load(cfg interface{}) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("failed to parse environment variables: %w", err)
	}
	return nil
}

// GetKafkaBrokers returns Kafka brokers as a slice
func (c *BaseConfig) GetKafkaBrokers() []string {
	return strings.Split(c.KafkaBrokers, ",")
}

// IsDevelopment returns true if running in development mode
func (c *BaseConfig) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *BaseConfig) IsProduction() bool {
	return c.Environment == "production"
}

// Validate validates the base configuration
func (c *BaseConfig) Validate() error {
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLogLevels, c.LogLevel) {
		return fmt.Errorf("invalid log level: %s (must be one of: %v)", c.LogLevel, validLogLevels)
	}

	validLogFormats := []string{"json", "console"}
	if !contains(validLogFormats, c.LogFormat) {
		return fmt.Errorf("invalid log format: %s (must be one of: %v)", c.LogFormat, validLogFormats)
	}

	if c.KafkaBrokers == "" {
		return fmt.Errorf("KAFKA_BROKERS is required")
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
