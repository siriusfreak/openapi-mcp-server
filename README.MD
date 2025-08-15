# OpenAPI MCP Server

A Model Context Protocol (MCP) server that provides LLM-friendly access to OpenAPI/Swagger specifications. This server enables AI models to understand and generate code for REST APIs by presenting API documentation in a simplified, natural language format.

## Features

- 🔄 **Multi-spec support**: Load multiple OpenAPI specifications from files or URLs
- 🔍 **Smart search**: Find relevant API operations using keywords
- 📝 **LLM-optimized output**: Natural language descriptions optimized for AI understanding
- 🚀 **Multiple transport modes**: Support for STDIO (local) and HTTP/SSE (remote) modes
- 🛠️ **Meta-tools approach**: Provides intelligent tools instead of overwhelming the LLM with every endpoint
- 📊 **Structured logging**: Uses Go's slog for structured, leveled logging

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/siriusfreak/openapi-mcp-server.git
cd openapi-mcp-server

# Install dependencies
make deps

# Build the server
make build
```

### Using Go Install

```bash
go install github.com/siriusfreak/openapi-mcp-server/cmd/server@latest
```

### Using Docker

```bash
# Build Docker image
make docker-build

# Run container
make docker-run
```

## Usage Examples with LLM

Once integrated with Claude or another LLM, you can use natural language queries:

### Basic Examples
```
"List all available APIs"
"Show me an overview of the petstore API"
"Find all user-related endpoints in the GitHub API"
"Get details about the GET /users/{id} operation"
```

### Advanced Search with Regex
```
"Find all GET operations in the API" 
→ Uses regex: ^GET

"Show me all endpoints that create or update resources"
→ Uses regex: ^(POST|PUT|PATCH)

"Find operations with path parameters"
→ Uses regex: /{\\w+}

"List all delete operations"
→ Uses regex: (?i)delete

"Find user order endpoints"
→ Uses regex: /users?/.*/orders
```

### STDIO Mode (for local integration)

```bash
# Run with a local swagger file
./build/openapi-mcp-server --specs ./path/to/swagger.json --mode stdio

# Run with multiple specs
./build/openapi-mcp-server \
  --specs ./api1.yaml \
  --specs https://api.example.com/swagger.json \
  --mode stdio

# Run with URL spec
./build/openapi-mcp-server \
  --specs https://petstore.swagger.io/v2/swagger.json \
  --mode stdio
```

### HTTP Mode (for remote access)

```bash
# Start the HTTP server
./build/openapi-mcp-server \
  --specs https://petstore.swagger.io/v2/swagger.json \
  --mode http \
  --port 3000

# With multiple specs
./build/openapi-mcp-server \
  --specs ./local-api.yaml \
  --specs https://api.github.com/swagger.json \
  --mode http \
  --port 8080
```

### Configuration File

Create a `config.yaml`:

```yaml
specs:
  - https://api.github.com/swagger.json
  - ./local-api.yaml
  - https://petstore.swagger.io/v2/swagger.json
port: 3000
mode: http
log_level: debug  # debug, info, warn, error
```

Run with config:

```bash
./build/openapi-mcp-server --config config.yaml
```

### Environment Variables

You can also use environment variables (prefixed with `MCP_`):

```bash
export MCP_MODE=http
export MCP_PORT=8080
export MCP_LOG_LEVEL=debug
./build/openapi-mcp-server --specs ./api.yaml
```

## Available MCP Tools

The server provides the following tools for LLMs:

### 1. `list_apis`
Lists all loaded API specifications with their IDs and basic information.

**Parameters:** None

**Example Response:**
```markdown
## Available APIs

1. **Swagger Petstore** (v1.0.0)
   ID: `swagger_petstore`
   This is a sample server Petstore server...
   Operations: 14
```

### 2. `get_api_overview`
Provides a comprehensive overview of a specific API.

**Parameters:**
- `api_id` (required): API identifier

**Example Response:**
```markdown
# Swagger Petstore API Overview

**Version:** 1.0.0
**Base URL:** https://petstore.swagger.io/v2

A sample server Petstore server...

**Total Operations:** 14

**Categories:**
- pet (8 operations)
- store (4 operations)
- user (8 operations)

**Authentication Methods:**
- apiKey: API Key authentication
- oauth2: OAuth2 authentication
```

### 3. `find_operations`
Searches for operations matching a keyword or **regex pattern**.

**Parameters:**
- `api_id` (required): API identifier
- `query` (required): Search keyword or regex pattern

**Search modes:**
- **Simple search**: Case-insensitive substring matching across operation paths, summaries, descriptions, tags, and IDs
- **Regex search**: Automatically detected when query contains regex metacharacters (^, $, ., *, +, ?, [, ], etc.)

**Regex Examples:**
- `^GET` - Find all GET operations
- `^(POST|PUT)` - Find all POST or PUT operations
- `/users?/` - Find operations with "/user/" or "/users/"
- `/pets/{\\w+}` - Find operations with path parameters like "/pets/{petId}"
- `(?i)delete` - Case-insensitive search for "delete"
- `.*Order# OpenAPI MCP Server

A Model Context Protocol (MCP) server that provides LLM-friendly access to OpenAPI/Swagger specifications. This server enables AI models to understand and generate code for REST APIs by presenting API documentation in a simplified, natural language format.

## Features

- 🔄 **Multi-spec support**: Load multiple OpenAPI specifications from files or URLs
- 🔍 **Smart search**: Find relevant API operations using keywords
- 📝 **LLM-optimized output**: Natural language descriptions optimized for AI understanding
- 🚀 **Multiple transport modes**: Support for STDIO (local) and HTTP/SSE (remote) modes
- 🛠️ **Meta-tools approach**: Provides intelligent tools instead of overwhelming the LLM with every endpoint
- 📊 **Structured logging**: Uses Go's slog for structured, leveled logging

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/siriusfreak/openapi-mcp-server.git
cd openapi-mcp-server

# Install dependencies
make deps

# Build the server
make build
```

### Using Go Install

```bash
go install github.com/siriusfreak/openapi-mcp-server/cmd/server@latest
```

### Using Docker

```bash
# Build Docker image
make docker-build

# Run container
make docker-run
```

## Usage

### STDIO Mode (for local integration)

```bash
# Run with a local swagger file
./build/openapi-mcp-server --specs ./path/to/swagger.json --mode stdio

# Run with multiple specs
./build/openapi-mcp-server \
  --specs ./api1.yaml \
  --specs https://api.example.com/swagger.json \
  --mode stdio

# Run with URL spec
./build/openapi-mcp-server \
  --specs https://petstore.swagger.io/v2/swagger.json \
  --mode stdio
```

### HTTP Mode (for remote access)

```bash
# Start the HTTP server
./build/openapi-mcp-server \
  --specs https://petstore.swagger.io/v2/swagger.json \
  --mode http \
  --port 3000

# With multiple specs
./build/openapi-mcp-server \
  --specs ./local-api.yaml \
  --specs https://api.github.com/swagger.json \
  --mode http \
  --port 8080
```

### Configuration File

Create a `config.yaml`:

```yaml
specs:
  - https://api.github.com/swagger.json
  - ./local-api.yaml
  - https://petstore.swagger.io/v2/swagger.json
port: 3000
mode: http
log_level: debug  # debug, info, warn, error
```

Run with config:

```bash
./build/openapi-mcp-server --config config.yaml
```

### Environment Variables

You can also use environment variables (prefixed with `MCP_`):

```bash
export MCP_MODE=http
export MCP_PORT=8080
export MCP_LOG_LEVEL=debug
./build/openapi-mcp-server --specs ./api.yaml
```

## Available MCP Tools

The server provides the following tools for LLMs:

### 1. `list_apis`
Lists all loaded API specifications with their IDs and basic information.

**Parameters:** None

**Example Response:**
```markdown
## Available APIs

1. **Swagger Petstore** (v1.0.0)
   ID: `swagger_petstore`
   This is a sample server Petstore server...
   Operations: 14
```

### 2. `get_api_overview`
Provides a comprehensive overview of a specific API.

**Parameters:**
- `api_id` (required): API identifier

**Example Response:**
```markdown
# Swagger Petstore API Overview

**Version:** 1.0.0
**Base URL:** https://petstore.swagger.io/v2

A sample server Petstore server...

**Total Operations:** 14

**Categories:**
- pet (8 operations)
- store (4 operations)
- user (8 operations)

**Authentication Methods:**
- apiKey: API Key authentication
- oauth2: OAuth2 authentication
```

- Find operations with IDs ending in "Order"
- `/users/\\d+/orders` - Find user order endpoints with numeric IDs

**Simple Search Examples:**
- `user` - Find all user-related operations
- `pet` - Find all pet-related operations
- `create` - Find operations with "create" in their description

**Example Response:**
```markdown
## Found Operations

- **GET /pet/{petId}** - Find pet by ID
  Operation ID: `getPetById`
- **PUT /pet** - Update an existing pet
  Operation ID: `updatePet`
- **POST /pet** - Add a new pet to the store
  Operation ID: `addPet`
```

### 4. `list_operations`
Lists all operations, optionally filtered by tag.

**Parameters:**
- `api_id` (required): API identifier
- `tag` (optional): Tag to filter by

### 5. `get_operation_details`
Provides detailed information about a specific operation.

**Parameters:**
- `api_id` (required): API identifier
- `operation_id` (required): Operation ID or "METHOD /path" format

**Example Response:**
```markdown
# GET /pet/{petId}

**Summary:** Find pet by ID

Returns a single pet

**Full URL:** `https://petstore.swagger.io/v2/pet/{petId}`

## Parameters

### Path Parameters

- **petId** (integer, int64) - **Required**
  ID of pet to return

## Responses

### 200 - successful operation
...
```

## Integration with Claude Desktop

Add to your Claude Desktop configuration (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "openapi": {
      "command": "/path/to/openapi-mcp-server",
      "args": ["--specs", "/path/to/swagger.json", "--mode", "stdio"]
    }
  }
}
```

Or for multiple specs:

```json
{
  "mcpServers": {
    "openapi": {
      "command": "/path/to/openapi-mcp-server",
      "args": [
        "--specs", "/path/to/api1.yaml",
        "--specs", "https://api.example.com/swagger.json",
        "--mode", "stdio"
      ]
    }
  }
}
```

## Integration with Cursor

Add to your Cursor settings:

```json
{
  "mcp.servers": {
    "openapi": {
      "command": "openapi-mcp-server",
      "args": ["--specs", "./api.yaml", "--mode", "stdio"]
    }
  }
}
```

## Development

### Project Structure

```
openapi-mcp-server/
├── cmd/server/          # Application entry point
├── internal/            # Private application code
│   ├── config/         # Configuration management
│   ├── handlers/       # MCP tool handlers
│   ├── parser/         # OpenAPI specification parsing
│   ├── registry/       # API registry management
│   ├── formatter/      # Response formatting for LLMs
│   └── models/         # Domain models
├── build/              # Build artifacts (generated)
├── Makefile           # Build automation
└── README.md          # This file
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Lint code (requires golangci-lint)
make lint

# Run with example spec
make run-example

# Run in HTTP mode
make run-http
```

### Testing

```bash
# Run all tests
make test

# Run with race detector
go test -race ./...

# Run specific package tests
go test -v ./internal/parser

# Generate coverage report
make test-coverage
```

## Docker Support

### Building Docker Image

```bash
# Build image
docker build -t openapi-mcp-server:latest .

# Or using make
make docker-build
```

### Running with Docker

```bash
# Run in HTTP mode
docker run -p 3000:3000 \
  openapi-mcp-server:latest \
  --specs https://petstore.swagger.io/v2/swagger.json \
  --mode http

# Run in STDIO mode (for testing)
docker run -it --rm \
  openapi-mcp-server:latest \
  --specs https://petstore.swagger.io/v2/swagger.json \
  --mode stdio

# With volume mount for local specs
docker run -p 3000:3000 \
  -v $(pwd)/specs:/specs \
  openapi-mcp-server:latest \
  --specs /specs/api.yaml \
  --mode http
```

## Logging

The server uses structured logging with slog. Log levels can be configured:

- `debug`: Detailed debugging information
- `info`: General information (default)
- `warn`: Warning messages
- `error`: Error messages only

Set log level via:
- Command line: `--log-level debug`
- Config file: `log_level: debug`
- Environment: `MCP_LOG_LEVEL=debug`

## Architecture

The server follows clean architecture principles:

- **Parser Layer**: Loads and validates OpenAPI specifications
- **Registry Layer**: Manages loaded APIs and provides search capabilities
- **Handler Layer**: Implements MCP tool logic
- **Formatter Layer**: Converts technical API details to LLM-friendly text
- **Transport Layer**: Handles STDIO/HTTP communication via mcp-go library

## Supported OpenAPI Versions

- OpenAPI 3.0.x
- OpenAPI 3.1.x
- Swagger 2.0 (automatically converted to OpenAPI 3.0)

## Limitations

- Request body and response schemas are limited to first 20 properties to avoid overwhelming output
- Search results are limited to 10 operations
- Only JSON content type is fully supported for request/response bodies
- No direct API execution - this server only provides documentation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Author

Created by [@siriusfreak](https://github.com/siriusfreak)

## Acknowledgments

- [mcp-go](https://github.com/mark3labs/mcp-go) - MCP protocol implementation for Go
- [kin-openapi](https://github.com/getkin/kin-openapi) - OpenAPI 3.0 parser for Go
- [Model Context Protocol](https://modelcontextprotocol.io/) - Protocol specification
