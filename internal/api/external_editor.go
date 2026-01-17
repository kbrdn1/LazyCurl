package api

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
)

// EditableField represents a field that can be edited externally
type EditableField string

const (
	EditableFieldBody    EditableField = "body"
	EditableFieldHeaders EditableField = "headers"
)

// EditorSource indicates the origin of editor configuration
type EditorSource string

const (
	EditorSourceVisual   EditorSource = "VISUAL"
	EditorSourceEditor   EditorSource = "EDITOR"
	EditorSourceFallback EditorSource = "fallback"
)

// ContentType represents detected content format
type ContentType string

const (
	ContentTypeJSON ContentType = "json"
	ContentTypeXML  ContentType = "xml"
	ContentTypeHTML ContentType = "html"
	ContentTypeText ContentType = "text"
)

// EditorConfig holds the parsed editor command configuration
type EditorConfig struct {
	// Binary is the path to the editor executable
	Binary string

	// Args are additional arguments to pass to the editor
	// (e.g., ["--wait"] for VS Code)
	Args []string

	// Source indicates where the config came from
	Source EditorSource
}

// ContentTypeExtensions maps content types to file extensions
var ContentTypeExtensions = map[ContentType]string{
	ContentTypeJSON: ".json",
	ContentTypeXML:  ".xml",
	ContentTypeHTML: ".html",
	ContentTypeText: ".txt",
}

// ErrNoEditorAvailable is returned when no editor can be found
var ErrNoEditorAvailable = errors.New("no editor available: set $EDITOR or $VISUAL environment variable")

// ErrEditorNotFound is returned when the configured editor binary doesn't exist
var ErrEditorNotFound = errors.New("editor not found in PATH")

// parseEditorCommand splits an editor command string into binary and args
// Handles commands like "code --wait" or "vim"
func parseEditorCommand(cmd string) []string {
	return strings.Fields(cmd)
}

// GetEditorConfig returns the resolved editor configuration
// by checking $VISUAL, $EDITOR, and fallback editors.
//
// Detection order:
// 1. $VISUAL environment variable
// 2. $EDITOR environment variable
// 3. Fallback: nano, vi (first available)
func GetEditorConfig() (*EditorConfig, error) {
	// Check $VISUAL first (preferred for graphical editors)
	if visual := os.Getenv("VISUAL"); visual != "" {
		parts := parseEditorCommand(visual)
		if len(parts) > 0 {
			return &EditorConfig{
				Binary: parts[0],
				Args:   parts[1:],
				Source: EditorSourceVisual,
			}, nil
		}
	}

	// Fall back to $EDITOR
	if editor := os.Getenv("EDITOR"); editor != "" {
		parts := parseEditorCommand(editor)
		if len(parts) > 0 {
			return &EditorConfig{
				Binary: parts[0],
				Args:   parts[1:],
				Source: EditorSourceEditor,
			}, nil
		}
	}

	// Fallback chain: nano, vi
	fallbacks := []string{"nano", "vi"}
	for _, fallback := range fallbacks {
		if path, err := exec.LookPath(fallback); err == nil {
			return &EditorConfig{
				Binary: path,
				Args:   nil,
				Source: EditorSourceFallback,
			}, nil
		}
	}

	return nil, ErrNoEditorAvailable
}

// Validate checks if the editor binary exists and is executable
func (ec *EditorConfig) Validate() error {
	if ec.Binary == "" {
		return errors.New("editor binary path is required")
	}

	// Check if binary exists and is executable
	if _, err := exec.LookPath(ec.Binary); err != nil {
		return ErrEditorNotFound
	}

	return nil
}

// DetectContentType analyzes content and returns its type.
// Uses heuristics to detect JSON, XML, HTML, or plain text.
func DetectContentType(content string) ContentType {
	trimmed := strings.TrimSpace(content)

	// Empty content defaults to text
	if trimmed == "" {
		return ContentTypeText
	}

	// JSON detection (starts with { or [)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		var js json.RawMessage
		if json.Unmarshal([]byte(trimmed), &js) == nil {
			return ContentTypeJSON
		}
	}

	// XML detection
	if strings.HasPrefix(trimmed, "<?xml") ||
		(strings.HasPrefix(trimmed, "<") &&
			strings.Contains(trimmed, "</") &&
			!strings.Contains(strings.ToLower(trimmed), "<html")) {
		return ContentTypeXML
	}

	// HTML detection
	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, "<!doctype html") ||
		strings.HasPrefix(lower, "<html") {
		return ContentTypeHTML
	}

	return ContentTypeText
}

// GetExtensionForContentType returns the file extension
// (with leading dot) for the given content type.
func GetExtensionForContentType(ct ContentType) string {
	if ext, ok := ContentTypeExtensions[ct]; ok {
		return ext
	}
	return ".txt"
}
