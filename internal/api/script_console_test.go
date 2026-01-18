package api

import (
	"strings"
	"testing"
)

func TestNewScriptConsole(t *testing.T) {
	console := NewScriptConsole()
	if console == nil {
		t.Fatal("NewScriptConsole() returned nil")
	}
	entries := console.GetEntries()
	if len(entries) != 0 {
		t.Errorf("new console should have 0 entries, got %d", len(entries))
	}
}

func TestScriptConsole_Log(t *testing.T) {
	console := NewScriptConsole()
	console.Log("hello", "world")

	entries := console.GetEntries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Level != LogLevelLog {
		t.Errorf("expected level %q, got %q", LogLevelLog, entries[0].Level)
	}
	if entries[0].Message != "hello world" {
		t.Errorf("expected message %q, got %q", "hello world", entries[0].Message)
	}
}

func TestScriptConsole_AllLevels(t *testing.T) {
	tests := []struct {
		method   func(*ScriptConsole, ...interface{})
		level    ConsoleLogLevel
		levelStr string
	}{
		{(*ScriptConsole).Log, LogLevelLog, "log"},
		{(*ScriptConsole).Info, LogLevelInfo, "info"},
		{(*ScriptConsole).Warn, LogLevelWarn, "warn"},
		{(*ScriptConsole).Error, LogLevelError, "error"},
		{(*ScriptConsole).Debug, LogLevelDebug, "debug"},
	}

	for _, tt := range tests {
		t.Run(tt.levelStr, func(t *testing.T) {
			console := NewScriptConsole()
			tt.method(console, "test message")

			entries := console.GetEntries()
			if len(entries) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(entries))
			}
			if entries[0].Level != tt.level {
				t.Errorf("expected level %q, got %q", tt.level, entries[0].Level)
			}
		})
	}
}

func TestScriptConsole_Clear(t *testing.T) {
	console := NewScriptConsole()
	console.Log("message 1")
	console.Log("message 2")

	if len(console.GetEntries()) != 2 {
		t.Fatal("expected 2 entries before clear")
	}

	console.Clear()

	if len(console.GetEntries()) != 0 {
		t.Error("expected 0 entries after clear")
	}
}

func TestScriptConsole_GetEntries_ReturnsCopy(t *testing.T) {
	console := NewScriptConsole()
	console.Log("original")

	entries1 := console.GetEntries()
	entries1[0].Message = "modified"

	entries2 := console.GetEntries()
	if entries2[0].Message == "modified" {
		t.Error("GetEntries should return a copy, not the original slice")
	}
}

func TestFormatArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected string
	}{
		{
			name:     "empty args",
			args:     []interface{}{},
			expected: "",
		},
		{
			name:     "single string",
			args:     []interface{}{"hello"},
			expected: "hello",
		},
		{
			name:     "multiple strings",
			args:     []interface{}{"hello", "world"},
			expected: "hello world",
		},
		{
			name:     "nil value",
			args:     []interface{}{nil},
			expected: "undefined",
		},
		{
			name:     "boolean true",
			args:     []interface{}{true},
			expected: "true",
		},
		{
			name:     "boolean false",
			args:     []interface{}{false},
			expected: "false",
		},
		{
			name:     "integer",
			args:     []interface{}{42},
			expected: "42",
		},
		{
			name:     "float",
			args:     []interface{}{3.14},
			expected: "3.14",
		},
		{
			name:     "empty object",
			args:     []interface{}{map[string]interface{}{}},
			expected: "{}",
		},
		{
			name:     "empty array",
			args:     []interface{}{[]interface{}{}},
			expected: "[]",
		},
		{
			name:     "mixed types",
			args:     []interface{}{"count:", 42, "active:", true},
			expected: "count: 42 active: true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatArgs(tt.args...)
			if result != tt.expected {
				t.Errorf("formatArgs() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatArg_Object(t *testing.T) {
	obj := map[string]interface{}{
		"name": "test",
		"age":  25,
	}

	result := formatArg(obj)

	// Object formatting may have different key order, so check for presence
	if !strings.Contains(result, "name: test") {
		t.Errorf("expected result to contain 'name: test', got %q", result)
	}
	if !strings.Contains(result, "age: 25") {
		t.Errorf("expected result to contain 'age: 25', got %q", result)
	}
	if !strings.HasPrefix(result, "{ ") || !strings.HasSuffix(result, " }") {
		t.Errorf("expected result to be wrapped in braces, got %q", result)
	}
}

func TestFormatArg_Array(t *testing.T) {
	arr := []interface{}{"a", "b", "c"}

	result := formatArg(arr)
	expected := "[ a, b, c ]"

	if result != expected {
		t.Errorf("formatArg() = %q, want %q", result, expected)
	}
}
