package ui

import (
	"github.com/kbrdn1/LazyCurl/internal/api"
)

// CurlImportedMsg is sent when a cURL command is successfully imported
type CurlImportedMsg struct {
	Request *api.CollectionRequest
}

// CurlExportedMsg is sent when a request is exported to clipboard
type CurlExportedMsg struct {
	Success bool
	Error   error
}

// ShowImportModalMsg triggers the import modal to open
type ShowImportModalMsg struct{}

// HideImportModalMsg triggers the import modal to close
type HideImportModalMsg struct{}

// OpenAPIImportedMsg is sent when an OpenAPI spec is successfully imported
type OpenAPIImportedMsg struct {
	Collection *api.CollectionFile
	Stats      OpenAPIImportStats
}

// OpenAPIImportStats contains import statistics
type OpenAPIImportStats struct {
	FolderCount  int
	RequestCount int
	WarningCount int
}

// ShowOpenAPIImportModalMsg triggers the OpenAPI import modal to open
type ShowOpenAPIImportModalMsg struct{}

// HideOpenAPIImportModalMsg triggers the OpenAPI import modal to close
type HideOpenAPIImportModalMsg struct{}

// OpenAPISpinnerTickMsg is sent to animate the import spinner
type OpenAPISpinnerTickMsg struct{}

// OpenAPIImportCompleteMsg is sent when async import completes
type OpenAPIImportCompleteMsg struct {
	Collection *api.CollectionFile
	Error      error
	SavePath   string
}
