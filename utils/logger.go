package utils

import (
	"log"
	"os"
	"strings"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

// Logger provides conditional logging based on environment
type Logger struct {
	level LogLevel
}

var (
	// Default logger instance
	defaultLogger *Logger
)

func init() {
	// Initialize default logger based on environment
	env := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch env {
	case "debug":
		defaultLogger = &Logger{level: LogLevelDebug}
	case "info":
		defaultLogger = &Logger{level: LogLevelInfo}
	case "warn":
		defaultLogger = &Logger{level: LogLevelWarn}
	case "error":
		defaultLogger = &Logger{level: LogLevelError}
	default:
		// Default to info level in production, debug in development
		if os.Getenv("GIN_MODE") == "release" {
			defaultLogger = &Logger{level: LogLevelInfo}
		} else {
			defaultLogger = &Logger{level: LogLevelDebug}
		}
	}
}

// NewLogger creates a new logger with the specified level
func NewLogger(level LogLevel) *Logger {
	return &Logger{level: level}
}

// Debug logs debug messages (only in debug mode)
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level >= LogLevelDebug {
		log.Printf("[DEBUG] "+format, v...)
	}
}

// Info logs info messages
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level >= LogLevelInfo {
		log.Printf("[INFO] "+format, v...)
	}
}

// Warn logs warning messages
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level >= LogLevelWarn {
		log.Printf("[WARN] "+format, v...)
	}
}

// Error logs error messages
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level >= LogLevelError {
		log.Printf("[ERROR] "+format, v...)
	}
}

// Package-level convenience functions using default logger

// Debug logs debug messages using the default logger
func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

// Info logs info messages using the default logger
func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

// Warn logs warning messages using the default logger
func Warn(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

// Error logs error messages using the default logger
func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

// IsDebugEnabled returns true if debug logging is enabled
func IsDebugEnabled() bool {
	return defaultLogger.level >= LogLevelDebug
}

// GetLogLevel returns the current log level
func GetLogLevel() LogLevel {
	return defaultLogger.level
}
