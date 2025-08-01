package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// AIWorkflowBuilder provides intelligent workflow creation capabilities
type AIWorkflowBuilder struct {
	logger           *zap.Logger
	nlpProcessor     *NLPProcessor
	patternLibrary   *PatternLibrary
	suggestionEngine *SuggestionEngine
	validationEngine *ValidationEngine
}

// NLPProcessor handles natural language workflow descriptions
type NLPProcessor struct {
	triggerPatterns map[string]*regexp.Regexp
	actionPatterns  map[string]*regexp.Regexp
	commonPhrases   map[string]string
}

// PatternLibrary stores common workflow patterns and templates
type PatternLibrary struct {
	Patterns map[string]WorkflowPattern `json:"patterns"`
}

// WorkflowPattern represents a reusable workflow template
type WorkflowPattern struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Triggers    []string               `json:"triggers"`
	Actions     []string               `json:"actions"`
	Template    map[string]interface{} `json:"template"`
	Usage       int                    `json:"usage"`
	Tags        []string               `json:"tags"`
}

// SuggestionEngine provides intelligent workflow suggestions
type SuggestionEngine struct {
	contextHistory []string
	userPatterns   map[string]int
}

// ValidationEngine checks workflow logic and provides improvement suggestions
type ValidationEngine struct {
	rules []ValidationRule
}

// ValidationRule defines a workflow validation constraint
type ValidationRule struct {
	Name        string                                     `json:"name"`
	Description string                                     `json:"description"`
	Check       func(workflow map[string]interface{}) bool `json:"-"`
	Suggestion  string                                     `json:"suggestion"`
}

// WorkflowRequest represents a natural language workflow creation request
type WorkflowRequest struct {
	Description string                 `json:"description"`
	Context     map[string]interface{} `json:"context,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
}

// WorkflowSuggestion represents an AI-generated workflow suggestion
type WorkflowSuggestion struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Workflow     map[string]interface{} `json:"workflow"`
	Confidence   float64                `json:"confidence"`
	Steps        []SuggestedStep        `json:"steps"`
	Alternatives []WorkflowSuggestion   `json:"alternatives,omitempty"`
}

// SuggestedStep represents a workflow step with AI suggestions
type SuggestedStep struct {
	Type         string                 `json:"type"`
	Action       string                 `json:"action"`
	Parameters   map[string]interface{} `json:"parameters"`
	Description  string                 `json:"description"`
	Confidence   float64                `json:"confidence"`
	Alternatives []string               `json:"alternatives,omitempty"`
}

// NewAIWorkflowBuilder creates a new AI workflow builder
func NewAIWorkflowBuilder(logger *zap.Logger) *AIWorkflowBuilder {
	builder := &AIWorkflowBuilder{
		logger:         logger,
		nlpProcessor:   initializeNLPProcessor(),
		patternLibrary: initializePatternLibrary(),
		suggestionEngine: &SuggestionEngine{
			contextHistory: make([]string, 0),
			userPatterns:   make(map[string]int),
		},
		validationEngine: initializeValidationEngine(),
	}

	return builder
}

// RegisterRoutes sets up the AI workflow builder API endpoints
func (ai *AIWorkflowBuilder) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/ai").Subrouter()

	// AI builder endpoints
	api.HandleFunc("/builder", ai.handleBuilderPage).Methods("GET")
	api.HandleFunc("/api/generate", ai.handleGenerateWorkflow).Methods("POST")
	api.HandleFunc("/api/suggest", ai.handleSuggestSteps).Methods("POST")
	api.HandleFunc("/api/validate", ai.handleValidateWorkflow).Methods("POST")
	api.HandleFunc("/api/patterns", ai.handleGetPatterns).Methods("GET")
	api.HandleFunc("/api/autocomplete", ai.handleAutoComplete).Methods("POST")
}

// handleBuilderPage serves the AI workflow builder interface
func (ai *AIWorkflowBuilder) handleBuilderPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(getAIBuilderHTML()))
}

// handleGenerateWorkflow creates a workflow from natural language description
func (ai *AIWorkflowBuilder) handleGenerateWorkflow(w http.ResponseWriter, r *http.Request) {
	var req WorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ai.logger.Info("Generating workflow from description", zap.String("description", req.Description))

	suggestion := ai.generateWorkflowFromNL(req.Description)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestion)
}

// handleSuggestSteps provides intelligent step suggestions
func (ai *AIWorkflowBuilder) handleSuggestSteps(w http.ResponseWriter, r *http.Request) {
	var context map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&context); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	suggestions := ai.suggestNextSteps(context)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

// handleValidateWorkflow validates workflow logic and provides suggestions
func (ai *AIWorkflowBuilder) handleValidateWorkflow(w http.ResponseWriter, r *http.Request) {
	var workflow map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	validation := ai.validateWorkflow(workflow)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(validation)
}

// handleGetPatterns returns available workflow patterns
func (ai *AIWorkflowBuilder) handleGetPatterns(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	patterns := ai.getPatterns(category)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(patterns)
}

// handleAutoComplete provides intelligent auto-completion
func (ai *AIWorkflowBuilder) handleAutoComplete(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Partial string `json:"partial"`
		Context string `json:"context"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	completions := ai.getAutoCompletions(req.Partial, req.Context)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completions)
}

// generateWorkflowFromNL converts natural language to workflow
func (ai *AIWorkflowBuilder) generateWorkflowFromNL(description string) WorkflowSuggestion {
	ai.logger.Info("Processing natural language description", zap.String("description", description))

	// Parse triggers from description
	triggers := ai.nlpProcessor.extractTriggers(description)
	actions := ai.nlpProcessor.extractActions(description)

	// Find matching patterns
	patterns := ai.findMatchingPatterns(triggers, actions)

	// Generate workflow structure
	workflow := ai.buildWorkflowFromPatterns(patterns, description)

	return WorkflowSuggestion{
		ID:          generateSuggestionID(),
		Name:        ai.generateWorkflowName(description),
		Description: description,
		Workflow:    workflow,
		Confidence:  ai.calculateConfidence(triggers, actions, patterns),
		Steps:       ai.generateSteps(workflow),
	}
}

// extractTriggers identifies trigger patterns in natural language
func (nlp *NLPProcessor) extractTriggers(text string) []string {
	var triggers []string
	text = strings.ToLower(text)

	for trigger, pattern := range nlp.triggerPatterns {
		if pattern.MatchString(text) {
			triggers = append(triggers, trigger)
		}
	}

	return triggers
}

// extractActions identifies action patterns in natural language
func (nlp *NLPProcessor) extractActions(text string) []string {
	var actions []string
	text = strings.ToLower(text)

	for action, pattern := range nlp.actionPatterns {
		if pattern.MatchString(text) {
			actions = append(actions, action)
		}
	}

	return actions
}

// findMatchingPatterns finds relevant workflow patterns
func (ai *AIWorkflowBuilder) findMatchingPatterns(triggers, actions []string) []WorkflowPattern {
	var matches []WorkflowPattern

	for _, pattern := range ai.patternLibrary.Patterns {
		score := ai.calculatePatternMatch(pattern, triggers, actions)
		if score > 0.3 { // 30% confidence threshold
			matches = append(matches, pattern)
		}
	}

	return matches
}

// calculatePatternMatch computes similarity between input and pattern
func (ai *AIWorkflowBuilder) calculatePatternMatch(pattern WorkflowPattern, triggers, actions []string) float64 {
	triggerMatch := ai.calculateArrayOverlap(pattern.Triggers, triggers)
	actionMatch := ai.calculateArrayOverlap(pattern.Actions, actions)

	return (triggerMatch + actionMatch) / 2.0
}

// calculateArrayOverlap computes overlap percentage between two string arrays
func (ai *AIWorkflowBuilder) calculateArrayOverlap(arr1, arr2 []string) float64 {
	if len(arr1) == 0 || len(arr2) == 0 {
		return 0.0
	}

	matches := 0
	for _, item1 := range arr1 {
		for _, item2 := range arr2 {
			if strings.Contains(strings.ToLower(item1), strings.ToLower(item2)) ||
				strings.Contains(strings.ToLower(item2), strings.ToLower(item1)) {
				matches++
				break
			}
		}
	}

	return float64(matches) / float64(len(arr1))
}

// buildWorkflowFromPatterns constructs workflow YAML from patterns
func (ai *AIWorkflowBuilder) buildWorkflowFromPatterns(patterns []WorkflowPattern, description string) map[string]interface{} {
	if len(patterns) == 0 {
		return ai.buildGenericWorkflow(description)
	}

	// Use the best matching pattern as base
	bestPattern := patterns[0]
	workflow := make(map[string]interface{})

	// Copy pattern template
	for key, value := range bestPattern.Template {
		workflow[key] = value
	}

	// Customize based on description
	workflow["name"] = ai.generateWorkflowName(description)
	workflow["description"] = description

	return workflow
}

// buildGenericWorkflow creates a basic workflow structure
func (ai *AIWorkflowBuilder) buildGenericWorkflow(description string) map[string]interface{} {
	return map[string]interface{}{
		"name":        ai.generateWorkflowName(description),
		"description": description,
		"version":     "1.0",
		"triggers": []map[string]interface{}{
			{
				"type": "http",
				"path": "/webhook",
			},
		},
		"steps": []map[string]interface{}{
			{
				"name":   "process",
				"action": "log.info",
				"params": map[string]string{
					"message": "Processing workflow: " + description,
				},
			},
		},
	}
}

// generateWorkflowName creates a descriptive workflow name
func (ai *AIWorkflowBuilder) generateWorkflowName(description string) string {
	words := strings.Fields(strings.ToLower(description))
	if len(words) > 4 {
		words = words[:4]
	}
	return strings.Join(words, "-") + "-workflow"
}

// generateSteps converts workflow to suggested steps
func (ai *AIWorkflowBuilder) generateSteps(workflow map[string]interface{}) []SuggestedStep {
	var steps []SuggestedStep

	if stepsData, exists := workflow["steps"]; exists {
		if stepsList, ok := stepsData.([]map[string]interface{}); ok {
			for _, step := range stepsList {
				suggestedStep := SuggestedStep{
					Type:        "action",
					Description: fmt.Sprintf("%v", step["name"]),
					Parameters:  step,
					Confidence:  0.8,
				}
				steps = append(steps, suggestedStep)
			}
		}
	}

	return steps
}

// initializeNLPProcessor sets up natural language processing patterns
func initializeNLPProcessor() *NLPProcessor {
	return &NLPProcessor{
		triggerPatterns: map[string]*regexp.Regexp{
			"http.request":    regexp.MustCompile(`(?i)(http|api|webhook|request|endpoint|url)`),
			"file.created":    regexp.MustCompile(`(?i)(file|upload|document|created|added|new file)`),
			"file.modified":   regexp.MustCompile(`(?i)(file.*modified|file.*changed|file.*updated)`),
			"file.deleted":    regexp.MustCompile(`(?i)(file.*deleted|file.*removed|file.*trashed)`),
			"database.change": regexp.MustCompile(`(?i)(database|db|record|table|insert|update|delete)`),
			"user.created":    regexp.MustCompile(`(?i)(user|customer|member|account|signup|register|created)`),
			"user.updated":    regexp.MustCompile(`(?i)(user.*updated|profile.*changed|account.*modified)`),
			"order.created":   regexp.MustCompile(`(?i)(order|purchase|transaction|payment|checkout|buy)`),
			"email.received":  regexp.MustCompile(`(?i)(email|mail|message|inbox|received)`),
			"scheduled":       regexp.MustCompile(`(?i)(schedule|time|daily|weekly|monthly|cron|periodic)`),
			"error.occurred":  regexp.MustCompile(`(?i)(error|exception|failure|crash|alert|issue)`),
			"system.event":    regexp.MustCompile(`(?i)(system|server|service|application|app)`),
		},
		actionPatterns: map[string]*regexp.Regexp{
			"email.send":      regexp.MustCompile(`(?i)(email|mail|send|notify|alert|message)`),
			"http.request":    regexp.MustCompile(`(?i)(http|api|call|request|fetch|get|post|put|delete)`),
			"database.insert": regexp.MustCompile(`(?i)(database|db|insert|save|store|create.*record)`),
			"database.update": regexp.MustCompile(`(?i)(database.*update|db.*update|modify.*record|change.*data)`),
			"file.process":    regexp.MustCompile(`(?i)(file|process|transform|convert|parse|read|write)`),
			"file.move":       regexp.MustCompile(`(?i)(move.*file|copy.*file|transfer.*file|backup)`),
			"notification":    regexp.MustCompile(`(?i)(notification|push|sms|slack|discord|teams|alert)`),
			"log.info":        regexp.MustCompile(`(?i)(log|record|track|audit|monitor|trace)`),
			"shell.exec":      regexp.MustCompile(`(?i)(shell|command|script|execute|run|bash|cmd)`),
			"transform.data":  regexp.MustCompile(`(?i)(transform|convert|format|parse|extract|process.*data)`),
			"validate.input":  regexp.MustCompile(`(?i)(validate|check|verify|test|ensure|confirm)`),
			"wait.delay":      regexp.MustCompile(`(?i)(wait|delay|sleep|pause|timeout|retry)`),
			"condition.if":    regexp.MustCompile(`(?i)(if|condition|check.*if|when.*then|unless)`),
		},
		commonPhrases: map[string]string{
			"when":    "trigger",
			"if":      "condition",
			"then":    "action",
			"send":    "email.send",
			"create":  "database.insert",
			"update":  "database.update",
			"delete":  "database.delete",
			"process": "file.process",
			"notify":  "notification",
			"log":     "log.info",
			"wait":    "wait.delay",
		},
	}
}

// initializePatternLibrary loads common workflow patterns
func initializePatternLibrary() *PatternLibrary {
	return &PatternLibrary{
		Patterns: map[string]WorkflowPattern{
			"user-onboarding": {
				ID:          "user-onboarding",
				Name:        "User Onboarding",
				Description: "Handle new user registration and onboarding",
				Category:    "User Management",
				Triggers:    []string{"user.created", "http.request"},
				Actions:     []string{"email.send", "database.insert", "log.info"},
				Template: map[string]interface{}{
					"name": "user-onboarding-workflow",
					"on": map[string]interface{}{
						"event": "user.created",
					},
					"workflow": []map[string]interface{}{
						{"name": "send-welcome", "action": "email.send", "message": "Welcome {{ .event.payload.name }}!"},
						{"name": "create-profile", "action": "database.insert", "table": "users"},
						{"name": "log-registration", "action": "log.info", "message": "New user registered"},
					},
				},
				Usage: 45,
				Tags:  []string{"user", "onboarding", "email", "registration"},
			},
			"file-processing": {
				ID:          "file-processing",
				Name:        "File Processing",
				Description: "Process uploaded files and perform operations",
				Category:    "File Management",
				Triggers:    []string{"file.created", "file.modified"},
				Actions:     []string{"file.process", "validate.input", "transform.data"},
				Template: map[string]interface{}{
					"name": "file-processing-workflow",
					"on": map[string]interface{}{
						"event": "file.created",
					},
					"workflow": []map[string]interface{}{
						{"name": "validate-file", "action": "validate.input", "type": "file"},
						{"name": "process-content", "action": "file.process", "operation": "parse"},
						{"name": "transform-data", "action": "transform.data", "format": "json"},
						{"name": "store-result", "action": "database.insert", "table": "processed_files"},
					},
				},
				Usage: 32,
				Tags:  []string{"file", "upload", "processing", "validation"},
			},
			"notification-system": {
				ID:          "notification-system",
				Name:        "Notification System",
				Description: "Send alerts and notifications based on events",
				Category:    "Communication",
				Triggers:    []string{"error.occurred", "system.event", "user.created"},
				Actions:     []string{"notification", "email.send", "log.info"},
				Template: map[string]interface{}{
					"name": "notification-workflow",
					"on": map[string]interface{}{
						"event": "error.occurred",
					},
					"workflow": []map[string]interface{}{
						{"name": "log-error", "action": "log.info", "level": "error"},
						{"name": "send-alert", "action": "notification", "channel": "slack"},
						{"name": "email-admin", "action": "email.send", "to": "admin@company.com"},
					},
				},
				Usage: 28,
				Tags:  []string{"notification", "alert", "error", "monitoring"},
			},
			"ecommerce-order": {
				ID:          "ecommerce-order",
				Name:        "E-commerce Order Processing",
				Description: "Handle order processing and fulfillment",
				Category:    "E-commerce",
				Triggers:    []string{"order.created", "payment.completed"},
				Actions:     []string{"database.update", "email.send", "notification"},
				Template: map[string]interface{}{
					"name": "order-processing-workflow",
					"on": map[string]interface{}{
						"event": "order.created",
					},
					"workflow": []map[string]interface{}{
						{"name": "validate-order", "action": "validate.input", "type": "order"},
						{"name": "process-payment", "action": "http.request", "url": "payment-api"},
						{"name": "update-inventory", "action": "database.update", "table": "inventory"},
						{"name": "send-confirmation", "action": "email.send", "template": "order-confirmation"},
						{"name": "notify-fulfillment", "action": "notification", "channel": "fulfillment-team"},
					},
				},
				Usage: 38,
				Tags:  []string{"ecommerce", "order", "payment", "fulfillment"},
			},
			"data-sync": {
				ID:          "data-sync",
				Name:        "Data Synchronization",
				Description: "Sync data between systems and databases",
				Category:    "Data Management",
				Triggers:    []string{"scheduled", "database.change"},
				Actions:     []string{"http.request", "database.insert", "transform.data"},
				Template: map[string]interface{}{
					"name": "data-sync-workflow",
					"on": map[string]interface{}{
						"event": "scheduled",
						"cron":  "0 */6 * * *",
					},
					"workflow": []map[string]interface{}{
						{"name": "fetch-source-data", "action": "http.request", "url": "{{ .source_api }}"},
						{"name": "transform-data", "action": "transform.data", "format": "json"},
						{"name": "sync-to-target", "action": "database.insert", "table": "synced_data"},
						{"name": "log-sync-status", "action": "log.info", "message": "Data sync completed"},
					},
				},
				Usage: 25,
				Tags:  []string{"data", "sync", "etl", "integration"},
			},
			"api-monitor": {
				ID:          "api-monitor",
				Name:        "API Health Monitor",
				Description: "Monitor API endpoints and alert on failures",
				Category:    "Monitoring",
				Triggers:    []string{"scheduled", "http.request"},
				Actions:     []string{"http.request", "notification", "log.info"},
				Template: map[string]interface{}{
					"name": "api-monitor-workflow",
					"on": map[string]interface{}{
						"event": "scheduled",
						"cron":  "*/5 * * * *",
					},
					"workflow": []map[string]interface{}{
						{"name": "check-api-health", "action": "http.request", "url": "{{ .api_endpoint }}/health"},
						{"name": "validate-response", "action": "validate.input", "type": "api_response"},
						{"name": "alert-if-down", "action": "notification", "channel": "ops-team", "if": "{{ .response.status != 200 }}"},
						{"name": "log-status", "action": "log.info", "message": "API health check completed"},
					},
				},
				Usage: 22,
				Tags:  []string{"monitoring", "api", "health", "alert"},
			},
		},
	}
}

// initializeValidationEngine sets up workflow validation rules
func initializeValidationEngine() *ValidationEngine {
	return &ValidationEngine{
		rules: []ValidationRule{
			{
				Name:        "has-trigger",
				Description: "Workflow must have at least one trigger",
				Check: func(workflow map[string]interface{}) bool {
					triggers, exists := workflow["triggers"]
					if !exists {
						return false
					}
					if triggersList, ok := triggers.([]interface{}); ok {
						return len(triggersList) > 0
					}
					return false
				},
				Suggestion: "Add at least one trigger to start the workflow",
			},
			{
				Name:        "has-steps",
				Description: "Workflow must have at least one step",
				Check: func(workflow map[string]interface{}) bool {
					steps, exists := workflow["steps"]
					if !exists {
						return false
					}
					if stepsList, ok := steps.([]interface{}); ok {
						return len(stepsList) > 0
					}
					return false
				},
				Suggestion: "Add at least one action step to the workflow",
			},
		},
	}
}

// Helper functions
func generateSuggestionID() string {
	return fmt.Sprintf("suggestion-%d", time.Now().UnixNano())
}

func (ai *AIWorkflowBuilder) calculateConfidence(triggers, actions []string, patterns []WorkflowPattern) float64 {
	if len(patterns) == 0 {
		return 0.3 // Low confidence for generic workflows
	}
	return 0.8 // High confidence for pattern-matched workflows
}

func (ai *AIWorkflowBuilder) suggestNextSteps(context map[string]interface{}) []SuggestedStep {
	// Implementation for suggesting next workflow steps
	return []SuggestedStep{}
}

func (ai *AIWorkflowBuilder) validateWorkflow(workflow map[string]interface{}) map[string]interface{} {
	results := make(map[string]interface{})
	var issues []string
	var suggestions []string

	for _, rule := range ai.validationEngine.rules {
		if !rule.Check(workflow) {
			issues = append(issues, rule.Description)
			suggestions = append(suggestions, rule.Suggestion)
		}
	}

	results["valid"] = len(issues) == 0
	results["issues"] = issues
	results["suggestions"] = suggestions

	return results
}

func (ai *AIWorkflowBuilder) getPatterns(category string) []WorkflowPattern {
	var patterns []WorkflowPattern

	for _, pattern := range ai.patternLibrary.Patterns {
		if category == "" || pattern.Category == category {
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

func (ai *AIWorkflowBuilder) getAutoCompletions(partial, context string) []string {
	// Implementation for auto-completion suggestions
	return []string{
		"send email notification",
		"update database record",
		"make HTTP request",
		"log execution details",
		"transform data format",
	}
}

// getAIBuilderHTML returns the AI workflow builder interface
func getAIBuilderHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI Workflow Builder - Reactor</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #0a0e1a; color: #e2e8f0; }
        .builder-container { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; padding: 20px; min-height: 100vh; }
        .input-panel, .output-panel { background: linear-gradient(135deg, #1e293b 0%, #334155 100%); border-radius: 12px; padding: 20px; border: 1px solid #475569; }
        .input-panel h2, .output-panel h2 { color: #60a5fa; margin-bottom: 20px; }
        .description-input { width: 100%; height: 150px; background: #1e293b; border: 1px solid #475569; border-radius: 8px; padding: 15px; color: #e2e8f0; font-family: inherit; resize: vertical; }
        .description-input:focus { outline: none; border-color: #60a5fa; }
        .generate-btn { background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%); color: white; border: none; padding: 12px 24px; border-radius: 8px; font-weight: bold; cursor: pointer; margin-top: 15px; }
        .generate-btn:hover { transform: translateY(-2px); box-shadow: 0 5px 15px rgba(59, 130, 246, 0.4); }
        .workflow-output { background: #1e293b; border: 1px solid #475569; border-radius: 8px; padding: 15px; min-height: 300px; font-family: monospace; white-space: pre-wrap; }
        .suggestions-list { margin-top: 20px; }
        .suggestion-item { background: #334155; padding: 10px; margin: 5px 0; border-radius: 6px; cursor: pointer; }
        .suggestion-item:hover { background: #475569; }
        .confidence-bar { width: 100%; height: 4px; background: #374151; border-radius: 2px; margin-top: 5px; }
        .confidence-fill { height: 100%; background: linear-gradient(90deg, #ef4444 0%, #f59e0b 50%, #10b981 100%); border-radius: 2px; }
        .ai-indicator { color: #10b981; font-size: 0.8em; margin-bottom: 10px; }
        .pattern-suggestions { margin-top: 20px; }
        .pattern-card { background: #374151; padding: 15px; margin: 10px 0; border-radius: 8px; border-left: 4px solid #60a5fa; }
        .pattern-card h4 { color: #60a5fa; margin-bottom: 5px; }
        .pattern-card p { color: #94a3b8; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="builder-container">
        <div class="input-panel">
            <h2>🤖 AI Workflow Builder</h2>
            <div class="ai-indicator">● AI Assistant Active</div>
            
            <label for="description">Describe your workflow in plain English:</label>
            <textarea id="description" class="description-input" 
                      placeholder="Example: When a customer signs up, send them a welcome email and create their profile in the database"></textarea>
            
            <button class="generate-btn" onclick="generateWorkflow()">✨ Generate Workflow</button>
            
            <div class="pattern-suggestions">
                <h3>💡 Popular Patterns</h3>
                <div class="pattern-card" onclick="usePattern('user-onboarding')">
                    <h4>User Onboarding</h4>
                    <p>Handle new user registration and welcome process</p>
                </div>
                <div class="pattern-card" onclick="usePattern('file-processing')">
                    <h4>File Processing</h4>
                    <p>Process uploaded files and perform operations</p>
                </div>
                <div class="pattern-card" onclick="usePattern('notification')">
                    <h4>Notification System</h4>
                    <p>Send alerts and notifications based on events</p>
                </div>
            </div>
        </div>
        
        <div class="output-panel">
            <h2>📋 Generated Workflow</h2>
            <div id="workflow-output" class="workflow-output">
                <em>Your AI-generated workflow will appear here...</em>
            </div>
            
            <div class="suggestions-list" id="suggestions-list" style="display: none;">
                <h3>🎯 AI Suggestions</h3>
                <div id="suggestions-content"></div>
            </div>
        </div>
    </div>
    
    <script>
        async function generateWorkflow() {
            const description = document.getElementById('description').value;
            if (!description.trim()) {
                alert('Please enter a workflow description');
                return;
            }
            
            document.getElementById('workflow-output').innerHTML = '<em>🤖 AI is generating your workflow...</em>';
            
            try {
                const response = await fetch('/ai/api/generate', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ description: description })
                });
                
                const result = await response.json();
                displayWorkflow(result);
            } catch (error) {
                document.getElementById('workflow-output').innerHTML = 
                    '<em style="color: #ef4444;">Error generating workflow. Please try again.</em>';
            }
        }
        
        function displayWorkflow(result) {
            const output = document.getElementById('workflow-output');
            const workflowYaml = formatWorkflowAsYAML(result.workflow);
            output.innerHTML = workflowYaml;
            
            // Show AI suggestions
            const suggestionsDiv = document.getElementById('suggestions-list');
            const suggestionsContent = document.getElementById('suggestions-content');
            
            let suggestionsHtml = '';
            suggestionsHtml += '<div class="suggestion-item">';
            suggestionsHtml += '<strong>' + result.name + '</strong><br>';
            suggestionsHtml += '<small>Confidence: ' + Math.round(result.confidence * 100) + '%</small>';
            suggestionsHtml += '<div class="confidence-bar"><div class="confidence-fill" style="width: ' + (result.confidence * 100) + '%"></div></div>';
            suggestionsHtml += '</div>';
            
            suggestionsContent.innerHTML = suggestionsHtml;
            suggestionsDiv.style.display = 'block';
        }
        
        function formatWorkflowAsYAML(workflow) {
            return JSON.stringify(workflow, null, 2)
                .replace(/"/g, '')
                .replace(/,/g, '')
                .replace(/\{/g, '')
                .replace(/\}/g, '')
                .replace(/\[/g, '')
                .replace(/\]/g, '');
        }
        
        function usePattern(patternId) {
            const patterns = {
                'user-onboarding': 'When a user registers, send welcome email and create their profile',
                'file-processing': 'When a file is uploaded, validate it and process the content',
                'notification': 'When an event occurs, send notifications to relevant users'
            };
            
            document.getElementById('description').value = patterns[patternId];
        }
        
        // Auto-complete functionality
        document.getElementById('description').addEventListener('input', function(e) {
            // Implementation for real-time suggestions
        });
    </script>
</body>
</html>`
}
