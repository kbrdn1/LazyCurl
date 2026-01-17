package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CurlGeneratorOptions configures output format
type CurlGeneratorOptions struct {
	Multiline     bool   // Use backslash continuation
	IndentString  string // Indentation for multiline (default: "  ")
	QuoteStyle    string // "single" or "double" (default: "single")
	IncludeMethod bool   // Always include -X even for GET (default: false)
}

// DefaultGeneratorOptions returns sensible defaults
func DefaultGeneratorOptions() CurlGeneratorOptions {
	return CurlGeneratorOptions{
		Multiline:     false,
		IndentString:  "  ",
		QuoteStyle:    "single",
		IncludeMethod: false,
	}
}

// GenerateCurlCommand generates a valid cURL command from a CollectionRequest
func GenerateCurlCommand(req *CollectionRequest) string {
	return GenerateCurlCommandWithOptions(req, DefaultGeneratorOptions())
}

// GenerateCurlCommandWithOptions generates a cURL command with custom formatting
func GenerateCurlCommandWithOptions(req *CollectionRequest, opts CurlGeneratorOptions) string {
	if req == nil || req.URL == "" {
		return ""
	}

	var parts []string
	parts = append(parts, "curl")

	// Method (omit GET unless forced or body is present)
	includeMethod := opts.IncludeMethod || req.Method != GET
	if includeMethod {
		parts = append(parts, "-X", string(req.Method))
	}

	// Headers (enabled only)
	for _, h := range req.Headers {
		if h.Enabled && h.Key != "" {
			headerValue := fmt.Sprintf("%s: %s", h.Key, h.Value)
			parts = append(parts, "-H", quote(headerValue, opts.QuoteStyle))
		}
	}

	// Auth
	if req.Auth != nil && req.Auth.Type == "basic" && req.Auth.Username != "" {
		authValue := req.Auth.Username
		if req.Auth.Password != "" {
			authValue += ":" + req.Auth.Password
		}
		parts = append(parts, "-u", quote(authValue, opts.QuoteStyle))
	}

	// Body
	if req.Body != nil && req.Body.Content != nil {
		bodyStr := formatBody(req.Body)
		if bodyStr != "" {
			// Use appropriate flag based on body type
			// --data-raw for raw content, -d for json/form-urlencoded
			bodyFlag := "-d"
			if req.Body.Type == "raw" {
				bodyFlag = "--data-raw"
			}
			parts = append(parts, bodyFlag, quote(bodyStr, opts.QuoteStyle))
		}
	}

	// URL (always quoted, always last)
	parts = append(parts, quote(req.URL, opts.QuoteStyle))

	if opts.Multiline {
		return formatMultiline(parts, opts.IndentString)
	}
	return strings.Join(parts, " ")
}

// quote wraps value in quotes with proper escaping
func quote(value, style string) string {
	if style == "double" {
		// Escape backslashes first, then double quotes
		escaped := strings.ReplaceAll(value, "\\", "\\\\")
		escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
		return fmt.Sprintf("\"%s\"", escaped)
	}
	// Single quote style - escape single quotes using shell syntax: ' -> '\''
	escaped := strings.ReplaceAll(value, "'", "'\\''")
	return fmt.Sprintf("'%s'", escaped)
}

// formatBody serializes body content to string
func formatBody(body *BodyConfig) string {
	if body == nil || body.Content == nil {
		return ""
	}

	switch v := body.Content.(type) {
	case string:
		return v
	case map[string]interface{}, []interface{}:
		// JSON content - serialize it
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	default:
		// Try to marshal as JSON first
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	}
}

// formatMultiline formats parts with backslash continuation
func formatMultiline(parts []string, indent string) string {
	if len(parts) <= 1 {
		return strings.Join(parts, " ")
	}

	var lines []string

	// First line is "curl" with continuation
	lines = append(lines, parts[0]+" \\")

	// Process remaining parts in pairs (flag + value) or singles
	i := 1
	for i < len(parts) {
		part := parts[i]

		// Check if this is a flag that takes a value
		if strings.HasPrefix(part, "-") && i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "-") {
			// Flag with value - keep together
			line := indent + part + " " + parts[i+1]
			i += 2

			// Add continuation if not last
			if i < len(parts) {
				line += " \\"
			}
			lines = append(lines, line)
		} else {
			// Single part (URL or flag without value)
			line := indent + part
			i++

			// Add continuation if not last
			if i < len(parts) {
				line += " \\"
			}
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

// GenerateCurlFromRequest generates a cURL command from a Request (simpler type)
func GenerateCurlFromRequest(req *Request) string {
	if req == nil || req.URL == "" {
		return ""
	}

	var parts []string
	parts = append(parts, "curl")

	// Method
	if req.Method != GET {
		parts = append(parts, "-X", string(req.Method))
	}

	// Headers
	for key, value := range req.Headers {
		headerValue := fmt.Sprintf("%s: %s", key, value)
		parts = append(parts, "-H", quote(headerValue, "single"))
	}

	// Body
	if req.Body != nil {
		var bodyStr string
		switch v := req.Body.(type) {
		case string:
			bodyStr = v
		default:
			jsonBytes, err := json.Marshal(v)
			if err == nil {
				bodyStr = string(jsonBytes)
			}
		}
		if bodyStr != "" {
			parts = append(parts, "-d", quote(bodyStr, "single"))
		}
	}

	// URL
	parts = append(parts, quote(req.URL, "single"))

	return strings.Join(parts, " ")
}
