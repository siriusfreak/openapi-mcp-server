package parser

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// Helper functions for working with OpenAPI types

// getSchemaType extracts the primary type from a Schema
// Since Type can be nil or a Types (which can be string or array)
func getSchemaType(schema *openapi3.Schema) string {
	if schema == nil || schema.Type == nil {
		// Check for composite schemas
		if len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 || len(schema.AllOf) > 0 {
			return "composite"
		}
		return ""
	}

	// Use the built-in Is() method to check types
	if schema.Type.Is("object") {
		return "object"
	}
	if schema.Type.Is("array") {
		return "array"
	}
	if schema.Type.Is("string") {
		return "string"
	}
	if schema.Type.Is("number") {
		return "number"
	}
	if schema.Type.Is("integer") {
		return "integer"
	}
	if schema.Type.Is("boolean") {
		return "boolean"
	}
	if schema.Type.Is("null") {
		return "null"
	}

	// If none match, return the first type if available
	types := schema.Type.Slice()
	if len(types) > 0 {
		return types[0]
	}

	return ""
}

// isRequiredField checks if a field is in the required list
func isRequiredField(fieldName string, required []string) bool {
	for _, req := range required {
		if req == fieldName {
			return true
		}
	}
	return false
}

// getResponseDescription safely extracts response description
func getResponseDescription(resp *openapi3.Response) string {
	if resp == nil || resp.Description == nil {
		return ""
	}
	return *resp.Description
}

// getSecuritySchemeType returns a readable security scheme type
func getSecuritySchemeType(scheme *openapi3.SecurityScheme) string {
	if scheme == nil {
		return ""
	}

	switch scheme.Type {
	case "apiKey":
		location := scheme.In
		name := scheme.Name
		return fmt.Sprintf("API Key (in %s: %s)", location, name)
	case "http":
		if scheme.Scheme != "" {
			if scheme.Scheme == "bearer" {
				format := scheme.BearerFormat
				if format != "" {
					return fmt.Sprintf("Bearer Token (%s)", format)
				}
				return "Bearer Token"
			}
			return fmt.Sprintf("HTTP %s", scheme.Scheme)
		}
		return "HTTP"
	case "oauth2":
		return "OAuth 2.0"
	case "openIdConnect":
		return "OpenID Connect"
	default:
		return scheme.Type
	}
}

// resolveSchemaRef resolves a schema reference if needed
func resolveSchemaRef(schemaRef *openapi3.SchemaRef) *openapi3.Schema {
	if schemaRef == nil {
		return nil
	}

	// The Value field contains the resolved schema
	return schemaRef.Value
}

// getParameterSchema safely extracts parameter schema
func getParameterSchema(param *openapi3.Parameter) *openapi3.Schema {
	if param == nil || param.Schema == nil {
		return nil
	}
	return param.Schema.Value
}

// mergeSecurityRequirements merges operation-level security with API-level security
func mergeSecurityRequirements(opSecurity *openapi3.SecurityRequirements, apiSecurity openapi3.SecurityRequirements) []string {
	requirements := make(map[string]bool)

	// If operation has security, use only operation security (it overrides API-level)
	if opSecurity != nil && len(*opSecurity) > 0 {
		for _, req := range *opSecurity {
			for name := range req {
				requirements[name] = true
			}
		}
	} else if len(apiSecurity) > 0 {
		// Otherwise use API-level security
		for _, req := range apiSecurity {
			for name := range req {
				requirements[name] = true
			}
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(requirements))
	for name := range requirements {
		result = append(result, name)
	}

	return result
}
