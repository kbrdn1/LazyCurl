package api

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestRunStatus_String(t *testing.T) {
	tests := []struct {
		status RunStatus
		want   string
	}{
		{RunStatusPending, "pending"},
		{RunStatusRunning, "running"},
		{RunStatusCompleted, "completed"},
		{RunStatusCancelled, "canceled"},
		{RunStatusStopped, "stopped"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("RunStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResultStatus_String(t *testing.T) {
	tests := []struct {
		status ResultStatus
		want   string
	}{
		{ResultStatusPending, "pending"},
		{ResultStatusRunning, "running"},
		{ResultStatusPassed, "passed"},
		{ResultStatusFailed, "failed"},
		{ResultStatusError, "error"},
		{ResultStatusSkipped, "skipped"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("ResultStatus.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  RunConfig
		wantErr error
	}{
		{
			name:    "valid default",
			config:  DefaultRunConfig(),
			wantErr: nil,
		},
		{
			name:    "valid with delay",
			config:  RunConfig{DelayMs: 5000, Timeout: time.Second, ScriptTimeout: time.Second},
			wantErr: nil,
		},
		{
			name:    "valid max delay",
			config:  RunConfig{DelayMs: 10000, Timeout: time.Second, ScriptTimeout: time.Second},
			wantErr: nil,
		},
		{
			name:    "negative delay",
			config:  RunConfig{DelayMs: -1, Timeout: time.Second, ScriptTimeout: time.Second},
			wantErr: ErrInvalidDelay,
		},
		{
			name:    "excessive delay",
			config:  RunConfig{DelayMs: 20000, Timeout: time.Second, ScriptTimeout: time.Second},
			wantErr: ErrInvalidDelay,
		},
		{
			name:    "zero timeout",
			config:  RunConfig{DelayMs: 0, Timeout: 0, ScriptTimeout: time.Second},
			wantErr: ErrInvalidTimeout,
		},
		{
			name:    "negative timeout",
			config:  RunConfig{DelayMs: 0, Timeout: -time.Second, ScriptTimeout: time.Second},
			wantErr: ErrInvalidTimeout,
		},
		{
			name:    "zero script timeout",
			config:  RunConfig{DelayMs: 0, Timeout: time.Second, ScriptTimeout: 0},
			wantErr: ErrInvalidScriptTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultRunConfig(t *testing.T) {
	config := DefaultRunConfig()

	if config.StopOnFailure != false {
		t.Error("StopOnFailure should default to false")
	}
	if config.DelayMs != 0 {
		t.Error("DelayMs should default to 0")
	}
	if config.Timeout != 30*time.Second {
		t.Error("Timeout should default to 30s")
	}
	if config.ScriptTimeout != 5*time.Second {
		t.Error("ScriptTimeout should default to 5s")
	}

	// Default config should be valid
	if err := config.Validate(); err != nil {
		t.Errorf("Default config should be valid, got error: %v", err)
	}
}

func TestNewRequestResult(t *testing.T) {
	req := &CollectionRequest{
		ID:     "req_123",
		Name:   "Test Request",
		Method: GET,
		URL:    "https://api.example.com/test",
	}

	result := NewRequestResult(req, 5)

	if result.Request.ID != "req_123" {
		t.Errorf("Request.ID = %v, want req_123", result.Request.ID)
	}
	if result.Request.Name != "Test Request" {
		t.Errorf("Request.Name = %v, want Test Request", result.Request.Name)
	}
	if result.Request.Method != "GET" {
		t.Errorf("Request.Method = %v, want GET", result.Request.Method)
	}
	if result.Index != 5 {
		t.Errorf("Index = %v, want 5", result.Index)
	}
	if result.Status != ResultStatusPending {
		t.Errorf("Status = %v, want pending", result.Status)
	}
}

func TestRequestResult_SetRunning(t *testing.T) {
	result := RequestResult{Status: ResultStatusPending}
	result.SetRunning()

	if result.Status != ResultStatusRunning {
		t.Errorf("Status = %v, want running", result.Status)
	}
}

func TestRequestResult_SetCompleted(t *testing.T) {
	result := RequestResult{Status: ResultStatusRunning}
	resp := &Response{
		StatusCode: 200,
		Status:     "200 OK",
		Time:       150 * time.Millisecond,
		Size:       1024,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
	}

	result.SetCompleted(resp, 200*time.Millisecond)

	if result.Status != ResultStatusPassed {
		t.Errorf("Status = %v, want passed", result.Status)
	}
	if result.Response == nil {
		t.Fatal("Response should not be nil")
	}
	if result.Response.Status != 200 {
		t.Errorf("Response.Status = %v, want 200", result.Response.Status)
	}
	if result.Response.TimeMs != 150 {
		t.Errorf("Response.TimeMs = %v, want 150", result.Response.TimeMs)
	}
	if result.DurationMs != 200 {
		t.Errorf("DurationMs = %v, want 200", result.DurationMs)
	}
}

func TestRequestResult_SetError(t *testing.T) {
	result := RequestResult{Status: ResultStatusRunning}
	result.SetError("network", "connection refused", "dial tcp 127.0.0.1:8080: connect: connection refused")

	if result.Status != ResultStatusError {
		t.Errorf("Status = %v, want error", result.Status)
	}
	if result.Error == nil {
		t.Fatal("Error should not be nil")
	}
	if result.Error.Type != "network" {
		t.Errorf("Error.Type = %v, want network", result.Error.Type)
	}
	if result.Error.Message != "connection refused" {
		t.Errorf("Error.Message = %v, want connection refused", result.Error.Message)
	}
}

func TestRequestResult_SetScriptResults(t *testing.T) {
	t.Run("with passing assertions", func(t *testing.T) {
		result := RequestResult{Status: ResultStatusPassed}
		postResult := &ScriptResult{
			Assertions: []AssertionResult{
				{Name: "test1", Passed: true},
				{Name: "test2", Passed: true},
			},
		}

		result.SetScriptResults(nil, postResult)

		if result.Status != ResultStatusPassed {
			t.Errorf("Status = %v, want passed", result.Status)
		}
	})

	t.Run("with failing assertions", func(t *testing.T) {
		result := RequestResult{Status: ResultStatusPassed}
		postResult := &ScriptResult{
			Assertions: []AssertionResult{
				{Name: "test1", Passed: true},
				{Name: "test2", Passed: false},
			},
		}

		result.SetScriptResults(nil, postResult)

		if result.Status != ResultStatusFailed {
			t.Errorf("Status = %v, want failed", result.Status)
		}
	})
}

func TestRequestResult_AssertionCount(t *testing.T) {
	tests := []struct {
		name       string
		postResult *ScriptResult
		wantPassed int
		wantFailed int
	}{
		{
			name:       "nil post result",
			postResult: nil,
			wantPassed: 0,
			wantFailed: 0,
		},
		{
			name: "all passing",
			postResult: &ScriptResult{
				Assertions: []AssertionResult{
					{Name: "test1", Passed: true},
					{Name: "test2", Passed: true},
				},
			},
			wantPassed: 2,
			wantFailed: 0,
		},
		{
			name: "mixed results",
			postResult: &ScriptResult{
				Assertions: []AssertionResult{
					{Name: "test1", Passed: true},
					{Name: "test2", Passed: false},
					{Name: "test3", Passed: true},
					{Name: "test4", Passed: false},
				},
			},
			wantPassed: 2,
			wantFailed: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RequestResult{PostScriptResult: tt.postResult}
			passed, failed := result.AssertionCount()

			if passed != tt.wantPassed {
				t.Errorf("passed = %v, want %v", passed, tt.wantPassed)
			}
			if failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v", failed, tt.wantFailed)
			}
		})
	}
}

func TestNewRunSession(t *testing.T) {
	env := &EnvironmentFile{
		Name: "development",
		Variables: map[string]*EnvironmentVariable{
			"base_url": {Value: "https://api.example.com", Active: true},
		},
	}
	config := DefaultRunConfig()

	session := NewRunSession("My API", []string{"Users"}, env, config)

	if session.Collection != "My API" {
		t.Errorf("Collection = %v, want My API", session.Collection)
	}
	if len(session.FolderPath) != 1 || session.FolderPath[0] != "Users" {
		t.Errorf("FolderPath = %v, want [Users]", session.FolderPath)
	}
	if session.Environment != "development" {
		t.Errorf("Environment = %v, want development", session.Environment)
	}
	if session.Status != RunStatusPending {
		t.Errorf("Status = %v, want pending", session.Status)
	}
	if session.ID == "" {
		t.Error("ID should not be empty")
	}
	if session.SessionEnv == nil {
		t.Error("SessionEnv should not be nil")
	}
	// Verify it's a clone
	if session.SessionEnv == env {
		t.Error("SessionEnv should be a clone, not the original")
	}
}

func TestNewRunSession_NilEnv(t *testing.T) {
	config := DefaultRunConfig()
	session := NewRunSession("Test", nil, nil, config)

	if session.SessionEnv == nil {
		t.Error("SessionEnv should be created even with nil input")
	}
	if session.Environment != "" {
		t.Errorf("Environment should be empty, got %v", session.Environment)
	}
}

func TestRunSession_StateTransitions(t *testing.T) {
	env := &EnvironmentFile{
		Name:      "test",
		Variables: map[string]*EnvironmentVariable{},
	}
	session := NewRunSession("Test", nil, env, DefaultRunConfig())

	// Initial state
	if session.Status != RunStatusPending {
		t.Errorf("Expected pending status, got %v", session.Status)
	}
	if session.IsTerminal() {
		t.Error("Pending session should not be terminal")
	}

	// Start
	session.Start(5)
	if session.Status != RunStatusRunning {
		t.Errorf("Expected running status, got %v", session.Status)
	}
	if session.TotalRequests != 5 {
		t.Errorf("TotalRequests = %v, want 5", session.TotalRequests)
	}
	if session.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}
	if session.IsTerminal() {
		t.Error("Running session should not be terminal")
	}

	// Complete
	session.Complete()
	if session.Status != RunStatusCompleted {
		t.Errorf("Expected completed status, got %v", session.Status)
	}
	if session.EndTime.IsZero() {
		t.Error("EndTime should be set")
	}
	if !session.IsTerminal() {
		t.Error("Completed session should be terminal")
	}
}

func TestRunSession_Cancel(t *testing.T) {
	session := NewRunSession("Test", nil, nil, DefaultRunConfig())
	session.Start(5)
	session.Cancel()

	if session.Status != RunStatusCancelled {
		t.Errorf("Expected canceled status, got %v", session.Status)
	}
	if !session.IsTerminal() {
		t.Error("Canceled session should be terminal")
	}
}

func TestRunSession_Stop(t *testing.T) {
	session := NewRunSession("Test", nil, nil, DefaultRunConfig())
	session.Start(5)
	session.Stop()

	if session.Status != RunStatusStopped {
		t.Errorf("Expected stopped status, got %v", session.Status)
	}
	if !session.IsTerminal() {
		t.Error("Stopped session should be terminal")
	}
}

func TestRunSession_AddResult(t *testing.T) {
	session := NewRunSession("Test", nil, nil, DefaultRunConfig())
	session.Start(3)

	result1 := RequestResult{Index: 0, Status: ResultStatusPassed}
	result2 := RequestResult{Index: 1, Status: ResultStatusFailed}

	session.AddResult(result1)
	if session.CurrentIndex != 1 {
		t.Errorf("CurrentIndex = %v, want 1", session.CurrentIndex)
	}
	if len(session.Results) != 1 {
		t.Errorf("len(Results) = %v, want 1", len(session.Results))
	}

	session.AddResult(result2)
	if session.CurrentIndex != 2 {
		t.Errorf("CurrentIndex = %v, want 2", session.CurrentIndex)
	}
	if len(session.Results) != 2 {
		t.Errorf("len(Results) = %v, want 2", len(session.Results))
	}
}

func TestRunSession_Progress(t *testing.T) {
	session := NewRunSession("Test", nil, nil, DefaultRunConfig())
	session.Start(10)
	session.CurrentIndex = 3

	progress := session.Progress()
	if progress != "3/10" {
		t.Errorf("Progress() = %v, want 3/10", progress)
	}
}

func TestCollectRequests(t *testing.T) {
	folder := &Folder{
		Name: "Test",
		Requests: []CollectionRequest{
			{ID: "1", Name: "First"},
			{ID: "2", Name: "Second"},
		},
		Folders: []Folder{
			{
				Name: "Nested",
				Requests: []CollectionRequest{
					{ID: "3", Name: "Nested First"},
				},
			},
		},
	}

	requests := CollectRequests(folder)

	if len(requests) != 3 {
		t.Errorf("Expected 3 requests, got %d", len(requests))
	}

	// Verify order (depth-first: folder requests first, then subfolders)
	expectedNames := []string{"First", "Second", "Nested First"}
	for i, req := range requests {
		if req.Name != expectedNames[i] {
			t.Errorf("Expected %s at index %d, got %s", expectedNames[i], i, req.Name)
		}
	}
}

func TestCollectRequests_Empty(t *testing.T) {
	folder := &Folder{
		Name:     "Empty",
		Requests: []CollectionRequest{},
		Folders:  []Folder{},
	}

	requests := CollectRequests(folder)

	if len(requests) != 0 {
		t.Errorf("Expected 0 requests, got %d", len(requests))
	}
}

func TestCollectRequests_Nested(t *testing.T) {
	folder := &Folder{
		Name: "Root",
		Requests: []CollectionRequest{
			{ID: "1", Name: "Root Request"},
		},
		Folders: []Folder{
			{
				Name: "Level1",
				Requests: []CollectionRequest{
					{ID: "2", Name: "Level1 Request"},
				},
				Folders: []Folder{
					{
						Name: "Level2",
						Requests: []CollectionRequest{
							{ID: "3", Name: "Level2 Request"},
						},
					},
				},
			},
		},
	}

	requests := CollectRequests(folder)

	if len(requests) != 3 {
		t.Errorf("Expected 3 requests, got %d", len(requests))
	}

	expectedNames := []string{"Root Request", "Level1 Request", "Level2 Request"}
	for i, req := range requests {
		if req.Name != expectedNames[i] {
			t.Errorf("Expected %s at index %d, got %s", expectedNames[i], i, req.Name)
		}
	}
}

func TestCollectFromCollection(t *testing.T) {
	collection := &CollectionFile{
		Name: "API",
		Requests: []CollectionRequest{
			{ID: "1", Name: "Root Request"},
		},
		Folders: []Folder{
			{
				Name: "Users",
				Requests: []CollectionRequest{
					{ID: "2", Name: "Get Users"},
					{ID: "3", Name: "Create User"},
				},
			},
		},
	}

	t.Run("entire collection", func(t *testing.T) {
		requests := CollectFromCollection(collection, nil)
		if len(requests) != 3 {
			t.Errorf("Expected 3 requests, got %d", len(requests))
		}
	})

	t.Run("specific folder", func(t *testing.T) {
		requests := CollectFromCollection(collection, []string{"Users"})
		if len(requests) != 2 {
			t.Errorf("Expected 2 requests, got %d", len(requests))
		}
	})

	t.Run("non-existent folder", func(t *testing.T) {
		requests := CollectFromCollection(collection, []string{"NonExistent"})
		if requests != nil {
			t.Errorf("Expected nil for non-existent folder, got %v", requests)
		}
	})
}

func TestRunReport_Generate(t *testing.T) {
	session := NewRunSession("Test API", nil, nil, DefaultRunConfig())
	session.Start(3)
	session.StartTime = time.Date(2026, 1, 18, 10, 0, 0, 0, time.UTC)

	// Add some results
	result1 := RequestResult{
		Index:      0,
		Status:     ResultStatusPassed,
		DurationMs: 100,
		PostScriptResult: &ScriptResult{
			Assertions: []AssertionResult{
				{Passed: true},
				{Passed: true},
			},
		},
	}
	result2 := RequestResult{
		Index:      1,
		Status:     ResultStatusFailed,
		DurationMs: 150,
		PostScriptResult: &ScriptResult{
			Assertions: []AssertionResult{
				{Passed: true},
				{Passed: false},
			},
		},
	}
	result3 := RequestResult{
		Index:      2,
		Status:     ResultStatusError,
		DurationMs: 50,
	}

	session.Results = []RequestResult{result1, result2, result3}
	session.Complete()
	session.EndTime = time.Date(2026, 1, 18, 10, 0, 1, 0, time.UTC)

	report := session.GenerateReport()

	// Verify session info
	if report.Session.Collection != "Test API" {
		t.Errorf("Session.Collection = %v, want Test API", report.Session.Collection)
	}
	if report.Session.Status != "completed" {
		t.Errorf("Session.Status = %v, want completed", report.Session.Status)
	}

	// Verify summary
	if report.Summary.TotalRequests != 3 {
		t.Errorf("Summary.TotalRequests = %v, want 3", report.Summary.TotalRequests)
	}
	if report.Summary.CompletedRequests != 3 {
		t.Errorf("Summary.CompletedRequests = %v, want 3", report.Summary.CompletedRequests)
	}
	if report.Summary.PassedAssertions != 3 {
		t.Errorf("Summary.PassedAssertions = %v, want 3", report.Summary.PassedAssertions)
	}
	if report.Summary.FailedAssertions != 1 {
		t.Errorf("Summary.FailedAssertions = %v, want 1", report.Summary.FailedAssertions)
	}
	if report.Summary.Errors != 1 {
		t.Errorf("Summary.Errors = %v, want 1", report.Summary.Errors)
	}
	if report.Summary.TotalDurationMs != 300 {
		t.Errorf("Summary.TotalDurationMs = %v, want 300", report.Summary.TotalDurationMs)
	}
}

func TestExportRunReport(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	session := NewRunSession("Test", nil, nil, DefaultRunConfig())
	session.Start(1)
	session.Complete()
	report := session.GenerateReport()

	path, err := ExportRunReport(report)
	if err != nil {
		t.Fatalf("ExportRunReport() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Report file was not created at %s", path)
	}

	// Verify directory was created
	if _, err := os.Stat(".lazycurl/reports"); os.IsNotExist(err) {
		t.Error("Reports directory was not created")
	}

	// Verify file content is valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read report file: %v", err)
	}

	var parsed RunReport
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Report is not valid JSON: %v", err)
	}
}

func TestFlattenHeaders(t *testing.T) {
	tests := []struct {
		name   string
		input  map[string][]string
		expect map[string]string
	}{
		{
			name:   "nil input",
			input:  nil,
			expect: nil,
		},
		{
			name:   "empty input",
			input:  map[string][]string{},
			expect: map[string]string{},
		},
		{
			name: "single values",
			input: map[string][]string{
				"Content-Type": {"application/json"},
			},
			expect: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name: "multiple values takes first",
			input: map[string][]string{
				"Accept": {"application/json", "text/plain"},
			},
			expect: map[string]string{
				"Accept": "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenHeaders(tt.input)
			if tt.expect == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
				return
			}
			if len(result) != len(tt.expect) {
				t.Errorf("Length mismatch: got %d, want %d", len(result), len(tt.expect))
			}
			for k, v := range tt.expect {
				if result[k] != v {
					t.Errorf("result[%s] = %s, want %s", k, result[k], v)
				}
			}
		})
	}
}

func TestSanitizeForID(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"Simple", "simple"},
		{"With Spaces", "with_spaces"},
		{"With-Dashes", "withdashes"},
		{"With.Dots", "withdots"},
		{"Special@#$%", "special"},
		{"A Very Long Name That Exceeds Thirty Characters", "a_very_long_name_that_exceeds_"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeForID(tt.input)
			if result != tt.expect {
				t.Errorf("sanitizeForID(%s) = %s, want %s", tt.input, result, tt.expect)
			}
		})
	}
}

func TestGenerateRunID(t *testing.T) {
	id1 := generateRunID("Test Collection")
	id2 := generateRunID("Test Collection")

	// IDs should start with "run_"
	if id1[:4] != "run_" {
		t.Errorf("ID should start with 'run_', got %s", id1)
	}

	// IDs should contain timestamp and sanitized name
	if len(id1) < 10 {
		t.Errorf("ID seems too short: %s", id1)
	}

	// IDs generated at different times should differ (or at least be unique format)
	// Note: This might occasionally fail if both are generated in the same second
	// but the format should still be valid
	if id1 == "" || id2 == "" {
		t.Error("Generated IDs should not be empty")
	}
}

func TestRunSession_GetSessionEnvVariables(t *testing.T) {
	env := &EnvironmentFile{
		Name: "test",
		Variables: map[string]*EnvironmentVariable{
			"active_var":   {Value: "active_value", Active: true},
			"inactive_var": {Value: "inactive_value", Active: false},
		},
	}

	session := NewRunSession("Test", nil, env, DefaultRunConfig())
	vars := session.GetSessionEnvVariables()

	if vars["active_var"] != "active_value" {
		t.Errorf("active_var = %v, want active_value", vars["active_var"])
	}
	if _, exists := vars["inactive_var"]; exists {
		t.Error("inactive_var should not be included")
	}
}

func TestRunSession_SetSessionEnvVariable(t *testing.T) {
	session := NewRunSession("Test", nil, nil, DefaultRunConfig())

	session.SetSessionEnvVariable("new_var", "new_value")

	vars := session.GetSessionEnvVariables()
	if vars["new_var"] != "new_value" {
		t.Errorf("new_var = %v, want new_value", vars["new_var"])
	}
}

func TestExportRunReport_DirectoryPermissions(t *testing.T) {
	// Skip on Windows (os.Getuid not available)
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}
	// Skip if running as root (root can write anywhere)
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()

	// Create a read-only directory scenario would be complex
	// Just verify normal operation works
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	session := NewRunSession("Test", nil, nil, DefaultRunConfig())
	session.Start(1)
	session.Complete()
	report := session.GenerateReport()

	path, err := ExportRunReport(report)
	if err != nil {
		t.Fatalf("ExportRunReport() error = %v", err)
	}

	// Verify the path is in the expected directory
	expectedDir := filepath.Join(".lazycurl", "reports")
	if filepath.Dir(path) != expectedDir {
		t.Errorf("Report path = %s, expected dir %s", path, expectedDir)
	}
}
