package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// ============================================================================
// ENUMS
// ============================================================================

// RunStatus represents the current state of a collection run.
type RunStatus string

const (
	RunStatusPending   RunStatus = "pending"   // Created but not started
	RunStatusRunning   RunStatus = "running"   // Actively executing requests
	RunStatusCompleted RunStatus = "completed" // All requests executed
	RunStatusCancelled RunStatus = "canceled"  // User canceled mid-run
	RunStatusStopped   RunStatus = "stopped"   // Halted due to failure
)

// String returns the display name for the status.
func (s RunStatus) String() string {
	return string(s)
}

// ResultStatus represents the outcome of a single request execution.
type ResultStatus string

const (
	ResultStatusPending ResultStatus = "pending" // Not yet executed
	ResultStatusRunning ResultStatus = "running" // Currently executing
	ResultStatusPassed  ResultStatus = "passed"  // All assertions passed
	ResultStatusFailed  ResultStatus = "failed"  // Some assertions failed
	ResultStatusError   ResultStatus = "error"   // HTTP or script error
	ResultStatusSkipped ResultStatus = "skipped" // Not executed (canceled/stopped)
)

// String returns the display name for the status.
func (s ResultStatus) String() string {
	return string(s)
}

// ============================================================================
// ERRORS
// ============================================================================

var (
	ErrInvalidDelay         = errors.New("delay must be between 0 and 10000 ms")
	ErrInvalidTimeout       = errors.New("timeout must be positive")
	ErrInvalidScriptTimeout = errors.New("script timeout must be positive")
	ErrNoRequests           = errors.New("no requests to run")
	ErrRunnerBusy           = errors.New("runner is already executing")
	ErrNoActiveRun          = errors.New("no active run")
)

// ============================================================================
// CONFIGURATION
// ============================================================================

// RunConfig holds configuration settings for a collection run.
type RunConfig struct {
	// StopOnFailure halts execution when an assertion fails or HTTP error occurs.
	StopOnFailure bool `json:"stopOnFailure"`

	// DelayMs is the delay in milliseconds between requests (0-10000).
	DelayMs int `json:"delayMs"`

	// Timeout is the HTTP request timeout.
	Timeout time.Duration `json:"-"`

	// ScriptTimeout is the JavaScript execution timeout.
	ScriptTimeout time.Duration `json:"-"`
}

// DefaultRunConfig returns the default configuration.
func DefaultRunConfig() RunConfig {
	return RunConfig{
		StopOnFailure: false,
		DelayMs:       0,
		Timeout:       30 * time.Second,
		ScriptTimeout: 5 * time.Second,
	}
}

// Validate checks if the configuration is valid.
func (c *RunConfig) Validate() error {
	if c.DelayMs < 0 || c.DelayMs > 10000 {
		return ErrInvalidDelay
	}
	if c.Timeout <= 0 {
		return ErrInvalidTimeout
	}
	if c.ScriptTimeout <= 0 {
		return ErrInvalidScriptTimeout
	}
	return nil
}

// ============================================================================
// REQUEST INFO & RESPONSE INFO
// ============================================================================

// RunnerRequestInfo contains metadata about the executed request.
type RunnerRequestInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Method      string `json:"method"`
	URL         string `json:"url"`
	ResolvedURL string `json:"resolvedUrl"`
}

// RunnerResponseInfo contains HTTP response data.
type RunnerResponseInfo struct {
	Status     int               `json:"status"`
	StatusText string            `json:"statusText"`
	TimeMs     int64             `json:"timeMs"`
	SizeBytes  int64             `json:"sizeBytes"`
	Headers    map[string]string `json:"headers,omitempty"`
}

// ============================================================================
// ERROR INFO
// ============================================================================

// RunnerErrorInfo contains error details.
type RunnerErrorInfo struct {
	Type    string `json:"type"` // "network", "timeout", "script"
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ============================================================================
// REQUEST RESULT
// ============================================================================

// RequestResult captures the complete outcome of a single request.
type RequestResult struct {
	Request          RunnerRequestInfo   `json:"request"`
	Response         *RunnerResponseInfo `json:"response,omitempty"`
	PreScriptResult  *ScriptResult       `json:"preScript,omitempty"`
	PostScriptResult *ScriptResult       `json:"postScript,omitempty"`
	Error            *RunnerErrorInfo    `json:"error,omitempty"`
	Duration         time.Duration       `json:"-"`
	DurationMs       int64               `json:"durationMs"`
	Index            int                 `json:"index"`
	Status           ResultStatus        `json:"status"`
}

// NewRequestResult creates a pending result for a request.
func NewRequestResult(req *CollectionRequest, index int) RequestResult {
	return RequestResult{
		Request: RunnerRequestInfo{
			ID:     req.ID,
			Name:   req.Name,
			Method: string(req.Method),
			URL:    req.URL,
		},
		Index:  index,
		Status: ResultStatusPending,
	}
}

// SetRunning marks the result as currently executing.
func (r *RequestResult) SetRunning() {
	r.Status = ResultStatusRunning
}

// SetCompleted finalizes the result with response data.
func (r *RequestResult) SetCompleted(resp *Response, duration time.Duration) {
	r.Response = &RunnerResponseInfo{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		TimeMs:     resp.Time.Milliseconds(),
		SizeBytes:  resp.Size,
		Headers:    flattenHeaders(resp.Headers),
	}
	r.Duration = duration
	r.DurationMs = duration.Milliseconds()
	r.Status = ResultStatusPassed // Updated by SetScriptResults if failures
}

// SetError marks the result as failed with error details.
func (r *RequestResult) SetError(errType, message, details string) {
	r.Error = &RunnerErrorInfo{
		Type:    errType,
		Message: message,
		Details: details,
	}
	r.Status = ResultStatusError
}

// SetScriptResults sets the pre and post script results.
func (r *RequestResult) SetScriptResults(pre, post *ScriptResult) {
	r.PreScriptResult = pre
	r.PostScriptResult = post

	// Update status based on assertions
	if post != nil && post.HasAssertionFailures() {
		r.Status = ResultStatusFailed
	}
}

// SetSkipped marks the result as skipped.
func (r *RequestResult) SetSkipped() {
	r.Status = ResultStatusSkipped
}

// AssertionCount returns passed and failed assertion counts.
func (r *RequestResult) AssertionCount() (passed, failed int) {
	if r.PostScriptResult == nil {
		return 0, 0
	}
	return r.PostScriptResult.PassedAssertionCount(), r.PostScriptResult.FailedAssertionCount()
}

// ============================================================================
// RUN SESSION
// ============================================================================

// RunSession represents a single collection run execution.
type RunSession struct {
	// Metadata
	ID          string    `json:"id"`
	Collection  string    `json:"collection"`
	FolderPath  []string  `json:"folderPath,omitempty"`
	Environment string    `json:"environment"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime,omitempty"`
	Status      RunStatus `json:"status"`

	// Configuration
	Config RunConfig `json:"config"`

	// Execution state
	Results       []RequestResult `json:"results"`
	CurrentIndex  int             `json:"-"`
	TotalRequests int             `json:"totalRequests"`

	// Session-scoped environment (modified by scripts)
	SessionEnv *EnvironmentFile `json:"-"`
}

// NewRunSession creates a new run session.
func NewRunSession(collection string, folderPath []string, env *EnvironmentFile, config RunConfig) *RunSession {
	envName := ""
	var sessionEnv *EnvironmentFile
	if env != nil {
		envName = env.Name
		sessionEnv = env.Clone()
	} else {
		sessionEnv = &EnvironmentFile{
			Name:      "default",
			Variables: make(map[string]*EnvironmentVariable),
		}
	}

	return &RunSession{
		ID:            generateRunID(collection),
		Collection:    collection,
		FolderPath:    folderPath,
		Environment:   envName,
		Status:        RunStatusPending,
		Config:        config,
		Results:       make([]RequestResult, 0),
		CurrentIndex:  0,
		TotalRequests: 0,
		SessionEnv:    sessionEnv,
	}
}

// Start transitions the session to running state.
func (s *RunSession) Start(totalRequests int) {
	s.StartTime = time.Now()
	s.Status = RunStatusRunning
	s.TotalRequests = totalRequests
}

// Complete transitions the session to completed state.
func (s *RunSession) Complete() {
	s.EndTime = time.Now()
	s.Status = RunStatusCompleted
}

// Cancel transitions the session to canceled state.
func (s *RunSession) Cancel() {
	s.EndTime = time.Now()
	s.Status = RunStatusCancelled
}

// Stop transitions the session to stopped state (failure halt).
func (s *RunSession) Stop() {
	s.EndTime = time.Now()
	s.Status = RunStatusStopped
}

// AddResult appends a request result and advances the index.
func (s *RunSession) AddResult(result RequestResult) {
	s.Results = append(s.Results, result)
	s.CurrentIndex++
}

// IsTerminal returns true if the session is in a terminal state.
func (s *RunSession) IsTerminal() bool {
	return s.Status == RunStatusCompleted ||
		s.Status == RunStatusCancelled ||
		s.Status == RunStatusStopped
}

// Progress returns current/total as a string.
func (s *RunSession) Progress() string {
	return fmt.Sprintf("%d/%d", s.CurrentIndex, s.TotalRequests)
}

// ============================================================================
// REPORT TYPES
// ============================================================================

// SessionInfo contains session metadata for the report.
type SessionInfo struct {
	ID          string    `json:"id"`
	Collection  string    `json:"collection"`
	FolderPath  []string  `json:"folderPath,omitempty"`
	Environment string    `json:"environment"`
	StartTime   string    `json:"startTime"`
	EndTime     string    `json:"endTime"`
	Status      string    `json:"status"`
	Settings    RunConfig `json:"settings"`
}

// SummaryStats contains aggregated statistics.
type SummaryStats struct {
	TotalRequests     int   `json:"totalRequests"`
	CompletedRequests int   `json:"completedRequests"`
	SkippedRequests   int   `json:"skippedRequests"`
	TotalAssertions   int   `json:"totalAssertions"`
	PassedAssertions  int   `json:"passedAssertions"`
	FailedAssertions  int   `json:"failedAssertions"`
	Errors            int   `json:"errors"`
	TotalDurationMs   int64 `json:"totalDurationMs"`
}

// RunReport is the complete JSON export structure.
type RunReport struct {
	Session SessionInfo     `json:"session"`
	Summary SummaryStats    `json:"summary"`
	Results []RequestResult `json:"results"`
}

// GenerateReport creates a report from a completed session.
func (s *RunSession) GenerateReport() *RunReport {
	report := &RunReport{
		Session: SessionInfo{
			ID:          s.ID,
			Collection:  s.Collection,
			FolderPath:  s.FolderPath,
			Environment: s.Environment,
			StartTime:   s.StartTime.Format(time.RFC3339),
			EndTime:     s.EndTime.Format(time.RFC3339),
			Status:      string(s.Status),
			Settings:    s.Config,
		},
		Results: s.Results,
	}

	// Calculate summary
	summary := SummaryStats{
		TotalRequests: s.TotalRequests,
	}

	for _, r := range s.Results {
		switch r.Status {
		case ResultStatusPassed, ResultStatusFailed:
			summary.CompletedRequests++
			passed, failed := r.AssertionCount()
			summary.TotalAssertions += passed + failed
			summary.PassedAssertions += passed
			summary.FailedAssertions += failed
		case ResultStatusError:
			summary.CompletedRequests++
			summary.Errors++
		case ResultStatusSkipped:
			summary.SkippedRequests++
		}
		summary.TotalDurationMs += r.DurationMs
	}

	report.Summary = summary
	return report
}

// ============================================================================
// REQUEST COLLECTION FUNCTIONS
// ============================================================================

// CollectRequests gathers all requests from a folder recursively (depth-first).
func CollectRequests(folder *Folder) []*CollectionRequest {
	var requests []*CollectionRequest

	// Add folder's direct requests first
	for i := range folder.Requests {
		requests = append(requests, &folder.Requests[i])
	}

	// Recurse into subfolders (depth-first)
	for i := range folder.Folders {
		requests = append(requests, CollectRequests(&folder.Folders[i])...)
	}

	return requests
}

// CollectFromCollection gathers requests from entire collection or specific folder.
func CollectFromCollection(collection *CollectionFile, folderPath []string) []*CollectionRequest {
	if len(folderPath) == 0 {
		// Collect from entire collection
		var requests []*CollectionRequest
		for i := range collection.Requests {
			requests = append(requests, &collection.Requests[i])
		}
		for i := range collection.Folders {
			requests = append(requests, CollectRequests(&collection.Folders[i])...)
		}
		return requests
	}

	// Find specific folder
	folder := collection.findFolder(collection.Folders, folderPath, 0)
	if folder == nil {
		return nil
	}
	return CollectRequests(folder)
}

// ============================================================================
// EXPORT FUNCTION
// ============================================================================

// ExportRunReport saves the report to .lazycurl/reports/
func ExportRunReport(report *RunReport) (string, error) {
	// Ensure directory exists
	dir := ".lazycurl/reports"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create reports directory: %w", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("run_%s.json", time.Now().Format("20060102_150405"))
	path := filepath.Join(dir, filename)

	// Marshal and write
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write report file: %w", err)
	}

	return path, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// generateRunID creates a unique run session ID.
func generateRunID(collection string) string {
	// Sanitize collection name
	sanitized := sanitizeForID(collection)
	return fmt.Sprintf("run_%d_%s", time.Now().Unix(), sanitized)
}

// sanitizeForID removes special characters from a string for use in IDs.
func sanitizeForID(s string) string {
	// Replace spaces with underscores
	s = strings.ReplaceAll(s, " ", "_")
	// Keep only alphanumeric and underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	s = reg.ReplaceAllString(s, "")
	// Limit length
	if len(s) > 30 {
		s = s[:30]
	}
	return strings.ToLower(s)
}

// flattenHeaders converts multi-value headers to single values.
func flattenHeaders(headers map[string][]string) map[string]string {
	if headers == nil {
		return nil
	}
	result := make(map[string]string)
	for k, v := range headers {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}
	return result
}

// GetSessionEnvVariables returns a simple map of environment variables for script execution.
func (s *RunSession) GetSessionEnvVariables() map[string]string {
	if s.SessionEnv == nil {
		return make(map[string]string)
	}
	result := make(map[string]string)
	for k, v := range s.SessionEnv.Variables {
		if v.Active {
			result[k] = v.Value
		}
	}
	return result
}

// SetSessionEnvVariable sets a variable in the session environment.
func (s *RunSession) SetSessionEnvVariable(key, value string) {
	if s.SessionEnv == nil {
		s.SessionEnv = &EnvironmentFile{
			Name:      "session",
			Variables: make(map[string]*EnvironmentVariable),
		}
	}
	s.SessionEnv.SetVariable(key, value)
}
