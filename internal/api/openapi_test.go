package api

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewOpenAPIImporter(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		wantErr bool
		errType ImportErrorType
	}{
		{
			name:    "valid OpenAPI 3.0 JSON",
			fixture: "minimal-3.0.json",
			wantErr: false,
		},
		{
			name:    "valid OpenAPI 3.0 YAML",
			fixture: "minimal-3.0.yaml",
			wantErr: false,
		},
		{
			name:    "valid OpenAPI 3.1 JSON",
			fixture: "petstore-3.1.json",
			wantErr: false,
		},
		{
			name:    "complex refs YAML",
			fixture: "complex-refs.yaml",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := readTestFixture(t, tt.fixture)

			importer, err := NewOpenAPIImporter(data)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				var importErr *ImportError
				if !errors.As(err, &importErr) {
					t.Errorf("expected ImportError, got %T", err)
					return
				}
				if importErr.Type != tt.errType {
					t.Errorf("expected error type %v, got %v", tt.errType, importErr.Type)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if importer == nil {
				t.Error("expected importer, got nil")
			}
		})
	}
}

func TestNewOpenAPIImporterFromFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		wantErr  bool
		errType  ImportErrorType
	}{
		{
			name:     "valid file",
			filePath: getTestFixturePath("minimal-3.0.json"),
			wantErr:  false,
		},
		{
			name:     "file not found",
			filePath: "/nonexistent/path/file.yaml",
			wantErr:  true,
			errType:  ErrFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importer, err := NewOpenAPIImporterFromFile(tt.filePath)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				var importErr *ImportError
				if !errors.As(err, &importErr) {
					t.Errorf("expected ImportError, got %T", err)
					return
				}
				if importErr.Type != tt.errType {
					t.Errorf("expected error type %v, got %v", tt.errType, importErr.Type)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if importer == nil {
				t.Error("expected importer, got nil")
			}
		})
	}
}

func TestOpenAPIImporter_GetVersion(t *testing.T) {
	tests := []struct {
		name        string
		fixture     string
		wantVersion string
	}{
		{
			name:        "OpenAPI 3.0.3",
			fixture:     "minimal-3.0.json",
			wantVersion: "3.0.3",
		},
		{
			name:        "OpenAPI 3.1.0",
			fixture:     "petstore-3.1.json",
			wantVersion: "3.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := readTestFixture(t, tt.fixture)
			importer, err := NewOpenAPIImporter(data)
			if err != nil {
				t.Fatalf("failed to create importer: %v", err)
			}

			version := importer.GetVersion()
			if version != tt.wantVersion {
				t.Errorf("expected version %q, got %q", tt.wantVersion, version)
			}
		})
	}
}

func TestOpenAPIImporter_ValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		wantErr bool
		errType ImportErrorType
	}{
		{
			name:    "OpenAPI 3.0 supported",
			fixture: "minimal-3.0.json",
			wantErr: false,
		},
		{
			name:    "OpenAPI 3.1 supported",
			fixture: "petstore-3.1.json",
			wantErr: false,
		},
		{
			name:    "Swagger 2.0 rejected",
			fixture: "swagger-2.0.json",
			wantErr: true,
			errType: ErrUnsupportedVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := readTestFixture(t, tt.fixture)
			importer, err := NewOpenAPIImporter(data)
			if err != nil {
				t.Fatalf("failed to create importer: %v", err)
			}

			err = importer.ValidateVersion()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				var importErr *ImportError
				if !errors.As(err, &importErr) {
					t.Errorf("expected ImportError, got %T", err)
					return
				}
				if importErr.Type != tt.errType {
					t.Errorf("expected error type %v, got %v", tt.errType, importErr.Type)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestOpenAPIImporter_Preview(t *testing.T) {
	tests := []struct {
		name             string
		fixture          string
		wantTitle        string
		wantEndpoints    int
		wantFolders      int
		wantServers      int
		wantErr          bool
		checkFolderNames []string
	}{
		{
			name:             "minimal 3.0 spec",
			fixture:          "minimal-3.0.json",
			wantTitle:        "Minimal API",
			wantEndpoints:    3, // GET /health, GET /users, POST /users
			wantFolders:      2, // System, Users
			wantServers:      1,
			checkFolderNames: []string{"System", "Users"},
		},
		{
			name:             "petstore 3.1 spec",
			fixture:          "petstore-3.1.json",
			wantTitle:        "Petstore API",
			wantEndpoints:    6, // GET/POST /pets, GET/PUT/DELETE /pets/{id}, GET /store/inventory
			wantFolders:      2, // Pets, Store
			wantServers:      2,
			checkFolderNames: []string{"Pets", "Store"},
		},
		{
			name:             "no tags spec",
			fixture:          "no-tags.yaml",
			wantTitle:        "No Tags API",
			wantEndpoints:    6,
			wantFolders:      1,
			wantServers:      1,
			checkFolderNames: []string{"Untagged"},
		},
		{
			name:    "swagger 2.0 rejected",
			fixture: "swagger-2.0.json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := readTestFixture(t, tt.fixture)
			importer, err := NewOpenAPIImporter(data)
			if err != nil {
				t.Fatalf("failed to create importer: %v", err)
			}

			preview, err := importer.Preview()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if preview.Title != tt.wantTitle {
				t.Errorf("expected title %q, got %q", tt.wantTitle, preview.Title)
			}

			if preview.EndpointCount != tt.wantEndpoints {
				t.Errorf("expected %d endpoints, got %d", tt.wantEndpoints, preview.EndpointCount)
			}

			if preview.FolderCount != tt.wantFolders {
				t.Errorf("expected %d folders, got %d", tt.wantFolders, preview.FolderCount)
			}

			if len(preview.Servers) != tt.wantServers {
				t.Errorf("expected %d servers, got %d", tt.wantServers, len(preview.Servers))
			}

			// Check folder names
			for _, expectedName := range tt.checkFolderNames {
				found := false
				for _, folder := range preview.Folders {
					if folder.Name == expectedName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected folder %q not found in preview", expectedName)
				}
			}
		})
	}
}

func TestOpenAPIImporter_ToCollection(t *testing.T) {
	tests := []struct {
		name           string
		fixture        string
		opts           ImportOptions
		wantName       string
		wantFolders    int
		wantTotalReqs  int
		wantErr        bool
		checkFirstURL  string
		checkHasMethod HTTPMethod
	}{
		{
			name:    "minimal spec default options",
			fixture: "minimal-3.0.json",
			opts: ImportOptions{
				IncludeExamples: true,
			},
			wantName:       "Minimal API",
			wantFolders:    2,
			wantTotalReqs:  3,
			checkFirstURL:  "https://api.example.com/v1",
			checkHasMethod: GET,
		},
		{
			name:    "minimal spec with custom name",
			fixture: "minimal-3.0.json",
			opts: ImportOptions{
				Name:            "Custom API Name",
				IncludeExamples: false,
			},
			wantName:    "Custom API Name",
			wantFolders: 2,
		},
		{
			name:    "minimal spec with custom base URL",
			fixture: "minimal-3.0.json",
			opts: ImportOptions{
				BaseURL:         "https://custom.api.com",
				IncludeExamples: true,
			},
			wantName:      "Minimal API",
			wantFolders:   2,
			checkFirstURL: "https://custom.api.com",
		},
		{
			name:    "petstore spec",
			fixture: "petstore-3.1.json",
			opts: ImportOptions{
				IncludeExamples: true,
			},
			wantName:      "Petstore API",
			wantFolders:   2,
			wantTotalReqs: 6,
		},
		{
			name:    "no tags creates Untagged folder",
			fixture: "no-tags.yaml",
			opts: ImportOptions{
				IncludeExamples: true,
			},
			wantName:    "No Tags API",
			wantFolders: 1, // Just "Untagged"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := readTestFixture(t, tt.fixture)
			importer, err := NewOpenAPIImporter(data)
			if err != nil {
				t.Fatalf("failed to create importer: %v", err)
			}

			collection, err := importer.ToCollection(tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if collection.Name != tt.wantName {
				t.Errorf("expected name %q, got %q", tt.wantName, collection.Name)
			}

			if len(collection.Folders) != tt.wantFolders {
				t.Errorf("expected %d folders, got %d", tt.wantFolders, len(collection.Folders))
			}

			// Count total requests
			if tt.wantTotalReqs > 0 {
				totalReqs := countRequests(collection)
				if totalReqs != tt.wantTotalReqs {
					t.Errorf("expected %d total requests, got %d", tt.wantTotalReqs, totalReqs)
				}
			}

			// Check first URL contains expected base
			if tt.checkFirstURL != "" {
				found := false
				for _, folder := range collection.Folders {
					for _, req := range folder.Requests {
						if strings.HasPrefix(req.URL, tt.checkFirstURL) {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					t.Errorf("expected URL starting with %q not found", tt.checkFirstURL)
				}
			}

			// Check for expected HTTP method
			if tt.checkHasMethod != "" {
				found := false
				for _, folder := range collection.Folders {
					for _, req := range folder.Requests {
						if req.Method == tt.checkHasMethod {
							found = true
							break
						}
					}
					if found {
						break
					}
				}
				if !found {
					t.Errorf("expected method %s not found", tt.checkHasMethod)
				}
			}
		})
	}
}

// Helper functions

func readTestFixture(t *testing.T, filename string) []byte {
	t.Helper()
	path := getTestFixturePath(filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", filename, err)
	}
	return data
}

func getTestFixturePath(filename string) string {
	return filepath.Join("..", "..", "testdata", "openapi", filename)
}

func countRequests(c *CollectionFile) int {
	count := len(c.Requests)
	for _, folder := range c.Folders {
		count += countFolderRequests(&folder)
	}
	return count
}

func countFolderRequests(f *Folder) int {
	count := len(f.Requests)
	for _, subfolder := range f.Folders {
		count += countFolderRequests(&subfolder)
	}
	return count
}

// T066: Tests for $ref resolution
func TestOpenAPIImporter_RefResolution(t *testing.T) {
	// Test that local $refs are properly resolved
	data := readTestFixture(t, "complex-refs.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert to collection: %v", err)
	}

	// Verify that refs to components/schemas are resolved
	// The complex-refs.yaml has $ref to Order, Customer, Product schemas
	foundOrdersFolder := false
	foundCustomersFolder := false
	foundProductsFolder := false

	for _, folder := range collection.Folders {
		switch folder.Name {
		case "Orders":
			foundOrdersFolder = true
			// Should have list and create operations
			if len(folder.Requests) < 2 {
				t.Errorf("Orders folder should have at least 2 requests, got %d", len(folder.Requests))
			}
		case "Customers":
			foundCustomersFolder = true
		case "Products":
			foundProductsFolder = true
		}
	}

	if !foundOrdersFolder {
		t.Error("expected Orders folder from tags")
	}
	if !foundCustomersFolder {
		t.Error("expected Customers folder from tags")
	}
	if !foundProductsFolder {
		t.Error("expected Products folder from tags")
	}

	// Verify that requests were generated from the spec
	// complex-refs.yaml defines operations that should be converted
	totalRequests := 0
	for _, folder := range collection.Folders {
		totalRequests += len(folder.Requests)
	}
	if totalRequests == 0 {
		t.Error("Expected requests to be generated from spec with $ref parameters")
	}
}

// T066: Test that $ref to requestBodies are resolved
func TestOpenAPIImporter_RefResolution_RequestBody(t *testing.T) {
	data := readTestFixture(t, "complex-refs.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert to collection: %v", err)
	}

	// Find POST /orders which has $ref to requestBodies/OrderBody
	for _, folder := range collection.Folders {
		for _, req := range folder.Requests {
			if req.Method == POST && strings.Contains(req.URL, "/orders") {
				// Should have a body from the $ref
				if req.Body == nil {
					t.Error("POST /orders should have a body from $ref requestBodies/OrderBody")
				} else if req.Body.Type == "" {
					t.Error("POST /orders body should have a type set")
				}
				// Should have Content-Type header
				hasContentType := false
				for _, h := range req.Headers {
					if h.Key == "Content-Type" {
						hasContentType = true
					}
				}
				if !hasContentType {
					t.Error("POST /orders should have Content-Type header")
				}
			}
		}
	}
}

// T067: Test circular reference handling
func TestOpenAPIImporter_CircularRefHandling(t *testing.T) {
	// complex-refs.yaml has Category schema with self-reference (parent: $ref Category)
	data := readTestFixture(t, "complex-refs.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	// This should complete without hanging due to circular refs
	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert to collection (may be circular ref issue): %v", err)
	}

	// Just verify we got a valid collection
	if collection == nil {
		t.Fatal("expected collection, got nil")
	}

	if collection.Name == "" {
		t.Error("expected collection name to be set")
	}

	// Verify folders were created despite circular refs
	if len(collection.Folders) == 0 {
		t.Error("expected at least one folder")
	}
}

// T067: Test deeply nested ref handling
func TestOpenAPIImporter_DeepNestedRefHandling(t *testing.T) {
	// complex-refs.yaml has Order -> Customer -> Address (nested refs)
	// and Order -> OrderItem -> Product -> Category -> Category (circular at end)
	data := readTestFixture(t, "complex-refs.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	// Test preview also works with circular refs
	preview, err := importer.Preview()
	if err != nil {
		t.Fatalf("preview failed (may be circular ref issue): %v", err)
	}

	// Should have endpoint count
	if preview.EndpointCount == 0 {
		t.Error("expected endpoints in preview")
	}

	// Should have folders
	if preview.FolderCount == 0 {
		t.Error("expected folders in preview")
	}
}
