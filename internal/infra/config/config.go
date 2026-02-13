package config

import (
	"os"
)

// Config holds the application configuration.
type Config struct {
	// Server settings
	Port string
	Env  string

	// Logging
	LogLevel  string
	LogFormat string
}

// Load loads configuration from environment variables.
func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		Env:       getEnv("ENV", "development"),
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),
	}
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// getEnv gets an environment variable with a fallback default.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
