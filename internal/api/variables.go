package api

import (
	"fmt"
	"math/rand" //nolint:gosec // G404: Used for non-security random data generation (test data, not crypto)
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

func init() {
	rand.Seed(time.Now().UnixNano()) //nolint:staticcheck // SA1019: Seeding for Go < 1.20 compatibility
}

// Variable patterns
var (
	variablePattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)
)

// System variable prefixes
const (
	SystemVarTimestamp  = "$timestamp"
	SystemVarDatetime   = "$datetime"
	SystemVarDate       = "$date"
	SystemVarTime       = "$time"
	SystemVarRandomInt  = "$randomInt"
	SystemVarRandomUUID = "$uuid"
	SystemVarGUID       = "$guid"
	SystemVarRandom     = "$random"
)

// ReplaceVariables replaces all variables in a string with their values from the environment
func ReplaceVariables(text string, env *EnvironmentFile) string {
	return variablePattern.ReplaceAllStringFunc(text, func(match string) string {
		// Extract variable name (remove {{ and }})
		varName := strings.TrimSpace(match[2 : len(match)-2])

		// Check for system variables first
		if strings.HasPrefix(varName, "$") {
			if value := getSystemVariable(varName); value != "" {
				return value
			}
		}

		// Check environment variables (only if active)
		if env != nil {
			if v, exists := env.Variables[varName]; exists && v.Active {
				return v.Value
			}
		}

		// Return original if not found (keep the placeholder)
		return match
	})
}

// ReplaceVariablesInRequest replaces variables in all parts of a request
func ReplaceVariablesInRequest(req *Request, env *EnvironmentFile) *Request {
	replaced := &Request{
		Method:  req.Method,
		URL:     ReplaceVariables(req.URL, env),
		Headers: make(map[string]string),
		Body:    req.Body,
		Timeout: req.Timeout,
	}

	// Replace in headers
	for key, value := range req.Headers {
		replacedKey := ReplaceVariables(key, env)
		replacedValue := ReplaceVariables(value, env)
		replaced.Headers[replacedKey] = replacedValue
	}

	// Replace in body if it's a string or map
	replaced.Body = replaceVariablesInBody(req.Body, env)

	return replaced
}

// replaceVariablesInBody recursively replaces variables in body
func replaceVariablesInBody(body interface{}, env *EnvironmentFile) interface{} {
	if body == nil {
		return nil
	}

	switch v := body.(type) {
	case string:
		return ReplaceVariables(v, env)

	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			replacedKey := ReplaceVariables(key, env)
			result[replacedKey] = replaceVariablesInBody(value, env)
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = replaceVariablesInBody(item, env)
		}
		return result

	default:
		return body
	}
}

// getSystemVariable returns the value of a system variable
func getSystemVariable(name string) string {
	now := time.Now()

	switch name {
	case SystemVarTimestamp:
		return fmt.Sprintf("%d", now.Unix())

	case SystemVarDatetime:
		return now.Format(time.RFC3339)

	case SystemVarDate:
		return now.Format("2006-01-02")

	case SystemVarTime:
		return now.Format("15:04:05")

	case SystemVarRandomInt:
		return fmt.Sprintf("%d", rand.Intn(1000000))

	case SystemVarRandomUUID, SystemVarGUID:
		return uuid.New().String()

	case SystemVarRandom:
		// Random string of 10 characters
		return generateRandomString(10)

	default:
		return ""
	}
}

// generateRandomString generates a random alphanumeric string
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// FindVariables finds all variable references in a string
func FindVariables(text string) []string {
	matches := variablePattern.FindAllStringSubmatch(text, -1)
	variables := make([]string, 0, len(matches))

	for _, match := range matches {
		if len(match) > 1 {
			varName := strings.TrimSpace(match[1])
			variables = append(variables, varName)
		}
	}

	return variables
}

// FindUnresolvedVariables finds variables that couldn't be resolved in the environment
func FindUnresolvedVariables(text string, env *EnvironmentFile) []string {
	allVars := FindVariables(text)
	unresolved := []string{}

	for _, varName := range allVars {
		// System variables are always resolved
		if strings.HasPrefix(varName, "$") {
			continue
		}

		// Check if variable exists in environment
		if env == nil || !env.HasVariable(varName) {
			unresolved = append(unresolved, varName)
		}
	}

	return unresolved
}

// ValidateVariables checks if all variables in a request can be resolved
func ValidateVariables(req *Request, env *EnvironmentFile) []string {
	unresolved := []string{}

	// Check URL
	unresolved = append(unresolved, FindUnresolvedVariables(req.URL, env)...)

	// Check headers
	for key, value := range req.Headers {
		unresolved = append(unresolved, FindUnresolvedVariables(key, env)...)
		unresolved = append(unresolved, FindUnresolvedVariables(value, env)...)
	}

	// Check body if it's a string
	if bodyStr, ok := req.Body.(string); ok {
		unresolved = append(unresolved, FindUnresolvedVariables(bodyStr, env)...)
	}

	// Remove duplicates
	return uniqueStrings(unresolved)
}

// uniqueStrings removes duplicate strings from a slice
func uniqueStrings(strings []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, str := range strings {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}

// PreviewVariableReplacement shows what the text would look like after variable replacement
func PreviewVariableReplacement(text string, env *EnvironmentFile) string {
	return ReplaceVariables(text, env)
}
