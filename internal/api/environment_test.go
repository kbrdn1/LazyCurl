package api

import (
	"os"
	"path/filepath"
	"testing"
)

// Helper to create environment variable
func newVar(value string, secret, active bool) *EnvironmentVariable {
	return &EnvironmentVariable{
		Value:  value,
		Secret: secret,
		Active: active,
	}
}

func TestLoadEnvironment(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with new format
	newFormatJSON := `{
		"name": "Development",
		"description": "Local development environment",
		"variables": {
			"base_url": {"value": "http://localhost:3000", "secret": false, "active": true},
			"api_key": {"value": "dev_key_123", "secret": true, "active": true}
		}
	}`

	newFormatPath := filepath.Join(tmpDir, "dev_new.json")
	if err := os.WriteFile(newFormatPath, []byte(newFormatJSON), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	env, err := LoadEnvironment(newFormatPath)
	if err != nil {
		t.Errorf("LoadEnvironment() error = %v", err)
	}
	if env.Name != "Development" {
		t.Errorf("Expected name 'Development', got '%s'", env.Name)
	}
	if env.Variables["base_url"].Value != "http://localhost:3000" {
		t.Errorf("Expected base_url 'http://localhost:3000', got '%s'", env.Variables["base_url"].Value)
	}

	// Test legacy format migration
	legacyJSON := `{
		"name": "Legacy",
		"variables": {
			"base_url": "http://localhost:3000",
			"api_key": "legacy_key"
		}
	}`

	legacyPath := filepath.Join(tmpDir, "legacy.json")
	if err := os.WriteFile(legacyPath, []byte(legacyJSON), 0644); err != nil {
		t.Fatalf("Failed to create legacy test file: %v", err)
	}

	legacyEnv, err := LoadEnvironment(legacyPath)
	if err != nil {
		t.Errorf("LoadEnvironment() for legacy format error = %v", err)
	}
	if legacyEnv.Variables["base_url"].Value != "http://localhost:3000" {
		t.Errorf("Legacy migration failed, expected 'http://localhost:3000', got '%s'", legacyEnv.Variables["base_url"].Value)
	}
	// api_key should be detected as secret
	if !legacyEnv.Variables["api_key"].Secret {
		t.Error("Expected api_key to be marked as secret after migration")
	}
}

func TestSaveEnvironment(t *testing.T) {
	tmpDir := t.TempDir()

	env := &EnvironmentFile{
		Name:        "Test",
		Description: "Test environment",
		Variables: map[string]*EnvironmentVariable{
			"var1": newVar("value1", false, true),
			"var2": newVar("value2", true, true),
		},
	}

	path := filepath.Join(tmpDir, "test.json")
	err := SaveEnvironment(env, path)
	if err != nil {
		t.Errorf("SaveEnvironment() error = %v", err)
	}

	// Load and verify
	loaded, err := LoadEnvironment(path)
	if err != nil {
		t.Errorf("Failed to load saved environment: %v", err)
	}
	if loaded.Name != env.Name {
		t.Errorf("Expected name '%s', got '%s'", env.Name, loaded.Name)
	}
	if loaded.Variables["var1"].Value != "value1" {
		t.Error("Variables not saved correctly")
	}
	if !loaded.Variables["var2"].Secret {
		t.Error("Secret flag not saved correctly")
	}
}

func TestLoadAllEnvironments(t *testing.T) {
	tmpDir := t.TempDir()
	envsDir := filepath.Join(tmpDir, "envs")
	os.MkdirAll(envsDir, 0755)

	// Create multiple environment files
	envs := []struct {
		name     string
		filename string
	}{
		{"Development", "dev.json"},
		{"Staging", "staging.json"},
		{"Production", "prod.json"},
	}

	for _, env := range envs {
		e := &EnvironmentFile{
			Name: env.name,
			Variables: map[string]*EnvironmentVariable{
				"test": newVar("value", false, true),
			},
		}
		if err := SaveEnvironment(e, filepath.Join(envsDir, env.filename)); err != nil {
			t.Fatalf("Failed to save environment: %v", err)
		}
	}

	loaded, err := LoadAllEnvironments(envsDir)
	if err != nil {
		t.Errorf("LoadAllEnvironments() error = %v", err)
	}

	if len(loaded) != 3 {
		t.Errorf("Expected 3 environments, got %d", len(loaded))
	}
}

func TestMergeEnvironments(t *testing.T) {
	env1 := &EnvironmentFile{
		Name: "Base",
		Variables: map[string]*EnvironmentVariable{
			"var1": newVar("value1", false, true),
			"var2": newVar("value2", false, true),
		},
	}

	env2 := &EnvironmentFile{
		Name: "Override",
		Variables: map[string]*EnvironmentVariable{
			"var2": newVar("overridden", false, true),
			"var3": newVar("value3", false, true),
		},
	}

	merged := MergeEnvironments(env1, env2)

	if merged.Variables["var1"].Value != "value1" {
		t.Error("var1 should be from env1")
	}
	if merged.Variables["var2"].Value != "overridden" {
		t.Error("var2 should be overridden by env2")
	}
	if merged.Variables["var3"].Value != "value3" {
		t.Error("var3 should be from env2")
	}
}

func TestGetVariable(t *testing.T) {
	env := &EnvironmentFile{
		Name: "Test",
		Variables: map[string]*EnvironmentVariable{
			"existing": newVar("value", false, true),
			"inactive": newVar("inactive_value", false, false),
		},
	}

	value, exists := env.GetVariable("existing")
	if !exists {
		t.Error("Expected variable to exist")
	}
	if value != "value" {
		t.Errorf("Expected 'value', got '%s'", value)
	}

	_, exists = env.GetVariable("nonexistent")
	if exists {
		t.Error("Expected variable to not exist")
	}

	// Inactive variables should not be returned
	_, exists = env.GetVariable("inactive")
	if exists {
		t.Error("Expected inactive variable to not be returned by GetVariable")
	}
}

func TestSetVariable(t *testing.T) {
	env := &EnvironmentFile{
		Name:      "Test",
		Variables: make(map[string]*EnvironmentVariable),
	}

	env.SetVariable("newvar", "newvalue")

	if env.Variables["newvar"].Value != "newvalue" {
		t.Error("Variable not set correctly")
	}
	if !env.Variables["newvar"].Active {
		t.Error("New variable should be active by default")
	}
}

func TestDeleteVariable(t *testing.T) {
	env := &EnvironmentFile{
		Name: "Test",
		Variables: map[string]*EnvironmentVariable{
			"var1": newVar("value1", false, true),
		},
	}

	env.DeleteVariable("var1")

	if _, exists := env.Variables["var1"]; exists {
		t.Error("Variable should be deleted")
	}
}

func TestToggleVariableSecret(t *testing.T) {
	env := &EnvironmentFile{
		Name: "Test",
		Variables: map[string]*EnvironmentVariable{
			"var1": newVar("value1", false, true),
		},
	}

	result := env.ToggleVariableSecret("var1")
	if !result {
		t.Error("Expected toggle to return true")
	}
	if !env.Variables["var1"].Secret {
		t.Error("Secret should be true after toggle")
	}

	result = env.ToggleVariableSecret("var1")
	if result {
		t.Error("Expected toggle to return false")
	}
	if env.Variables["var1"].Secret {
		t.Error("Secret should be false after second toggle")
	}
}

func TestToggleVariableActive(t *testing.T) {
	env := &EnvironmentFile{
		Name: "Test",
		Variables: map[string]*EnvironmentVariable{
			"var1": newVar("value1", false, true),
		},
	}

	result := env.ToggleVariableActive("var1")
	if result {
		t.Error("Expected toggle to return false")
	}
	if env.Variables["var1"].Active {
		t.Error("Active should be false after toggle")
	}

	result = env.ToggleVariableActive("var1")
	if !result {
		t.Error("Expected toggle to return true")
	}
	if !env.Variables["var1"].Active {
		t.Error("Active should be true after second toggle")
	}
}

func TestValidateEnvironment(t *testing.T) {
	tests := []struct {
		name    string
		env     *EnvironmentFile
		wantErr bool
	}{
		{
			name: "Valid environment",
			env: &EnvironmentFile{
				Name:      "Valid",
				Variables: map[string]*EnvironmentVariable{},
			},
			wantErr: false,
		},
		{
			name: "Missing name",
			env: &EnvironmentFile{
				Variables: map[string]*EnvironmentVariable{},
			},
			wantErr: true,
		},
		{
			name: "Nil variables",
			env: &EnvironmentFile{
				Name: "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvironment(tt.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnvironment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClone(t *testing.T) {
	original := &EnvironmentFile{
		Name:        "Original",
		Description: "Original description",
		Variables: map[string]*EnvironmentVariable{
			"var1": newVar("value1", true, true),
		},
	}

	clone := original.Clone()

	// Verify clone is equal
	if clone.Name != original.Name {
		t.Error("Clone name doesn't match")
	}
	if clone.Variables["var1"].Value != "value1" {
		t.Error("Clone variables don't match")
	}
	if !clone.Variables["var1"].Secret {
		t.Error("Clone should preserve secret flag")
	}

	// Verify it's a deep copy
	clone.Variables["var1"].Value = "modified"
	if original.Variables["var1"].Value == "modified" {
		t.Error("Clone modified original")
	}
}

func TestGetVariableNames(t *testing.T) {
	env := &EnvironmentFile{
		Name: "Test",
		Variables: map[string]*EnvironmentVariable{
			"var1": newVar("value1", false, true),
			"var2": newVar("value2", false, true),
			"var3": newVar("value3", false, true),
		},
	}

	names := env.GetVariableNames()
	if len(names) != 3 {
		t.Errorf("Expected 3 variable names, got %d", len(names))
	}
}

func TestHasVariable(t *testing.T) {
	env := &EnvironmentFile{
		Name: "Test",
		Variables: map[string]*EnvironmentVariable{
			"existing": newVar("value", false, true),
		},
	}

	if !env.HasVariable("existing") {
		t.Error("Expected HasVariable to return true")
	}
	if env.HasVariable("nonexistent") {
		t.Error("Expected HasVariable to return false")
	}
}
