package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	Port              string
	KeycloakURL       string
	KeycloakRealm     string
	KeycloakClientID  string
	KeycloakSecret    string
	RedirectURL       string
	KeycloakAdmin     string
	KeycloakAdminPass string
	TemplatesPath     string
	NATSURL           string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		KeycloakURL:       getEnv("KEYCLOAK_URL", "http://keycloak:8080"),
		KeycloakRealm:     getEnv("KEYCLOAK_REALM", "microservice"),
		KeycloakClientID:  getEnv("KEYCLOAK_CLIENT_ID", "web-client"),
		KeycloakSecret:    getEnv("KEYCLOAK_CLIENT_SECRET", "web-client-secret"),
		RedirectURL:       getEnv("REDIRECT_URL", "http://localhost:8080/auth/callback"),
		KeycloakAdmin:     getEnv("KEYCLOAK_ADMIN", "admin"),
		KeycloakAdminPass: getEnv("KEYCLOAK_ADMIN_PASSWORD", "admin"),
		TemplatesPath:     getEnv("TEMPLATES_PATH", "templates/*.html"),
		NATSURL:           getEnv("NATS_URL", "nats://nats:4222"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

