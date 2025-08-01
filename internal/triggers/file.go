package triggers

import (
	"context"
	"path/filepath"
	"time"

	"github.com/logimos/conduktr/internal/engine"
	"github.com/logimos/conduktr/internal/persistence"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// FileTrigger handles file system event triggers
type FileTrigger struct {
	logger  *zap.Logger
	engine  *engine.Engine
	watcher *fsnotify.Watcher
	done    chan bool
}

// NewFileTrigger creates a new file trigger
func NewFileTrigger(logger *zap.Logger, engine *engine.Engine) *FileTrigger {
	return &FileTrigger{
		logger: logger,
		engine: engine,
		done:   make(chan bool),
	}
}

// Start starts watching for file system events
func (f *FileTrigger) Start(watchDir string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	f.watcher = watcher

	go func() {
		defer watcher.Close()

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				f.logger.Debug("File system event",
					zap.String("file", event.Name),
					zap.String("op", event.Op.String()))

				f.handleFileEvent(event)

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				f.logger.Error("File watcher error", zap.Error(err))

			case <-f.done:
				return
			}
		}
	}()

	// Add the directory to watch
	err = watcher.Add(watchDir)
	if err != nil {
		return err
	}

	f.logger.Info("File trigger started", zap.String("watch_dir", watchDir))
	return nil
}

// Stop stops the file trigger
func (f *FileTrigger) Stop() {
	close(f.done)
	if f.watcher != nil {
		f.watcher.Close()
	}
}

// handleFileEvent processes a file system event
func (f *FileTrigger) handleFileEvent(event fsnotify.Event) {
	var eventType string

	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		eventType = "file.created"
	case event.Op&fsnotify.Write == fsnotify.Write:
		eventType = "file.modified"
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		eventType = "file.deleted"
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		eventType = "file.renamed"
	case event.Op&fsnotify.Chmod == fsnotify.Chmod:
		eventType = "file.chmod"
	default:
		// Unknown operation, skip
		return
	}

	// Check if there's a workflow for this event type
	workflow, exists := f.engine.GetWorkflowForEvent(eventType)
	if !exists {
		return
	}

	// Create event context
	eventCtx := &persistence.EventContext{
		Event: &persistence.Event{
			Type: eventType,
			Payload: map[string]interface{}{
				"file_path": event.Name,
				"file_name": filepath.Base(event.Name),
				"file_dir":  filepath.Dir(event.Name),
				"file_ext":  filepath.Ext(event.Name),
			},
			Metadata: map[string]interface{}{
				"operation": event.Op.String(),
			},
			Timestamp: time.Now().Unix(),
		},
		Variables: make(map[string]interface{}),
	}

	f.logger.Info("Triggering workflow for file event",
		zap.String("event", eventType),
		zap.String("file", event.Name),
		zap.String("workflow", workflow.Name))

	// Execute workflow asynchronously
	go func() {
		ctx := context.Background()
		instanceID, err := f.engine.ExecuteWorkflow(ctx, workflow, eventCtx)
		if err != nil {
			f.logger.Error("Workflow execution failed",
				zap.String("instance_id", instanceID),
				zap.Error(err))
		} else {
			f.logger.Info("Workflow execution completed",
				zap.String("instance_id", instanceID))
		}
	}()
}
