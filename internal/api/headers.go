package api

import (
	"fmt"
	"sort"
	"strings"
)

// HeaderEntry represents a single HTTP header
type HeaderEntry struct {
	Key   string
	Value string
}

// HeaderList is an ordered list of headers
type HeaderList []HeaderEntry

// HeadersToText converts a headers map to text format.
// Each header is formatted as "Key: Value" on its own line.
// Headers are sorted alphabetically by key for consistency.
func HeadersToText(headers map[string]string) string {
	if len(headers) == 0 {
		return ""
	}

	// Extract and sort keys for deterministic output
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build lines in sorted key order
	var lines []string
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %s", key, headers[key]))
	}
	return strings.Join(lines, "\n")
}

// TextToHeaders parses text format back to headers map.
// Lines without ": " separator are ignored.
// Whitespace is trimmed from keys and values.
func TextToHeaders(text string) map[string]string {
	headers := make(map[string]string)

	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split on first ": " (colon + space)
		if idx := strings.Index(line, ": "); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.TrimSpace(line[idx+2:])
			if key != "" {
				headers[key] = value
			}
		}
	}

	return headers
}

// ValidateHeaderText checks if text is valid header format.
// Returns list of warnings for lines that couldn't be parsed.
func ValidateHeaderText(text string) []string {
	var warnings []string

	for i, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line has valid format
		if !strings.Contains(line, ": ") {
			warnings = append(warnings, fmt.Sprintf("line %d: missing ': ' separator", i+1))
		} else {
			idx := strings.Index(line, ": ")
			if idx == 0 {
				warnings = append(warnings, fmt.Sprintf("line %d: empty header key", i+1))
			}
		}
	}

	return warnings
}
