package api

import (
	"testing"

	"github.com/dop251/goja"
)

func TestSetupLCBase64(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		{
			name:     "encode simple text",
			script:   `lc.base64.encode("Hello, World!")`,
			expected: "SGVsbG8sIFdvcmxkIQ==",
		},
		{
			name:     "decode valid base64",
			script:   `lc.base64.decode("SGVsbG8sIFdvcmxkIQ==")`,
			expected: "Hello, World!",
		},
		{
			name:     "round-trip encode then decode",
			script:   `lc.base64.decode(lc.base64.encode("test string"))`,
			expected: "test string",
		},
		{
			name:     "encode empty string",
			script:   `lc.base64.encode("")`,
			expected: "",
		},
		{
			name:     "decode empty string",
			script:   `lc.base64.decode("")`,
			expected: "",
		},
		{
			name:     "decode invalid base64 returns empty",
			script:   `lc.base64.decode("not valid base64!!!")`,
			expected: "",
		},
		{
			name:     "encode unicode characters",
			script:   `lc.base64.encode("café ☕")`,
			expected: "Y2Fmw6kg4piV",
		},
		{
			name:     "btoa global function",
			script:   `btoa("Hello")`,
			expected: "SGVsbG8=",
		},
		{
			name:     "atob global function",
			script:   `atob("SGVsbG8=")`,
			expected: "Hello",
		},
		{
			name:     "btoa and atob round-trip",
			script:   `atob(btoa("round trip test"))`,
			expected: "round trip test",
		},
		{
			name:     "atob with invalid input",
			script:   `atob("invalid!!!base64")`,
			expected: "",
		},
		{
			name:     "encode with no arguments",
			script:   `lc.base64.encode()`,
			expected: "",
		},
		{
			name:     "decode with no arguments",
			script:   `lc.base64.decode()`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := goja.New()
			executor := &gojaExecutor{globals: NewScriptGlobals()}

			// Create lc object
			lc := vm.NewObject()
			err := executor.setupLCBase64(vm, lc)
			if err != nil {
				t.Fatalf("setupLCBase64 failed: %v", err)
			}
			if err := vm.Set("lc", lc); err != nil {
				t.Fatalf("Failed to set lc: %v", err)
			}

			// Execute script
			result, err := vm.RunString(tt.script)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}

			// Check result
			got := result.String()
			if got != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestBase64EncodingConsistency(t *testing.T) {
	// Verify that lc.base64.encode and btoa produce the same output
	testCases := []string{
		"Hello",
		"test123",
		"Special chars: !@#$%^&*()",
		"",
	}

	for _, input := range testCases {
		t.Run("consistency_"+input, func(t *testing.T) {
			vm := goja.New()
			executor := &gojaExecutor{globals: NewScriptGlobals()}

			lc := vm.NewObject()
			if err := executor.setupLCBase64(vm, lc); err != nil {
				t.Fatalf("setupLCBase64 failed: %v", err)
			}
			if err := vm.Set("lc", lc); err != nil {
				t.Fatalf("Failed to set lc: %v", err)
			}
			if err := vm.Set("testInput", input); err != nil {
				t.Fatalf("Failed to set testInput: %v", err)
			}

			// Compare both methods
			script := `
				var lcResult = lc.base64.encode(testInput);
				var btoaResult = btoa(testInput);
				lcResult === btoaResult ? "match" : "mismatch";
			`
			result, err := vm.RunString(script)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}

			if result.String() != "match" {
				t.Error("lc.base64.encode and btoa produced different results")
			}
		})
	}
}
