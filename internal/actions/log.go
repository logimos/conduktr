package actions

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// LogAction implements logging actions
type LogAction struct {
	logger *zap.Logger
}

// NewLogAction creates a new log action
func NewLogAction(logger *zap.Logger) *LogAction {
	return &LogAction{
		logger: logger,
	}
}

// Execute logs a message
func (l *LogAction) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Parse input parameters
	message, ok := input["message"].(string)
	if !ok || message == "" {
		return nil, fmt.Errorf("message parameter is required")
	}

	// Get log level (default to info)
	level := "info"
	if lvl, ok := input["level"].(string); ok {
		level = lvl
	}

	// Get additional fields
	fields := make([]zap.Field, 0)
	if extraFields, ok := input["fields"].(map[string]interface{}); ok {
		for key, value := range extraFields {
			fields = append(fields, zap.Any(key, value))
		}
	}

	// Log the message at the appropriate level
	switch level {
	case "debug":
		l.logger.Debug(message, fields...)
	case "info":
		l.logger.Info(message, fields...)
	case "warn", "warning":
		l.logger.Warn(message, fields...)
	case "error":
		l.logger.Error(message, fields...)
	default:
		l.logger.Info(message, fields...)
	}

	result := map[string]interface{}{
		"message": message,
		"level":   level,
		"success": true,
	}

	return result, nil
}
