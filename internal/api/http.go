package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// HTTPMethod represents HTTP request methods
type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	PATCH  HTTPMethod = "PATCH"
	DELETE HTTPMethod = "DELETE"
	HEAD   HTTPMethod = "HEAD"
	OPTIONS HTTPMethod = "OPTIONS"
)

// Request represents an HTTP request
type Request struct {
	Method  HTTPMethod
	URL     string
	Headers map[string]string
	Body    interface{}
	Timeout time.Duration
}

// Response represents an HTTP response
type Response struct {
	StatusCode int
	Status     string
	Headers    map[string][]string
	Body       string
	Time       time.Duration
	Size       int64
}

// Client handles HTTP requests
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new HTTP client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Send sends an HTTP request and returns the response
func (c *Client) Send(req *Request) (*Response, error) {
	start := time.Now()

	// Prepare body
	var bodyReader io.Reader
	if req.Body != nil {
		jsonBody, err := json.Marshal(req.Body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(string(req.Method), req.URL, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Set default Content-Type if body exists and not set
	if req.Body != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Send request
	if req.Timeout > 0 {
		c.httpClient.Timeout = req.Timeout
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	elapsed := time.Since(start)

	return &Response{
		StatusCode: httpResp.StatusCode,
		Status:     httpResp.Status,
		Headers:    httpResp.Header,
		Body:       string(bodyBytes),
		Time:       elapsed,
		Size:       int64(len(bodyBytes)),
	}, nil
}

// Collection represents a collection of requests
type Collection struct {
	Name        string
	Description string
	Requests    []*Request
}

// Environment represents environment variables for requests
type Environment struct {
	Name      string
	Variables map[string]string
}
