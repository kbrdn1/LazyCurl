package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateTempFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		contentType ContentType
		wantExt     string
	}{
		{
			name:        "JSON content",
			content:     `{"key": "value"}`,
			contentType: ContentTypeJSON,
			wantExt:     ".json",
		},
		{
			name:        "XML content",
			content:     `<?xml version="1.0"?><root/>`,
			contentType: ContentTypeXML,
			wantExt:     ".xml",
		},
		{
			name:        "HTML content",
			content:     `<!DOCTYPE html><html></html>`,
			contentType: ContentTypeHTML,
			wantExt:     ".html",
		},
		{
			name:        "Plain text",
			content:     "Hello, World!",
			contentType: ContentTypeText,
			wantExt:     ".txt",
		},
		{
			name:        "Empty content",
			content:     "",
			contentType: ContentTypeText,
			wantExt:     ".txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := CreateTempFile(tt.content, tt.contentType)
			if err != nil {
				t.Fatalf("CreateTempFile() error = %v", err)
			}
			defer CleanupTempFile(info)

			// Verify file exists
			if _, err := os.Stat(info.Path); os.IsNotExist(err) {
				t.Error("temp file was not created")
			}

			// Verify extension
			if !strings.HasSuffix(info.Path, tt.wantExt) {
				t.Errorf("file extension = %q, want suffix %q", filepath.Ext(info.Path), tt.wantExt)
			}

			// Verify content type stored
			if info.ContentType != tt.contentType {
				t.Errorf("ContentType = %v, want %v", info.ContentType, tt.contentType)
			}

			// Verify original content stored
			if info.OriginalContent != tt.content {
				t.Errorf("OriginalContent = %q, want %q", info.OriginalContent, tt.content)
			}

			// Verify file content matches
			content, err := os.ReadFile(info.Path)
			if err != nil {
				t.Fatalf("failed to read temp file: %v", err)
			}
			if string(content) != tt.content {
				t.Errorf("file content = %q, want %q", string(content), tt.content)
			}
		})
	}
}

func TestReadTempFile(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() (*TempFileInfo, func())
		modifyContent  string
		wantContent    string
		wantErr        bool
		wantErrContain string
	}{
		{
			name: "read original content",
			setup: func() (*TempFileInfo, func()) {
				info, _ := CreateTempFile("test content", ContentTypeText)
				return info, func() { CleanupTempFile(info) }
			},
			wantContent: "test content",
			wantErr:     false,
		},
		{
			name: "read modified content",
			setup: func() (*TempFileInfo, func()) {
				info, _ := CreateTempFile("original", ContentTypeText)
				return info, func() { CleanupTempFile(info) }
			},
			modifyContent: "modified content",
			wantContent:   "modified content",
			wantErr:       false,
		},
		{
			name: "non-existent file",
			setup: func() (*TempFileInfo, func()) {
				return &TempFileInfo{Path: "/nonexistent/path/file.txt"}, func() {}
			},
			wantErr: true,
		},
		{
			name: "nil info",
			setup: func() (*TempFileInfo, func()) {
				return nil, func() {}
			},
			wantErr:        true,
			wantErrContain: "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, cleanup := tt.setup()
			defer cleanup()

			// Apply modification if specified
			if tt.modifyContent != "" && info != nil {
				if err := os.WriteFile(info.Path, []byte(tt.modifyContent), 0644); err != nil {
					t.Fatalf("failed to modify temp file: %v", err)
				}
			}

			content, err := ReadTempFile(info)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadTempFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErrContain != "" && err != nil {
				if !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Errorf("ReadTempFile() error = %v, want error containing %q", err, tt.wantErrContain)
				}
			}
			if !tt.wantErr && content != tt.wantContent {
				t.Errorf("ReadTempFile() = %q, want %q", content, tt.wantContent)
			}
		})
	}
}

func TestCleanupTempFile(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *TempFileInfo
		wantErr bool
	}{
		{
			name: "cleanup existing file",
			setup: func() *TempFileInfo {
				info, _ := CreateTempFile("test", ContentTypeText)
				return info
			},
			wantErr: false,
		},
		{
			name: "cleanup already deleted file",
			setup: func() *TempFileInfo {
				info, _ := CreateTempFile("test", ContentTypeText)
				os.Remove(info.Path)
				return info
			},
			wantErr: false,
		},
		{
			name: "cleanup nil info",
			setup: func() *TempFileInfo {
				return nil
			},
			wantErr: false, // Should be no-op for nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := tt.setup()

			err := CleanupTempFile(info)
			if (err != nil) != tt.wantErr {
				t.Errorf("CleanupTempFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify file is gone (only if info is not nil and had a real path)
			if info != nil && info.Path != "" {
				if _, statErr := os.Stat(info.Path); !os.IsNotExist(statErr) {
					t.Error("temp file was not deleted")
				}
			}
		})
	}
}

func TestHasContentChanged(t *testing.T) {
	tests := []struct {
		name        string
		original    string
		modified    string
		wantChanged bool
	}{
		{
			name:        "no change",
			original:    "same content",
			modified:    "same content",
			wantChanged: false,
		},
		{
			name:        "content changed",
			original:    "original",
			modified:    "modified",
			wantChanged: true,
		},
		{
			name:        "whitespace added",
			original:    "text",
			modified:    "text ",
			wantChanged: true,
		},
		{
			name:        "empty to content",
			original:    "",
			modified:    "something",
			wantChanged: true,
		},
		{
			name:        "content to empty",
			original:    "something",
			modified:    "",
			wantChanged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := CreateTempFile(tt.original, ContentTypeText)
			if err != nil {
				t.Fatalf("CreateTempFile() error = %v", err)
			}
			defer CleanupTempFile(info)

			// Modify if needed
			if tt.original != tt.modified {
				if err := os.WriteFile(info.Path, []byte(tt.modified), 0644); err != nil {
					t.Fatalf("failed to modify file: %v", err)
				}
			}

			changed, err := HasContentChanged(info)
			if err != nil {
				t.Fatalf("HasContentChanged() error = %v", err)
			}

			if changed != tt.wantChanged {
				t.Errorf("HasContentChanged() = %v, want %v", changed, tt.wantChanged)
			}
		})
	}
}

func TestTempFileInfo_Extension(t *testing.T) {
	tests := []struct {
		name  string
		input ContentType
		want  string
	}{
		{name: "JSON extension", input: ContentTypeJSON, want: ".json"},
		{name: "XML extension", input: ContentTypeXML, want: ".xml"},
		{name: "HTML extension", input: ContentTypeHTML, want: ".html"},
		{name: "Text extension", input: ContentTypeText, want: ".txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := CreateTempFile("content", tt.input)
			if err != nil {
				t.Fatalf("CreateTempFile() error = %v", err)
			}
			defer CleanupTempFile(info)

			if info.Extension != tt.want {
				t.Errorf("Extension = %q, want %q", info.Extension, tt.want)
			}
		})
	}
}

func TestHasContentChanged_NilInfo(t *testing.T) {
	_, err := HasContentChanged(nil)
	if err == nil {
		t.Error("HasContentChanged(nil) expected error, got nil")
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("HasContentChanged(nil) error = %v, want error containing 'nil'", err)
	}
}
