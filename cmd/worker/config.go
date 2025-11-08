package main

import (
	"flag"
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

	// gRPC Unix socket configuration
	GRPCUnixSocket string // Path to Unix socket file for gRPC (if enabled)

	EnableHTTP      bool
	EnableWebSocket bool
	EnableGRPC      bool

	MaxLengthBreak int
}

// GetConfig loads configuration with priority: CLI flags > environment variables > defaults
func GetConfig() *Config {
	// Define command line flags
	logLevel := flag.String("log-level", "", "Log level (debug, info, warn, error)")
	workDir := flag.String("work-dir", "", "Working directory")
	serverHost := flag.String("host", "", "Server host address")
	serverPort := flag.String("port", "", "Server port")
	grpcUnixSocket := flag.String("grpc-unix-socket", "", "Path to Unix socket file for gRPC")
	enableHTTP := flag.String("enable-http", "", "Enable HTTP server (true/false)")
	enableWebSocket := flag.String("enable-websocket", "", "Enable WebSocket server (true/false)")
	enableGRPC := flag.String("enable-grpc", "", "Enable gRPC server (true/false)")
	maxLengthBreak := flag.Int("max-length-break", 0, "Max text length before auto-splitting (default 200)")

	flag.Parse()

	// Helper function to get value with priority: flag > env > default
	getConfigValue := func(flagValue, envKey, defaultValue string) string {
		if flagValue != "" {
			return flagValue
		}
		return getEnvOrDefault(envKey, defaultValue)
	}

	// Helper function to get boolean value with priority
	getBoolConfigValue := func(flagValue, envKey, defaultValue string) bool {
		if flagValue != "" {
			return parseBool(flagValue)
		}
		return parseBool(getEnvOrDefault(envKey, defaultValue))
	}

	// Helper function to get int value with priority
	getIntConfigValue := func(flagValue int, envKey string, defaultValue int) int {
		if flagValue != 0 {
			return flagValue
		}
		if envVal := os.Getenv(envKey); envVal != "" {
			if v, err := strconv.Atoi(envVal); err == nil {
				return v
			}
		}
		return defaultValue
	}

	return &Config{
		LogLevel: getConfigValue(*logLevel, "LOG_LEVEL", "info"),
		WorkDir:  getConfigValue(*workDir, "WORK_DIR", "./"),

		// Unified server configuration
		ServerHost: getConfigValue(*serverHost, "SERVER_HOST", "0.0.0.0"),
		ServerPort: getConfigValue(*serverPort, "SERVER_PORT", "8988"),

		// gRPC Unix socket configuration
		GRPCUnixSocket: getConfigValue(*grpcUnixSocket, "GRPC_UNIX_SOCKET", ""),

		EnableHTTP:      getBoolConfigValue(*enableHTTP, "ENABLE_HTTP", "true"),
		EnableWebSocket: getBoolConfigValue(*enableWebSocket, "ENABLE_WEBSOCKET", "true"),
		EnableGRPC:      getBoolConfigValue(*enableGRPC, "ENABLE_GRPC", "true"),

		MaxLengthBreak: getIntConfigValue(*maxLengthBreak, "MAX_LENGTH_BREAK", 200),
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
