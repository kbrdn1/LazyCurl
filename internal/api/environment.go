package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EnvironmentVariable represents a variable with metadata
type EnvironmentVariable struct {
	Value    string `json:"value"`
	Secret   bool   `json:"secret,omitempty"`
	Active   bool   `json:"active"`
}

// EnvironmentFile represents an environment configuration file
type EnvironmentFile struct {
	Name        string                          `json:"name"`
	Description string                          `json:"description,omitempty"`
	Variables   map[string]*EnvironmentVariable `json:"variables"`
	FilePath    string                          `json:"-"` // Internal: path to the file
}

// LoadEnvironment loads an environment from a JSON file
// Supports both new format (with EnvironmentVariable) and legacy format (simple string values)
func LoadEnvironment(path string) (*EnvironmentFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment file: %w", err)
	}

	// First try to parse as new format
	var env EnvironmentFile
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("failed to parse environment JSON: %w", err)
	}

	// Check if variables need migration from legacy format
	// Try to detect if it's legacy format by attempting to unmarshal as simple strings
	var legacyEnv struct {
		Name        string            `json:"name"`
		Description string            `json:"description,omitempty"`
		Variables   map[string]string `json:"variables"`
	}
	if json.Unmarshal(data, &legacyEnv) == nil && len(legacyEnv.Variables) > 0 {
		// Check if first value is a string (legacy) or object (new)
		var rawVars map[string]json.RawMessage
		if err := json.Unmarshal(data, &struct {
			Variables *map[string]json.RawMessage `json:"variables"`
		}{&rawVars}); err == nil {
			for _, v := range rawVars {
				// If it starts with quote, it's a string (legacy format)
				if len(v) > 0 && v[0] == '"' {
					// Migrate to new format
					env.Variables = make(map[string]*EnvironmentVariable)
					for name, value := range legacyEnv.Variables {
						env.Variables[name] = &EnvironmentVariable{
							Value:  value,
							Secret: isSecretKey(name),
							Active: true,
						}
					}
					break
				}
				break
			}
		}
	}

	// Initialize variables map if nil
	if env.Variables == nil {
		env.Variables = make(map[string]*EnvironmentVariable)
	}

	// Set file path
	env.FilePath = path

	return &env, nil
}

// isSecretKey checks if a variable name suggests it should be secret
func isSecretKey(name string) bool {
	nameLower := strings.ToLower(name)
	secretKeywords := []string{"password", "secret", "token", "key", "api_key", "apikey", "auth", "credential"}
	for _, keyword := range secretKeywords {
		if strings.Contains(nameLower, keyword) {
			return true
		}
	}
	return false
}

// SaveEnvironment saves an environment to a JSON file
func SaveEnvironment(env *EnvironmentFile, path string) error {
	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal environment: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write environment file: %w", err)
	}

	return nil
}

// LoadAllEnvironments loads all environments from a directory
func LoadAllEnvironments(dir string) ([]*EnvironmentFile, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []*EnvironmentFile{}, nil
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read environments directory: %w", err)
	}

	var environments []*EnvironmentFile
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, file.Name())
		env, err := LoadEnvironment(path)
		if err != nil {
			// Log error but continue loading other environments
			fmt.Printf("Warning: failed to load environment %s: %v\n", file.Name(), err)
			continue
		}

		environments = append(environments, env)
	}

	return environments, nil
}

// MergeEnvironments merges multiple environments, with later environments overriding earlier ones
func MergeEnvironments(envs ...*EnvironmentFile) *EnvironmentFile {
	if len(envs) == 0 {
		return &EnvironmentFile{
			Name:      "Empty",
			Variables: make(map[string]*EnvironmentVariable),
		}
	}

	merged := &EnvironmentFile{
		Name:      envs[0].Name,
		Variables: make(map[string]*EnvironmentVariable),
	}

	// Merge all variables
	for _, env := range envs {
		if env == nil {
			continue
		}
		for key, value := range env.Variables {
			merged.Variables[key] = &EnvironmentVariable{
				Value:  value.Value,
				Secret: value.Secret,
				Active: value.Active,
			}
		}
	}

	return merged
}

// GetVariable retrieves a variable value from the environment (only if active)
func (e *EnvironmentFile) GetVariable(name string) (string, bool) {
	v, exists := e.Variables[name]
	if !exists || !v.Active {
		return "", false
	}
	return v.Value, true
}

// GetVariableRaw retrieves a variable regardless of active state
func (e *EnvironmentFile) GetVariableRaw(name string) (*EnvironmentVariable, bool) {
	v, exists := e.Variables[name]
	return v, exists
}

// SetVariable sets a variable value in the environment
func (e *EnvironmentFile) SetVariable(name, value string) {
	if e.Variables == nil {
		e.Variables = make(map[string]*EnvironmentVariable)
	}
	if existing, ok := e.Variables[name]; ok {
		existing.Value = value
	} else {
		e.Variables[name] = &EnvironmentVariable{
			Value:  value,
			Secret: isSecretKey(name),
			Active: true,
		}
	}
}

// SetVariableFull sets a variable with all metadata
func (e *EnvironmentFile) SetVariableFull(name string, v *EnvironmentVariable) {
	if e.Variables == nil {
		e.Variables = make(map[string]*EnvironmentVariable)
	}
	e.Variables[name] = v
}

// DeleteVariable removes a variable from the environment
func (e *EnvironmentFile) DeleteVariable(name string) {
	delete(e.Variables, name)
}

// ToggleVariableSecret toggles the secret flag of a variable
func (e *EnvironmentFile) ToggleVariableSecret(name string) bool {
	if v, ok := e.Variables[name]; ok {
		v.Secret = !v.Secret
		return v.Secret
	}
	return false
}

// ToggleVariableActive toggles the active flag of a variable
func (e *EnvironmentFile) ToggleVariableActive(name string) bool {
	if v, ok := e.Variables[name]; ok {
		v.Active = !v.Active
		return v.Active
	}
	return false
}

// ValidateEnvironment validates an environment structure
func ValidateEnvironment(env *EnvironmentFile) error {
	if env.Name == "" {
		return fmt.Errorf("environment name is required")
	}
	if env.Variables == nil {
		return fmt.Errorf("environment variables map is nil")
	}
	return nil
}

// Clone creates a deep copy of the environment
func (e *EnvironmentFile) Clone() *EnvironmentFile {
	clone := &EnvironmentFile{
		Name:        e.Name,
		Description: e.Description,
		FilePath:    e.FilePath,
		Variables:   make(map[string]*EnvironmentVariable),
	}

	for k, v := range e.Variables {
		clone.Variables[k] = &EnvironmentVariable{
			Value:  v.Value,
			Secret: v.Secret,
			Active: v.Active,
		}
	}

	return clone
}

// GetVariableNames returns all variable names in the environment
func (e *EnvironmentFile) GetVariableNames() []string {
	names := make([]string, 0, len(e.Variables))
	for name := range e.Variables {
		names = append(names, name)
	}
	return names
}

// HasVariable checks if a variable exists in the environment
func (e *EnvironmentFile) HasVariable(name string) bool {
	_, exists := e.Variables[name]
	return exists
}
