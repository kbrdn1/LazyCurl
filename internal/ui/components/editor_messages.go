package components

import (
	"time"

	"github.com/kbrdn1/LazyCurl/internal/api"
)

// EditorErrorType categorizes editor errors
type EditorErrorType string

const (
	EditorErrorNoEditor    EditorErrorType = "no_editor"
	EditorErrorNotFound    EditorErrorType = "not_found"
	EditorErrorTempFile    EditorErrorType = "temp_file"
	EditorErrorProcess     EditorErrorType = "process"
	EditorErrorReadContent EditorErrorType = "read_content"
)

// ExternalEditorRequestMsg requests opening an external editor
type ExternalEditorRequestMsg struct {
	// Field to edit (body or headers)
	Field api.EditableField

	// Content is the current content to edit
	Content string

	// ContentType is the detected or specified content type
	ContentType api.ContentType
}

// ExternalEditorFinishedMsg indicates editor has closed
type ExternalEditorFinishedMsg struct {
	// Field that was edited
	Field api.EditableField

	// Content is the new content (if successful)
	Content string

	// OriginalContent for comparison
	OriginalContent string

	// Changed indicates if content was modified
	Changed bool

	// Err is non-nil if an error occurred
	Err error

	// Duration is how long the editor was open
	Duration time.Duration
}

// ExternalEditorErrorMsg indicates editor could not be opened
type ExternalEditorErrorMsg struct {
	// Field that was being edited
	Field api.EditableField

	// Err describes what went wrong
	Err error

	// ErrorType categorizes the error
	ErrorType EditorErrorType
}
