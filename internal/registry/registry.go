package registry

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"sync"

	"github.com/siriusfreak/openapi-mcp-server/internal/models"
	"github.com/siriusfreak/openapi-mcp-server/internal/parser"
)

// Registry manages loaded API specifications
type Registry struct {
	mu     sync.RWMutex
	apis   map[string]*models.APIInfo
	logger *slog.Logger
}

// New creates a new API registry
func New(logger *slog.Logger) *Registry {
	return &Registry{
		apis:   make(map[string]*models.APIInfo),
		logger: logger,
	}
}

// LoadSpec loads an OpenAPI specification from file or URL
func (r *Registry) LoadSpec(path string) error {
	loader := parser.NewLoader(r.logger)
	apiInfo, err := loader.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load spec: %w", err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate unique ID
	id := r.generateID(apiInfo.Name)
	apiInfo.ID = id
	r.apis[id] = apiInfo

	r.logger.Info("Registered API",
		"id", id,
		"name", apiInfo.Name,
		"operations", len(apiInfo.Operations))

	return nil
}

// GetAPI retrieves an API by ID
func (r *Registry) GetAPI(id string) (*models.APIInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Try exact match first
	if api, exists := r.apis[id]; exists {
		return api, nil
	}

	// Try case-insensitive match
	lowerId := strings.ToLower(id)
	for key, api := range r.apis {
		if strings.ToLower(key) == lowerId {
			return api, nil
		}
	}

	return nil, fmt.Errorf("API '%s' not found", id)
}

// ListAPIs returns all loaded APIs
func (r *Registry) ListAPIs() []*models.APIInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	apis := make([]*models.APIInfo, 0, len(r.apis))
	for _, api := range r.apis {
		apis = append(apis, api)
	}
	return apis
}

// Count returns the number of loaded APIs
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.apis)
}

// FindOperations searches for operations matching a query (supports regex)
func (r *Registry) FindOperations(apiID, query string) ([]*models.OperationInfo, error) {
	api, err := r.GetAPI(apiID)
	if err != nil {
		return nil, err
	}

	var results []*models.OperationInfo

	// Try to compile as regex first
	var isRegex bool
	var re *regexp.Regexp

	// Check if query looks like a regex (contains regex metacharacters)
	if containsRegexChars(query) {
		re, err = regexp.Compile(query)
		if err != nil {
			// If regex compilation fails, fall back to simple string search
			r.logger.Debug("Regex compilation failed, using simple search",
				"query", query,
				"error", err)
			isRegex = false
		} else {
			isRegex = true
			r.logger.Debug("Using regex search",
				"api", apiID,
				"regex", query)
		}
	} else {
		// Simple string search for queries without regex chars
		isRegex = false
		query = strings.ToLower(query)
		r.logger.Debug("Using simple search",
			"api", apiID,
			"query", query)
	}

	for _, op := range api.Operations {
		var matches bool

		if isRegex {
			matches = matchesRegex(op, re)
		} else {
			matches = matchesSimpleQuery(op, query)
		}

		if matches {
			results = append(results, op)
			if len(results) >= 10 { // Limit results
				break
			}
		}
	}

	r.logger.Debug("Found operations",
		"count", len(results),
		"isRegex", isRegex)

	return results, nil
}

// GetOperation retrieves a specific operation
func (r *Registry) GetOperation(apiID, operationID string) (*models.OperationInfo, error) {
	api, err := r.GetAPI(apiID)
	if err != nil {
		return nil, err
	}

	// Try exact match by ID
	for _, op := range api.Operations {
		if op.ID == operationID {
			return op, nil
		}
	}

	// Try match by METHOD /path format
	for _, op := range api.Operations {
		methodPath := fmt.Sprintf("%s %s", op.Method, op.Path)
		if methodPath == operationID {
			return op, nil
		}
	}

	return nil, fmt.Errorf("operation '%s' not found in API '%s'", operationID, apiID)
}

func (r *Registry) generateID(name string) string {
	// Simple ID generation - can be enhanced
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "_")
	id = strings.ReplaceAll(id, "-", "_")

	// Ensure uniqueness
	baseID := id
	counter := 1
	for {
		if _, exists := r.apis[id]; !exists {
			break
		}
		id = fmt.Sprintf("%s_%d", baseID, counter)
		counter++
	}

	return id
}

func matchesSimpleQuery(op *models.OperationInfo, query string) bool {
	// Check in path
	if strings.Contains(strings.ToLower(op.Path), query) {
		return true
	}

	// Check in summary
	if strings.Contains(strings.ToLower(op.Summary), query) {
		return true
	}

	// Check in description
	if strings.Contains(strings.ToLower(op.Description), query) {
		return true
	}

	// Check in tags
	for _, tag := range op.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}

	// Check in operation ID
	if strings.Contains(strings.ToLower(op.ID), query) {
		return true
	}

	return false
}

func matchesRegex(op *models.OperationInfo, re *regexp.Regexp) bool {
	// Check in path
	if re.MatchString(op.Path) {
		return true
	}

	// Check in summary
	if re.MatchString(op.Summary) {
		return true
	}

	// Check in description
	if re.MatchString(op.Description) {
		return true
	}

	// Check in tags
	for _, tag := range op.Tags {
		if re.MatchString(tag) {
			return true
		}
	}

	// Check in operation ID
	if re.MatchString(op.ID) {
		return true
	}

	// Check in method
	if re.MatchString(op.Method) {
		return true
	}

	return false
}

// containsRegexChars checks if string contains regex metacharacters
func containsRegexChars(s string) bool {
	regexChars := []string{
		"^", "$", ".", "*", "+", "?",
		"[", "]", "{", "}", "(", ")",
		"|", "\\",
	}

	for _, char := range regexChars {
		if strings.Contains(s, char) {
			return true
		}
	}

	return false
}
