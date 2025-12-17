package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	Port              string
	FlinkJobManagerURL string
	DatabaseURL       string
	DataProcessorURL  string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		FlinkJobManagerURL: getEnv("FLINK_JOBMANAGER_URL", "http://flink-jobmanager:8081"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable"),
		DataProcessorURL:  getEnv("DATA_PROCESSOR_URL", "http://data-processor:8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

