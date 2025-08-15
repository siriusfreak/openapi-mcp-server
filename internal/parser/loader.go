package parser

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/siriusfreak/openapi-mcp-server/internal/models"
)

// Loader handles loading and parsing OpenAPI specifications
type Loader struct {
	loader *openapi3.Loader
	logger *slog.Logger
}

// NewLoader creates a new OpenAPI specification loader
func NewLoader(logger *slog.Logger) *Loader {
	return &Loader{
		loader: openapi3.NewLoader(),
		logger: logger,
	}
}

// Load loads an OpenAPI specification from file or URL
func (l *Loader) Load(source string) (*models.APIInfo, error) {
	var doc *openapi3.T
	var err error

	l.logger.Info("Loading OpenAPI spec", "source", source)

	// Check if source is URL
	if isURL(source) {
		doc, err = l.loadFromURL(source)
	} else {
		doc, err = l.loadFromFile(source)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	// Validate the document
	ctx := context.Background()

	// Use the standard Validate method
	if err := doc.Validate(ctx); err != nil {
		l.logger.Warn("OpenAPI spec validation warning",
			"source", source,
			"error", err)
		// Continue anyway - some specs have minor validation issues
		// Many real-world specs have minor issues that don't prevent usage
	}

	// Convert to internal model
	converter := NewConverter(l.logger)
	apiInfo := converter.Convert(doc)

	l.logger.Info("Successfully parsed OpenAPI spec",
		"name", apiInfo.Name,
		"version", apiInfo.Version,
		"operations", len(apiInfo.Operations))

	return apiInfo, nil
}

func (l *Loader) loadFromFile(path string) (*openapi3.T, error) {
	l.logger.Debug("Loading from file", "path", path)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	doc, err := l.loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI document: %w", err)
	}

	return doc, nil
}

func (l *Loader) loadFromURL(urlStr string) (*openapi3.T, error) {
	l.logger.Debug("Loading from URL", "url", urlStr)

	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Save to a temporary file
	tempFile, err := os.CreateTemp("", "openapi-*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.ReadFrom(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	slog.Debug("Temporary file created", "path", tempFile.Name())
	doc, err := l.loader.LoadFromFile(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI document from URL: %w", err)
	}

	return doc, nil
}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}
