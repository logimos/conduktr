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
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`
	Tags         []string               `json:"tags"`
	Author       string                 `json:"author"`
	Version      string                 `json:"version"`
	Created      time.Time              `json:"created"`
	Updated      time.Time              `json:"updated"`
	Downloads    int                    `json:"downloads"`
	Rating       float64                `json:"rating"`
	Reviews      int                    `json:"reviews"`
	Icon         string                 `json:"icon"`
	Screenshots  []string               `json:"screenshots"`
	Template     WorkflowSpec           `json:"template"`
	Dependencies []string               `json:"dependencies"`
	License      string                 `json:"license"`
	Price        float64                `json:"price"`
	IsFree       bool                   `json:"isFree"`
	Featured     bool                   `json:"featured"`
	Verified     bool                   `json:"verified"`
	UseCase      string                 `json:"useCase"`
	Complexity   string                 `json:"complexity"`
	Metadata     map[string]interface{} `json:"metadata"`
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
	w.Write([]byte(getMarketplaceHTML()))
}

func getMarketplaceHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reactor Marketplace - Workflow Templates</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; 
            background: linear-gradient(135deg, #0f1419 0%, #1a1f2e 100%);
            color: #e2e8f0; 
            line-height: 1.6;
        }
        .marketplace-container { 
            max-width: 1400px; 
            margin: 0 auto; 
            padding: 20px; 
        }
        .header { 
            text-align: center; 
            margin-bottom: 40px; 
            padding: 40px 0;
        }
        .header h1 { 
            color: #4fd1c7; 
            font-size: 3rem; 
            margin-bottom: 10px;
            text-shadow: 0 0 20px rgba(79, 209, 199, 0.3);
        }
        .header p { 
            color: #94a3b8; 
            font-size: 1.2rem; 
        }
        .search-bar { 
            background: #1a1f2e; 
            border: 1px solid #2d3748; 
            border-radius: 12px; 
            padding: 20px; 
            margin-bottom: 30px;
            display: flex;
            gap: 15px;
            align-items: center;
        }
        .search-input { 
            flex: 1; 
            background: #2d3748; 
            border: 1px solid #4a5568; 
            border-radius: 8px; 
            padding: 12px 16px; 
            color: #e2e8f0; 
            font-size: 1rem;
        }
        .search-input:focus { 
            outline: none; 
            border-color: #4fd1c7; 
            box-shadow: 0 0 0 3px rgba(79, 209, 199, 0.1);
        }
        .filter-select { 
            background: #2d3748; 
            border: 1px solid #4a5568; 
            border-radius: 8px; 
            padding: 12px 16px; 
            color: #e2e8f0; 
            font-size: 1rem;
        }
        .templates-grid { 
            display: grid; 
            grid-template-columns: repeat(auto-fill, minmax(350px, 1fr)); 
            gap: 25px; 
            margin-top: 30px;
        }
        .template-card { 
            background: linear-gradient(135deg, #1a1f2e 0%, #2d3748 100%);
            border: 1px solid #4a5568; 
            border-radius: 16px; 
            padding: 25px; 
            transition: all 0.3s ease;
            position: relative;
            overflow: hidden;
        }
        .template-card:hover { 
            transform: translateY(-5px); 
            border-color: #4fd1c7; 
            box-shadow: 0 10px 30px rgba(79, 209, 199, 0.2);
        }
        .template-icon { 
            font-size: 2.5rem; 
            margin-bottom: 15px; 
            display: block;
        }
        .template-title { 
            color: #4fd1c7; 
            font-size: 1.3rem; 
            font-weight: 600; 
            margin-bottom: 10px;
        }
        .template-description { 
            color: #94a3b8; 
            margin-bottom: 15px; 
            line-height: 1.5;
        }
        .template-meta { 
            display: flex; 
            justify-content: space-between; 
            align-items: center; 
            margin-bottom: 20px;
        }
        .template-rating { 
            color: #fbbf24; 
            font-weight: 600;
        }
        .template-downloads { 
            color: #60a5fa; 
            font-size: 0.9rem;
        }
        .template-tags { 
            display: flex; 
            flex-wrap: wrap; 
            gap: 8px; 
            margin-bottom: 20px;
        }
        .tag { 
            background: #374151; 
            color: #d1d5db; 
            padding: 4px 12px; 
            border-radius: 20px; 
            font-size: 0.8rem;
        }
        .template-actions { 
            display: flex; 
            gap: 10px;
        }
        .btn { 
            padding: 10px 20px; 
            border-radius: 8px; 
            font-weight: 600; 
            cursor: pointer; 
            border: none; 
            transition: all 0.3s ease;
        }
        .btn-primary { 
            background: linear-gradient(135deg, #4fd1c7 0%, #38b2ac 100%); 
            color: white;
        }
        .btn-primary:hover { 
            transform: translateY(-2px); 
            box-shadow: 0 5px 15px rgba(79, 209, 199, 0.4);
        }
        .btn-secondary { 
            background: transparent; 
            color: #4fd1c7; 
            border: 1px solid #4fd1c7;
        }
        .btn-secondary:hover { 
            background: #4fd1c7; 
            color: white;
        }
        .featured-badge { 
            position: absolute; 
            top: 15px; 
            right: 15px; 
            background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%); 
            color: white; 
            padding: 4px 12px; 
            border-radius: 20px; 
            font-size: 0.8rem; 
            font-weight: 600;
        }
        .free-badge { 
            background: linear-gradient(135deg, #10b981 0%, #059669 100%); 
            color: white; 
            padding: 4px 12px; 
            border-radius: 20px; 
            font-size: 0.8rem; 
            font-weight: 600;
            margin-left: 10px;
        }
        .categories { 
            display: flex; 
            gap: 15px; 
            margin-bottom: 30px; 
            flex-wrap: wrap;
        }
        .category-btn { 
            background: #2d3748; 
            color: #e2e8f0; 
            border: 1px solid #4a5568; 
            border-radius: 25px; 
            padding: 10px 20px; 
            cursor: pointer; 
            transition: all 0.3s ease;
        }
        .category-btn.active { 
            background: #4fd1c7; 
            color: white; 
            border-color: #4fd1c7;
        }
        .loading { 
            text-align: center; 
            padding: 40px; 
            color: #94a3b8;
        }
        .stats-bar { 
            background: #1a1f2e; 
            border: 1px solid #2d3748; 
            border-radius: 12px; 
            padding: 20px; 
            margin-bottom: 30px;
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
        }
        .stat-item { 
            text-align: center;
        }
        .stat-number { 
            color: #4fd1c7; 
            font-size: 2rem; 
            font-weight: 600;
        }
        .stat-label { 
            color: #94a3b8; 
            font-size: 0.9rem;
        }
    </style>
</head>
<body>
    <div class="marketplace-container">
        <div class="header">
            <h1>🚀 Reactor Marketplace</h1>
            <p>Discover and deploy powerful workflow templates</p>
        </div>
        
        <div class="stats-bar">
            <div class="stat-item">
                <div class="stat-number" id="total-templates">0</div>
                <div class="stat-label">Templates</div>
            </div>
            <div class="stat-item">
                <div class="stat-number" id="total-downloads">0</div>
                <div class="stat-label">Downloads</div>
            </div>
            <div class="stat-item">
                <div class="stat-number" id="avg-rating">0.0</div>
                <div class="stat-label">Avg Rating</div>
            </div>
            <div class="stat-item">
                <div class="stat-number" id="free-templates">0</div>
                <div class="stat-label">Free Templates</div>
            </div>
        </div>
        
        <div class="search-bar">
            <input type="text" class="search-input" id="search-input" placeholder="Search templates...">
            <select class="filter-select" id="category-filter">
                <option value="">All Categories</option>
                <option value="Data Processing">Data Processing</option>
                <option value="E-commerce">E-commerce</option>
                <option value="User Management">User Management</option>
                <option value="File Management">File Management</option>
                <option value="Communication">Communication</option>
                <option value="Monitoring">Monitoring</option>
            </select>
            <select class="filter-select" id="complexity-filter">
                <option value="">All Complexity</option>
                <option value="Simple">Simple</option>
                <option value="Medium">Medium</option>
                <option value="Complex">Complex</option>
            </select>
        </div>
        
        <div class="categories">
            <button class="category-btn active" data-category="">All</button>
            <button class="category-btn" data-category="featured">Featured</button>
            <button class="category-btn" data-category="free">Free</button>
            <button class="category-btn" data-category="new">New</button>
            <button class="category-btn" data-category="popular">Popular</button>
        </div>
        
        <div class="templates-grid" id="templates-grid">
            <div class="loading">Loading templates...</div>
        </div>
    </div>
    
    <script>
        // Load templates on page load
        document.addEventListener('DOMContentLoaded', function() {
            loadTemplates();
            loadStats();
        });
        
        // Search and filter functionality
        document.getElementById('search-input').addEventListener('input', debounce(filterTemplates, 300));
        document.getElementById('category-filter').addEventListener('change', filterTemplates);
        document.getElementById('complexity-filter').addEventListener('change', filterTemplates);
        
        // Category buttons
        document.querySelectorAll('.category-btn').forEach(btn => {
            btn.addEventListener('click', function() {
                document.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
                this.classList.add('active');
                filterTemplates();
            });
        });
        
        async function loadTemplates() {
            try {
                const response = await fetch('/api/marketplace/templates');
                const templates = await response.json();
                displayTemplates(templates);
            } catch (error) {
                document.getElementById('templates-grid').innerHTML = 
                    '<div class="loading">Error loading templates. Please try again.</div>';
            }
        }
        
        async function loadStats() {
            try {
                const response = await fetch('/api/marketplace/stats');
                const stats = await response.json();
                document.getElementById('total-templates').textContent = stats.total_templates || 0;
                document.getElementById('total-downloads').textContent = stats.total_downloads || 0;
                document.getElementById('avg-rating').textContent = (stats.avg_rating || 0).toFixed(1);
                document.getElementById('free-templates').textContent = stats.free_templates || 0;
            } catch (error) {
                console.error('Error loading stats:', error);
            }
        }
        
        function displayTemplates(templates) {
            const grid = document.getElementById('templates-grid');
            grid.innerHTML = '';
            
            templates.forEach(template => {
                const card = createTemplateCard(template);
                grid.appendChild(card);
            });
        }
        
        function createTemplateCard(template) {
            const card = document.createElement('div');
            card.className = 'template-card';
            
            const featuredBadge = template.featured ? '<div class="featured-badge">⭐ Featured</div>' : '';
            const freeBadge = template.isFree ? '<span class="free-badge">Free</span>' : '';
            
            card.innerHTML = featuredBadge +
                '<div class="template-icon">' + template.icon + '</div>' +
                '<div class="template-title">' + template.name + '</div>' +
                '<div class="template-description">' + template.description + '</div>' +
                '<div class="template-meta">' +
                    '<div class="template-rating">★ ' + template.rating + '</div>' +
                    '<div class="template-downloads">📥 ' + template.downloads + ' downloads</div>' +
                '</div>' +
                '<div class="template-tags">' +
                    template.tags.map(function(tag) { return '<span class="tag">' + tag + '</span>'; }).join('') +
                '</div>' +
                '<div class="template-actions">' +
                    '<button class="btn btn-primary" onclick="downloadTemplate(\'' + template.id + '\')">' +
                        'Download ' + freeBadge +
                    '</button>' +
                    '<button class="btn btn-secondary" onclick="viewTemplate(\'' + template.id + '\')">' +
                        'View Details' +
                    '</button>' +
                '</div>';
            
            return card;
        }
        
        function filterTemplates() {
            const searchTerm = document.getElementById('search-input').value.toLowerCase();
            const category = document.getElementById('category-filter').value;
            const complexity = document.getElementById('complexity-filter').value;
            const activeCategory = document.querySelector('.category-btn.active').dataset.category;
            
            // Implement filtering logic here
            loadTemplates(); // For now, just reload
        }
        
        async function downloadTemplate(templateId) {
            try {
                const response = await fetch('/api/marketplace/templates/' + templateId + '/download', {
                    method: 'POST'
                });
                
                if (response.ok) {
                    alert('Template downloaded successfully!');
                } else {
                    alert('Error downloading template');
                }
            } catch (error) {
                alert('Error downloading template');
            }
        }
        
        function viewTemplate(templateId) {
            window.location.href = '/marketplace/template/' + templateId;
        }
        
        function debounce(func, wait) {
            let timeout;
            return function executedFunction(...args) {
                const later = () => {
                    clearTimeout(timeout);
                    func(...args);
                };
                clearTimeout(timeout);
                timeout = setTimeout(later, wait);
            };
        }
    </script>
</body>
</html>`
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
			UseCase:     "Process online orders automatically",
			Complexity:  "Complex",
			License:     "Commercial",
			Template: WorkflowSpec{
				Name:        "E-commerce Order Processor",
				Description: "Handle order processing and fulfillment",
				Version:     "2.1.0",
				Triggers: []TriggerSpec{
					{
						Type: "http",
						Config: map[string]interface{}{
							"path": "/webhook/order-created",
						},
					},
				},
				Steps: []StepSpec{
					{
						Name:   "validate-order",
						Action: "validate",
						Config: map[string]interface{}{
							"type": "order",
						},
					},
					{
						Name:   "process-payment",
						Action: "http",
						Config: map[string]interface{}{
							"url":    "{{.payment_gateway}}",
							"method": "POST",
						},
					},
					{
						Name:   "update-inventory",
						Action: "database",
						Config: map[string]interface{}{
							"operation": "update",
							"table":     "inventory",
						},
					},
					{
						Name:   "send-confirmation",
						Action: "email",
						Config: map[string]interface{}{
							"template": "order-confirmation",
						},
					},
					{
						Name:   "notify-fulfillment",
						Action: "notification",
						Config: map[string]interface{}{
							"channel": "fulfillment-team",
						},
					},
				},
			},
		},
		{
			ID:          "user-onboarding",
			Name:        "User Onboarding Flow",
			Description: "Complete user registration and onboarding process",
			Category:    "User Management",
			Tags:        []string{"user", "onboarding", "registration", "welcome"},
			Author:      "Reactor Team",
			Version:     "1.2.0",
			Created:     time.Now().AddDate(0, -1, 0),
			Updated:     time.Now().AddDate(0, 0, -2),
			Downloads:   567,
			Rating:      4.7,
			Reviews:     23,
			Icon:        "👤",
			Screenshots: []string{"/screenshots/onboarding-1.png"},
			IsFree:      true,
			Featured:    false,
			Verified:    true,
			UseCase:     "Handle new user registration",
			Complexity:  "Simple",
			License:     "MIT",
			Template: WorkflowSpec{
				Name:        "User Onboarding Flow",
				Description: "Welcome new users and set up their accounts",
				Version:     "1.2.0",
				Triggers: []TriggerSpec{
					{
						Type: "http",
						Config: map[string]interface{}{
							"path": "/webhook/user-created",
						},
					},
				},
				Steps: []StepSpec{
					{
						Name:   "send-welcome-email",
						Action: "email",
						Config: map[string]interface{}{
							"template": "welcome-email",
						},
					},
					{
						Name:   "create-user-profile",
						Action: "database",
						Config: map[string]interface{}{
							"operation": "insert",
							"table":     "user_profiles",
						},
					},
					{
						Name:   "send-setup-guide",
						Action: "email",
						Config: map[string]interface{}{
							"template": "setup-guide",
						},
					},
				},
			},
		},
		{
			ID:          "api-monitor",
			Name:        "API Health Monitor",
			Description: "Monitor API endpoints and alert on failures",
			Category:    "Monitoring",
			Tags:        []string{"monitoring", "api", "health", "alert"},
			Author:      "DevOps Pro",
			Version:     "1.0.0",
			Created:     time.Now().AddDate(0, -2, 0),
			Updated:     time.Now().AddDate(0, -1, 0),
			Downloads:   445,
			Rating:      4.6,
			Reviews:     18,
			Icon:        "🔍",
			Screenshots: []string{"/screenshots/api-monitor-1.png"},
			IsFree:      true,
			Featured:    false,
			Verified:    true,
			UseCase:     "Monitor API health and performance",
			Complexity:  "Medium",
			License:     "MIT",
			Template: WorkflowSpec{
				Name:        "API Health Monitor",
				Description: "Check API endpoints and alert on issues",
				Version:     "1.0.0",
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
							"url":    "{{.api_endpoint}}/health",
							"method": "GET",
						},
					},
					{
						Name:   "validate-response",
						Action: "validate",
						Config: map[string]interface{}{
							"type": "api_response",
						},
					},
					{
						Name:   "alert-if-down",
						Action: "notification",
						Config: map[string]interface{}{
							"channel":   "ops-team",
							"condition": "{{.response.status != 200}}",
						},
					},
				},
			},
		},
		{
			ID:          "file-processor",
			Name:        "File Processing Pipeline",
			Description: "Process uploaded files with validation and transformation",
			Category:    "File Management",
			Tags:        []string{"file", "upload", "processing", "validation"},
			Author:      "Data Team",
			Version:     "1.1.0",
			Created:     time.Now().AddDate(0, -1, 0),
			Updated:     time.Now().AddDate(0, 0, -1),
			Downloads:   334,
			Rating:      4.5,
			Reviews:     15,
			Icon:        "📁",
			Screenshots: []string{"/screenshots/file-processor-1.png"},
			IsFree:      true,
			Featured:    false,
			Verified:    true,
			UseCase:     "Process uploaded files automatically",
			Complexity:  "Medium",
			License:     "MIT",
			Template: WorkflowSpec{
				Name:        "File Processing Pipeline",
				Description: "Validate and process uploaded files",
				Version:     "1.1.0",
				Triggers: []TriggerSpec{
					{
						Type: "file",
						Config: map[string]interface{}{
							"path": "/uploads",
						},
					},
				},
				Steps: []StepSpec{
					{
						Name:   "validate-file",
						Action: "validate",
						Config: map[string]interface{}{
							"type": "file",
						},
					},
					{
						Name:   "process-content",
						Action: "transform",
						Config: map[string]interface{}{
							"type": "file_content",
						},
					},
					{
						Name:   "store-result",
						Action: "database",
						Config: map[string]interface{}{
							"operation": "insert",
							"table":     "processed_files",
						},
					},
				},
			},
		},
	}
}
