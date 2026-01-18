package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/kbrdn1/LazyCurl/internal/api"
)

func TestNewScriptConsoleFormatter(t *testing.T) {
	f := NewScriptConsoleFormatter()
	if f == nil {
		t.Fatal("NewScriptConsoleFormatter() returned nil")
	}
}

func TestScriptConsoleFormatter_FormatEntry(t *testing.T) {
	f := NewScriptConsoleFormatter()
	ts := time.Date(2025, 1, 17, 14, 30, 45, 0, time.UTC)

	tests := []struct {
		name     string
		entry    api.ConsoleLogEntry
		width    int
		contains []string
	}{
		{
			name: "log level",
			entry: api.ConsoleLogEntry{
				Level:     api.LogLevelLog,
				Message:   "test message",
				Timestamp: ts,
			},
			width:    80,
			contains: []string{"14:30:45", "test message"},
		},
		{
			name: "info level",
			entry: api.ConsoleLogEntry{
				Level:     api.LogLevelInfo,
				Message:   "info message",
				Timestamp: ts,
			},
			width:    80,
			contains: []string{"14:30:45", "ℹ", "info message"},
		},
		{
			name: "warn level",
			entry: api.ConsoleLogEntry{
				Level:     api.LogLevelWarn,
				Message:   "warning message",
				Timestamp: ts,
			},
			width:    80,
			contains: []string{"14:30:45", "⚠", "warning message"},
		},
		{
			name: "error level",
			entry: api.ConsoleLogEntry{
				Level:     api.LogLevelError,
				Message:   "error message",
				Timestamp: ts,
			},
			width:    80,
			contains: []string{"14:30:45", "✖", "error message"},
		},
		{
			name: "debug level",
			entry: api.ConsoleLogEntry{
				Level:     api.LogLevelDebug,
				Message:   "debug message",
				Timestamp: ts,
			},
			width:    80,
			contains: []string{"14:30:45", "◌", "debug message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.FormatEntry(tt.entry, tt.width)
			for _, want := range tt.contains {
				if !strings.Contains(result, want) {
					t.Errorf("FormatEntry() missing %q in output: %s", want, result)
				}
			}
		})
	}
}

func TestScriptConsoleFormatter_FormatEntry_Truncation(t *testing.T) {
	f := NewScriptConsoleFormatter()
	ts := time.Now()

	longMessage := strings.Repeat("a", 200)
	entry := api.ConsoleLogEntry{
		Level:     api.LogLevelLog,
		Message:   longMessage,
		Timestamp: ts,
	}

	result := f.FormatEntry(entry, 50) // Small width to force truncation

	if !strings.Contains(result, "...") {
		t.Error("Long message should be truncated with ellipsis")
	}
}

func TestScriptConsoleFormatter_FormatEntries(t *testing.T) {
	f := NewScriptConsoleFormatter()
	ts := time.Now()

	t.Run("empty entries", func(t *testing.T) {
		result := f.FormatEntries(nil, 80, 10)
		if !strings.Contains(result, "No console output") {
			t.Error("Empty entries should show 'No console output'")
		}
	})

	t.Run("multiple entries", func(t *testing.T) {
		entries := []api.ConsoleLogEntry{
			{Level: api.LogLevelLog, Message: "first", Timestamp: ts},
			{Level: api.LogLevelInfo, Message: "second", Timestamp: ts},
			{Level: api.LogLevelWarn, Message: "third", Timestamp: ts},
		}

		result := f.FormatEntries(entries, 80, 10)

		if !strings.Contains(result, "first") {
			t.Error("Should contain 'first'")
		}
		if !strings.Contains(result, "second") {
			t.Error("Should contain 'second'")
		}
		if !strings.Contains(result, "third") {
			t.Error("Should contain 'third'")
		}
	})

	t.Run("limited lines", func(t *testing.T) {
		entries := make([]api.ConsoleLogEntry, 10)
		for i := range entries {
			entries[i] = api.ConsoleLogEntry{
				Level:     api.LogLevelLog,
				Message:   "message",
				Timestamp: ts,
			}
		}

		result := f.FormatEntries(entries, 80, 5)

		if !strings.Contains(result, "earlier entries hidden") {
			t.Error("Should indicate hidden entries when exceeding maxLines")
		}
	})
}

func TestScriptConsoleFormatter_FormatHeader(t *testing.T) {
	f := NewScriptConsoleFormatter()

	t.Run("pre-request", func(t *testing.T) {
		result := f.FormatHeader(ScriptTypePreRequest, 80)
		if !strings.Contains(result, "Pre-Request") {
			t.Error("Pre-request header should contain 'Pre-Request'")
		}
	})

	t.Run("post-response", func(t *testing.T) {
		result := f.FormatHeader(ScriptTypePostResponse, 80)
		if !strings.Contains(result, "Post-Response") {
			t.Error("Post-response header should contain 'Post-Response'")
		}
	})
}

func TestScriptConsoleFormatter_FormatAssertionResults(t *testing.T) {
	f := NewScriptConsoleFormatter()

	t.Run("no assertions", func(t *testing.T) {
		result := f.FormatAssertionResults(nil, 80)
		if !strings.Contains(result, "No assertions") {
			t.Error("Empty assertions should show 'No assertions'")
		}
	})

	t.Run("all passed", func(t *testing.T) {
		results := []api.AssertionResult{
			{Name: "test 1", Passed: true},
			{Name: "test 2", Passed: true},
		}

		result := f.FormatAssertionResults(results, 80)

		if !strings.Contains(result, "2 passed") {
			t.Error("Should show 2 passed")
		}
		if !strings.Contains(result, "✓") {
			t.Error("Should show pass icon")
		}
		if strings.Contains(result, "failed") {
			t.Error("Should not mention failed when all passed")
		}
	})

	t.Run("with failures", func(t *testing.T) {
		results := []api.AssertionResult{
			{Name: "test 1", Passed: true},
			{Name: "test 2", Passed: false, Expected: 200, Actual: 404, Message: "wrong status"},
		}

		result := f.FormatAssertionResults(results, 80)

		if !strings.Contains(result, "1 passed") {
			t.Error("Should show 1 passed")
		}
		if !strings.Contains(result, "1 failed") {
			t.Error("Should show 1 failed")
		}
		if !strings.Contains(result, "✗") {
			t.Error("Should show fail icon")
		}
		if !strings.Contains(result, "Expected") {
			t.Error("Should show expected value")
		}
		if !strings.Contains(result, "Actual") {
			t.Error("Should show actual value")
		}
	})
}

func TestScriptConsoleFormatter_FormatEnvChanges(t *testing.T) {
	f := NewScriptConsoleFormatter()

	t.Run("no changes", func(t *testing.T) {
		result := f.FormatEnvChanges(nil, 80)
		if !strings.Contains(result, "No environment changes") {
			t.Error("Empty changes should show 'No environment changes'")
		}
	})

	t.Run("set new variable", func(t *testing.T) {
		changes := []api.EnvChange{
			{Type: api.EnvChangeSet, Name: "API_KEY", Value: "secret123"},
		}

		result := f.FormatEnvChanges(changes, 80)

		if !strings.Contains(result, "+") {
			t.Error("New variable should show + icon")
		}
		if !strings.Contains(result, "API_KEY") {
			t.Error("Should contain variable name")
		}
		if !strings.Contains(result, "set") {
			t.Error("Should indicate 'set' action")
		}
	})

	t.Run("update existing variable", func(t *testing.T) {
		changes := []api.EnvChange{
			{Type: api.EnvChangeSet, Name: "API_KEY", Value: "new_value", Previous: "old_value"},
		}

		result := f.FormatEnvChanges(changes, 80)

		if !strings.Contains(result, "~") {
			t.Error("Updated variable should show ~ icon")
		}
		if !strings.Contains(result, "updated") {
			t.Error("Should indicate 'updated' action")
		}
	})

	t.Run("unset variable", func(t *testing.T) {
		changes := []api.EnvChange{
			{Type: api.EnvChangeUnset, Name: "OLD_VAR", Previous: "value"},
		}

		result := f.FormatEnvChanges(changes, 80)

		if !strings.Contains(result, "-") {
			t.Error("Unset variable should show - icon")
		}
		if !strings.Contains(result, "unset") {
			t.Error("Should indicate 'unset' action")
		}
	})
}

func TestScriptConsoleFormatter_FormatScriptError(t *testing.T) {
	f := NewScriptConsoleFormatter()

	t.Run("nil error", func(t *testing.T) {
		result := f.FormatScriptError(nil, 80)
		if result != "" {
			t.Error("Nil error should return empty string")
		}
	})

	t.Run("basic error", func(t *testing.T) {
		err := &api.ScriptErrorInfo{
			Type:    "SyntaxError",
			Message: "Unexpected token",
		}

		result := f.FormatScriptError(err, 80)

		if !strings.Contains(result, "SyntaxError") {
			t.Error("Should contain error type")
		}
		if !strings.Contains(result, "Unexpected token") {
			t.Error("Should contain error message")
		}
	})

	t.Run("error with line/column", func(t *testing.T) {
		err := &api.ScriptErrorInfo{
			Type:    "SyntaxError",
			Message: "Unexpected token",
			Line:    5,
			Column:  10,
		}

		result := f.FormatScriptError(err, 80)

		if !strings.Contains(result, "line 5") {
			t.Error("Should contain line number")
		}
		if !strings.Contains(result, "column 10") {
			t.Error("Should contain column number")
		}
	})

	t.Run("error with stack trace", func(t *testing.T) {
		err := &api.ScriptErrorInfo{
			Type:       "Error",
			Message:    "Something went wrong",
			StackTrace: "at foo (script.js:10)\nat bar (script.js:20)",
		}

		result := f.FormatScriptError(err, 80)

		if !strings.Contains(result, "Stack trace") {
			t.Error("Should show stack trace header")
		}
		if !strings.Contains(result, "at foo") {
			t.Error("Should contain stack trace content")
		}
	})
}

func TestScriptConsoleFormatter_getLevelStyle(t *testing.T) {
	f := NewScriptConsoleFormatter()

	tests := []struct {
		level    api.ConsoleLogLevel
		wantIcon string
	}{
		{api.LogLevelLog, "●"},
		{api.LogLevelInfo, "ℹ"},
		{api.LogLevelWarn, "⚠"},
		{api.LogLevelError, "✖"},
		{api.LogLevelDebug, "◌"},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			icon, _ := f.getLevelStyle(tt.level)
			if icon != tt.wantIcon {
				t.Errorf("getLevelStyle(%s) icon = %q, want %q", tt.level, icon, tt.wantIcon)
			}
		})
	}
}
