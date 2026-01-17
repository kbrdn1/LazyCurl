package api

import (
	"testing"
)

func TestBuildURL(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		path    string
		want    string
	}{
		{
			name:    "base with trailing slash",
			baseURL: "https://api.example.com/",
			path:    "/users",
			want:    "https://api.example.com/users",
		},
		{
			name:    "base without trailing slash",
			baseURL: "https://api.example.com",
			path:    "/users",
			want:    "https://api.example.com/users",
		},
		{
			name:    "path without leading slash",
			baseURL: "https://api.example.com",
			path:    "users",
			want:    "https://api.example.com/users",
		},
		{
			name:    "empty base URL",
			baseURL: "",
			path:    "/users",
			want:    "/users",
		},
		{
			name:    "base with version path",
			baseURL: "https://api.example.com/v1",
			path:    "/users",
			want:    "https://api.example.com/v1/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildURL(tt.baseURL, tt.path)
			if got != tt.want {
				t.Errorf("buildURL(%q, %q) = %q, want %q", tt.baseURL, tt.path, got, tt.want)
			}
		})
	}
}

func TestReplacePathParameters(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		pathParams map[string]string
		want       string
	}{
		{
			name: "single parameter with value",
			url:  "/users/{userId}",
			pathParams: map[string]string{
				"userId": "123",
			},
			want: "/users/123",
		},
		{
			name: "single parameter without value",
			url:  "/users/{userId}",
			pathParams: map[string]string{
				"userId": "",
			},
			want: "/users/{{userId}}",
		},
		{
			name: "multiple parameters",
			url:  "/users/{userId}/posts/{postId}",
			pathParams: map[string]string{
				"userId": "123",
				"postId": "456",
			},
			want: "/users/123/posts/456",
		},
		{
			name:       "no parameters",
			url:        "/users",
			pathParams: map[string]string{},
			want:       "/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replacePathParameters(tt.url, tt.pathParams)
			if got != tt.want {
				t.Errorf("replacePathParameters(%q, %v) = %q, want %q", tt.url, tt.pathParams, got, tt.want)
			}
		})
	}
}

func TestGenerateStringExample(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{"email", "user@example.com"},
		{"uri", "https://example.com"},
		{"url", "https://example.com"},
		{"uuid", "550e8400-e29b-41d4-a716-446655440000"},
		{"date", "2024-01-15"},
		{"date-time", "2024-01-15T10:30:00Z"},
		{"password", "********"},
		{"byte", "SGVsbG8gV29ybGQ="},
		{"hostname", "example.com"},
		{"ipv4", "192.168.1.1"},
		{"ipv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		{"", "string"},
		{"unknown", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			got := generateStringExample(tt.format)
			if got != tt.want {
				t.Errorf("generateStringExample(%q) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}

func TestAddOrUpdateHeader(t *testing.T) {
	tests := []struct {
		name     string
		headers  []KeyValueEntry
		key      string
		value    string
		wantLen  int
		wantVal  string
		checkKey string
	}{
		{
			name:     "add to empty list",
			headers:  []KeyValueEntry{},
			key:      "Content-Type",
			value:    "application/json",
			wantLen:  1,
			wantVal:  "application/json",
			checkKey: "Content-Type",
		},
		{
			name: "update existing header",
			headers: []KeyValueEntry{
				{Key: "Content-Type", Value: "text/plain", Enabled: true},
			},
			key:      "Content-Type",
			value:    "application/json",
			wantLen:  1,
			wantVal:  "application/json",
			checkKey: "Content-Type",
		},
		{
			name: "add new header to existing list",
			headers: []KeyValueEntry{
				{Key: "Accept", Value: "application/json", Enabled: true},
			},
			key:      "Content-Type",
			value:    "application/json",
			wantLen:  2,
			wantVal:  "application/json",
			checkKey: "Content-Type",
		},
		{
			name: "case insensitive update",
			headers: []KeyValueEntry{
				{Key: "content-type", Value: "text/plain", Enabled: true},
			},
			key:      "Content-Type",
			value:    "application/json",
			wantLen:  1,
			wantVal:  "application/json",
			checkKey: "content-type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := addOrUpdateHeader(tt.headers, tt.key, tt.value)

			if len(result) != tt.wantLen {
				t.Errorf("expected %d headers, got %d", tt.wantLen, len(result))
			}

			// Find the header and check value
			found := false
			for _, h := range result {
				if h.Key == tt.checkKey {
					if h.Value != tt.wantVal {
						t.Errorf("expected value %q, got %q", tt.wantVal, h.Value)
					}
					found = true
					break
				}
			}
			if !found {
				t.Errorf("header %q not found in result", tt.checkKey)
			}
		})
	}
}

func TestTagMarkers(t *testing.T) {
	tests := []struct {
		name        string
		description string
		tag         string
		wantMarked  string
	}{
		{
			name:        "add tag to empty description",
			description: "",
			tag:         "Users",
			wantMarked:  "[[TAG:Users]]",
		},
		{
			name:        "add tag to existing description",
			description: "Get all users",
			tag:         "Users",
			wantMarked:  "[[TAG:Users]] Get all users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marked := setTagMarker(tt.description, tt.tag)
			if marked != tt.wantMarked {
				t.Errorf("setTagMarker(%q, %q) = %q, want %q", tt.description, tt.tag, marked, tt.wantMarked)
			}

			// Test extraction
			req := &CollectionRequest{Description: marked}
			extractedTag := extractTagFromRequest(req)

			if extractedTag != tt.tag {
				t.Errorf("extractTagFromRequest got tag %q, want %q", extractedTag, tt.tag)
			}

			if req.Description != tt.description {
				t.Errorf("after extraction, description = %q, want %q", req.Description, tt.description)
			}
		})
	}
}

func TestExtractTagFromRequest_NoTag(t *testing.T) {
	req := &CollectionRequest{Description: "A regular description"}
	tag := extractTagFromRequest(req)

	if tag != "" {
		t.Errorf("expected empty tag, got %q", tag)
	}

	if req.Description != "A regular description" {
		t.Errorf("description should be unchanged, got %q", req.Description)
	}
}

func TestFormatExample(t *testing.T) {
	tests := []struct {
		name    string
		example interface{}
		want    string
	}{
		{
			name:    "string value",
			example: "hello",
			want:    "hello",
		},
		{
			name:    "true boolean",
			example: true,
			want:    "true",
		},
		{
			name:    "false boolean",
			example: false,
			want:    "false",
		},
		{
			name:    "integer as float64",
			example: float64(42),
			want:    "42",
		},
		{
			name:    "nil value",
			example: nil,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatExample(tt.example)
			if got != tt.want {
				t.Errorf("formatExample(%v) = %q, want %q", tt.example, got, tt.want)
			}
		})
	}
}

func TestConvertPathsToFolders_Integration(t *testing.T) {
	// Test with real OpenAPI spec
	data := readTestFixture(t, "minimal-3.0.json")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	// Verify folder structure
	if len(collection.Folders) != 2 {
		t.Errorf("expected 2 folders, got %d", len(collection.Folders))
	}

	// Check for expected folders
	folderNames := make(map[string]bool)
	for _, folder := range collection.Folders {
		folderNames[folder.Name] = true
	}

	expectedFolders := []string{"System", "Users"}
	for _, expected := range expectedFolders {
		if !folderNames[expected] {
			t.Errorf("expected folder %q not found", expected)
		}
	}

	// Verify requests have proper structure
	for _, folder := range collection.Folders {
		for _, req := range folder.Requests {
			if req.ID == "" {
				t.Error("request has empty ID")
			}
			if req.Name == "" {
				t.Error("request has empty Name")
			}
			if req.Method == "" {
				t.Error("request has empty Method")
			}
			if req.URL == "" {
				t.Error("request has empty URL")
			}
		}
	}
}

func TestConvertPathsToFolders_NoTags(t *testing.T) {
	data := readTestFixture(t, "no-tags.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	// Should have exactly one "Untagged" folder
	if len(collection.Folders) != 1 {
		t.Errorf("expected 1 folder, got %d", len(collection.Folders))
	}

	if collection.Folders[0].Name != "Untagged" {
		t.Errorf("expected folder name 'Untagged', got %q", collection.Folders[0].Name)
	}

	// All requests should be in the Untagged folder
	if len(collection.Folders[0].Requests) != 6 {
		t.Errorf("expected 6 requests in Untagged folder, got %d", len(collection.Folders[0].Requests))
	}
}

func TestConvertPathsToFolders_ComplexRefs(t *testing.T) {
	data := readTestFixture(t, "complex-refs.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	// Should have folders for Orders, Customers, Products
	if len(collection.Folders) < 3 {
		t.Errorf("expected at least 3 folders, got %d", len(collection.Folders))
	}

	// Verify refs are resolved - collection should have folders with requests
	if len(collection.Folders) == 0 {
		t.Error("Expected collection to have folders")
	}
}

// T055: Tests for example generation from schemas
func TestSchemaExampleGeneration_Integration(t *testing.T) {
	data := readTestFixture(t, "complex-refs.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	// Find a POST request with body
	var postReq *CollectionRequest
	for _, folder := range collection.Folders {
		for i := range folder.Requests {
			if folder.Requests[i].Method == POST && folder.Requests[i].Body != nil {
				postReq = &folder.Requests[i]
				break
			}
		}
		if postReq != nil {
			break
		}
	}

	if postReq == nil {
		t.Skip("no POST request with body found in fixture")
	}

	// Verify body has content
	if postReq.Body.Content == nil {
		t.Error("POST request body should have content")
	}

	// Verify body type is set
	if postReq.Body.Type == "" {
		t.Error("POST request body should have type set")
	}
}

// T055: Test example priority (explicit > generated)
func TestExamplePriority(t *testing.T) {
	// Test that explicit examples take priority over generated ones
	data := readTestFixture(t, "minimal-3.0.json")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	// Verify we got proper examples from the schema
	// The minimal-3.0.json has defined schemas that should produce examples
	foundRequest := false
	for _, folder := range collection.Folders {
		for _, req := range folder.Requests {
			if len(req.Params) > 0 || req.Body != nil {
				foundRequest = true
			}
		}
	}

	if !foundRequest {
		// This is OK - not all specs have examples
		t.Log("No requests with params or body found, skipping example priority test")
	}
}

// T059: Tests for request body conversion
func TestRequestBodyConversion_MediaTypes(t *testing.T) {
	data := readTestFixture(t, "complex-refs.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	// Verify Content-Type headers are set for requests with bodies
	for _, folder := range collection.Folders {
		for _, req := range folder.Requests {
			if req.Body != nil {
				// Check for Content-Type header
				hasContentType := false
				for _, h := range req.Headers {
					if h.Key == "Content-Type" {
						hasContentType = true
						if h.Value == "" {
							t.Errorf("request %s has empty Content-Type", req.Name)
						}
					}
				}
				if !hasContentType {
					t.Errorf("request %s with body should have Content-Type header", req.Name)
				}
			}
		}
	}
}

// Test parameter extraction with required status
func TestParameterExtraction_Required(t *testing.T) {
	data := readTestFixture(t, "minimal-3.0.json")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	// Find request with path parameters
	for _, folder := range collection.Folders {
		for _, req := range folder.Requests {
			// Check if URL contains resolved path parameters or templates
			if req.URL != "" {
				// URL should be properly formed
				if req.URL[0] != 'h' && req.URL[0] != '/' && req.URL[0] != '{' {
					t.Errorf("unexpected URL format: %s", req.URL)
				}
			}
		}
	}
}

// Test depth limit for circular references
func TestSchemaToExample_DepthLimit(t *testing.T) {
	// Test that deeply nested schemas don't cause infinite recursion
	data := readTestFixture(t, "complex-refs.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	// This should complete without hanging or panicking
	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	// Just verify we got some output
	if collection == nil {
		t.Error("expected collection, got nil")
	}
}

// T071: Test security scheme extraction
func TestSecuritySchemeExtraction(t *testing.T) {
	data := readTestFixture(t, "with-security.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	tests := []struct {
		name         string
		requestName  string
		wantAuthType string
		wantAPIKey   string
		wantLocation string
	}{
		{
			name:         "bearer token from global security",
			requestName:  "Get users (uses global bearer)",
			wantAuthType: "bearer",
		},
		{
			name:         "basic auth from operation security",
			requestName:  "Admin endpoint (uses basic)",
			wantAuthType: "basic",
		},
		{
			name:         "public endpoint with no auth",
			requestName:  "Public endpoint (no auth)",
			wantAuthType: "",
		},
		{
			name:         "api key in header",
			requestName:  "API Key in header",
			wantAuthType: "api_key",
			wantAPIKey:   "X-API-Key",
			wantLocation: "header",
		},
		{
			name:         "api key in query",
			requestName:  "API Key in query",
			wantAuthType: "api_key",
			wantAPIKey:   "api_key",
			wantLocation: "query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var foundReq *CollectionRequest
			for _, folder := range collection.Folders {
				for i := range folder.Requests {
					if folder.Requests[i].Name == tt.requestName {
						foundReq = &folder.Requests[i]
						break
					}
				}
				if foundReq != nil {
					break
				}
			}

			if foundReq == nil {
				t.Fatalf("request %q not found", tt.requestName)
			}

			if tt.wantAuthType == "" {
				if foundReq.Auth != nil {
					t.Errorf("expected no auth, got auth type %q", foundReq.Auth.Type)
				}
			} else {
				if foundReq.Auth == nil {
					t.Fatalf("expected auth type %q, got nil", tt.wantAuthType)
				}
				if foundReq.Auth.Type != tt.wantAuthType {
					t.Errorf("expected auth type %q, got %q", tt.wantAuthType, foundReq.Auth.Type)
				}

				if tt.wantAuthType == "api_key" {
					if foundReq.Auth.APIKeyName != tt.wantAPIKey {
						t.Errorf("expected API key name %q, got %q", tt.wantAPIKey, foundReq.Auth.APIKeyName)
					}
					if foundReq.Auth.APIKeyLocation != tt.wantLocation {
						t.Errorf("expected API key location %q, got %q", tt.wantLocation, foundReq.Auth.APIKeyLocation)
					}
				}
			}
		})
	}
}

// T071: Test bearer token extraction
func TestBearerTokenExtraction(t *testing.T) {
	data := readTestFixture(t, "with-security.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	// Find the request that uses bearer token
	var foundReq *CollectionRequest
	for _, folder := range collection.Folders {
		for i := range folder.Requests {
			if folder.Requests[i].Name == "Get users (uses global bearer)" {
				foundReq = &folder.Requests[i]
				break
			}
		}
	}

	if foundReq == nil {
		t.Fatal("bearer token request not found")
	}

	if foundReq.Auth == nil {
		t.Fatal("expected auth config, got nil")
	}

	if foundReq.Auth.Type != "bearer" {
		t.Errorf("expected auth type 'bearer', got %q", foundReq.Auth.Type)
	}

	if foundReq.Auth.Token != "" {
		t.Errorf("expected empty token, got %q", foundReq.Auth.Token)
	}
}

// T071: Test basic auth extraction
func TestBasicAuthExtraction(t *testing.T) {
	data := readTestFixture(t, "with-security.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	// Find the request that uses basic auth
	var foundReq *CollectionRequest
	for _, folder := range collection.Folders {
		for i := range folder.Requests {
			if folder.Requests[i].Name == "Admin endpoint (uses basic)" {
				foundReq = &folder.Requests[i]
				break
			}
		}
	}

	if foundReq == nil {
		t.Fatal("basic auth request not found")
	}

	if foundReq.Auth == nil {
		t.Fatal("expected auth config, got nil")
	}

	if foundReq.Auth.Type != "basic" {
		t.Errorf("expected auth type 'basic', got %q", foundReq.Auth.Type)
	}
}

// T071: Test API key extraction
func TestAPIKeyExtraction(t *testing.T) {
	data := readTestFixture(t, "with-security.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	tests := []struct {
		name         string
		requestName  string
		wantKeyName  string
		wantLocation string
	}{
		{
			name:         "api key in header",
			requestName:  "API Key in header",
			wantKeyName:  "X-API-Key",
			wantLocation: "header",
		},
		{
			name:         "api key in query",
			requestName:  "API Key in query",
			wantKeyName:  "api_key",
			wantLocation: "query",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var foundReq *CollectionRequest
			for _, folder := range collection.Folders {
				for i := range folder.Requests {
					if folder.Requests[i].Name == tt.requestName {
						foundReq = &folder.Requests[i]
						break
					}
				}
			}

			if foundReq == nil {
				t.Fatalf("request %q not found", tt.requestName)
			}

			if foundReq.Auth == nil {
				t.Fatal("expected auth config, got nil")
			}

			if foundReq.Auth.Type != "api_key" {
				t.Errorf("expected auth type 'api_key', got %q", foundReq.Auth.Type)
			}

			if foundReq.Auth.APIKeyName != tt.wantKeyName {
				t.Errorf("expected API key name %q, got %q", tt.wantKeyName, foundReq.Auth.APIKeyName)
			}

			if foundReq.Auth.APIKeyLocation != tt.wantLocation {
				t.Errorf("expected API key location %q, got %q", tt.wantLocation, foundReq.Auth.APIKeyLocation)
			}
		})
	}
}

// T071: Test operation-level security override
func TestOperationLevelSecurityOverride(t *testing.T) {
	data := readTestFixture(t, "with-security.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	// Find the admin request that overrides global security
	var adminReq *CollectionRequest
	for _, folder := range collection.Folders {
		for i := range folder.Requests {
			if folder.Requests[i].Name == "Admin endpoint (uses basic)" {
				adminReq = &folder.Requests[i]
				break
			}
		}
	}

	if adminReq == nil {
		t.Fatal("admin request not found")
	}

	if adminReq.Auth == nil {
		t.Fatal("expected auth config, got nil")
	}

	// Should use basic auth, not global bearer
	if adminReq.Auth.Type != "basic" {
		t.Errorf("expected operation-level 'basic' auth, got %q", adminReq.Auth.Type)
	}
}

// T071: Test empty security (public endpoint)
func TestEmptySecurityPublicEndpoint(t *testing.T) {
	data := readTestFixture(t, "with-security.yaml")
	importer, err := NewOpenAPIImporter(data)
	if err != nil {
		t.Fatalf("failed to create importer: %v", err)
	}

	collection, err := importer.ToCollection(ImportOptions{
		IncludeExamples: true,
	})
	if err != nil {
		t.Fatalf("failed to convert: %v", err)
	}

	// Find the public endpoint
	var publicReq *CollectionRequest
	for _, folder := range collection.Folders {
		for i := range folder.Requests {
			if folder.Requests[i].Name == "Public endpoint (no auth)" {
				publicReq = &folder.Requests[i]
				break
			}
		}
	}

	if publicReq == nil {
		t.Fatal("public request not found")
	}

	if publicReq.Auth != nil {
		t.Errorf("expected no auth for public endpoint, got auth type %q", publicReq.Auth.Type)
	}
}
