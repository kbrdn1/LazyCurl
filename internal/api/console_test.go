package api

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewConsoleEntry(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		wantStatus ConsoleEntryStatus
	}{
		{"success 200", 200, nil, StatusSuccess},
		{"success 201", 201, nil, StatusSuccess},
		{"success 204", 204, nil, StatusSuccess},
		{"redirect 301", 301, nil, StatusRedirect},
		{"redirect 302", 302, nil, StatusRedirect},
		{"redirect 304", 304, nil, StatusRedirect},
		{"client error 400", 400, nil, StatusClientError},
		{"client error 401", 401, nil, StatusClientError},
		{"client error 404", 404, nil, StatusClientError},
		{"server error 500", 500, nil, StatusServerError},
		{"server error 502", 502, nil, StatusServerError},
		{"server error 503", 503, nil, StatusServerError},
		{"network error", 0, errors.New("connection refused"), StatusNetworkError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &Request{Method: GET, URL: "http://test.com"}
			var resp *Response
			if tt.err == nil {
				resp = &Response{StatusCode: tt.statusCode}
			}
			entry := NewConsoleEntry(req, resp, tt.err, time.Second)
			if entry.Status != tt.wantStatus {
				t.Errorf("got status %v, want %v", entry.Status, tt.wantStatus)
			}
			if entry.ID == "" {
				t.Error("expected non-empty ID")
			}
			if entry.Request != req {
				t.Error("expected request to be stored")
			}
		})
	}
}

func TestNewConsoleEntryPanicsOnNilRequest(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on nil request")
		}
	}()
	NewConsoleEntry(nil, nil, nil, time.Second)
}

func TestConsoleEntryHasError(t *testing.T) {
	req := &Request{Method: GET, URL: "http://test.com"}

	entryWithError := NewConsoleEntry(req, nil, errors.New("test error"), time.Second)
	if !entryWithError.HasError() {
		t.Error("expected HasError to return true")
	}

	entryWithoutError := NewConsoleEntry(req, &Response{StatusCode: 200}, nil, time.Second)
	if entryWithoutError.HasError() {
		t.Error("expected HasError to return false")
	}
}

func TestConsoleEntryIsSuccess(t *testing.T) {
	req := &Request{Method: GET, URL: "http://test.com"}

	successEntry := NewConsoleEntry(req, &Response{StatusCode: 200}, nil, time.Second)
	if !successEntry.IsSuccess() {
		t.Error("expected IsSuccess to return true for 200")
	}

	errorEntry := NewConsoleEntry(req, &Response{StatusCode: 404}, nil, time.Second)
	if errorEntry.IsSuccess() {
		t.Error("expected IsSuccess to return false for 404")
	}
}

func TestConsoleEntryGetStatusCode(t *testing.T) {
	req := &Request{Method: GET, URL: "http://test.com"}

	entry := NewConsoleEntry(req, &Response{StatusCode: 201}, nil, time.Second)
	if entry.GetStatusCode() != 201 {
		t.Errorf("expected 201, got %d", entry.GetStatusCode())
	}

	errorEntry := NewConsoleEntry(req, nil, errors.New("error"), time.Second)
	if errorEntry.GetStatusCode() != 0 {
		t.Errorf("expected 0 for error entry, got %d", errorEntry.GetStatusCode())
	}
}

func TestConsoleEntryFormatTimestamp(t *testing.T) {
	req := &Request{Method: GET, URL: "http://test.com"}
	entry := NewConsoleEntry(req, &Response{StatusCode: 200}, nil, time.Second)

	ts := entry.FormatTimestamp()
	if len(ts) != 8 || ts[2] != ':' || ts[5] != ':' {
		t.Errorf("unexpected timestamp format: %s", ts)
	}
}

func TestConsoleEntryFormatDuration(t *testing.T) {
	req := &Request{Method: GET, URL: "http://test.com"}

	tests := []struct {
		name     string
		duration time.Duration
		err      error
		want     string
	}{
		{"microseconds", 500 * time.Microsecond, nil, "500µs"},
		{"milliseconds", 125 * time.Millisecond, nil, "125ms"},
		{"seconds", 1500 * time.Millisecond, nil, "1.5s"},
		{"error", 100 * time.Millisecond, errors.New("error"), "Err"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *Response
			if tt.err == nil {
				resp = &Response{StatusCode: 200}
			}
			entry := NewConsoleEntry(req, resp, tt.err, tt.duration)
			got := entry.FormatDuration()
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestConsoleEntryFormatSize(t *testing.T) {
	req := &Request{Method: GET, URL: "http://test.com"}

	tests := []struct {
		name string
		size int64
		want string
	}{
		{"bytes", 512, "512B"},
		{"kilobytes", 2560, "2.5KB"},
		{"megabytes", 1572864, "1.5MB"},
		{"zero", 0, "0B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := NewConsoleEntry(req, &Response{StatusCode: 200, Size: tt.size}, nil, time.Second)
			got := entry.FormatSize()
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}

	// Test nil response
	errorEntry := NewConsoleEntry(req, nil, errors.New("error"), time.Second)
	if errorEntry.FormatSize() != "-" {
		t.Errorf("expected '-' for nil response, got %s", errorEntry.FormatSize())
	}
}

func TestConsoleEntryCopyHeaders(t *testing.T) {
	req := &Request{
		Method:  GET,
		URL:     "http://test.com",
		Headers: map[string]string{"Authorization": "Bearer token"},
	}
	resp := &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
	}

	entry := NewConsoleEntry(req, resp, nil, time.Second)
	headers := entry.CopyHeaders()

	if headers == "" {
		t.Error("expected non-empty headers")
	}
	if !strings.Contains(headers, "Authorization: Bearer token") {
		t.Error("expected request header to be included")
	}
	if !strings.Contains(headers, "Content-Type: application/json") {
		t.Error("expected response header to be included")
	}
}

func TestConsoleEntryCopyBody(t *testing.T) {
	req := &Request{Method: GET, URL: "http://test.com"}
	resp := &Response{StatusCode: 200, Body: `{"key": "value"}`}

	entry := NewConsoleEntry(req, resp, nil, time.Second)
	body := entry.CopyBody()

	if body != `{"key": "value"}` {
		t.Errorf("unexpected body: %s", body)
	}

	// Test nil response
	errorEntry := NewConsoleEntry(req, nil, errors.New("error"), time.Second)
	if errorEntry.CopyBody() != "" {
		t.Error("expected empty body for error entry")
	}
}

func TestConsoleEntryCopyError(t *testing.T) {
	req := &Request{Method: GET, URL: "http://test.com"}

	errorEntry := NewConsoleEntry(req, nil, errors.New("connection refused"), time.Second)
	if errorEntry.CopyError() != "connection refused" {
		t.Errorf("unexpected error: %s", errorEntry.CopyError())
	}

	successEntry := NewConsoleEntry(req, &Response{StatusCode: 200}, nil, time.Second)
	if successEntry.CopyError() != "" {
		t.Error("expected empty error for success entry")
	}
}

func TestConsoleEntryCopyAll(t *testing.T) {
	req := &Request{
		Method:  POST,
		URL:     "http://test.com/api",
		Headers: map[string]string{"Content-Type": "application/json"},
		Body:    `{"name": "test"}`,
	}
	resp := &Response{
		StatusCode: 201,
		Status:     "201 Created",
		Headers:    map[string][]string{"Location": {"/api/1"}},
		Body:       `{"id": 1}`,
		Size:       10,
	}

	entry := NewConsoleEntry(req, resp, nil, 125*time.Millisecond)
	all := entry.CopyAll()

	if !strings.Contains(all, "POST http://test.com/api") {
		t.Error("expected request method and URL")
	}
	if !strings.Contains(all, "201 Created") {
		t.Error("expected response status")
	}
	if !strings.Contains(all, "=== REQUEST ===") {
		t.Error("expected REQUEST section")
	}
	if !strings.Contains(all, "=== RESPONSE ===") {
		t.Error("expected RESPONSE section")
	}
}

func TestConsoleEntryCopyCookies(t *testing.T) {
	req := &Request{Method: GET, URL: "http://test.com"}

	// Test with Set-Cookie header (standard case)
	resp := &Response{
		StatusCode: 200,
		Headers: map[string][]string{
			"Set-Cookie": {"session=abc123", "token=xyz789"},
		},
	}
	entry := NewConsoleEntry(req, resp, nil, time.Second)
	cookies := entry.CopyCookies()

	if !strings.Contains(cookies, "session=abc123") {
		t.Error("expected session cookie to be included")
	}
	if !strings.Contains(cookies, "token=xyz789") {
		t.Error("expected token cookie to be included")
	}

	// Test with lowercase header name (case-insensitivity)
	respLower := &Response{
		StatusCode: 200,
		Headers: map[string][]string{
			"set-cookie": {"auth=test"},
		},
	}
	entryLower := NewConsoleEntry(req, respLower, nil, time.Second)
	cookiesLower := entryLower.CopyCookies()

	if !strings.Contains(cookiesLower, "auth=test") {
		t.Error("expected case-insensitive cookie header matching")
	}

	// Test with no cookies
	respNoCookies := &Response{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
	}
	entryNoCookies := NewConsoleEntry(req, respNoCookies, nil, time.Second)
	if entryNoCookies.CopyCookies() != "" {
		t.Error("expected empty string when no cookies")
	}

	// Test with nil response
	errorEntry := NewConsoleEntry(req, nil, errors.New("error"), time.Second)
	if errorEntry.CopyCookies() != "" {
		t.Error("expected empty string for error entry")
	}
}

func TestConsoleEntryCopyInfo(t *testing.T) {
	req := &Request{Method: POST, URL: "http://test.com/api"}
	resp := &Response{
		StatusCode: 201,
		Status:     "201 Created",
		Size:       1024,
	}
	entry := NewConsoleEntry(req, resp, nil, 125*time.Millisecond)
	info := entry.CopyInfo()

	if !strings.Contains(info, "Method: POST") {
		t.Error("expected method to be included")
	}
	if !strings.Contains(info, "URL: http://test.com/api") {
		t.Error("expected URL to be included")
	}
	if !strings.Contains(info, "Status: 201 Created") {
		t.Error("expected status to be included")
	}
	if !strings.Contains(info, "Duration:") {
		t.Error("expected duration to be included")
	}

	// Test with error entry
	errorEntry := NewConsoleEntry(req, nil, errors.New("connection refused"), time.Second)
	errorInfo := errorEntry.CopyInfo()

	if !strings.Contains(errorInfo, "Error: connection refused") {
		t.Error("expected error message to be included")
	}
}

func TestConsoleHistoryAdd(t *testing.T) {
	h := NewConsoleHistory(10)

	req := &Request{Method: GET, URL: "http://test.com"}
	entry := NewConsoleEntry(req, &Response{StatusCode: 200}, nil, time.Second)

	id := h.Add(*entry)
	if id != entry.ID {
		t.Error("expected returned ID to match entry ID")
	}
	if h.Len() != 1 {
		t.Errorf("expected length 1, got %d", h.Len())
	}
}

func TestConsoleHistoryMaxSize(t *testing.T) {
	h := NewConsoleHistory(3)
	req := &Request{Method: GET, URL: "http://test.com"}

	for i := 0; i < 5; i++ {
		entry := ConsoleEntry{ID: fmt.Sprintf("entry-%d", i)}
		entry.Request = req
		h.Add(entry)
	}

	if h.Len() != 3 {
		t.Errorf("expected 3 entries, got %d", h.Len())
	}

	// Check that oldest entries were removed
	entries := h.GetAll()
	if entries[0].ID != "entry-2" {
		t.Errorf("expected oldest to be entry-2, got %s", entries[0].ID)
	}
}

func TestConsoleHistoryGet(t *testing.T) {
	h := NewConsoleHistory(10)
	req := &Request{Method: GET, URL: "http://test.com"}
	entry := NewConsoleEntry(req, &Response{StatusCode: 200}, nil, time.Second)

	h.Add(*entry)

	found, ok := h.Get(entry.ID)
	if !ok {
		t.Error("expected to find entry by ID")
	}
	if found.ID != entry.ID {
		t.Error("found entry has wrong ID")
	}

	_, ok = h.Get("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent entry")
	}
}

func TestConsoleHistoryGetReversed(t *testing.T) {
	h := NewConsoleHistory(10)
	req := &Request{Method: GET, URL: "http://test.com"}

	for i := 0; i < 3; i++ {
		entry := ConsoleEntry{ID: fmt.Sprintf("entry-%d", i), Request: req}
		h.Add(entry)
	}

	reversed := h.GetReversed()
	if reversed[0].ID != "entry-2" {
		t.Errorf("expected first element to be newest (entry-2), got %s", reversed[0].ID)
	}
	if reversed[2].ID != "entry-0" {
		t.Errorf("expected last element to be oldest (entry-0), got %s", reversed[2].ID)
	}
}

func TestConsoleHistoryGetByIndex(t *testing.T) {
	h := NewConsoleHistory(10)
	req := &Request{Method: GET, URL: "http://test.com"}

	for i := 0; i < 3; i++ {
		entry := ConsoleEntry{ID: fmt.Sprintf("entry-%d", i), Request: req}
		h.Add(entry)
	}

	// Index 0 should be newest
	entry, ok := h.GetByIndex(0)
	if !ok {
		t.Error("expected to find entry at index 0")
	}
	if entry.ID != "entry-2" {
		t.Errorf("expected entry-2 at index 0, got %s", entry.ID)
	}

	// Index 2 should be oldest
	entry, ok = h.GetByIndex(2)
	if !ok {
		t.Error("expected to find entry at index 2")
	}
	if entry.ID != "entry-0" {
		t.Errorf("expected entry-0 at index 2, got %s", entry.ID)
	}

	// Out of bounds
	_, ok = h.GetByIndex(-1)
	if ok {
		t.Error("expected not to find entry at index -1")
	}
	_, ok = h.GetByIndex(10)
	if ok {
		t.Error("expected not to find entry at index 10")
	}
}

func TestConsoleHistoryClear(t *testing.T) {
	h := NewConsoleHistory(10)
	req := &Request{Method: GET, URL: "http://test.com"}

	for i := 0; i < 3; i++ {
		entry := ConsoleEntry{ID: fmt.Sprintf("entry-%d", i), Request: req}
		h.Add(entry)
	}

	h.Clear()
	if !h.IsEmpty() {
		t.Error("expected history to be empty after Clear")
	}
	if h.Len() != 0 {
		t.Errorf("expected length 0, got %d", h.Len())
	}
}

func TestConsoleHistoryIsEmpty(t *testing.T) {
	h := NewConsoleHistory(10)
	if !h.IsEmpty() {
		t.Error("expected new history to be empty")
	}

	req := &Request{Method: GET, URL: "http://test.com"}
	entry := ConsoleEntry{ID: "test", Request: req}
	h.Add(entry)

	if h.IsEmpty() {
		t.Error("expected history to not be empty after Add")
	}
}

func TestConsoleHistoryDefaultMaxSize(t *testing.T) {
	h := NewConsoleHistory(0)
	// Default should be 1000
	req := &Request{Method: GET, URL: "http://test.com"}

	for i := 0; i < 1001; i++ {
		entry := ConsoleEntry{ID: fmt.Sprintf("entry-%d", i), Request: req}
		h.Add(entry)
	}

	if h.Len() != 1000 {
		t.Errorf("expected 1000 entries with default max, got %d", h.Len())
	}
}
