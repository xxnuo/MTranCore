package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// LogLevel represents different log levels
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger wraps the standard logger with level control
type Logger struct {
	level  LogLevel
	logger *log.Logger
}

var globalLogger *Logger

func init() {
	// Initialize with a default logger to avoid nil panics before InitLogger is called
	globalLogger = &Logger{
		level:  LogLevelInfo,
		logger: log.New(os.Stdout, "", 0),
	}
}

// InitLogger initializes the global logger with specified level
func InitLogger(levelStr string) {
	levelStr = strings.ToLower(strings.TrimSpace(levelStr))

	var level LogLevel
	switch levelStr {
	case "debug":
		level = LogLevelDebug
	case "info":
		level = LogLevelInfo
	case "warn", "warning":
		level = LogLevelWarn
	case "error":
		level = LogLevelError
	default:
		level = LogLevelInfo
	}

	globalLogger = &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", 0),
	}
}

// formatMessage formats a log message with timestamp and level
func (l *Logger) formatMessage(level string, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	message := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s [%s] %s", timestamp, level, message)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LogLevelDebug {
		l.logger.Println(l.formatMessage("DEBUG", format, args...))
	}
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LogLevelInfo {
		l.logger.Println(l.formatMessage("INFO", format, args...))
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LogLevelWarn {
		l.logger.Println(l.formatMessage("WARN", format, args...))
	}
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LogLevelError {
		l.logger.Println(l.formatMessage("ERROR", format, args...))
	}
}

// Fatal logs a fatal error message and exits
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.logger.Println(l.formatMessage("FATAL", format, args...))
	os.Exit(1)
}

// GetWriter returns an io.Writer that can be used by third-party libraries
// The writer will output logs at the specified level
func (l *Logger) GetWriter(level LogLevel) io.Writer {
	return &logWriter{logger: l, level: level}
}

// logWriter implements io.Writer for redirecting third-party logs
type logWriter struct {
	logger *Logger
	level  LogLevel
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	message := strings.TrimSpace(string(p))
	if message == "" {
		return len(p), nil
	}

	switch w.level {
	case LogLevelDebug:
		w.logger.Debug("%s", message)
	case LogLevelInfo:
		w.logger.Info("%s", message)
	case LogLevelWarn:
		w.logger.Warn("%s", message)
	case LogLevelError:
		w.logger.Error("%s", message)
	}

	return len(p), nil
}

// Global convenience functions
func Debug(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(format, args...)
	}
}

func Info(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Error(format, args...)
	}
}

func Fatal(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Fatal(format, args...)
	}
	os.Exit(1)
}

func GetLogger() *Logger {
	return globalLogger
}

// DiscardLogger is a logger that discards all output
type DiscardLogger struct{}

func (d *DiscardLogger) Printf(format string, args ...interface{}) {
	// Discard all output
}
