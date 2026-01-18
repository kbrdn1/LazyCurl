package ui

import (
	"github.com/kbrdn1/LazyCurl/internal/api"
)

// ScriptType indicates when a script runs
type ScriptType string

const (
	ScriptTypePreRequest   ScriptType = "pre-request"
	ScriptTypePostResponse ScriptType = "post-response"
)

// ScriptExecutionRequestMsg requests script execution
type ScriptExecutionRequestMsg struct {
	Type     ScriptType          // When this script runs
	Script   string              // JavaScript code to execute
	Request  *api.ScriptRequest  // Request data (always provided)
	Response *api.ScriptResponse // Response data (only for post-response)
	Env      *api.Environment    // Environment for variable access
}

// ScriptExecutionResultMsg contains script execution results
type ScriptExecutionResultMsg struct {
	Type    ScriptType         // Which script ran
	Result  *api.ScriptResult  // Execution result
	Request *api.ScriptRequest // Modified request (for pre-request scripts)
	Error   error              // Execution error (if any)
}

// ScriptConsoleOutputMsg notifies UI of new console output
type ScriptConsoleOutputMsg struct {
	Entries []api.ConsoleLogEntry // Console log entries from script
	Type    ScriptType            // Which script produced this output
}

// ScriptAssertionResultsMsg notifies UI of assertion results
type ScriptAssertionResultsMsg struct {
	Assertions []api.AssertionResult // Test assertion results
	AllPassed  bool                  // True if all assertions passed
	PassCount  int                   // Number of passed assertions
	FailCount  int                   // Number of failed assertions
}

// ScriptEnvChangesMsg notifies UI of environment variable changes
type ScriptEnvChangesMsg struct {
	Changes []api.EnvChange // Environment variable changes
	Applied bool            // True if changes were applied
}

// ScriptErrorDisplayMsg requests error display in UI
type ScriptErrorDisplayMsg struct {
	Error   *api.ScriptErrorInfo // Error details
	Type    ScriptType           // Which script failed
	Context string               // Additional context (e.g., "line 5")
}
