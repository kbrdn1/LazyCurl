package api

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ConsoleLogLevel represents log severity
type ConsoleLogLevel string

const (
	LogLevelLog   ConsoleLogLevel = "log"
	LogLevelInfo  ConsoleLogLevel = "info"
	LogLevelWarn  ConsoleLogLevel = "warn"
	LogLevelError ConsoleLogLevel = "error"
	LogLevelDebug ConsoleLogLevel = "debug"
)

// ConsoleLogEntry represents a single console output
type ConsoleLogEntry struct {
	Level     ConsoleLogLevel `json:"level"`
	Message   string          `json:"message"`
	Timestamp time.Time       `json:"timestamp"`
}

// ScriptConsole collects console output from scripts
type ScriptConsole struct {
	entries []ConsoleLogEntry
	mu      sync.Mutex
}

// NewScriptConsole creates a new console capture instance
func NewScriptConsole() *ScriptConsole {
	return &ScriptConsole{
		entries: make([]ConsoleLogEntry, 0),
	}
}

// Log adds a log level message
func (c *ScriptConsole) Log(args ...interface{}) {
	c.addEntry(LogLevelLog, args...)
}

// Info adds an info level message
func (c *ScriptConsole) Info(args ...interface{}) {
	c.addEntry(LogLevelInfo, args...)
}

// Warn adds a warning level message
func (c *ScriptConsole) Warn(args ...interface{}) {
	c.addEntry(LogLevelWarn, args...)
}

// Error adds an error level message
func (c *ScriptConsole) Error(args ...interface{}) {
	c.addEntry(LogLevelError, args...)
}

// Debug adds a debug level message
func (c *ScriptConsole) Debug(args ...interface{}) {
	c.addEntry(LogLevelDebug, args...)
}

// GetEntries returns all logged entries
func (c *ScriptConsole) GetEntries() []ConsoleLogEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Return a copy to avoid concurrent modification
	result := make([]ConsoleLogEntry, len(c.entries))
	copy(result, c.entries)
	return result
}

// Clear removes all logged entries
func (c *ScriptConsole) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make([]ConsoleLogEntry, 0)
}

// addEntry adds a new entry to the console
func (c *ScriptConsole) addEntry(level ConsoleLogLevel, args ...interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, ConsoleLogEntry{
		Level:     level,
		Message:   formatArgs(args...),
		Timestamp: time.Now(),
	})
}

// formatArgs converts arguments to a formatted string similar to JavaScript console
func formatArgs(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}

	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = formatArg(arg)
	}
	return strings.Join(parts, " ")
}

// formatArg formats a single argument for console output
func formatArg(arg interface{}) string {
	if arg == nil {
		return "undefined"
	}

	switch v := arg.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case map[string]interface{}:
		return formatObject(v)
	case []interface{}:
		return formatArray(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatObject formats a map as a JSON-like string
func formatObject(obj map[string]interface{}) string {
	if len(obj) == 0 {
		return "{}"
	}

	parts := make([]string, 0, len(obj))
	for k, v := range obj {
		parts = append(parts, fmt.Sprintf("%s: %s", k, formatArg(v)))
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

// formatArray formats a slice as a JSON-like string
func formatArray(arr []interface{}) string {
	if len(arr) == 0 {
		return "[]"
	}

	parts := make([]string, len(arr))
	for i, v := range arr {
		parts[i] = formatArg(v)
	}
	return "[ " + strings.Join(parts, ", ") + " ]"
}
