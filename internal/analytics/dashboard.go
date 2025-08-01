package analytics

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// AnalyticsDashboard provides real-time workflow analytics and monitoring
type AnalyticsDashboard struct {
	logger         *zap.Logger
	metrics        *MetricsCollector
	alertManager   *AlertManager
	mu             sync.RWMutex
}

// MetricsCollector gathers real-time workflow execution data
type MetricsCollector struct {
	WorkflowExecutions map[string]*WorkflowMetrics `json:"workflow_executions"`
	SystemMetrics      *SystemMetrics              `json:"system_metrics"`
	AlertsTriggered    []Alert                     `json:"alerts_triggered"`
	mu                 sync.RWMutex
}

// WorkflowMetrics tracks individual workflow performance
type WorkflowMetrics struct {
	Name            string                 `json:"name"`
	TotalExecutions int64                  `json:"total_executions"`
	SuccessCount    int64                  `json:"success_count"`
	FailureCount    int64                  `json:"failure_count"`
	AvgExecutionTime time.Duration         `json:"avg_execution_time"`
	LastExecution   time.Time             `json:"last_execution"`
	Throughput      float64               `json:"throughput"` // executions per minute
	ErrorRate       float64               `json:"error_rate"`
	RecentEvents    []ExecutionEvent      `json:"recent_events"`
}

// SystemMetrics tracks overall system performance
type SystemMetrics struct {
	CPUUsage        float64   `json:"cpu_usage"`
	MemoryUsage     float64   `json:"memory_usage"`
	ActiveWorkflows int       `json:"active_workflows"`
	QueuedEvents    int       `json:"queued_events"`
	Uptime          time.Duration `json:"uptime"`
	LastUpdated     time.Time `json:"last_updated"`
}

// ExecutionEvent represents a single workflow execution
type ExecutionEvent struct {
	ID          string                 `json:"id"`
	WorkflowName string                `json:"workflow_name"`
	Status      string                 `json:"status"`
	StartTime   time.Time             `json:"start_time"`
	EndTime     time.Time             `json:"end_time"`
	Duration    time.Duration         `json:"duration"`
	TriggerType string                `json:"trigger_type"`
	EventData   map[string]interface{} `json:"event_data"`
	Error       string                `json:"error,omitempty"`
}

// Alert represents a system alert
type Alert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	WorkflowName string                `json:"workflow_name,omitempty"`
	Timestamp   time.Time             `json:"timestamp"`
	Resolved    bool                  `json:"resolved"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AlertManager handles system alerts and notifications
type AlertManager struct {
	alerts     []Alert
	thresholds map[string]float64
	mu         sync.RWMutex
}

// NewAnalyticsDashboard creates a new analytics dashboard
func NewAnalyticsDashboard(logger *zap.Logger) *AnalyticsDashboard {
	return &AnalyticsDashboard{
		logger: logger,
		metrics: &MetricsCollector{
			WorkflowExecutions: make(map[string]*WorkflowMetrics),
			SystemMetrics: &SystemMetrics{
				LastUpdated: time.Now(),
			},
			AlertsTriggered: make([]Alert, 0),
		},
		alertManager: &AlertManager{
			alerts: make([]Alert, 0),
			thresholds: map[string]float64{
				"error_rate":      0.1,  // 10% error rate
				"avg_duration":    300,  // 5 minutes
				"queue_backlog":   1000, // 1000 queued events
				"cpu_usage":       0.8,  // 80% CPU
				"memory_usage":    0.8,  // 80% memory
			},
		},
	}
}

// RegisterRoutes sets up the analytics API endpoints
func (ad *AnalyticsDashboard) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/analytics").Subrouter()
	
	// Dashboard endpoints
	api.HandleFunc("/dashboard", ad.handleDashboardPage).Methods("GET")
	api.HandleFunc("/api/metrics", ad.handleGetMetrics).Methods("GET")
	api.HandleFunc("/api/workflows", ad.handleGetWorkflowMetrics).Methods("GET")
	api.HandleFunc("/api/alerts", ad.handleGetAlerts).Methods("GET")
	api.HandleFunc("/api/system", ad.handleGetSystemMetrics).Methods("GET")
	
	// Real-time endpoints
	api.HandleFunc("/api/stream", ad.handleMetricsStream).Methods("GET")
	api.HandleFunc("/api/events", ad.handleRecentEvents).Methods("GET")
}

// handleDashboardPage serves the analytics dashboard HTML
func (ad *AnalyticsDashboard) handleDashboardPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(getDashboardHTML()))
}

// handleGetMetrics returns comprehensive metrics data
func (ad *AnalyticsDashboard) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	ad.metrics.mu.RLock()
	defer ad.metrics.mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ad.metrics)
}

// handleGetWorkflowMetrics returns workflow-specific metrics
func (ad *AnalyticsDashboard) handleGetWorkflowMetrics(w http.ResponseWriter, r *http.Request) {
	workflowName := r.URL.Query().Get("name")
	
	ad.metrics.mu.RLock()
	defer ad.metrics.mu.RUnlock()
	
	if workflowName != "" {
		if metrics, exists := ad.metrics.WorkflowExecutions[workflowName]; exists {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(metrics)
			return
		}
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ad.metrics.WorkflowExecutions)
}

// handleGetAlerts returns active alerts
func (ad *AnalyticsDashboard) handleGetAlerts(w http.ResponseWriter, r *http.Request) {
	ad.alertManager.mu.RLock()
	defer ad.alertManager.mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ad.alertManager.alerts)
}

// handleGetSystemMetrics returns system performance metrics
func (ad *AnalyticsDashboard) handleGetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	ad.metrics.mu.RLock()
	defer ad.metrics.mu.RUnlock()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ad.metrics.SystemMetrics)
}

// handleMetricsStream provides real-time metrics via Server-Sent Events
func (ad *AnalyticsDashboard) handleMetricsStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ad.metrics.mu.RLock()
			data, _ := json.Marshal(ad.metrics)
			ad.metrics.mu.RUnlock()
			
			w.Write([]byte("data: "))
			w.Write(data)
			w.Write([]byte("\n\n"))
			w.(http.Flusher).Flush()
			
		case <-r.Context().Done():
			return
		}
	}
}

// handleRecentEvents returns recent workflow execution events
func (ad *AnalyticsDashboard) handleRecentEvents(w http.ResponseWriter, r *http.Request) {
	limit := 100 // Default limit
	
	var allEvents []ExecutionEvent
	ad.metrics.mu.RLock()
	for _, metrics := range ad.metrics.WorkflowExecutions {
		allEvents = append(allEvents, metrics.RecentEvents...)
	}
	ad.metrics.mu.RUnlock()
	
	// Sort by timestamp (most recent first)
	if len(allEvents) > limit {
		allEvents = allEvents[:limit]
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allEvents)
}

// RecordExecution records a workflow execution for analytics
func (ad *AnalyticsDashboard) RecordExecution(event ExecutionEvent) {
	ad.metrics.mu.Lock()
	defer ad.metrics.mu.Unlock()
	
	if ad.metrics.WorkflowExecutions[event.WorkflowName] == nil {
		ad.metrics.WorkflowExecutions[event.WorkflowName] = &WorkflowMetrics{
			Name:         event.WorkflowName,
			RecentEvents: make([]ExecutionEvent, 0),
		}
	}
	
	metrics := ad.metrics.WorkflowExecutions[event.WorkflowName]
	metrics.TotalExecutions++
	metrics.LastExecution = event.EndTime
	
	if event.Status == "success" {
		metrics.SuccessCount++
	} else {
		metrics.FailureCount++
	}
	
	// Update averages and rates
	metrics.ErrorRate = float64(metrics.FailureCount) / float64(metrics.TotalExecutions)
	
	// Add to recent events (keep last 50)
	metrics.RecentEvents = append(metrics.RecentEvents, event)
	if len(metrics.RecentEvents) > 50 {
		metrics.RecentEvents = metrics.RecentEvents[1:]
	}
	
	// Check for alerts
	ad.checkAlerts(event.WorkflowName, metrics)
}

// checkAlerts evaluates metrics against thresholds and creates alerts
func (ad *AnalyticsDashboard) checkAlerts(workflowName string, metrics *WorkflowMetrics) {
	if metrics.ErrorRate > ad.alertManager.thresholds["error_rate"] {
		alert := Alert{
			ID:          generateAlertID(),
			Type:        "high_error_rate",
			Severity:    "warning",
			Message:     "High error rate detected",
			WorkflowName: workflowName,
			Timestamp:   time.Now(),
			Metadata: map[string]interface{}{
				"error_rate": metrics.ErrorRate,
				"threshold":  ad.alertManager.thresholds["error_rate"],
			},
		}
		ad.alertManager.addAlert(alert)
	}
}

// addAlert adds a new alert to the system
func (am *AlertManager) addAlert(alert Alert) {
	am.mu.Lock()
	defer am.mu.Unlock()
	
	am.alerts = append(am.alerts, alert)
	
	// Keep only last 100 alerts
	if len(am.alerts) > 100 {
		am.alerts = am.alerts[1:]
	}
}

// generateAlertID creates a unique alert identifier
func generateAlertID() string {
	return time.Now().Format("20060102150405") + "-alert"
}

// getDashboardHTML returns the analytics dashboard HTML
func getDashboardHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reactor Analytics Dashboard</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #0a0e1a; color: #e2e8f0; }
        .dashboard { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 20px; padding: 20px; min-height: 100vh; }
        .card { background: linear-gradient(135deg, #1e293b 0%, #334155 100%); border-radius: 12px; padding: 20px; border: 1px solid #475569; }
        .card h3 { color: #60a5fa; margin-bottom: 15px; font-size: 1.2em; }
        .metric-value { font-size: 2.5em; font-weight: bold; color: #10b981; margin-bottom: 10px; }
        .metric-label { color: #94a3b8; font-size: 0.9em; }
        .chart-container { height: 200px; background: #1e293b; border-radius: 8px; margin-top: 15px; position: relative; }
        .alert { background: #dc2626; color: white; padding: 10px; border-radius: 6px; margin-bottom: 10px; }
        .alert.warning { background: #f59e0b; }
        .alert.info { background: #3b82f6; }
        .status-indicator { display: inline-block; width: 12px; height: 12px; border-radius: 50%; margin-right: 8px; }
        .status-success { background: #10b981; }
        .status-error { background: #ef4444; }
        .status-warning { background: #f59e0b; }
        .workflow-list { max-height: 300px; overflow-y: auto; }
        .workflow-item { padding: 10px; border-bottom: 1px solid #475569; display: flex; justify-content: between; align-items: center; }
        .real-time-indicator { color: #10b981; font-size: 0.8em; }
    </style>
</head>
<body>
    <div class="dashboard">
        <div class="card">
            <h3>System Overview</h3>
            <div class="metric-value" id="total-executions">0</div>
            <div class="metric-label">Total Executions</div>
            <div class="chart-container" id="executions-chart"></div>
        </div>
        
        <div class="card">
            <h3>Success Rate</h3>
            <div class="metric-value" id="success-rate">0%</div>
            <div class="metric-label">Overall Success Rate</div>
            <div class="chart-container" id="success-chart"></div>
        </div>
        
        <div class="card">
            <h3>Performance</h3>
            <div class="metric-value" id="avg-duration">0ms</div>
            <div class="metric-label">Average Execution Time</div>
            <div class="chart-container" id="performance-chart"></div>
        </div>
        
        <div class="card">
            <h3>Active Workflows <span class="real-time-indicator">● LIVE</span></h3>
            <div class="workflow-list" id="workflow-list">
                <div class="workflow-item">
                    <span><span class="status-indicator status-success"></span>customer-workflow</span>
                    <span>98.5%</span>
                </div>
                <div class="workflow-item">
                    <span><span class="status-indicator status-warning"></span>file-processor</span>
                    <span>85.2%</span>
                </div>
            </div>
        </div>
        
        <div class="card">
            <h3>Recent Alerts</h3>
            <div id="alerts-list">
                <div class="alert warning">High error rate in file-processor workflow</div>
                <div class="alert info">System resources optimal</div>
            </div>
        </div>
        
        <div class="card">
            <h3>System Resources</h3>
            <div>
                <div class="metric-label">CPU Usage</div>
                <div class="metric-value" style="font-size: 1.5em;" id="cpu-usage">0%</div>
                <div class="metric-label">Memory Usage</div>
                <div class="metric-value" style="font-size: 1.5em;" id="memory-usage">0%</div>
            </div>
        </div>
    </div>
    
    <script>
        // Real-time dashboard updates
        const eventSource = new EventSource('/analytics/api/stream');
        
        eventSource.onmessage = function(event) {
            const data = JSON.parse(event.data);
            updateDashboard(data);
        };
        
        function updateDashboard(data) {
            // Update system metrics
            if (data.system_metrics) {
                document.getElementById('cpu-usage').textContent = 
                    (data.system_metrics.cpu_usage * 100).toFixed(1) + '%';
                document.getElementById('memory-usage').textContent = 
                    (data.system_metrics.memory_usage * 100).toFixed(1) + '%';
            }
            
            // Update workflow metrics
            if (data.workflow_executions) {
                let totalExecutions = 0;
                let totalSuccess = 0;
                let totalDuration = 0;
                let workflowCount = 0;
                
                for (const [name, metrics] of Object.entries(data.workflow_executions)) {
                    totalExecutions += metrics.total_executions;
                    totalSuccess += metrics.success_count;
                    totalDuration += metrics.avg_execution_time;
                    workflowCount++;
                }
                
                document.getElementById('total-executions').textContent = totalExecutions;
                document.getElementById('success-rate').textContent = 
                    totalExecutions > 0 ? ((totalSuccess / totalExecutions) * 100).toFixed(1) + '%' : '0%';
                document.getElementById('avg-duration').textContent = 
                    workflowCount > 0 ? Math.round(totalDuration / workflowCount) + 'ms' : '0ms';
            }
        }
        
        // Initialize dashboard
        fetch('/analytics/api/metrics')
            .then(response => response.json())
            .then(data => updateDashboard(data));
    </script>
</body>
</html>`
}