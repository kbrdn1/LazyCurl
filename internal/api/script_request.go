package api

import (
	"strings"
)

// ScriptRequest represents mutable request data for scripts
type ScriptRequest struct {
	name     string
	method   string
	url      string
	headers  map[string]string
	body     string
	modified bool
}

// NewScriptRequest creates a ScriptRequest from a CollectionRequest
func NewScriptRequest(req *CollectionRequest) *ScriptRequest {
	if req == nil {
		return &ScriptRequest{
			name:    "",
			method:  "GET",
			url:     "",
			headers: make(map[string]string),
			body:    "",
		}
	}

	// Convert headers from KeyValueEntry slice to map
	headers := make(map[string]string)
	for _, h := range req.Headers {
		if h.Enabled {
			headers[h.Key] = h.Value
		}
	}
	// Also include legacy headers if present
	for k, v := range req.HeadersMap {
		headers[k] = v
	}

	// Extract body content
	body := ""
	if req.Body != nil && req.Body.Content != nil {
		switch v := req.Body.Content.(type) {
		case string:
			body = v
		}
	}

	return &ScriptRequest{
		name:     req.Name,
		method:   string(req.Method),
		url:      req.URL,
		headers:  headers,
		body:     body,
		modified: false,
	}
}

// NewScriptRequestFromHTTP creates a ScriptRequest from an HTTP Request
func NewScriptRequestFromHTTP(req *Request) *ScriptRequest {
	if req == nil {
		return &ScriptRequest{
			name:    "",
			method:  "GET",
			url:     "",
			headers: make(map[string]string),
			body:    "",
		}
	}

	// Copy headers
	headers := make(map[string]string)
	for k, v := range req.Headers {
		headers[k] = v
	}

	// Extract body content
	body := ""
	if req.Body != nil {
		switch v := req.Body.(type) {
		case string:
			body = v
		}
	}

	return &ScriptRequest{
		name:     "",
		method:   string(req.Method),
		url:      req.URL,
		headers:  headers,
		body:     body,
		modified: false,
	}
}

// Name returns the request name
func (r *ScriptRequest) Name() string {
	return r.name
}

// Method returns the HTTP method (readonly)
func (r *ScriptRequest) Method() string {
	return r.method
}

// URL returns the request URL
func (r *ScriptRequest) URL() string {
	return r.url
}

// SetURL sets the request URL and marks as modified
func (r *ScriptRequest) SetURL(url string) {
	r.url = url
	r.modified = true
}

// Body returns the request body
func (r *ScriptRequest) Body() string {
	return r.body
}

// SetBody sets the request body and marks as modified
func (r *ScriptRequest) SetBody(body string) {
	r.body = body
	r.modified = true
}

// GetHeader returns a header value (case-insensitive)
func (r *ScriptRequest) GetHeader(name string) string {
	// Case-insensitive header lookup
	nameLower := strings.ToLower(name)
	for k, v := range r.headers {
		if strings.ToLower(k) == nameLower {
			return v
		}
	}
	return ""
}

// SetHeader sets or updates a header value and marks as modified
func (r *ScriptRequest) SetHeader(name, value string) {
	// Remove existing header with same name (case-insensitive)
	nameLower := strings.ToLower(name)
	for k := range r.headers {
		if strings.ToLower(k) == nameLower {
			delete(r.headers, k)
		}
	}
	r.headers[name] = value
	r.modified = true
}

// RemoveHeader removes a header (case-insensitive) and marks as modified
func (r *ScriptRequest) RemoveHeader(name string) {
	nameLower := strings.ToLower(name)
	for k := range r.headers {
		if strings.ToLower(k) == nameLower {
			delete(r.headers, k)
			r.modified = true
			return
		}
	}
}

// Headers returns a copy of all headers
func (r *ScriptRequest) Headers() map[string]string {
	result := make(map[string]string, len(r.headers))
	for k, v := range r.headers {
		result[k] = v
	}
	return result
}

// IsModified returns true if the request was modified by a script
func (r *ScriptRequest) IsModified() bool {
	return r.modified
}

// ApplyTo applies the modifications to a CollectionRequest
func (r *ScriptRequest) ApplyTo(req *CollectionRequest) {
	if req == nil || !r.modified {
		return
	}

	// Update URL
	req.URL = r.url

	// Update headers - convert map back to KeyValueEntry slice
	req.Headers = make([]KeyValueEntry, 0, len(r.headers))
	for k, v := range r.headers {
		req.Headers = append(req.Headers, KeyValueEntry{
			Key:     k,
			Value:   v,
			Enabled: true,
		})
	}

	// Update body if we have body config
	if r.body != "" {
		if req.Body == nil {
			req.Body = &BodyConfig{
				Type:    "raw",
				Content: r.body,
			}
		} else {
			req.Body.Content = r.body
		}
	}
}
