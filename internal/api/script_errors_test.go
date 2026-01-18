package api

import (
	"errors"
	"testing"
	"time"
)

func TestScriptExecutionError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ScriptExecutionError
		expected string
	}{
		{
			name: "error with line and column",
			err: &ScriptExecutionError{
				Message: "undefined variable 'foo'",
				Line:    10,
				Column:  5,
			},
			expected: "script error at line 10, column 5: undefined variable 'foo'",
		},
		{
			name: "error without line info",
			err: &ScriptExecutionError{
				Message: "general error",
				Line:    0,
			},
			expected: "script error: general error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestScriptExecutionError_Unwrap(t *testing.T) {
	cause := errors.New("underlying cause")
	err := &ScriptExecutionError{
		Message: "wrapper",
		Cause:   cause,
	}

	if unwrapped := err.Unwrap(); !errors.Is(unwrapped, cause) {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestScriptTimeoutError_Error(t *testing.T) {
	err := &ScriptTimeoutError{
		Timeout: 5 * time.Second,
	}

	expected := "script execution timed out after 5s"
	if got := err.Error(); got != expected {
		t.Errorf("Error() = %q, want %q", got, expected)
	}
}

func TestScriptSyntaxError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ScriptSyntaxError
		expected string
	}{
		{
			name: "error with line and column",
			err: &ScriptSyntaxError{
				Message: "unexpected token",
				Line:    5,
				Column:  10,
			},
			expected: "syntax error at line 5, column 10: unexpected token",
		},
		{
			name: "error without line info",
			err: &ScriptSyntaxError{
				Message: "invalid syntax",
				Line:    0,
			},
			expected: "syntax error: invalid syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}
