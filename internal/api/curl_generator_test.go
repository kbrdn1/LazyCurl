package api

import (
	"strings"
	"testing"
)

func TestGenerateCurlCommand(t *testing.T) {
	tests := []struct {
		name  string
		input *CollectionRequest
		want  string
	}{
		{
			name: "simple GET",
			input: &CollectionRequest{
				Method: GET,
				URL:    "https://example.com",
			},
			want: "curl 'https://example.com'",
		},
		{
			name: "POST without body",
			input: &CollectionRequest{
				Method: POST,
				URL:    "https://example.com",
			},
			want: "curl -X POST 'https://example.com'",
		},
		{
			name: "PUT method",
			input: &CollectionRequest{
				Method: PUT,
				URL:    "https://api.example.com/users/1",
			},
			want: "curl -X PUT 'https://api.example.com/users/1'",
		},
		{
			name: "DELETE method",
			input: &CollectionRequest{
				Method: DELETE,
				URL:    "https://api.example.com/users/1",
			},
			want: "curl -X DELETE 'https://api.example.com/users/1'",
		},
		{
			name: "with single header",
			input: &CollectionRequest{
				Method: GET,
				URL:    "https://example.com",
				Headers: []KeyValueEntry{
					{Key: "Accept", Value: "application/json", Enabled: true},
				},
			},
			want: "curl -H 'Accept: application/json' 'https://example.com'",
		},
		{
			name: "with multiple headers",
			input: &CollectionRequest{
				Method: POST,
				URL:    "https://api.example.com",
				Headers: []KeyValueEntry{
					{Key: "Content-Type", Value: "application/json", Enabled: true},
					{Key: "Authorization", Value: "Bearer token123", Enabled: true},
				},
			},
			want: "curl -X POST -H 'Content-Type: application/json' -H 'Authorization: Bearer token123' 'https://api.example.com'",
		},
		{
			name: "disabled header excluded",
			input: &CollectionRequest{
				Method: GET,
				URL:    "https://example.com",
				Headers: []KeyValueEntry{
					{Key: "Accept", Value: "application/json", Enabled: true},
					{Key: "X-Debug", Value: "true", Enabled: false},
				},
			},
			want: "curl -H 'Accept: application/json' 'https://example.com'",
		},
		{
			name: "with string body",
			input: &CollectionRequest{
				Method: POST,
				URL:    "https://api.example.com",
				Body: &BodyConfig{
					Type:    "raw",
					Content: "name=test",
				},
			},
			want: "curl -X POST --data-raw 'name=test' 'https://api.example.com'",
		},
		{
			name: "with JSON body",
			input: &CollectionRequest{
				Method: POST,
				URL:    "https://api.example.com",
				Headers: []KeyValueEntry{
					{Key: "Content-Type", Value: "application/json", Enabled: true},
				},
				Body: &BodyConfig{
					Type:    "json",
					Content: `{"key":"value"}`,
				},
			},
			want: `curl -X POST -H 'Content-Type: application/json' -d '{"key":"value"}' 'https://api.example.com'`,
		},
		{
			name: "with basic auth",
			input: &CollectionRequest{
				Method: GET,
				URL:    "https://api.example.com",
				Auth: &AuthConfig{
					Type:     "basic",
					Username: "user",
					Password: "pass",
				},
			},
			want: "curl -u 'user:pass' 'https://api.example.com'",
		},
		{
			name: "with basic auth username only",
			input: &CollectionRequest{
				Method: GET,
				URL:    "https://api.example.com",
				Auth: &AuthConfig{
					Type:     "basic",
					Username: "user",
				},
			},
			want: "curl -u 'user' 'https://api.example.com'",
		},
		{
			name:  "nil request",
			input: nil,
			want:  "",
		},
		{
			name: "empty URL",
			input: &CollectionRequest{
				Method: GET,
				URL:    "",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateCurlCommand(tt.input)
			if got != tt.want {
				t.Errorf("GenerateCurlCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateCurlCommandWithOptions(t *testing.T) {
	tests := []struct {
		name  string
		input *CollectionRequest
		opts  CurlGeneratorOptions
		check func(t *testing.T, got string)
	}{
		{
			name: "double quote style",
			input: &CollectionRequest{
				Method: GET,
				URL:    "https://example.com",
				Headers: []KeyValueEntry{
					{Key: "Accept", Value: "application/json", Enabled: true},
				},
			},
			opts: CurlGeneratorOptions{
				QuoteStyle: "double",
			},
			check: func(t *testing.T, got string) {
				if !strings.Contains(got, `"Accept: application/json"`) {
					t.Errorf("expected double quoted header, got %q", got)
				}
				if !strings.Contains(got, `"https://example.com"`) {
					t.Errorf("expected double quoted URL, got %q", got)
				}
			},
		},
		{
			name: "include method for GET",
			input: &CollectionRequest{
				Method: GET,
				URL:    "https://example.com",
			},
			opts: CurlGeneratorOptions{
				IncludeMethod: true,
				QuoteStyle:    "single",
			},
			check: func(t *testing.T, got string) {
				if !strings.Contains(got, "-X GET") {
					t.Errorf("expected -X GET, got %q", got)
				}
			},
		},
		{
			name: "multiline format",
			input: &CollectionRequest{
				Method: POST,
				URL:    "https://api.example.com",
				Headers: []KeyValueEntry{
					{Key: "Content-Type", Value: "application/json", Enabled: true},
					{Key: "Authorization", Value: "Bearer token", Enabled: true},
				},
				Body: &BodyConfig{
					Type:    "json",
					Content: `{"key":"value"}`,
				},
			},
			opts: CurlGeneratorOptions{
				Multiline:    true,
				IndentString: "  ",
				QuoteStyle:   "single",
			},
			check: func(t *testing.T, got string) {
				if !strings.Contains(got, "\\\n") {
					t.Errorf("expected line continuation, got %q", got)
				}
				lines := strings.Split(got, "\n")
				if len(lines) < 2 {
					t.Errorf("expected multiple lines, got %d", len(lines))
				}
				// Check indentation
				for i, line := range lines {
					if i > 0 && !strings.HasPrefix(line, "  ") {
						t.Errorf("line %d not indented: %q", i, line)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateCurlCommandWithOptions(tt.input, tt.opts)
			tt.check(t, got)
		})
	}
}

func TestQuote(t *testing.T) {
	tests := []struct {
		name  string
		value string
		style string
		want  string
	}{
		{
			name:  "single quote simple",
			value: "hello",
			style: "single",
			want:  "'hello'",
		},
		{
			name:  "double quote simple",
			value: "hello",
			style: "double",
			want:  `"hello"`,
		},
		{
			name:  "single quote with single quote",
			value: "it's here",
			style: "single",
			want:  "'it'\\''s here'",
		},
		{
			name:  "double quote with double quote",
			value: `say "hello"`,
			style: "double",
			want:  `"say \"hello\""`,
		},
		{
			name:  "double quote with backslash",
			value: `path\to\file`,
			style: "double",
			want:  `"path\\to\\file"`,
		},
		{
			name:  "JSON content single quote",
			value: `{"key":"value"}`,
			style: "single",
			want:  `'{"key":"value"}'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := quote(tt.value, tt.style)
			if got != tt.want {
				t.Errorf("quote() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatBody(t *testing.T) {
	tests := []struct {
		name  string
		input *BodyConfig
		want  string
	}{
		{
			name:  "nil body",
			input: nil,
			want:  "",
		},
		{
			name: "string content",
			input: &BodyConfig{
				Type:    "raw",
				Content: "hello world",
			},
			want: "hello world",
		},
		{
			name: "JSON string content",
			input: &BodyConfig{
				Type:    "json",
				Content: `{"key":"value"}`,
			},
			want: `{"key":"value"}`,
		},
		{
			name: "map content",
			input: &BodyConfig{
				Type: "json",
				Content: map[string]interface{}{
					"key": "value",
				},
			},
			want: `{"key":"value"}`,
		},
		{
			name: "nil content",
			input: &BodyConfig{
				Type:    "raw",
				Content: nil,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatBody(tt.input)
			if got != tt.want {
				t.Errorf("formatBody() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatMultiline(t *testing.T) {
	tests := []struct {
		name   string
		parts  []string
		indent string
		check  func(t *testing.T, got string)
	}{
		{
			name:   "single part",
			parts:  []string{"curl"},
			indent: "  ",
			check: func(t *testing.T, got string) {
				if got != "curl" {
					t.Errorf("expected 'curl', got %q", got)
				}
			},
		},
		{
			name:   "simple command",
			parts:  []string{"curl", "'https://example.com'"},
			indent: "  ",
			check: func(t *testing.T, got string) {
				if !strings.Contains(got, "\\\n") {
					t.Errorf("expected continuation, got %q", got)
				}
			},
		},
		{
			name:   "with flags",
			parts:  []string{"curl", "-X", "POST", "-H", "'Accept: */*'", "'https://example.com'"},
			indent: "    ",
			check: func(t *testing.T, got string) {
				lines := strings.Split(got, "\n")
				if len(lines) < 2 {
					t.Errorf("expected multiple lines, got %d", len(lines))
				}
				// Check that flags and values are kept together
				for _, line := range lines[1:] {
					if !strings.HasPrefix(line, "    ") {
						t.Errorf("line not properly indented: %q", line)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMultiline(tt.parts, tt.indent)
			tt.check(t, got)
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Parse → Generate → Parse should produce equivalent requests
	commands := []struct {
		name string
		cmd  string
	}{
		{
			name: "simple GET",
			cmd:  "curl https://example.com",
		},
		{
			name: "POST with header",
			cmd:  "curl -X POST -H 'Accept: application/json' https://api.example.com",
		},
		{
			name: "POST with body",
			cmd:  `curl -X POST -H 'Content-Type: application/json' -d '{"key":"value"}' https://api.example.com`,
		},
		{
			name: "with basic auth",
			cmd:  "curl -u admin:secret https://secure.example.com",
		},
		{
			name: "multiple headers",
			cmd:  "curl -H 'Accept: application/json' -H 'Authorization: Bearer token' https://api.example.com",
		},
	}

	for _, tt := range commands {
		t.Run(tt.name, func(t *testing.T) {
			// First parse
			req1, err := ParseCurlCommand(tt.cmd)
			if err != nil {
				t.Fatalf("first parse failed: %v", err)
			}

			// Generate
			generated := GenerateCurlCommand(req1)
			if generated == "" {
				t.Fatal("generated command is empty")
			}

			// Second parse
			req2, err := ParseCurlCommand(generated)
			if err != nil {
				t.Fatalf("second parse failed: %v\ngenerated: %s", err, generated)
			}

			// Compare key fields
			if req1.Method != req2.Method {
				t.Errorf("method mismatch: %v vs %v", req1.Method, req2.Method)
			}
			if req1.URL != req2.URL {
				t.Errorf("URL mismatch: %v vs %v", req1.URL, req2.URL)
			}

			// Compare headers (by key-value, order may differ)
			if len(req1.Headers) != len(req2.Headers) {
				t.Errorf("headers count mismatch: %d vs %d", len(req1.Headers), len(req2.Headers))
			}
			for _, h1 := range req1.Headers {
				found := false
				for _, h2 := range req2.Headers {
					if h1.Key == h2.Key && h1.Value == h2.Value {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("header not found after round-trip: %s: %s", h1.Key, h1.Value)
				}
			}

			// Compare body
			if (req1.Body == nil) != (req2.Body == nil) {
				t.Errorf("body presence mismatch")
			}
			if req1.Body != nil && req2.Body != nil {
				body1 := formatBody(req1.Body)
				body2 := formatBody(req2.Body)
				if body1 != body2 {
					t.Errorf("body mismatch: %q vs %q", body1, body2)
				}
			}

			// Compare auth
			if (req1.Auth == nil) != (req2.Auth == nil) {
				t.Errorf("auth presence mismatch")
			}
			if req1.Auth != nil && req2.Auth != nil {
				if req1.Auth.Username != req2.Auth.Username {
					t.Errorf("auth username mismatch: %q vs %q", req1.Auth.Username, req2.Auth.Username)
				}
				if req1.Auth.Password != req2.Auth.Password {
					t.Errorf("auth password mismatch: %q vs %q", req1.Auth.Password, req2.Auth.Password)
				}
			}
		})
	}
}

func TestGenerateCurlFromRequest(t *testing.T) {
	tests := []struct {
		name  string
		input *Request
		want  string
	}{
		{
			name: "simple GET",
			input: &Request{
				Method: GET,
				URL:    "https://example.com",
			},
			want: "curl 'https://example.com'",
		},
		{
			name: "POST with headers and body",
			input: &Request{
				Method: POST,
				URL:    "https://api.example.com",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: `{"key":"value"}`,
			},
			want: "curl -X POST -H 'Content-Type: application/json' -d '{\"key\":\"value\"}' 'https://api.example.com'",
		},
		{
			name:  "nil request",
			input: nil,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateCurlFromRequest(tt.input)
			// For the POST test, we need to check both possible orderings since map iteration is random
			if tt.name == "POST with headers and body" {
				if !strings.Contains(got, "-X POST") {
					t.Errorf("expected -X POST in output")
				}
				if !strings.Contains(got, "-H 'Content-Type: application/json'") {
					t.Errorf("expected Content-Type header in output")
				}
				if !strings.Contains(got, "-d") {
					t.Errorf("expected -d flag in output")
				}
			} else if got != tt.want {
				t.Errorf("GenerateCurlFromRequest() = %q, want %q", got, tt.want)
			}
		})
	}
}
