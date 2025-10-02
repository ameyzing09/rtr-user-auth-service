package messaging

import (
	"fmt"
	"rtr-user-auth-service/utils"
)

// LoggerAdapter adapts utils.Logger to messaging.Logger interface
type LoggerAdapter struct {
	logger *utils.Logger
}

// NewLoggerAdapter creates a new logger adapter
func NewLoggerAdapter(logger *utils.Logger) *LoggerAdapter {
	return &LoggerAdapter{logger: logger}
}

// Info logs an info message with optional key-value pairs
func (l *LoggerAdapter) Info(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.logger.Info("%s %s", msg, formatFields(fields...))
	} else {
		l.logger.Info(msg)
	}
}

// Error logs an error message with optional key-value pairs
func (l *LoggerAdapter) Error(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.logger.Error("%s %s", msg, formatFields(fields...))
	} else {
		l.logger.Error(msg)
	}
}

// Warn logs a warning message with optional key-value pairs
func (l *LoggerAdapter) Warn(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.logger.Warn("%s %s", msg, formatFields(fields...))
	} else {
		l.logger.Warn(msg)
	}
}

// formatFields formats key-value pairs for logging
func formatFields(fields ...interface{}) string {
	if len(fields) == 0 {
		return ""
	}

	if len(fields)%2 != 0 {
		return fmt.Sprintf("%v", fields)
	}

	var result string
	for i := 0; i < len(fields); i += 2 {
		if i > 0 {
			result += " "
		}
		result += fmt.Sprintf("%v=%v", fields[i], fields[i+1])
	}
	return result
}
