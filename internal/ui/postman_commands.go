package ui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kbrdn1/LazyCurl/internal/api"
	"github.com/kbrdn1/LazyCurl/internal/import/postman"
)

// ImportPostmanFile imports a Postman collection or environment file.
// It auto-detects the file type and imports accordingly.
func ImportPostmanFile(filePath string) tea.Cmd {
	return func() tea.Msg {
		// Detect file type
		fileType, err := postman.DetectFileType(filePath)
		if err != nil {
			return PostmanImportErrorMsg{Error: fmt.Errorf("failed to detect file type: %w", err)}
		}

		switch fileType {
		case postman.FileTypeCollection:
			result, err := postman.ImportCollection(filePath)
			if err != nil {
				return PostmanImportErrorMsg{Error: fmt.Errorf("failed to import collection: %w", err)}
			}
			return PostmanImportedMsg{
				Collection: result.Collection,
				Summary:    result.FormatSummary(),
				IsEnv:      false,
			}

		case postman.FileTypeEnvironment:
			result, err := postman.ImportEnvironment(filePath)
			if err != nil {
				return PostmanImportErrorMsg{Error: fmt.Errorf("failed to import environment: %w", err)}
			}
			return PostmanImportedMsg{
				Environment: result.Environment,
				Summary:     result.FormatSummary(),
				IsEnv:       true,
			}

		default:
			return PostmanImportErrorMsg{Error: fmt.Errorf("unrecognized file format: not a valid Postman collection or environment")}
		}
	}
}

// ExportCollectionToPostman exports a LazyCurl collection to Postman format.
func ExportCollectionToPostman(collection *api.CollectionFile, outputPath string) tea.Cmd {
	return func() tea.Msg {
		if collection == nil {
			return PostmanExportedMsg{
				Success: false,
				Error:   fmt.Errorf("no collection to export"),
			}
		}

		// Ensure directory exists
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return PostmanExportedMsg{
				Success: false,
				Error:   fmt.Errorf("failed to create directory: %w", err),
			}
		}

		// Export to file
		if err := postman.ExportCollection(collection, outputPath); err != nil {
			return PostmanExportedMsg{
				Success: false,
				Error:   fmt.Errorf("failed to export collection: %w", err),
			}
		}

		return PostmanExportedMsg{
			Success:  true,
			FilePath: outputPath,
		}
	}
}

// ExportEnvironmentToPostman exports a LazyCurl environment to Postman format.
func ExportEnvironmentToPostman(env *api.EnvironmentFile, outputPath string) tea.Cmd {
	return func() tea.Msg {
		if env == nil {
			return PostmanExportedMsg{
				Success: false,
				Error:   fmt.Errorf("no environment to export"),
			}
		}

		// Ensure directory exists
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return PostmanExportedMsg{
				Success: false,
				Error:   fmt.Errorf("failed to create directory: %w", err),
			}
		}

		// Export to file
		if err := postman.ExportEnvironment(env, outputPath); err != nil {
			return PostmanExportedMsg{
				Success: false,
				Error:   fmt.Errorf("failed to export environment: %w", err),
			}
		}

		return PostmanExportedMsg{
			Success:  true,
			FilePath: outputPath,
		}
	}
}

// SaveImportedCollection saves an imported collection to the workspace.
func SaveImportedCollection(collection *api.CollectionFile, workspacePath string) error {
	if collection == nil {
		return fmt.Errorf("no collection to save")
	}

	// Create collections directory if needed
	collectionsDir := filepath.Join(workspacePath, ".lazycurl", "collections")
	if err := os.MkdirAll(collectionsDir, 0755); err != nil {
		return fmt.Errorf("failed to create collections directory: %w", err)
	}

	// Generate filename from collection name
	filename := sanitizeFilename(collection.Name) + ".json"
	outputPath := filepath.Join(collectionsDir, filename)

	// Check if file exists and add suffix if needed
	outputPath = ensureUniqueFilename(outputPath)

	// Save collection
	return api.SaveCollection(collection, outputPath)
}

// SaveImportedEnvironment saves an imported environment to the workspace.
func SaveImportedEnvironment(env *api.EnvironmentFile, workspacePath string) error {
	if env == nil {
		return fmt.Errorf("no environment to save")
	}

	// Create environments directory if needed
	envsDir := filepath.Join(workspacePath, ".lazycurl", "environments")
	if err := os.MkdirAll(envsDir, 0755); err != nil {
		return fmt.Errorf("failed to create environments directory: %w", err)
	}

	// Generate filename from environment name
	filename := sanitizeFilename(env.Name) + ".json"
	outputPath := filepath.Join(envsDir, filename)

	// Check if file exists and add suffix if needed
	outputPath = ensureUniqueFilename(outputPath)

	// Save environment
	return api.SaveEnvironment(env, outputPath)
}

// sanitizeFilename removes or replaces characters that are invalid in filenames.
func sanitizeFilename(name string) string {
	// Replace common invalid characters
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch c {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			result = append(result, '_')
		case ' ':
			result = append(result, '_')
		default:
			result = append(result, c)
		}
	}

	// Ensure non-empty filename
	if len(result) == 0 {
		return "untitled"
	}

	return string(result)
}

// ensureUniqueFilename adds a numeric suffix if the file already exists.
func ensureUniqueFilename(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := filepath.Ext(path)
	base := path[:len(path)-len(ext)]

	for i := 1; i < 100; i++ {
		newPath := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}

	// Fallback with timestamp
	return fmt.Sprintf("%s_%d%s", base, os.Getpid(), ext)
}
