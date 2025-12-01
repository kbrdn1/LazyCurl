package api

import (
	"strings"
	"testing"
)

func TestReplaceVariables(t *testing.T) {
	env := &EnvironmentFile{
		Name: "Test",
		Variables: map[string]*EnvironmentVariable{
			"base_url": {Value: "https://api.example.com", Active: true},
			"api_key":  {Value: "secret123", Active: true},
			"version":  {Value: "v1", Active: true},
		},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple variable",
			input:    "{{base_url}}/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "Multiple variables",
			input:    "{{base_url}}/{{version}}/users",
			expected: "https://api.example.com/v1/users",
		},
		{
			name:     "Variable in header value",
			input:    "Bearer {{api_key}}",
			expected: "Bearer secret123",
		},
		{
			name:     "No variables",
			input:    "https://api.example.com/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "Undefined variable",
			input:    "{{undefined}}",
			expected: "{{undefined}}",
		},
		{
			name:     "Mixed defined and undefined",
			input:    "{{base_url}}/{{undefined}}",
			expected: "https://api.example.com/{{undefined}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceVariables(tt.input, env)
			if result != tt.expected {
				t.Errorf("ReplaceVariables() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestReplaceVariablesInactiveVariable(t *testing.T) {
	env := &EnvironmentFile{
		Name: "Test",
		Variables: map[string]*EnvironmentVariable{
			"active_var":   {Value: "active_value", Active: true},
			"inactive_var": {Value: "inactive_value", Active: false},
		},
	}

	// Active variable should be replaced
	result := ReplaceVariables("{{active_var}}", env)
	if result != "active_value" {
		t.Errorf("Active variable not replaced: got %v", result)
	}

	// Inactive variable should NOT be replaced
	result = ReplaceVariables("{{inactive_var}}", env)
	if result != "{{inactive_var}}" {
		t.Errorf("Inactive variable should not be replaced: got %v", result)
	}
}

func TestSystemVariables(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		shouldContain string
	}{
		{
			name:          "Timestamp",
			input:         "{{$timestamp}}",
			shouldContain: "", // Just check it's a number
		},
		{
			name:          "Datetime",
			input:         "{{$datetime}}",
			shouldContain: "T", // ISO format contains T
		},
		{
			name:          "Date",
			input:         "{{$date}}",
			shouldContain: "-", // Date format contains dashes
		},
		{
			name:          "UUID",
			input:         "{{$uuid}}",
			shouldContain: "-", // UUID contains dashes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceVariables(tt.input, nil)

			// Should not contain the original placeholder
			if result == tt.input {
				t.Errorf("System variable not replaced: %v", tt.input)
			}

			// Should contain expected pattern if specified
			if tt.shouldContain != "" && !strings.Contains(result, tt.shouldContain) {
				t.Errorf("Result '%v' does not contain expected pattern '%v'", result, tt.shouldContain)
			}
		})
	}
}

func TestReplaceVariablesInRequest(t *testing.T) {
	env := &EnvironmentFile{
		Name: "Test",
		Variables: map[string]*EnvironmentVariable{
			"base_url": {Value: "https://api.example.com", Active: true},
			"token":    {Value: "secret123", Active: true},
			"user_id":  {Value: "42", Active: true},
		},
	}

	req := &Request{
		Method: POST,
		URL:    "{{base_url}}/users/{{user_id}}",
		Headers: map[string]string{
			"Authorization": "Bearer {{token}}",
			"Content-Type":  "application/json",
		},
		Body: map[string]interface{}{
			"name": "John",
			"age":  30,
		},
	}

	replaced := ReplaceVariablesInRequest(req, env)

	if replaced.URL != "https://api.example.com/users/42" {
		t.Errorf("URL not replaced correctly: %v", replaced.URL)
	}

	if replaced.Headers["Authorization"] != "Bearer secret123" {
		t.Errorf("Header not replaced correctly: %v", replaced.Headers["Authorization"])
	}

	// Original should not be modified
	if req.URL == replaced.URL {
		t.Error("Original request was modified")
	}
}

func TestReplaceVariablesInBody(t *testing.T) {
	env := &EnvironmentFile{
		Name: "Test",
		Variables: map[string]*EnvironmentVariable{
			"username": {Value: "john_doe", Active: true},
			"email":    {Value: "john@example.com", Active: true},
		},
	}

	tests := []struct {
		name     string
		body     interface{}
		expected string
	}{
		{
			name:     "String body",
			body:     "Hello {{username}}",
			expected: "Hello john_doe",
		},
		{
			name: "Map body",
			body: map[string]interface{}{
				"user":  "{{username}}",
				"email": "{{email}}",
			},
			expected: "john_doe", // Check one value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replaceVariablesInBody(tt.body, env)

			if str, ok := result.(string); ok {
				if str != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, str)
				}
			} else if m, ok := result.(map[string]interface{}); ok {
				if m["user"] != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, m["user"])
				}
			}
		})
	}
}

func TestFindVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Single variable",
			input:    "{{base_url}}/users",
			expected: []string{"base_url"},
		},
		{
			name:     "Multiple variables",
			input:    "{{base_url}}/{{version}}/users",
			expected: []string{"base_url", "version"},
		},
		{
			name:     "System variable",
			input:    "{{$timestamp}}",
			expected: []string{"$timestamp"},
		},
		{
			name:     "No variables",
			input:    "https://api.example.com",
			expected: []string{},
		},
		{
			name:     "Variable with spaces",
			input:    "{{ base_url }}",
			expected: []string{"base_url"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindVariables(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d variables, got %d", len(tt.expected), len(result))
			}
			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected variable '%s', got '%s'", expected, result[i])
				}
			}
		})
	}
}

func TestFindUnresolvedVariables(t *testing.T) {
	env := &EnvironmentFile{
		Name: "Test",
		Variables: map[string]*EnvironmentVariable{
			"defined": {Value: "value", Active: true},
		},
	}

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "All resolved",
			input:    "{{defined}}",
			expected: []string{},
		},
		{
			name:     "Unresolved variable",
			input:    "{{undefined}}",
			expected: []string{"undefined"},
		},
		{
			name:     "Mixed",
			input:    "{{defined}}/{{undefined}}",
			expected: []string{"undefined"},
		},
		{
			name:     "System variable",
			input:    "{{$timestamp}}",
			expected: []string{}, // System vars are always resolved
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindUnresolvedVariables(tt.input, env)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d unresolved, got %d", len(tt.expected), len(result))
			}
		})
	}
}

func TestValidateVariables(t *testing.T) {
	env := &EnvironmentFile{
		Name: "Test",
		Variables: map[string]*EnvironmentVariable{
			"base_url": {Value: "https://api.example.com", Active: true},
		},
	}

	tests := []struct {
		name               string
		req                *Request
		expectedUnresolved int
	}{
		{
			name: "All resolved",
			req: &Request{
				Method: GET,
				URL:    "{{base_url}}/users",
				Headers: map[string]string{
					"Authorization": "Bearer token",
				},
			},
			expectedUnresolved: 0,
		},
		{
			name: "Unresolved in URL",
			req: &Request{
				Method: GET,
				URL:    "{{undefined}}/users",
			},
			expectedUnresolved: 1,
		},
		{
			name: "Unresolved in header",
			req: &Request{
				Method: GET,
				URL:    "{{base_url}}/users",
				Headers: map[string]string{
					"Authorization": "Bearer {{undefined_token}}",
				},
			},
			expectedUnresolved: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unresolved := ValidateVariables(tt.req, env)
			if len(unresolved) != tt.expectedUnresolved {
				t.Errorf("Expected %d unresolved variables, got %d: %v",
					tt.expectedUnresolved, len(unresolved), unresolved)
			}
		})
	}
}

func TestPreviewVariableReplacement(t *testing.T) {
	env := &EnvironmentFile{
		Name: "Test",
		Variables: map[string]*EnvironmentVariable{
			"base_url": {Value: "https://api.example.com", Active: true},
		},
	}

	input := "{{base_url}}/users"
	expected := "https://api.example.com/users"

	result := PreviewVariableReplacement(input, env)
	if result != expected {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestUniqueStrings(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b", "d"}
	result := uniqueStrings(input)

	// Should have 4 unique strings
	if len(result) != 4 {
		t.Errorf("Expected 4 unique strings, got %d", len(result))
	}

	// Check all unique values are present
	seen := make(map[string]bool)
	for _, s := range result {
		if seen[s] {
			t.Errorf("Duplicate string found: %s", s)
		}
		seen[s] = true
	}
}
