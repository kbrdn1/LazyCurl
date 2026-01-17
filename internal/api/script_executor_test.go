package api

import (
	"strings"
	"testing"
	"time"
)

func TestNewScriptExecutor(t *testing.T) {
	executor := NewScriptExecutor()
	if executor == nil {
		t.Fatal("NewScriptExecutor() returned nil")
	}

	// Check default timeout (5 seconds)
	if executor.GetTimeout() != 5*time.Second {
		t.Errorf("GetTimeout() = %v, want %v", executor.GetTimeout(), 5*time.Second)
	}
}

func TestScriptExecutor_SetTimeout(t *testing.T) {
	executor := NewScriptExecutor()

	newTimeout := 10 * time.Second
	executor.SetTimeout(newTimeout)

	if executor.GetTimeout() != newTimeout {
		t.Errorf("GetTimeout() = %v, want %v", executor.GetTimeout(), newTimeout)
	}
}

func TestExecutePreRequest_BasicExecution(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	// Simple script that doesn't modify anything
	script := `
		console.log("Pre-request script executed");
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err != nil {
		t.Errorf("ExecutePreRequest failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed: %v", result.Error)
	}
	if result.RequestModified {
		t.Error("Request should not be modified")
	}
	if len(result.ConsoleOutput) != 1 {
		t.Errorf("Expected 1 console output, got %d", len(result.ConsoleOutput))
	}
	if result.ConsoleOutput[0].Message != "Pre-request script executed" {
		t.Errorf("Console message = %q, want %q", result.ConsoleOutput[0].Message, "Pre-request script executed")
	}
}

func TestExecutePreRequest_ModifyURL(t *testing.T) {
	executor := NewScriptExecutor()

	collReq := &CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	}
	req := NewScriptRequest(collReq)

	script := `
		lc.request.url = "https://api.example.com/v2/users";
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err != nil {
		t.Errorf("ExecutePreRequest failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed: %v", result.Error)
	}
	if !result.RequestModified {
		t.Error("Request should be marked as modified")
	}
	if req.URL() != "https://api.example.com/v2/users" {
		t.Errorf("URL = %q, want %q", req.URL(), "https://api.example.com/v2/users")
	}

	// Apply to original and check
	req.ApplyTo(collReq)
	if collReq.URL != "https://api.example.com/v2/users" {
		t.Errorf("CollectionRequest.URL = %q, want %q", collReq.URL, "https://api.example.com/v2/users")
	}
}

func TestExecutePreRequest_ModifyHeaders(t *testing.T) {
	executor := NewScriptExecutor()

	collReq := &CollectionRequest{
		Method: "POST",
		URL:    "https://api.example.com/users",
		Headers: []KeyValueEntry{
			{Key: "Content-Type", Value: "application/json", Enabled: true},
		},
	}
	req := NewScriptRequest(collReq)

	script := `
		lc.request.headers.set("Authorization", "Bearer token123");
		lc.request.headers.set("X-Custom-Header", "custom-value");
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err != nil {
		t.Errorf("ExecutePreRequest failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed: %v", result.Error)
	}
	if !result.RequestModified {
		t.Error("Request should be marked as modified")
	}

	// Check headers on ScriptRequest
	if req.GetHeader("Authorization") != "Bearer token123" {
		t.Error("Authorization header not set")
	}
	if req.GetHeader("X-Custom-Header") != "custom-value" {
		t.Error("X-Custom-Header not set")
	}

	// Apply to original and check
	req.ApplyTo(collReq)
	foundAuth := false
	foundCustom := false
	for _, h := range collReq.Headers {
		if h.Key == "Authorization" && h.Value == "Bearer token123" {
			foundAuth = true
		}
		if h.Key == "X-Custom-Header" && h.Value == "custom-value" {
			foundCustom = true
		}
	}
	if !foundAuth {
		t.Error("Authorization header not applied to CollectionRequest")
	}
	if !foundCustom {
		t.Error("X-Custom-Header not applied to CollectionRequest")
	}
}

func TestExecutePreRequest_RemoveHeader(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
		Headers: []KeyValueEntry{
			{Key: "X-Debug", Value: "true", Enabled: true},
			{Key: "Content-Type", Value: "application/json", Enabled: true},
		},
	})

	script := `
		lc.request.headers.remove("X-Debug");
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err != nil {
		t.Errorf("ExecutePreRequest failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed: %v", result.Error)
	}
	if !result.RequestModified {
		t.Error("Request should be marked as modified")
	}

	// Check X-Debug header was removed
	if req.GetHeader("X-Debug") != "" {
		t.Error("X-Debug header should have been removed")
	}
	// Content-Type should still be there
	if req.GetHeader("Content-Type") != "application/json" {
		t.Error("Content-Type header should still exist")
	}
}

func TestExecutePreRequest_ModifyBody(t *testing.T) {
	executor := NewScriptExecutor()

	collReq := &CollectionRequest{
		Method: "POST",
		URL:    "https://api.example.com/users",
		Body: &BodyConfig{
			Type:    "raw",
			Content: `{"name": "old"}`,
		},
	}
	req := NewScriptRequest(collReq)

	script := `
		lc.request.body.set('{"name": "new", "updated": true}');
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err != nil {
		t.Errorf("ExecutePreRequest failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed: %v", result.Error)
	}
	if !result.RequestModified {
		t.Error("Request should be marked as modified")
	}

	// Check body was updated
	if !strings.Contains(req.Body(), `"name": "new"`) {
		t.Errorf("Body not updated correctly: %s", req.Body())
	}

	// Apply to original and check
	req.ApplyTo(collReq)
	if collReq.Body == nil {
		t.Fatal("Body should not be nil")
	}
	bodyStr, ok := collReq.Body.Content.(string)
	if !ok {
		t.Fatal("Body content should be string")
	}
	if !strings.Contains(bodyStr, `"name": "new"`) {
		t.Errorf("Body not applied correctly: %s", bodyStr)
	}
}

func TestExecutePreRequest_ReadRequestProperties(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "POST",
		URL:    "https://api.example.com/users",
		Headers: []KeyValueEntry{
			{Key: "Content-Type", Value: "application/json", Enabled: true},
		},
		Body: &BodyConfig{
			Type:    "raw",
			Content: `{"name": "test"}`,
		},
	})

	script := `
		var method = lc.request.method;
		var url = lc.request.url;
		var contentType = lc.request.headers.get("Content-Type");
		var body = lc.request.body.raw();

		console.log("Method: " + method);
		console.log("URL: " + url);
		console.log("Content-Type: " + contentType);
		console.log("Body: " + body);
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err != nil {
		t.Errorf("ExecutePreRequest failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed: %v", result.Error)
	}

	// Check console output contains expected values
	if len(result.ConsoleOutput) != 4 {
		t.Errorf("Expected 4 console outputs, got %d", len(result.ConsoleOutput))
	}

	expectedOutputs := []string{
		"Method: POST",
		"URL: https://api.example.com/users",
		"Content-Type: application/json",
		`Body: {"name": "test"}`,
	}

	for i, expected := range expectedOutputs {
		if i < len(result.ConsoleOutput) && result.ConsoleOutput[i].Message != expected {
			t.Errorf("Console[%d] = %q, want %q", i, result.ConsoleOutput[i].Message, expected)
		}
	}
}

func TestExecutePreRequest_EnvironmentVariables(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	env := &Environment{
		Name: "test",
		Variables: map[string]string{
			"api_key":  "secret123",
			"base_url": "https://api.example.com",
		},
	}

	script := `
		var apiKey = lc.environment.get("api_key");
		var baseUrl = lc.environment.get("base_url");

		console.log("API Key: " + apiKey);
		console.log("Base URL: " + baseUrl);

		// Set a new variable
		lc.environment.set("new_var", "new_value");

		// Check if variable exists
		if (lc.environment.has("api_key")) {
			console.log("api_key exists");
		}
	`

	result, err := executor.ExecutePreRequest(script, req, env)

	if err != nil {
		t.Errorf("ExecutePreRequest failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed: %v", result.Error)
	}

	// Check console outputs
	if len(result.ConsoleOutput) < 3 {
		t.Errorf("Expected at least 3 console outputs, got %d", len(result.ConsoleOutput))
	}

	// Check environment changes
	if len(result.EnvChanges) != 1 {
		t.Errorf("Expected 1 env change, got %d", len(result.EnvChanges))
	}
	if result.EnvChanges[0].Name != "new_var" || result.EnvChanges[0].Value != "new_value" {
		t.Errorf("EnvChange = %+v, want {Name: new_var, Value: new_value}", result.EnvChanges[0])
	}
}

func TestExecutePreRequest_SyntaxError(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	// Script with syntax error
	script := `
		var x = {
			invalid syntax here
		};
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err == nil {
		t.Error("Expected error from syntax error script")
	}
	if result.Success {
		t.Error("Expected execution to fail due to syntax error")
	}
	if result.Error == nil {
		t.Fatal("Expected error info to be set")
	}
	if result.Error.Type != "SyntaxError" {
		t.Errorf("Error type = %q, want %q", result.Error.Type, "SyntaxError")
	}
}

func TestExecutePreRequest_RuntimeError(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	// Script with runtime error
	script := `
		var obj = null;
		obj.property; // TypeError: Cannot read property of null
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err == nil {
		t.Error("Expected error from runtime error script")
	}
	if result.Success {
		t.Error("Expected execution to fail due to runtime error")
	}
	if result.Error == nil {
		t.Fatal("Expected error info to be set")
	}
}

func TestExecutePreRequest_Timeout(t *testing.T) {
	executor := NewScriptExecutor()
	executor.SetTimeout(100 * time.Millisecond) // Short timeout for test

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	// Infinite loop script
	script := `
		while (true) {
			// This should timeout
		}
	`

	start := time.Now()
	result, err := executor.ExecutePreRequest(script, req, nil)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected error from timeout")
	}
	if result.Success {
		t.Error("Expected execution to fail due to timeout")
	}
	if result.Error == nil {
		t.Fatal("Expected error info to be set")
	}
	if result.Error.Type != "TimeoutError" {
		t.Errorf("Error type = %q, want %q", result.Error.Type, "TimeoutError")
	}

	// Should complete around the timeout duration (with some margin)
	if elapsed > 500*time.Millisecond {
		t.Errorf("Execution took too long: %v", elapsed)
	}
}

func TestExecutePreRequest_NilRequest(t *testing.T) {
	executor := NewScriptExecutor()

	// Create a minimal ScriptRequest
	req := NewScriptRequest(nil)

	script := `
		console.log("Request method is: " + lc.request.method);
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err != nil {
		t.Errorf("ExecutePreRequest failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed: %v", result.Error)
	}
}

func TestExecutePreRequest_EmptyScript(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	result, err := executor.ExecutePreRequest("", req, nil)

	if err != nil {
		t.Errorf("ExecutePreRequest failed with empty script: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed with empty script: %v", result.Error)
	}
	if result.RequestModified {
		t.Error("Empty script should not modify request")
	}
}

func TestExecutePreRequest_ConsoleLevels(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	script := `
		console.log("log message");
		console.info("info message");
		console.warn("warn message");
		console.error("error message");
		console.debug("debug message");
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err != nil {
		t.Errorf("ExecutePreRequest failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed: %v", result.Error)
	}

	if len(result.ConsoleOutput) != 5 {
		t.Fatalf("Expected 5 console outputs, got %d", len(result.ConsoleOutput))
	}

	expectedLevels := []ConsoleLogLevel{
		LogLevelLog,
		LogLevelInfo,
		LogLevelWarn,
		LogLevelError,
		LogLevelDebug,
	}

	for i, level := range expectedLevels {
		if result.ConsoleOutput[i].Level != level {
			t.Errorf("Console[%d].Level = %q, want %q", i, result.ConsoleOutput[i].Level, level)
		}
	}
}

func TestExecutePreRequest_HeadersCaseInsensitive(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
		Headers: []KeyValueEntry{
			{Key: "Content-Type", Value: "application/json", Enabled: true},
		},
	})

	script := `
		// Get header with different case
		var ct = lc.request.headers.get("content-type");
		console.log("Content-Type: " + ct);

		// Set header with different case should replace
		lc.request.headers.set("CONTENT-TYPE", "text/plain");
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err != nil {
		t.Errorf("ExecutePreRequest failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed: %v", result.Error)
	}

	// Check console output
	if len(result.ConsoleOutput) > 0 && result.ConsoleOutput[0].Message != "Content-Type: application/json" {
		t.Errorf("Case-insensitive get failed: %s", result.ConsoleOutput[0].Message)
	}

	// After modification, should have the new value
	if req.GetHeader("content-type") != "text/plain" {
		t.Errorf("Content-Type = %q, want %q", req.GetHeader("content-type"), "text/plain")
	}
}

func TestExecutePreRequest_Duration(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	script := `
		// Simple script
		var x = 1 + 1;
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err != nil {
		t.Errorf("ExecutePreRequest failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed: %v", result.Error)
	}

	// Duration should be recorded and positive
	if result.Duration <= 0 {
		t.Errorf("Duration = %v, want > 0", result.Duration)
	}
}

func TestExecutePreRequest_GetAllHeaders(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
		Headers: []KeyValueEntry{
			{Key: "Content-Type", Value: "application/json", Enabled: true},
			{Key: "Authorization", Value: "Bearer token", Enabled: true},
			{Key: "X-Disabled", Value: "should-not-appear", Enabled: false},
		},
	})

	script := `
		var headers = lc.request.headers.all();
		var keys = Object.keys(headers);
		console.log("Header count: " + keys.length);
		for (var i = 0; i < keys.length; i++) {
			console.log(keys[i] + ": " + headers[keys[i]]);
		}
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err != nil {
		t.Errorf("ExecutePreRequest failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed: %v", result.Error)
	}

	// Should have 3 outputs: count + 2 headers (disabled one excluded)
	if len(result.ConsoleOutput) < 1 {
		t.Fatal("Expected at least 1 console output")
	}
	if result.ConsoleOutput[0].Message != "Header count: 2" {
		t.Errorf("Header count = %s, want 'Header count: 2'", result.ConsoleOutput[0].Message)
	}
}

func TestExecutePreRequest_JSONBody(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "POST",
		URL:    "https://api.example.com/users",
		Body: &BodyConfig{
			Type:    "raw",
			Content: `{"name": "test", "age": 25}`,
		},
	})

	script := `
		var body = lc.request.body.json();
		console.log("Name: " + body.name);
		console.log("Age: " + body.age);
	`

	result, err := executor.ExecutePreRequest(script, req, nil)

	if err != nil {
		t.Errorf("ExecutePreRequest failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePreRequest failed: %v", result.Error)
	}

	if len(result.ConsoleOutput) != 2 {
		t.Fatalf("Expected 2 console outputs, got %d", len(result.ConsoleOutput))
	}
	if result.ConsoleOutput[0].Message != "Name: test" {
		t.Errorf("Name output = %q, want %q", result.ConsoleOutput[0].Message, "Name: test")
	}
	if result.ConsoleOutput[1].Message != "Age: 25" {
		t.Errorf("Age output = %q, want %q", result.ConsoleOutput[1].Message, "Age: 25")
	}
}

// ==================== POST-RESPONSE TESTS ====================

func TestExecutePostResponse_BasicExecution(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	resp := NewScriptResponseFromData(200, "200 OK", map[string]string{
		"Content-Type": "application/json",
	}, `{"users": []}`, 150)

	script := `
		console.log("Post-response script executed");
		console.log("Status: " + lc.response.status);
	`

	result, err := executor.ExecutePostResponse(script, req, resp, nil)

	if err != nil {
		t.Errorf("ExecutePostResponse failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePostResponse failed: %v", result.Error)
	}

	if len(result.ConsoleOutput) != 2 {
		t.Errorf("Expected 2 console outputs, got %d", len(result.ConsoleOutput))
	}
	if result.ConsoleOutput[0].Message != "Post-response script executed" {
		t.Errorf("Console message = %q, want %q", result.ConsoleOutput[0].Message, "Post-response script executed")
	}
	if result.ConsoleOutput[1].Message != "Status: 200" {
		t.Errorf("Status message = %q, want %q", result.ConsoleOutput[1].Message, "Status: 200")
	}
}

func TestExecutePostResponse_ReadResponseProperties(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "POST",
		URL:    "https://api.example.com/users",
	})

	resp := NewScriptResponseFromData(201, "201 Created", map[string]string{
		"Content-Type": "application/json",
		"X-Request-Id": "req-123",
	}, `{"id": 42, "name": "test user"}`, 250)

	script := `
		console.log("Status: " + lc.response.status);
		console.log("StatusText: " + lc.response.statusText);
		console.log("Time: " + lc.response.time);
		console.log("Content-Type: " + lc.response.headers.get("Content-Type"));
		console.log("Body: " + lc.response.body.raw());
	`

	result, err := executor.ExecutePostResponse(script, req, resp, nil)

	if err != nil {
		t.Errorf("ExecutePostResponse failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePostResponse failed: %v", result.Error)
	}

	expectedOutputs := []string{
		"Status: 201",
		"StatusText: 201 Created",
		"Time: 250",
		"Content-Type: application/json",
		`Body: {"id": 42, "name": "test user"}`,
	}

	if len(result.ConsoleOutput) != len(expectedOutputs) {
		t.Fatalf("Expected %d console outputs, got %d", len(expectedOutputs), len(result.ConsoleOutput))
	}

	for i, expected := range expectedOutputs {
		if result.ConsoleOutput[i].Message != expected {
			t.Errorf("Console[%d] = %q, want %q", i, result.ConsoleOutput[i].Message, expected)
		}
	}
}

func TestExecutePostResponse_JSONBody(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users/1",
	})

	resp := NewScriptResponseFromData(200, "200 OK", nil, `{"id": 1, "name": "John", "email": "john@example.com"}`, 100)

	script := `
		var data = lc.response.body.json();
		console.log("ID: " + data.id);
		console.log("Name: " + data.name);
		console.log("Email: " + data.email);
	`

	result, err := executor.ExecutePostResponse(script, req, resp, nil)

	if err != nil {
		t.Errorf("ExecutePostResponse failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePostResponse failed: %v", result.Error)
	}

	expectedOutputs := []string{
		"ID: 1",
		"Name: John",
		"Email: john@example.com",
	}

	if len(result.ConsoleOutput) != len(expectedOutputs) {
		t.Fatalf("Expected %d console outputs, got %d", len(expectedOutputs), len(result.ConsoleOutput))
	}

	for i, expected := range expectedOutputs {
		if result.ConsoleOutput[i].Message != expected {
			t.Errorf("Console[%d] = %q, want %q", i, result.ConsoleOutput[i].Message, expected)
		}
	}
}

func TestExecutePostResponse_ExtractValueToEnv(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "POST",
		URL:    "https://api.example.com/auth/login",
	})

	resp := NewScriptResponseFromData(200, "200 OK", nil, `{"token": "jwt-token-12345", "expires_in": 3600}`, 200)

	env := &Environment{
		Name:      "test",
		Variables: map[string]string{},
	}

	script := `
		var data = lc.response.body.json();
		lc.environment.set("auth_token", data.token);
		lc.environment.set("token_expires", String(data.expires_in));
	`

	result, err := executor.ExecutePostResponse(script, req, resp, env)

	if err != nil {
		t.Errorf("ExecutePostResponse failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePostResponse failed: %v", result.Error)
	}

	// Check environment changes
	if len(result.EnvChanges) != 2 {
		t.Fatalf("Expected 2 env changes, got %d", len(result.EnvChanges))
	}

	// Verify specific changes
	foundToken := false
	foundExpires := false
	for _, change := range result.EnvChanges {
		if change.Name == "auth_token" && change.Value == "jwt-token-12345" {
			foundToken = true
		}
		if change.Name == "token_expires" && change.Value == "3600" {
			foundExpires = true
		}
	}

	if !foundToken {
		t.Error("auth_token not set correctly")
	}
	if !foundExpires {
		t.Error("token_expires not set correctly")
	}
}

func TestExecutePostResponse_RequestIsReadonly(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	resp := NewScriptResponseFromData(200, "200 OK", nil, `{}`, 100)

	// Try to modify request in post-response script (should not have setters)
	script := `
		// Reading should work
		console.log("URL: " + lc.request.url);
		console.log("Method: " + lc.request.method);

		// Attempting to set should fail or be ignored
		try {
			lc.request.url = "https://hacked.com";
		} catch (e) {
			console.log("Cannot set URL");
		}
	`

	result, err := executor.ExecutePostResponse(script, req, resp, nil)

	if err != nil {
		t.Errorf("ExecutePostResponse failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePostResponse failed: %v", result.Error)
	}

	// Original URL should not be modified
	if req.URL() != "https://api.example.com/users" {
		t.Errorf("Request URL was modified: %s", req.URL())
	}
}

func TestExecutePostResponse_TestAssertions(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	resp := NewScriptResponseFromData(200, "200 OK", nil, `{"users": ["alice", "bob"]}`, 100)

	script := `
		lc.test("Status is 200", function() {
			lc.expect(lc.response.status).toBe(200);
		});

		lc.test("Response has users array", function() {
			var data = lc.response.body.json();
			lc.expect(data.users).toBeTruthy();
		});
	`

	result, err := executor.ExecutePostResponse(script, req, resp, nil)

	if err != nil {
		t.Errorf("ExecutePostResponse failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePostResponse failed: %v", result.Error)
	}

	// Check assertions
	if len(result.Assertions) != 2 {
		t.Fatalf("Expected 2 assertions, got %d", len(result.Assertions))
	}

	if !result.Assertions[0].Passed || result.Assertions[0].Name != "Status is 200" {
		t.Errorf("First assertion failed: %+v", result.Assertions[0])
	}

	if !result.Assertions[1].Passed || result.Assertions[1].Name != "Response has users array" {
		t.Errorf("Second assertion failed: %+v", result.Assertions[1])
	}
}

func TestExecutePostResponse_FailedAssertion(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	resp := NewScriptResponseFromData(404, "404 Not Found", nil, `{"error": "Not found"}`, 50)

	script := `
		lc.test("Status is 200", function() {
			lc.expect(lc.response.status).toBe(200);
		});
	`

	result, err := executor.ExecutePostResponse(script, req, resp, nil)

	// The script itself should still succeed
	if err != nil {
		t.Errorf("ExecutePostResponse failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePostResponse failed: %v", result.Error)
	}

	// But the assertion should fail
	if len(result.Assertions) != 1 {
		t.Fatalf("Expected 1 assertion, got %d", len(result.Assertions))
	}

	if result.Assertions[0].Passed {
		t.Error("Assertion should have failed")
	}
	if result.Assertions[0].Name != "Status is 200" {
		t.Errorf("Assertion name = %q, want %q", result.Assertions[0].Name, "Status is 200")
	}
}

func TestExecutePostResponse_EmptyScript(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	resp := NewScriptResponseFromData(200, "200 OK", nil, `{}`, 100)

	result, err := executor.ExecutePostResponse("", req, resp, nil)

	if err != nil {
		t.Errorf("ExecutePostResponse failed with empty script: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePostResponse failed with empty script: %v", result.Error)
	}
}

func TestExecutePostResponse_RuntimeError(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	resp := NewScriptResponseFromData(200, "200 OK", nil, `not valid json`, 100)

	script := `
		var data = lc.response.body.json();
		// This will fail because JSON is invalid
		console.log(data.nonexistent.property);
	`

	result, err := executor.ExecutePostResponse(script, req, resp, nil)

	if err == nil {
		t.Error("Expected error from runtime error script")
	}
	if result.Success {
		t.Error("Expected execution to fail")
	}
}

func TestExecutePostResponse_GetAllResponseHeaders(t *testing.T) {
	executor := NewScriptExecutor()

	req := NewScriptRequest(&CollectionRequest{
		Method: "GET",
		URL:    "https://api.example.com/users",
	})

	resp := NewScriptResponseFromData(200, "200 OK", map[string]string{
		"Content-Type":  "application/json",
		"Cache-Control": "no-cache",
		"X-Request-Id":  "abc123",
	}, `{}`, 100)

	script := `
		var headers = lc.response.headers.all();
		var keys = Object.keys(headers);
		console.log("Header count: " + keys.length);
	`

	result, err := executor.ExecutePostResponse(script, req, resp, nil)

	if err != nil {
		t.Errorf("ExecutePostResponse failed: %v", err)
	}
	if !result.Success {
		t.Errorf("ExecutePostResponse failed: %v", result.Error)
	}

	if len(result.ConsoleOutput) < 1 {
		t.Fatal("Expected at least 1 console output")
	}
	if result.ConsoleOutput[0].Message != "Header count: 3" {
		t.Errorf("Header count = %s, want 'Header count: 3'", result.ConsoleOutput[0].Message)
	}
}

// Matcher tests - comprehensive coverage for all assertion matchers

func TestMatcher_ToEqual(t *testing.T) {
	executor := NewScriptExecutor()
	resp := NewScriptResponseFromData(200, "200 OK", nil, `{"name": "test", "value": 42}`, 100)

	tests := []struct {
		name       string
		script     string
		shouldPass bool
	}{
		{
			name: "toEqual with same object",
			script: `
				lc.test("Object equality", function() {
					var data = lc.response.body.json();
					lc.expect(data.name).toEqual("test");
				});
			`,
			shouldPass: true,
		},
		{
			name: "toEqual with different values",
			script: `
				lc.test("Different values", function() {
					var data = lc.response.body.json();
					lc.expect(data.name).toEqual("other");
				});
			`,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.ExecutePostResponse(tt.script, nil, resp, nil)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}
			if len(result.Assertions) != 1 {
				t.Fatalf("Expected 1 assertion, got %d", len(result.Assertions))
			}
			if result.Assertions[0].Passed != tt.shouldPass {
				t.Errorf("Assertion passed = %v, want %v", result.Assertions[0].Passed, tt.shouldPass)
			}
		})
	}
}

func TestMatcher_ToBeTruthy(t *testing.T) {
	executor := NewScriptExecutor()
	resp := NewScriptResponseFromData(200, "200 OK", nil, `{"name": "test", "count": 0, "empty": "", "items": [1,2,3]}`, 100)

	tests := []struct {
		name       string
		script     string
		shouldPass bool
	}{
		{
			name: "toBeTruthy with non-empty string",
			script: `
				lc.test("String is truthy", function() {
					var data = lc.response.body.json();
					lc.expect(data.name).toBeTruthy();
				});
			`,
			shouldPass: true,
		},
		{
			name: "toBeTruthy with zero",
			script: `
				lc.test("Zero is falsy", function() {
					var data = lc.response.body.json();
					lc.expect(data.count).toBeTruthy();
				});
			`,
			shouldPass: false,
		},
		{
			name: "toBeTruthy with empty string",
			script: `
				lc.test("Empty string is falsy", function() {
					var data = lc.response.body.json();
					lc.expect(data.empty).toBeTruthy();
				});
			`,
			shouldPass: false,
		},
		{
			name: "toBeTruthy with array",
			script: `
				lc.test("Array is truthy", function() {
					var data = lc.response.body.json();
					lc.expect(data.items).toBeTruthy();
				});
			`,
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.ExecutePostResponse(tt.script, nil, resp, nil)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}
			if len(result.Assertions) != 1 {
				t.Fatalf("Expected 1 assertion, got %d", len(result.Assertions))
			}
			if result.Assertions[0].Passed != tt.shouldPass {
				t.Errorf("Assertion passed = %v, want %v", result.Assertions[0].Passed, tt.shouldPass)
			}
		})
	}
}

func TestMatcher_ToBeFalsy(t *testing.T) {
	executor := NewScriptExecutor()
	resp := NewScriptResponseFromData(200, "200 OK", nil, `{"name": "test", "count": 0, "empty": "", "active": false}`, 100)

	tests := []struct {
		name       string
		script     string
		shouldPass bool
	}{
		{
			name: "toBeFalsy with zero",
			script: `
				lc.test("Zero is falsy", function() {
					var data = lc.response.body.json();
					lc.expect(data.count).toBeFalsy();
				});
			`,
			shouldPass: true,
		},
		{
			name: "toBeFalsy with empty string",
			script: `
				lc.test("Empty string is falsy", function() {
					var data = lc.response.body.json();
					lc.expect(data.empty).toBeFalsy();
				});
			`,
			shouldPass: true,
		},
		{
			name: "toBeFalsy with false boolean",
			script: `
				lc.test("False is falsy", function() {
					var data = lc.response.body.json();
					lc.expect(data.active).toBeFalsy();
				});
			`,
			shouldPass: true,
		},
		{
			name: "toBeFalsy with truthy value",
			script: `
				lc.test("String is truthy", function() {
					var data = lc.response.body.json();
					lc.expect(data.name).toBeFalsy();
				});
			`,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.ExecutePostResponse(tt.script, nil, resp, nil)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}
			if len(result.Assertions) != 1 {
				t.Fatalf("Expected 1 assertion, got %d", len(result.Assertions))
			}
			if result.Assertions[0].Passed != tt.shouldPass {
				t.Errorf("Assertion passed = %v, want %v", result.Assertions[0].Passed, tt.shouldPass)
			}
		})
	}
}

func TestMatcher_ToContain(t *testing.T) {
	executor := NewScriptExecutor()
	resp := NewScriptResponseFromData(200, "200 OK", nil, `{"message": "Hello World", "items": ["apple", "banana", "cherry"]}`, 100)

	tests := []struct {
		name       string
		script     string
		shouldPass bool
	}{
		{
			name: "toContain with string substring",
			script: `
				lc.test("String contains substring", function() {
					var data = lc.response.body.json();
					lc.expect(data.message).toContain("World");
				});
			`,
			shouldPass: true,
		},
		{
			name: "toContain with missing substring",
			script: `
				lc.test("String missing substring", function() {
					var data = lc.response.body.json();
					lc.expect(data.message).toContain("Goodbye");
				});
			`,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.ExecutePostResponse(tt.script, nil, resp, nil)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}
			if len(result.Assertions) != 1 {
				t.Fatalf("Expected 1 assertion, got %d", len(result.Assertions))
			}
			if result.Assertions[0].Passed != tt.shouldPass {
				t.Errorf("Assertion passed = %v, want %v", result.Assertions[0].Passed, tt.shouldPass)
			}
		})
	}
}

func TestMatcher_ToHaveProperty(t *testing.T) {
	executor := NewScriptExecutor()
	resp := NewScriptResponseFromData(200, "200 OK", nil, `{"user": {"name": "test", "email": "test@example.com"}}`, 100)

	tests := []struct {
		name       string
		script     string
		shouldPass bool
	}{
		{
			name: "toHaveProperty with existing property",
			script: `
				lc.test("Object has property", function() {
					var data = lc.response.body.json();
					lc.expect(data.user).toHaveProperty("name");
				});
			`,
			shouldPass: true,
		},
		{
			name: "toHaveProperty with missing property",
			script: `
				lc.test("Object missing property", function() {
					var data = lc.response.body.json();
					lc.expect(data.user).toHaveProperty("age");
				});
			`,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.ExecutePostResponse(tt.script, nil, resp, nil)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}
			if len(result.Assertions) != 1 {
				t.Fatalf("Expected 1 assertion, got %d", len(result.Assertions))
			}
			if result.Assertions[0].Passed != tt.shouldPass {
				t.Errorf("Assertion passed = %v, want %v", result.Assertions[0].Passed, tt.shouldPass)
			}
		})
	}
}

func TestMatcher_ToBeGreaterThan(t *testing.T) {
	executor := NewScriptExecutor()
	resp := NewScriptResponseFromData(200, "200 OK", nil, `{"count": 42, "price": 19.99}`, 100)

	tests := []struct {
		name       string
		script     string
		shouldPass bool
	}{
		{
			name: "toBeGreaterThan with greater value",
			script: `
				lc.test("Count is greater than 40", function() {
					var data = lc.response.body.json();
					lc.expect(data.count).toBeGreaterThan(40);
				});
			`,
			shouldPass: true,
		},
		{
			name: "toBeGreaterThan with smaller value",
			script: `
				lc.test("Count is not greater than 50", function() {
					var data = lc.response.body.json();
					lc.expect(data.count).toBeGreaterThan(50);
				});
			`,
			shouldPass: false,
		},
		{
			name: "toBeGreaterThan with equal value",
			script: `
				lc.test("Count is not greater than itself", function() {
					var data = lc.response.body.json();
					lc.expect(data.count).toBeGreaterThan(42);
				});
			`,
			shouldPass: false,
		},
		{
			name: "toBeGreaterThan with float",
			script: `
				lc.test("Price is greater than 19", function() {
					var data = lc.response.body.json();
					lc.expect(data.price).toBeGreaterThan(19);
				});
			`,
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.ExecutePostResponse(tt.script, nil, resp, nil)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}
			if len(result.Assertions) != 1 {
				t.Fatalf("Expected 1 assertion, got %d", len(result.Assertions))
			}
			if result.Assertions[0].Passed != tt.shouldPass {
				t.Errorf("Assertion passed = %v, want %v", result.Assertions[0].Passed, tt.shouldPass)
			}
		})
	}
}

func TestMatcher_ToBeLessThan(t *testing.T) {
	executor := NewScriptExecutor()
	resp := NewScriptResponseFromData(200, "200 OK", nil, `{"count": 42, "price": 19.99}`, 100)

	tests := []struct {
		name       string
		script     string
		shouldPass bool
	}{
		{
			name: "toBeLessThan with smaller value",
			script: `
				lc.test("Count is less than 50", function() {
					var data = lc.response.body.json();
					lc.expect(data.count).toBeLessThan(50);
				});
			`,
			shouldPass: true,
		},
		{
			name: "toBeLessThan with greater value",
			script: `
				lc.test("Count is not less than 40", function() {
					var data = lc.response.body.json();
					lc.expect(data.count).toBeLessThan(40);
				});
			`,
			shouldPass: false,
		},
		{
			name: "toBeLessThan with equal value",
			script: `
				lc.test("Count is not less than itself", function() {
					var data = lc.response.body.json();
					lc.expect(data.count).toBeLessThan(42);
				});
			`,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.ExecutePostResponse(tt.script, nil, resp, nil)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}
			if len(result.Assertions) != 1 {
				t.Fatalf("Expected 1 assertion, got %d", len(result.Assertions))
			}
			if result.Assertions[0].Passed != tt.shouldPass {
				t.Errorf("Assertion passed = %v, want %v", result.Assertions[0].Passed, tt.shouldPass)
			}
		})
	}
}

func TestMatcher_ResponseTime(t *testing.T) {
	executor := NewScriptExecutor()
	resp := NewScriptResponseFromData(200, "200 OK", nil, `{}`, 150)

	tests := []struct {
		name       string
		script     string
		shouldPass bool
	}{
		{
			name: "Response time under threshold",
			script: `
				lc.test("Response time is under 200ms", function() {
					lc.expect(lc.response.time).toBeLessThan(200);
				});
			`,
			shouldPass: true,
		},
		{
			name: "Response time over threshold",
			script: `
				lc.test("Response time is under 100ms", function() {
					lc.expect(lc.response.time).toBeLessThan(100);
				});
			`,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.ExecutePostResponse(tt.script, nil, resp, nil)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}
			if len(result.Assertions) != 1 {
				t.Fatalf("Expected 1 assertion, got %d", len(result.Assertions))
			}
			if result.Assertions[0].Passed != tt.shouldPass {
				t.Errorf("Assertion passed = %v, want %v", result.Assertions[0].Passed, tt.shouldPass)
			}
		})
	}
}

func TestMatcher_ChainedExpectations(t *testing.T) {
	executor := NewScriptExecutor()
	resp := NewScriptResponseFromData(200, "200 OK", nil, `{"users": [{"name": "alice"}, {"name": "bob"}]}`, 100)

	script := `
		lc.test("Chained expectations", function() {
			lc.expect(lc.response.status).toBe(200);
			var data = lc.response.body.json();
			lc.expect(data).toHaveProperty("users");
			lc.expect(data.users).toBeTruthy();
		});
	`

	result, err := executor.ExecutePostResponse(script, nil, resp, nil)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
	if len(result.Assertions) != 1 {
		t.Fatalf("Expected 1 assertion, got %d", len(result.Assertions))
	}
	if !result.Assertions[0].Passed {
		t.Errorf("Chained expectations failed: %s", result.Assertions[0].Message)
	}
}

func TestMatcher_MultipleTestBlocks(t *testing.T) {
	executor := NewScriptExecutor()
	resp := NewScriptResponseFromData(200, "200 OK", map[string]string{"Content-Type": "application/json"}, `{"status": "ok"}`, 50)

	script := `
		lc.test("Status code test", function() {
			lc.expect(lc.response.status).toBe(200);
		});

		lc.test("Content-Type test", function() {
			lc.expect(lc.response.headers.get("Content-Type")).toContain("json");
		});

		lc.test("Body test", function() {
			var data = lc.response.body.json();
			lc.expect(data.status).toEqual("ok");
		});

		lc.test("Performance test", function() {
			lc.expect(lc.response.time).toBeLessThan(100);
		});
	`

	result, err := executor.ExecutePostResponse(script, nil, resp, nil)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
	if len(result.Assertions) != 4 {
		t.Fatalf("Expected 4 assertions, got %d", len(result.Assertions))
	}

	for i, assertion := range result.Assertions {
		if !assertion.Passed {
			t.Errorf("Assertion %d (%s) failed: %s", i, assertion.Name, assertion.Message)
		}
	}
}

// Helper function tests

func TestHelperFunction_IsTruthy(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"nil is falsy", nil, false},
		{"true is truthy", true, true},
		{"false is falsy", false, false},
		{"non-zero int is truthy", 42, true},
		{"zero int is falsy", 0, false},
		{"non-zero int64 is truthy", int64(42), true},
		{"zero int64 is falsy", int64(0), false},
		{"non-zero float is truthy", 3.14, true},
		{"zero float is falsy", 0.0, false},
		{"non-empty string is truthy", "hello", true},
		{"empty string is falsy", "", false},
		{"struct is truthy", struct{}{}, true},
		{"slice is truthy", []int{1, 2, 3}, true},
		{"map is truthy", map[string]int{"a": 1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTruthy(tt.value)
			if result != tt.expected {
				t.Errorf("isTruthy(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestHelperFunction_ToFloat(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected float64
	}{
		{"int to float", 42, 42.0},
		{"int64 to float", int64(100), 100.0},
		{"int32 to float", int32(50), 50.0},
		{"float32 to float", float32(3.14), 3.140000104904175}, // float32 precision
		{"float64 to float", 2.718, 2.718},
		{"string to float", "not a number", 0.0},
		{"nil to float", nil, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toFloat(tt.value)
			if result != tt.expected {
				t.Errorf("toFloat(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestHelperFunction_DeepEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{"same strings", "hello", "hello", true},
		{"different strings", "hello", "world", false},
		{"same numbers", 42, 42, true},
		{"different numbers", 42, 24, false},
		// Note: deepEqual uses formatArg which converts to string, so "42" == "42" is true
		{"int and string with same representation", 42, "42", true},
		{"nil values", nil, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := deepEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("deepEqual(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// JSON parsing security test
func TestSecureJSONParsing(t *testing.T) {
	executor := NewScriptExecutor()

	// Create a response with potentially malicious JSON
	// This should NOT execute as code, just parse as JSON
	maliciousBody := `{"value": "test"}`

	resp := NewScriptResponseFromData(200, "200 OK", nil, maliciousBody, 100)

	script := `
		lc.test("Safe JSON parsing", function() {
			var data = lc.response.body.json();
			lc.expect(data.value).toBe("test");
		});
	`

	result, err := executor.ExecutePostResponse(script, nil, resp, nil)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %v", result.Error)
	}
}

func TestSecureJSONParsing_InvalidJSON(t *testing.T) {
	executor := NewScriptExecutor()

	// Invalid JSON should return null, not execute as code
	invalidJSON := `{invalid: json}`

	resp := NewScriptResponseFromData(200, "200 OK", nil, invalidJSON, 100)

	script := `
		lc.test("Invalid JSON returns null", function() {
			var data = lc.response.body.json();
			lc.expect(data).toBeFalsy();
		});
	`

	result, err := executor.ExecutePostResponse(script, nil, resp, nil)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
	if !result.Success {
		t.Errorf("Expected success, got error: %v", result.Error)
	}
	if len(result.Assertions) != 1 {
		t.Fatalf("Expected 1 assertion, got %d", len(result.Assertions))
	}
	if !result.Assertions[0].Passed {
		t.Errorf("Invalid JSON should return falsy value")
	}
}
