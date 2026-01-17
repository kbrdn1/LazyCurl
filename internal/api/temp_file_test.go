package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestHasContentChanged_FileDeleted(t *testing.T) {
	// Create a temp file
	info, err := CreateTempFile("test content", ContentTypeText)
	if err != nil {
		t.Fatalf("CreateTempFile() error = %v", err)
	}

	// Delete the file before checking for changes
	if err := os.Remove(info.Path); err != nil {
		t.Fatalf("failed to delete temp file: %v", err)
	}

	// HasContentChanged should return error when file is deleted
	_, err = HasContentChanged(info)
	if err == nil {
		t.Error("HasContentChanged() expected error for deleted file, got nil")
	}
}

func TestCleanupTempFile_EmptyPath(t *testing.T) {
	// Test with empty path in TempFileInfo
	info := &TempFileInfo{Path: ""}

	// Should not panic and return the os.Remove error for empty path
	err := CleanupTempFile(info)
	// Empty path removal behavior varies by OS, but shouldn't panic
	// On most systems, this will return an error (path does not exist)
	_ = err // Just verify no panic
}

func TestCleanupTempFile_DirectoryNotEmpty(t *testing.T) {
	// Create a temp directory with a file inside
	tmpDir, err := os.MkdirTemp("", "lazycurl-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Clean up at the end

	// Create a file inside the directory
	tmpFile := filepath.Join(tmpDir, "testfile.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file in temp dir: %v", err)
	}

	// Try to cleanup the directory (should fail because it's not empty)
	info := &TempFileInfo{Path: tmpDir}
	err = CleanupTempFile(info)

	// Should return an error because os.Remove can't delete non-empty directory
	if err == nil {
		t.Error("CleanupTempFile() expected error for non-empty directory, got nil")
	}
}

func TestCreateTempFile_AllContentTypes(t *testing.T) {
	// Ensure all content types create files correctly
	contentTypes := []struct {
		contentType ContentType
		wantExt     string
	}{
		{ContentTypeJSON, ".json"},
		{ContentTypeXML, ".xml"},
		{ContentTypeHTML, ".html"},
		{ContentTypeText, ".txt"},
		{ContentType("unknown"), ".txt"}, // Unknown defaults to .txt
	}

	for _, tt := range contentTypes {
		t.Run(string(tt.contentType), func(t *testing.T) {
			info, err := CreateTempFile("test", tt.contentType)
			if err != nil {
				t.Fatalf("CreateTempFile() error = %v", err)
			}
			defer CleanupTempFile(info)

			if info.Extension != tt.wantExt {
				t.Errorf("Extension = %q, want %q", info.Extension, tt.wantExt)
			}

			if info.ContentType != tt.contentType {
				t.Errorf("ContentType = %v, want %v", info.ContentType, tt.contentType)
			}

			if info.CreatedAt.IsZero() {
				t.Error("CreatedAt should not be zero")
			}
		})
	}
}

func TestReadTempFile_EmptyPath(t *testing.T) {
	// Test with empty path in TempFileInfo
	info := &TempFileInfo{Path: ""}

	_, err := ReadTempFile(info)
	if err == nil {
		t.Error("ReadTempFile() expected error for empty path, got nil")
	}
}

func TestReadTempFile_NonExistentPath(t *testing.T) {
	// Test with non-existent path
	info := &TempFileInfo{Path: "/nonexistent/path/that/does/not/exist.txt"}

	_, err := ReadTempFile(info)
	if err == nil {
		t.Error("ReadTempFile() expected error for non-existent path, got nil")
	}
}

func TestTempFileInfo_FieldsPopulated(t *testing.T) {
	content := "test content for field validation"
	info, err := CreateTempFile(content, ContentTypeJSON)
	if err != nil {
		t.Fatalf("CreateTempFile() error = %v", err)
	}
	defer CleanupTempFile(info)

	// Verify all fields are properly populated
	if info.Path == "" {
		t.Error("Path should not be empty")
	}

	if info.OriginalContent != content {
		t.Errorf("OriginalContent = %q, want %q", info.OriginalContent, content)
	}

	if info.ContentType != ContentTypeJSON {
		t.Errorf("ContentType = %v, want %v", info.ContentType, ContentTypeJSON)
	}

	if info.Extension != ".json" {
		t.Errorf("Extension = %q, want %q", info.Extension, ".json")
	}

	if info.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}

	// CreatedAt should be recent (within last minute)
	if time.Since(info.CreatedAt) > time.Minute {
		t.Error("CreatedAt should be recent")
	}
}
