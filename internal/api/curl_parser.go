package api

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// ParseError provides context about parsing failures
type ParseError struct {
	Message  string
	Position int
	Line     int
	Column   int
	Context  string
}

// Error implements the error interface
func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d, col %d: %s", e.Line, e.Column, e.Message)
}

// FormatWithContext returns a multi-line error message with visual indicator
func (e *ParseError) FormatWithContext() string {
	if e.Context == "" {
		return e.Error()
	}
	indicator := strings.Repeat(" ", e.Column-1) + "^"
	return fmt.Sprintf("%s\n  %s\n  %s", e.Error(), e.Context, indicator)
}

// TokenType represents the type of a lexer token
type TokenType int

const (
	TokenWord   TokenType = iota // Unquoted word
	TokenString                  // Quoted string (quotes stripped)
	TokenFlag                    // Flag starting with - or --
	TokenEquals                  // = sign (for --flag=value)
	TokenEOF                     // End of input
)

// Token represents a lexer token
type Token struct {
	Type     TokenType
	Value    string
	Position int
	Line     int
	Column   int
}

// ParsedCurlCommand holds the result of parsing a cURL command
type ParsedCurlCommand struct {
	Method    HTTPMethod
	URL       string
	Headers   []KeyValueEntry
	Body      string
	BasicAuth *BasicAuthCreds
	UserAgent string
	Cookies   []string
	Insecure  bool
	RawFlags  []string // Unrecognized flags
}

// BasicAuthCreds holds parsed basic auth credentials
type BasicAuthCreds struct {
	Username string
	Password string
}

// normalizeMultiline joins lines with backslash or backtick continuation
// while preserving whitespace inside quoted strings
func normalizeMultiline(cmd string) string {
	// Handle Unix backslash continuation
	cmd = strings.ReplaceAll(cmd, "\\\n", " ")
	cmd = strings.ReplaceAll(cmd, "\\\r\n", " ")
	// Handle PowerShell backtick continuation
	cmd = strings.ReplaceAll(cmd, "`\n", " ")
	cmd = strings.ReplaceAll(cmd, "`\r\n", " ")
	// Normalize whitespace outside of quotes
	return normalizeWhitespacePreservingQuotes(cmd)
}

// normalizeWhitespacePreservingQuotes collapses whitespace runs to single space
// but preserves whitespace inside single or double quoted strings
func normalizeWhitespacePreservingQuotes(s string) string {
	var result strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	lastWasSpace := false
	escaped := false

	for _, ch := range s {
		if escaped {
			result.WriteRune(ch)
			escaped = false
			lastWasSpace = false
			continue
		}

		if ch == '\\' && inDoubleQuote {
			result.WriteRune(ch)
			escaped = true
			continue
		}

		if ch == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
			result.WriteRune(ch)
			lastWasSpace = false
			continue
		}

		if ch == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
			result.WriteRune(ch)
			lastWasSpace = false
			continue
		}

		// Inside quotes - preserve all characters
		if inSingleQuote || inDoubleQuote {
			result.WriteRune(ch)
			lastWasSpace = false
			continue
		}

		// Outside quotes - collapse whitespace
		if unicode.IsSpace(ch) {
			if !lastWasSpace {
				result.WriteRune(' ')
				lastWasSpace = true
			}
		} else {
			result.WriteRune(ch)
			lastWasSpace = false
		}
	}

	return strings.TrimSpace(result.String())
}

// tokenize splits the cURL command into tokens
func tokenize(cmd string) ([]Token, error) {
	var tokens []Token
	runes := []rune(cmd)
	i := 0
	line := 1
	col := 1

	for i < len(runes) {
		ch := runes[i]

		// Skip whitespace
		if unicode.IsSpace(ch) {
			if ch == '\n' {
				line++
				col = 1
			} else {
				col++
			}
			i++
			continue
		}

		startPos := i
		startCol := col

		// Handle flags
		if ch == '-' {
			flagStart := i
			i++
			col++

			// Check for long flag (--flag)
			if i < len(runes) && runes[i] == '-' {
				i++
				col++
			}

			// Read flag name
			for i < len(runes) && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '-' || runes[i] == '_') {
				i++
				col++
			}

			flagValue := string(runes[flagStart:i])
			tokens = append(tokens, Token{
				Type:     TokenFlag,
				Value:    flagValue,
				Position: startPos,
				Line:     line,
				Column:   startCol,
			})

			// Check for --flag=value syntax
			if i < len(runes) && runes[i] == '=' {
				tokens = append(tokens, Token{
					Type:     TokenEquals,
					Value:    "=",
					Position: i,
					Line:     line,
					Column:   col,
				})
				i++
				col++
			}
			continue
		}

		// Handle single-quoted string
		if ch == '\'' {
			i++
			col++
			start := i
			for i < len(runes) && runes[i] != '\'' {
				if runes[i] == '\n' {
					line++
					col = 1
				} else {
					col++
				}
				i++
			}
			if i >= len(runes) {
				return nil, &ParseError{
					Message:  "unclosed single quote",
					Position: startPos,
					Line:     line,
					Column:   startCol,
					Context:  extractContext(cmd, startPos),
				}
			}
			tokens = append(tokens, Token{
				Type:     TokenString,
				Value:    string(runes[start:i]),
				Position: startPos,
				Line:     line,
				Column:   startCol,
			})
			i++
			col++
			continue
		}

		// Handle double-quoted string
		if ch == '"' {
			i++
			col++
			var value strings.Builder
			for i < len(runes) && runes[i] != '"' {
				if runes[i] == '\\' && i+1 < len(runes) {
					// Handle escape sequences
					nextChar := runes[i+1]
					switch nextChar {
					case '"', '\\', '/':
						value.WriteRune(nextChar)
						i += 2
						col += 2
						continue
					case 'n':
						value.WriteRune('\n')
						i += 2
						col += 2
						continue
					case 't':
						value.WriteRune('\t')
						i += 2
						col += 2
						continue
					case 'r':
						value.WriteRune('\r')
						i += 2
						col += 2
						continue
					}
				}
				if runes[i] == '\n' {
					line++
					col = 1
				} else {
					col++
				}
				value.WriteRune(runes[i])
				i++
			}
			if i >= len(runes) {
				return nil, &ParseError{
					Message:  "unclosed double quote",
					Position: startPos,
					Line:     line,
					Column:   startCol,
					Context:  extractContext(cmd, startPos),
				}
			}
			tokens = append(tokens, Token{
				Type:     TokenString,
				Value:    value.String(),
				Position: startPos,
				Line:     line,
				Column:   startCol,
			})
			i++
			col++
			continue
		}

		// Handle $'...' ANSI-C quoting
		if ch == '$' && i+1 < len(runes) && runes[i+1] == '\'' {
			i += 2
			col += 2
			var value strings.Builder
			for i < len(runes) && runes[i] != '\'' {
				if runes[i] == '\\' && i+1 < len(runes) {
					nextChar := runes[i+1]
					switch nextChar {
					case '\'', '\\':
						value.WriteRune(nextChar)
						i += 2
						col += 2
						continue
					case 'n':
						value.WriteRune('\n')
						i += 2
						col += 2
						continue
					case 't':
						value.WriteRune('\t')
						i += 2
						col += 2
						continue
					case 'r':
						value.WriteRune('\r')
						i += 2
						col += 2
						continue
					}
				}
				if runes[i] == '\n' {
					line++
					col = 1
				} else {
					col++
				}
				value.WriteRune(runes[i])
				i++
			}
			if i >= len(runes) {
				return nil, &ParseError{
					Message:  "unclosed ANSI-C quote",
					Position: startPos,
					Line:     line,
					Column:   startCol,
					Context:  extractContext(cmd, startPos),
				}
			}
			tokens = append(tokens, Token{
				Type:     TokenString,
				Value:    value.String(),
				Position: startPos,
				Line:     line,
				Column:   startCol,
			})
			i++
			col++
			continue
		}

		// Handle unquoted word
		start := i
		for i < len(runes) && !unicode.IsSpace(runes[i]) && runes[i] != '\'' && runes[i] != '"' && runes[i] != '=' {
			// Handle escaped character in unquoted word
			if runes[i] == '\\' && i+1 < len(runes) {
				i += 2
				col += 2
				continue
			}
			i++
			col++
		}
		if i > start {
			tokens = append(tokens, Token{
				Type:     TokenWord,
				Value:    string(runes[start:i]),
				Position: startPos,
				Line:     line,
				Column:   startCol,
			})
		}
	}

	tokens = append(tokens, Token{
		Type:     TokenEOF,
		Value:    "",
		Position: len(runes),
		Line:     line,
		Column:   col,
	})

	return tokens, nil
}

// extractContext extracts a snippet of the input around the given position
func extractContext(input string, pos int) string {
	start := pos - 20
	if start < 0 {
		start = 0
	}
	end := pos + 20
	if end > len(input) {
		end = len(input)
	}
	return strings.TrimSpace(input[start:end])
}

// parseTokens extracts flags and values from tokens into ParsedCurlCommand
func parseTokens(tokens []Token) (*ParsedCurlCommand, error) {
	parsed := &ParsedCurlCommand{
		Method:   GET, // Default method
		Headers:  []KeyValueEntry{},
		RawFlags: []string{},
	}

	var urls []string
	var bodies []string

	i := 0
	for i < len(tokens) {
		token := tokens[i]

		if token.Type == TokenEOF {
			break
		}

		// Skip "curl" command
		if (token.Type == TokenWord || token.Type == TokenString) && strings.ToLower(token.Value) == "curl" {
			i++
			continue
		}

		// Handle flags
		if token.Type == TokenFlag {
			flag := token.Value
			i++

			// Get flag value (could be after = or next token)
			var flagValue string
			hasValue := false

			if i < len(tokens) && tokens[i].Type == TokenEquals {
				i++
				if i < len(tokens) && (tokens[i].Type == TokenString || tokens[i].Type == TokenWord) {
					flagValue = tokens[i].Value
					hasValue = true
					i++
				}
			} else if i < len(tokens) && (tokens[i].Type == TokenString || tokens[i].Type == TokenWord) {
				// Flags that require values
				needsValue := map[string]bool{
					"-X": true, "--request": true,
					"-H": true, "--header": true,
					"-d": true, "--data": true, "--data-raw": true, "--data-binary": true,
					"-u": true, "--user": true,
					"-A": true, "--user-agent": true,
					"--cookie": true, "-b": true,
					"-F": true, "--form": true,
					"-o": true, "--output": true,
				}
				if needsValue[flag] {
					flagValue = tokens[i].Value
					hasValue = true
					i++
				}
			}

			// Process the flag
			switch flag {
			case "-X", "--request":
				if hasValue {
					parsed.Method = HTTPMethod(strings.ToUpper(flagValue))
				}
			case "-H", "--header":
				if hasValue {
					header := parseHeader(flagValue)
					if header != nil {
						parsed.Headers = append(parsed.Headers, *header)
					}
				}
			case "-d", "--data", "--data-raw":
				if hasValue {
					bodies = append(bodies, flagValue)
				}
			case "--data-binary":
				if hasValue {
					// Handle @filename syntax - store as-is for now
					bodies = append(bodies, flagValue)
				}
			case "-u", "--user":
				if hasValue {
					parsed.BasicAuth = parseBasicAuth(flagValue)
				}
			case "-A", "--user-agent":
				if hasValue {
					parsed.UserAgent = flagValue
				}
			case "--cookie", "-b":
				if hasValue {
					parsed.Cookies = append(parsed.Cookies, flagValue)
				}
			case "-k", "--insecure":
				parsed.Insecure = true
			case "-F", "--form":
				// Form data - store as warning for now
				if hasValue {
					parsed.RawFlags = append(parsed.RawFlags, fmt.Sprintf("%s=%s (multipart form not fully supported)", flag, flagValue))
				}
			case "-s", "--silent", "-S", "--show-error", "-L", "--location", "--compressed", "-v", "--verbose":
				// Silently ignored flags (don't affect request content)
			case "-o", "--output":
				// Ignored - we display in UI
			default:
				// Unrecognized flag
				if hasValue {
					parsed.RawFlags = append(parsed.RawFlags, fmt.Sprintf("%s=%s", flag, flagValue))
				} else {
					parsed.RawFlags = append(parsed.RawFlags, flag)
				}
			}
			continue
		}

		// Non-flag tokens are URLs or words
		if token.Type == TokenString || token.Type == TokenWord {
			// Check if it looks like a URL
			value := token.Value
			if looksLikeURL(value) {
				urls = append(urls, value)
			}
			i++
			continue
		}

		i++
	}

	// Set URL
	if len(urls) == 0 {
		return nil, &ParseError{
			Message: "URL is required",
		}
	}
	parsed.URL = urls[0]

	// Set body (concatenate multiple -d flags with &)
	if len(bodies) > 0 {
		parsed.Body = strings.Join(bodies, "&")
		// If body is present and method is still GET, default to POST
		if parsed.Method == GET {
			parsed.Method = POST
		}
	}

	// Convert User-Agent to header if set
	if parsed.UserAgent != "" {
		parsed.Headers = append(parsed.Headers, KeyValueEntry{
			Key:     "User-Agent",
			Value:   parsed.UserAgent,
			Enabled: true,
		})
	}

	// Convert cookies to header if set
	if len(parsed.Cookies) > 0 {
		parsed.Headers = append(parsed.Headers, KeyValueEntry{
			Key:     "Cookie",
			Value:   strings.Join(parsed.Cookies, "; "),
			Enabled: true,
		})
	}

	return parsed, nil
}

// parseHeader parses a header string in "Key: Value" format
func parseHeader(header string) *KeyValueEntry {
	parts := strings.SplitN(header, ":", 2)
	if len(parts) != 2 {
		return nil
	}
	return &KeyValueEntry{
		Key:     strings.TrimSpace(parts[0]),
		Value:   strings.TrimSpace(parts[1]),
		Enabled: true,
	}
}

// parseBasicAuth parses "username:password" format
func parseBasicAuth(auth string) *BasicAuthCreds {
	parts := strings.SplitN(auth, ":", 2)
	if len(parts) == 1 {
		return &BasicAuthCreds{
			Username: parts[0],
			Password: "",
		}
	}
	return &BasicAuthCreds{
		Username: parts[0],
		Password: parts[1],
	}
}

// looksLikeURL checks if a string appears to be a URL
func looksLikeURL(s string) bool {
	// Check for common URL schemes
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return true
	}
	// Check for URL-like patterns without scheme (e.g., localhost:8080, example.com/path)
	if strings.Contains(s, "://") {
		return true
	}
	// Check for common patterns
	if strings.HasPrefix(s, "localhost") || strings.HasPrefix(s, "127.0.0.1") {
		return true
	}
	// Check for domain-like pattern (contains dot or colon for port)
	if strings.Contains(s, ".") && !strings.HasPrefix(s, "-") && !strings.HasPrefix(s, ".") {
		return true
	}
	return false
}

// ToCollectionRequest converts ParsedCurlCommand to CollectionRequest
func (p *ParsedCurlCommand) ToCollectionRequest() *CollectionRequest {
	// Convert shell variables ($VAR, ${VAR}) to LazyCurl variables ({{VAR}})
	url := detectAndConvertVariables(p.URL)

	// Convert variables in headers
	headers := make([]KeyValueEntry, len(p.Headers))
	for i, h := range p.Headers {
		headers[i] = KeyValueEntry{
			Key:     h.Key,
			Value:   detectAndConvertVariables(h.Value),
			Enabled: h.Enabled,
		}
	}

	req := &CollectionRequest{
		ID:      GenerateID(),
		Name:    extractNameFromURL(p.URL),
		Method:  p.Method,
		URL:     url,
		Headers: headers,
	}

	// Set body if present
	if p.Body != "" {
		bodyType := "raw"
		// Try to detect JSON
		trimmed := strings.TrimSpace(p.Body)
		if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
			(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
			bodyType = "json"
		}
		req.Body = &BodyConfig{
			Type:    bodyType,
			Content: detectAndConvertVariables(p.Body),
		}
	}

	// Set auth if present
	if p.BasicAuth != nil {
		req.Auth = &AuthConfig{
			Type:     "basic",
			Username: p.BasicAuth.Username,
			Password: p.BasicAuth.Password,
		}
	}

	return req
}

// extractNameFromURL generates a request name from the URL
func extractNameFromURL(url string) string {
	// Remove scheme
	name := url
	name = strings.TrimPrefix(name, "http://")
	name = strings.TrimPrefix(name, "https://")

	// Get path part
	if idx := strings.Index(name, "/"); idx != -1 {
		path := name[idx:]
		// Remove query string
		if qIdx := strings.Index(path, "?"); qIdx != -1 {
			path = path[:qIdx]
		}
		if path != "/" && path != "" {
			// Get last segment
			segments := strings.Split(strings.Trim(path, "/"), "/")
			if len(segments) > 0 {
				return segments[len(segments)-1]
			}
		}
	}

	// Fallback to hostname
	if idx := strings.Index(name, "/"); idx != -1 {
		name = name[:idx]
	}
	if idx := strings.Index(name, ":"); idx != -1 {
		name = name[:idx]
	}
	return name
}

// ParseCurlCommand parses a cURL command string into a CollectionRequest
func ParseCurlCommand(cmd string) (*CollectionRequest, error) {
	if strings.TrimSpace(cmd) == "" {
		return nil, &ParseError{Message: "empty cURL command"}
	}

	// Normalize multiline commands
	normalized := normalizeMultiline(cmd)

	// Check for curl command (case-insensitive)
	lower := strings.ToLower(normalized)
	if !strings.Contains(lower, "curl") {
		return nil, &ParseError{Message: "command must start with 'curl'"}
	}

	// Tokenize
	tokens, err := tokenize(normalized)
	if err != nil {
		return nil, err
	}

	// Parse tokens
	parsed, err := parseTokens(tokens)
	if err != nil {
		return nil, err
	}

	// Convert to CollectionRequest
	return parsed.ToCollectionRequest(), nil
}

// ValidateCurlCommand performs quick validation of a cURL command
func ValidateCurlCommand(cmd string) error {
	if strings.TrimSpace(cmd) == "" {
		return &ParseError{Message: "empty cURL command"}
	}

	normalized := normalizeMultiline(cmd)
	lower := strings.ToLower(normalized)

	if !strings.Contains(lower, "curl") {
		return &ParseError{Message: "command must start with 'curl'"}
	}

	// Check for balanced quotes
	singleQuotes := 0
	doubleQuotes := 0
	escaped := false

	for _, ch := range normalized {
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '\'' && doubleQuotes%2 == 0 {
			singleQuotes++
		}
		if ch == '"' && singleQuotes%2 == 0 {
			doubleQuotes++
		}
	}

	if singleQuotes%2 != 0 {
		return &ParseError{Message: "unclosed single quote"}
	}
	if doubleQuotes%2 != 0 {
		return &ParseError{Message: "unclosed double quote"}
	}

	return nil
}

// detectAndConvertVariables converts shell variables ($VAR, ${VAR}) to {{VAR}} syntax
func detectAndConvertVariables(value string) string {
	// Pattern for ${VAR} syntax
	bracePattern := regexp.MustCompile(`\$\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)
	value = bracePattern.ReplaceAllString(value, "{{$1}}")

	// Pattern for $VAR syntax (not followed by special chars that would make it not a variable)
	simplePattern := regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_]*)`)
	value = simplePattern.ReplaceAllString(value, "{{$1}}")

	return value
}
