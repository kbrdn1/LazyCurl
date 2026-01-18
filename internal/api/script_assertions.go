package api

import (
	"sync"
)

// AssertionResult represents the outcome of a test assertion
type AssertionResult struct {
	Name     string      `json:"name"`
	Passed   bool        `json:"passed"`
	Expected interface{} `json:"expected,omitempty"`
	Actual   interface{} `json:"actual,omitempty"`
	Message  string      `json:"message,omitempty"`
}

// AssertionCollector gathers test results during script execution
type AssertionCollector struct {
	results []AssertionResult
	mu      sync.Mutex
}

// NewAssertionCollector creates a new collector
func NewAssertionCollector() *AssertionCollector {
	return &AssertionCollector{
		results: make([]AssertionResult, 0),
	}
}

// RegisterTest adds a test result to the collector
func (c *AssertionCollector) RegisterTest(name string, passed bool, expected, actual interface{}, message string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = append(c.results, AssertionResult{
		Name:     name,
		Passed:   passed,
		Expected: expected,
		Actual:   actual,
		Message:  message,
	})
}

// GetResults returns all assertion results
func (c *AssertionCollector) GetResults() []AssertionResult {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Return a copy to avoid concurrent modification
	result := make([]AssertionResult, len(c.results))
	copy(result, c.results)
	return result
}

// AllPassed returns true if all assertions passed
func (c *AssertionCollector) AllPassed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, r := range c.results {
		if !r.Passed {
			return false
		}
	}
	return true
}

// FailureCount returns the number of failed assertions
func (c *AssertionCollector) FailureCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	count := 0
	for _, r := range c.results {
		if !r.Passed {
			count++
		}
	}
	return count
}

// Clear removes all results
func (c *AssertionCollector) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.results = make([]AssertionResult, 0)
}
