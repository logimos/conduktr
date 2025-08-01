package actions

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Action defines the interface for workflow actions
type Action interface {
	Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}

// Registry manages available workflow actions
type Registry struct {
	logger  *zap.Logger
	actions map[string]Action
}

// NewRegistry creates a new action registry with built-in actions
func NewRegistry(logger *zap.Logger) *Registry {
	registry := &Registry{
		logger:  logger,
		actions: make(map[string]Action),
	}

	// Register built-in actions
	registry.RegisterAction("http.request", NewHTTPAction(logger))
	registry.RegisterAction("shell.exec", NewShellAction(logger))
	registry.RegisterAction("log.info", NewLogAction(logger))

	return registry
}

// RegisterAction registers a new action
func (r *Registry) RegisterAction(name string, action Action) {
	r.actions[name] = action
	r.logger.Info("Action registered", zap.String("name", name))
}

// GetAction retrieves an action by name
func (r *Registry) GetAction(name string) (Action, error) {
	action, exists := r.actions[name]
	if !exists {
		return nil, fmt.Errorf("action not found: %s", name)
	}
	return action, nil
}

// ListActions returns all registered action names
func (r *Registry) ListActions() []string {
	names := make([]string, 0, len(r.actions))
	for name := range r.actions {
		names = append(names, name)
	}
	return names
}
