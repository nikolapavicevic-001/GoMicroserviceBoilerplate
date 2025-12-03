package config

import (
	"fmt"

	"github.com/yourorg/boilerplate/shared/config"
)

// Config holds the configuration for user-service
type Config struct {
	config.BaseConfig

	// Server
	GRPCPort int `env:"USER_SERVICE_GRPC_PORT" envDefault:"50051"`

	// Database
	MongoURI     string `env:"USER_SERVICE_MONGO_URI" envDefault:"mongodb://localhost:27017"`
	MongoDB      string `env:"USER_SERVICE_MONGO_DB" envDefault:"users_db"`
	MongoTimeout int    `env:"USER_SERVICE_MONGO_TIMEOUT" envDefault:"10"`

	// Kafka
	KafkaGroupID string `env:"USER_SERVICE_KAFKA_GROUP" envDefault:"user-service"`
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

	if err := cfg.validateUserService(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validateUserService() error {
	if c.GRPCPort < 1024 || c.GRPCPort > 65535 {
		return fmt.Errorf("invalid GRPC port: %d", c.GRPCPort)
	}

	if c.MongoURI == "" {
		return fmt.Errorf("MONGO_URI is required")
	}

	if c.MongoDB == "" {
		return fmt.Errorf("MONGO_DB is required")
	}

	return nil
}
