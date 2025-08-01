package triggers

import (
	"context"
	"time"

	"github.com/logimos/conduktr/internal/engine"
	"github.com/logimos/conduktr/internal/persistence"

	"go.uber.org/zap"
)

// executeWorkflow is a helper function that all triggers can use to execute workflows
func executeWorkflow(ctx context.Context, engine *engine.Engine, logger *zap.Logger, eventType string, contextData map[string]interface{}) {
	workflow, exists := engine.GetWorkflowForEvent(eventType)
	if !exists {
		logger.Warn("No workflow found for event", zap.String("event", eventType))
		return
	}

	eventCtx := &persistence.EventContext{
		Event: &persistence.Event{
			Type:      eventType,
			Payload:   contextData,
			Timestamp: time.Now().Unix(),
		},
		Variables: contextData,
	}

	if _, err := engine.ExecuteWorkflow(ctx, workflow, eventCtx); err != nil {
		logger.Error("Workflow execution failed",
			zap.String("event", eventType),
			zap.Error(err))
	}
}
