package persistence

import (
        "bytes"
        "encoding/json"
        "fmt"
        "os"
        "path/filepath"
        "reflect"
        "text/template"
        "time"
)

// WorkflowInstance represents a running workflow instance
type WorkflowInstance struct {
        ID           string                 `json:"id"`
        WorkflowName string                 `json:"workflow_name"`
        Status       string                 `json:"status"`
        StartTime    time.Time              `json:"start_time"`
        EndTime      *time.Time             `json:"end_time,omitempty"`
        Context      *EventContext          `json:"context"`
        Steps        []StepExecution        `json:"steps"`
        Error        string                 `json:"error,omitempty"`
}

// StepExecution represents the execution of a single workflow step
type StepExecution struct {
        Name      string                 `json:"name"`
        Status    string                 `json:"status"`
        StartTime time.Time              `json:"start_time"`
        EndTime   *time.Time             `json:"end_time,omitempty"`
        Input     map[string]interface{} `json:"input"`
        Output    map[string]interface{} `json:"output"`
        Error     string                 `json:"error,omitempty"`
        Retries   int                    `json:"retries"`
}

// Event represents an incoming event that triggers a workflow
type Event struct {
        Type      string                 `json:"type"`
        Payload   map[string]interface{} `json:"payload"`
        Metadata  map[string]interface{} `json:"metadata"`
        Timestamp int64                  `json:"timestamp"`
}

// EventContext holds the context for a workflow execution
type EventContext struct {
        Event     *Event                 `json:"event"`
        Variables map[string]interface{} `json:"variables"`
}

// ResolveTemplate resolves template variables in a string using the event context
func (ctx *EventContext) ResolveTemplate(templateStr string) (string, error) {
        // Import the template package to resolve templates
        // Since we moved the EventContext here, we can import template without cycles
        return resolveTemplate(templateStr, ctx)
}

// GetVariable retrieves a variable from the context
func (ctx *EventContext) GetVariable(name string) (interface{}, bool) {
        value, exists := ctx.Variables[name]
        return value, exists
}

// SetVariable sets a variable in the context
func (ctx *EventContext) SetVariable(name string, value interface{}) {
        ctx.Variables[name] = value
}

// Store defines the interface for workflow persistence
type Store interface {
        SaveWorkflowInstance(instance *WorkflowInstance) error
        GetWorkflowInstance(instanceID string) (*WorkflowInstance, error)
        ListWorkflowInstances() ([]*WorkflowInstance, error)
}

// JSONPersistence implements file-based JSON persistence
type JSONPersistence struct {
        dataDir string
}

// NewJSONPersistence creates a new JSON persistence store
func NewJSONPersistence(dataDir string) *JSONPersistence {
        // Create data directory if it doesn't exist
        os.MkdirAll(dataDir, 0755)
        
        return &JSONPersistence{
                dataDir: dataDir,
        }
}

// SaveWorkflowInstance saves a workflow instance to a JSON file
func (j *JSONPersistence) SaveWorkflowInstance(instance *WorkflowInstance) error {
        filename := filepath.Join(j.dataDir, fmt.Sprintf("%s.json", instance.ID))
        
        data, err := json.MarshalIndent(instance, "", "  ")
        if err != nil {
                return fmt.Errorf("failed to marshal instance: %w", err)
        }

        if err := os.WriteFile(filename, data, 0644); err != nil {
                return fmt.Errorf("failed to write instance file: %w", err)
        }

        return nil
}

// GetWorkflowInstance retrieves a workflow instance from a JSON file
func (j *JSONPersistence) GetWorkflowInstance(instanceID string) (*WorkflowInstance, error) {
        filename := filepath.Join(j.dataDir, fmt.Sprintf("%s.json", instanceID))
        
        data, err := os.ReadFile(filename)
        if err != nil {
                if os.IsNotExist(err) {
                        return nil, fmt.Errorf("instance not found: %s", instanceID)
                }
                return nil, fmt.Errorf("failed to read instance file: %w", err)
        }

        var instance WorkflowInstance
        if err := json.Unmarshal(data, &instance); err != nil {
                return nil, fmt.Errorf("failed to unmarshal instance: %w", err)
        }

        return &instance, nil
}

// ListWorkflowInstances retrieves all workflow instances
func (j *JSONPersistence) ListWorkflowInstances() ([]*WorkflowInstance, error) {
        files, err := filepath.Glob(filepath.Join(j.dataDir, "*.json"))
        if err != nil {
                return nil, fmt.Errorf("failed to list instance files: %w", err)
        }

        instances := make([]*WorkflowInstance, 0, len(files))
        
        for _, file := range files {
                data, err := os.ReadFile(file)
                if err != nil {
                        continue // Skip files that can't be read
                }

                var instance WorkflowInstance
                if err := json.Unmarshal(data, &instance); err != nil {
                        continue // Skip files that can't be parsed
                }

                instances = append(instances, &instance)
        }

        return instances, nil
}

// resolveTemplate resolves template variables in a string using the provided context
func resolveTemplate(templateStr string, ctx *EventContext) (string, error) {
        if templateStr == "" {
                return "", nil
        }

        // Create template data structure
        templateData := map[string]interface{}{
                "event": map[string]interface{}{
                        "type":      ctx.Event.Type,
                        "payload":   ctx.Event.Payload,
                        "metadata":  ctx.Event.Metadata,
                        "timestamp": ctx.Event.Timestamp,
                },
                "variables": ctx.Variables,
        }

        // Create template with custom functions
        tmpl, err := template.New("workflow").
                Funcs(templateFunctions()).
                Parse(templateStr)
        if err != nil {
                return "", fmt.Errorf("template parse error: %w", err)
        }

        // Execute template
        var buf bytes.Buffer
        if err := tmpl.Execute(&buf, templateData); err != nil {
                return "", fmt.Errorf("template execution error: %w", err)
        }

        return buf.String(), nil
}

// templateFunctions returns custom template functions
func templateFunctions() template.FuncMap {
        return template.FuncMap{
                "default": func(defaultValue, value interface{}) interface{} {
                        if value == nil || isEmptyValue(reflect.ValueOf(value)) {
                                return defaultValue
                        }
                        return value
                },
                "empty": func(value interface{}) bool {
                        if value == nil {
                                return true
                        }
                        return isEmptyValue(reflect.ValueOf(value))
                },
                "not": func(value bool) bool {
                        return !value
                },
                "eq": func(a, b interface{}) bool {
                        return reflect.DeepEqual(a, b)
                },
                "ne": func(a, b interface{}) bool {
                        return !reflect.DeepEqual(a, b)
                },
                "contains": func(haystack, needle string) bool {
                        return bytes.Contains([]byte(haystack), []byte(needle))
                },
        }
}

// isEmptyValue checks if a reflect.Value is empty
func isEmptyValue(v reflect.Value) bool {
        switch v.Kind() {
        case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
                return v.Len() == 0
        case reflect.Bool:
                return !v.Bool()
        case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
                return v.Int() == 0
        case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
                return v.Uint() == 0
        case reflect.Float32, reflect.Float64:
                return v.Float() == 0
        case reflect.Interface, reflect.Ptr:
                return v.IsNil()
        }
        return false
}
