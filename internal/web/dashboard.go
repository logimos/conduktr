package web

import (
	"html/template"
	"net/http"

	"github.com/logimos/conduktr/internal/engine"
)

const dashboardHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Reactor Workflow Engine</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; text-align: center; margin-bottom: 30px; }
        .section { margin: 30px 0; }
        .workflow-card { background: #f8f9fa; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #007bff; }
        .endpoint { background: #e9ecef; padding: 10px; margin: 5px 0; border-radius: 3px; font-family: monospace; }
        .status-healthy { color: #28a745; font-weight: bold; }
        .instance { background: #fff3cd; padding: 10px; margin: 5px 0; border-radius: 3px; border-left: 3px solid #ffc107; }
        ul { list-style-type: none; padding: 0; }
        li { margin: 5px 0; }
        .api-example { background: #f8f9fa; padding: 15px; border-radius: 5px; margin: 10px 0; }
        code { background: #e9ecef; padding: 2px 5px; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔄 Reactor Workflow Engine</h1>
        
        <div class="section">
            <h2>System Status</h2>
            <p class="status-healthy">✅ Engine Running</p>
            <p class="status-healthy">✅ HTTP Server Active (Port 5000)</p>
            <p class="status-healthy">✅ File Monitoring Active</p>
        </div>

        <div class="section">
            <h2>Available Workflows</h2>
            <div class="workflow-card">
                <h3>customer-creation-workflow</h3>
                <p><strong>Trigger:</strong> customer.created</p>
                <p><strong>Steps:</strong> Validate → Notify Team → Create Welcome File → Log Completion</p>
            </div>
            <div class="workflow-card">
                <h3>file-processing-workflow</h3>
                <p><strong>Trigger:</strong> file.created</p>
                <p><strong>Steps:</strong> Log Event → Process Text File → Backup File → Notify Complete</p>
            </div>
        </div>

        <div class="section">
            <h2>API Endpoints</h2>
            <div class="endpoint">GET /health - Health check</div>
            <div class="endpoint">POST /webhook/{event} - Trigger workflow by event type</div>
            <div class="endpoint">POST /events - Trigger workflow with event in payload</div>
            <div class="endpoint">GET /workflows - List all workflows</div>
            <div class="endpoint">GET /instances/{id} - Get workflow instance details</div>
        </div>

        <div class="section">
            <h2>Quick Test</h2>
            <div class="api-example">
                <h4>Trigger Customer Workflow:</h4>
                <code>curl -X POST http://localhost:5000/webhook/customer.created \<br>
                &nbsp;&nbsp;-H "Content-Type: application/json" \<br>
                &nbsp;&nbsp;-d '{"name": "John Doe", "email": "john@example.com", "id": "12345"}'</code>
            </div>
            <div class="api-example">
                <h4>Check Health:</h4>
                <code>curl http://localhost:5000/health</code>
            </div>
        </div>

        <div class="section">
            <h2>Built-in Actions</h2>
            <ul>
                <li><strong>http.request</strong> - Make HTTP requests to external APIs</li>
                <li><strong>shell.exec</strong> - Execute shell commands</li>
                <li><strong>log.info</strong> - Log messages with different levels</li>
            </ul>
        </div>
    </div>
</body>
</html>
`

// DashboardHandler serves the workflow dashboard
type DashboardHandler struct {
	engine *engine.Engine
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(engine *engine.Engine) *DashboardHandler {
	return &DashboardHandler{engine: engine}
}

// ServeHTTP handles dashboard requests
func (d *DashboardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.New("dashboard").Parse(dashboardHTML)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Template execution error", http.StatusInternalServerError)
		return
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
	Workflows int    `json:"workflows"`
}
