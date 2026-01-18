package api

import (
	"sync"
)

// ScriptGlobals provides global variable storage accessible across all scripts
// Unlike ScriptEnvironment which is scoped to a single request execution,
// ScriptGlobals persists across multiple requests in a session
type ScriptGlobals struct {
	variables map[string]interface{}
	mu        sync.RWMutex
}

// NewScriptGlobals creates a new global variable store
func NewScriptGlobals() *ScriptGlobals {
	return &ScriptGlobals{
		variables: make(map[string]interface{}),
	}
}

// Get retrieves a global variable value
func (g *ScriptGlobals) Get(name string) interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if val, ok := g.variables[name]; ok {
		return val
	}
	return nil
}

// Set sets a global variable value
func (g *ScriptGlobals) Set(name string, value interface{}) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.variables[name] = value
}

// Has checks if a global variable exists
func (g *ScriptGlobals) Has(name string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, ok := g.variables[name]
	return ok
}

// Unset removes a global variable
func (g *ScriptGlobals) Unset(name string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.variables, name)
}

// Clear removes all global variables
func (g *ScriptGlobals) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.variables = make(map[string]interface{})
}

// All returns a copy of all global variables
func (g *ScriptGlobals) All() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make(map[string]interface{}, len(g.variables))
	for k, v := range g.variables {
		result[k] = v
	}
	return result
}
