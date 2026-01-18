package api

import (
	"testing"

	"github.com/dop251/goja"
)

func setupInfoVM(t *testing.T, info *ScriptInfo) *goja.Runtime {
	t.Helper()
	vm := goja.New()
	executor := &gojaExecutor{globals: NewScriptGlobals()}

	lc := vm.NewObject()
	if err := executor.setupLCInfo(vm, lc, info); err != nil {
		t.Fatalf("setupLCInfo failed: %v", err)
	}
	if err := vm.Set("lc", lc); err != nil {
		t.Fatalf("Failed to set lc: %v", err)
	}

	return vm
}

func TestScriptInfo_Defaults(t *testing.T) {
	info := NewScriptInfo()

	if info.Iteration != 1 {
		t.Errorf("Default iteration should be 1, got %d", info.Iteration)
	}
	if info.ScriptType != "" {
		t.Errorf("Default scriptType should be empty, got %s", info.ScriptType)
	}
}

func TestSetupLCInfo_ScriptType(t *testing.T) {
	tests := []struct {
		name       string
		scriptType string
		expected   string
	}{
		{
			name:       "pre-request",
			scriptType: "pre-request",
			expected:   "pre-request",
		},
		{
			name:       "post-response",
			scriptType: "post-response",
			expected:   "post-response",
		},
		{
			name:       "empty",
			scriptType: "",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ScriptInfo{ScriptType: tt.scriptType}
			vm := setupInfoVM(t, info)

			result, err := vm.RunString(`lc.info.scriptType`)
			if err != nil {
				t.Fatalf("Script failed: %v", err)
			}

			if result.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.String())
			}
		})
	}
}

func TestSetupLCInfo_RequestName(t *testing.T) {
	info := &ScriptInfo{RequestName: "Get Users"}
	vm := setupInfoVM(t, info)

	result, err := vm.RunString(`lc.info.requestName`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	if result.String() != "Get Users" {
		t.Errorf("Expected 'Get Users', got %s", result.String())
	}
}

func TestSetupLCInfo_RequestNameEmpty(t *testing.T) {
	info := &ScriptInfo{RequestName: ""}
	vm := setupInfoVM(t, info)

	result, err := vm.RunString(`lc.info.requestName`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	if !goja.IsUndefined(result) {
		t.Errorf("Expected undefined for empty requestName, got %v", result)
	}
}

func TestSetupLCInfo_RequestId(t *testing.T) {
	info := &ScriptInfo{RequestID: "req_123"}
	vm := setupInfoVM(t, info)

	result, err := vm.RunString(`lc.info.requestId`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	if result.String() != "req_123" {
		t.Errorf("Expected 'req_123', got %s", result.String())
	}
}

func TestSetupLCInfo_CollectionName(t *testing.T) {
	info := &ScriptInfo{CollectionName: "My API"}
	vm := setupInfoVM(t, info)

	result, err := vm.RunString(`lc.info.collectionName`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	if result.String() != "My API" {
		t.Errorf("Expected 'My API', got %s", result.String())
	}
}

func TestSetupLCInfo_CollectionNameEmpty(t *testing.T) {
	info := &ScriptInfo{CollectionName: ""}
	vm := setupInfoVM(t, info)

	result, err := vm.RunString(`lc.info.collectionName`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	if !goja.IsUndefined(result) {
		t.Errorf("Expected undefined for empty collectionName, got %v", result)
	}
}

func TestSetupLCInfo_EnvironmentName(t *testing.T) {
	info := &ScriptInfo{EnvironmentName: "production"}
	vm := setupInfoVM(t, info)

	result, err := vm.RunString(`lc.info.environmentName`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	if result.String() != "production" {
		t.Errorf("Expected 'production', got %s", result.String())
	}
}

func TestSetupLCInfo_Iteration(t *testing.T) {
	tests := []struct {
		name      string
		iteration int
		expected  int64
	}{
		{
			name:      "first iteration",
			iteration: 1,
			expected:  1,
		},
		{
			name:      "fifth iteration",
			iteration: 5,
			expected:  5,
		},
		{
			name:      "hundredth iteration",
			iteration: 100,
			expected:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ScriptInfo{Iteration: tt.iteration}
			vm := setupInfoVM(t, info)

			result, err := vm.RunString(`lc.info.iteration`)
			if err != nil {
				t.Fatalf("Script failed: %v", err)
			}

			if result.ToInteger() != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result.ToInteger())
			}
		})
	}
}

func TestSetupLCInfo_ReadOnly(t *testing.T) {
	info := &ScriptInfo{
		ScriptType:  "pre-request",
		RequestName: "Test",
		Iteration:   1,
	}
	vm := setupInfoVM(t, info)

	// Try to modify (should not change actual value due to read-only property)
	// Note: In strict mode this would throw, but in non-strict mode it silently fails
	_, err := vm.RunString(`
		lc.info.scriptType = "modified";
		lc.info.iteration = 999;
	`)
	if err != nil {
		// Expected to fail or be ignored
		t.Logf("Assignment attempt: %v", err)
	}

	// Verify values unchanged
	result, _ := vm.RunString(`lc.info.scriptType`)
	if result.String() != "pre-request" {
		t.Error("scriptType should be read-only")
	}

	result, _ = vm.RunString(`lc.info.iteration`)
	if result.ToInteger() != 1 {
		t.Error("iteration should be read-only")
	}
}

func TestSetupLCInfo_FullContext(t *testing.T) {
	info := &ScriptInfo{
		ScriptType:      "post-response",
		RequestName:     "Create User",
		RequestID:       "req_abc123",
		CollectionName:  "User API",
		EnvironmentName: "staging",
		Iteration:       3,
	}
	vm := setupInfoVM(t, info)

	script := `
		var ctx = {
			type: lc.info.scriptType,
			name: lc.info.requestName,
			id: lc.info.requestId,
			collection: lc.info.collectionName,
			env: lc.info.environmentName,
			iter: lc.info.iteration
		};
		JSON.stringify(ctx);
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	expected := `{"type":"post-response","name":"Create User","id":"req_abc123","collection":"User API","env":"staging","iter":3}`
	if result.String() != expected {
		t.Errorf("Expected %s\nGot %s", expected, result.String())
	}
}
