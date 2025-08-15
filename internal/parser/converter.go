package parser

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/siriusfreak/openapi-mcp-server/internal/models"
)

// Converter converts OpenAPI specs to internal models
type Converter struct {
	logger *slog.Logger
}

// NewConverter creates a new converter
func NewConverter(logger *slog.Logger) *Converter {
	return &Converter{
		logger: logger,
	}
}

// Convert converts OpenAPI document to internal model
func (c *Converter) Convert(doc *openapi3.T) *models.APIInfo {
	api := &models.APIInfo{
		Name:        doc.Info.Title,
		Description: doc.Info.Description,
		Version:     doc.Info.Version,
		Operations:  make([]*models.OperationInfo, 0),
		Tags:        make(map[string][]*models.OperationInfo),
	}

	if doc.Components != nil {
		api.Security = c.convertSecurity(doc.Components.SecuritySchemes)
	}

	// Extract base URL
	if len(doc.Servers) > 0 {
		api.BaseURL = doc.Servers[0].URL

		// Process server variables if any
		if len(doc.Servers[0].Variables) > 0 {
			c.logger.Debug("Server has variables",
				"url", api.BaseURL,
				"variables", len(doc.Servers[0].Variables))
		}
	}

	c.logger.Debug("Converting OpenAPI document",
		"title", api.Name,
		"servers", len(doc.Servers),
		"paths", len(doc.Paths.Map()))

	// Process all paths and operations
	for path, pathItem := range doc.Paths.Map() {
		c.processPath(api, path, pathItem)
	}

	return api
}

func (c *Converter) processPath(api *models.APIInfo, path string, pathItem *openapi3.PathItem) {
	operations := map[string]*openapi3.Operation{
		"GET":     pathItem.Get,
		"POST":    pathItem.Post,
		"PUT":     pathItem.Put,
		"DELETE":  pathItem.Delete,
		"PATCH":   pathItem.Patch,
		"HEAD":    pathItem.Head,
		"OPTIONS": pathItem.Options,
	}

	for method, op := range operations {
		if op == nil {
			continue
		}

		opInfo := c.convertOperation(method, path, op)
		api.Operations = append(api.Operations, opInfo)

		// if no tags are defined, use "default" tag
		if len(opInfo.Tags) == 0 {
			c.logger.Debug("Operation has no tags, using default",
				"method", method,
				"path", path,
				"operation_id", opInfo.ID)

			opInfo.Tags = []string{"default"}
			api.Tags["default"] = append(api.Tags["default"], opInfo)
		} else {
			// Group by tags
			for _, tag := range opInfo.Tags {
				api.Tags[tag] = append(api.Tags[tag], opInfo)
			}
		}
	}
}

func (c *Converter) convertOperation(method, path string, op *openapi3.Operation) *models.OperationInfo {
	opInfo := &models.OperationInfo{
		Method:      method,
		Path:        path,
		Summary:     op.Summary,
		Description: op.Description,
		Tags:        op.Tags,
		Parameters:  c.convertParameters(op.Parameters),
		Responses:   make(map[string]*models.ResponseInfo),
	}

	// Generate operation ID
	if op.OperationID != "" {
		opInfo.ID = op.OperationID
	} else {
		// Generate ID from method and path
		opInfo.ID = fmt.Sprintf("%s_%s",
			strings.ToLower(method),
			strings.ReplaceAll(
				strings.ReplaceAll(
					strings.ReplaceAll(path, "/", "_"),
					"{", ""),
				"}", ""))
	}

	c.logger.Debug("Converting operation",
		"id", opInfo.ID,
		"method", method,
		"path", path)

	// Convert request body
	if op.RequestBody != nil {
		opInfo.RequestBody = c.convertRequestBody(op.RequestBody)
	}

	// Convert responses - Responses is now a pointer and required
	if op.Responses != nil {
		for code, resp := range op.Responses.Map() {
			opInfo.Responses[code] = c.convertResponse(resp)
		}
	}

	// Extract security requirements - Security is now a pointer to SecurityRequirements
	if op.Security != nil && *op.Security != nil {
		for _, secReq := range *op.Security {
			for name := range secReq {
				opInfo.Security = append(opInfo.Security, name)
			}
		}
	}

	return opInfo
}

func (c *Converter) convertParameters(params openapi3.Parameters) []models.ParameterInfo {
	result := make([]models.ParameterInfo, 0, len(params))

	for _, paramRef := range params {
		if paramRef.Value == nil {
			continue
		}

		param := paramRef.Value
		paramInfo := models.ParameterInfo{
			Name:        param.Name,
			In:          param.In,
			Required:    param.Required,
			Description: param.Description,
		}

		// Extract type information from schema using helper
		if paramSchema := getParameterSchema(param); paramSchema != nil {
			paramInfo.Type = getSchemaType(paramSchema)
			paramInfo.Format = paramSchema.Format
			paramInfo.Default = paramSchema.Default
			paramInfo.Example = paramSchema.Example

			// If no example, check for enum values
			if paramInfo.Example == nil && len(paramSchema.Enum) > 0 {
				paramInfo.Example = paramSchema.Enum[0]
			}
		}

		result = append(result, paramInfo)
	}

	return result
}

func (c *Converter) convertRequestBody(body *openapi3.RequestBodyRef) *models.RequestBodyInfo {
	if body.Value == nil {
		return nil
	}

	reqBody := &models.RequestBodyInfo{
		Required:    body.Value.Required,
		Description: body.Value.Description,
	}

	// Focus on JSON content type
	if content, exists := body.Value.Content["application/json"]; exists {
		reqBody.ContentType = "application/json"
		if content.Schema != nil {
			reqBody.Schema = c.convertSchema(content.Schema)
		}
		reqBody.Example = content.Example
	} else if len(body.Value.Content) > 0 {
		// Take the first content type if JSON not available
		for contentType, content := range body.Value.Content {
			reqBody.ContentType = contentType
			if content.Schema != nil {
				reqBody.Schema = c.convertSchema(content.Schema)
			}
			reqBody.Example = content.Example
			break
		}
	}

	return reqBody
}

func (c *Converter) convertResponse(resp *openapi3.ResponseRef) *models.ResponseInfo {
	if resp.Value == nil {
		return nil
	}

	respInfo := &models.ResponseInfo{
		Description: getResponseDescription(resp.Value),
		Headers:     make(map[string]string),
	}

	// Extract headers
	for name, header := range resp.Value.Headers {
		if header.Value != nil {
			respInfo.Headers[name] = header.Value.Description
		}
	}

	// Focus on JSON content
	if content, exists := resp.Value.Content["application/json"]; exists {
		respInfo.ContentType = "application/json"
		if content.Schema != nil {
			respInfo.Schema = c.convertSchema(content.Schema)
		}
		respInfo.Example = content.Example
	} else if len(resp.Value.Content) > 0 {
		// Take the first content type if JSON not available
		for contentType, content := range resp.Value.Content {
			respInfo.ContentType = contentType
			if content.Schema != nil {
				respInfo.Schema = c.convertSchema(content.Schema)
			}
			respInfo.Example = content.Example
			break
		}
	}

	return respInfo
}

func (c *Converter) convertSchema(schemaRef *openapi3.SchemaRef) *models.SchemaInfo {
	if schemaRef == nil || schemaRef.Value == nil {
		return nil
	}

	schema := schemaRef.Value
	schemaInfo := &models.SchemaInfo{
		Description: schema.Description,
		Required:    schema.Required,
		Example:     schema.Example,
		Format:      schema.Format,
	}

	// Use helper function to get type
	schemaInfo.Type = getSchemaType(schema)

	// Convert properties for objects
	if schemaInfo.Type == "object" && len(schema.Properties) > 0 {
		schemaInfo.Properties = make(map[string]*models.SchemaInfo)

		// Limit properties to avoid overwhelming output
		propCount := 0
		for name, propRef := range schema.Properties {
			if propCount >= 20 { // Limit to first 20 properties
				c.logger.Debug("Truncating properties",
					"total", len(schema.Properties),
					"shown", propCount)
				break
			}
			schemaInfo.Properties[name] = c.convertSchema(propRef)
			propCount++
		}
	}

	// Convert items for arrays
	if schemaInfo.Type == "array" && schema.Items != nil {
		schemaInfo.Items = c.convertSchema(schema.Items)
	}

	// Handle enum values
	if len(schema.Enum) > 0 && schemaInfo.Example == nil {
		// Set first enum value as example
		schemaInfo.Example = schema.Enum[0]
	}

	return schemaInfo
}

func (c *Converter) convertSecurity(schemes openapi3.SecuritySchemes) []models.SecurityScheme {
	var result []models.SecurityScheme

	for name, schemeRef := range schemes {
		if schemeRef.Value == nil {
			continue
		}

		scheme := schemeRef.Value
		secScheme := models.SecurityScheme{
			Type:        scheme.Type,
			Description: scheme.Description,
		}

		// Add more specific information based on type
		switch scheme.Type {
		case "http":
			secScheme.Scheme = scheme.Scheme
			if scheme.Scheme == "bearer" && scheme.BearerFormat != "" {
				secScheme.Description = fmt.Sprintf("Bearer %s", scheme.BearerFormat)
			}
		case "apiKey":
			if secScheme.Description == "" {
				secScheme.Description = fmt.Sprintf("API Key in %s: %s", scheme.In, scheme.Name)
			}
		case "oauth2":
			if secScheme.Description == "" {
				secScheme.Description = "OAuth 2.0 authentication"
			}
		case "openIdConnect":
			if secScheme.Description == "" {
				secScheme.Description = "OpenID Connect authentication"
			}
		}

		c.logger.Debug("Converted security scheme",
			"name", name,
			"type", secScheme.Type)

		result = append(result, secScheme)
	}

	return result
}
