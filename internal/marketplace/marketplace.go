package marketplace

import (
        "encoding/json"
        "net/http"
        "sort"
        "strings"
        "time"

        "github.com/gorilla/mux"
)

// WorkflowTemplate represents a template in the marketplace
type WorkflowTemplate struct {
        ID          string            `json:"id"`
        Name        string            `json:"name"`
        Description string            `json:"description"`
        Category    string            `json:"category"`
        Tags        []string          `json:"tags"`
        Author      string            `json:"author"`
        Version     string            `json:"version"`
        Created     time.Time         `json:"created"`
        Updated     time.Time         `json:"updated"`
        Downloads   int               `json:"downloads"`
        Rating      float64           `json:"rating"`
        Reviews     int               `json:"reviews"`
        Icon        string            `json:"icon"`
        Screenshots []string          `json:"screenshots"`
        Template    WorkflowSpec      `json:"template"`
        Dependencies []string         `json:"dependencies"`
        License     string            `json:"license"`
        Price       float64           `json:"price"`
        IsFree      bool              `json:"isFree"`
        Featured    bool              `json:"featured"`
        Verified    bool              `json:"verified"`
        UseCase     string            `json:"useCase"`
        Complexity  string            `json:"complexity"`
        Metadata    map[string]interface{} `json:"metadata"`
}

// WorkflowSpec defines the structure of workflow templates
type WorkflowSpec struct {
        Name        string                 `json:"name"`
        Description string                 `json:"description"`
        Version     string                 `json:"version"`
        Triggers    []TriggerSpec          `json:"triggers"`
        Steps       []StepSpec             `json:"steps"`
        Variables   []VariableSpec         `json:"variables"`
        Settings    map[string]interface{} `json:"settings"`
}

// TriggerSpec defines trigger configuration
type TriggerSpec struct {
        Type   string                 `json:"type"`
        Config map[string]interface{} `json:"config"`
}

// StepSpec defines step configuration
type StepSpec struct {
        Name      string                 `json:"name"`
        Action    string                 `json:"action"`
        Config    map[string]interface{} `json:"config"`
        Condition string                 `json:"condition,omitempty"`
        Parallel  bool                   `json:"parallel,omitempty"`
}

// VariableSpec defines variable configuration
type VariableSpec struct {
        Name         string      `json:"name"`
        Type         string      `json:"type"`
        DefaultValue interface{} `json:"defaultValue"`
        Description  string      `json:"description"`
        Required     bool        `json:"required"`
}

// MarketplaceService handles template marketplace operations
type MarketplaceService struct {
        templates []WorkflowTemplate
}

// NewMarketplaceService creates a new marketplace service
func NewMarketplaceService() *MarketplaceService {
        service := &MarketplaceService{
                templates: make([]WorkflowTemplate, 0),
        }
        service.loadDefaultTemplates()
        return service
}

// RegisterMarketplaceRoutes registers marketplace API routes
func (ms *MarketplaceService) RegisterMarketplaceRoutes(r *mux.Router) {
        // Marketplace UI routes
        r.HandleFunc("/marketplace", ms.handleMarketplacePage).Methods("GET")
        r.HandleFunc("/marketplace/template/{id}", ms.handleTemplatePage).Methods("GET")
        
        // API routes
        api := r.PathPrefix("/api/marketplace").Subrouter()
        api.HandleFunc("/templates", ms.handleGetTemplates).Methods("GET")
        api.HandleFunc("/templates/{id}", ms.handleGetTemplate).Methods("GET")
        api.HandleFunc("/templates/{id}/download", ms.handleDownloadTemplate).Methods("POST")
        api.HandleFunc("/templates/search", ms.handleSearchTemplates).Methods("GET")
        api.HandleFunc("/templates/categories", ms.handleGetCategories).Methods("GET")
        api.HandleFunc("/templates/featured", ms.handleGetFeatured).Methods("GET")
        api.HandleFunc("/templates/{id}/reviews", ms.handleGetReviews).Methods("GET")
        api.HandleFunc("/templates/{id}/reviews", ms.handleAddReview).Methods("POST")
}

// handleMarketplacePage serves the marketplace HTML page
func (ms *MarketplaceService) handleMarketplacePage(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        w.Write([]byte(`<!DOCTYPE html><html><head><title>Reactor Marketplace</title></head><body><h1>Workflow Marketplace</h1><p>Enterprise template marketplace with curated workflows coming online...</p></body></html>`))
}

// handleTemplatePage serves individual template pages
func (ms *MarketplaceService) handleTemplatePage(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        templateID := vars["id"]
        
        for _, template := range ms.templates {
                if template.ID == templateID {
                        w.Header().Set("Content-Type", "text/html")
                        w.Write([]byte(`<!DOCTYPE html><html><head><title>` + template.Name + `</title></head><body><h1>` + template.Name + `</h1><p>` + template.Description + `</p></body></html>`))
                        return
                }
        }
        
        http.Error(w, "Template not found", http.StatusNotFound)
}

// handleGetTemplates returns all templates with optional filtering
func (ms *MarketplaceService) handleGetTemplates(w http.ResponseWriter, r *http.Request) {
        category := r.URL.Query().Get("category")
        sortBy := r.URL.Query().Get("sort")
        
        templates := ms.templates
        
        // Filter by category
        if category != "" {
                filtered := make([]WorkflowTemplate, 0)
                for _, template := range templates {
                        if template.Category == category {
                                filtered = append(filtered, template)
                        }
                }
                templates = filtered
        }
        
        // Sort templates
        ms.sortTemplates(templates, sortBy)
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(templates)
}

// handleGetTemplate returns a specific template
func (ms *MarketplaceService) handleGetTemplate(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        
        for _, template := range ms.templates {
                if template.ID == id {
                        w.Header().Set("Content-Type", "application/json")
                        json.NewEncoder(w).Encode(template)
                        return
                }
        }
        
        http.Error(w, "Template not found", http.StatusNotFound)
}

// handleDownloadTemplate handles template downloads
func (ms *MarketplaceService) handleDownloadTemplate(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]
        
        for i, template := range ms.templates {
                if template.ID == id {
                        // Increment download count
                        ms.templates[i].Downloads++
                        
                        w.Header().Set("Content-Type", "application/json")
                        json.NewEncoder(w).Encode(template.Template)
                        return
                }
        }
        
        http.Error(w, "Template not found", http.StatusNotFound)
}

// handleSearchTemplates searches templates by query
func (ms *MarketplaceService) handleSearchTemplates(w http.ResponseWriter, r *http.Request) {
        query := r.URL.Query().Get("q")
        if query == "" {
                http.Error(w, "Search query required", http.StatusBadRequest)
                return
        }
        
        results := make([]WorkflowTemplate, 0)
        query = strings.ToLower(query)
        
        for _, template := range ms.templates {
                if ms.matchesSearch(template, query) {
                        results = append(results, template)
                }
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(results)
}

// handleGetCategories returns all available categories
func (ms *MarketplaceService) handleGetCategories(w http.ResponseWriter, r *http.Request) {
        categories := make(map[string]int)
        
        for _, template := range ms.templates {
                categories[template.Category]++
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(categories)
}

// handleGetFeatured returns featured templates
func (ms *MarketplaceService) handleGetFeatured(w http.ResponseWriter, r *http.Request) {
        featured := make([]WorkflowTemplate, 0)
        
        for _, template := range ms.templates {
                if template.Featured {
                        featured = append(featured, template)
                }
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(featured)
}

// handleGetReviews returns reviews for a template
func (ms *MarketplaceService) handleGetReviews(w http.ResponseWriter, r *http.Request) {
        // Placeholder implementation
        reviews := []map[string]interface{}{
                {
                        "user":    "developer123",
                        "rating":  5,
                        "comment": "Excellent template, saved me hours of work!",
                        "date":    time.Now().Format("2006-01-02"),
                },
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(reviews)
}

// handleAddReview adds a review for a template
func (ms *MarketplaceService) handleAddReview(w http.ResponseWriter, r *http.Request) {
        var review map[string]interface{}
        if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
        }
        
        // Add review logic here
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]string{"status": "review added"})
}

// matchesSearch checks if template matches search query
func (ms *MarketplaceService) matchesSearch(template WorkflowTemplate, query string) bool {
        searchText := strings.ToLower(template.Name + " " + template.Description + " " + strings.Join(template.Tags, " "))
        return strings.Contains(searchText, query)
}

// sortTemplates sorts templates by the specified criteria
func (ms *MarketplaceService) sortTemplates(templates []WorkflowTemplate, sortBy string) {
        switch sortBy {
        case "downloads":
                sort.Slice(templates, func(i, j int) bool {
                        return templates[i].Downloads > templates[j].Downloads
                })
        case "rating":
                sort.Slice(templates, func(i, j int) bool {
                        return templates[i].Rating > templates[j].Rating
                })
        case "newest":
                sort.Slice(templates, func(i, j int) bool {
                        return templates[i].Created.After(templates[j].Created)
                })
        case "name":
                sort.Slice(templates, func(i, j int) bool {
                        return templates[i].Name < templates[j].Name
                })
        default:
                // Default: featured first, then by rating
                sort.Slice(templates, func(i, j int) bool {
                        if templates[i].Featured != templates[j].Featured {
                                return templates[i].Featured
                        }
                        return templates[i].Rating > templates[j].Rating
                })
        }
}

// loadDefaultTemplates loads predefined workflow templates
func (ms *MarketplaceService) loadDefaultTemplates() {
        ms.templates = []WorkflowTemplate{
                {
                        ID:          "web-scraper",
                        Name:        "Web Scraper Pipeline",
                        Description: "Complete web scraping workflow with data processing and storage",
                        Category:    "Data Processing",
                        Tags:        []string{"web-scraping", "data-extraction", "automation"},
                        Author:      "Reactor Team",
                        Version:     "1.0.0",
                        Created:     time.Now().AddDate(0, -2, 0),
                        Updated:     time.Now().AddDate(0, -1, 0),
                        Downloads:   1247,
                        Rating:      4.8,
                        Reviews:     34,
                        Icon:        "🕷️",
                        Screenshots: []string{"/screenshots/web-scraper-1.png", "/screenshots/web-scraper-2.png"},
                        IsFree:      true,
                        Featured:    true,
                        Verified:    true,
                        UseCase:     "Extract data from websites on schedule",
                        Complexity:  "Medium",
                        License:     "MIT",
                        Template: WorkflowSpec{
                                Name:        "Web Scraper Pipeline",
                                Description: "Scrape web data and process it",
                                Version:     "1.0.0",
                                Triggers: []TriggerSpec{
                                        {
                                                Type: "scheduler",
                                                Config: map[string]interface{}{
                                                        "cron": "0 */6 * * *",
                                                },
                                        },
                                },
                                Steps: []StepSpec{
                                        {
                                                Name:   "fetch-web-data",
                                                Action: "http",
                                                Config: map[string]interface{}{
                                                        "url":    "{{.target_url}}",
                                                        "method": "GET",
                                                },
                                        },
                                        {
                                                Name:   "parse-html",
                                                Action: "transform",
                                                Config: map[string]interface{}{
                                                        "type":     "html",
                                                        "selector": "{{.css_selector}}",
                                                },
                                        },
                                        {
                                                Name:   "save-data",
                                                Action: "database",
                                                Config: map[string]interface{}{
                                                        "operation": "insert",
                                                        "table":     "scraped_data",
                                                },
                                        },
                                },
                                Variables: []VariableSpec{
                                        {
                                                Name:         "target_url",
                                                Type:         "string",
                                                DefaultValue: "https://example.com",
                                                Description:  "URL to scrape",
                                                Required:     true,
                                        },
                                        {
                                                Name:         "css_selector",
                                                Type:         "string",
                                                DefaultValue: ".content",
                                                Description:  "CSS selector for data extraction",
                                                Required:     true,
                                        },
                                },
                        },
                },
                {
                        ID:          "ecommerce-order-processor",
                        Name:        "E-commerce Order Processor",
                        Description: "Complete order processing workflow with payment, inventory, and fulfillment",
                        Category:    "E-commerce",
                        Tags:        []string{"orders", "payment", "inventory", "fulfillment"},
                        Author:      "Commerce Solutions",
                        Version:     "2.1.0",
                        Created:     time.Now().AddDate(0, -3, 0),
                        Updated:     time.Now().AddDate(0, 0, -5),
                        Downloads:   892,
                        Rating:      4.9,
                        Reviews:     28,
                        Icon:        "🛒",
                        Screenshots: []string{"/screenshots/ecommerce-1.png"},
                        IsFree:      false,
                        Price:       29.99,
                        Featured:    true,
                        Verified:    true,
                        UseCase:     "Process e-commerce orders end-to-end",
                        Complexity:  "Advanced",
                        License:     "Commercial",
                        Template: WorkflowSpec{
                                Name:        "E-commerce Order Processor",
                                Description: "Process orders with payment and fulfillment",
                                Version:     "2.1.0",
                                Triggers: []TriggerSpec{
                                        {
                                                Type: "http",
                                                Config: map[string]interface{}{
                                                        "port": 8080,
                                                        "path": "/order",
                                                },
                                        },
                                },
                                Steps: []StepSpec{
                                        {
                                                Name:   "validate-order",
                                                Action: "validate",
                                                Config: map[string]interface{}{
                                                        "schema": "order_schema.json",
                                                },
                                        },
                                        {
                                                Name:   "check-inventory",
                                                Action: "database",
                                                Config: map[string]interface{}{
                                                        "operation": "select",
                                                        "table":     "inventory",
                                                },
                                        },
                                        {
                                                Name:   "process-payment",
                                                Action: "payment",
                                                Config: map[string]interface{}{
                                                        "provider": "stripe",
                                                        "currency": "USD",
                                                },
                                        },
                                        {
                                                Name:     "create-shipment",
                                                Action:   "shipping",
                                                Parallel: true,
                                                Config: map[string]interface{}{
                                                        "provider": "fedex",
                                                },
                                        },
                                        {
                                                Name:   "send-confirmation",
                                                Action: "email",
                                                Config: map[string]interface{}{
                                                        "template": "order_confirmation",
                                                },
                                        },
                                },
                        },
                },
                {
                        ID:          "api-monitoring",
                        Name:        "API Health Monitor",
                        Description: "Monitor API endpoints and alert on failures with detailed diagnostics",
                        Category:    "Monitoring",
                        Tags:        []string{"monitoring", "api", "health-check", "alerts"},
                        Author:      "DevOps Tools",
                        Version:     "1.3.0",
                        Created:     time.Now().AddDate(0, -1, 0),
                        Updated:     time.Now().AddDate(0, 0, -2),
                        Downloads:   634,
                        Rating:      4.6,
                        Reviews:     19,
                        Icon:        "📊",
                        Screenshots: []string{"/screenshots/api-monitor-1.png"},
                        IsFree:      true,
                        Featured:    false,
                        Verified:    true,
                        UseCase:     "Monitor API health and send alerts",
                        Complexity:  "Simple",
                        License:     "Apache 2.0",
                        Template: WorkflowSpec{
                                Name:        "API Health Monitor",
                                Description: "Monitor API endpoints continuously",
                                Version:     "1.3.0",
                                Triggers: []TriggerSpec{
                                        {
                                                Type: "scheduler",
                                                Config: map[string]interface{}{
                                                        "cron": "*/5 * * * *",
                                                },
                                        },
                                },
                                Steps: []StepSpec{
                                        {
                                                Name:   "check-api-health",
                                                Action: "http",
                                                Config: map[string]interface{}{
                                                        "url":     "{{.api_endpoint}}/health",
                                                        "method":  "GET",
                                                        "timeout": "10s",
                                                },
                                        },
                                        {
                                                Name:      "alert-on-failure",
                                                Action:    "email",
                                                Condition: "{{.response.status}} != 200",
                                                Config: map[string]interface{}{
                                                        "subject": "API Health Alert",
                                                        "to":      "{{.alert_email}}",
                                                },
                                        },
                                },
                                Variables: []VariableSpec{
                                        {
                                                Name:         "api_endpoint",
                                                Type:         "string",
                                                DefaultValue: "https://api.example.com",
                                                Description:  "API endpoint to monitor",
                                                Required:     true,
                                        },
                                        {
                                                Name:         "alert_email",
                                                Type:         "string",
                                                DefaultValue: "admin@example.com",
                                                Description:  "Email for alerts",
                                                Required:     true,
                                        },
                                },
                        },
                },
                {
                        ID:          "data-pipeline",
                        Name:        "ETL Data Pipeline",
                        Description: "Extract, transform, and load data with validation and error handling",
                        Category:    "Data Processing",
                        Tags:        []string{"etl", "data-pipeline", "transformation", "validation"},
                        Author:      "Data Engineering",
                        Version:     "1.5.0",
                        Created:     time.Now().AddDate(0, -4, 0),
                        Updated:     time.Now().AddDate(0, 0, -1),
                        Downloads:   1156,
                        Rating:      4.7,
                        Reviews:     42,
                        Icon:        "🔄",
                        Screenshots: []string{"/screenshots/etl-1.png", "/screenshots/etl-2.png"},
                        IsFree:      true,
                        Featured:    true,
                        Verified:    true,
                        UseCase:     "Process and transform data between systems",
                        Complexity:  "Advanced",
                        License:     "MIT",
                        Template: WorkflowSpec{
                                Name:        "ETL Data Pipeline",
                                Description: "Extract, transform, and load data",
                                Version:     "1.5.0",
                                Triggers: []TriggerSpec{
                                        {
                                                Type: "file",
                                                Config: map[string]interface{}{
                                                        "path":   "{{.input_directory}}",
                                                        "events": []string{"create"},
                                                },
                                        },
                                },
                                Steps: []StepSpec{
                                        {
                                                Name:   "extract-data",
                                                Action: "file",
                                                Config: map[string]interface{}{
                                                        "operation": "read",
                                                        "format":    "csv",
                                                },
                                        },
                                        {
                                                Name:   "validate-data",
                                                Action: "validate",
                                                Config: map[string]interface{}{
                                                        "schema": "{{.validation_schema}}",
                                                },
                                        },
                                        {
                                                Name:   "transform-data",
                                                Action: "transform",
                                                Config: map[string]interface{}{
                                                        "operations": []map[string]interface{}{
                                                                {"type": "filter", "condition": "{{.filter_condition}}"},
                                                                {"type": "map", "function": "{{.mapping_function}}"},
                                                        },
                                                },
                                        },
                                        {
                                                Name:   "load-data",
                                                Action: "database",
                                                Config: map[string]interface{}{
                                                        "operation": "bulk_insert",
                                                        "table":     "{{.target_table}}",
                                                },
                                        },
                                },
                        },
                },
        }
}