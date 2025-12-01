package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCollection(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create a valid collection file
	validJSON := `{
		"name": "Test Collection",
		"description": "A test collection",
		"requests": [
			{
				"id": "req1",
				"name": "Get Users",
				"method": "GET",
				"url": "https://api.example.com/users"
			}
		]
	}`

	validPath := filepath.Join(tmpDir, "valid.json")
	if err := os.WriteFile(validPath, []byte(validJSON), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test loading valid collection
	collection, err := LoadCollection(validPath)
	if err != nil {
		t.Errorf("LoadCollection() error = %v", err)
	}
	if collection.Name != "Test Collection" {
		t.Errorf("Expected name 'Test Collection', got '%s'", collection.Name)
	}
	if len(collection.Requests) != 1 {
		t.Errorf("Expected 1 request, got %d", len(collection.Requests))
	}

	// Test loading non-existent file
	_, err = LoadCollection(filepath.Join(tmpDir, "nonexistent.json"))
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test loading invalid JSON
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	if err := os.WriteFile(invalidPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	_, err = LoadCollection(invalidPath)
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestSaveCollection(t *testing.T) {
	tmpDir := t.TempDir()

	collection := &CollectionFile{
		Name:        "Test Collection",
		Description: "A test collection",
		Requests: []CollectionRequest{
			{
				ID:     "req1",
				Name:   "Test Request",
				Method: GET,
				URL:    "https://api.example.com/test",
			},
		},
	}

	path := filepath.Join(tmpDir, "test.json")
	err := SaveCollection(collection, path)
	if err != nil {
		t.Errorf("SaveCollection() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Collection file was not created")
	}

	// Load and verify
	loaded, err := LoadCollection(path)
	if err != nil {
		t.Errorf("Failed to load saved collection: %v", err)
	}
	if loaded.Name != collection.Name {
		t.Errorf("Expected name '%s', got '%s'", collection.Name, loaded.Name)
	}
}

func TestLoadAllCollections(t *testing.T) {
	tmpDir := t.TempDir()
	collectionsDir := filepath.Join(tmpDir, "collections")
	os.MkdirAll(collectionsDir, 0755)

	// Create multiple collection files
	collections := []struct {
		name     string
		filename string
	}{
		{"Collection 1", "col1.json"},
		{"Collection 2", "col2.json"},
		{"Collection 3", "col3.json"},
	}

	for _, col := range collections {
		c := &CollectionFile{
			Name:     col.name,
			Requests: []CollectionRequest{},
		}
		if err := SaveCollection(c, filepath.Join(collectionsDir, col.filename)); err != nil {
			t.Fatalf("Failed to save collection: %v", err)
		}
	}

	// Create a non-JSON file (should be ignored)
	os.WriteFile(filepath.Join(collectionsDir, "readme.txt"), []byte("test"), 0644)

	// Load all collections
	loaded, err := LoadAllCollections(collectionsDir)
	if err != nil {
		t.Errorf("LoadAllCollections() error = %v", err)
	}

	if len(loaded) != 3 {
		t.Errorf("Expected 3 collections, got %d", len(loaded))
	}

	// Test loading from non-existent directory
	loaded, err = LoadAllCollections(filepath.Join(tmpDir, "nonexistent"))
	if err != nil {
		t.Errorf("Expected no error for non-existent directory, got %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("Expected 0 collections, got %d", len(loaded))
	}
}

func TestCollectionRequestToRequest(t *testing.T) {
	cr := &CollectionRequest{
		ID:     "test1",
		Name:   "Test Request",
		Method: POST,
		URL:    "https://api.example.com/test",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: map[string]interface{}{
			"key": "value",
		},
	}

	req := cr.ToRequest()
	if req.Method != cr.Method {
		t.Errorf("Expected method %s, got %s", cr.Method, req.Method)
	}
	if req.URL != cr.URL {
		t.Errorf("Expected URL %s, got %s", cr.URL, req.URL)
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Error("Headers not converted correctly")
	}
}

func TestFromRequest(t *testing.T) {
	req := &Request{
		Method: GET,
		URL:    "https://api.example.com/users",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
	}

	cr := FromRequest(req, "Get Users")
	if cr.Name != "Get Users" {
		t.Errorf("Expected name 'Get Users', got '%s'", cr.Name)
	}
	if cr.Method != req.Method {
		t.Errorf("Expected method %s, got %s", req.Method, cr.Method)
	}
	if cr.ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestFindRequest(t *testing.T) {
	collection := &CollectionFile{
		Name: "Test",
		Requests: []CollectionRequest{
			{ID: "req1", Name: "Request 1", Method: GET, URL: "http://test1.com"},
			{ID: "req2", Name: "Request 2", Method: POST, URL: "http://test2.com"},
		},
		Folders: []Folder{
			{
				Name: "Folder 1",
				Requests: []CollectionRequest{
					{ID: "req3", Name: "Request 3", Method: PUT, URL: "http://test3.com"},
				},
			},
		},
	}

	// Test finding top-level request
	req := collection.FindRequest("req1")
	if req == nil {
		t.Error("Expected to find req1")
	} else if req.Name != "Request 1" {
		t.Errorf("Expected 'Request 1', got '%s'", req.Name)
	}

	// Test finding request in folder
	req = collection.FindRequest("req3")
	if req == nil {
		t.Error("Expected to find req3")
	} else if req.Name != "Request 3" {
		t.Errorf("Expected 'Request 3', got '%s'", req.Name)
	}

	// Test finding non-existent request
	req = collection.FindRequest("nonexistent")
	if req != nil {
		t.Error("Expected nil for non-existent request")
	}
}

func TestAddRequest(t *testing.T) {
	collection := &CollectionFile{
		Name:     "Test",
		Requests: []CollectionRequest{},
	}

	req := &CollectionRequest{
		Name:   "New Request",
		Method: GET,
		URL:    "http://test.com",
	}

	collection.AddRequest(req)

	if len(collection.Requests) != 1 {
		t.Errorf("Expected 1 request, got %d", len(collection.Requests))
	}
	if collection.Requests[0].ID == "" {
		t.Error("Expected ID to be generated")
	}
}

func TestCreateFolder(t *testing.T) {
	collection := &CollectionFile{
		Name:    "Test",
		Folders: []Folder{},
	}

	collection.CreateFolder("New Folder")

	if len(collection.Folders) != 1 {
		t.Errorf("Expected 1 folder, got %d", len(collection.Folders))
	}
	if collection.Folders[0].Name != "New Folder" {
		t.Errorf("Expected 'New Folder', got '%s'", collection.Folders[0].Name)
	}
}

func TestValidateCollection(t *testing.T) {
	tests := []struct {
		name       string
		collection *CollectionFile
		wantErr    bool
	}{
		{
			name: "Valid collection",
			collection: &CollectionFile{
				Name: "Valid",
				Requests: []CollectionRequest{
					{Name: "Test", Method: GET, URL: "http://test.com"},
				},
			},
			wantErr: false,
		},
		{
			name: "Missing collection name",
			collection: &CollectionFile{
				Name: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid request - missing name",
			collection: &CollectionFile{
				Name: "Test",
				Requests: []CollectionRequest{
					{Method: GET, URL: "http://test.com"},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid request - missing method",
			collection: &CollectionFile{
				Name: "Test",
				Requests: []CollectionRequest{
					{Name: "Test", URL: "http://test.com"},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid request - missing URL",
			collection: &CollectionFile{
				Name: "Test",
				Requests: []CollectionRequest{
					{Name: "Test", Method: GET},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCollection(tt.collection)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCollection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
