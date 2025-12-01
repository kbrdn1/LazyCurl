package format

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
)

// ContentType represents the detected content type
type ContentType string

const (
	ContentTypeJSON       ContentType = "json"
	ContentTypeXML        ContentType = "xml"
	ContentTypeHTML       ContentType = "html"
	ContentTypeText       ContentType = "text"
	ContentTypeJavaScript ContentType = "javascript"
	ContentTypeUnknown    ContentType = "unknown"
)

// DetectContentType attempts to detect the content type from headers and content
func DetectContentType(contentType string, body []byte) ContentType {
	// Check Content-Type header first
	contentTypeLower := strings.ToLower(contentType)

	if strings.Contains(contentTypeLower, "application/json") || strings.Contains(contentTypeLower, "text/json") {
		return ContentTypeJSON
	}
	if strings.Contains(contentTypeLower, "application/xml") || strings.Contains(contentTypeLower, "text/xml") {
		return ContentTypeXML
	}
	if strings.Contains(contentTypeLower, "text/html") {
		return ContentTypeHTML
	}
	if strings.Contains(contentTypeLower, "application/javascript") || strings.Contains(contentTypeLower, "text/javascript") {
		return ContentTypeJavaScript
	}
	if strings.Contains(contentTypeLower, "text/") {
		return ContentTypeText
	}

	// Try to detect from content
	bodyStr := string(body)
	trimmed := strings.TrimSpace(bodyStr)

	if len(trimmed) == 0 {
		return ContentTypeText
	}

	// Check for JSON
	if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
		return ContentTypeJSON
	}

	// Check for XML
	if strings.HasPrefix(trimmed, "<?xml") || strings.HasPrefix(trimmed, "<") {
		return ContentTypeXML
	}

	// Check for HTML
	if strings.Contains(strings.ToLower(trimmed), "<!doctype html") ||
		strings.Contains(strings.ToLower(trimmed), "<html") {
		return ContentTypeHTML
	}

	return ContentTypeText
}

// FormatJSON formats JSON with proper indentation
func FormatJSON(data []byte, indent string) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	formatted, err := json.MarshalIndent(parsed, "", indent)
	if err != nil {
		return "", fmt.Errorf("failed to format JSON: %w", err)
	}

	return string(formatted), nil
}

// FormatXML formats XML with proper indentation
func FormatXML(data []byte, indent string) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	var parsed interface{}
	if err := xml.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("invalid XML: %w", err)
	}

	formatted, err := xml.MarshalIndent(parsed, "", indent)
	if err != nil {
		return "", fmt.Errorf("failed to format XML: %w", err)
	}

	return string(formatted), nil
}

// Format automatically detects content type and formats accordingly
func Format(contentType string, body []byte) (string, ContentType, error) {
	detected := DetectContentType(contentType, body)

	switch detected {
	case ContentTypeJSON:
		formatted, err := FormatJSON(body, "  ")
		if err != nil {
			// If JSON parsing fails, return as text
			return string(body), ContentTypeText, nil
		}
		return formatted, ContentTypeJSON, nil

	case ContentTypeXML:
		formatted, err := FormatXML(body, "  ")
		if err != nil {
			// If XML parsing fails, return as text
			return string(body), ContentTypeText, nil
		}
		return formatted, ContentTypeXML, nil

	default:
		return string(body), detected, nil
	}
}

// MinifyJSON removes all unnecessary whitespace from JSON
func MinifyJSON(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	minified, err := json.Marshal(parsed)
	if err != nil {
		return "", fmt.Errorf("failed to minify JSON: %w", err)
	}

	return string(minified), nil
}

// ValidateJSON checks if the data is valid JSON
func ValidateJSON(data []byte) error {
	var parsed interface{}
	return json.Unmarshal(data, &parsed)
}

// PrettyPrint formats the response body for display in the TUI
func PrettyPrint(contentType string, body []byte, maxLength int) string {
	if len(body) == 0 {
		return "(empty response)"
	}

	formatted, detectedType, err := Format(contentType, body)
	if err != nil {
		return string(body)
	}

	// Truncate if too long
	if maxLength > 0 && len(formatted) > maxLength {
		formatted = formatted[:maxLength] + "\n\n... (truncated)"
	}

	// Add type indicator
	typeIndicator := ""
	switch detectedType {
	case ContentTypeJSON:
		typeIndicator = "📄 JSON"
	case ContentTypeXML:
		typeIndicator = "📄 XML"
	case ContentTypeHTML:
		typeIndicator = "🌐 HTML"
	case ContentTypeJavaScript:
		typeIndicator = "📜 JavaScript"
	case ContentTypeText:
		typeIndicator = "📝 Text"
	default:
		typeIndicator = "📦 Binary"
	}

	return fmt.Sprintf("%s\n\n%s", typeIndicator, formatted)
}
