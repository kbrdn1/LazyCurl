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
