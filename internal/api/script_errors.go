package api

import (
	"fmt"
	"time"
)

// ScriptExecutionError wraps script runtime errors with location information
type ScriptExecutionError struct {
	Message    string
	Line       int
	Column     int
	StackTrace string
	Cause      error
}

// Error implements the error interface
func (e *ScriptExecutionError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("script error at line %d, column %d: %s", e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("script error: %s", e.Message)
}

// Unwrap returns the underlying cause
func (e *ScriptExecutionError) Unwrap() error {
	return e.Cause
}

// ScriptTimeoutError indicates script exceeded time limit
type ScriptTimeoutError struct {
	Timeout time.Duration
}

// Error implements the error interface
func (e *ScriptTimeoutError) Error() string {
	return fmt.Sprintf("script execution timed out after %v", e.Timeout)
}

// ScriptSyntaxError indicates JavaScript syntax error
type ScriptSyntaxError struct {
	Message string
	Line    int
	Column  int
}

// Error implements the error interface
func (e *ScriptSyntaxError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("syntax error at line %d, column %d: %s", e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("syntax error: %s", e.Message)
}
