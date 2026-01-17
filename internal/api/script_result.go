package api

import (
	"errors"
	"time"
)

// ScriptResult contains the outcome of script execution
type ScriptResult struct {
	// Execution status
	Success  bool          `json:"success"`
	Duration time.Duration `json:"duration"`

	// Error information (if failed)
	Error *ScriptErrorInfo `json:"error,omitempty"`

	// Console output
	ConsoleOutput []ConsoleLogEntry `json:"console_output"`

	// Assertion results
	Assertions []AssertionResult `json:"assertions"`

	// Environment changes made by script
	EnvChanges []EnvChange `json:"env_changes"`

	// Request modifications (pre-request only)
	RequestModified bool `json:"request_modified"`
}

// ScriptErrorInfo contains detailed error information for display
type ScriptErrorInfo struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Line       int    `json:"line,omitempty"`
	Column     int    `json:"column,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// NewScriptResult creates a new successful result
func NewScriptResult() *ScriptResult {
	return &ScriptResult{
		Success:       true,
		ConsoleOutput: make([]ConsoleLogEntry, 0),
		Assertions:    make([]AssertionResult, 0),
		EnvChanges:    make([]EnvChange, 0),
	}
}

// NewScriptResultWithError creates a result with an error
func NewScriptResultWithError(err error) *ScriptResult {
	result := NewScriptResult()
	result.Success = false
	result.SetError(err)
	return result
}

// SetError sets the error information from an error
func (r *ScriptResult) SetError(err error) {
	if err == nil {
		r.Error = nil
		return
	}

	r.Success = false
	r.Error = &ScriptErrorInfo{
		Message: err.Error(),
	}

	// Extract detailed info from typed errors
	{
		var e *ScriptExecutionError
		var e1 *ScriptSyntaxError
		var e2 *ScriptTimeoutError
		switch {
		case errors.As(err, &e):
			r.Error.Type = "ExecutionError"
			r.Error.Message = e.Message
			r.Error.Line = e.Line
			r.Error.Column = e.Column
			r.Error.StackTrace = e.StackTrace
		case errors.As(err, &e1):
			r.Error.Type = "SyntaxError"
			r.Error.Message = e1.Message
			r.Error.Line = e1.Line
			r.Error.Column = e1.Column
		case errors.As(err, &e2):
			r.Error.Type = "TimeoutError"
			r.Error.Message = e2.Error()
		default:
			r.Error.Type = "Error"
		}
	}
}

// HasAssertionFailures returns true if any assertion failed
func (r *ScriptResult) HasAssertionFailures() bool {
	for _, a := range r.Assertions {
		if !a.Passed {
			return true
		}
	}
	return false
}

// PassedAssertionCount returns the number of passed assertions
func (r *ScriptResult) PassedAssertionCount() int {
	count := 0
	for _, a := range r.Assertions {
		if a.Passed {
			count++
		}
	}
	return count
}

// FailedAssertionCount returns the number of failed assertions
func (r *ScriptResult) FailedAssertionCount() int {
	count := 0
	for _, a := range r.Assertions {
		if !a.Passed {
			count++
		}
	}
	return count
}
