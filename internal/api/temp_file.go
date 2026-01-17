package api

import (
	"os"
	"time"
)

// TempFileInfo holds metadata about a temporary file
type TempFileInfo struct {
	// Path is the absolute path to the temp file
	Path string

	// OriginalContent is the content before editing
	// Used for comparison and recovery
	OriginalContent string

	// ContentType is the detected content type
	ContentType ContentType

	// Extension is the file extension (with dot)
	Extension string

	// CreatedAt is when the file was created
	CreatedAt time.Time
}

// CreateTempFile creates a temporary file with the given content.
// The file extension is determined by the content type.
//
// The temp file is created in the system temp directory with
// the prefix "lazycurl-" for identification.
//
// Caller is responsible for cleanup via CleanupTempFile.
func CreateTempFile(content string, contentType ContentType) (*TempFileInfo, error) {
	extension := GetExtensionForContentType(contentType)

	// Create temp file with appropriate extension
	tmpFile, err := os.CreateTemp("", "lazycurl-*"+extension)
	if err != nil {
		return nil, err
	}

	// Write content to file
	if _, err := tmpFile.WriteString(content); err != nil {
		_ = tmpFile.Close()           // Best effort cleanup
		_ = os.Remove(tmpFile.Name()) // Best effort cleanup
		return nil, err
	}

	// Close file handle (required for some editors)
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpFile.Name()) // Best effort cleanup
		return nil, err
	}

	return &TempFileInfo{
		Path:            tmpFile.Name(),
		OriginalContent: content,
		ContentType:     contentType,
		Extension:       extension,
		CreatedAt:       time.Now(),
	}, nil
}

// ReadTempFile reads the current content of a temp file.
//
// Returns the content as string, or error if file is unreadable.
func ReadTempFile(info *TempFileInfo) (string, error) {
	content, err := os.ReadFile(info.Path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// CleanupTempFile removes the temporary file.
// Safe to call multiple times; ignores "file not found" errors.
func CleanupTempFile(info *TempFileInfo) error {
	err := os.Remove(info.Path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// HasContentChanged compares original content with current file content.
// Returns true if content has been modified.
func HasContentChanged(info *TempFileInfo) (bool, error) {
	currentContent, err := ReadTempFile(info)
	if err != nil {
		return false, err
	}
	return currentContent != info.OriginalContent, nil
}
