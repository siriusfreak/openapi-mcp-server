package registry

import (
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/siriusfreak/openapi-mcp-server/internal/models"
)

func TestFindOperations(t *testing.T) {
	// Create test logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create registry with test data
	reg := New(logger)

	// Create test API with various operations
	testAPI := &models.APIInfo{
		ID:      "test_api",
		Name:    "Test API",
		Version: "1.0.0",
		Operations: []*models.OperationInfo{
			{
				ID:          "getPet",
				Method:      "GET",
				Path:        "/pets/{petId}",
				Summary:     "Get a pet by ID",
				Description: "Returns a single pet",
				Tags:        []string{"pets"},
			},
			{
				ID:          "listPets",
				Method:      "GET",
				Path:        "/pets",
				Summary:     "List all pets",
				Description: "Returns an array of pets",
				Tags:        []string{"pets"},
			},
			{
				ID:          "createPet",
				Method:      "POST",
				Path:        "/pets",
				Summary:     "Create a new pet",
				Description: "Creates a new pet in the store",
				Tags:        []string{"pets"},
			},
			{
				ID:          "updatePet",
				Method:      "PUT",
				Path:        "/pets/{petId}",
				Summary:     "Update an existing pet",
				Description: "Updates a pet",
				Tags:        []string{"pets"},
			},
			{
				ID:          "deletePet",
				Method:      "DELETE",
				Path:        "/pets/{petId}",
				Summary:     "Delete a pet",
				Description: "Deletes a pet from the store",
				Tags:        []string{"pets"},
			},
			{
				ID:          "getUser",
				Method:      "GET",
				Path:        "/users/{userId}",
				Summary:     "Get user by ID",
				Description: "Returns a single user",
				Tags:        []string{"users"},
			},
			{
				ID:          "listUsers",
				Method:      "GET",
				Path:        "/users",
				Summary:     "List all users",
				Description: "Returns an array of users",
				Tags:        []string{"users"},
			},
			{
				ID:          "getUserOrders",
				Method:      "GET",
				Path:        "/users/{userId}/orders",
				Summary:     "Get user orders",
				Description: "Returns orders for a specific user",
				Tags:        []string{"users", "orders"},
			},
		},
	}

	// Add test API to registry
	reg.apis["test_api"] = testAPI

	tests := []struct {
		name        string
		query       string
		expectCount int
		expectFirst string // ID of first expected result
		isRegex     bool
	}{
		// Simple string searches
		{
			name:        "Simple keyword search - pet",
			query:       "pet",
			expectCount: 5,
			expectFirst: "getPet",
			isRegex:     false,
		},
		{
			name:        "Simple keyword search - user",
			query:       "user",
			expectCount: 3,
			expectFirst: "getUser",
			isRegex:     false,
		},
		{
			name:        "Case insensitive search",
			query:       "PET",
			expectCount: 5,
			expectFirst: "getPet",
			isRegex:     false,
		},

		// Regex searches
		{
			name:        "Regex: All GET operations",
			query:       "^GET",
			expectCount: 5,
			expectFirst: "getPet",
			isRegex:     true,
		},
		{
			name:        "Regex: Paths with IDs",
			query:       "/\\w+/{\\w+Id}",
			expectCount: 4,
			expectFirst: "getPet",
			isRegex:     true,
		},
		{
			name:        "Regex: Operations ending with 'Pet'",
			query:       ".*Pet$",
			expectCount: 5,
			expectFirst: "getPet",
			isRegex:     true,
		},
		{
			name:        "Regex: POST or PUT operations",
			query:       "^(POST|PUT)",
			expectCount: 2,
			expectFirst: "createPet",
			isRegex:     true,
		},
		{
			name:        "Regex: User-related paths",
			query:       "/users?(/|$)",
			expectCount: 3,
			expectFirst: "getUser",
			isRegex:     true,
		},
		{
			name:        "Regex: Operations with 'create' or 'update'",
			query:       "(?i)(create|update)",
			expectCount: 2,
			expectFirst: "createPet",
			isRegex:     true,
		},
		{
			name:        "Regex: Any delete operation",
			query:       "(?i)delete",
			expectCount: 1,
			expectFirst: "deletePet",
			isRegex:     true,
		},
		{
			name:        "Invalid regex falls back to simple search",
			query:       "[invalid",
			expectCount: 0,
			expectFirst: "",
			isRegex:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := reg.FindOperations("test_api", tt.query)
			if err != nil {
				t.Fatalf("FindOperations returned error: %v", err)
			}

			if len(results) != tt.expectCount {
				t.Errorf("Expected %d results, got %d", tt.expectCount, len(results))
				for _, op := range results {
					t.Logf("  Found: %s %s %s", op.ID, op.Method, op.Path)
				}
			}

			if tt.expectCount > 0 && len(results) > 0 {
				if results[0].ID != tt.expectFirst {
					t.Errorf("Expected first result ID to be %s, got %s",
						tt.expectFirst, results[0].ID)
				}
			}
		})
	}
}

func TestContainsRegexChars(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"simple", false},
		{"user", false},
		{"get_pet", false},
		{"^GET", true},
		{"user.*", true},
		{"/pets?", true},
		{"[a-z]+", true},
		{"(GET|POST)", true},
		{"users$", true},
		{"\\d+", true},
		{"normal-string", false},
		{"under_score", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsRegexChars(tt.input)
			if result != tt.expected {
				t.Errorf("containsRegexChars(%q) = %v, want %v",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestFindOperationsWithLimit(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	reg := New(logger)

	// Create API with many operations
	testAPI := &models.APIInfo{
		ID:         "test_api",
		Name:       "Test API",
		Operations: make([]*models.OperationInfo, 0, 20),
	}

	// Add 20 operations
	for i := 0; i < 20; i++ {
		testAPI.Operations = append(testAPI.Operations, &models.OperationInfo{
			ID:      fmt.Sprintf("operation_%d", i),
			Method:  "GET",
			Path:    fmt.Sprintf("/resource_%d", i),
			Summary: "Test operation",
		})
	}

	reg.apis["test_api"] = testAPI

	// Search for all operations (should be limited to 10)
	results, err := reg.FindOperations("test_api", "operation")
	if err != nil {
		t.Fatalf("FindOperations returned error: %v", err)
	}

	if len(results) != 10 {
		t.Errorf("Expected results to be limited to 10, got %d", len(results))
	}
}
