// Package postman provides import and export functionality for Postman Collection v2.1 format.
//
// This package enables bidirectional conversion between Postman collections/environments
// and LazyCurl's internal formats. It supports:
//
//   - Importing Postman Collection v2.1 files
//   - Importing Postman Environment files
//   - Exporting LazyCurl collections to Postman format
//   - Exporting LazyCurl environments to Postman format
//   - Auto-detecting file types (collection vs environment)
//
// # Import Example
//
//	result, err := postman.ImportCollection("/path/to/collection.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if result.HasWarnings() {
//	    for _, w := range result.Summary.Warnings {
//	        log.Printf("Warning: %s", w)
//	    }
//	}
//	// Use result.Collection
//
// # Export Example
//
//	err := postman.ExportCollection(collection, "/path/to/export.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Supported Features
//
// The following Postman features are fully supported:
//   - Collections with nested folders (unlimited depth)
//   - All HTTP methods (GET, POST, PUT, PATCH, DELETE, etc.)
//   - Request headers with enabled/disabled state
//   - Query parameters
//   - Body types: raw (JSON, text, XML), urlencoded, formdata
//   - Authentication: Bearer, Basic, API Key
//   - Environment variables with secret/enabled flags
//
// # Unsupported Features
//
// The following Postman features generate warnings but don't prevent import:
//   - Pre-request scripts (stored but not executed)
//   - Test scripts (stored but not executed)
//   - OAuth 2.0 authentication (warning generated)
//   - GraphQL body type (imported as raw)
//   - File uploads (path preserved only)
package postman
