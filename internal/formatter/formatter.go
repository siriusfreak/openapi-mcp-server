package formatter

import (
	"fmt"
	"strings"

	"github.com/siriusfreak/openapi-mcp-server/internal/models"
)

// Formatter formats API information for LLM consumption
type Formatter struct{}

// New creates a new formatter
func New() *Formatter {
	return &Formatter{}
}

// FormatAPIList formats a list of APIs
func (f *Formatter) FormatAPIList(apis []*models.APIInfo) string {
	var sb strings.Builder
	sb.WriteString("## Available APIs\n\n")

	for i, api := range apis {
		sb.WriteString(fmt.Sprintf("%d. **%s** (v%s)\n", i+1, api.Name, api.Version))
		sb.WriteString(fmt.Sprintf("   ID: `%s`\n", api.ID))
		if api.Description != "" {
			desc := api.Description
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			sb.WriteString(fmt.Sprintf("   %s\n", desc))
		}
		sb.WriteString(fmt.Sprintf("   Operations: %d\n\n", len(api.Operations)))
	}

	return sb.String()
}

// FormatAPIOverview formats an API overview
func (f *Formatter) FormatAPIOverview(api *models.APIInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s API Overview\n\n", api.Name))
	sb.WriteString(fmt.Sprintf("**Version:** %s\n", api.Version))

	if api.BaseURL != "" {
		sb.WriteString(fmt.Sprintf("**Base URL:** %s\n", api.BaseURL))
	}

	if api.Description != "" {
		sb.WriteString(fmt.Sprintf("\n%s\n", api.Description))
	}

	// Statistics
	sb.WriteString(fmt.Sprintf("\n**Total Operations:** %d\n", len(api.Operations)))

	// Tags/Categories
	if len(api.Tags) > 0 {
		sb.WriteString("\n**Categories:**\n")
		for tag, ops := range api.Tags {
			sb.WriteString(fmt.Sprintf("- %s (%d operations)\n", tag, len(ops)))
		}
	}

	// Security schemes
	if len(api.Security) > 0 {
		sb.WriteString("\n**Authentication Methods:**\n")
		for _, sec := range api.Security {
			sb.WriteString(fmt.Sprintf("- %s", sec.Type))
			if sec.Scheme != "" {
				sb.WriteString(fmt.Sprintf(" (%s)", sec.Scheme))
			}
			if sec.Description != "" {
				sb.WriteString(fmt.Sprintf(": %s", sec.Description))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// FormatOperationList formats a list of operations
func (f *Formatter) FormatOperationList(operations []*models.OperationInfo) string {
	var sb strings.Builder
	sb.WriteString("## Found Operations\n\n")

	for _, op := range operations {
		sb.WriteString(fmt.Sprintf("- **%s %s**", op.Method, op.Path))
		if op.Summary != "" {
			sb.WriteString(fmt.Sprintf(" - %s", op.Summary))
		}
		sb.WriteString("\n")
		if op.ID != "" {
			sb.WriteString(fmt.Sprintf("  Operation ID: `%s`\n", op.ID))
		}
	}

	return sb.String()
}

// FormatOperationsByTag formats operations grouped by tag
func (f *Formatter) FormatOperationsByTag(tags map[string][]*models.OperationInfo) string {
	var sb strings.Builder
	sb.WriteString("## API Operations by Category\n\n")

	if len(tags) == 0 {
		sb.WriteString("No categorized operations found.\n")
		return sb.String()
	}

	for tag, ops := range tags {
		if tag == "" {
			tag = "Uncategorized"
		}
		sb.WriteString(fmt.Sprintf("### %s\n\n", tag))
		for _, op := range ops {
			sb.WriteString(fmt.Sprintf("- **%s %s**", op.Method, op.Path))
			if op.Summary != "" {
				sb.WriteString(fmt.Sprintf(" - %s", op.Summary))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatOperationDetails formats detailed operation information
func (f *Formatter) FormatOperationDetails(op *models.OperationInfo, baseURL string) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# %s %s\n\n", op.Method, op.Path))

	// Description
	if op.Summary != "" {
		sb.WriteString(fmt.Sprintf("**Summary:** %s\n\n", op.Summary))
	}
	if op.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", op.Description))
	}

	// Full URL
	if baseURL != "" {
		sb.WriteString(fmt.Sprintf("**Full URL:** `%s%s`\n\n", baseURL, op.Path))
	}

	// Authentication
	if len(op.Security) > 0 {
		sb.WriteString("**Authentication Required:**\n")
		for _, sec := range op.Security {
			sb.WriteString(fmt.Sprintf("- %s\n", sec))
		}
		sb.WriteString("\n")
	}

	// Parameters
	if len(op.Parameters) > 0 {
		sb.WriteString("## Parameters\n\n")

		// Group by type
		pathParams := f.filterParams(op.Parameters, "path")
		queryParams := f.filterParams(op.Parameters, "query")
		headerParams := f.filterParams(op.Parameters, "header")

		if len(pathParams) > 0 {
			sb.WriteString("### Path Parameters\n\n")
			f.formatParams(&sb, pathParams)
		}

		if len(queryParams) > 0 {
			sb.WriteString("### Query Parameters\n\n")
			f.formatParams(&sb, queryParams)
		}

		if len(headerParams) > 0 {
			sb.WriteString("### Header Parameters\n\n")
			f.formatParams(&sb, headerParams)
		}
	}

	// Request Body
	if op.RequestBody != nil {
		sb.WriteString("## Request Body\n\n")
		sb.WriteString(fmt.Sprintf("**Content-Type:** %s\n", op.RequestBody.ContentType))
		sb.WriteString(fmt.Sprintf("**Required:** %v\n\n", op.RequestBody.Required))

		if op.RequestBody.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n\n", op.RequestBody.Description))
		}

		if op.RequestBody.Schema != nil {
			sb.WriteString("**Schema:**\n```json\n")
			f.formatSchema(&sb, op.RequestBody.Schema, 0)
			sb.WriteString("\n```\n\n")
		}

		if op.RequestBody.Example != nil {
			sb.WriteString("**Example:**\n```json\n")
			sb.WriteString(fmt.Sprintf("%v\n", op.RequestBody.Example))
			sb.WriteString("```\n\n")
		}
	}

	// Responses
	if len(op.Responses) > 0 {
		sb.WriteString("## Responses\n\n")

		// Focus on success responses first
		for code, resp := range op.Responses {
			if code[0] == '2' { // 2xx responses
				f.formatResponse(&sb, code, resp)
			}
		}

		// Then error responses
		for code, resp := range op.Responses {
			if code[0] != '2' { // Non-2xx responses
				f.formatResponse(&sb, code, resp)
			}
		}
	}

	// Example usage
	sb.WriteString("## Example Usage\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString(fmt.Sprintf("curl -X %s \\\n", op.Method))
	if baseURL != "" {
		sb.WriteString(fmt.Sprintf("  '%s%s' \\\n", baseURL, f.examplePath(op)))
	}
	if len(op.Security) > 0 {
		sb.WriteString("  -H 'Authorization: Bearer YOUR_TOKEN' \\\n")
	}
	if op.RequestBody != nil && op.RequestBody.ContentType == "application/json" {
		sb.WriteString("  -H 'Content-Type: application/json' \\\n")
		sb.WriteString("  -d '{...}'\n")
	}
	sb.WriteString("```\n")

	return sb.String()
}

func (f *Formatter) formatResponse(sb *strings.Builder, code string, resp *models.ResponseInfo) {
	sb.WriteString(fmt.Sprintf("### %s - %s\n\n", code, resp.Description))

	if resp.Schema != nil {
		sb.WriteString("**Response Schema:**\n```json\n")
		f.formatSchema(sb, resp.Schema, 0)
		sb.WriteString("\n```\n\n")
	}

	if resp.Example != nil {
		sb.WriteString("**Example Response:**\n```json\n")
		sb.WriteString(fmt.Sprintf("%v\n", resp.Example))
		sb.WriteString("```\n\n")
	}

	if len(resp.Headers) > 0 {
		sb.WriteString("**Response Headers:**\n")
		for name, desc := range resp.Headers {
			sb.WriteString(fmt.Sprintf("- `%s`: %s\n", name, desc))
		}
		sb.WriteString("\n")
	}
}

func (f *Formatter) filterParams(params []models.ParameterInfo, in string) []models.ParameterInfo {
	var filtered []models.ParameterInfo
	for _, p := range params {
		if p.In == in {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func (f *Formatter) formatParams(sb *strings.Builder, params []models.ParameterInfo) {
	for _, p := range params {
		sb.WriteString(fmt.Sprintf("- **%s**", p.Name))
		if p.Type != "" {
			sb.WriteString(fmt.Sprintf(" (%s", p.Type))
			if p.Format != "" {
				sb.WriteString(fmt.Sprintf(", %s", p.Format))
			}
			sb.WriteString(")")
		}
		if p.Required {
			sb.WriteString(" - **Required**")
		}
		sb.WriteString("\n")

		if p.Description != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", p.Description))
		}
		if p.Default != nil {
			sb.WriteString(fmt.Sprintf("  Default: `%v`\n", p.Default))
		}
		if p.Example != nil {
			sb.WriteString(fmt.Sprintf("  Example: `%v`\n", p.Example))
		}
		sb.WriteString("\n")
	}
}

func (f *Formatter) formatSchema(sb *strings.Builder, schema *models.SchemaInfo, indent int) {
	indentStr := strings.Repeat("  ", indent)

	if schema.Type == "object" && schema.Properties != nil {
		sb.WriteString(fmt.Sprintf("%s{\n", indentStr))
		for name, prop := range schema.Properties {
			required := ""
			for _, req := range schema.Required {
				if req == name {
					required = " (required)"
					break
				}
			}
			sb.WriteString(fmt.Sprintf("%s  \"%s\": ", indentStr, name))

			if prop.Type == "object" || prop.Type == "array" {
				f.formatSchema(sb, prop, indent+1)
			} else {
				typeStr := prop.Type
				if prop.Format != "" {
					typeStr += fmt.Sprintf(" (%s)", prop.Format)
				}
				sb.WriteString(fmt.Sprintf("%s%s", typeStr, required))
				if prop.Description != "" && len(prop.Description) < 50 {
					sb.WriteString(fmt.Sprintf(" // %s", prop.Description))
				}
			}
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf("%s}", indentStr))
	} else if schema.Type == "array" && schema.Items != nil {
		sb.WriteString("[\n")
		f.formatSchema(sb, schema.Items, indent+1)
		sb.WriteString(fmt.Sprintf("\n%s]", indentStr))
	} else {
		sb.WriteString(schema.Type)
		if schema.Format != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", schema.Format))
		}
	}
}

func (f *Formatter) examplePath(op *models.OperationInfo) string {
	path := op.Path
	// Replace path parameters with example values
	for _, param := range op.Parameters {
		if param.In == "path" {
			placeholder := fmt.Sprintf("{%s}", param.Name)
			exampleValue := "123" // Default example
			if param.Example != nil {
				exampleValue = fmt.Sprintf("%v", param.Example)
			} else if param.Name == "id" || strings.HasSuffix(param.Name, "_id") {
				exampleValue = "123"
			} else if param.Name == "username" || param.Name == "user" {
				exampleValue = "john_doe"
			} else if param.Name == "email" {
				exampleValue = "user@example.com"
			}
			path = strings.ReplaceAll(path, placeholder, exampleValue)
		}
	}
	return path
}
