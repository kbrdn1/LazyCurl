package api

import (
	"errors"
	"os"
	"runtime"
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
		{
			name:       "VISUAL whitespace only falls through to EDITOR",
			visual:     "   ",
			editor:     "nano",
			wantBinary: "nano",
			wantSource: EditorSourceEditor,
		},
		{
			name:           "Both whitespace only falls through to fallback",
			visual:         "   ",
			editor:         "   ",
			wantSource:     EditorSourceFallback,
			skipIfNoEditor: true,
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
	// Skip on Windows - sh is not available
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	tests := []struct {
		name    string
		input   EditorConfig
		wantErr bool
	}{
		{
			name: "valid editor",
			input: EditorConfig{
				Binary: "sh", // sh should exist on all Unix systems
				Args:   []string{"-c", "true"},
				Source: EditorSourceVisual,
			},
			wantErr: false,
		},
		{
			name: "non-existent editor",
			input: EditorConfig{
				Binary: "definitely-not-a-real-editor-12345",
				Args:   nil,
				Source: EditorSourceVisual,
			},
			wantErr: true,
		},
		{
			name: "empty binary",
			input: EditorConfig{
				Binary: "",
				Args:   nil,
				Source: EditorSourceVisual,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ContentType
	}{
		{
			name:  "JSON object",
			input: `{"key": "value"}`,
			want:  ContentTypeJSON,
		},
		{
			name:  "JSON array",
			input: `[1, 2, 3]`,
			want:  ContentTypeJSON,
		},
		{
			name:  "JSON with whitespace",
			input: `  { "name": "test" }  `,
			want:  ContentTypeJSON,
		},
		{
			name:  "XML declaration",
			input: `<?xml version="1.0" encoding="UTF-8"?>`,
			want:  ContentTypeXML,
		},
		{
			name:  "XML root element",
			input: `<root><child>value</child></root>`,
			want:  ContentTypeXML,
		},
		{
			name:  "HTML doctype",
			input: `<!DOCTYPE html><html></html>`,
			want:  ContentTypeHTML,
		},
		{
			name:  "HTML tag",
			input: `<html><body>Hello</body></html>`,
			want:  ContentTypeHTML,
		},
		{
			name:  "Plain text",
			input: `Hello, World!`,
			want:  ContentTypeText,
		},
		{
			name:  "Empty string",
			input: ``,
			want:  ContentTypeText,
		},
		{
			name:  "Whitespace only",
			input: `   `,
			want:  ContentTypeText,
		},
		{
			name:  "Multiline JSON",
			input: "{\n  \"key\": \"value\"\n}",
			want:  ContentTypeJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectContentType(tt.input)
			if got != tt.want {
				t.Errorf("DetectContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetExtensionForContentType(t *testing.T) {
	tests := []struct {
		name  string
		input ContentType
		want  string
	}{
		{name: "JSON type", input: ContentTypeJSON, want: ".json"},
		{name: "XML type", input: ContentTypeXML, want: ".xml"},
		{name: "HTML type", input: ContentTypeHTML, want: ".html"},
		{name: "Text type", input: ContentTypeText, want: ".txt"},
		{name: "Unknown type", input: "unknown", want: ".txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetExtensionForContentType(tt.input)
			if got != tt.want {
				t.Errorf("GetExtensionForContentType(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseEditorCommand(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "simple command",
			input: "vim",
			want:  []string{"vim"},
		},
		{
			name:  "command with one arg",
			input: "code --wait",
			want:  []string{"code", "--wait"},
		},
		{
			name:  "command with multiple args",
			input: "emacs -nw --no-splash",
			want:  []string{"emacs", "-nw", "--no-splash"},
		},
		{
			name:  "command with extra whitespace",
			input: "  vim   -u  NONE  ",
			want:  []string{"vim", "-u", "NONE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEditorCommand(tt.input)
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

func TestGetEditorConfig_NoEditorAvailable(t *testing.T) {
	// This test documents the expected behavior when no editor is available.
	// It only runs meaningfully on systems without nano or vi installed.
	// On most Unix systems, this test verifies fallback detection works.

	// Save original env vars
	originalVisual := os.Getenv("VISUAL")
	originalEditor := os.Getenv("EDITOR")
	defer func() {
		os.Setenv("VISUAL", originalVisual)
		os.Setenv("EDITOR", originalEditor)
	}()

	// Unset both environment variables
	os.Unsetenv("VISUAL")
	os.Unsetenv("EDITOR")

	config, err := GetEditorConfig()

	// If we get an error, it should be ErrNoEditorAvailable
	if err != nil {
		if !errors.Is(err, ErrNoEditorAvailable) {
			t.Errorf("GetEditorConfig() error = %v, want ErrNoEditorAvailable", err)
		}
		// Test passes - no editor was available
		return
	}

	// If we got a config, verify it's from fallback
	if config.Source != EditorSourceFallback {
		t.Errorf("GetEditorConfig() Source = %v, want EditorSourceFallback", config.Source)
	}

	// Verify the binary is one of the expected fallbacks
	if config.Binary == "" {
		t.Error("GetEditorConfig() Binary should not be empty for fallback")
	}
}

func TestEditorSource_String(t *testing.T) {
	tests := []struct {
		source EditorSource
		want   string
	}{
		{EditorSourceVisual, "VISUAL"},
		{EditorSourceEditor, "EDITOR"},
		{EditorSourceFallback, "fallback"},
		{EditorSource("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(string(tt.source), func(t *testing.T) {
			if got := string(tt.source); got != tt.want {
				t.Errorf("EditorSource string = %q, want %q", got, tt.want)
			}
		})
	}
}
