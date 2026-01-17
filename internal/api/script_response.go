package api

import (
	"net/http"
	"strings"
	"time"
)

// ScriptResponse represents immutable response data for scripts
type ScriptResponse struct {
	status     int
	statusText string
	headers    map[string]string
	body       string
	time       int64 // Response time in milliseconds
}

// NewScriptResponse creates a ScriptResponse from an HTTP response
func NewScriptResponse(resp *http.Response, body string, duration time.Duration) *ScriptResponse {
	if resp == nil {
		return &ScriptResponse{
			status:     0,
			statusText: "",
			headers:    make(map[string]string),
			body:       "",
			time:       0,
		}
	}

	// Convert headers to map (single value per header)
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &ScriptResponse{
		status:     resp.StatusCode,
		statusText: resp.Status,
		headers:    headers,
		body:       body,
		time:       duration.Milliseconds(),
	}
}

// NewScriptResponseFromData creates a ScriptResponse from raw data
func NewScriptResponseFromData(status int, statusText string, headers map[string]string, body string, timeMs int64) *ScriptResponse {
	if headers == nil {
		headers = make(map[string]string)
	}
	return &ScriptResponse{
		status:     status,
		statusText: statusText,
		headers:    headers,
		body:       body,
		time:       timeMs,
	}
}

// Status returns the HTTP status code
func (r *ScriptResponse) Status() int {
	return r.status
}

// StatusText returns the full status text (e.g., "200 OK")
func (r *ScriptResponse) StatusText() string {
	return r.statusText
}

// Time returns the response time in milliseconds
func (r *ScriptResponse) Time() int64 {
	return r.time
}

// Body returns the response body as string
func (r *ScriptResponse) Body() string {
	return r.body
}

// GetHeader returns a header value (case-insensitive)
func (r *ScriptResponse) GetHeader(name string) string {
	// Case-insensitive header lookup
	nameLower := strings.ToLower(name)
	for k, v := range r.headers {
		if strings.ToLower(k) == nameLower {
			return v
		}
	}
	return ""
}

// Headers returns a copy of all headers
func (r *ScriptResponse) Headers() map[string]string {
	result := make(map[string]string, len(r.headers))
	for k, v := range r.headers {
		result[k] = v
	}
	return result
}
