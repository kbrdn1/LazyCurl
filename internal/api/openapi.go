package api

import (
	"fmt"
	"os"
	"strings"

	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// ImportErrorType categorizes import errors.
type ImportErrorType int

const (
	// ErrFileNotFound indicates the spec file does not exist
	ErrFileNotFound ImportErrorType = iota

	// ErrFileUnreadable indicates the file cannot be read
	ErrFileUnreadable

	// ErrInvalidFormat indicates the file is not valid JSON or YAML
	ErrInvalidFormat

	// ErrUnsupportedVersion indicates OpenAPI 2.0 or unknown version
	ErrUnsupportedVersion

	// ErrValidationFailed indicates the spec failed validation
	ErrValidationFailed

	// ErrRefResolutionFailed indicates a $ref could not be resolved
	ErrRefResolutionFailed

	// ErrConversionFailed indicates an error during collection building
	ErrConversionFailed
)

// String returns a string representation of the error type.
func (t ImportErrorType) String() string {
	switch t {
	case ErrFileNotFound:
		return "file_not_found"
	case ErrFileUnreadable:
		return "file_unreadable"
	case ErrInvalidFormat:
		return "invalid_format"
	case ErrUnsupportedVersion:
		return "unsupported_version"
	case ErrValidationFailed:
		return "validation_failed"
	case ErrRefResolutionFailed:
		return "ref_resolution_failed"
	case ErrConversionFailed:
		return "conversion_failed"
	default:
		return "unknown"
	}
}

// ImportError is returned when import fails.
type ImportError struct {
	// Type categorizes the error
	Type ImportErrorType

	// Message is the user-friendly error description
	Message string

	// Details provides technical details for debugging
	Details string

	// Line is the source line number (1-based, 0 if unavailable)
	Line int

	// Column is the source column number (1-based, 0 if unavailable)
	Column int

	// Cause is the underlying error
	Cause error
}

// Error implements the error interface.
func (e *ImportError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s (line %d)", e.Message, e.Line)
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *ImportError) Unwrap() error {
	return e.Cause
}

// ImportOptions configures the import behavior.
type ImportOptions struct {
	// Name overrides the collection name (default: uses info.title)
	Name string

	// OutputPath specifies where to save the collection (default: .lazycurl/collections/<name>.json)
	OutputPath string

	// BaseURL overrides the server URL from the spec
	BaseURL string

	// IncludeExamples enables example generation for request bodies
	IncludeExamples bool

	// CreateEnvironment creates an environment with server variables
	CreateEnvironment bool
}

// FolderPreview describes a folder to be created from a tag.
type FolderPreview struct {
	// Name is the tag/folder name
	Name string

	// Description is the tag description
	Description string

	// RequestCount is the number of operations with this tag
	RequestCount int
}

// ImportPreview provides statistics about an import before execution.
type ImportPreview struct {
	// SpecVersion is the OpenAPI version (e.g., "3.0.3", "3.1.0")
	SpecVersion string

	// Title is the API title from info.title
	Title string

	// Description is the API description from info.description
	Description string

	// EndpointCount is the total number of operations
	EndpointCount int

	// FolderCount is the number of tags (folders to create)
	FolderCount int

	// Folders lists tag names and their operation counts
	Folders []FolderPreview

	// Servers lists available server URLs
	Servers []string

	// Warnings lists non-fatal issues encountered during parsing
	Warnings []string
}

// OpenAPIImporter handles OpenAPI spec parsing and conversion
type OpenAPIImporter struct {
	doc     libopenapi.Document
	model   *libopenapi.DocumentModel[v3.Document]
	rawData []byte
}

// NewOpenAPIImporter creates an importer from file data
func NewOpenAPIImporter(data []byte) (*OpenAPIImporter, error) {
	doc, err := libopenapi.NewDocument(data)
	if err != nil {
		return nil, &ImportError{
			Type:    ErrInvalidFormat,
			Message: "Invalid OpenAPI specification",
			Details: err.Error(),
			Cause:   err,
		}
	}

	return &OpenAPIImporter{
		doc:     doc,
		rawData: data,
	}, nil
}

// NewOpenAPIImporterFromFile creates an importer from a file path
func NewOpenAPIImporterFromFile(filePath string) (*OpenAPIImporter, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &ImportError{
				Type:    ErrFileNotFound,
				Message: fmt.Sprintf("File not found: %s", filePath),
				Cause:   err,
			}
		}
		return nil, &ImportError{
			Type:    ErrFileUnreadable,
			Message: fmt.Sprintf("Cannot read file: %s", filePath),
			Details: err.Error(),
			Cause:   err,
		}
	}

	return NewOpenAPIImporter(data)
}

// GetVersion returns the OpenAPI version (e.g., "3.0.3", "3.1.0")
func (i *OpenAPIImporter) GetVersion() string {
	return i.doc.GetVersion()
}

// GetSpecInfo returns the spec info (title, description)
func (i *OpenAPIImporter) GetSpecInfo() (title, description string) {
	model, err := i.doc.BuildV3Model()
	if err != nil || model == nil {
		return "", ""
	}
	if model.Model.Info != nil {
		return model.Model.Info.Title, model.Model.Info.Description
	}
	return "", ""
}

// ValidateVersion checks if the OpenAPI version is supported (3.0.x or 3.1.x)
func (i *OpenAPIImporter) ValidateVersion() error {
	version := i.GetVersion()

	// Check for OpenAPI 2.0 (Swagger)
	if strings.HasPrefix(version, "2.") {
		return &ImportError{
			Type:    ErrUnsupportedVersion,
			Message: "OpenAPI 2.0 (Swagger) is not supported. Please convert to OpenAPI 3.x.",
			Details: fmt.Sprintf("Detected version: %s", version),
		}
	}

	// Check for supported 3.x versions
	if !strings.HasPrefix(version, "3.0") && !strings.HasPrefix(version, "3.1") {
		return &ImportError{
			Type:    ErrUnsupportedVersion,
			Message: fmt.Sprintf("Unsupported OpenAPI version: %s", version),
			Details: "Supported versions: 3.0.x, 3.1.x",
		}
	}

	return nil
}

// BuildV3Model builds and caches the V3 model
func (i *OpenAPIImporter) BuildV3Model() (*libopenapi.DocumentModel[v3.Document], error) {
	if i.model != nil {
		return i.model, nil
	}

	model, err := i.doc.BuildV3Model()
	if err != nil {
		return nil, &ImportError{
			Type:    ErrValidationFailed,
			Message: "Failed to parse OpenAPI specification",
			Details: err.Error(),
			Cause:   err,
		}
	}

	if model == nil {
		return nil, &ImportError{
			Type:    ErrValidationFailed,
			Message: "Failed to build OpenAPI model",
			Details: "The document could not be parsed as a valid OpenAPI 3.x specification",
		}
	}

	i.model = model
	return model, nil
}

// Preview returns import statistics without creating collection
func (i *OpenAPIImporter) Preview() (*ImportPreview, error) {
	if err := i.ValidateVersion(); err != nil {
		return nil, err
	}

	model, err := i.BuildV3Model()
	if err != nil {
		return nil, err
	}

	preview := &ImportPreview{
		SpecVersion: i.GetVersion(),
		Warnings:    []string{},
	}

	// Extract info
	if model.Model.Info != nil {
		preview.Title = model.Model.Info.Title
		preview.Description = model.Model.Info.Description
	}

	if preview.Title == "" {
		preview.Warnings = append(preview.Warnings, "Missing info.title, will use filename as collection name")
	}

	// Extract servers
	if model.Model.Servers != nil {
		for _, server := range model.Model.Servers {
			if server.URL != "" {
				preview.Servers = append(preview.Servers, server.URL)
			}
		}
	}

	if len(preview.Servers) == 0 {
		preview.Warnings = append(preview.Warnings, "No servers defined, requests will use relative URLs")
	}

	// Count endpoints and gather tag statistics
	tagCounts := make(map[string]int)
	tagDescriptions := make(map[string]string)
	hasUntagged := false

	// Extract tag descriptions from top-level tags
	if model.Model.Tags != nil {
		for _, tag := range model.Model.Tags {
			tagDescriptions[tag.Name] = tag.Description
		}
	}

	// Count operations by tag
	if model.Model.Paths != nil {
		for pair := model.Model.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
			pathItem := pair.Value()

			// Count each HTTP method as an endpoint
			operations := getOperationsFromPathItem(pathItem)
			for _, op := range operations {
				preview.EndpointCount++

				if len(op.Tags) == 0 {
					hasUntagged = true
					tagCounts["Untagged"]++
				} else {
					// Use first tag for folder organization
					tag := op.Tags[0]
					tagCounts[tag]++
				}
			}
		}
	}

	// Build folder previews
	for tag, count := range tagCounts {
		preview.Folders = append(preview.Folders, FolderPreview{
			Name:         tag,
			Description:  tagDescriptions[tag],
			RequestCount: count,
		})
	}
	preview.FolderCount = len(preview.Folders)

	if hasUntagged {
		preview.Warnings = append(preview.Warnings, "Some operations have no tags and will be placed in 'Untagged' folder")
	}

	return preview, nil
}

// ToCollection converts the spec to a LazyCurl collection
func (i *OpenAPIImporter) ToCollection(opts ImportOptions) (*CollectionFile, error) {
	if err := i.ValidateVersion(); err != nil {
		return nil, err
	}

	model, err := i.BuildV3Model()
	if err != nil {
		return nil, err
	}

	// Determine collection name
	name := opts.Name
	if name == "" && model.Model.Info != nil {
		name = model.Model.Info.Title
	}
	if name == "" {
		name = "Imported API"
	}

	// Determine base URL
	baseURL := opts.BaseURL
	if baseURL == "" && model.Model.Servers != nil && len(model.Model.Servers) > 0 {
		baseURL = model.Model.Servers[0].URL
	}

	// Create collection
	collection := &CollectionFile{
		Name: name,
	}

	if model.Model.Info != nil {
		collection.Description = model.Model.Info.Description
	}

	// Convert paths to requests organized by tags
	if model.Model.Paths != nil {
		collection.Folders = convertPathsToFolders(model.Model.Paths, &model.Model, baseURL, opts.IncludeExamples)
	}

	return collection, nil
}

// getOperationsFromPathItem extracts all operations from a path item
func getOperationsFromPathItem(pathItem *v3.PathItem) []*v3.Operation {
	var operations []*v3.Operation

	if pathItem.Get != nil {
		operations = append(operations, pathItem.Get)
	}
	if pathItem.Post != nil {
		operations = append(operations, pathItem.Post)
	}
	if pathItem.Put != nil {
		operations = append(operations, pathItem.Put)
	}
	if pathItem.Delete != nil {
		operations = append(operations, pathItem.Delete)
	}
	if pathItem.Patch != nil {
		operations = append(operations, pathItem.Patch)
	}
	if pathItem.Head != nil {
		operations = append(operations, pathItem.Head)
	}
	if pathItem.Options != nil {
		operations = append(operations, pathItem.Options)
	}

	return operations
}
