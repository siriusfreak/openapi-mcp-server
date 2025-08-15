package handlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/siriusfreak/openapi-mcp-server/internal/formatter"
	"github.com/siriusfreak/openapi-mcp-server/internal/models"
	"github.com/siriusfreak/openapi-mcp-server/internal/registry"
)

// Handlers contains all MCP tool handlers
type Handlers struct {
	registry  *registry.Registry
	formatter *formatter.Formatter
	logger    *slog.Logger
}

// New creates new handlers instance
func New(reg *registry.Registry, logger *slog.Logger) *Handlers {
	return &Handlers{
		registry:  reg,
		formatter: formatter.New(),
		logger:    logger,
	}
}

// ListAPIs lists all available APIs
func (h *Handlers) ListAPIs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	h.logger.Debug("Handling list_apis request")

	apis := h.registry.ListAPIs()

	if len(apis) == 0 {
		return mcp.NewToolResultText("No APIs loaded"), nil
	}

	result := h.formatter.FormatAPIList(apis)
	h.logger.Debug("Returning API list", "count", len(apis))

	return mcp.NewToolResultText(result), nil
}

// GetAPIOverview returns overview of a specific API
func (h *Handlers) GetAPIOverview(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiID, err := req.RequireString("api_id")
	if err != nil {
		h.logger.Error("Missing required parameter", "param", "api_id", "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	h.logger.Debug("Handling get_api_overview request", "api_id", apiID)

	api, err := h.registry.GetAPI(apiID)
	if err != nil {
		h.logger.Warn("API not found", "api_id", apiID, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	overview := h.formatter.FormatAPIOverview(api)
	return mcp.NewToolResultText(overview), nil
}

// FindOperations searches for operations matching a query
func (h *Handlers) FindOperations(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiID, err := req.RequireString("api_id")
	if err != nil {
		h.logger.Error("Missing required parameter", "param", "api_id", "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	query, err := req.RequireString("query")
	if err != nil {
		h.logger.Error("Missing required parameter", "param", "query", "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	h.logger.Debug("Handling find_operations request",
		"api_id", apiID,
		"query", query)

	operations, err := h.registry.FindOperations(apiID, query)
	if err != nil {
		h.logger.Warn("Failed to find operations",
			"api_id", apiID,
			"query", query,
			"error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	if len(operations) == 0 {
		return mcp.NewToolResultText(fmt.Sprintf("No operations found matching pattern '%s'", query)), nil
	}

	result := h.formatter.FormatOperationList(operations)
	h.logger.Debug("Found operations",
		"count", len(operations))

	return mcp.NewToolResultText(result), nil
}

// ListOperations lists all operations in an API
func (h *Handlers) ListOperations(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiID, err := req.RequireString("api_id")
	if err != nil {
		h.logger.Error("Missing required parameter", "param", "api_id", "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	api, err := h.registry.GetAPI(apiID)
	if err != nil {
		h.logger.Warn("API not found", "api_id", apiID, "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	tag := req.GetString("tag", "")

	h.logger.Debug("Handling list_operations request",
		"api_id", apiID,
		"tag", tag)

	var operations []*models.OperationInfo
	if tag != "" {
		operations = api.Tags[tag]
		if len(operations) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No operations found with tag '%s'", tag)), nil
		}
	} else {
		operations = api.Operations
	}

	if len(operations) == 0 {
		return mcp.NewToolResultText("No operations found"), nil
	}

	h.logger.Debug("Found operations",
		"api_id", apiID,
		"tag", tag,
		"count", len(operations))
	var result string
	if tag != "" {
		result = h.formatter.FormatOperationList(operations)
	} else {
		result = h.formatter.FormatOperationsByTag(api.Tags)
	}

	return mcp.NewToolResultText(result), nil
}

// GetOperationDetails returns detailed information about an operation
func (h *Handlers) GetOperationDetails(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiID, err := req.RequireString("api_id")
	if err != nil {
		h.logger.Error("Missing required parameter", "param", "api_id", "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	operationID, err := req.RequireString("operation_id")
	if err != nil {
		h.logger.Error("Missing required parameter", "param", "operation_id", "error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	h.logger.Debug("Handling get_operation_details request",
		"api_id", apiID,
		"operation_id", operationID)

	operation, err := h.registry.GetOperation(apiID, operationID)
	if err != nil {
		h.logger.Warn("Operation not found",
			"api_id", apiID,
			"operation_id", operationID,
			"error", err)
		return mcp.NewToolResultError(err.Error()), nil
	}

	api, _ := h.registry.GetAPI(apiID)
	details := h.formatter.FormatOperationDetails(operation, api.BaseURL)

	h.logger.Debug("Returning operation details",
		"method", operation.Method,
		"path", operation.Path)

	return mcp.NewToolResultText(details), nil
}
