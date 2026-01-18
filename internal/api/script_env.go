package api

import (
	"sync"
)

// EnvChangeType represents the type of environment change
type EnvChangeType string

const (
	EnvChangeSet   EnvChangeType = "set"
	EnvChangeUnset EnvChangeType = "unset"
)

// EnvChange represents a single environment variable modification
type EnvChange struct {
	Type     EnvChangeType `json:"type"`
	Name     string        `json:"name"`
	Value    string        `json:"value,omitempty"` // Only for "set"
	Previous string        `json:"previous,omitempty"`
}

// ScriptEnvironment wraps Environment for script access with change tracking
type ScriptEnvironment struct {
	env     *Environment
	changes []EnvChange
	mu      sync.Mutex
}

// NewScriptEnvironment wraps an environment for script use
func NewScriptEnvironment(env *Environment) *ScriptEnvironment {
	return &ScriptEnvironment{
		env:     env,
		changes: make([]EnvChange, 0),
	}
}

// NewScriptEnvironmentFromFile creates a ScriptEnvironment from an EnvironmentFile
func NewScriptEnvironmentFromFile(envFile *EnvironmentFile) *ScriptEnvironment {
	return NewScriptEnvironment(EnvironmentFromFile(envFile))
}

// EnvironmentFromFile converts an EnvironmentFile to a simple Environment
func EnvironmentFromFile(envFile *EnvironmentFile) *Environment {
	if envFile == nil {
		return nil
	}

	env := &Environment{
		Name:      envFile.Name,
		Variables: make(map[string]string),
	}

	for key, varDef := range envFile.Variables {
		if varDef != nil && varDef.Active {
			env.Variables[key] = varDef.Value
		}
	}

	return env
}

// Get retrieves an environment variable value
func (e *ScriptEnvironment) Get(name string) string {
	// Check if we have a pending change for this variable
	e.mu.Lock()
	defer e.mu.Unlock()
	for i := len(e.changes) - 1; i >= 0; i-- {
		if e.changes[i].Name == name {
			if e.changes[i].Type == EnvChangeUnset {
				return ""
			}
			return e.changes[i].Value
		}
	}
	// Fall back to the original environment
	if e.env != nil && e.env.Variables != nil {
		if val, ok := e.env.Variables[name]; ok {
			return val
		}
	}
	return ""
}

// Set sets an environment variable value and tracks the change
func (e *ScriptEnvironment) Set(name, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get the previous value
	previous := ""
	if e.env != nil && e.env.Variables != nil {
		if val, ok := e.env.Variables[name]; ok {
			previous = val
		}
	}

	// Track the change
	e.changes = append(e.changes, EnvChange{
		Type:     EnvChangeSet,
		Name:     name,
		Value:    value,
		Previous: previous,
	})
}

// Unset removes an environment variable and tracks the change
func (e *ScriptEnvironment) Unset(name string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Get the previous value
	previous := ""
	if e.env != nil && e.env.Variables != nil {
		if val, ok := e.env.Variables[name]; ok {
			previous = val
		}
	}

	// Track the change
	e.changes = append(e.changes, EnvChange{
		Type:     EnvChangeUnset,
		Name:     name,
		Previous: previous,
	})
}

// Has checks if an environment variable exists
func (e *ScriptEnvironment) Has(name string) bool {
	// Check pending changes first
	e.mu.Lock()
	defer e.mu.Unlock()
	for i := len(e.changes) - 1; i >= 0; i-- {
		if e.changes[i].Name == name {
			return e.changes[i].Type == EnvChangeSet
		}
	}
	// Check original environment
	if e.env != nil && e.env.Variables != nil {
		_, ok := e.env.Variables[name]
		return ok
	}
	return false
}

// GetChanges returns all environment variable modifications
func (e *ScriptEnvironment) GetChanges() []EnvChange {
	e.mu.Lock()
	defer e.mu.Unlock()
	// Return a copy to avoid concurrent modification
	result := make([]EnvChange, len(e.changes))
	copy(result, e.changes)
	return result
}

// ApplyChanges persists the changes to the underlying environment
func (e *ScriptEnvironment) ApplyChanges() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.env == nil {
		return nil
	}

	// Initialize Variables map if nil
	if e.env.Variables == nil {
		e.env.Variables = make(map[string]string)
	}

	for _, change := range e.changes {
		switch change.Type {
		case EnvChangeSet:
			e.env.Variables[change.Name] = change.Value
		case EnvChangeUnset:
			delete(e.env.Variables, change.Name)
		}
	}

	return nil
}
