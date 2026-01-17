package postman

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DetectFileType determines if a file is a Postman collection or environment.
func DetectFileType(filePath string) (FileType, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return FileTypeUnknown, fmt.Errorf("failed to read file: %w", err)
	}
	return DetectFileTypeFromBytes(data), nil
}

// DetectFileTypeFromBytes determines the file type from raw JSON bytes.
func DetectFileTypeFromBytes(data []byte) FileType {
	// Parse minimally to detect type
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return FileTypeUnknown
	}

	// Check for collection: info.schema containing "collection/v2"
	if infoRaw, ok := raw["info"]; ok {
		var info struct {
			Schema string `json:"schema"`
		}
		if err := json.Unmarshal(infoRaw, &info); err == nil {
			if strings.Contains(info.Schema, "collection/v2") {
				return FileTypeCollection
			}
		}
	}

	// Check for environment: _postman_variable_scope = "environment"
	if scopeRaw, ok := raw["_postman_variable_scope"]; ok {
		var scope string
		if err := json.Unmarshal(scopeRaw, &scope); err == nil {
			if scope == "environment" {
				return FileTypeEnvironment
			}
		}
	}

	// Fallback: check for values array with key/value structure (environment pattern)
	if valuesRaw, ok := raw["values"]; ok {
		var values []map[string]interface{}
		if err := json.Unmarshal(valuesRaw, &values); err == nil {
			if len(values) > 0 {
				// Check if first item has "key" field (environment structure)
				if _, hasKey := values[0]["key"]; hasKey {
					return FileTypeEnvironment
				}
			}
			// Empty values array with "name" field is likely an environment
			if _, hasName := raw["name"]; hasName {
				return FileTypeEnvironment
			}
		}
	}

	return FileTypeUnknown
}
