package models

// APIInfo represents a parsed OpenAPI specification
type APIInfo struct {
	ID          string
	Name        string
	Description string
	Version     string
	BaseURL     string
	Operations  []*OperationInfo
	Tags        map[string][]*OperationInfo
	Security    []SecurityScheme
}

// OperationInfo represents a single API operation
type OperationInfo struct {
	ID          string
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Parameters  []ParameterInfo
	RequestBody *RequestBodyInfo
	Responses   map[string]*ResponseInfo
	Security    []string
}

// ParameterInfo represents an operation parameter
type ParameterInfo struct {
	Name        string
	In          string // path, query, header, cookie
	Type        string
	Format      string
	Required    bool
	Description string
	Default     interface{}
	Example     interface{}
}

// RequestBodyInfo represents request body schema
type RequestBodyInfo struct {
	ContentType string
	Required    bool
	Description string
	Schema      *SchemaInfo
	Example     interface{}
}

// ResponseInfo represents API response
type ResponseInfo struct {
	Description string
	ContentType string
	Schema      *SchemaInfo
	Example     interface{}
	Headers     map[string]string
}

// SchemaInfo represents a data schema
type SchemaInfo struct {
	Type        string
	Format      string
	Description string
	Properties  map[string]*SchemaInfo
	Required    []string
	Items       *SchemaInfo // for arrays
	Example     interface{}
}

// SecurityScheme represents an authentication method
type SecurityScheme struct {
	Type        string
	Scheme      string
	Description string
}
