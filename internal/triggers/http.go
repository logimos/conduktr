package triggers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/logimos/conduktr/internal/engine"
	"github.com/logimos/conduktr/internal/persistence"
	"github.com/logimos/conduktr/internal/web"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// HTTPTrigger handles HTTP-based event triggers
type HTTPTrigger struct {
	logger *zap.Logger
	engine *engine.Engine
	server *http.Server
	port   int
	router *mux.Router
}

// EventPayload represents the payload of an HTTP event
type EventPayload struct {
	Event string                 `json:"event"`
	Data  map[string]interface{} `json:"data"`
}

// NewHTTPTrigger creates a new HTTP trigger
func NewHTTPTrigger(logger *zap.Logger, engine *engine.Engine, port int) *HTTPTrigger {
	return &HTTPTrigger{
		logger: logger,
		engine: engine,
		port:   port,
		router: mux.NewRouter(),
	}
}

// RegisterAdvancedRoutes registers routes for advanced services
func (h *HTTPTrigger) RegisterAdvancedRoutes(analyticsDashboard interface{}, aiBuilder interface{}, integrationHub interface{}) {
	// Register analytics dashboard routes if it has RegisterRoutes method
	if analytics, ok := analyticsDashboard.(interface{ RegisterRoutes(*mux.Router) }); ok {
		analytics.RegisterRoutes(h.router)
		h.logger.Info("Registered analytics dashboard routes")
	}

	// Register AI builder routes if it has RegisterRoutes method
	if ai, ok := aiBuilder.(interface{ RegisterRoutes(*mux.Router) }); ok {
		ai.RegisterRoutes(h.router)
		h.logger.Info("Registered AI builder routes")
	}

	// Register integration hub routes if it has RegisterRoutes method
	if hub, ok := integrationHub.(interface{ RegisterRoutes(*mux.Router) }); ok {
		hub.RegisterRoutes(h.router)
		h.logger.Info("Registered integration hub routes")
	}
}

// RegisterMarketplaceRoutes registers marketplace routes
func (h *HTTPTrigger) RegisterMarketplaceRoutes(marketplaceService interface{}) {
	if marketplace, ok := marketplaceService.(interface{ RegisterMarketplaceRoutes(*mux.Router) }); ok {
		marketplace.RegisterMarketplaceRoutes(h.router)
		h.logger.Info("Registered marketplace routes")
	}
}

// Start starts the HTTP trigger server
func (h *HTTPTrigger) Start() error {
	// Advanced Dashboard endpoint
	advancedDashboard := web.NewAdvancedDashboardHandler(h.engine, h.logger)
	h.router.Handle("/", advancedDashboard).Methods("GET")

	// API endpoints for dashboard data
	h.router.HandleFunc("/metrics", advancedDashboard.HandleMetrics).Methods("GET")
	h.router.HandleFunc("/logs", advancedDashboard.HandleLogs).Methods("GET")

	// Event webhook endpoint
	h.router.HandleFunc("/webhook/{event}", h.handleWebhook).Methods("POST")
	h.router.HandleFunc("/events", h.handleEvent).Methods("POST")

	// Health check endpoint
	h.router.HandleFunc("/health", h.handleHealth).Methods("GET")

	// Workflow management endpoints
	h.router.HandleFunc("/workflows", h.handleListWorkflows).Methods("GET")
	h.router.HandleFunc("/instances/{id}", h.handleGetInstance).Methods("GET")

	h.server = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", h.port),
		Handler:      h.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	h.logger.Info("Starting HTTP trigger server", zap.Int("port", h.port))
	return h.server.ListenAndServe()
}

// Stop stops the HTTP trigger server
func (h *HTTPTrigger) Stop(ctx context.Context) error {
	if h.server != nil {
		return h.server.Shutdown(ctx)
	}
	return nil
}

// handleWebhook handles webhook-style events with event type in URL
func (h *HTTPTrigger) handleWebhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventType := vars["event"]

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.logger.Error("Failed to decode webhook payload", zap.Error(err))
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	h.triggerWorkflow(w, r, eventType, payload)
}

// handleEvent handles generic events with event type in payload
func (h *HTTPTrigger) handleEvent(w http.ResponseWriter, r *http.Request) {
	var eventPayload EventPayload
	if err := json.NewDecoder(r.Body).Decode(&eventPayload); err != nil {
		h.logger.Error("Failed to decode event payload", zap.Error(err))
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if eventPayload.Event == "" {
		http.Error(w, "Event type is required", http.StatusBadRequest)
		return
	}

	h.triggerWorkflow(w, r, eventPayload.Event, eventPayload.Data)
}

// triggerWorkflow triggers a workflow for the given event
func (h *HTTPTrigger) triggerWorkflow(w http.ResponseWriter, r *http.Request, eventType string, data map[string]interface{}) {
	workflow, exists := h.engine.GetWorkflowForEvent(eventType)
	if !exists {
		h.logger.Warn("No workflow found for event", zap.String("event", eventType))
		http.Error(w, fmt.Sprintf("No workflow found for event: %s", eventType), http.StatusNotFound)
		return
	}

	// Create event context
	eventCtx := &persistence.EventContext{
		Event: &persistence.Event{
			Type:      eventType,
			Payload:   data,
			Metadata:  make(map[string]interface{}),
			Timestamp: time.Now().Unix(),
		},
		Variables: make(map[string]interface{}),
	}

	// Add request metadata
	eventCtx.Event.Metadata["remote_addr"] = r.RemoteAddr
	eventCtx.Event.Metadata["user_agent"] = r.UserAgent()

	h.logger.Info("Triggering workflow",
		zap.String("event", eventType),
		zap.String("workflow", workflow.Name))

	// Execute workflow asynchronously
	go func() {
		ctx := context.Background()
		instanceID, err := h.engine.ExecuteWorkflow(ctx, workflow, eventCtx)
		if err != nil {
			h.logger.Error("Workflow execution failed",
				zap.String("instance_id", instanceID),
				zap.Error(err))
		} else {
			h.logger.Info("Workflow execution completed",
				zap.String("instance_id", instanceID))
		}
	}()

	// Return immediate response
	response := map[string]interface{}{
		"status":    "accepted",
		"event":     eventType,
		"workflow":  workflow.Name,
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth handles health check requests
func (h *HTTPTrigger) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListWorkflows lists all registered workflows
func (h *HTTPTrigger) handleListWorkflows(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, you'd get this from the engine
	response := map[string]interface{}{
		"workflows": []string{}, // Placeholder
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetInstance retrieves a workflow instance
func (h *HTTPTrigger) handleGetInstance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["id"]

	instance, err := h.engine.GetWorkflowInstance(instanceID)
	if err != nil {
		h.logger.Error("Failed to get workflow instance",
			zap.String("instance_id", instanceID),
			zap.Error(err))
		http.Error(w, "Instance not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instance)
}
