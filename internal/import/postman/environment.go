package postman

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kbrdn1/LazyCurl/internal/api"
)

// ImportEnvironment imports a Postman Environment file and converts it to LazyCurl format.
func ImportEnvironment(filePath string) (*ImportResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return ImportEnvironmentFromBytes(data)
}

// ImportEnvironmentFromBytes imports a Postman Environment from raw JSON bytes.
func ImportEnvironmentFromBytes(data []byte) (*ImportResult, error) {
	pe, err := parseEnvironment(data)
	if err != nil {
		return nil, err
	}

	if err := validateEnvironment(pe); err != nil {
		return nil, err
	}

	env, summary := convertEnvironment(pe)
	return &ImportResult{
		Environment: env,
		Summary:     *summary,
	}, nil
}

// parseEnvironment parses JSON bytes into a Environment struct.
func parseEnvironment(data []byte) (*Environment, error) {
	var pe Environment
	if err := json.Unmarshal(data, &pe); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return &pe, nil
}

// validateEnvironment validates that the parsed data is a valid Postman Environment.
func validateEnvironment(pe *Environment) error {
	if pe.Name == "" {
		return fmt.Errorf("invalid environment: name is required")
	}
	return nil
}

// convertEnvironment converts a Environment to a LazyCurl EnvironmentFile.
func convertEnvironment(pe *Environment) (*api.EnvironmentFile, *ImportSummary) {
	summary := &ImportSummary{
		EnvironmentName: pe.Name,
	}

	env := &api.EnvironmentFile{
		Name:      pe.Name,
		Variables: make(map[string]*api.EnvironmentVariable),
	}

	for _, v := range pe.Values {
		summary.VariablesCount++

		// Map Postman type=secret to Secret=true
		isSecret := v.Type == "secret"

		env.Variables[v.Key] = &api.EnvironmentVariable{
			Value:  v.Value,
			Secret: isSecret,
			Active: v.Enabled,
		}
	}

	return env, summary
}
