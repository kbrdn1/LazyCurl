package api

import (
	"testing"

	"github.com/dop251/goja"
)

func TestSetupLCCrypto(t *testing.T) {
	tests := []struct {
		name     string
		script   string
		expected string
	}{
		// MD5 tests
		{
			name:     "md5 of empty string",
			script:   `lc.crypto.md5("")`,
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:     "md5 of hello",
			script:   `lc.crypto.md5("hello")`,
			expected: "5d41402abc4b2a76b9719d911017c592",
		},
		{
			name:     "md5 with no arguments",
			script:   `lc.crypto.md5()`,
			expected: "",
		},
		// SHA1 tests
		{
			name:     "sha1 of empty string",
			script:   `lc.crypto.sha1("")`,
			expected: "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		},
		{
			name:     "sha1 of hello",
			script:   `lc.crypto.sha1("hello")`,
			expected: "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d",
		},
		{
			name:     "sha1 with no arguments",
			script:   `lc.crypto.sha1()`,
			expected: "",
		},
		// SHA256 tests
		{
			name:     "sha256 of empty string",
			script:   `lc.crypto.sha256("")`,
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "sha256 of hello",
			script:   `lc.crypto.sha256("hello")`,
			expected: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		},
		{
			name:     "sha256 with no arguments",
			script:   `lc.crypto.sha256()`,
			expected: "",
		},
		// SHA512 tests
		{
			name:     "sha512 of empty string",
			script:   `lc.crypto.sha512("")`,
			expected: "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
		},
		{
			name:     "sha512 of hello",
			script:   `lc.crypto.sha512("hello")`,
			expected: "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043",
		},
		{
			name:     "sha512 with no arguments",
			script:   `lc.crypto.sha512()`,
			expected: "",
		},
		// HMAC-SHA256 tests
		{
			name:     "hmacSha256 with key",
			script:   `lc.crypto.hmacSha256("message", "secret")`,
			expected: "8b5f48702995c1598c573db1e21866a9b825d4a794d169d7060a03605796360b",
		},
		{
			name:     "hmacSha256 empty message",
			script:   `lc.crypto.hmacSha256("", "secret")`,
			expected: "f9e66e179b6747ae54108f82f8ade8b3c25d76fd30afde6c395822c530196169",
		},
		{
			name:     "hmacSha256 with no arguments",
			script:   `lc.crypto.hmacSha256()`,
			expected: "",
		},
		{
			name:     "hmacSha256 with one argument",
			script:   `lc.crypto.hmacSha256("message")`,
			expected: "",
		},
		// HMAC-SHA512 tests
		{
			name:     "hmacSha512 with key",
			script:   `lc.crypto.hmacSha512("message", "secret")`,
			expected: "1bba587c730eedba31f53abb0b6ca589e09de4e894ee455e6140807399759adaafa069eec7c01647bb173dcb17f55d22af49a18071b748c5c2edd7f7a829c632",
		},
		{
			name:     "hmacSha512 empty message",
			script:   `lc.crypto.hmacSha512("", "secret")`,
			expected: "b0e9650c5faf9cd8ae02276671545424104589b3656731ec193b25d01b07561c27637c2d4d68389d6cf5007a8632c26ec89ba80a01c77a6cdd389ec28db43901",
		},
		{
			name:     "hmacSha512 with no arguments",
			script:   `lc.crypto.hmacSha512()`,
			expected: "",
		},
		// HMAC-SHA1 tests (for OAuth compatibility)
		{
			name:     "hmacSha1 with key",
			script:   `lc.crypto.hmacSha1("message", "secret")`,
			expected: "0caf649feee4953d87bf903ac1176c45e028df16",
		},
		{
			name:     "hmacSha1 empty message",
			script:   `lc.crypto.hmacSha1("", "secret")`,
			expected: "25af6174a0fcecc4d346680a72b7ce644b9a88e8",
		},
		// Unicode tests
		{
			name:     "sha256 of unicode",
			script:   `lc.crypto.sha256("café ☕")`,
			expected: "a7e46d54289812af2aa5b08c2fbab5d24bccfc6586df55b187272c8a2a31c85f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := goja.New()
			executor := &gojaExecutor{globals: NewScriptGlobals()}

			// Create lc object
			lc := vm.NewObject()
			err := executor.setupLCCrypto(vm, lc)
			if err != nil {
				t.Fatalf("setupLCCrypto failed: %v", err)
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

func TestCryptoHashDeterminism(t *testing.T) {
	// Verify that hashes produce the same output for the same input
	vm := goja.New()
	executor := &gojaExecutor{globals: NewScriptGlobals()}

	lc := vm.NewObject()
	if err := executor.setupLCCrypto(vm, lc); err != nil {
		t.Fatalf("setupLCCrypto failed: %v", err)
	}
	if err := vm.Set("lc", lc); err != nil {
		t.Fatalf("Failed to set lc: %v", err)
	}

	// Run same hash multiple times
	script := `
		var h1 = lc.crypto.sha256("test");
		var h2 = lc.crypto.sha256("test");
		var h3 = lc.crypto.sha256("test");
		h1 === h2 && h2 === h3 ? "deterministic" : "random";
	`
	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	if result.String() != "deterministic" {
		t.Error("Hash function is not deterministic")
	}
}

func TestHMACWithDifferentKeys(t *testing.T) {
	// Verify that HMAC produces different outputs with different keys
	vm := goja.New()
	executor := &gojaExecutor{globals: NewScriptGlobals()}

	lc := vm.NewObject()
	if err := executor.setupLCCrypto(vm, lc); err != nil {
		t.Fatalf("setupLCCrypto failed: %v", err)
	}
	if err := vm.Set("lc", lc); err != nil {
		t.Fatalf("Failed to set lc: %v", err)
	}

	script := `
		var h1 = lc.crypto.hmacSha256("message", "key1");
		var h2 = lc.crypto.hmacSha256("message", "key2");
		h1 !== h2 ? "different" : "same";
	`
	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	if result.String() != "different" {
		t.Error("HMAC should produce different results with different keys")
	}
}
