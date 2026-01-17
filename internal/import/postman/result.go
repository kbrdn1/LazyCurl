package postman

import (
	"fmt"
	"strings"

	"github.com/kbrdn1/LazyCurl/internal/api"
)

// ImportResult represents the result of an import operation.
type ImportResult struct {
	Collection  *api.CollectionFile  // Non-nil if collection import succeeded
	Environment *api.EnvironmentFile // Non-nil if environment import succeeded
	Summary     ImportSummary
	Error       error
}

// Success returns true if the import succeeded without errors.
func (r *ImportResult) Success() bool {
	return r.Error == nil && (r.Collection != nil || r.Environment != nil)
}

// HasWarnings returns true if there are warnings in the summary.
func (r *ImportResult) HasWarnings() bool {
	return len(r.Summary.Warnings) > 0
}

// FormatSummary returns a human-readable summary string.
func (r *ImportResult) FormatSummary() string {
	if r.Collection != nil {
		return r.Summary.FormatCollectionSummary()
	}
	if r.Environment != nil {
		return r.Summary.FormatEnvironmentSummary()
	}
	if r.Error != nil {
		return fmt.Sprintf("Import failed: %s", r.Error)
	}
	return "No import performed"
}

// ImportSummary contains statistics and messages from an import operation.
type ImportSummary struct {
	CollectionName  string
	RequestsCount   int
	FoldersCount    int
	EnvironmentName string
	VariablesCount  int
	Warnings        []string
	Errors          []string
}

// AddWarningf adds a warning message to the summary.
func (s *ImportSummary) AddWarningf(format string, args ...interface{}) {
	s.Warnings = append(s.Warnings, fmt.Sprintf(format, args...))
}

// AddErrorf adds an error message to the summary.
func (s *ImportSummary) AddErrorf(format string, args ...interface{}) {
	s.Errors = append(s.Errors, fmt.Sprintf(format, args...))
}

// FormatCollectionSummary returns a formatted string for collection import.
func (s *ImportSummary) FormatCollectionSummary() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Imported \"%s\"", s.CollectionName))

	var stats []string
	if s.RequestsCount > 0 {
		stats = append(stats, fmt.Sprintf("%d requests", s.RequestsCount))
	}
	if s.FoldersCount > 0 {
		stats = append(stats, fmt.Sprintf("%d folders", s.FoldersCount))
	}
	if len(stats) > 0 {
		parts = append(parts, strings.Join(stats, ", "))
	}

	if len(s.Warnings) > 0 {
		parts = append(parts, fmt.Sprintf("%d warnings", len(s.Warnings)))
	}

	return strings.Join(parts, " - ")
}

// FormatEnvironmentSummary returns a formatted string for environment import.
func (s *ImportSummary) FormatEnvironmentSummary() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Imported \"%s\"", s.EnvironmentName))

	if s.VariablesCount > 0 {
		parts = append(parts, fmt.Sprintf("%d variables", s.VariablesCount))
	}

	if len(s.Warnings) > 0 {
		parts = append(parts, fmt.Sprintf("%d warnings", len(s.Warnings)))
	}

	return strings.Join(parts, " - ")
}

// FileType represents the detected type of a Postman file.
type FileType int

const (
	// FileTypeUnknown indicates the file type could not be determined.
	FileTypeUnknown FileType = iota
	// FileTypeCollection indicates a Postman Collection v2.1 file.
	FileTypeCollection
	// FileTypeEnvironment indicates a Postman Environment file.
	FileTypeEnvironment
)

// String returns a string representation of the FileType.
func (t FileType) String() string {
	switch t {
	case FileTypeCollection:
		return "Collection"
	case FileTypeEnvironment:
		return "Environment"
	default:
		return "Unknown"
	}
}
