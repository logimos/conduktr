package actions

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ShellAction implements shell command execution
type ShellAction struct {
	logger *zap.Logger
}

// NewShellAction creates a new shell action
func NewShellAction(logger *zap.Logger) *ShellAction {
	return &ShellAction{
		logger: logger,
	}
}

// Execute runs a shell command
func (s *ShellAction) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Parse input parameters
	command, ok := input["command"].(string)
	if !ok || command == "" {
		return nil, fmt.Errorf("command parameter is required")
	}

	// Set working directory if provided
	workDir := ""
	if wd, ok := input["working_dir"].(string); ok {
		workDir = wd
	}

	// Set timeout (default 30 seconds)
	timeout := 30 * time.Second
	if t, ok := input["timeout"].(float64); ok {
		timeout = time.Duration(t) * time.Second
	}

	// Parse environment variables
	env := os.Environ()
	if envVars, ok := input["env"].(map[string]interface{}); ok {
		for key, value := range envVars {
			env = append(env, fmt.Sprintf("%s=%v", key, value))
		}
	}

	s.logger.Info("Executing shell command", 
		zap.String("command", command),
		zap.String("working_dir", workDir))

	// Create context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Parse command (simple space-based splitting)
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	// Create command
	cmd := exec.CommandContext(cmdCtx, parts[0], parts[1:]...)
	
	if workDir != "" {
		cmd.Dir = workDir
	}
	
	cmd.Env = env

	// Execute command and capture output
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	result := map[string]interface{}{
		"command":    command,
		"output":     outputStr,
		"success":    err == nil,
		"exit_code":  0,
	}

	if err != nil {
		// Try to get exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			result["exit_code"] = exitError.ExitCode()
		} else {
			result["exit_code"] = -1
		}
		result["error"] = err.Error()
		
		s.logger.Error("Shell command failed", 
			zap.String("command", command),
			zap.Error(err),
			zap.String("output", outputStr))
		
		return result, fmt.Errorf("command failed: %w", err)
	}

	s.logger.Info("Shell command completed", 
		zap.String("command", command),
		zap.String("output", outputStr))

	return result, nil
}
