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
	// Create a temp file
	originalContent := "test content"
	info, err := CreateTempFile(originalContent, ContentTypeText)
	if err != nil {
		t.Fatalf("CreateTempFile() error = %v", err)
	}
	defer CleanupTempFile(info)

	// Read the content
	content, err := ReadTempFile(info)
	if err != nil {
		t.Fatalf("ReadTempFile() error = %v", err)
	}

	if content != originalContent {
		t.Errorf("ReadTempFile() = %q, want %q", content, originalContent)
	}

	// Modify the file
	newContent := "modified content"
	if err := os.WriteFile(info.Path, []byte(newContent), 0644); err != nil {
		t.Fatalf("failed to modify temp file: %v", err)
	}

	// Read again
	content, err = ReadTempFile(info)
	if err != nil {
		t.Fatalf("ReadTempFile() after modification error = %v", err)
	}

	if content != newContent {
		t.Errorf("ReadTempFile() after modification = %q, want %q", content, newContent)
	}
}

func TestReadTempFile_NonExistent(t *testing.T) {
	info := &TempFileInfo{
		Path: "/nonexistent/path/file.txt",
	}

	_, err := ReadTempFile(info)
	if err == nil {
		t.Error("ReadTempFile() expected error for non-existent file")
	}
}

func TestCleanupTempFile(t *testing.T) {
	// Create a temp file
	info, err := CreateTempFile("test", ContentTypeText)
	if err != nil {
		t.Fatalf("CreateTempFile() error = %v", err)
	}

	// Verify it exists
	if _, err := os.Stat(info.Path); os.IsNotExist(err) {
		t.Fatal("temp file was not created")
	}

	// Cleanup
	if err := CleanupTempFile(info); err != nil {
		t.Fatalf("CleanupTempFile() error = %v", err)
	}

	// Verify it's gone
	if _, err := os.Stat(info.Path); !os.IsNotExist(err) {
		t.Error("temp file was not deleted")
	}
}

func TestCleanupTempFile_AlreadyDeleted(t *testing.T) {
	// Create and immediately delete
	info, err := CreateTempFile("test", ContentTypeText)
	if err != nil {
		t.Fatalf("CreateTempFile() error = %v", err)
	}
	os.Remove(info.Path)

	// Cleanup should not error
	if err := CleanupTempFile(info); err != nil {
		t.Errorf("CleanupTempFile() should not error for already deleted file: %v", err)
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
		contentType ContentType
		wantExt     string
	}{
		{ContentTypeJSON, ".json"},
		{ContentTypeXML, ".xml"},
		{ContentTypeHTML, ".html"},
		{ContentTypeText, ".txt"},
	}

	for _, tt := range tests {
		t.Run(string(tt.contentType), func(t *testing.T) {
			info, err := CreateTempFile("content", tt.contentType)
			if err != nil {
				t.Fatalf("CreateTempFile() error = %v", err)
			}
			defer CleanupTempFile(info)

			if info.Extension != tt.wantExt {
				t.Errorf("Extension = %q, want %q", info.Extension, tt.wantExt)
			}
		})
	}
}
