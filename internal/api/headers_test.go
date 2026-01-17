package api

import (
	"strings"
	"testing"
)

func TestHeadersToText(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]string
		want  []string // Lines that should be present (order may vary)
	}{
		{
			name:  "empty headers",
			input: map[string]string{},
			want:  nil,
		},
		{
			name: "single header",
			input: map[string]string{
				"Content-Type": "application/json",
			},
			want: []string{"Content-Type: application/json"},
		},
		{
			name: "multiple headers",
			input: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token123",
			},
			want: []string{
				"Authorization: Bearer token123",
				"Content-Type: application/json",
			},
		},
		{
			name: "header with empty value",
			input: map[string]string{
				"X-Empty": "",
			},
			want: []string{"X-Empty: "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HeadersToText(tt.input)

			if len(tt.input) == 0 {
				if got != "" {
					t.Errorf("HeadersToText() = %q, want empty string", got)
				}
				return
			}

			// Check all expected lines are present
			for _, line := range tt.want {
				if !strings.Contains(got, line) {
					t.Errorf("HeadersToText() missing line %q, got %q", line, got)
				}
			}

			// Check line count matches
			gotLines := strings.Split(got, "\n")
			if len(gotLines) != len(tt.want) {
				t.Errorf("HeadersToText() has %d lines, want %d", len(gotLines), len(tt.want))
			}
		})
	}
}

func TestTextToHeaders(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			name:  "empty text",
			input: "",
			want:  map[string]string{},
		},
		{
			name:  "single header",
			input: "Content-Type: application/json",
			want: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name:  "multiple headers",
			input: "Content-Type: application/json\nAuthorization: Bearer token",
			want: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token",
			},
		},
		{
			name:  "header with value containing colon",
			input: "X-Custom: value:with:colons",
			want: map[string]string{
				"X-Custom": "value:with:colons",
			},
		},
		{
			name:  "skip blank lines",
			input: "Content-Type: application/json\n\nAuthorization: Bearer token\n",
			want: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token",
			},
		},
		{
			name:  "trim whitespace",
			input: "  Content-Type  :   application/json  ",
			want: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			name:  "skip malformed lines",
			input: "Content-Type: application/json\nmalformed line without colon\nAuthorization: Bearer token",
			want: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer token",
			},
		},
		{
			name:  "skip line with only colon",
			input: ":\nContent-Type: application/json",
			want: map[string]string{
				"Content-Type": "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TextToHeaders(tt.input)

			if len(got) != len(tt.want) {
				t.Errorf("TextToHeaders() has %d headers, want %d", len(got), len(tt.want))
			}

			for key, wantValue := range tt.want {
				gotValue, ok := got[key]
				if !ok {
					t.Errorf("TextToHeaders() missing key %q", key)
					continue
				}
				if gotValue != wantValue {
					t.Errorf("TextToHeaders()[%q] = %q, want %q", key, gotValue, wantValue)
				}
			}
		})
	}
}

func TestHeadersRoundTrip(t *testing.T) {
	original := map[string]string{
		"Content-Type":    "application/json",
		"Authorization":   "Bearer secret-token",
		"Accept":          "*/*",
		"X-Custom-Header": "custom-value",
	}

	// Convert to text and back
	text := HeadersToText(original)
	result := TextToHeaders(text)

	// Verify all headers are preserved
	if len(result) != len(original) {
		t.Errorf("round trip changed header count: got %d, want %d", len(result), len(original))
	}

	for key, wantValue := range original {
		gotValue, ok := result[key]
		if !ok {
			t.Errorf("round trip lost key %q", key)
			continue
		}
		if gotValue != wantValue {
			t.Errorf("round trip changed value for %q: got %q, want %q", key, gotValue, wantValue)
		}
	}
}

func TestValidateHeaderText(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantWarnings int
	}{
		{
			name:         "valid headers",
			input:        "Content-Type: application/json\nAuthorization: Bearer token",
			wantWarnings: 0,
		},
		{
			name:         "empty text",
			input:        "",
			wantWarnings: 0,
		},
		{
			name:         "missing separator",
			input:        "Content-Type application/json",
			wantWarnings: 1,
		},
		{
			name:         "empty key",
			input:        ": application/json",
			wantWarnings: 1,
		},
		{
			name:         "multiple issues",
			input:        "valid: header\nmissing separator\n: empty key",
			wantWarnings: 2,
		},
		{
			name:         "blank lines ignored",
			input:        "Content-Type: application/json\n\n\nAuthorization: Bearer token",
			wantWarnings: 0,
		},
		{
			name:         "whitespace-only lines ignored",
			input:        "Content-Type: application/json\n   \nAuthorization: Bearer token",
			wantWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := ValidateHeaderText(tt.input)
			if len(warnings) != tt.wantWarnings {
				t.Errorf("ValidateHeaderText() returned %d warnings, want %d: %v", len(warnings), tt.wantWarnings, warnings)
			}
		})
	}
}

func TestHeadersToText_SortedOutput(t *testing.T) {
	// Verify that output is consistently sorted
	headers := map[string]string{
		"Z-Header": "z-value",
		"A-Header": "a-value",
		"M-Header": "m-value",
	}

	text := HeadersToText(headers)
	lines := strings.Split(text, "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	// Verify sorted order (A, M, Z)
	expectedOrder := []string{
		"A-Header: a-value",
		"M-Header: m-value",
		"Z-Header: z-value",
	}

	for i, expected := range expectedOrder {
		if lines[i] != expected {
			t.Errorf("line %d = %q, want %q", i, lines[i], expected)
		}
	}
}
