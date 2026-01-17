package api

import (
	"sync"
	"testing"
)

func TestNewAssertionCollector(t *testing.T) {
	ac := NewAssertionCollector()

	if ac == nil {
		t.Fatal("NewAssertionCollector() returned nil")
	}
	if len(ac.GetResults()) != 0 {
		t.Error("New collector should have empty results")
	}
	if !ac.AllPassed() {
		t.Error("Empty collector should report AllPassed() as true")
	}
	if ac.FailureCount() != 0 {
		t.Error("Empty collector should have 0 failures")
	}
}

func TestAssertionCollector_RegisterTest_Passed(t *testing.T) {
	ac := NewAssertionCollector()

	ac.RegisterTest("test passed", true, 200, 200, "")

	results := ac.GetResults()
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Name != "test passed" {
		t.Errorf("Name = %q, want %q", r.Name, "test passed")
	}
	if !r.Passed {
		t.Error("Passed should be true")
	}
	if r.Expected != 200 {
		t.Errorf("Expected = %v, want 200", r.Expected)
	}
	if r.Actual != 200 {
		t.Errorf("Actual = %v, want 200", r.Actual)
	}
	if r.Message != "" {
		t.Errorf("Message = %q, want empty", r.Message)
	}
}

func TestAssertionCollector_RegisterTest_Failed(t *testing.T) {
	ac := NewAssertionCollector()

	ac.RegisterTest("test failed", false, 200, 404, "status code mismatch")

	results := ac.GetResults()
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	r := results[0]
	if r.Name != "test failed" {
		t.Errorf("Name = %q, want %q", r.Name, "test failed")
	}
	if r.Passed {
		t.Error("Passed should be false")
	}
	if r.Expected != 200 {
		t.Errorf("Expected = %v, want 200", r.Expected)
	}
	if r.Actual != 404 {
		t.Errorf("Actual = %v, want 404", r.Actual)
	}
	if r.Message != "status code mismatch" {
		t.Errorf("Message = %q, want %q", r.Message, "status code mismatch")
	}
}

func TestAssertionCollector_MultipleAssertions(t *testing.T) {
	ac := NewAssertionCollector()

	ac.RegisterTest("test 1", true, nil, nil, "")
	ac.RegisterTest("test 2", true, "expected", "expected", "")
	ac.RegisterTest("test 3", false, true, false, "boolean mismatch")
	ac.RegisterTest("test 4", true, 100, 100, "")
	ac.RegisterTest("test 5", false, "a", "b", "string mismatch")

	results := ac.GetResults()
	if len(results) != 5 {
		t.Fatalf("Expected 5 results, got %d", len(results))
	}

	// Verify order is preserved
	expectedNames := []string{"test 1", "test 2", "test 3", "test 4", "test 5"}
	for i, name := range expectedNames {
		if results[i].Name != name {
			t.Errorf("results[%d].Name = %q, want %q", i, results[i].Name, name)
		}
	}
}

func TestAssertionCollector_AllPassed(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*AssertionCollector)
		expected bool
	}{
		{
			name:     "empty collector",
			setup:    func(ac *AssertionCollector) {},
			expected: true,
		},
		{
			name: "all passed",
			setup: func(ac *AssertionCollector) {
				ac.RegisterTest("t1", true, nil, nil, "")
				ac.RegisterTest("t2", true, nil, nil, "")
				ac.RegisterTest("t3", true, nil, nil, "")
			},
			expected: true,
		},
		{
			name: "one failure",
			setup: func(ac *AssertionCollector) {
				ac.RegisterTest("t1", true, nil, nil, "")
				ac.RegisterTest("t2", false, nil, nil, "failed")
				ac.RegisterTest("t3", true, nil, nil, "")
			},
			expected: false,
		},
		{
			name: "all failures",
			setup: func(ac *AssertionCollector) {
				ac.RegisterTest("t1", false, nil, nil, "")
				ac.RegisterTest("t2", false, nil, nil, "")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewAssertionCollector()
			tt.setup(ac)
			if ac.AllPassed() != tt.expected {
				t.Errorf("AllPassed() = %v, want %v", ac.AllPassed(), tt.expected)
			}
		})
	}
}

func TestAssertionCollector_FailureCount(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*AssertionCollector)
		expected int
	}{
		{
			name:     "empty collector",
			setup:    func(ac *AssertionCollector) {},
			expected: 0,
		},
		{
			name: "all passed",
			setup: func(ac *AssertionCollector) {
				ac.RegisterTest("t1", true, nil, nil, "")
				ac.RegisterTest("t2", true, nil, nil, "")
			},
			expected: 0,
		},
		{
			name: "one failure",
			setup: func(ac *AssertionCollector) {
				ac.RegisterTest("t1", true, nil, nil, "")
				ac.RegisterTest("t2", false, nil, nil, "")
				ac.RegisterTest("t3", true, nil, nil, "")
			},
			expected: 1,
		},
		{
			name: "multiple failures",
			setup: func(ac *AssertionCollector) {
				ac.RegisterTest("t1", false, nil, nil, "")
				ac.RegisterTest("t2", true, nil, nil, "")
				ac.RegisterTest("t3", false, nil, nil, "")
				ac.RegisterTest("t4", false, nil, nil, "")
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := NewAssertionCollector()
			tt.setup(ac)
			if ac.FailureCount() != tt.expected {
				t.Errorf("FailureCount() = %d, want %d", ac.FailureCount(), tt.expected)
			}
		})
	}
}

func TestAssertionCollector_GetResults_ReturnsCopy(t *testing.T) {
	ac := NewAssertionCollector()
	ac.RegisterTest("original", true, nil, nil, "")

	results1 := ac.GetResults()
	results1[0].Name = "modified"

	// Original should not be modified
	results2 := ac.GetResults()
	if results2[0].Name != "original" {
		t.Error("GetResults() should return a copy, not the original slice")
	}
}

func TestAssertionCollector_Clear(t *testing.T) {
	ac := NewAssertionCollector()
	ac.RegisterTest("t1", true, nil, nil, "")
	ac.RegisterTest("t2", false, nil, nil, "")

	if len(ac.GetResults()) != 2 {
		t.Fatal("Expected 2 results before clear")
	}

	ac.Clear()

	if len(ac.GetResults()) != 0 {
		t.Error("Clear() should remove all results")
	}
	if !ac.AllPassed() {
		t.Error("After Clear(), AllPassed() should return true")
	}
	if ac.FailureCount() != 0 {
		t.Error("After Clear(), FailureCount() should return 0")
	}
}

func TestAssertionCollector_ConcurrentAccess(t *testing.T) {
	ac := NewAssertionCollector()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ac.RegisterTest("test", n%2 == 0, nil, nil, "")
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ac.GetResults()
			_ = ac.AllPassed()
			_ = ac.FailureCount()
		}()
	}

	wg.Wait()

	// Verify final count
	if len(ac.GetResults()) != 100 {
		t.Errorf("Expected 100 results after concurrent writes, got %d", len(ac.GetResults()))
	}

	// Should have 50 passed (even numbers) and 50 failed (odd numbers)
	if ac.FailureCount() != 50 {
		t.Errorf("Expected 50 failures, got %d", ac.FailureCount())
	}
}

func TestAssertionCollector_DifferentValueTypes(t *testing.T) {
	ac := NewAssertionCollector()

	// Test with different value types
	ac.RegisterTest("nil values", true, nil, nil, "")
	ac.RegisterTest("string values", false, "expected", "actual", "")
	ac.RegisterTest("int values", true, 42, 42, "")
	ac.RegisterTest("float values", false, 3.14, 3.15, "precision issue")
	ac.RegisterTest("bool values", true, true, true, "")
	ac.RegisterTest("slice values", false, []int{1, 2}, []int{1, 3}, "slice mismatch")
	ac.RegisterTest("map values", true, map[string]int{"a": 1}, map[string]int{"a": 1}, "")

	results := ac.GetResults()
	if len(results) != 7 {
		t.Fatalf("Expected 7 results, got %d", len(results))
	}

	// Verify types are preserved
	if results[2].Expected != 42 {
		t.Error("Int value should be preserved")
	}
	if results[3].Expected != 3.14 {
		t.Error("Float value should be preserved")
	}
}

func TestAssertionResult_Struct(t *testing.T) {
	r := AssertionResult{
		Name:     "status check",
		Passed:   false,
		Expected: 200,
		Actual:   404,
		Message:  "unexpected status code",
	}

	if r.Name != "status check" {
		t.Errorf("Name = %q, want %q", r.Name, "status check")
	}
	if r.Passed {
		t.Error("Passed should be false")
	}
	if r.Expected != 200 {
		t.Errorf("Expected = %v, want 200", r.Expected)
	}
	if r.Actual != 404 {
		t.Errorf("Actual = %v, want 404", r.Actual)
	}
	if r.Message != "unexpected status code" {
		t.Errorf("Message = %q, want %q", r.Message, "unexpected status code")
	}
}
