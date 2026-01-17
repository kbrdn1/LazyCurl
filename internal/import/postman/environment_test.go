package postman

import (
	"path/filepath"
	"testing"
)

func TestImportEnvironment_Simple(t *testing.T) {
	result, err := ImportEnvironment(filepath.Join("testdata", "simple_environment.json"))
	if err != nil {
		t.Fatalf("ImportEnvironment failed: %v", err)
	}

	if !result.Success() {
		t.Fatal("Expected successful import")
	}

	if result.Environment == nil {
		t.Fatal("Expected environment to be non-nil")
	}

	// Check environment name
	if result.Environment.Name != "Development" {
		t.Errorf("Expected name 'Development', got '%s'", result.Environment.Name)
	}

	// Check variable count
	if result.Summary.VariablesCount != 5 {
		t.Errorf("Expected 5 variables, got %d", result.Summary.VariablesCount)
	}

	// Check specific variables
	if v, ok := result.Environment.Variables["base_url"]; !ok {
		t.Error("Expected 'base_url' variable")
	} else {
		if v.Value != "https://api.example.com" {
			t.Errorf("Expected base_url value 'https://api.example.com', got '%s'", v.Value)
		}
		if v.Secret {
			t.Error("Expected base_url to not be secret")
		}
		if !v.Active {
			t.Error("Expected base_url to be active")
		}
	}
}

func TestImportEnvironment_SecretVariables(t *testing.T) {
	result, err := ImportEnvironment(filepath.Join("testdata", "simple_environment.json"))
	if err != nil {
		t.Fatalf("ImportEnvironment failed: %v", err)
	}

	// Check api_key is marked as secret
	if v, ok := result.Environment.Variables["api_key"]; !ok {
		t.Error("Expected 'api_key' variable")
	} else {
		if !v.Secret {
			t.Error("Expected api_key to be marked as secret")
		}
	}

	// Check access_token is marked as secret
	if v, ok := result.Environment.Variables["access_token"]; !ok {
		t.Error("Expected 'access_token' variable")
	} else {
		if !v.Secret {
			t.Error("Expected access_token to be marked as secret")
		}
	}
}

func TestImportEnvironment_DisabledVariables(t *testing.T) {
	result, err := ImportEnvironment(filepath.Join("testdata", "simple_environment.json"))
	if err != nil {
		t.Fatalf("ImportEnvironment failed: %v", err)
	}

	// Check old_endpoint is disabled
	if v, ok := result.Environment.Variables["old_endpoint"]; !ok {
		t.Error("Expected 'old_endpoint' variable")
	} else {
		if v.Active {
			t.Error("Expected old_endpoint to be disabled (Active=false)")
		}
	}
}

func TestImportEnvironment_Empty(t *testing.T) {
	jsonData := []byte(`{
		"name": "Empty Environment",
		"values": [],
		"_postman_variable_scope": "environment"
	}`)

	result, err := ImportEnvironmentFromBytes(jsonData)
	if err != nil {
		t.Fatalf("ImportEnvironmentFromBytes failed: %v", err)
	}

	if !result.Success() {
		t.Fatal("Expected successful import")
	}

	if result.Summary.VariablesCount != 0 {
		t.Errorf("Expected 0 variables, got %d", result.Summary.VariablesCount)
	}

	if len(result.Environment.Variables) != 0 {
		t.Errorf("Expected empty variables map, got %d entries", len(result.Environment.Variables))
	}
}

func TestImportEnvironment_MissingName(t *testing.T) {
	jsonData := []byte(`{
		"values": [
			{"key": "test", "value": "value", "enabled": true}
		]
	}`)

	_, err := ImportEnvironmentFromBytes(jsonData)
	if err == nil {
		t.Fatal("Expected error for missing name")
	}

	if !contains(err.Error(), "name") {
		t.Errorf("Expected name required error, got: %v", err)
	}
}

func TestImportEnvironment_InvalidJSON(t *testing.T) {
	jsonData := []byte(`{invalid json}`)

	_, err := ImportEnvironmentFromBytes(jsonData)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}

func TestImportEnvironment_FileNotFound(t *testing.T) {
	_, err := ImportEnvironment("nonexistent_environment.json")
	if err == nil {
		t.Fatal("Expected error for missing file")
	}
}

func TestImportEnvironment_FormatSummary(t *testing.T) {
	result, err := ImportEnvironment(filepath.Join("testdata", "simple_environment.json"))
	if err != nil {
		t.Fatalf("ImportEnvironment failed: %v", err)
	}

	summary := result.FormatSummary()
	if !contains(summary, "Development") {
		t.Errorf("Summary should contain environment name, got: %s", summary)
	}
	if !contains(summary, "5 variables") {
		t.Errorf("Summary should contain variable count, got: %s", summary)
	}
}
