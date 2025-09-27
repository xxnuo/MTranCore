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

	EnableHTTP      bool
	HTTPHost        string
	HTTPPort        string

	EnableWebSocket bool
	WebSocketHost   string
	WebSocketPort   string

	EnableGRPC bool
	GRPCHost   string
	GRPCPort   string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		LogLevel: getEnvOrDefault("LOG_LEVEL", "info"),
		WorkDir:  getEnvOrDefault("WORK_DIR", "./"),

		EnableHTTP: parseBool(getEnvOrDefault("ENABLE_HTTP", "true")),
		HTTPHost:   getEnvOrDefault("HTTP_HOST", "0.0.0.0"),
		HTTPPort:   getEnvOrDefault("HTTP_PORT", "8989"),

		EnableWebSocket: parseBool(getEnvOrDefault("ENABLE_WEBSOCKET", "true")),
		WebSocketHost:   getEnvOrDefault("WEBSOCKET_HOST", "0.0.0.0"),
		WebSocketPort:   getEnvOrDefault("WEBSOCKET_PORT", "8990"),

		EnableGRPC: parseBool(getEnvOrDefault("ENABLE_GRPC", "true")),
		GRPCHost:   getEnvOrDefault("GRPC_HOST", "0.0.0.0"),
		GRPCPort:   getEnvOrDefault("GRPC_PORT", "8991"),
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
