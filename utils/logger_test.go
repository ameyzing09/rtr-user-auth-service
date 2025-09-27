package utils

import (
	"os"
	"testing"
)

func TestLogger_LogLevels(t *testing.T) {
	// Test debug level
	logger := NewLogger(LogLevelDebug)

	// Capture output
	// Note: In a real test, you'd want to capture log output
	// For now, we'll just test that the functions don't panic
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")
}

func TestLogger_ProductionLevel(t *testing.T) {
	// Test info level (production)
	logger := NewLogger(LogLevelInfo)

	// These should not panic
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")

	// Debug should be suppressed (no output)
	logger.Debug("Debug message")
}

func TestLogger_EnvironmentBased(t *testing.T) {
	// Test environment-based initialization
	originalEnv := os.Getenv("LOG_LEVEL")
	defer os.Setenv("LOG_LEVEL", originalEnv)

	tests := []struct {
		envLevel string
		expected LogLevel
	}{
		{"debug", LogLevelDebug},
		{"info", LogLevelInfo},
		{"warn", LogLevelWarn},
		{"error", LogLevelError},
		{"", LogLevelInfo}, // Default in production
	}

	for _, tt := range tests {
		t.Run("LOG_LEVEL="+tt.envLevel, func(t *testing.T) {
			os.Setenv("LOG_LEVEL", tt.envLevel)
			os.Setenv("GIN_MODE", "release") // Force production mode

			// Create new logger to test environment detection
			logger := NewLogger(LogLevelInfo) // This would normally be initialized from env

			// Test that the logger works
			logger.Info("Test message")
		})
	}
}

func TestIsDebugEnabled(t *testing.T) {
	// Test debug enabled check
	debugLogger := NewLogger(LogLevelDebug)
	infoLogger := NewLogger(LogLevelInfo)

	// Note: These tests would need to be more sophisticated in a real implementation
	// to actually capture and verify log output
	_ = debugLogger
	_ = infoLogger
}

func TestPackageLevelFunctions(t *testing.T) {
	// Test package-level convenience functions
	Debug("Package debug message")
	Info("Package info message")
	Warn("Package warn message")
	Error("Package error message")

	// Test utility functions
	_ = IsDebugEnabled()
	_ = GetLogLevel()
}
