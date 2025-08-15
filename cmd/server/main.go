package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/siriusfreak/openapi-mcp-server/internal/config"
	"github.com/siriusfreak/openapi-mcp-server/internal/handlers"
	"github.com/siriusfreak/openapi-mcp-server/internal/registry"
)

var (
	cfgFile string
	specs   []string
	port    int
	mode    string // "stdio" or "http"
	logger  *slog.Logger
)

var rootCmd = &cobra.Command{
	Use:   "openapi-mcp-server",
	Short: "MCP server for OpenAPI specifications",
	Long:  `A Model Context Protocol server that provides LLM-friendly access to OpenAPI/Swagger specifications`,
	Run:   runServer,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	rootCmd.PersistentFlags().StringSliceVarP(&specs, "specs", "s", []string{}, "OpenAPI spec files or URLs")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 3000, "HTTP server port")
	rootCmd.PersistentFlags().StringVarP(&mode, "mode", "m", "stdio", "Server mode: stdio or http")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Failed to execute command", "error", err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Override with command line flags if provided
	if len(specs) > 0 {
		cfg.Specs = specs
	}
	if port != 3000 {
		cfg.Port = port
	}
	if mode != "stdio" {
		cfg.Mode = mode
	}

	// Initialize logger
	logger = setupLogger(cfg.LogLevel)

	// Create API registry
	apiRegistry := registry.New(logger)

	// Load OpenAPI specifications
	logger.Info("Loading OpenAPI specifications...")
	for _, specPath := range cfg.Specs {
		if err := apiRegistry.LoadSpec(specPath); err != nil {
			logger.Error("Failed to load specification",
				"spec", specPath,
				"error", err)
			continue
		}
		logger.Info("Successfully loaded specification", "spec", specPath)
	}

	if apiRegistry.Count() == 0 {
		logger.Error("No valid specifications loaded")
		os.Exit(1)
	}

	// Create MCP server
	mcpServer := createMCPServer(apiRegistry)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down server...")
		cancel()
	}()

	// Start server based on mode
	switch cfg.Mode {
	case "stdio":
		logger.Info("Starting MCP server in STDIO mode...")
		if err := server.ServeStdio(mcpServer); err != nil {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	case "http":
		logger.Info("Starting MCP server in HTTP mode", "port", cfg.Port)
		// Create HTTP handler for MCP server
		handler := server.NewSSEServer(mcpServer)

		// Create HTTP server
		httpServer := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Port),
			Handler: handler,
		}

		// Start server in goroutine
		go func() {
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Error("HTTP server error", "error", err)
				os.Exit(1)
			}
		}()

		// Wait for shutdown signal
		<-ctx.Done()

		// Graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error("Server shutdown error", "error", err)
		}

		logger.Info("Server stopped")
	default:
		logger.Error("Unknown mode", "mode", cfg.Mode)
		os.Exit(1)
	}
}

func createMCPServer(registry *registry.Registry) *server.MCPServer {
	s := server.NewMCPServer(
		"OpenAPI MCP Server",
		"1.0.0",
		server.WithToolCapabilities(false),
	)

	// Register tools
	registerTools(s, registry)

	return s
}

func registerTools(s *server.MCPServer, reg *registry.Registry) {
	// Create handler instances with registry
	h := handlers.New(reg, logger)

	// Register list_apis tool
	listAPIsTool := mcp.NewTool("list_apis",
		mcp.WithDescription("List all available API specifications"),
	)
	s.AddTool(listAPIsTool, h.ListAPIs)

	// Register get_api_overview tool
	overviewTool := mcp.NewTool("get_api_overview",
		mcp.WithDescription("Get a comprehensive overview of an API"),
		mcp.WithString("api_id",
			mcp.Required(),
			mcp.Description("API identifier (name or alias)"),
		),
	)
	s.AddTool(overviewTool, h.GetAPIOverview)

	// Register find_operations tool
	findTool := mcp.NewTool("find_operations",
		mcp.WithDescription("Search for API operations by keyword or regex pattern"),
		mcp.WithString("api_id",
			mcp.Required(),
			mcp.Description("API identifier"),
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search keyword or regex pattern (e.g., 'user', '^GET.*pet', '/users?/\\d+'"),
		),
	)
	s.AddTool(findTool, h.FindOperations)

	// Register list_operations tool
	listOpsTool := mcp.NewTool("list_operations",
		mcp.WithDescription("List all operations in an API, optionally filtered by tag"),
		mcp.WithString("api_id",
			mcp.Required(),
			mcp.Description("API identifier"),
		),
		mcp.WithString("tag",
			mcp.Description("Optional tag to filter operations"),
		),
	)
	s.AddTool(listOpsTool, h.ListOperations)

	// Register get_operation_details tool
	detailsTool := mcp.NewTool("get_operation_details",
		mcp.WithDescription("Get detailed information about a specific API operation"),
		mcp.WithString("api_id",
			mcp.Required(),
			mcp.Description("API identifier"),
		),
		mcp.WithString("operation_id",
			mcp.Required(),
			mcp.Description("Operation ID or 'METHOD /path' format"),
		),
	)
	s.AddTool(detailsTool, h.GetOperationDetails)
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level

	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
