package web

import (
        "encoding/json"
        "fmt"
        "net/http"
        "time"

        "github.com/gorilla/mux"
)

// NodeType represents different types of workflow nodes
type NodeType string

const (
        NodeTypeTrigger    NodeType = "trigger"
        NodeTypeAction     NodeType = "action"
        NodeTypeCondition  NodeType = "condition"
        NodeTypeParallel   NodeType = "parallel"
        NodeTypeSubflow    NodeType = "subflow"
        NodeTypeDelay      NodeType = "delay"
)

// WorkflowNode represents a node in the visual workflow designer
type WorkflowNode struct {
        ID          string                 `json:"id"`
        Type        NodeType               `json:"type"`
        Name        string                 `json:"name"`
        Description string                 `json:"description"`
        Position    NodePosition           `json:"position"`
        Config      map[string]interface{} `json:"config"`
        Inputs      []NodeConnection       `json:"inputs"`
        Outputs     []NodeConnection       `json:"outputs"`
}

// NodePosition represents the visual position of a node
type NodePosition struct {
        X float64 `json:"x"`
        Y float64 `json:"y"`
}

// NodeConnection represents connections between nodes
type NodeConnection struct {
        NodeID     string `json:"nodeId"`
        OutputPort string `json:"outputPort"`
        InputPort  string `json:"inputPort"`
}

// VisualWorkflow represents a complete visual workflow
type VisualWorkflow struct {
        ID          string         `json:"id"`
        Name        string         `json:"name"`
        Description string         `json:"description"`
        Version     string         `json:"version"`
        Created     time.Time      `json:"created"`
        Modified    time.Time      `json:"modified"`
        Nodes       []WorkflowNode `json:"nodes"`
        Variables   []Variable     `json:"variables"`
        Settings    WorkflowSettings `json:"settings"`
}

// Variable represents workflow variables
type Variable struct {
        Name         string      `json:"name"`
        Type         string      `json:"type"`
        DefaultValue interface{} `json:"defaultValue"`
        Description  string      `json:"description"`
}

// WorkflowSettings contains workflow execution settings
type WorkflowSettings struct {
        RetryCount    int           `json:"retryCount"`
        RetryDelay    time.Duration `json:"retryDelay"`
        Timeout       time.Duration `json:"timeout"`
        Parallel      bool          `json:"parallel"`
        ErrorHandling string        `json:"errorHandling"`
}

// NodeTemplate represents pre-built node templates
type NodeTemplate struct {
        ID          string                 `json:"id"`
        Type        NodeType               `json:"type"`
        Name        string                 `json:"name"`
        Description string                 `json:"description"`
        Category    string                 `json:"category"`
        Icon        string                 `json:"icon"`
        Config      map[string]interface{} `json:"config"`
        Inputs      []PortDefinition       `json:"inputs"`
        Outputs     []PortDefinition       `json:"outputs"`
}

// PortDefinition defines input/output ports for nodes
type PortDefinition struct {
        Name        string `json:"name"`
        Type        string `json:"type"`
        Required    bool   `json:"required"`
        Description string `json:"description"`
}

// DesignerService handles visual workflow designer operations
type DesignerService struct {
        workflows []VisualWorkflow
        templates []NodeTemplate
}

// NewDesignerService creates a new designer service
func NewDesignerService() *DesignerService {
        return &DesignerService{
                workflows: make([]VisualWorkflow, 0),
                templates: getDefaultNodeTemplates(),
        }
}

// RegisterDesignerRoutes registers all designer-related routes
func (ds *DesignerService) RegisterDesignerRoutes(r *mux.Router) {
        // Designer UI routes
        r.HandleFunc("/designer", ds.handleDesignerPage).Methods("GET")
        r.HandleFunc("/designer/{id}", ds.handleDesignerPage).Methods("GET")
        
        // API routes
        api := r.PathPrefix("/api/designer").Subrouter()
        api.HandleFunc("/workflows", ds.handleGetWorkflows).Methods("GET")
        api.HandleFunc("/workflows", ds.handleCreateWorkflow).Methods("POST")
        api.HandleFunc("/workflows/{id}", ds.handleGetWorkflow).Methods("GET")
        api.HandleFunc("/workflows/{id}", ds.handleUpdateWorkflow).Methods("PUT")
        api.HandleFunc("/workflows/{id}", ds.handleDeleteWorkflow).Methods("DELETE")
        api.HandleFunc("/workflows/{id}/export", ds.handleExportWorkflow).Methods("GET")
        api.HandleFunc("/workflows/import", ds.handleImportWorkflow).Methods("POST")
        api.HandleFunc("/templates", ds.handleGetTemplates).Methods("GET")
        api.HandleFunc("/validate", ds.handleValidateWorkflow).Methods("POST")
        api.HandleFunc("/generate-yaml", ds.handleGenerateYAML).Methods("POST")
}

// handleDesignerPage serves the visual workflow designer page
func (ds *DesignerService) handleDesignerPage(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        w.Write([]byte(`<!DOCTYPE html><html><head><title>Reactor Visual Designer</title></head><body><h1>Visual Workflow Designer</h1><p>Advanced drag-and-drop workflow designer interface coming online...</p></body></html>`))
}

// handleGetWorkflows returns all visual workflows
func (ds *DesignerService) handleGetWorkflows(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(ds.workflows)
}

// handleCreateWorkflow creates a new visual workflow
func (ds *DesignerService) handleCreateWorkflow(w http.ResponseWriter, r *http.Request) {
        var workflow VisualWorkflow
        if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }
        
        workflow.ID = generateID()
        workflow.Created = time.Now()
        workflow.Modified = time.Now()
        workflow.Version = "1.0.0"
        
        ds.workflows = append(ds.workflows, workflow)
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(workflow)
}

// handleGetWorkflow returns a specific visual workflow
func (ds *DesignerService) handleGetWorkflow(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        
        for _, workflow := range ds.workflows {
                if workflow.ID == id {
                        w.Header().Set("Content-Type", "application/json")
                        json.NewEncoder(w).Encode(workflow)
                        return
                }
        }
        
        http.Error(w, "Workflow not found", http.StatusNotFound)
}

// handleUpdateWorkflow updates an existing visual workflow
func (ds *DesignerService) handleUpdateWorkflow(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        
        var updatedWorkflow VisualWorkflow
        if err := json.NewDecoder(r.Body).Decode(&updatedWorkflow); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }
        
        for i, workflow := range ds.workflows {
                if workflow.ID == id {
                        updatedWorkflow.ID = id
                        updatedWorkflow.Created = workflow.Created
                        updatedWorkflow.Modified = time.Now()
                        ds.workflows[i] = updatedWorkflow
                        
                        w.Header().Set("Content-Type", "application/json")
                        json.NewEncoder(w).Encode(updatedWorkflow)
                        return
                }
        }
        
        http.Error(w, "Workflow not found", http.StatusNotFound)
}

// handleDeleteWorkflow deletes a visual workflow
func (ds *DesignerService) handleDeleteWorkflow(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        
        for i, workflow := range ds.workflows {
                if workflow.ID == id {
                        ds.workflows = append(ds.workflows[:i], ds.workflows[i+1:]...)
                        w.WriteHeader(http.StatusNoContent)
                        return
                }
        }
        
        http.Error(w, "Workflow not found", http.StatusNotFound)
}

// handleGetTemplates returns available node templates
func (ds *DesignerService) handleGetTemplates(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(ds.templates)
}

// handleValidateWorkflow validates a visual workflow
func (ds *DesignerService) handleValidateWorkflow(w http.ResponseWriter, r *http.Request) {
        var workflow VisualWorkflow
        if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }
        
        errors := ds.validateWorkflow(workflow)
        
        response := map[string]interface{}{
                "valid":  len(errors) == 0,
                "errors": errors,
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
}

// handleGenerateYAML generates YAML from visual workflow
func (ds *DesignerService) handleGenerateYAML(w http.ResponseWriter, r *http.Request) {
        var workflow VisualWorkflow
        if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }
        
        yaml, err := ds.generateYAMLFromWorkflow(workflow)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        
        response := map[string]string{
                "yaml": yaml,
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
}

// handleExportWorkflow exports workflow as JSON
func (ds *DesignerService) handleExportWorkflow(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        
        for _, workflow := range ds.workflows {
                if workflow.ID == id {
                        w.Header().Set("Content-Type", "application/json")
                        w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.json\"", workflow.Name))
                        json.NewEncoder(w).Encode(workflow)
                        return
                }
        }
        
        http.Error(w, "Workflow not found", http.StatusNotFound)
}

// handleImportWorkflow imports workflow from JSON
func (ds *DesignerService) handleImportWorkflow(w http.ResponseWriter, r *http.Request) {
        var workflow VisualWorkflow
        if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }
        
        workflow.ID = generateID()
        workflow.Created = time.Now()
        workflow.Modified = time.Now()
        
        ds.workflows = append(ds.workflows, workflow)
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(workflow)
}

// validateWorkflow validates a visual workflow structure
func (ds *DesignerService) validateWorkflow(workflow VisualWorkflow) []string {
        var errors []string
        
        if workflow.Name == "" {
                errors = append(errors, "Workflow name is required")
        }
        
        if len(workflow.Nodes) == 0 {
                errors = append(errors, "Workflow must have at least one node")
        }
        
        // Check for trigger nodes
        hasTrigger := false
        for _, node := range workflow.Nodes {
                if node.Type == NodeTypeTrigger {
                        hasTrigger = true
                        break
                }
        }
        
        if !hasTrigger {
                errors = append(errors, "Workflow must have at least one trigger node")
        }
        
        // Validate node connections
        nodeIDs := make(map[string]bool)
        for _, node := range workflow.Nodes {
                nodeIDs[node.ID] = true
        }
        
        for _, node := range workflow.Nodes {
                for _, output := range node.Outputs {
                        if !nodeIDs[output.NodeID] {
                                errors = append(errors, fmt.Sprintf("Node %s references non-existent node %s", node.ID, output.NodeID))
                        }
                }
        }
        
        return errors
}

// generateYAMLFromWorkflow converts visual workflow to YAML
func (ds *DesignerService) generateYAMLFromWorkflow(workflow VisualWorkflow) (string, error) {
        yaml := fmt.Sprintf(`name: %s
description: %s
version: %s

`, workflow.Name, workflow.Description, workflow.Version)
        
        // Add triggers
        yaml += "triggers:\n"
        for _, node := range workflow.Nodes {
                if node.Type == NodeTypeTrigger {
                        yaml += fmt.Sprintf("  - type: %s\n", node.Config["type"])
                        if config, ok := node.Config["config"].(map[string]interface{}); ok {
                                for key, value := range config {
                                        yaml += fmt.Sprintf("    %s: %v\n", key, value)
                                }
                        }
                }
        }
        
        // Add steps
        yaml += "\nsteps:\n"
        for _, node := range workflow.Nodes {
                if node.Type == NodeTypeAction {
                        yaml += fmt.Sprintf("  - name: %s\n", node.Name)
                        yaml += fmt.Sprintf("    action: %s\n", node.Config["action"])
                        if config, ok := node.Config["config"].(map[string]interface{}); ok {
                                yaml += "    config:\n"
                                for key, value := range config {
                                        yaml += fmt.Sprintf("      %s: %v\n", key, value)
                                }
                        }
                }
        }
        
        return yaml, nil
}

// getDefaultNodeTemplates returns predefined node templates
func getDefaultNodeTemplates() []NodeTemplate {
        return []NodeTemplate{
                {
                        ID:          "http-trigger",
                        Type:        NodeTypeTrigger,
                        Name:        "HTTP Trigger",
                        Description: "Trigger workflow on HTTP request",
                        Category:    "Triggers",
                        Icon:        "🌐",
                        Config: map[string]interface{}{
                                "type": "http",
                                "config": map[string]interface{}{
                                        "port": 8080,
                                        "path": "/webhook",
                                },
                        },
                        Outputs: []PortDefinition{{Name: "request", Type: "object", Description: "HTTP request data"}},
                },
                {
                        ID:          "file-trigger",
                        Type:        NodeTypeTrigger,
                        Name:        "File Trigger",
                        Description: "Trigger workflow on file changes",
                        Category:    "Triggers",
                        Icon:        "📁",
                        Config: map[string]interface{}{
                                "type": "file",
                                "config": map[string]interface{}{
                                        "path":   "./watch",
                                        "events": []string{"create", "modify"},
                                },
                        },
                        Outputs: []PortDefinition{{Name: "file", Type: "object", Description: "File event data"}},
                },
                {
                        ID:          "redis-trigger",
                        Type:        NodeTypeTrigger,
                        Name:        "Redis Trigger",
                        Description: "Trigger workflow on Redis events",
                        Category:    "Triggers",
                        Icon:        "🔴",
                        Config: map[string]interface{}{
                                "type": "redis",
                                "config": map[string]interface{}{
                                        "host":    "localhost:6379",
                                        "channel": "events",
                                },
                        },
                        Outputs: []PortDefinition{{Name: "message", Type: "object", Description: "Redis message data"}},
                },
                {
                        ID:          "http-action",
                        Type:        NodeTypeAction,
                        Name:        "HTTP Request",
                        Description: "Make HTTP request",
                        Category:    "Actions",
                        Icon:        "🚀",
                        Config: map[string]interface{}{
                                "action": "http",
                                "config": map[string]interface{}{
                                        "url":    "https://api.example.com",
                                        "method": "POST",
                                },
                        },
                        Inputs:  []PortDefinition{{Name: "data", Type: "object", Description: "Request data"}},
                        Outputs: []PortDefinition{{Name: "response", Type: "object", Description: "HTTP response"}},
                },
                {
                        ID:          "email-action",
                        Type:        NodeTypeAction,
                        Name:        "Send Email",
                        Description: "Send email notification",
                        Category:    "Actions",
                        Icon:        "📧",
                        Config: map[string]interface{}{
                                "action": "email",
                                "config": map[string]interface{}{
                                        "smtp_host": "smtp.gmail.com",
                                        "smtp_port": 587,
                                },
                        },
                        Inputs: []PortDefinition{{Name: "message", Type: "object", Description: "Email content"}},
                },
                {
                        ID:          "condition",
                        Type:        NodeTypeCondition,
                        Name:        "Condition",
                        Description: "Conditional branching",
                        Category:    "Logic",
                        Icon:        "🔀",
                        Config: map[string]interface{}{
                                "condition": "{{.data.status}} == 'success'",
                        },
                        Inputs:  []PortDefinition{{Name: "data", Type: "object", Description: "Input data"}},
                        Outputs: []PortDefinition{
                                {Name: "true", Type: "object", Description: "True branch"},
                                {Name: "false", Type: "object", Description: "False branch"},
                        },
                },
                {
                        ID:          "parallel",
                        Type:        NodeTypeParallel,
                        Name:        "Parallel Execution",
                        Description: "Execute multiple branches in parallel",
                        Category:    "Flow Control",
                        Icon:        "⚡",
                        Config:      map[string]interface{}{},
                        Inputs:      []PortDefinition{{Name: "data", Type: "object", Description: "Input data"}},
                        Outputs:     []PortDefinition{{Name: "results", Type: "array", Description: "Parallel results"}},
                },
                {
                        ID:          "delay",
                        Type:        NodeTypeDelay,
                        Name:        "Delay",
                        Description: "Add delay to workflow",
                        Category:    "Flow Control",
                        Icon:        "⏱️",
                        Config: map[string]interface{}{
                                "duration": "5s",
                        },
                        Inputs:  []PortDefinition{{Name: "data", Type: "object", Description: "Input data"}},
                        Outputs: []PortDefinition{{Name: "data", Type: "object", Description: "Output data"}},
                },
        }
}

// generateID generates a unique ID
func generateID() string {
        return fmt.Sprintf("wf_%d", time.Now().UnixNano())
}