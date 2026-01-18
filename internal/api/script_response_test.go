package api

import (
	"net/http"
	"testing"
	"time"
)

func TestNewScriptResponse_NilResponse(t *testing.T) {
	sr := NewScriptResponse(nil, "", 0)

	if sr == nil {
		t.Fatal("NewScriptResponse(nil) returned nil")
	}
	if sr.Status() != 0 {
		t.Errorf("Status() = %d, want 0", sr.Status())
	}
	if sr.StatusText() != "" {
		t.Errorf("StatusText() = %q, want empty", sr.StatusText())
	}
	if sr.Body() != "" {
		t.Errorf("Body() = %q, want empty", sr.Body())
	}
	if sr.Time() != 0 {
		t.Errorf("Time() = %d, want 0", sr.Time())
	}
}

func TestNewScriptResponse_ValidResponse(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header: http.Header{
			"Content-Type":  []string{"application/json"},
			"X-Custom":      []string{"value"},
			"Cache-Control": []string{"no-cache", "no-store"},
		},
	}

	body := `{"name": "test", "success": true}`
	duration := 150 * time.Millisecond

	sr := NewScriptResponse(resp, body, duration)

	if sr.Status() != 200 {
		t.Errorf("Status() = %d, want 200", sr.Status())
	}
	if sr.StatusText() != "200 OK" {
		t.Errorf("StatusText() = %q, want %q", sr.StatusText(), "200 OK")
	}
	if sr.Body() != body {
		t.Errorf("Body() = %q, want %q", sr.Body(), body)
	}
	if sr.Time() != 150 {
		t.Errorf("Time() = %d, want 150", sr.Time())
	}
}

func TestNewScriptResponseFromData(t *testing.T) {
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Request-Id": "abc123",
	}

	sr := NewScriptResponseFromData(201, "201 Created", headers, `{"id": 1}`, 50)

	if sr.Status() != 201 {
		t.Errorf("Status() = %d, want 201", sr.Status())
	}
	if sr.StatusText() != "201 Created" {
		t.Errorf("StatusText() = %q, want %q", sr.StatusText(), "201 Created")
	}
	if sr.Body() != `{"id": 1}` {
		t.Errorf("Body() = %q, want %q", sr.Body(), `{"id": 1}`)
	}
	if sr.Time() != 50 {
		t.Errorf("Time() = %d, want 50", sr.Time())
	}
}

func TestNewScriptResponseFromData_NilHeaders(t *testing.T) {
	sr := NewScriptResponseFromData(200, "200 OK", nil, "body", 100)

	// Should not panic
	headers := sr.Headers()
	if headers == nil {
		t.Error("Headers() should not return nil")
	}
	if len(headers) != 0 {
		t.Errorf("len(Headers()) = %d, want 0", len(headers))
	}
}

func TestScriptResponse_GetHeader(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			"X-Custom-Value": []string{"test123"},
		},
	}

	sr := NewScriptResponse(resp, "", 0)

	t.Run("exact match", func(t *testing.T) {
		if sr.GetHeader("Content-Type") != "application/json" {
			t.Errorf("GetHeader('Content-Type') = %q, want %q", sr.GetHeader("Content-Type"), "application/json")
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		if sr.GetHeader("content-type") != "application/json" {
			t.Error("GetHeader should be case-insensitive")
		}
		if sr.GetHeader("CONTENT-TYPE") != "application/json" {
			t.Error("GetHeader should be case-insensitive (uppercase)")
		}
	})

	t.Run("non-existent header", func(t *testing.T) {
		if sr.GetHeader("X-Does-Not-Exist") != "" {
			t.Error("GetHeader should return empty string for non-existent header")
		}
	})
}

func TestScriptResponse_Headers(t *testing.T) {
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Request-Id": "abc123",
	}

	sr := NewScriptResponseFromData(200, "200 OK", headers, "", 0)

	t.Run("returns copy", func(t *testing.T) {
		h := sr.Headers()
		h["Content-Type"] = "modified"

		// Original should not be modified
		if sr.GetHeader("Content-Type") != "application/json" {
			t.Error("Headers() should return a copy")
		}
	})

	t.Run("contains all headers", func(t *testing.T) {
		h := sr.Headers()
		if len(h) != 2 {
			t.Errorf("len(Headers()) = %d, want 2", len(h))
		}
		if h["Content-Type"] != "application/json" {
			t.Error("Missing Content-Type header")
		}
		if h["X-Request-Id"] != "abc123" {
			t.Error("Missing X-Request-Id header")
		}
	})
}

func TestScriptResponse_DifferentStatusCodes(t *testing.T) {
	tests := []struct {
		code int
		text string
	}{
		{200, "200 OK"},
		{201, "201 Created"},
		{204, "204 No Content"},
		{301, "301 Moved Permanently"},
		{400, "400 Bad Request"},
		{401, "401 Unauthorized"},
		{403, "403 Forbidden"},
		{404, "404 Not Found"},
		{500, "500 Internal Server Error"},
		{502, "502 Bad Gateway"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			sr := NewScriptResponseFromData(tt.code, tt.text, nil, "", 0)
			if sr.Status() != tt.code {
				t.Errorf("Status() = %d, want %d", sr.Status(), tt.code)
			}
			if sr.StatusText() != tt.text {
				t.Errorf("StatusText() = %q, want %q", sr.StatusText(), tt.text)
			}
		})
	}
}
