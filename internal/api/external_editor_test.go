package api

import (
	"os"
	"testing"
)

func TestGetEditorConfig(t *testing.T) {
	// Save original env vars
	originalVisual := os.Getenv("VISUAL")
	originalEditor := os.Getenv("EDITOR")
	defer func() {
		os.Setenv("VISUAL", originalVisual)
		os.Setenv("EDITOR", originalEditor)
	}()

	tests := []struct {
		name           string
		visual         string
		editor         string
		wantBinary     string
		wantSource     EditorSource
		wantArgsLen    int
		wantErr        bool
		skipIfNoEditor bool
	}{
		{
			name:       "VISUAL takes priority",
			visual:     "vim",
			editor:     "nano",
			wantBinary: "vim",
			wantSource: EditorSourceVisual,
		},
		{
			name:       "EDITOR used when VISUAL not set",
			visual:     "",
			editor:     "nano",
			wantBinary: "nano",
			wantSource: EditorSourceEditor,
		},
		{
			name:        "VISUAL with args",
			visual:      "code --wait",
			editor:      "",
			wantBinary:  "code",
			wantSource:  EditorSourceVisual,
			wantArgsLen: 1,
		},
		{
			name:           "Fallback when neither set",
			visual:         "",
			editor:         "",
			wantSource:     EditorSourceFallback,
			skipIfNoEditor: true, // Skip if fallback editors aren't available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("VISUAL", tt.visual)
			os.Setenv("EDITOR", tt.editor)

			config, err := GetEditorConfig()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				if tt.skipIfNoEditor {
					t.Skip("No fallback editor available")
				}
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantBinary != "" && config.Binary != tt.wantBinary {
				t.Errorf("Binary = %q, want %q", config.Binary, tt.wantBinary)
			}

			if config.Source != tt.wantSource {
				t.Errorf("Source = %v, want %v", config.Source, tt.wantSource)
			}

			if tt.wantArgsLen > 0 && len(config.Args) != tt.wantArgsLen {
				t.Errorf("Args len = %d, want %d", len(config.Args), tt.wantArgsLen)
			}
		})
	}
}

func TestEditorConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  EditorConfig
		wantErr bool
	}{
		{
			name: "valid editor",
			config: EditorConfig{
				Binary: "sh", // sh should exist on all Unix systems
				Args:   []string{"-c", "true"},
				Source: EditorSourceVisual,
			},
			wantErr: false,
		},
		{
			name: "non-existent editor",
			config: EditorConfig{
				Binary: "definitely-not-a-real-editor-12345",
				Args:   nil,
				Source: EditorSourceVisual,
			},
			wantErr: true,
		},
		{
			name: "empty binary",
			config: EditorConfig{
				Binary: "",
				Args:   nil,
				Source: EditorSourceVisual,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    ContentType
	}{
		{
			name:    "JSON object",
			content: `{"key": "value"}`,
			want:    ContentTypeJSON,
		},
		{
			name:    "JSON array",
			content: `[1, 2, 3]`,
			want:    ContentTypeJSON,
		},
		{
			name:    "JSON with whitespace",
			content: `  { "name": "test" }  `,
			want:    ContentTypeJSON,
		},
		{
			name:    "XML declaration",
			content: `<?xml version="1.0" encoding="UTF-8"?>`,
			want:    ContentTypeXML,
		},
		{
			name:    "XML root element",
			content: `<root><child>value</child></root>`,
			want:    ContentTypeXML,
		},
		{
			name:    "HTML doctype",
			content: `<!DOCTYPE html><html></html>`,
			want:    ContentTypeHTML,
		},
		{
			name:    "HTML tag",
			content: `<html><body>Hello</body></html>`,
			want:    ContentTypeHTML,
		},
		{
			name:    "Plain text",
			content: `Hello, World!`,
			want:    ContentTypeText,
		},
		{
			name:    "Empty string",
			content: ``,
			want:    ContentTypeText,
		},
		{
			name:    "Whitespace only",
			content: `   `,
			want:    ContentTypeText,
		},
		{
			name:    "Multiline JSON",
			content: "{\n  \"key\": \"value\"\n}",
			want:    ContentTypeJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectContentType(tt.content)
			if got != tt.want {
				t.Errorf("DetectContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetExtensionForContentType(t *testing.T) {
	tests := []struct {
		contentType ContentType
		want        string
	}{
		{ContentTypeJSON, ".json"},
		{ContentTypeXML, ".xml"},
		{ContentTypeHTML, ".html"},
		{ContentTypeText, ".txt"},
		{"unknown", ".txt"},
	}

	for _, tt := range tests {
		t.Run(string(tt.contentType), func(t *testing.T) {
			got := GetExtensionForContentType(tt.contentType)
			if got != tt.want {
				t.Errorf("GetExtensionForContentType(%v) = %v, want %v", tt.contentType, got, tt.want)
			}
		})
	}
}

func TestParseEditorCommand(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    []string
	}{
		{
			name:    "simple command",
			command: "vim",
			want:    []string{"vim"},
		},
		{
			name:    "command with one arg",
			command: "code --wait",
			want:    []string{"code", "--wait"},
		},
		{
			name:    "command with multiple args",
			command: "emacs -nw --no-splash",
			want:    []string{"emacs", "-nw", "--no-splash"},
		},
		{
			name:    "command with extra whitespace",
			command: "  vim   -u  NONE  ",
			want:    []string{"vim", "-u", "NONE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEditorCommand(tt.command)
			if len(got) != len(tt.want) {
				t.Errorf("parseEditorCommand() len = %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseEditorCommand()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
