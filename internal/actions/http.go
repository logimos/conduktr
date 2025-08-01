package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HTTPAction implements HTTP request actions
type HTTPAction struct {
	logger *zap.Logger
	client *http.Client
}

// NewHTTPAction creates a new HTTP action
func NewHTTPAction(logger *zap.Logger) *HTTPAction {
	return &HTTPAction{
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Execute performs an HTTP request
func (h *HTTPAction) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Parse input parameters
	url, ok := input["url"].(string)
	if !ok || url == "" {
		return nil, fmt.Errorf("url parameter is required")
	}

	method := "GET"
	if m, ok := input["method"].(string); ok {
		method = m
	}

	var body io.Reader
	if bodyData, ok := input["body"]; ok {
		bodyBytes, err := json.Marshal(bodyData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if headers, ok := input["headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", value))
		}
	}

	// Set default content type for POST/PUT requests with body
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	h.logger.Info("Executing HTTP request", 
		zap.String("method", method),
		zap.String("url", url))

	// Perform request
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse JSON response if possible
	var responseData interface{}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &responseData); err != nil {
			// If JSON parsing fails, use raw string
			responseData = string(responseBody)
		}
	}

	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     make(map[string]string),
		"body":        responseData,
		"success":     resp.StatusCode >= 200 && resp.StatusCode < 300,
	}

	// Convert headers to string map
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	result["headers"] = headers

	h.logger.Info("HTTP request completed", 
		zap.String("url", url),
		zap.Int("status_code", resp.StatusCode))

	if !result["success"].(bool) {
		return result, fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	return result, nil
}
