package api

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/dop251/goja"
)

func setupVariablesVM(t *testing.T) *goja.Runtime {
	t.Helper()
	vm := goja.New()
	executor := &gojaExecutor{globals: NewScriptGlobals()}

	lc := vm.NewObject()
	if err := executor.setupLCVariables(vm, lc); err != nil {
		t.Fatalf("setupLCVariables failed: %v", err)
	}
	if err := vm.Set("lc", lc); err != nil {
		t.Fatalf("Failed to set lc: %v", err)
	}

	return vm
}

func TestVariablesUUID(t *testing.T) {
	vm := setupVariablesVM(t)

	// Test UUID format (v4)
	result, err := vm.RunString(`lc.variables.uuid()`)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	uuid := result.String()

	// UUID v4 format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(uuid) {
		t.Errorf("UUID format invalid: %s", uuid)
	}

	// Test uniqueness
	result2, _ := vm.RunString(`lc.variables.uuid()`)
	if uuid == result2.String() {
		t.Error("Two consecutive UUIDs should be different")
	}
}

func TestVariablesTimestamp(t *testing.T) {
	vm := setupVariablesVM(t)

	// Test timestamp is reasonable
	before := time.Now().Unix()
	result, err := vm.RunString(`lc.variables.timestamp()`)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
	after := time.Now().Unix()

	ts := result.ToInteger()
	if ts < before || ts > after {
		t.Errorf("Timestamp %d not in expected range [%d, %d]", ts, before, after)
	}
}

func TestVariablesTimestampMs(t *testing.T) {
	vm := setupVariablesVM(t)

	before := time.Now().UnixMilli()
	result, err := vm.RunString(`lc.variables.timestampMs()`)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
	after := time.Now().UnixMilli()

	ts := result.ToInteger()
	if ts < before || ts > after {
		t.Errorf("TimestampMs %d not in expected range [%d, %d]", ts, before, after)
	}
}

func TestVariablesIsoTimestamp(t *testing.T) {
	vm := setupVariablesVM(t)

	result, err := vm.RunString(`lc.variables.isoTimestamp()`)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	iso := result.String()

	// Should parse as RFC3339
	_, parseErr := time.Parse(time.RFC3339, iso)
	if parseErr != nil {
		t.Errorf("isoTimestamp not in RFC3339 format: %s, error: %v", iso, parseErr)
	}

	// Should end with Z (UTC)
	if !strings.HasSuffix(iso, "Z") {
		t.Errorf("isoTimestamp should be UTC (end with Z): %s", iso)
	}
}

func TestVariablesRandomInt(t *testing.T) {
	vm := setupVariablesVM(t)

	tests := []struct {
		name   string
		script string
		minVal int64
		maxVal int64
	}{
		{
			name:   "default range",
			script: `lc.variables.randomInt()`,
			minVal: 0,
			maxVal: 100,
		},
		{
			name:   "custom range",
			script: `lc.variables.randomInt(10, 20)`,
			minVal: 10,
			maxVal: 20,
		},
		{
			name:   "single value min only",
			script: `lc.variables.randomInt(5)`,
			minVal: 5,
			maxVal: 100,
		},
		{
			name:   "same min and max",
			script: `lc.variables.randomInt(42, 42)`,
			minVal: 42,
			maxVal: 42,
		},
		{
			name:   "reversed range (auto-swaps)",
			script: `lc.variables.randomInt(100, 50)`,
			minVal: 50,
			maxVal: 100,
		},
		{
			name:   "negative range",
			script: `lc.variables.randomInt(-10, 10)`,
			minVal: -10,
			maxVal: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Run multiple times to test randomness
			for i := 0; i < 100; i++ {
				result, err := vm.RunString(tt.script)
				if err != nil {
					t.Fatalf("Script execution failed: %v", err)
				}

				val := result.ToInteger()
				if val < tt.minVal || val > tt.maxVal {
					t.Errorf("randomInt %d not in range [%d, %d]", val, tt.minVal, tt.maxVal)
				}
			}
		})
	}
}

func TestVariablesRandomFloat(t *testing.T) {
	vm := setupVariablesVM(t)

	// Test multiple times
	for i := 0; i < 100; i++ {
		result, err := vm.RunString(`lc.variables.randomFloat()`)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		val := result.ToFloat()
		if val < 0 || val >= 1 {
			t.Errorf("randomFloat %f not in range [0, 1)", val)
		}
	}
}

func TestVariablesRandomString(t *testing.T) {
	vm := setupVariablesVM(t)

	tests := []struct {
		name     string
		script   string
		expected int
	}{
		{
			name:     "default length",
			script:   `lc.variables.randomString()`,
			expected: 16,
		},
		{
			name:     "custom length",
			script:   `lc.variables.randomString(8)`,
			expected: 8,
		},
		{
			name:     "length 1",
			script:   `lc.variables.randomString(1)`,
			expected: 1,
		},
		{
			name:     "zero length becomes 1",
			script:   `lc.variables.randomString(0)`,
			expected: 1,
		},
		{
			name:     "negative length becomes 1",
			script:   `lc.variables.randomString(-5)`,
			expected: 1,
		},
	}

	alphanumeric := regexp.MustCompile(`^[a-zA-Z0-9]+$`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.RunString(tt.script)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}

			str := result.String()
			if len(str) != tt.expected {
				t.Errorf("Expected length %d, got %d", tt.expected, len(str))
			}

			if !alphanumeric.MatchString(str) {
				t.Errorf("randomString contains non-alphanumeric chars: %s", str)
			}
		})
	}
}

func TestVariablesRandomHex(t *testing.T) {
	vm := setupVariablesVM(t)

	tests := []struct {
		name     string
		script   string
		expected int
	}{
		{
			name:     "default length",
			script:   `lc.variables.randomHex()`,
			expected: 16,
		},
		{
			name:     "custom length",
			script:   `lc.variables.randomHex(32)`,
			expected: 32,
		},
	}

	hexRegex := regexp.MustCompile(`^[0-9a-f]+$`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.RunString(tt.script)
			if err != nil {
				t.Fatalf("Script execution failed: %v", err)
			}

			str := result.String()
			if len(str) != tt.expected {
				t.Errorf("Expected length %d, got %d", tt.expected, len(str))
			}

			if !hexRegex.MatchString(str) {
				t.Errorf("randomHex contains non-hex chars: %s", str)
			}
		})
	}
}

func TestVariablesRandomEmail(t *testing.T) {
	vm := setupVariablesVM(t)

	emailRegex := regexp.MustCompile(`^[a-z0-9]+@[a-z.]+$`)

	for i := 0; i < 10; i++ {
		result, err := vm.RunString(`lc.variables.randomEmail()`)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		email := result.String()
		if !emailRegex.MatchString(email) {
			t.Errorf("Invalid email format: %s", email)
		}

		if !strings.Contains(email, "@") {
			t.Errorf("Email missing @: %s", email)
		}
	}
}

func TestVariablesRandomFirstName(t *testing.T) {
	vm := setupVariablesVM(t)

	validNames := map[string]bool{
		"Alice": true, "Bob": true, "Charlie": true, "Diana": true,
		"Edward": true, "Fiona": true, "George": true, "Hannah": true,
		"Ivan": true, "Julia": true, "Kevin": true, "Laura": true,
		"Michael": true, "Nancy": true, "Oscar": true, "Patricia": true,
		"Quinn": true, "Rachel": true, "Samuel": true, "Teresa": true,
		"Ulrich": true, "Victoria": true, "William": true, "Xena": true,
		"Yuri": true, "Zoe": true,
	}

	for i := 0; i < 20; i++ {
		result, err := vm.RunString(`lc.variables.randomFirstName()`)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		name := result.String()
		if !validNames[name] {
			t.Errorf("Unknown first name: %s", name)
		}
	}
}

func TestVariablesRandomLastName(t *testing.T) {
	vm := setupVariablesVM(t)

	validNames := map[string]bool{
		"Anderson": true, "Brown": true, "Clark": true, "Davis": true,
		"Evans": true, "Foster": true, "Garcia": true, "Harris": true,
		"Ivanov": true, "Johnson": true, "King": true, "Lee": true,
		"Martinez": true, "Nelson": true, "O'Brien": true, "Patel": true,
		"Quinn": true, "Roberts": true, "Smith": true, "Taylor": true,
		"Underwood": true, "Vargas": true, "Wilson": true, "Xavier": true,
		"Young": true, "Zhang": true,
	}

	for i := 0; i < 20; i++ {
		result, err := vm.RunString(`lc.variables.randomLastName()`)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		name := result.String()
		if !validNames[name] {
			t.Errorf("Unknown last name: %s", name)
		}
	}
}

func TestVariablesRandomBoolean(t *testing.T) {
	vm := setupVariablesVM(t)

	trueCount := 0
	falseCount := 0

	for i := 0; i < 100; i++ {
		result, err := vm.RunString(`lc.variables.randomBoolean()`)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		if result.ToBoolean() {
			trueCount++
		} else {
			falseCount++
		}
	}

	// Should have some of each (statistically very unlikely to have all of one)
	if trueCount == 0 {
		t.Error("Expected some true values")
	}
	if falseCount == 0 {
		t.Error("Expected some false values")
	}
}
