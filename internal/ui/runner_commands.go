package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kbrdn1/LazyCurl/internal/api"
)

// ============================================================================
// RUNNER COMMAND FUNCTIONS
// ============================================================================

// ExecuteCollectionRunCmd initializes and starts a collection run.
func ExecuteCollectionRunCmd(
	collection *api.CollectionFile,
	folderPath []string,
	env *api.EnvironmentFile,
	config api.RunConfig,
) tea.Cmd {
	return func() tea.Msg {
		// Validate configuration
		if err := config.Validate(); err != nil {
			return RunnerErrorMsg{Error: err}
		}

		// Collect requests from collection or folder
		requests := api.CollectFromCollection(collection, folderPath)
		if len(requests) == 0 {
			return RunnerErrorMsg{Error: api.ErrNoRequests}
		}

		// Create session
		session := api.NewRunSession(collection.Name, folderPath, env, config)
		session.Start(len(requests))

		return RunnerStartedMsg{
			Session:  session,
			Requests: requests,
		}
	}
}

// ExecuteNextRequestCmd executes the next request in the run.
func ExecuteNextRequestCmd(
	session *api.RunSession,
	requests []*api.CollectionRequest,
	client *api.Client,
	executor api.ScriptExecutor,
) tea.Cmd {
	return func() tea.Msg {
		// Check if run is already complete or canceled
		if session.IsTerminal() {
			report := session.GenerateReport()
			return RunnerCompleteMsg{
				Session: session,
				Report:  report,
			}
		}

		// Check if all requests completed
		if session.CurrentIndex >= len(requests) {
			session.Complete()
			report := session.GenerateReport()
			return RunnerCompleteMsg{
				Session: session,
				Report:  report,
			}
		}

		// Get current request
		collReq := requests[session.CurrentIndex]
		result := api.NewRequestResult(collReq, session.CurrentIndex)
		result.SetRunning()

		startTime := time.Now()

		// Convert to HTTP request
		httpReq := collReq.ToRequest()

		// Execute pre-request script FIRST if present
		// This allows scripts to modify environment variables before substitution
		var preScriptResult *api.ScriptResult
		preScript := getPreRequestScript(collReq)
		if preScript != "" {
			// Create script request from raw (unsubstituted) HTTP request
			scriptReq := api.NewScriptRequestFromHTTP(httpReq)
			env := sessionEnvToEnvironment(session.SessionEnv)

			var err error
			preScriptResult, err = executor.ExecutePreRequest(preScript, scriptReq, env)
			if err != nil {
				result.SetError("script", "Pre-request script error", err.Error())
				result.Duration = time.Since(startTime)
				result.DurationMs = result.Duration.Milliseconds()
				session.AddResult(result)

				// Check stop on failure
				if session.Config.StopOnFailure {
					session.Stop()
					markRemainingSkipped(session, requests)
					return RunnerCompleteMsg{
						Session: session,
						Report:  session.GenerateReport(),
					}
				}

				return runnerRequestComplete(session, result)
			}

			// Apply environment changes from pre-request script FIRST
			// This ensures new variables are available for substitution
			applyEnvChanges(session, preScriptResult)

			// Apply script modifications to request (URL, headers, body changes)
			applyScriptModifications(httpReq, scriptReq)
		}

		// Apply variable substitution from session environment AFTER pre-request script
		// This allows scripts to define variables that get substituted
		httpReq = api.ReplaceVariablesInRequest(httpReq, session.SessionEnv)

		// Store resolved URL
		result.Request.ResolvedURL = httpReq.URL

		// Send HTTP request
		resp, err := client.Send(httpReq)
		if err != nil {
			errType := "network"
			if isTimeoutError(err) {
				errType = "timeout"
			}
			result.SetError(errType, "HTTP request failed", err.Error())
			result.Duration = time.Since(startTime)
			result.DurationMs = result.Duration.Milliseconds()
			result.PreScriptResult = preScriptResult
			session.AddResult(result)

			// Check stop on failure
			if session.Config.StopOnFailure {
				session.Stop()
				markRemainingSkipped(session, requests)
				return RunnerCompleteMsg{
					Session: session,
					Report:  session.GenerateReport(),
				}
			}

			return runnerRequestComplete(session, result)
		}

		// Set completed with response
		result.SetCompleted(resp, time.Since(startTime))

		// Execute post-response script if present
		var postScriptResult *api.ScriptResult
		postScript := getPostResponseScript(collReq)
		if postScript != "" {
			scriptReq := api.NewScriptRequestFromHTTP(httpReq)
			scriptResp := createScriptResponseFromHTTP(resp)
			env := sessionEnvToEnvironment(session.SessionEnv)

			var err error
			postScriptResult, err = executor.ExecutePostResponse(postScript, scriptReq, scriptResp, env)
			if err != nil {
				result.SetError("script", "Post-response script error", err.Error())
			} else {
				// Apply environment changes from post-response script
				applyEnvChanges(session, postScriptResult)
			}
		}

		// Set script results
		result.SetScriptResults(preScriptResult, postScriptResult)

		// Finalize timing
		result.Duration = time.Since(startTime)
		result.DurationMs = result.Duration.Milliseconds()

		// Add result to session
		session.AddResult(result)

		// Check stop on failure (for failed assertions)
		if session.Config.StopOnFailure && (result.Status == api.ResultStatusFailed || result.Status == api.ResultStatusError) {
			session.Stop()
			markRemainingSkipped(session, requests)
			return RunnerCompleteMsg{
				Session: session,
				Report:  session.GenerateReport(),
			}
		}

		return runnerRequestComplete(session, result)
	}
}

// DelayCmd returns a command that waits for the specified delay.
func DelayCmd(delayMs int) tea.Cmd {
	if delayMs <= 0 {
		return nil
	}
	return tea.Tick(time.Duration(delayMs)*time.Millisecond, func(t time.Time) tea.Msg {
		return RunnerDelayCompleteMsg{}
	})
}

// ExportRunReportCmd exports the run report to a file.
func ExportRunReportCmd(report *api.RunReport) tea.Cmd {
	return func() tea.Msg {
		path, err := api.ExportRunReport(report)
		return RunnerExportedMsg{
			FilePath: path,
			Error:    err,
		}
	}
}

// CancelRunCmd cancels the current run.
func CancelRunCmd(session *api.RunSession, requests []*api.CollectionRequest) tea.Cmd {
	return func() tea.Msg {
		session.Cancel()
		markRemainingSkipped(session, requests)
		return RunnerCancelledMsg{
			Session: session,
			Report:  session.GenerateReport(),
		}
	}
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// runnerRequestComplete creates a RunnerRequestCompleteMsg.
func runnerRequestComplete(session *api.RunSession, result api.RequestResult) tea.Msg {
	return RunnerRequestCompleteMsg{
		Result:  result,
		Session: session,
	}
}

// markRemainingSkipped marks all remaining requests as skipped.
func markRemainingSkipped(session *api.RunSession, requests []*api.CollectionRequest) {
	for i := session.CurrentIndex; i < len(requests); i++ {
		result := api.NewRequestResult(requests[i], i)
		result.SetSkipped()
		session.Results = append(session.Results, result)
	}
}

// sessionEnvToEnvironment converts EnvironmentFile to Environment for script execution.
func sessionEnvToEnvironment(envFile *api.EnvironmentFile) *api.Environment {
	if envFile == nil {
		return &api.Environment{
			Name:      "default",
			Variables: make(map[string]string),
		}
	}

	env := &api.Environment{
		Name:      envFile.Name,
		Variables: make(map[string]string),
	}
	for k, v := range envFile.Variables {
		if v.Active {
			env.Variables[k] = v.Value
		}
	}
	return env
}

// getPreRequestScript extracts the pre-request script from a CollectionRequest.
func getPreRequestScript(req *api.CollectionRequest) string {
	if req.Scripts == nil {
		return ""
	}
	return req.Scripts.PreRequest
}

// getPostResponseScript extracts the post-response script from a CollectionRequest.
func getPostResponseScript(req *api.CollectionRequest) string {
	if req.Scripts == nil {
		return ""
	}
	return req.Scripts.PostRequest
}

// createScriptResponseFromHTTP creates a ScriptResponse from an HTTP Response.
func createScriptResponseFromHTTP(resp *api.Response) *api.ScriptResponse {
	if resp == nil {
		return api.NewScriptResponseFromData(0, "", nil, "", 0)
	}

	// Flatten headers
	headers := make(map[string]string)
	for k, v := range resp.Headers {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return api.NewScriptResponseFromData(
		resp.StatusCode,
		resp.Status,
		headers,
		resp.Body,
		resp.Time.Milliseconds(),
	)
}

// applyScriptModifications applies script modifications to the HTTP request.
func applyScriptModifications(httpReq *api.Request, scriptReq *api.ScriptRequest) {
	// Apply URL changes
	if scriptReq.URL() != "" && scriptReq.IsModified() {
		httpReq.URL = scriptReq.URL()
	}

	// Apply header changes
	for k, v := range scriptReq.Headers() {
		httpReq.Headers[k] = v
	}

	// Apply body changes
	if scriptReq.IsBodyModified() {
		body := scriptReq.Body()
		httpReq.Body = &body
	}
}

// applyEnvChanges applies environment changes from script result to session.
func applyEnvChanges(session *api.RunSession, result *api.ScriptResult) {
	if result == nil {
		return
	}
	for _, change := range result.EnvChanges {
		switch change.Type {
		case api.EnvChangeSet:
			session.SetSessionEnvVariable(change.Name, change.Value)
		case api.EnvChangeUnset:
			if session.SessionEnv != nil {
				delete(session.SessionEnv.Variables, change.Name)
			}
		}
	}
}

// isTimeoutError checks if an error is a timeout error.
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "timeout") || strings.Contains(errStr, "deadline exceeded")
}

// ============================================================================
// CHAIN COMMANDS
// ============================================================================

// ChainNextRequestCmd chains the next request execution with optional delay.
func ChainNextRequestCmd(
	session *api.RunSession,
	requests []*api.CollectionRequest,
	client *api.Client,
	executor api.ScriptExecutor,
) tea.Cmd {
	// Check for delay
	if session.Config.DelayMs > 0 {
		// Return delay command, then execute next
		return tea.Sequence(
			DelayCmd(session.Config.DelayMs),
			func() tea.Msg {
				return RunnerExecuteNextMsg{}
			},
		)
	}

	// No delay, execute immediately
	return func() tea.Msg {
		return RunnerExecuteNextMsg{}
	}
}

// FormatRunnerError formats a runner error for display.
func FormatRunnerError(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("Runner error: %s", err.Error())
}
