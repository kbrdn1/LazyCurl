package format

import (
	"strings"
	"testing"
)

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        []byte
		expected    ContentType
	}{
		{
			name:        "JSON from header",
			contentType: "application/json",
			body:        []byte(`{"key": "value"}`),
			expected:    ContentTypeJSON,
		},
		{
			name:        "JSON from body",
			contentType: "",
			body:        []byte(`{"key": "value"}`),
			expected:    ContentTypeJSON,
		},
		{
			name:        "JSON array",
			contentType: "",
			body:        []byte(`[1, 2, 3]`),
			expected:    ContentTypeJSON,
		},
		{
			name:        "XML from header",
			contentType: "application/xml",
			body:        []byte(`<root></root>`),
			expected:    ContentTypeXML,
		},
		{
			name:        "XML from body",
			contentType: "",
			body:        []byte(`<?xml version="1.0"?><root></root>`),
			expected:    ContentTypeXML,
		},
		{
			name:        "HTML",
			contentType: "text/html",
			body:        []byte(`<!DOCTYPE html><html></html>`),
			expected:    ContentTypeHTML,
		},
		{
			name:        "Plain text",
			contentType: "text/plain",
			body:        []byte(`Just some text`),
			expected:    ContentTypeText,
		},
		{
			name:        "Empty body",
			contentType: "",
			body:        []byte{},
			expected:    ContentTypeText,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectContentType(tt.contentType, tt.body)
			if result != tt.expected {
				t.Errorf("DetectContentType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		indent        string
		shouldContain []string
		wantErr       bool
	}{
		{
			name:          "Simple object",
			input:         []byte(`{"name":"John","age":30}`),
			indent:        "  ",
			shouldContain: []string{"\"name\": \"John\"", "\"age\": 30"},
			wantErr:       false,
		},
		{
			name:          "Nested object",
			input:         []byte(`{"user":{"name":"John","address":{"city":"NYC"}}}`),
			indent:        "  ",
			shouldContain: []string{"\"user\"", "\"name\": \"John\"", "\"city\": \"NYC\""},
			wantErr:       false,
		},
		{
			name:          "Array",
			input:         []byte(`[1,2,3]`),
			indent:        "  ",
			shouldContain: []string{"[\n  1,\n  2,\n  3\n]"},
			wantErr:       false,
		},
		{
			name:          "Empty",
			input:         []byte{},
			indent:        "  ",
			shouldContain: []string{""},
			wantErr:       false,
		},
		{
			name:          "Invalid JSON",
			input:         []byte(`{invalid}`),
			indent:        "  ",
			shouldContain: []string{},
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatJSON(tt.input, tt.indent)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for _, expected := range tt.shouldContain {
					if !strings.Contains(result, expected) {
						t.Errorf("FormatJSON() result does not contain expected string: %v", expected)
					}
				}
			}
		})
	}
}

func TestMinifyJSON(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		shouldContain []string
		wantErr       bool
	}{
		{
			name:          "Formatted JSON",
			input:         []byte("{\n  \"name\": \"John\",\n  \"age\": 30\n}"),
			shouldContain: []string{"\"name\":\"John\"", "\"age\":30"},
			wantErr:       false,
		},
		{
			name:          "Already minified",
			input:         []byte(`{"name":"John"}`),
			shouldContain: []string{"\"name\":\"John\""},
			wantErr:       false,
		},
		{
			name:          "Empty",
			input:         []byte{},
			shouldContain: []string{""},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MinifyJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MinifyJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				for _, expected := range tt.shouldContain {
					if !strings.Contains(result, expected) {
						t.Errorf("MinifyJSON() result does not contain expected string: %v", expected)
					}
				}
			}
		})
	}
}

func TestValidateJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "Valid JSON object",
			input:   []byte(`{"key": "value"}`),
			wantErr: false,
		},
		{
			name:    "Valid JSON array",
			input:   []byte(`[1, 2, 3]`),
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			input:   []byte(`{key: value}`),
			wantErr: true,
		},
		{
			name:    "Empty",
			input:   []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name             string
		contentType      string
		body             []byte
		expectedType     ContentType
		expectedContains string
	}{
		{
			name:             "Format JSON",
			contentType:      "application/json",
			body:             []byte(`{"name":"John"}`),
			expectedType:     ContentTypeJSON,
			expectedContains: "\"name\": \"John\"",
		},
		{
			name:             "Format plain text",
			contentType:      "text/plain",
			body:             []byte(`Hello World`),
			expectedType:     ContentTypeText,
			expectedContains: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, detectedType, err := Format(tt.contentType, tt.body)
			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}
			if detectedType != tt.expectedType {
				t.Errorf("Format() type = %v, want %v", detectedType, tt.expectedType)
			}
			if !strings.Contains(result, tt.expectedContains) {
				t.Errorf("Format() result does not contain expected string: %v", tt.expectedContains)
			}
		})
	}
}

func TestPrettyPrint(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        []byte
		maxLength   int
		contains    []string
	}{
		{
			name:        "JSON response",
			contentType: "application/json",
			body:        []byte(`{"users":[{"name":"John"}]}`),
			maxLength:   0,
			contains:    []string{"📄 JSON", "\"users\"", "\"name\": \"John\""},
		},
		{
			name:        "Empty response",
			contentType: "",
			body:        []byte{},
			maxLength:   0,
			contains:    []string{"(empty response)"},
		},
		{
			name:        "Truncated response",
			contentType: "text/plain",
			body:        []byte(strings.Repeat("a", 200)),
			maxLength:   50,
			contains:    []string{"(truncated)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PrettyPrint(tt.contentType, tt.body, tt.maxLength)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("PrettyPrint() result does not contain expected string: %v", expected)
				}
			}
		})
	}
}
