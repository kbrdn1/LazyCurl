package api

import (
	"testing"
)

func TestNewScriptRequest(t *testing.T) {
	t.Run("with nil request", func(t *testing.T) {
		sr := NewScriptRequest(nil)
		if sr == nil {
			t.Fatal("NewScriptRequest(nil) returned nil")
		}
		if sr.Method() != "GET" {
			t.Errorf("Method() = %q, want %q", sr.Method(), "GET")
		}
		if sr.URL() != "" {
			t.Errorf("URL() = %q, want empty", sr.URL())
		}
		if sr.IsModified() {
			t.Error("IsModified() should be false for new request")
		}
	})

	t.Run("with valid request", func(t *testing.T) {
		req := &CollectionRequest{
			Method: "POST",
			URL:    "https://api.example.com/users",
			Headers: []KeyValueEntry{
				{Key: "Content-Type", Value: "application/json", Enabled: true},
				{Key: "Authorization", Value: "Bearer token123", Enabled: true},
				{Key: "X-Disabled", Value: "should-not-appear", Enabled: false},
			},
			Body: &BodyConfig{
				Type:    "raw",
				Content: `{"name": "test"}`,
			},
		}

		sr := NewScriptRequest(req)

		if sr.Method() != "POST" {
			t.Errorf("Method() = %q, want %q", sr.Method(), "POST")
		}
		if sr.URL() != "https://api.example.com/users" {
			t.Errorf("URL() = %q, want %q", sr.URL(), "https://api.example.com/users")
		}
		if sr.GetHeader("Content-Type") != "application/json" {
			t.Errorf("GetHeader('Content-Type') = %q, want %q", sr.GetHeader("Content-Type"), "application/json")
		}
		if sr.GetHeader("X-Disabled") != "" {
			t.Error("disabled header should not be included")
		}
		if sr.Body() != `{"name": "test"}` {
			t.Errorf("Body() = %q, want %q", sr.Body(), `{"name": "test"}`)
		}
	})
}

func TestScriptRequest_SetURL(t *testing.T) {
	sr := NewScriptRequest(nil)

	sr.SetURL("https://new.example.com/api")

	if sr.URL() != "https://new.example.com/api" {
		t.Errorf("URL() = %q, want %q", sr.URL(), "https://new.example.com/api")
	}
	if !sr.IsModified() {
		t.Error("IsModified() should be true after SetURL")
	}
}

func TestScriptRequest_SetBody(t *testing.T) {
	sr := NewScriptRequest(nil)

	sr.SetBody(`{"key": "value"}`)

	if sr.Body() != `{"key": "value"}` {
		t.Errorf("Body() = %q, want %q", sr.Body(), `{"key": "value"}`)
	}
	if !sr.IsModified() {
		t.Error("IsModified() should be true after SetBody")
	}
}

func TestScriptRequest_Headers(t *testing.T) {
	sr := NewScriptRequest(nil)

	t.Run("SetHeader", func(t *testing.T) {
		sr.SetHeader("X-Custom", "value1")
		if sr.GetHeader("X-Custom") != "value1" {
			t.Errorf("GetHeader('X-Custom') = %q, want %q", sr.GetHeader("X-Custom"), "value1")
		}
		if !sr.IsModified() {
			t.Error("IsModified() should be true after SetHeader")
		}
	})

	t.Run("case-insensitive get", func(t *testing.T) {
		sr.SetHeader("Content-Type", "application/json")
		if sr.GetHeader("content-type") != "application/json" {
			t.Error("GetHeader should be case-insensitive")
		}
	})

	t.Run("SetHeader replaces existing case-insensitive", func(t *testing.T) {
		sr.SetHeader("CONTENT-TYPE", "text/plain")
		headers := sr.Headers()
		// Should only have one Content-Type header
		count := 0
		for k := range headers {
			if k == "CONTENT-TYPE" || k == "Content-Type" {
				count++
			}
		}
		if count != 1 {
			t.Errorf("expected 1 Content-Type header, got %d", count)
		}
	})

	t.Run("RemoveHeader", func(t *testing.T) {
		sr.SetHeader("X-ToRemove", "value")
		sr.RemoveHeader("x-toremove") // case-insensitive
		if sr.GetHeader("X-ToRemove") != "" {
			t.Error("header should be removed")
		}
	})

	t.Run("Headers returns copy", func(t *testing.T) {
		sr.SetHeader("Original", "value")
		headers := sr.Headers()
		headers["Original"] = "modified"

		if sr.GetHeader("Original") != "value" {
			t.Error("Headers() should return a copy")
		}
	})
}

func TestScriptRequest_ApplyTo(t *testing.T) {
	original := &CollectionRequest{
		Method: "GET",
		URL:    "https://old.example.com",
		Headers: []KeyValueEntry{
			{Key: "Old-Header", Value: "old", Enabled: true},
		},
	}

	sr := NewScriptRequest(original)
	sr.SetURL("https://new.example.com")
	sr.SetHeader("New-Header", "new")
	sr.SetBody(`{"updated": true}`)

	sr.ApplyTo(original)

	if original.URL != "https://new.example.com" {
		t.Errorf("URL not updated: %q", original.URL)
	}

	// Check headers were updated
	foundNewHeader := false
	for _, h := range original.Headers {
		if h.Key == "New-Header" && h.Value == "new" {
			foundNewHeader = true
		}
	}
	if !foundNewHeader {
		t.Error("new header not applied")
	}

	if original.Body == nil || original.Body.Content != `{"updated": true}` {
		t.Error("body not updated")
	}
}

func TestScriptRequest_ApplyTo_NilRequest(t *testing.T) {
	sr := NewScriptRequest(nil)
	sr.SetURL("https://example.com")

	// Should not panic
	sr.ApplyTo(nil)
}

func TestScriptRequest_ApplyTo_NotModified(t *testing.T) {
	original := &CollectionRequest{
		URL: "https://original.com",
	}

	sr := NewScriptRequest(original)
	// Don't modify anything

	sr.ApplyTo(original)

	if original.URL != "https://original.com" {
		t.Error("should not modify request if not modified")
	}
}
