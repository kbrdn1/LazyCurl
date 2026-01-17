package postman

import (
	"path/filepath"
	"testing"

	"github.com/kbrdn1/LazyCurl/internal/api"
)

func TestImportCollection_Simple(t *testing.T) {
	result, err := ImportCollection(filepath.Join("testdata", "simple_collection.json"))
	if err != nil {
		t.Fatalf("ImportCollection failed: %v", err)
	}

	if !result.Success() {
		t.Fatal("Expected successful import")
	}

	if result.Collection == nil {
		t.Fatal("Expected collection to be non-nil")
	}

	// Check collection name
	if result.Collection.Name != "Simple API" {
		t.Errorf("Expected name 'Simple API', got '%s'", result.Collection.Name)
	}

	// Check request count
	if result.Summary.RequestsCount != 5 {
		t.Errorf("Expected 5 requests, got %d", result.Summary.RequestsCount)
	}

	// Check folder count
	if result.Summary.FoldersCount != 0 {
		t.Errorf("Expected 0 folders, got %d", result.Summary.FoldersCount)
	}

	// Check requests
	if len(result.Collection.Requests) != 5 {
		t.Errorf("Expected 5 requests, got %d", len(result.Collection.Requests))
	}

	// Verify HTTP methods
	expectedMethods := map[string]api.HTTPMethod{
		"Get Users":   "GET",
		"Create User": "POST",
		"Update User": "PUT",
		"Delete User": "DELETE",
		"Patch User":  "PATCH",
	}
	for _, req := range result.Collection.Requests {
		if expected, ok := expectedMethods[req.Name]; ok {
			if req.Method != expected {
				t.Errorf("Request '%s': expected method %s, got %s", req.Name, expected, req.Method)
			}
		}
	}
}

func TestImportCollection_Nested(t *testing.T) {
	result, err := ImportCollection(filepath.Join("testdata", "nested_collection.json"))
	if err != nil {
		t.Fatalf("ImportCollection failed: %v", err)
	}

	if !result.Success() {
		t.Fatal("Expected successful import")
	}

	// Check collection name
	if result.Collection.Name != "Nested API" {
		t.Errorf("Expected name 'Nested API', got '%s'", result.Collection.Name)
	}

	// Check folder count (Authentication, Users, Users/Admin, Users/Admin/Permissions = 4)
	if result.Summary.FoldersCount != 4 {
		t.Errorf("Expected 4 folders, got %d", result.Summary.FoldersCount)
	}

	// Check request count
	if result.Summary.RequestsCount != 6 {
		t.Errorf("Expected 6 requests, got %d", result.Summary.RequestsCount)
	}

	// Check top-level structure
	if len(result.Collection.Folders) != 2 {
		t.Errorf("Expected 2 top-level folders, got %d", len(result.Collection.Folders))
	}
	if len(result.Collection.Requests) != 1 {
		t.Errorf("Expected 1 top-level request, got %d", len(result.Collection.Requests))
	}

	// Check nested folder depth (Users > Admin > Permissions)
	usersFolder := findFolder(result.Collection.Folders, "Users")
	if usersFolder == nil {
		t.Fatal("Expected 'Users' folder")
	}
	adminFolder := findFolder(usersFolder.Folders, "Admin")
	if adminFolder == nil {
		t.Fatal("Expected 'Admin' subfolder")
	}
	permissionsFolder := findFolder(adminFolder.Folders, "Permissions")
	if permissionsFolder == nil {
		t.Fatal("Expected 'Permissions' subfolder (3 levels deep)")
	}
}

func TestImportCollection_AllBodyTypes(t *testing.T) {
	result, err := ImportCollection(filepath.Join("testdata", "all_body_types.json"))
	if err != nil {
		t.Fatalf("ImportCollection failed: %v", err)
	}

	if !result.Success() {
		t.Fatal("Expected successful import")
	}

	tests := []struct {
		name     string
		bodyType string
	}{
		{"Raw JSON", "json"},
		{"Raw Text", "raw"},
		{"Raw XML", "raw"},
		{"URL Encoded", "form-data"},
		{"Form Data", "form-data"},
		{"No Body", ""},
	}

	for _, tt := range tests {
		req := findRequest(result.Collection.Requests, tt.name)
		if req == nil {
			t.Errorf("Request '%s' not found", tt.name)
			continue
		}

		if tt.bodyType == "" {
			if req.Body != nil {
				t.Errorf("Request '%s': expected no body, got %v", tt.name, req.Body)
			}
		} else {
			if req.Body == nil {
				t.Errorf("Request '%s': expected body, got nil", tt.name)
			} else if req.Body.Type != tt.bodyType {
				t.Errorf("Request '%s': expected body type '%s', got '%s'", tt.name, tt.bodyType, req.Body.Type)
			}
		}
	}

	// Check file upload warning
	if !result.HasWarnings() {
		t.Error("Expected warnings for file upload")
	}
}

func TestImportCollection_WithAuth(t *testing.T) {
	result, err := ImportCollection(filepath.Join("testdata", "with_auth.json"))
	if err != nil {
		t.Fatalf("ImportCollection failed: %v", err)
	}

	if !result.Success() {
		t.Fatal("Expected successful import")
	}

	tests := []struct {
		name     string
		authType string
	}{
		{"Bearer Token Auth", "bearer"},
		{"Basic Auth", "basic"},
		{"API Key Header", "api_key"},
		{"API Key Query", "api_key"},
		{"No Auth", "none"},
	}

	for _, tt := range tests {
		req := findRequest(result.Collection.Requests, tt.name)
		if req == nil {
			t.Errorf("Request '%s' not found", tt.name)
			continue
		}

		if tt.authType == "none" {
			if req.Auth != nil && req.Auth.Type != "none" {
				t.Errorf("Request '%s': expected no auth or 'none', got '%s'", tt.name, req.Auth.Type)
			}
		} else {
			if req.Auth == nil {
				t.Errorf("Request '%s': expected auth config, got nil", tt.name)
			} else if req.Auth.Type != tt.authType {
				t.Errorf("Request '%s': expected auth type '%s', got '%s'", tt.name, tt.authType, req.Auth.Type)
			}
		}
	}

	// Check OAuth2 warning
	if !result.HasWarnings() {
		t.Error("Expected warnings for OAuth2")
	}
	foundOAuthWarning := false
	for _, w := range result.Summary.Warnings {
		if contains(w, "OAuth 2.0") {
			foundOAuthWarning = true
			break
		}
	}
	if !foundOAuthWarning {
		t.Error("Expected OAuth 2.0 warning")
	}
}

func TestImportCollection_WithScripts(t *testing.T) {
	result, err := ImportCollection(filepath.Join("testdata", "with_scripts.json"))
	if err != nil {
		t.Fatalf("ImportCollection failed: %v", err)
	}

	if !result.Success() {
		t.Fatal("Expected successful import")
	}

	// Check request with scripts
	req := findRequest(result.Collection.Requests, "Request with Scripts")
	if req == nil {
		t.Fatal("Request 'Request with Scripts' not found")
	}

	if req.Scripts == nil {
		t.Error("Expected scripts to be set")
	} else {
		if req.Scripts.PreRequest == "" {
			t.Error("Expected pre-request script")
		}
		if req.Scripts.PostRequest == "" {
			t.Error("Expected post-request (test) script")
		}
	}

	// Check warnings for scripts
	if !result.HasWarnings() {
		t.Error("Expected warnings for scripts")
	}
	if len(result.Summary.Warnings) != 2 {
		t.Errorf("Expected 2 warnings (pre-request + test), got %d", len(result.Summary.Warnings))
	}

	// Check request without scripts
	simpleReq := findRequest(result.Collection.Requests, "Request without Scripts")
	if simpleReq == nil {
		t.Fatal("Request 'Request without Scripts' not found")
	}
	if simpleReq.Scripts != nil {
		t.Error("Expected no scripts for simple request")
	}
}

func TestImportCollection_DisabledHeaders(t *testing.T) {
	// Create test data with disabled headers
	jsonData := []byte(`{
		"info": {
			"name": "Headers Test",
			"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
		},
		"item": [{
			"name": "Test Request",
			"request": {
				"method": "GET",
				"header": [
					{"key": "Active-Header", "value": "value1", "disabled": false},
					{"key": "Disabled-Header", "value": "value2", "disabled": true}
				],
				"url": {"raw": "http://test.com"}
			}
		}]
	}`)

	result, err := ImportCollectionFromBytes(jsonData)
	if err != nil {
		t.Fatalf("ImportCollectionFromBytes failed: %v", err)
	}

	if len(result.Collection.Requests) != 1 {
		t.Fatal("Expected 1 request")
	}

	headers := result.Collection.Requests[0].Headers
	if len(headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(headers))
	}

	for _, h := range headers {
		switch h.Key {
		case "Active-Header":
			if !h.Enabled {
				t.Error("Active-Header should be enabled")
			}
		case "Disabled-Header":
			if h.Enabled {
				t.Error("Disabled-Header should be disabled")
			}
		}
	}
}

func TestImportCollection_QueryParams(t *testing.T) {
	jsonData := []byte(`{
		"info": {
			"name": "Query Params Test",
			"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
		},
		"item": [{
			"name": "Search",
			"request": {
				"method": "GET",
				"url": {
					"raw": "http://test.com/search?q=test&limit=10",
					"query": [
						{"key": "q", "value": "test"},
						{"key": "limit", "value": "10"},
						{"key": "disabled_param", "value": "ignore", "disabled": true}
					]
				}
			}
		}]
	}`)

	result, err := ImportCollectionFromBytes(jsonData)
	if err != nil {
		t.Fatalf("ImportCollectionFromBytes failed: %v", err)
	}

	params := result.Collection.Requests[0].Params
	if len(params) != 3 {
		t.Errorf("Expected 3 params, got %d", len(params))
	}

	// Check disabled param
	for _, p := range params {
		if p.Key == "disabled_param" && p.Enabled {
			t.Error("disabled_param should be disabled")
		}
	}
}

func TestImportCollection_EmptyCollection(t *testing.T) {
	jsonData := []byte(`{
		"info": {
			"name": "Empty Collection",
			"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
		},
		"item": []
	}`)

	result, err := ImportCollectionFromBytes(jsonData)
	if err != nil {
		t.Fatalf("ImportCollectionFromBytes failed: %v", err)
	}

	if !result.Success() {
		t.Fatal("Expected successful import")
	}

	if result.Summary.RequestsCount != 0 {
		t.Errorf("Expected 0 requests, got %d", result.Summary.RequestsCount)
	}
	if result.Summary.FoldersCount != 0 {
		t.Errorf("Expected 0 folders, got %d", result.Summary.FoldersCount)
	}
}

func TestImportCollection_InvalidJSON(t *testing.T) {
	_, err := ImportCollection(filepath.Join("testdata", "invalid_json.json"))
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}

	if !contains(err.Error(), "parse") && !contains(err.Error(), "JSON") {
		t.Errorf("Expected JSON parse error, got: %v", err)
	}
}

func TestImportCollection_NotPostmanFormat(t *testing.T) {
	_, err := ImportCollection(filepath.Join("testdata", "not_postman.json"))
	if err == nil {
		t.Fatal("Expected error for non-Postman file")
	}

	// Error can be about missing name (no info.name field) or schema validation
	if !contains(err.Error(), "schema") && !contains(err.Error(), "Postman") && !contains(err.Error(), "name") {
		t.Errorf("Expected validation error (schema/name/Postman), got: %v", err)
	}
}

func TestImportCollection_MissingName(t *testing.T) {
	jsonData := []byte(`{
		"info": {
			"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
		},
		"item": []
	}`)

	_, err := ImportCollectionFromBytes(jsonData)
	if err == nil {
		t.Fatal("Expected error for missing name")
	}

	if !contains(err.Error(), "name") {
		t.Errorf("Expected name required error, got: %v", err)
	}
}

func TestImportCollection_FileNotFound(t *testing.T) {
	_, err := ImportCollection("nonexistent_file.json")
	if err == nil {
		t.Fatal("Expected error for missing file")
	}

	if !contains(err.Error(), "read file") {
		t.Errorf("Expected file read error, got: %v", err)
	}
}

// Helper functions

func findFolder(folders []api.Folder, name string) *api.Folder {
	for i := range folders {
		if folders[i].Name == name {
			return &folders[i]
		}
	}
	return nil
}

func findRequest(requests []api.CollectionRequest, name string) *api.CollectionRequest {
	for i := range requests {
		if requests[i].Name == name {
			return &requests[i]
		}
	}
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
