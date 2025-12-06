package api

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ConsoleEntryStatus represents the visual status of a console entry
type ConsoleEntryStatus int

const (
	StatusSuccess      ConsoleEntryStatus = iota // 2xx
	StatusRedirect                               // 3xx
	StatusClientError                            // 4xx
	StatusServerError                            // 5xx
	StatusNetworkError                           // Connection failures
)

// ConsoleEntry represents a single request/response pair in the console
type ConsoleEntry struct {
	ID        string
	Timestamp time.Time
	Request   *Request
	Response  *Response
	Error     error
	Duration  time.Duration
	Status    ConsoleEntryStatus
}

// NewConsoleEntry creates a new console entry from a completed request
func NewConsoleEntry(req *Request, resp *Response, err error, duration time.Duration) *ConsoleEntry {
	if req == nil {
		panic("NewConsoleEntry: request must not be nil")
	}

	entry := &ConsoleEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Request:   req,
		Response:  resp,
		Error:     err,
		Duration:  duration,
	}
	entry.Status = entry.computeStatus()
	return entry
}

// computeStatus determines the status based on response or error
func (e *ConsoleEntry) computeStatus() ConsoleEntryStatus {
	if e.Error != nil {
		return StatusNetworkError
	}
	if e.Response == nil {
		return StatusNetworkError
	}
	switch {
	case e.Response.StatusCode >= 200 && e.Response.StatusCode < 300:
		return StatusSuccess
	case e.Response.StatusCode >= 300 && e.Response.StatusCode < 400:
		return StatusRedirect
	case e.Response.StatusCode >= 400 && e.Response.StatusCode < 500:
		return StatusClientError
	default:
		// Treats 5xx and any unexpected codes (< 100 or >= 600) as server errors
		return StatusServerError
	}
}

// HasError returns true if the entry represents a failed request
func (e *ConsoleEntry) HasError() bool {
	return e.Error != nil
}

// IsSuccess returns true if response is 2xx
func (e *ConsoleEntry) IsSuccess() bool {
	return e.Status == StatusSuccess
}

// GetStatusCode returns the HTTP status code, or 0 if error
func (e *ConsoleEntry) GetStatusCode() int {
	if e.Response == nil {
		return 0
	}
	return e.Response.StatusCode
}

// FormatTimestamp returns timestamp in HH:MM:SS format
func (e *ConsoleEntry) FormatTimestamp() string {
	return e.Timestamp.Format("15:04:05")
}

// FormatDuration returns human-readable duration (e.g., "125ms", "1.2s")
func (e *ConsoleEntry) FormatDuration() string {
	if e.Error != nil {
		return "Err"
	}
	if e.Duration < time.Millisecond {
		return fmt.Sprintf("%dµs", e.Duration.Microseconds())
	}
	if e.Duration < time.Second {
		return fmt.Sprintf("%dms", e.Duration.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", e.Duration.Seconds())
}

// FormatSize returns human-readable response size (e.g., "2.4KB", "1.2MB")
func (e *ConsoleEntry) FormatSize() string {
	if e.Response == nil {
		return "-"
	}
	size := e.Response.Size
	if size < 0 {
		return "-"
	}
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1fKB", float64(size)/1024)
	}
	return fmt.Sprintf("%.1fMB", float64(size)/(1024*1024))
}

// CopyHeaders returns formatted headers string for clipboard
func (e *ConsoleEntry) CopyHeaders() string {
	if e.Response == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("--- Request Headers ---\n")
	if e.Request != nil && e.Request.Headers != nil {
		for key, value := range e.Request.Headers {
			sb.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
	}
	sb.WriteString("\n--- Response Headers ---\n")
	for key, values := range e.Response.Headers {
		for _, value := range values {
			sb.WriteString(fmt.Sprintf("%s: %s\n", key, value))
		}
	}
	return sb.String()
}

// CopyBody returns response body for clipboard
func (e *ConsoleEntry) CopyBody() string {
	if e.Response == nil {
		return ""
	}
	return e.Response.Body
}

// CopyError returns formatted error message for clipboard
func (e *ConsoleEntry) CopyError() string {
	if e.Error == nil {
		return ""
	}
	return e.Error.Error()
}

// CopyCookies returns cookies from response headers for clipboard
func (e *ConsoleEntry) CopyCookies() string {
	if e.Response == nil || e.Response.Headers == nil {
		return ""
	}
	var sb strings.Builder
	// HTTP headers are case-insensitive; check all variants
	for key, cookies := range e.Response.Headers {
		if strings.EqualFold(key, "Set-Cookie") {
			for _, cookie := range cookies {
				sb.WriteString(cookie)
				sb.WriteString("\n")
			}
		}
	}
	return strings.TrimSpace(sb.String())
}

// CopyInfo returns request/response summary info for clipboard
func (e *ConsoleEntry) CopyInfo() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Time: %s\n", e.FormatTimestamp()))
	if e.Request != nil {
		sb.WriteString(fmt.Sprintf("Method: %s\n", e.Request.Method))
		sb.WriteString(fmt.Sprintf("URL: %s\n", e.Request.URL))
	}
	if e.Response != nil {
		sb.WriteString(fmt.Sprintf("Status: %s\n", e.Response.Status))
		sb.WriteString(fmt.Sprintf("Size: %s\n", e.FormatSize()))
	}
	sb.WriteString(fmt.Sprintf("Duration: %s\n", e.FormatDuration()))
	if e.Error != nil {
		sb.WriteString(fmt.Sprintf("Error: %s\n", e.Error.Error()))
	}
	return strings.TrimSpace(sb.String())
}

// CopyAll returns complete request/response for clipboard
func (e *ConsoleEntry) CopyAll() string {
	var sb strings.Builder

	// Request section
	sb.WriteString("=== REQUEST ===\n")
	if e.Request != nil {
		sb.WriteString(fmt.Sprintf("%s %s\n", e.Request.Method, e.Request.URL))
		if e.Request.Headers != nil {
			sb.WriteString("\nHeaders:\n")
			for key, value := range e.Request.Headers {
				sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
			}
		}
		if e.Request.Body != nil {
			sb.WriteString(fmt.Sprintf("\nBody:\n%v\n", e.Request.Body))
		}
	}

	// Response section
	sb.WriteString("\n=== RESPONSE ===\n")
	if e.Error != nil {
		sb.WriteString(fmt.Sprintf("Error: %s\n", e.Error.Error()))
	} else if e.Response != nil {
		sb.WriteString(fmt.Sprintf("%s (%s, %s)\n",
			e.Response.Status,
			e.FormatDuration(),
			e.FormatSize()))
		sb.WriteString("\nHeaders:\n")
		for key, values := range e.Response.Headers {
			for _, value := range values {
				sb.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
			}
		}
		if e.Response.Body != "" {
			sb.WriteString(fmt.Sprintf("\nBody:\n%s\n", e.Response.Body))
		}
	}

	return sb.String()
}

// ConsoleHistory manages a collection of console entries with thread-safe access
type ConsoleHistory struct {
	mu      sync.RWMutex
	entries []ConsoleEntry
	maxSize int
}

// NewConsoleHistory creates a new history manager
func NewConsoleHistory(maxSize int) *ConsoleHistory {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &ConsoleHistory{
		entries: make([]ConsoleEntry, 0),
		maxSize: maxSize,
	}
}

// Add appends a new entry to history (thread-safe)
func (h *ConsoleHistory) Add(entry ConsoleEntry) string {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.entries) >= h.maxSize {
		h.entries = h.entries[1:] // Remove oldest
	}
	h.entries = append(h.entries, entry)
	return entry.ID
}

// Get retrieves an entry by ID (thread-safe)
func (h *ConsoleHistory) Get(id string) (*ConsoleEntry, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for i := range h.entries {
		if h.entries[i].ID == id {
			// Return a copy to avoid data races
			entry := h.entries[i]
			return &entry, true
		}
	}
	return nil, false
}

// GetAll returns all entries in chronological order (oldest first, thread-safe)
func (h *ConsoleHistory) GetAll() []ConsoleEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]ConsoleEntry, len(h.entries))
	copy(result, h.entries)
	return result
}

// GetReversed returns entries in reverse chronological order (newest first, thread-safe)
func (h *ConsoleHistory) GetReversed() []ConsoleEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]ConsoleEntry, len(h.entries))
	for i, j := 0, len(h.entries)-1; j >= 0; i, j = i+1, j-1 {
		result[i] = h.entries[j]
	}
	return result
}

// GetByIndex returns entry at display index (0 = newest, thread-safe)
func (h *ConsoleHistory) GetByIndex(idx int) (*ConsoleEntry, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if idx < 0 || idx >= len(h.entries) {
		return nil, false
	}
	// Convert display index (0 = newest) to internal index
	internalIdx := len(h.entries) - 1 - idx
	// Return a copy to avoid data races
	entry := h.entries[internalIdx]
	return &entry, true
}

// Len returns the number of entries (thread-safe)
func (h *ConsoleHistory) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.entries)
}

// Clear removes all entries (thread-safe)
func (h *ConsoleHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = make([]ConsoleEntry, 0)
}

// IsEmpty returns true if no entries (thread-safe)
func (h *ConsoleHistory) IsEmpty() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.entries) == 0
}
