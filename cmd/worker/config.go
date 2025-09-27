package main

import (
	"os"
	"strconv"
	"strings"
)

// Config holds the worker service configuration
type Config struct {
	LogLevel string
	WorkDir  string

	// Unified server configuration
	ServerHost string
	ServerPort string

	EnableHTTP      bool
	EnableWebSocket bool
	EnableGRPC      bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		LogLevel: getEnvOrDefault("LOG_LEVEL", "info"),
		WorkDir:  getEnvOrDefault("WORK_DIR", "./"),

		// Unified server configuration
		ServerHost: getEnvOrDefault("SERVER_HOST", "0.0.0.0"),
		ServerPort: getEnvOrDefault("SERVER_PORT", "8988"),

		EnableHTTP:      parseBool(getEnvOrDefault("ENABLE_HTTP", "true")),
		EnableWebSocket: parseBool(getEnvOrDefault("ENABLE_WEBSOCKET", "true")),
		EnableGRPC:      parseBool(getEnvOrDefault("ENABLE_GRPC", "true")),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseBool parses boolean values from environment variables
// Supports: true/false, 1/0, yes/no (case insensitive)
func parseBool(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		// Try standard parsing as fallback
		result, _ := strconv.ParseBool(value)
		return result
	}
}
