package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"time"

	"github.com/logimos/conduktr/internal/engine"

	"go.uber.org/zap"
)

const advancedDashboardHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Reactor - Advanced Workflow Dashboard</title>
    <meta charset="utf-8">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            background: #0f1419; 
            color: #e6e6e6; 
            line-height: 1.6; 
        }
        
        .navbar {
            background: #1a1f2e;
            padding: 1rem 2rem;
            border-bottom: 1px solid #2d3748;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .navbar h1 {
            color: #4fd1c7;
            font-size: 1.5rem;
            font-weight: 600;
        }
        
        .status-indicators {
            display: flex;
            gap: 1rem;
        }
        
        .status-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background: #10b981;
            animation: pulse 2s infinite;
        }
        
        @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
        
        .container { 
            max-width: 1400px; 
            margin: 0 auto; 
            padding: 2rem; 
        }
        
        .dashboard-grid {
            display: grid;
            grid-template-columns: 1fr 1fr 1fr;
            gap: 2rem;
            margin-bottom: 2rem;
        }
        
        .card {
            background: #1a1f2e;
            border: 1px solid #2d3748;
            border-radius: 8px;
            padding: 1.5rem;
            transition: all 0.3s ease;
        }
        
        .card:hover {
            border-color: #4fd1c7;
            transform: translateY(-2px);
            box-shadow: 0 4px 20px rgba(79, 209, 199, 0.1);
        }
        
        .card h3 {
            color: #4fd1c7;
            margin-bottom: 1rem;
            font-size: 1.1rem;
        }
        
        .metric {
            display: flex;
            justify-content: space-between;
            margin: 0.5rem 0;
        }
        
        .metric-value {
            color: #10b981;
            font-weight: 600;
        }
        
        .workflow-section {
            margin: 2rem 0;
        }
        
        .section-title {
            color: #4fd1c7;
            font-size: 1.3rem;
            margin-bottom: 1.5rem;
            padding-bottom: 0.5rem;
            border-bottom: 2px solid #2d3748;
        }
        
        .workflow-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
            gap: 1.5rem;
        }
        
        .workflow-card {
            background: #1a1f2e;
            border: 1px solid #2d3748;
            border-radius: 8px;
            overflow: hidden;
        }
        
        .workflow-header {
            background: linear-gradient(135deg, #4fd1c7, #3b82f6);
            padding: 1rem;
            color: #0f1419;
            font-weight: 600;
        }
        
        .workflow-body {
            padding: 1.5rem;
        }
        
        .workflow-steps {
            margin: 1rem 0;
        }
        
        .step {
            display: flex;
            align-items: center;
            padding: 0.5rem;
            margin: 0.5rem 0;
            background: #0f1419;
            border-radius: 4px;
            border-left: 3px solid #4fd1c7;
        }
        
        .step-number {
            background: #4fd1c7;
            color: #0f1419;
            width: 24px;
            height: 24px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 0.8rem;
            font-weight: 600;
            margin-right: 1rem;
        }
        
        .execution-log {
            background: #0f1419;
            border: 1px solid #2d3748;
            border-radius: 8px;
            padding: 1.5rem;
            margin: 2rem 0;
            max-height: 400px;
            overflow-y: auto;
        }
        
        .log-entry {
            display: flex;
            padding: 0.5rem 0;
            border-bottom: 1px solid #2d3748;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.9rem;
        }
        
        .log-timestamp {
            color: #6b7280;
            margin-right: 1rem;
            min-width: 120px;
        }
        
        .log-level {
            margin-right: 1rem;
            font-weight: 600;
            min-width: 60px;
        }
        
        .log-level.info { color: #3b82f6; }
        .log-level.warn { color: #f59e0b; }
        .log-level.error { color: #ef4444; }
        
        .log-message {
            color: #e6e6e6;
        }
        
        .trigger-buttons {
            display: flex;
            gap: 1rem;
            margin: 2rem 0;
        }
        
        .btn {
            background: linear-gradient(135deg, #4fd1c7, #3b82f6);
            color: #0f1419;
            border: none;
            padding: 0.75rem 1.5rem;
            border-radius: 6px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
        }
        
        .btn:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(79, 209, 199, 0.3);
        }
        
        .flow-diagram {
            background: #0f1419;
            border: 1px solid #2d3748;
            border-radius: 8px;
            padding: 2rem;
            margin: 2rem 0;
            position: relative;
        }
        
        .flow-node {
            background: #1a1f2e;
            border: 2px solid #4fd1c7;
            border-radius: 8px;
            padding: 1rem;
            margin: 1rem;
            display: inline-block;
            position: relative;
            min-width: 120px;
            text-align: center;
        }
        
        .flow-node.executing {
            border-color: #f59e0b;
            background: linear-gradient(135deg, #f59e0b22, #f59e0b11);
            animation: pulse 1s infinite;
        }
        
        .flow-node.completed {
            border-color: #10b981;
            background: linear-gradient(135deg, #10b98122, #10b98111);
        }
        
        .flow-node.failed {
            border-color: #ef4444;
            background: linear-gradient(135deg, #ef444422, #ef444411);
        }
        
        .flow-arrow {
            position: absolute;
            color: #4fd1c7;
            font-size: 1.5rem;
            top: 50%;
            transform: translateY(-50%);
        }
        
        .metrics-charts {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 2rem;
            margin: 2rem 0;
        }
        
        .chart-container {
            background: #1a1f2e;
            border: 1px solid #2d3748;
            border-radius: 8px;
            padding: 1.5rem;
        }
        
        .refresh-indicator {
            position: fixed;
            top: 20px;
            right: 20px;
            background: #4fd1c7;
            color: #0f1419;
            padding: 0.5rem 1rem;
            border-radius: 20px;
            font-size: 0.9rem;
            font-weight: 600;
            opacity: 0;
            transition: opacity 0.3s ease;
        }
        
        .refresh-indicator.show {
            opacity: 1;
        }
        
        .api-docs {
            background: #1a1f2e;
            border: 1px solid #2d3748;
            border-radius: 8px;
            padding: 1.5rem;
            margin: 2rem 0;
        }
        
        .endpoint {
            background: #0f1419;
            border: 1px solid #2d3748;
            border-radius: 4px;
            padding: 1rem;
            margin: 0.5rem 0;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.9rem;
        }
        
        .method {
            color: #10b981;
            font-weight: 600;
            margin-right: 1rem;
        }
        
        .method.post { color: #3b82f6; }
        .method.get { color: #10b981; }
        .method.delete { color: #ef4444; }
        
        @media (max-width: 768px) {
            .dashboard-grid, .workflow-grid, .metrics-charts {
                grid-template-columns: 1fr;
            }
            
            .container {
                padding: 1rem;
            }
        }
    </style>
</head>
<body>
    <div class="navbar">
        <h1>🔄 Reactor Workflow Engine</h1>
        <div class="status-indicators">
            <div class="status-dot" title="Engine Status"></div>
            <span>System Online</span>
        </div>
    </div>

    <div class="container">
        <!-- System Metrics -->
        <div class="dashboard-grid">
            <div class="card">
                <h3>📊 System Metrics</h3>
                <div class="metric">
                    <span>Active Workflows</span>
                    <span class="metric-value" id="active-workflows">2</span>
                </div>
                <div class="metric">
                    <span>Total Executions</span>
                    <span class="metric-value" id="total-executions">247</span>
                </div>
                <div class="metric">
                    <span>Success Rate</span>
                    <span class="metric-value" id="success-rate">94.3%</span>
                </div>
                <div class="metric">
                    <span>Avg Execution Time</span>
                    <span class="metric-value" id="avg-time">2.4s</span>
                </div>
            </div>

            <div class="card">
                <h3>🔥 Trigger Statistics</h3>
                <div class="metric">
                    <span>HTTP Triggers</span>
                    <span class="metric-value">156</span>
                </div>
                <div class="metric">
                    <span>File Events</span>
                    <span class="metric-value">89</span>
                </div>
                <div class="metric">
                    <span>Database Changes</span>
                    <span class="metric-value">12</span>
                </div>
                <div class="metric">
                    <span>Redis Events</span>
                    <span class="metric-value">23</span>
                </div>
            </div>

            <div class="card">
                <h3>⚡ Performance</h3>
                <div class="metric">
                    <span>CPU Usage</span>
                    <span class="metric-value">12%</span>
                </div>
                <div class="metric">
                    <span>Memory Usage</span>
                    <span class="metric-value">64MB</span>
                </div>
                <div class="metric">
                    <span>Active Goroutines</span>
                    <span class="metric-value">18</span>
                </div>
                <div class="metric">
                    <span>Queue Depth</span>
                    <span class="metric-value">3</span>
                </div>
            </div>
        </div>

        <!-- Active Workflows -->
        <div class="workflow-section">
            <h2 class="section-title">🔄 Active Workflows</h2>
            <div class="workflow-grid">
                <div class="workflow-card">
                    <div class="workflow-header">
                        customer-creation-workflow
                    </div>
                    <div class="workflow-body">
                        <p><strong>Trigger:</strong> customer.created (HTTP)</p>
                        <p><strong>Last Execution:</strong> 2 minutes ago</p>
                        <p><strong>Success Rate:</strong> 98.2%</p>
                        
                        <div class="workflow-steps">
                            <div class="step">
                                <div class="step-number">1</div>
                                <span>Validate Customer Data</span>
                            </div>
                            <div class="step">
                                <div class="step-number">2</div>
                                <span>Send Welcome Email</span>
                            </div>
                            <div class="step">
                                <div class="step-number">3</div>
                                <span>Create User Profile</span>
                            </div>
                            <div class="step">
                                <div class="step-number">4</div>
                                <span>Log Completion</span>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="workflow-card">
                    <div class="workflow-header">
                        file-processing-workflow
                    </div>
                    <div class="workflow-body">
                        <p><strong>Trigger:</strong> file.created (File System)</p>
                        <p><strong>Last Execution:</strong> 5 minutes ago</p>
                        <p><strong>Success Rate:</strong> 89.7%</p>
                        
                        <div class="workflow-steps">
                            <div class="step">
                                <div class="step-number">1</div>
                                <span>Detect File Change</span>
                            </div>
                            <div class="step">
                                <div class="step-number">2</div>
                                <span>Process File Content</span>
                            </div>
                            <div class="step">
                                <div class="step-number">3</div>
                                <span>Create Backup</span>
                            </div>
                            <div class="step">
                                <div class="step-number">4</div>
                                <span>Notify Completion</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Execution Flow Diagram -->
        <div class="workflow-section">
            <h2 class="section-title">🌊 Current Execution Flow</h2>
            <div class="flow-diagram">
                <div class="flow-node completed">Trigger Received</div>
                <span class="flow-arrow">→</span>
                <div class="flow-node executing">Validate Input</div>
                <span class="flow-arrow">→</span>
                <div class="flow-node">Execute Actions</div>
                <span class="flow-arrow">→</span>
                <div class="flow-node">Complete</div>
            </div>
        </div>

        <!-- Real-time Execution Log -->
        <div class="workflow-section">
            <h2 class="section-title">📋 Real-time Execution Log</h2>
            <div class="execution-log" id="execution-log">
                <div class="log-entry">
                    <span class="log-timestamp">14:32:15</span>
                    <span class="log-level info">INFO</span>
                    <span class="log-message">Workflow 'customer-creation-workflow' triggered by HTTP event</span>
                </div>
                <div class="log-entry">
                    <span class="log-timestamp">14:32:16</span>
                    <span class="log-level info">INFO</span>
                    <span class="log-message">Step 'validate_customer' completed successfully</span>
                </div>
                <div class="log-entry">
                    <span class="log-timestamp">14:32:17</span>
                    <span class="log-level warn">WARN</span>
                    <span class="log-message">Email service response time: 2.3s (threshold: 2s)</span>
                </div>
                <div class="log-entry">
                    <span class="log-timestamp">14:32:18</span>
                    <span class="log-level info">INFO</span>
                    <span class="log-message">Workflow execution completed - Duration: 3.2s</span>
                </div>
            </div>
        </div>

        <!-- Test Triggers -->
        <div class="workflow-section">
            <h2 class="section-title">🧪 Test Workflow Triggers</h2>
            <div class="trigger-buttons">
                <button class="btn" onclick="triggerCustomerWorkflow()">Trigger Customer Workflow</button>
                <button class="btn" onclick="triggerFileWorkflow()">Trigger File Workflow</button>
                <button class="btn" onclick="triggerDatabaseEvent()">Trigger Database Event</button>
                <button class="btn" onclick="triggerRedisEvent()">Trigger Redis Event</button>
            </div>
        </div>

        <!-- API Documentation -->
        <div class="workflow-section">
            <h2 class="section-title">📚 API Endpoints</h2>
            <div class="api-docs">
                <div class="endpoint">
                    <span class="method get">GET</span>
                    <span>/health</span>
                    <span style="color: #6b7280; margin-left: 1rem;">System health check</span>
                </div>
                <div class="endpoint">
                    <span class="method post">POST</span>
                    <span>/webhook/{event}</span>
                    <span style="color: #6b7280; margin-left: 1rem;">Trigger workflow by event type</span>
                </div>
                <div class="endpoint">
                    <span class="method post">POST</span>
                    <span>/events</span>
                    <span style="color: #6b7280; margin-left: 1rem;">Trigger workflow with event payload</span>
                </div>
                <div class="endpoint">
                    <span class="method get">GET</span>
                    <span>/workflows</span>
                    <span style="color: #6b7280; margin-left: 1rem;">List all registered workflows</span>
                </div>
                <div class="endpoint">
                    <span class="method get">GET</span>
                    <span>/instances/{id}</span>
                    <span style="color: #6b7280; margin-left: 1rem;">Get workflow instance details</span>
                </div>
                <div class="endpoint">
                    <span class="method get">GET</span>
                    <span>/metrics</span>
                    <span style="color: #6b7280; margin-left: 1rem;">System metrics and statistics</span>
                </div>
                <div class="endpoint">
                    <span class="method get">GET</span>
                    <span>/logs</span>
                    <span style="color: #6b7280; margin-left: 1rem;">Real-time execution logs</span>
                </div>
            </div>
        </div>
    </div>

    <div class="refresh-indicator" id="refresh-indicator">
        Refreshing...
    </div>

    <script>
        // Auto-refresh dashboard every 5 seconds
        setInterval(() => {
            const indicator = document.getElementById('refresh-indicator');
            indicator.classList.add('show');
            
            // Simulate data refresh
            updateMetrics();
            updateLogs();
            
            setTimeout(() => {
                indicator.classList.remove('show');
            }, 1000);
        }, 5000);

        function updateMetrics() {
            // Simulate metric updates
            const executions = document.getElementById('total-executions');
            if (executions) {
                const current = parseInt(executions.textContent);
                executions.textContent = current + Math.floor(Math.random() * 3);
            }
        }

        function updateLogs() {
            const logContainer = document.getElementById('execution-log');
            const now = new Date();
            const timestamp = now.toTimeString().split(' ')[0];
            
            const messages = [
                'Workflow execution started',
                'Database connection established', 
                'Redis event processed',
                'File monitoring active',
                'HTTP trigger received'
            ];
            
            const levels = ['info', 'warn', 'error'];
            const level = levels[Math.floor(Math.random() * levels.length)];
            const message = messages[Math.floor(Math.random() * messages.length)];
            
            const logEntry = document.createElement('div');
            logEntry.className = 'log-entry';
            logEntry.innerHTML = ` + "`" + `
                <span class="log-timestamp">${timestamp}</span>
                <span class="log-level ${level}">${level.toUpperCase()}</span>
                <span class="log-message">${message}</span>
            ` + "`" + `;
            
            logContainer.insertBefore(logEntry, logContainer.firstChild);
            
            // Keep only last 20 entries
            while (logContainer.children.length > 20) {
                logContainer.removeChild(logContainer.lastChild);
            }
        }

        async function triggerCustomerWorkflow() {
            try {
                const response = await fetch('/webhook/customer.created', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        name: 'Test User',
                        email: 'test@example.com',
                        id: Math.random().toString(36).substr(2, 9)
                    })
                });
                
                if (response.ok) {
                    alert('Customer workflow triggered successfully!');
                } else {
                    alert('Failed to trigger workflow');
                }
            } catch (error) {
                alert('Error: ' + error.message);
            }
        }

        async function triggerFileWorkflow() {
            alert('File workflow will be triggered when files are added to the workflows directory');
        }

        async function triggerDatabaseEvent() {
            alert('Database triggers require configuration - see documentation');
        }

        async function triggerRedisEvent() {
            alert('Redis triggers require Redis connection - see configuration');
        }
    </script>
</body>
</html>
`

// AdvancedDashboardHandler serves the advanced dashboard
type AdvancedDashboardHandler struct {
	engine *engine.Engine
	logger *zap.Logger
}

// NewAdvancedDashboardHandler creates a new advanced dashboard handler
func NewAdvancedDashboardHandler(engine *engine.Engine, logger *zap.Logger) *AdvancedDashboardHandler {
	return &AdvancedDashboardHandler{
		engine: engine,
		logger: logger,
	}
}

// ServeHTTP handles advanced dashboard requests
func (a *AdvancedDashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.New("advanced-dashboard").Parse(advancedDashboardHTML)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// MetricsResponse represents system metrics
type MetricsResponse struct {
	ActiveWorkflows  int     `json:"active_workflows"`
	TotalExecutions  int     `json:"total_executions"`
	SuccessRate      float64 `json:"success_rate"`
	AvgExecutionTime float64 `json:"avg_execution_time"`
	HTTPTriggers     int     `json:"http_triggers"`
	FileTriggers     int     `json:"file_triggers"`
	DatabaseTriggers int     `json:"database_triggers"`
	RedisTriggers    int     `json:"redis_triggers"`
	CPUUsage         float64 `json:"cpu_usage"`
	MemoryUsage      int64   `json:"memory_usage"`
	ActiveGoroutines int     `json:"active_goroutines"`
	QueueDepth       int     `json:"queue_depth"`
}

// HandleMetrics returns system metrics as JSON
func (a *AdvancedDashboardHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := MetricsResponse{
		ActiveWorkflows:  2,
		TotalExecutions:  247,
		SuccessRate:      94.3,
		AvgExecutionTime: 2.4,
		HTTPTriggers:     156,
		FileTriggers:     89,
		DatabaseTriggers: 12,
		RedisTriggers:    23,
		CPUUsage:         12.0,
		MemoryUsage:      64 * 1024 * 1024, // 64MB
		ActiveGoroutines: 18,
		QueueDepth:       3,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// HandleLogs returns recent execution logs
func (a *AdvancedDashboardHandler) HandleLogs(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would fetch from a log storage system
	logs := []LogEntry{
		{
			Timestamp: time.Now().Add(-2 * time.Minute),
			Level:     "INFO",
			Message:   "Workflow 'customer-creation-workflow' triggered by HTTP event",
			Context:   map[string]interface{}{"workflow": "customer-creation-workflow", "trigger": "http"},
		},
		{
			Timestamp: time.Now().Add(-1 * time.Minute),
			Level:     "INFO",
			Message:   "Step 'validate_customer' completed successfully",
			Context:   map[string]interface{}{"step": "validate_customer", "duration": "0.5s"},
		},
		{
			Timestamp: time.Now().Add(-30 * time.Second),
			Level:     "WARN",
			Message:   "Email service response time: 2.3s (threshold: 2s)",
			Context:   map[string]interface{}{"service": "email", "response_time": "2.3s"},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}
