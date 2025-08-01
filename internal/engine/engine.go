package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/logimos/conduktr/internal/actions"
	"github.com/logimos/conduktr/internal/persistence"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Engine is the core workflow execution engine
type Engine struct {
	logger      *zap.Logger
	registry    *actions.Registry
	persistence persistence.Store
	workflows   map[string]*Workflow
}

// NewEngine creates a new workflow engine
func NewEngine(logger *zap.Logger, store persistence.Store) *Engine {
	return &Engine{
		logger:      logger,
		registry:    actions.NewRegistry(logger),
		persistence: store,
		workflows:   make(map[string]*Workflow),
	}
}

// RegisterWorkflow registers a workflow with the engine
func (e *Engine) RegisterWorkflow(workflow *Workflow) {
	e.workflows[workflow.On.Event] = workflow
	e.logger.Info("Workflow registered",
		zap.String("name", workflow.Name),
		zap.String("event", workflow.On.Event))
}

// GetWorkflowForEvent returns the workflow associated with an event type
func (e *Engine) GetWorkflowForEvent(eventType string) (*Workflow, bool) {
	workflow, exists := e.workflows[eventType]
	return workflow, exists
}

// ExecuteWorkflow executes a workflow with the given event context
func (e *Engine) ExecuteWorkflow(ctx context.Context, workflow *Workflow, eventCtx *persistence.EventContext) (string, error) {
	instanceID := uuid.New().String()

	instance := &persistence.WorkflowInstance{
		ID:           instanceID,
		WorkflowName: workflow.Name,
		Status:       "running",
		StartTime:    time.Now(),
		Context:      eventCtx,
		Steps:        make([]persistence.StepExecution, 0),
	}

	e.logger.Info("Starting workflow execution",
		zap.String("instance_id", instanceID),
		zap.String("workflow", workflow.Name))

	// Save initial instance state
	if err := e.persistence.SaveWorkflowInstance(instance); err != nil {
		e.logger.Error("Failed to save workflow instance", zap.Error(err))
	}

	// Execute workflow steps
	for _, step := range workflow.Workflow {
		stepExec := persistence.StepExecution{
			Name:      step.Name,
			Status:    "running",
			StartTime: time.Now(),
			Input:     make(map[string]interface{}),
			Retries:   0,
		}

		// Check conditions
		if step.If != "" {
			shouldExecute, err := e.evaluateCondition(step.If, eventCtx)
			if err != nil {
				stepExec.Status = "failed"
				stepExec.Error = fmt.Sprintf("condition evaluation failed: %v", err)
				instance.Steps = append(instance.Steps, stepExec)
				continue
			}
			if !shouldExecute {
				stepExec.Status = "skipped"
				now := time.Now()
				stepExec.EndTime = &now
				instance.Steps = append(instance.Steps, stepExec)
				continue
			}
		}

		// Execute step with retry logic
		var err error
		maxRetries := 1
		if step.Retry != nil && step.Retry.Max > 0 {
			maxRetries = step.Retry.Max
		}

		for attempt := 0; attempt < maxRetries; attempt++ {
			if attempt > 0 {
				// Apply backoff
				backoffDuration := e.calculateBackoff(step.Retry, attempt)
				e.logger.Info("Retrying step after backoff",
					zap.String("step", step.Name),
					zap.Int("attempt", attempt+1),
					zap.Duration("backoff", backoffDuration))
				time.Sleep(backoffDuration)
			}

			stepExec.Retries = attempt
			err = e.executeStep(ctx, &step, eventCtx, &stepExec)
			if err == nil {
				break
			}

			e.logger.Warn("Step execution failed",
				zap.String("step", step.Name),
				zap.Int("attempt", attempt+1),
				zap.Error(err))
		}

		if err != nil {
			stepExec.Status = "failed"
			stepExec.Error = err.Error()
			instance.Status = "failed"
			instance.Error = fmt.Sprintf("Step '%s' failed: %v", step.Name, err)

			now := time.Now()
			stepExec.EndTime = &now
			instance.EndTime = &now
			instance.Steps = append(instance.Steps, stepExec)

			// Save failed state
			e.persistence.SaveWorkflowInstance(instance)

			return instanceID, fmt.Errorf("workflow failed at step '%s': %w", step.Name, err)
		}

		stepExec.Status = "completed"
		now := time.Now()
		stepExec.EndTime = &now
		instance.Steps = append(instance.Steps, stepExec)

		// Update instance
		if err := e.persistence.SaveWorkflowInstance(instance); err != nil {
			e.logger.Error("Failed to save workflow instance", zap.Error(err))
		}
	}

	// Mark workflow as completed
	instance.Status = "completed"
	now := time.Now()
	instance.EndTime = &now

	if err := e.persistence.SaveWorkflowInstance(instance); err != nil {
		e.logger.Error("Failed to save completed workflow instance", zap.Error(err))
	}

	e.logger.Info("Workflow execution completed",
		zap.String("instance_id", instanceID),
		zap.String("workflow", workflow.Name))

	return instanceID, nil
}

// executeStep executes a single workflow step
func (e *Engine) executeStep(ctx context.Context, step *WorkflowStep, eventCtx *persistence.EventContext, stepExec *persistence.StepExecution) error {
	action, err := e.registry.GetAction(step.Action)
	if err != nil {
		return fmt.Errorf("action not found: %s", step.Action)
	}

	// Prepare step input by resolving templates
	stepInput := make(map[string]interface{})
	for key, value := range step.Config {
		resolvedValue, err := eventCtx.ResolveTemplate(fmt.Sprintf("%v", value))
		if err != nil {
			return fmt.Errorf("template resolution failed for %s: %w", key, err)
		}
		stepInput[key] = resolvedValue
	}

	stepExec.Input = stepInput

	// Execute the action
	output, err := action.Execute(ctx, stepInput)
	if err != nil {
		return err
	}

	stepExec.Output = output

	// Update context with step output
	if stepExec.Output != nil {
		eventCtx.Variables[step.Name] = stepExec.Output
	}

	return nil
}

// evaluateCondition evaluates a condition string against the event context
func (e *Engine) evaluateCondition(condition string, eventCtx *persistence.EventContext) (bool, error) {
	// Simple condition evaluation - in a real implementation, you'd use a more sophisticated expression evaluator
	resolvedCondition, err := eventCtx.ResolveTemplate(condition)
	if err != nil {
		return false, err
	}

	// Basic true/false evaluation
	switch resolvedCondition {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no", "":
		return false, nil
	default:
		// For more complex conditions, you'd implement a proper expression evaluator
		return resolvedCondition != "", nil
	}
}

// calculateBackoff calculates the backoff duration for retries
func (e *Engine) calculateBackoff(retry *RetryConfig, attempt int) time.Duration {
	if retry == nil {
		return time.Second
	}

	baseDuration := time.Second
	if retry.Backoff == "exponential" {
		return baseDuration * time.Duration(1<<uint(attempt))
	}

	return baseDuration
}

// GetWorkflowInstance retrieves a workflow instance by ID
func (e *Engine) GetWorkflowInstance(instanceID string) (*persistence.WorkflowInstance, error) {
	return e.persistence.GetWorkflowInstance(instanceID)
}
