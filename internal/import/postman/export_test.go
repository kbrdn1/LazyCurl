package postman

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kbrdn1/LazyCurl/internal/api"
)

func TestExportCollectionToBytes(t *testing.T) {
	collection := &api.CollectionFile{
		Name:        "Test Collection",
		Description: "A test collection",
		Requests: []api.CollectionRequest{
			{
				ID:     "req_1",
				Name:   "Get Users",
				Method: "GET",
				URL:    "https://api.example.com/users",
				Headers: []api.KeyValueEntry{
					{Key: "Accept", Value: "application/json", Enabled: true},
				},
			},
			{
				ID:     "req_2",
				Name:   "Create User",
				Method: "POST",
				URL:    "https://api.example.com/users",
				Body: &api.BodyConfig{
					Type:    "json",
					Content: map[string]interface{}{"name": "John"},
				},
			},
		},
	}

	data, err := ExportCollectionToBytes(collection)
	if err != nil {
		t.Fatalf("ExportCollectionToBytes failed: %v", err)
	}

	// Verify it's valid JSON
	var exported Collection
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Exported JSON is invalid: %v", err)
	}

	// Verify collection name
	if exported.Info.Name != "Test Collection" {
		t.Errorf("Expected name 'Test Collection', got '%s'", exported.Info.Name)
	}

	// Verify schema
	if exported.Info.Schema != postmanSchemaV21 {
		t.Errorf("Expected schema '%s', got '%s'", postmanSchemaV21, exported.Info.Schema)
	}

	// Verify UUID is generated
	if exported.Info.PostmanID == "" {
		t.Error("Expected PostmanID to be generated")
	}

	// Verify request count
	if len(exported.Item) != 2 {
		t.Errorf("Expected 2 items, got %d", len(exported.Item))
	}
}

func TestExportCollection_RoundTrip(t *testing.T) {
	// Import a collection
	result, err := ImportCollection(filepath.Join("testdata", "simple_collection.json"))
	if err != nil {
		t.Fatalf("ImportCollection failed: %v", err)
	}

	// Export it
	exportedData, err := ExportCollectionToBytes(result.Collection)
	if err != nil {
		t.Fatalf("ExportCollectionToBytes failed: %v", err)
	}

	// Re-import the exported data
	reImportResult, err := ImportCollectionFromBytes(exportedData)
	if err != nil {
		t.Fatalf("Re-import failed: %v", err)
	}

	// Verify key properties preserved
	if reImportResult.Collection.Name != result.Collection.Name {
		t.Errorf("Name mismatch: original '%s', round-trip '%s'",
			result.Collection.Name, reImportResult.Collection.Name)
	}

	if reImportResult.Summary.RequestsCount != result.Summary.RequestsCount {
		t.Errorf("Request count mismatch: original %d, round-trip %d",
			result.Summary.RequestsCount, reImportResult.Summary.RequestsCount)
	}
}

func TestExportCollection_WithFolders(t *testing.T) {
	collection := &api.CollectionFile{
		Name: "Nested Collection",
		Folders: []api.Folder{
			{
				Name: "Users",
				Requests: []api.CollectionRequest{
					{ID: "req_1", Name: "Get Users", Method: "GET", URL: "/users"},
				},
				Folders: []api.Folder{
					{
						Name: "Admin",
						Requests: []api.CollectionRequest{
							{ID: "req_2", Name: "Get Admins", Method: "GET", URL: "/users/admins"},
						},
					},
				},
			},
		},
		Requests: []api.CollectionRequest{
			{ID: "req_3", Name: "Health", Method: "GET", URL: "/health"},
		},
	}

	data, err := ExportCollectionToBytes(collection)
	if err != nil {
		t.Fatalf("ExportCollectionToBytes failed: %v", err)
	}

	var exported Collection
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Exported JSON is invalid: %v", err)
	}

	// Verify structure: 1 folder + 1 top-level request
	if len(exported.Item) != 2 {
		t.Errorf("Expected 2 top-level items, got %d", len(exported.Item))
	}

	// Find Users folder
	var usersFolder *Item
	for i := range exported.Item {
		if exported.Item[i].Name == "Users" {
			usersFolder = &exported.Item[i]
			break
		}
	}
	if usersFolder == nil {
		t.Fatal("Expected 'Users' folder")
	}

	// Verify nested structure
	if len(usersFolder.Item) != 2 { // 1 request + 1 subfolder
		t.Errorf("Expected 2 items in Users folder, got %d", len(usersFolder.Item))
	}
}

func TestExportCollection_WithAuth(t *testing.T) {
	collection := &api.CollectionFile{
		Name: "Auth Collection",
		Requests: []api.CollectionRequest{
			{
				ID:     "req_1",
				Name:   "Bearer Auth",
				Method: "GET",
				URL:    "/protected",
				Auth: &api.AuthConfig{
					Type:  "bearer",
					Token: "my-token",
				},
			},
			{
				ID:     "req_2",
				Name:   "Basic Auth",
				Method: "GET",
				URL:    "/basic-protected",
				Auth: &api.AuthConfig{
					Type:     "basic",
					Username: "user",
					Password: "pass",
				},
			},
			{
				ID:     "req_3",
				Name:   "API Key Auth",
				Method: "GET",
				URL:    "/api-key-protected",
				Auth: &api.AuthConfig{
					Type:           "api_key",
					APIKeyName:     "X-API-Key",
					APIKeyValue:    "secret-key",
					APIKeyLocation: "header",
				},
			},
		},
	}

	data, err := ExportCollectionToBytes(collection)
	if err != nil {
		t.Fatalf("ExportCollectionToBytes failed: %v", err)
	}

	var exported Collection
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Exported JSON is invalid: %v", err)
	}

	// Verify bearer auth
	bearerReq := exported.Item[0]
	if bearerReq.Request.Auth == nil || bearerReq.Request.Auth.Type != "bearer" {
		t.Error("Expected bearer auth type")
	}
	bearerToken := getAuthValue(bearerReq.Request.Auth.Bearer, "token")
	if bearerToken != "my-token" {
		t.Errorf("Expected bearer token 'my-token', got '%s'", bearerToken)
	}

	// Verify basic auth
	basicReq := exported.Item[1]
	if basicReq.Request.Auth == nil || basicReq.Request.Auth.Type != "basic" {
		t.Error("Expected basic auth type")
	}
	username := getAuthValue(basicReq.Request.Auth.Basic, "username")
	if username != "user" {
		t.Errorf("Expected username 'user', got '%s'", username)
	}

	// Verify API key auth
	apiKeyReq := exported.Item[2]
	if apiKeyReq.Request.Auth == nil || apiKeyReq.Request.Auth.Type != "apikey" {
		t.Error("Expected apikey auth type")
	}
}

func TestExportCollection_ToFile(t *testing.T) {
	collection := &api.CollectionFile{
		Name: "File Export Test",
		Requests: []api.CollectionRequest{
			{ID: "req_1", Name: "Test", Method: "GET", URL: "/test"},
		},
	}

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "exported_collection.json")

	err := ExportCollection(collection, filePath)
	if err != nil {
		t.Fatalf("ExportCollection failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Expected exported file to exist")
	}

	// Verify file is valid Postman collection
	fileType, err := DetectFileType(filePath)
	if err != nil {
		t.Fatalf("DetectFileType failed: %v", err)
	}
	if fileType != FileTypeCollection {
		t.Errorf("Expected exported file to be detected as Collection, got %s", fileType)
	}
}

func TestExportEnvironmentToBytes(t *testing.T) {
	env := &api.EnvironmentFile{
		Name: "Test Environment",
		Variables: map[string]*api.EnvironmentVariable{
			"base_url": {Value: "https://api.example.com", Secret: false, Active: true},
			"api_key":  {Value: "secret-key", Secret: true, Active: true},
			"disabled": {Value: "old-value", Secret: false, Active: false},
		},
	}

	data, err := ExportEnvironmentToBytes(env)
	if err != nil {
		t.Fatalf("ExportEnvironmentToBytes failed: %v", err)
	}

	// Verify it's valid JSON
	var exported Environment
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Exported JSON is invalid: %v", err)
	}

	// Verify environment name
	if exported.Name != "Test Environment" {
		t.Errorf("Expected name 'Test Environment', got '%s'", exported.Name)
	}

	// Verify variable count
	if len(exported.Values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(exported.Values))
	}

	// Verify secret mapping
	for _, v := range exported.Values {
		if v.Key == "api_key" && v.Type != "secret" {
			t.Error("Expected api_key to have type 'secret'")
		}
		if v.Key == "base_url" && v.Type != "default" {
			t.Error("Expected base_url to have type 'default'")
		}
		if v.Key == "disabled" && v.Enabled {
			t.Error("Expected disabled variable to have Enabled=false")
		}
	}
}

func TestExportEnvironment_RoundTrip(t *testing.T) {
	// Import an environment
	result, err := ImportEnvironment(filepath.Join("testdata", "simple_environment.json"))
	if err != nil {
		t.Fatalf("ImportEnvironment failed: %v", err)
	}

	// Export it
	exportedData, err := ExportEnvironmentToBytes(result.Environment)
	if err != nil {
		t.Fatalf("ExportEnvironmentToBytes failed: %v", err)
	}

	// Re-import the exported data
	reImportResult, err := ImportEnvironmentFromBytes(exportedData)
	if err != nil {
		t.Fatalf("Re-import failed: %v", err)
	}

	// Verify key properties preserved
	if reImportResult.Environment.Name != result.Environment.Name {
		t.Errorf("Name mismatch: original '%s', round-trip '%s'",
			result.Environment.Name, reImportResult.Environment.Name)
	}

	if reImportResult.Summary.VariablesCount != result.Summary.VariablesCount {
		t.Errorf("Variable count mismatch: original %d, round-trip %d",
			result.Summary.VariablesCount, reImportResult.Summary.VariablesCount)
	}

	// Verify specific variable preserved
	originalVar := result.Environment.Variables["api_key"]
	reImportedVar := reImportResult.Environment.Variables["api_key"]
	if originalVar.Value != reImportedVar.Value {
		t.Errorf("api_key value mismatch: original '%s', round-trip '%s'",
			originalVar.Value, reImportedVar.Value)
	}
	if originalVar.Secret != reImportedVar.Secret {
		t.Errorf("api_key secret mismatch: original %v, round-trip %v",
			originalVar.Secret, reImportedVar.Secret)
	}
}

func TestExportEnvironment_ToFile(t *testing.T) {
	env := &api.EnvironmentFile{
		Name: "File Export Test",
		Variables: map[string]*api.EnvironmentVariable{
			"test": {Value: "value", Secret: false, Active: true},
		},
	}

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "exported_environment.json")

	err := ExportEnvironment(env, filePath)
	if err != nil {
		t.Fatalf("ExportEnvironment failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("Expected exported file to exist")
	}

	// Verify file is valid Postman environment
	fileType, err := DetectFileType(filePath)
	if err != nil {
		t.Fatalf("DetectFileType failed: %v", err)
	}
	if fileType != FileTypeEnvironment {
		t.Errorf("Expected exported file to be detected as Environment, got %s", fileType)
	}
}

func TestExportCollection_BodyTypes(t *testing.T) {
	collection := &api.CollectionFile{
		Name: "Body Types",
		Requests: []api.CollectionRequest{
			{
				ID:     "req_1",
				Name:   "JSON Body",
				Method: "POST",
				URL:    "/json",
				Body:   &api.BodyConfig{Type: "json", Content: `{"key": "value"}`},
			},
			{
				ID:     "req_2",
				Name:   "Raw Body",
				Method: "POST",
				URL:    "/raw",
				Body:   &api.BodyConfig{Type: "raw", Content: "plain text"},
			},
			{
				ID:     "req_3",
				Name:   "Form Data",
				Method: "POST",
				URL:    "/form",
				Body: &api.BodyConfig{
					Type: "form-data",
					Content: []map[string]interface{}{
						{"key": "field1", "value": "value1", "type": "text"},
					},
				},
			},
		},
	}

	data, err := ExportCollectionToBytes(collection)
	if err != nil {
		t.Fatalf("ExportCollectionToBytes failed: %v", err)
	}

	var exported Collection
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Exported JSON is invalid: %v", err)
	}

	// Verify JSON body
	jsonReq := exported.Item[0]
	if jsonReq.Request.Body == nil || jsonReq.Request.Body.Mode != "raw" {
		t.Error("Expected raw mode for JSON body")
	}
	if jsonReq.Request.Body.Options == nil || jsonReq.Request.Body.Options.Raw.Language != "json" {
		t.Error("Expected json language for JSON body")
	}

	// Verify raw body
	rawReq := exported.Item[1]
	if rawReq.Request.Body == nil || rawReq.Request.Body.Mode != "raw" {
		t.Error("Expected raw mode for raw body")
	}
	if rawReq.Request.Body.Options == nil || rawReq.Request.Body.Options.Raw.Language != "text" {
		t.Error("Expected text language for raw body")
	}

	// Verify form data
	formReq := exported.Item[2]
	if formReq.Request.Body == nil || formReq.Request.Body.Mode != "formdata" {
		t.Error("Expected formdata mode for form body")
	}
}

func TestExportCollection_WithScripts(t *testing.T) {
	collection := &api.CollectionFile{
		Name: "Scripts Collection",
		Requests: []api.CollectionRequest{
			{
				ID:     "req_1",
				Name:   "With Scripts",
				Method: "GET",
				URL:    "/test",
				Scripts: &api.ScriptConfig{
					PreRequest:  "console.log('pre');",
					PostRequest: "console.log('post');",
				},
			},
		},
	}

	data, err := ExportCollectionToBytes(collection)
	if err != nil {
		t.Fatalf("ExportCollectionToBytes failed: %v", err)
	}

	var exported Collection
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Exported JSON is invalid: %v", err)
	}

	// Verify scripts exported as events
	item := exported.Item[0]
	if len(item.Event) != 2 {
		t.Errorf("Expected 2 events, got %d", len(item.Event))
	}

	var hasPrerequest, hasTest bool
	for _, event := range item.Event {
		if event.Listen == "prerequest" {
			hasPrerequest = true
		}
		if event.Listen == "test" {
			hasTest = true
		}
	}

	if !hasPrerequest {
		t.Error("Expected prerequest event")
	}
	if !hasTest {
		t.Error("Expected test event")
	}
}
