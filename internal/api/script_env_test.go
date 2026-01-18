package api

import (
	"testing"
)

func TestNewScriptEnvironment_NilEnv(t *testing.T) {
	se := NewScriptEnvironment(nil)
	if se == nil {
		t.Fatal("NewScriptEnvironment(nil) returned nil")
	}

	// Should not panic with nil env
	if se.Get("test") != "" {
		t.Error("Get on nil env should return empty string")
	}
	if se.Has("test") {
		t.Error("Has on nil env should return false")
	}
}

func TestNewScriptEnvironment_WithEnv(t *testing.T) {
	env := &Environment{
		Name: "test",
		Variables: map[string]string{
			"api_key":  "secret123",
			"base_url": "https://api.example.com",
		},
	}

	se := NewScriptEnvironment(env)

	if se.Get("api_key") != "secret123" {
		t.Errorf("Get('api_key') = %q, want %q", se.Get("api_key"), "secret123")
	}
	if se.Get("base_url") != "https://api.example.com" {
		t.Errorf("Get('base_url') = %q, want %q", se.Get("base_url"), "https://api.example.com")
	}
}

func TestScriptEnvironment_Get(t *testing.T) {
	env := &Environment{
		Name: "test",
		Variables: map[string]string{
			"existing": "value",
		},
	}

	se := NewScriptEnvironment(env)

	t.Run("existing variable", func(t *testing.T) {
		if se.Get("existing") != "value" {
			t.Errorf("Get('existing') = %q, want %q", se.Get("existing"), "value")
		}
	})

	t.Run("non-existing variable", func(t *testing.T) {
		if se.Get("nonexistent") != "" {
			t.Error("Get should return empty string for non-existent variable")
		}
	})
}

func TestScriptEnvironment_Set(t *testing.T) {
	env := &Environment{
		Name: "test",
		Variables: map[string]string{
			"existing": "old_value",
		},
	}

	se := NewScriptEnvironment(env)

	t.Run("set new variable", func(t *testing.T) {
		se.Set("new_var", "new_value")

		// Should be accessible via Get immediately
		if se.Get("new_var") != "new_value" {
			t.Errorf("Get('new_var') = %q, want %q", se.Get("new_var"), "new_value")
		}

		// Should be tracked in changes
		changes := se.GetChanges()
		found := false
		for _, c := range changes {
			if c.Name == "new_var" && c.Value == "new_value" && c.Type == EnvChangeSet {
				found = true
				break
			}
		}
		if !found {
			t.Error("Set should track the change")
		}
	})

	t.Run("update existing variable", func(t *testing.T) {
		se.Set("existing", "updated_value")

		if se.Get("existing") != "updated_value" {
			t.Errorf("Get('existing') = %q, want %q", se.Get("existing"), "updated_value")
		}

		// Should track previous value
		changes := se.GetChanges()
		found := false
		for _, c := range changes {
			if c.Name == "existing" && c.Previous == "old_value" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Set should track previous value")
		}
	})
}

func TestScriptEnvironment_Unset(t *testing.T) {
	env := &Environment{
		Name: "test",
		Variables: map[string]string{
			"to_remove": "value",
		},
	}

	se := NewScriptEnvironment(env)

	se.Unset("to_remove")

	// Should return empty after unset
	if se.Get("to_remove") != "" {
		t.Error("Get should return empty after Unset")
	}

	// Should not exist anymore
	if se.Has("to_remove") {
		t.Error("Has should return false after Unset")
	}

	// Should be tracked in changes
	changes := se.GetChanges()
	found := false
	for _, c := range changes {
		if c.Name == "to_remove" && c.Type == EnvChangeUnset && c.Previous == "value" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Unset should track the change with previous value")
	}
}

func TestScriptEnvironment_Has(t *testing.T) {
	env := &Environment{
		Name: "test",
		Variables: map[string]string{
			"existing": "value",
		},
	}

	se := NewScriptEnvironment(env)

	t.Run("existing variable", func(t *testing.T) {
		if !se.Has("existing") {
			t.Error("Has should return true for existing variable")
		}
	})

	t.Run("non-existing variable", func(t *testing.T) {
		if se.Has("nonexistent") {
			t.Error("Has should return false for non-existent variable")
		}
	})

	t.Run("after Set", func(t *testing.T) {
		se.Set("new_var", "value")
		if !se.Has("new_var") {
			t.Error("Has should return true after Set")
		}
	})

	t.Run("after Unset", func(t *testing.T) {
		se.Unset("existing")
		if se.Has("existing") {
			t.Error("Has should return false after Unset")
		}
	})
}

func TestScriptEnvironment_GetChanges(t *testing.T) {
	env := &Environment{
		Name:      "test",
		Variables: map[string]string{},
	}

	se := NewScriptEnvironment(env)

	// Initially no changes
	if len(se.GetChanges()) != 0 {
		t.Error("GetChanges should return empty list initially")
	}

	// Add some changes
	se.Set("var1", "value1")
	se.Set("var2", "value2")
	se.Unset("var1")

	changes := se.GetChanges()
	if len(changes) != 3 {
		t.Errorf("Expected 3 changes, got %d", len(changes))
	}
}

func TestScriptEnvironment_GetChanges_ReturnsCopy(t *testing.T) {
	env := &Environment{
		Name:      "test",
		Variables: map[string]string{},
	}

	se := NewScriptEnvironment(env)
	se.Set("var1", "value1")

	changes1 := se.GetChanges()
	changes1[0].Value = "modified"

	// Original should not be modified
	changes2 := se.GetChanges()
	if changes2[0].Value != "value1" {
		t.Error("GetChanges should return a copy")
	}
}

func TestScriptEnvironment_ApplyChanges(t *testing.T) {
	env := &Environment{
		Name: "test",
		Variables: map[string]string{
			"existing": "old_value",
		},
	}

	se := NewScriptEnvironment(env)

	// Make changes
	se.Set("new_var", "new_value")
	se.Set("existing", "updated_value")
	se.Unset("existing")

	// Apply changes
	err := se.ApplyChanges()
	if err != nil {
		t.Errorf("ApplyChanges failed: %v", err)
	}

	// Check env was updated
	if env.Variables["new_var"] != "new_value" {
		t.Errorf("new_var = %q, want %q", env.Variables["new_var"], "new_value")
	}

	// existing should be removed (last operation was Unset)
	if _, exists := env.Variables["existing"]; exists {
		t.Error("existing should have been removed")
	}
}

func TestScriptEnvironment_ApplyChanges_NilEnv(t *testing.T) {
	se := NewScriptEnvironment(nil)
	se.Set("var", "value")

	// Should not panic
	err := se.ApplyChanges()
	if err != nil {
		t.Errorf("ApplyChanges with nil env returned error: %v", err)
	}
}

func TestScriptEnvironment_ApplyChanges_NilVariables(t *testing.T) {
	env := &Environment{
		Name:      "test",
		Variables: nil, // nil Variables map
	}

	se := NewScriptEnvironment(env)
	se.Set("new_var", "value")

	err := se.ApplyChanges()
	if err != nil {
		t.Errorf("ApplyChanges failed: %v", err)
	}

	// Variables map should be initialized
	if env.Variables == nil {
		t.Error("ApplyChanges should initialize nil Variables map")
	}
	if env.Variables["new_var"] != "value" {
		t.Errorf("new_var = %q, want %q", env.Variables["new_var"], "value")
	}
}

func TestScriptEnvironment_MultipleSetsSameKey(t *testing.T) {
	env := &Environment{
		Name:      "test",
		Variables: map[string]string{},
	}

	se := NewScriptEnvironment(env)

	// Set same variable multiple times
	se.Set("var", "value1")
	se.Set("var", "value2")
	se.Set("var", "value3")

	// Get should return the latest value
	if se.Get("var") != "value3" {
		t.Errorf("Get('var') = %q, want %q", se.Get("var"), "value3")
	}

	// Changes should have all three operations recorded
	changes := se.GetChanges()
	if len(changes) != 3 {
		t.Errorf("Expected 3 changes, got %d", len(changes))
	}

	// After apply, should have final value
	_ = se.ApplyChanges()
	if env.Variables["var"] != "value3" {
		t.Errorf("env.Variables['var'] = %q, want %q", env.Variables["var"], "value3")
	}
}

func TestScriptEnvironment_SetThenUnset(t *testing.T) {
	env := &Environment{
		Name:      "test",
		Variables: map[string]string{},
	}

	se := NewScriptEnvironment(env)

	// Set then immediately unset
	se.Set("temp_var", "temp_value")
	se.Unset("temp_var")

	// Should not exist
	if se.Has("temp_var") {
		t.Error("temp_var should not exist after Set+Unset")
	}
	if se.Get("temp_var") != "" {
		t.Error("Get should return empty for Set+Unset variable")
	}

	// After apply, should not be in env
	_ = se.ApplyChanges()
	if _, exists := env.Variables["temp_var"]; exists {
		t.Error("temp_var should not exist in env after Apply")
	}
}
