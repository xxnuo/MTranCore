package main

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

type arrayFlags []string

func (a *arrayFlags) String() string {
	return strings.Join(*a, ", ")
}

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

type Config struct {
	LogLevel string
	ModelDir string

	ServerHost string
	ServerPort string

	GRPCUnixSocket string

	EnableHTTP      bool
	EnableWebSocket bool
	EnableGRPC      bool

	MaxLengthBreak int

	ModelPath            string
	LexicalShortlistPath string
	VocabularyPaths      []string
}

func GetConfig() *Config {
	logLevel := flag.String("log-level", "", "Log level (debug, info, warn, error)")
	modelDir := flag.String("model-dir", "", "Model directory (auto poweron)")
	serverHost := flag.String("host", "", "Server host address")
	serverPort := flag.String("port", "", "Server port")
	grpcUnixSocket := flag.String("grpc-unix-socket", "", "Path to Unix socket file for gRPC")
	enableHTTP := flag.String("enable-http", "", "Enable HTTP server (true/false)")
	enableWebSocket := flag.String("enable-websocket", "", "Enable WebSocket server (true/false)")
	enableGRPC := flag.String("enable-grpc", "", "Enable gRPC server (true/false)")
	maxLengthBreak := flag.Int("max-length-break", 0, "Max text length before auto-splitting (default 200)")

	modelPath := flag.String("model-path", "", "Model file path (auto poweron)")
	lexicalShortlistPath := flag.String("lexical-shortlist-path", "", "Lexical shortlist file path (auto poweron)")
	var vocabularyPaths arrayFlags
	flag.Var(&vocabularyPaths, "vocabulary-path", "Vocabulary file path (auto poweron, can be specified multiple times)")

	flag.Parse()

	getConfigValue := func(flagValue, envKey, defaultValue string) string {
		if flagValue != "" {
			return flagValue
		}
		return getEnvOrDefault(envKey, defaultValue)
	}

	getBoolConfigValue := func(flagValue, envKey, defaultValue string) bool {
		if flagValue != "" {
			return parseBool(flagValue)
		}
		return parseBool(getEnvOrDefault(envKey, defaultValue))
	}

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
		ModelDir: getConfigValue(*modelDir, "MODEL_DIR", ""),

		ServerHost: getConfigValue(*serverHost, "SERVER_HOST", "0.0.0.0"),
		ServerPort: getConfigValue(*serverPort, "SERVER_PORT", "8988"),

		GRPCUnixSocket: getConfigValue(*grpcUnixSocket, "GRPC_UNIX_SOCKET", ""),

		EnableHTTP:      getBoolConfigValue(*enableHTTP, "ENABLE_HTTP", "true"),
		EnableWebSocket: getBoolConfigValue(*enableWebSocket, "ENABLE_WEBSOCKET", "true"),
		EnableGRPC:      getBoolConfigValue(*enableGRPC, "ENABLE_GRPC", "true"),

		MaxLengthBreak: getIntConfigValue(*maxLengthBreak, "MAX_LENGTH_BREAK", 200),

		ModelPath:            getConfigValue(*modelPath, "MODEL_PATH", ""),
		LexicalShortlistPath: getConfigValue(*lexicalShortlistPath, "LEXICAL_SHORTLIST_PATH", ""),
		VocabularyPaths:      vocabularyPaths,
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
