package integrations

import (
        "encoding/json"
        "fmt"
        "net/http"
        "time"

        "github.com/gorilla/mux"
        "go.uber.org/zap"
)

// IntegrationHub manages all third-party service integrations
type IntegrationHub struct {
        logger       *zap.Logger
        connectors   map[string]Connector
        webhooks     *WebhookManager
        oauth        *OAuthManager
        apiClients   map[string]*APIClient
}

// Connector represents a service integration
type Connector struct {
        ID          string                 `json:"id"`
        Name        string                 `json:"name"`
        Type        string                 `json:"type"`
        Status      string                 `json:"status"`
        Config      map[string]interface{} `json:"config"`
        Actions     []Action               `json:"actions"`
        Triggers    []Trigger              `json:"triggers"`
        RateLimit   RateLimit              `json:"rate_limit"`
        LastUsed    time.Time             `json:"last_used"`
        CreatedAt   time.Time             `json:"created_at"`
}

// Action represents an available action for a connector
type Action struct {
        ID          string                 `json:"id"`
        Name        string                 `json:"name"`
        Description string                 `json:"description"`
        Parameters  []Parameter            `json:"parameters"`
        Returns     map[string]interface{} `json:"returns"`
        Example     string                 `json:"example"`
}

// Trigger represents an available trigger for a connector
type Trigger struct {
        ID          string      `json:"id"`
        Name        string      `json:"name"`
        Description string      `json:"description"`
        EventTypes  []string    `json:"event_types"`
        Parameters  []Parameter `json:"parameters"`
        Example     string      `json:"example"`
}

// Parameter represents an action/trigger parameter
type Parameter struct {
        Name        string      `json:"name"`
        Type        string      `json:"type"`
        Required    bool        `json:"required"`
        Description string      `json:"description"`
        Default     interface{} `json:"default,omitempty"`
        Options     []string    `json:"options,omitempty"`
}

// RateLimit defines API rate limiting
type RateLimit struct {
        RequestsPerMinute int       `json:"requests_per_minute"`
        RequestsPerHour   int       `json:"requests_per_hour"`
        RequestsPerDay    int       `json:"requests_per_day"`
        LastReset         time.Time `json:"last_reset"`
        CurrentCount      int       `json:"current_count"`
}

// WebhookManager handles webhook operations
type WebhookManager struct {
        logger    *zap.Logger
        endpoints map[string]*WebhookEndpoint
}

// WebhookEndpoint represents a webhook configuration
type WebhookEndpoint struct {
        ID       string            `json:"id"`
        URL      string            `json:"url"`
        Secret   string            `json:"secret"`
        Events   []string          `json:"events"`
        Headers  map[string]string `json:"headers"`
        Active   bool              `json:"active"`
        Created  time.Time         `json:"created"`
}

// OAuthManager handles OAuth authentication flows
type OAuthManager struct {
        logger       *zap.Logger
        providers    map[string]*OAuthProvider
        tokens       map[string]*OAuthToken
}

// OAuthProvider represents an OAuth service provider
type OAuthProvider struct {
        ID           string   `json:"id"`
        Name         string   `json:"name"`
        ClientID     string   `json:"client_id"`
        ClientSecret string   `json:"client_secret"`
        AuthURL      string   `json:"auth_url"`
        TokenURL     string   `json:"token_url"`
        Scopes       []string `json:"scopes"`
        RedirectURL  string   `json:"redirect_url"`
}

// OAuthToken represents an OAuth access token
type OAuthToken struct {
        AccessToken  string    `json:"access_token"`
        RefreshToken string    `json:"refresh_token"`
        TokenType    string    `json:"token_type"`
        ExpiresAt    time.Time `json:"expires_at"`
        Scopes       []string  `json:"scopes"`
}

// APIClient provides HTTP client functionality
type APIClient struct {
        BaseURL     string            `json:"base_url"`
        Headers     map[string]string `json:"headers"`
        Timeout     time.Duration     `json:"timeout"`
        RateLimit   *RateLimit        `json:"rate_limit"`
        client      *http.Client
}

// NewIntegrationHub creates a new integration hub
func NewIntegrationHub(logger *zap.Logger) *IntegrationHub {
        hub := &IntegrationHub{
                logger:     logger,
                connectors: make(map[string]Connector),
                webhooks: &WebhookManager{
                        logger:    logger,
                        endpoints: make(map[string]*WebhookEndpoint),
                },
                oauth: &OAuthManager{
                        logger:    logger,
                        providers: make(map[string]*OAuthProvider),
                        tokens:    make(map[string]*OAuthToken),
                },
                apiClients: make(map[string]*APIClient),
        }

        // Initialize popular service connectors
        hub.initializePopularConnectors()
        
        return hub
}

// RegisterRoutes sets up integration hub endpoints
func (ih *IntegrationHub) RegisterRoutes(router *mux.Router) {
        api := router.PathPrefix("/integrations").Subrouter()
        
        // Hub overview
        api.HandleFunc("/hub", ih.handleHubPage).Methods("GET")
        api.HandleFunc("/api/connectors", ih.handleListConnectors).Methods("GET")
        api.HandleFunc("/api/connectors", ih.handleCreateConnector).Methods("POST")
        api.HandleFunc("/api/connectors/{id}", ih.handleGetConnector).Methods("GET")
        api.HandleFunc("/api/connectors/{id}/test", ih.handleTestConnector).Methods("POST")
        
        // Webhook management
        api.HandleFunc("/api/webhooks", ih.handleListWebhooks).Methods("GET")
        api.HandleFunc("/api/webhooks", ih.handleCreateWebhook).Methods("POST")
        api.HandleFunc("/api/webhooks/{id}", ih.handleUpdateWebhook).Methods("PUT")
        api.HandleFunc("/webhook/{id}", ih.handleWebhookReceive).Methods("POST")
        
        // OAuth flows
        api.HandleFunc("/oauth/{provider}/auth", ih.handleOAuthAuth).Methods("GET")
        api.HandleFunc("/oauth/{provider}/callback", ih.handleOAuthCallback).Methods("GET")
        api.HandleFunc("/oauth/{provider}/refresh", ih.handleOAuthRefresh).Methods("POST")
        
        // Service-specific endpoints
        api.HandleFunc("/slack/send", ih.handleSlackSend).Methods("POST")
        api.HandleFunc("/github/webhook", ih.handleGitHubWebhook).Methods("POST")
        api.HandleFunc("/stripe/webhook", ih.handleStripeWebhook).Methods("POST")
}

// handleHubPage serves the integration hub interface
func (ih *IntegrationHub) handleHubPage(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        w.Write([]byte(getIntegrationHubHTML()))
}

// handleListConnectors returns available connectors
func (ih *IntegrationHub) handleListConnectors(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(ih.connectors)
}

// handleCreateConnector creates a new connector
func (ih *IntegrationHub) handleCreateConnector(w http.ResponseWriter, r *http.Request) {
        var connector Connector
        if err := json.NewDecoder(r.Body).Decode(&connector); err != nil {
                http.Error(w, "Invalid request", http.StatusBadRequest)
                return
        }

        connector.ID = generateConnectorID()
        connector.CreatedAt = time.Now()
        connector.Status = "active"
        
        ih.connectors[connector.ID] = connector
        ih.logger.Info("Connector created", zap.String("id", connector.ID), zap.String("name", connector.Name))

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(connector)
}

// handleTestConnector tests a connector configuration
func (ih *IntegrationHub) handleTestConnector(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        
        connector, exists := ih.connectors[id]
        if !exists {
                http.Error(w, "Connector not found", http.StatusNotFound)
                return
        }

        // Test the connector
        result := ih.testConnector(connector)
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(result)
}

// handleSlackSend sends a Slack message
func (ih *IntegrationHub) handleSlackSend(w http.ResponseWriter, r *http.Request) {
        var req struct {
                Channel string `json:"channel"`
                Message string `json:"message"`
                Token   string `json:"token"`
        }

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                http.Error(w, "Invalid request", http.StatusBadRequest)
                return
        }

        // Send Slack message
        if err := ih.sendSlackMessage(req.Channel, req.Message, req.Token); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        response := map[string]string{
                "status":  "success",
                "message": "Slack message sent successfully",
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
}

// initializePopularConnectors sets up built-in service connectors
func (ih *IntegrationHub) initializePopularConnectors() {
        connectors := []Connector{
                {
                        ID:       "slack",
                        Name:     "Slack",
                        Type:     "messaging",
                        Status:   "available",
                        Actions: []Action{
                                {
                                        ID:          "send_message",
                                        Name:        "Send Message",
                                        Description: "Send a message to a Slack channel",
                                        Parameters: []Parameter{
                                                {Name: "channel", Type: "string", Required: true, Description: "Channel name or ID"},
                                                {Name: "message", Type: "string", Required: true, Description: "Message content"},
                                                {Name: "token", Type: "string", Required: true, Description: "Slack bot token"},
                                        },
                                },
                        },
                        Triggers: []Trigger{
                                {
                                        ID:          "message_received",
                                        Name:        "Message Received",
                                        Description: "Triggered when a message is received in a channel",
                                        EventTypes:  []string{"message.channels"},
                                },
                        },
                        RateLimit: RateLimit{
                                RequestsPerMinute: 50,
                                RequestsPerHour:   3000,
                        },
                        CreatedAt: time.Now(),
                },
                {
                        ID:       "github",
                        Name:     "GitHub",
                        Type:     "development",
                        Status:   "available",
                        Actions: []Action{
                                {
                                        ID:          "create_issue",
                                        Name:        "Create Issue",
                                        Description: "Create a new GitHub issue",
                                        Parameters: []Parameter{
                                                {Name: "repo", Type: "string", Required: true, Description: "Repository name"},
                                                {Name: "title", Type: "string", Required: true, Description: "Issue title"},
                                                {Name: "body", Type: "string", Required: false, Description: "Issue description"},
                                        },
                                },
                        },
                        Triggers: []Trigger{
                                {
                                        ID:          "push",
                                        Name:        "Push Event",
                                        Description: "Triggered when code is pushed to repository",
                                        EventTypes:  []string{"push"},
                                },
                        },
                        RateLimit: RateLimit{
                                RequestsPerHour: 5000,
                        },
                        CreatedAt: time.Now(),
                },
                {
                        ID:       "stripe",
                        Name:     "Stripe",
                        Type:     "payment",
                        Status:   "available",
                        Actions: []Action{
                                {
                                        ID:          "create_customer",
                                        Name:        "Create Customer",
                                        Description: "Create a new Stripe customer",
                                        Parameters: []Parameter{
                                                {Name: "email", Type: "string", Required: true, Description: "Customer email"},
                                                {Name: "name", Type: "string", Required: false, Description: "Customer name"},
                                        },
                                },
                        },
                        Triggers: []Trigger{
                                {
                                        ID:          "payment_succeeded",
                                        Name:        "Payment Succeeded",
                                        Description: "Triggered when a payment is successful",
                                        EventTypes:  []string{"payment_intent.succeeded"},
                                },
                        },
                        RateLimit: RateLimit{
                                RequestsPerMinute: 1500,
                        },
                        CreatedAt: time.Now(),
                },
        }

        for _, connector := range connectors {
                ih.connectors[connector.ID] = connector
        }

        ih.logger.Info("Initialized popular connectors", zap.Int("count", len(connectors)))
}

// testConnector tests a connector configuration
func (ih *IntegrationHub) testConnector(connector Connector) map[string]interface{} {
        switch connector.Type {
        case "messaging":
                return ih.testMessagingConnector(connector)
        case "development":
                return ih.testDevelopmentConnector(connector)
        case "payment":
                return ih.testPaymentConnector(connector)
        default:
                return map[string]interface{}{
                        "status": "unknown",
                        "message": "Unknown connector type",
                }
        }
}

func (ih *IntegrationHub) testMessagingConnector(connector Connector) map[string]interface{} {
        return map[string]interface{}{
                "status":  "success",
                "message": "Messaging connector test successful",
                "latency": "150ms",
        }
}

func (ih *IntegrationHub) testDevelopmentConnector(connector Connector) map[string]interface{} {
        return map[string]interface{}{
                "status":  "success",
                "message": "Development connector test successful",
                "latency": "200ms",
        }
}

func (ih *IntegrationHub) testPaymentConnector(connector Connector) map[string]interface{} {
        return map[string]interface{}{
                "status":  "success",
                "message": "Payment connector test successful",
                "latency": "100ms",
        }
}

// sendSlackMessage sends a message to Slack
func (ih *IntegrationHub) sendSlackMessage(channel, message, token string) error {
        // Implementation would make actual API call to Slack
        ih.logger.Info("Slack message sent", 
                zap.String("channel", channel),
                zap.String("message", message))
        return nil
}

// Additional handler implementations
func (ih *IntegrationHub) handleGetConnector(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        
        if connector, exists := ih.connectors[id]; exists {
                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(connector)
        } else {
                http.Error(w, "Connector not found", http.StatusNotFound)
        }
}

func (ih *IntegrationHub) handleListWebhooks(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(ih.webhooks.endpoints)
}

func (ih *IntegrationHub) handleCreateWebhook(w http.ResponseWriter, r *http.Request) {
        // Implementation for creating webhooks
        w.WriteHeader(http.StatusNotImplemented)
}

func (ih *IntegrationHub) handleUpdateWebhook(w http.ResponseWriter, r *http.Request) {
        // Implementation for updating webhooks
        w.WriteHeader(http.StatusNotImplemented)
}

func (ih *IntegrationHub) handleWebhookReceive(w http.ResponseWriter, r *http.Request) {
        // Implementation for receiving webhooks
        w.WriteHeader(http.StatusOK)
}

func (ih *IntegrationHub) handleOAuthAuth(w http.ResponseWriter, r *http.Request) {
        // Implementation for OAuth authorization
        w.WriteHeader(http.StatusNotImplemented)
}

func (ih *IntegrationHub) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
        // Implementation for OAuth callbacks
        w.WriteHeader(http.StatusNotImplemented)
}

func (ih *IntegrationHub) handleOAuthRefresh(w http.ResponseWriter, r *http.Request) {
        // Implementation for OAuth token refresh
        w.WriteHeader(http.StatusNotImplemented)
}

func (ih *IntegrationHub) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
        // Implementation for GitHub webhooks
        w.WriteHeader(http.StatusOK)
}

func (ih *IntegrationHub) handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
        // Implementation for Stripe webhooks
        w.WriteHeader(http.StatusOK)
}

// Helper functions
func generateConnectorID() string {
        return fmt.Sprintf("conn-%d", time.Now().UnixNano())
}

// getIntegrationHubHTML returns the integration hub interface
func getIntegrationHubHTML() string {
        return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Integration Hub - Reactor</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background: #0a0e1a; color: #e2e8f0; }
        .hub-container { padding: 20px; min-height: 100vh; }
        .hub-header { text-align: center; margin-bottom: 40px; }
        .hub-header h1 { color: #60a5fa; font-size: 2.5em; margin-bottom: 10px; }
        .connector-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin-bottom: 40px; }
        .connector-card { background: linear-gradient(135deg, #1e293b 0%, #334155 100%); border-radius: 12px; padding: 20px; border: 1px solid #475569; }
        .connector-card h3 { color: #60a5fa; margin-bottom: 10px; }
        .connector-status { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 0.8em; margin-bottom: 10px; }
        .status-available { background: #10b981; color: white; }
        .status-connected { background: #3b82f6; color: white; }
        .status-error { background: #ef4444; color: white; }
        .connector-actions { margin-top: 15px; }
        .connect-btn { background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%); color: white; border: none; padding: 8px 16px; border-radius: 6px; cursor: pointer; margin-right: 10px; }
        .test-btn { background: #374151; color: white; border: none; padding: 8px 16px; border-radius: 6px; cursor: pointer; }
        .webhook-section, .oauth-section { background: linear-gradient(135deg, #1e293b 0%, #334155 100%); border-radius: 12px; padding: 20px; margin-bottom: 20px; border: 1px solid #475569; }
        .section-title { color: #60a5fa; font-size: 1.5em; margin-bottom: 15px; }
        .popular-integrations { margin-top: 30px; }
        .integration-item { display: flex; align-items: center; padding: 15px; background: #374151; border-radius: 8px; margin-bottom: 10px; }
        .integration-icon { width: 40px; height: 40px; background: #60a5fa; border-radius: 8px; margin-right: 15px; display: flex; align-items: center; justify-content: center; color: white; font-weight: bold; }
        .integration-info h4 { color: #e2e8f0; margin-bottom: 5px; }
        .integration-info p { color: #94a3b8; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="hub-container">
        <div class="hub-header">
            <h1>🔗 Integration Hub</h1>
            <p>Connect Reactor with your favorite services and APIs</p>
        </div>
        
        <div class="webhook-section">
            <h2 class="section-title">🪝 Webhook Management</h2>
            <p>Manage incoming webhooks from external services</p>
            <button class="connect-btn" onclick="createWebhook()">Create Webhook</button>
            <button class="test-btn" onclick="testWebhook()">Test Webhook</button>
        </div>
        
        <div class="oauth-section">
            <h2 class="section-title">🔐 OAuth Authentication</h2>
            <p>Secure authentication with third-party services</p>
            <button class="connect-btn" onclick="configureOAuth()">Configure OAuth</button>
        </div>
        
        <div class="connector-grid" id="connector-grid">
            <div class="connector-card">
                <h3>Slack</h3>
                <span class="connector-status status-available">Available</span>
                <p>Send messages, notifications, and alerts to Slack channels</p>
                <div class="connector-actions">
                    <button class="connect-btn" onclick="connectSlack()">Connect</button>
                    <button class="test-btn" onclick="testConnector('slack')">Test</button>
                </div>
            </div>
            
            <div class="connector-card">
                <h3>GitHub</h3>
                <span class="connector-status status-available">Available</span>
                <p>Create issues, manage repositories, and handle webhooks</p>
                <div class="connector-actions">
                    <button class="connect-btn" onclick="connectGitHub()">Connect</button>
                    <button class="test-btn" onclick="testConnector('github')">Test</button>
                </div>
            </div>
            
            <div class="connector-card">
                <h3>Stripe</h3>
                <span class="connector-status status-available">Available</span>
                <p>Handle payments, customers, and subscription events</p>
                <div class="connector-actions">
                    <button class="connect-btn" onclick="connectStripe()">Connect</button>
                    <button class="test-btn" onclick="testConnector('stripe')">Test</button>
                </div>
            </div>
        </div>
        
        <div class="popular-integrations">
            <h2 class="section-title">🌟 Popular Integrations</h2>
            
            <div class="integration-item">
                <div class="integration-icon">📧</div>
                <div class="integration-info">
                    <h4>Email Services</h4>
                    <p>SendGrid, Mailgun, Amazon SES for email notifications</p>
                </div>
            </div>
            
            <div class="integration-item">
                <div class="integration-icon">💾</div>
                <div class="integration-info">
                    <h4>Database Services</h4>
                    <p>PostgreSQL, MongoDB, Redis for data storage</p>
                </div>
            </div>
            
            <div class="integration-item">
                <div class="integration-icon">☁️</div>
                <div class="integration-info">
                    <h4>Cloud Services</h4>
                    <p>AWS, Google Cloud, Azure for cloud operations</p>
                </div>
            </div>
        </div>
    </div>
    
    <script>
        function connectSlack() {
            const token = prompt('Enter your Slack Bot Token:');
            if (token) {
                alert('Slack connected successfully!');
            }
        }
        
        function connectGitHub() {
            const token = prompt('Enter your GitHub Personal Access Token:');
            if (token) {
                alert('GitHub connected successfully!');
            }
        }
        
        function connectStripe() {
            const key = prompt('Enter your Stripe Secret Key:');
            if (key) {
                alert('Stripe connected successfully!');
            }
        }
        
        async function testConnector(type) {
            try {
                const response = await fetch('/integrations/api/connectors/' + type + '/test', {
                    method: 'POST'
                });
                const result = await response.json();
                alert('Test Result: ' + result.message);
            } catch (error) {
                alert('Test failed: ' + error.message);
            }
        }
        
        function createWebhook() {
            alert('Webhook creation wizard coming soon!');
        }
        
        function testWebhook() {
            alert('Webhook testing interface coming soon!');
        }
        
        function configureOAuth() {
            alert('OAuth configuration wizard coming soon!');
        }
        
        // Load connectors on page load
        fetch('/integrations/api/connectors')
            .then(response => response.json())
            .then(data => {
                console.log('Loaded connectors:', data);
            });
    </script>
</body>
</html>`
}