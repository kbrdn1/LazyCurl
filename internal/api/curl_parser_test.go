package api

import (
	"errors"
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "simple command",
			input:     "curl https://example.com",
			wantCount: 3, // curl, url, EOF
			wantErr:   false,
		},
		{
			name:      "with flag",
			input:     "curl -X POST https://example.com",
			wantCount: 5, // curl, -X, POST, url, EOF
			wantErr:   false,
		},
		{
			name:      "single quoted string",
			input:     "curl -H 'Content-Type: application/json' https://example.com",
			wantCount: 5, // curl, -H, string, url, EOF
			wantErr:   false,
		},
		{
			name:      "double quoted string",
			input:     `curl -d "{\"key\":\"value\"}" https://example.com`,
			wantCount: 5, // curl, -d, string, url, EOF
			wantErr:   false,
		},
		{
			name:      "unclosed single quote",
			input:     "curl -H 'Content-Type: application/json https://example.com",
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "unclosed double quote",
			input:     `curl -d "{\"key\":\"value\" https://example.com`,
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:      "long flag with equals",
			input:     "curl --request=POST https://example.com",
			wantCount: 6, // curl, --request, =, POST, url, EOF
			wantErr:   false,
		},
		{
			name:      "ANSI-C quoting",
			input:     "curl -d $'line1\\nline2' https://example.com",
			wantCount: 5, // curl, -d, string, url, EOF
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := tokenize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("tokenize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(tokens) != tt.wantCount {
				t.Errorf("tokenize() got %d tokens, want %d", len(tokens), tt.wantCount)
				for i, tok := range tokens {
					t.Logf("  token %d: type=%d value=%q", i, tok.Type, tok.Value)
				}
			}
		})
	}
}

func TestNormalizeMultiline(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single line",
			input: "curl https://example.com",
			want:  "curl https://example.com",
		},
		{
			name:  "backslash continuation",
			input: "curl \\\n  -X POST \\\n  https://example.com",
			want:  "curl -X POST https://example.com",
		},
		{
			name:  "backtick continuation",
			input: "curl `\n  -X POST `\n  https://example.com",
			want:  "curl -X POST https://example.com",
		},
		{
			name:  "windows line endings",
			input: "curl \\\r\n  -X POST \\\r\n  https://example.com",
			want:  "curl -X POST https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeMultiline(tt.input)
			if got != tt.want {
				t.Errorf("normalizeMultiline() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseCurlCommand(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantMethod HTTPMethod
		wantURL    string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "simple GET",
			input:      "curl https://example.com",
			wantMethod: GET,
			wantURL:    "https://example.com",
			wantErr:    false,
		},
		{
			name:       "explicit GET",
			input:      "curl -X GET https://example.com",
			wantMethod: GET,
			wantURL:    "https://example.com",
			wantErr:    false,
		},
		{
			name:       "POST method",
			input:      "curl -X POST https://example.com",
			wantMethod: POST,
			wantURL:    "https://example.com",
			wantErr:    false,
		},
		{
			name:       "PUT method",
			input:      "curl -X PUT https://example.com",
			wantMethod: PUT,
			wantURL:    "https://example.com",
			wantErr:    false,
		},
		{
			name:       "DELETE method",
			input:      "curl -X DELETE https://example.com",
			wantMethod: DELETE,
			wantURL:    "https://example.com",
			wantErr:    false,
		},
		{
			name:       "PATCH method",
			input:      "curl -X PATCH https://example.com",
			wantMethod: PATCH,
			wantURL:    "https://example.com",
			wantErr:    false,
		},
		{
			name:       "long form method",
			input:      "curl --request POST https://example.com",
			wantMethod: POST,
			wantURL:    "https://example.com",
			wantErr:    false,
		},
		{
			name:       "method with equals",
			input:      "curl --request=POST https://example.com",
			wantMethod: POST,
			wantURL:    "https://example.com",
			wantErr:    false,
		},
		{
			name:       "implicit POST with data",
			input:      "curl -d 'data' https://example.com",
			wantMethod: POST,
			wantURL:    "https://example.com",
			wantErr:    false,
		},
		{
			name:       "quoted URL",
			input:      "curl 'https://example.com/path?query=value'",
			wantMethod: GET,
			wantURL:    "https://example.com/path?query=value",
			wantErr:    false,
		},
		{
			name:       "multiline command",
			input:      "curl \\\n  -X POST \\\n  https://example.com",
			wantMethod: POST,
			wantURL:    "https://example.com",
			wantErr:    false,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
			errMsg:  "empty cURL command",
		},
		{
			name:    "whitespace only",
			input:   "   \n\t  ",
			wantErr: true,
			errMsg:  "empty cURL command",
		},
		{
			name:    "no curl command",
			input:   "wget https://example.com",
			wantErr: true,
			errMsg:  "command must start with 'curl'",
		},
		{
			name:    "no URL",
			input:   "curl -X POST",
			wantErr: true,
			errMsg:  "URL is required",
		},
		{
			name:    "unclosed quote",
			input:   "curl 'https://example.com",
			wantErr: true,
			errMsg:  "unclosed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseCurlCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCurlCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.errMsg != "" && err != nil {
					var parseErr *ParseError
					if errors.As(err, &parseErr) {
						if parseErr.Message != tt.errMsg && !strings.Contains(parseErr.Message, tt.errMsg) {
							t.Errorf("ParseCurlCommand() error message = %q, want containing %q", parseErr.Message, tt.errMsg)
						}
					}
				}
				return
			}
			if req.Method != tt.wantMethod {
				t.Errorf("ParseCurlCommand() method = %v, want %v", req.Method, tt.wantMethod)
			}
			if req.URL != tt.wantURL {
				t.Errorf("ParseCurlCommand() URL = %v, want %v", req.URL, tt.wantURL)
			}
		})
	}
}

func TestParseCurlCommandHeaders(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantHeaders map[string]string
	}{
		{
			name:  "single header",
			input: "curl -H 'Accept: application/json' https://example.com",
			wantHeaders: map[string]string{
				"Accept": "application/json",
			},
		},
		{
			name:  "multiple headers",
			input: "curl -H 'Accept: application/json' -H 'Content-Type: application/json' https://example.com",
			wantHeaders: map[string]string{
				"Accept":       "application/json",
				"Content-Type": "application/json",
			},
		},
		{
			name:  "long form header",
			input: "curl --header 'Authorization: Bearer token123' https://example.com",
			wantHeaders: map[string]string{
				"Authorization": "Bearer token123",
			},
		},
		{
			name:  "header with double quotes",
			input: `curl -H "Accept: application/json" https://example.com`,
			wantHeaders: map[string]string{
				"Accept": "application/json",
			},
		},
		{
			name:  "user agent flag",
			input: "curl -A 'MyAgent/1.0' https://example.com",
			wantHeaders: map[string]string{
				"User-Agent": "MyAgent/1.0",
			},
		},
		{
			name:  "cookie flag",
			input: "curl --cookie 'session=abc123' https://example.com",
			wantHeaders: map[string]string{
				"Cookie": "session=abc123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseCurlCommand(tt.input)
			if err != nil {
				t.Fatalf("ParseCurlCommand() error = %v", err)
			}

			for wantKey, wantValue := range tt.wantHeaders {
				found := false
				for _, h := range req.Headers {
					if h.Key == wantKey {
						found = true
						if h.Value != wantValue {
							t.Errorf("header %q = %q, want %q", wantKey, h.Value, wantValue)
						}
						break
					}
				}
				if !found {
					t.Errorf("header %q not found", wantKey)
				}
			}
		})
	}
}

func TestParseCurlCommandBody(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantBody string
		wantType string
	}{
		{
			name:     "simple data",
			input:    "curl -d 'name=test' https://example.com",
			wantBody: "name=test",
			wantType: "raw",
		},
		{
			name:     "JSON data",
			input:    `curl -d '{"key":"value"}' https://example.com`,
			wantBody: `{"key":"value"}`,
			wantType: "json",
		},
		{
			name:     "data-raw flag",
			input:    "curl --data-raw 'raw content' https://example.com",
			wantBody: "raw content",
			wantType: "raw",
		},
		{
			name:     "multiple data flags concatenated",
			input:    "curl -d 'a=1' -d 'b=2' https://example.com",
			wantBody: "a=1&b=2",
			wantType: "raw",
		},
		{
			name:     "JSON array",
			input:    `curl -d '[1,2,3]' https://example.com`,
			wantBody: `[1,2,3]`,
			wantType: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseCurlCommand(tt.input)
			if err != nil {
				t.Fatalf("ParseCurlCommand() error = %v", err)
			}

			if req.Body == nil {
				t.Fatal("expected body to be set")
			}

			content, ok := req.Body.Content.(string)
			if !ok {
				t.Fatalf("expected body content to be string, got %T", req.Body.Content)
			}

			if content != tt.wantBody {
				t.Errorf("body content = %q, want %q", content, tt.wantBody)
			}

			if req.Body.Type != tt.wantType {
				t.Errorf("body type = %q, want %q", req.Body.Type, tt.wantType)
			}
		})
	}
}

func TestParseCurlCommandAuth(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantUsername string
		wantPassword string
	}{
		{
			name:         "basic auth with password",
			input:        "curl -u user:pass https://example.com",
			wantUsername: "user",
			wantPassword: "pass",
		},
		{
			name:         "basic auth long form",
			input:        "curl --user admin:secret https://example.com",
			wantUsername: "admin",
			wantPassword: "secret",
		},
		{
			name:         "basic auth without password",
			input:        "curl -u username https://example.com",
			wantUsername: "username",
			wantPassword: "",
		},
		{
			name:         "basic auth with colon in password",
			input:        "curl -u user:pass:word https://example.com",
			wantUsername: "user",
			wantPassword: "pass:word",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseCurlCommand(tt.input)
			if err != nil {
				t.Fatalf("ParseCurlCommand() error = %v", err)
			}

			if req.Auth == nil {
				t.Fatal("expected auth to be set")
			}

			if req.Auth.Type != "basic" {
				t.Errorf("auth type = %q, want %q", req.Auth.Type, "basic")
			}

			if req.Auth.Username != tt.wantUsername {
				t.Errorf("username = %q, want %q", req.Auth.Username, tt.wantUsername)
			}

			if req.Auth.Password != tt.wantPassword {
				t.Errorf("password = %q, want %q", req.Auth.Password, tt.wantPassword)
			}
		})
	}
}

func TestValidateCurlCommand(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid command",
			input:   "curl https://example.com",
			wantErr: false,
		},
		{
			name:    "valid with quotes",
			input:   "curl -H 'Accept: */*' https://example.com",
			wantErr: false,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no curl",
			input:   "wget https://example.com",
			wantErr: true,
		},
		{
			name:    "unclosed single quote",
			input:   "curl -H 'Accept https://example.com",
			wantErr: true,
		},
		{
			name:    "unclosed double quote",
			input:   `curl -H "Accept https://example.com`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCurlCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCurlCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetectAndConvertVariables(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple variable",
			input: "$TOKEN",
			want:  "{{TOKEN}}",
		},
		{
			name:  "braced variable",
			input: "${BASE_URL}",
			want:  "{{BASE_URL}}",
		},
		{
			name:  "multiple variables",
			input: "$BASE_URL/users/$USER_ID",
			want:  "{{BASE_URL}}/users/{{USER_ID}}",
		},
		{
			name:  "variable in URL",
			input: "https://api.example.com/$VERSION/users",
			want:  "https://api.example.com/{{VERSION}}/users",
		},
		{
			name:  "no variables",
			input: "https://api.example.com/users",
			want:  "https://api.example.com/users",
		},
		{
			name:  "underscore in variable",
			input: "$API_KEY_SECRET",
			want:  "{{API_KEY_SECRET}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectAndConvertVariables(tt.input)
			if got != tt.want {
				t.Errorf("detectAndConvertVariables() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractNameFromURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "with path",
			input: "https://api.example.com/users",
			want:  "users",
		},
		{
			name:  "with nested path",
			input: "https://api.example.com/v1/users/profile",
			want:  "profile",
		},
		{
			name:  "root path",
			input: "https://api.example.com/",
			want:  "api.example.com",
		},
		{
			name:  "no path",
			input: "https://api.example.com",
			want:  "api.example.com",
		},
		{
			name:  "with query string",
			input: "https://api.example.com/search?q=test",
			want:  "search",
		},
		{
			name:  "with port",
			input: "http://localhost:8080/api/users",
			want:  "users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractNameFromURL(tt.input)
			if got != tt.want {
				t.Errorf("extractNameFromURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseError_FormatWithContext(t *testing.T) {
	err := &ParseError{
		Message:  "unclosed quote",
		Position: 20,
		Line:     1,
		Column:   20,
		Context:  "-H 'Content-Type: app",
	}

	formatted := err.FormatWithContext()
	if formatted == "" {
		t.Error("FormatWithContext() returned empty string")
	}
	if !strings.Contains(formatted, "unclosed quote") {
		t.Error("FormatWithContext() should contain error message")
	}
	if !strings.Contains(formatted, "^") {
		t.Error("FormatWithContext() should contain position indicator")
	}
}

// TestEdgeCases tests various edge cases for cURL parsing
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantURL    string
		wantMethod HTTPMethod
		wantErr    bool
	}{
		{
			name:       "ignored flags - silent",
			input:      "curl -s https://example.com",
			wantURL:    "https://example.com",
			wantMethod: GET,
			wantErr:    false,
		},
		{
			name:       "ignored flags - verbose and location",
			input:      "curl -v -L https://example.com",
			wantURL:    "https://example.com",
			wantMethod: GET,
			wantErr:    false,
		},
		{
			name:       "ignored flags - compressed",
			input:      "curl --compressed https://api.example.com",
			wantURL:    "https://api.example.com",
			wantMethod: GET,
			wantErr:    false,
		},
		{
			name:       "special chars in URL query",
			input:      "curl 'https://api.example.com/search?q=hello%20world&limit=10'",
			wantURL:    "https://api.example.com/search?q=hello%20world&limit=10",
			wantMethod: GET,
			wantErr:    false,
		},
		{
			name:       "unicode in header value",
			input:      "curl -H 'X-Custom: café' https://example.com",
			wantURL:    "https://example.com",
			wantMethod: GET,
			wantErr:    false,
		},
		{
			name:       "form data warning",
			input:      "curl -F 'file=@test.txt' https://example.com/upload",
			wantURL:    "https://example.com/upload",
			wantMethod: GET, // -F doesn't auto-set POST (use -X POST -F for multipart)
			wantErr:    false,
		},
		{
			name:       "insecure flag",
			input:      "curl -k https://self-signed.example.com",
			wantURL:    "https://self-signed.example.com",
			wantMethod: GET,
			wantErr:    false,
		},
		{
			name:       "user agent flag",
			input:      "curl -A 'MyApp/1.0' https://example.com",
			wantURL:    "https://example.com",
			wantMethod: GET,
			wantErr:    false,
		},
		{
			name:       "cookie flag",
			input:      "curl -b 'session=abc123' https://example.com",
			wantURL:    "https://example.com",
			wantMethod: GET,
			wantErr:    false,
		},
		{
			name:       "multiple flags combined",
			input:      "curl -s -L -k -v --compressed -X POST https://api.example.com",
			wantURL:    "https://api.example.com",
			wantMethod: POST,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := ParseCurlCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCurlCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if req.URL != tt.wantURL {
					t.Errorf("URL = %q, want %q", req.URL, tt.wantURL)
				}
				if req.Method != tt.wantMethod {
					t.Errorf("Method = %v, want %v", req.Method, tt.wantMethod)
				}
			}
		})
	}
}

// TestUserAgentHeader verifies -A flag converts to User-Agent header
func TestUserAgentHeader(t *testing.T) {
	req, err := ParseCurlCommand("curl -A 'Mozilla/5.0' https://example.com")
	if err != nil {
		t.Fatalf("ParseCurlCommand() error = %v", err)
	}

	found := false
	for _, h := range req.Headers {
		if h.Key == "User-Agent" && h.Value == "Mozilla/5.0" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected User-Agent header from -A flag")
	}
}

// TestCookieHeader verifies -b flag converts to Cookie header
func TestCookieHeader(t *testing.T) {
	req, err := ParseCurlCommand("curl -b 'session=abc; token=xyz' https://example.com")
	if err != nil {
		t.Fatalf("ParseCurlCommand() error = %v", err)
	}

	found := false
	for _, h := range req.Headers {
		if h.Key == "Cookie" && h.Value == "session=abc; token=xyz" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected Cookie header from -b flag")
	}
}
