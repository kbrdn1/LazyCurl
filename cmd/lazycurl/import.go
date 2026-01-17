package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kbrdn1/LazyCurl/internal/api"
	"github.com/kbrdn1/LazyCurl/internal/config"
	"github.com/kbrdn1/LazyCurl/internal/import/postman"
)

// ImportCommand handles the import subcommand
type ImportCommand struct {
	Format     string // "auto", "openapi", "postman"
	FilePath   string // Path to file to import
	Name       string // Override collection name
	Output     string // Custom output path
	DryRun     bool   // Preview only, don't save
	JSONOutput bool   // Output as JSON
}

// ParseImportArgs parses import command arguments
func ParseImportArgs(args []string) (*ImportCommand, error) {
	cmd := &ImportCommand{Format: "auto"} // Default to auto-detection

	if len(args) < 1 {
		return nil, fmt.Errorf("usage: lazycurl import <file> [options]\n       lazycurl import <format> <file> [options]\n\nFormats:\n  auto       Auto-detect format (default)\n  openapi    Import OpenAPI 3.x specification (JSON/YAML)\n  postman    Import Postman collection or environment\n\nOptions:\n  --format FORMAT  Specify import format (auto, openapi, postman)\n  --name NAME      Override collection name\n  --output PATH    Custom output path\n  --dry-run        Preview without saving\n  --json           Output results as JSON")
	}

	// Check if first arg is a format or a file
	firstArg := args[0]
	if firstArg == "openapi" || firstArg == "postman" || firstArg == "auto" {
		// Format specified as first argument
		if len(args) < 2 {
			return nil, fmt.Errorf("file path required after format")
		}
		cmd.Format = firstArg
		cmd.FilePath = args[1]
		args = args[2:] // Skip format and file for flag parsing
	} else {
		// File specified directly, format will be auto-detected
		cmd.FilePath = firstArg
		args = args[1:] // Skip file for flag parsing
	}

	// Parse optional flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--format requires a value")
			}
			i++
			cmd.Format = args[i]
		case "--name":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--name requires a value")
			}
			i++
			cmd.Name = args[i]
		case "--output":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--output requires a value")
			}
			i++
			cmd.Output = args[i]
		case "--dry-run":
			cmd.DryRun = true
		case "--json":
			cmd.JSONOutput = true
		default:
			if args[i][0] == '-' {
				return nil, fmt.Errorf("unknown option: %s", args[i])
			}
		}
	}

	return cmd, nil
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	Success        bool     `json:"success"`
	CollectionName string   `json:"collection_name,omitempty"`
	FilePath       string   `json:"file_path,omitempty"`
	FolderCount    int      `json:"folder_count,omitempty"`
	RequestCount   int      `json:"request_count,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
	Error          string   `json:"error,omitempty"`
	ErrorLine      int      `json:"error_line,omitempty"`
	ErrorColumn    int      `json:"error_column,omitempty"`
}

// RunImportCommand executes the import command
func RunImportCommand(cmd *ImportCommand) error {
	switch cmd.Format {
	case "auto":
		return runAutoDetectImport(cmd)
	case "openapi":
		return runOpenAPIImport(cmd)
	case "postman":
		return runPostmanImport(cmd)
	default:
		return fmt.Errorf("unsupported format: %s. Supported formats: auto, openapi, postman", cmd.Format)
	}
}

// runAutoDetectImport auto-detects file format and routes to appropriate importer
func runAutoDetectImport(cmd *ImportCommand) error {
	// Try Postman detection first (faster)
	fileType, err := postman.DetectFileType(cmd.FilePath)
	if err == nil && fileType != postman.FileTypeUnknown {
		// Detected Postman format
		return runPostmanImport(cmd)
	}

	// Fall back to OpenAPI
	return runOpenAPIImport(cmd)
}

// runPostmanImport handles Postman collection/environment import
func runPostmanImport(cmd *ImportCommand) error {
	// Detect file type
	fileType, err := postman.DetectFileType(cmd.FilePath)
	if err != nil {
		return handleImportError(cmd, fmt.Errorf("failed to detect file type: %w", err))
	}

	switch fileType {
	case postman.FileTypeCollection:
		return runPostmanCollectionImport(cmd)
	case postman.FileTypeEnvironment:
		return runPostmanEnvironmentImport(cmd)
	default:
		return handleImportError(cmd, fmt.Errorf("unrecognized file format: not a valid Postman collection or environment"))
	}
}

// runPostmanCollectionImport handles Postman collection import
func runPostmanCollectionImport(cmd *ImportCommand) error {
	// Import collection
	result, err := postman.ImportCollection(cmd.FilePath)
	if err != nil {
		return handleImportError(cmd, err)
	}

	// If dry-run, show preview and exit
	if cmd.DryRun {
		return outputPostmanCollectionPreview(cmd, result)
	}

	// Override name if provided
	if cmd.Name != "" {
		result.Collection.Name = cmd.Name
	}

	// Determine output path
	outputPath := cmd.Output
	if outputPath == "" {
		workspacePath, err := config.GetWorkspacePath()
		if err != nil {
			return handleImportError(cmd, fmt.Errorf("failed to get workspace path: %w", err))
		}
		collectionsDir := filepath.Join(workspacePath, ".lazycurl", "collections")
		if err := os.MkdirAll(collectionsDir, 0755); err != nil {
			return handleImportError(cmd, fmt.Errorf("failed to create collections directory: %w", err))
		}
		filename := sanitizeFilename(result.Collection.Name) + ".json"
		outputPath = filepath.Join(collectionsDir, filename)
	}

	// Save collection
	result.Collection.FilePath = outputPath
	if err := api.SaveCollection(result.Collection, outputPath); err != nil {
		return handleImportError(cmd, fmt.Errorf("failed to save collection: %w", err))
	}

	// Output result
	importResult := ImportResult{
		Success:        true,
		CollectionName: result.Collection.Name,
		FilePath:       outputPath,
		FolderCount:    result.Summary.FoldersCount,
		RequestCount:   result.Summary.RequestsCount,
		Warnings:       result.Summary.Warnings,
	}

	return outputResult(cmd, importResult)
}

// runPostmanEnvironmentImport handles Postman environment import
func runPostmanEnvironmentImport(cmd *ImportCommand) error {
	// Import environment
	result, err := postman.ImportEnvironment(cmd.FilePath)
	if err != nil {
		return handleImportError(cmd, err)
	}

	// If dry-run, show preview and exit
	if cmd.DryRun {
		return outputPostmanEnvironmentPreview(cmd, result)
	}

	// Override name if provided
	if cmd.Name != "" {
		result.Environment.Name = cmd.Name
	}

	// Determine output path
	outputPath := cmd.Output
	if outputPath == "" {
		workspacePath, err := config.GetWorkspacePath()
		if err != nil {
			return handleImportError(cmd, fmt.Errorf("failed to get workspace path: %w", err))
		}
		envsDir := filepath.Join(workspacePath, ".lazycurl", "environments")
		if err := os.MkdirAll(envsDir, 0755); err != nil {
			return handleImportError(cmd, fmt.Errorf("failed to create environments directory: %w", err))
		}
		filename := sanitizeFilename(result.Environment.Name) + ".json"
		outputPath = filepath.Join(envsDir, filename)
	}

	// Save environment
	if err := api.SaveEnvironment(result.Environment, outputPath); err != nil {
		return handleImportError(cmd, fmt.Errorf("failed to save environment: %w", err))
	}

	// Output result
	importResult := ImportResult{
		Success:        true,
		CollectionName: result.Environment.Name,
		FilePath:       outputPath,
		Warnings:       result.Summary.Warnings,
	}

	return outputResult(cmd, importResult)
}

// runOpenAPIImport handles OpenAPI import
func runOpenAPIImport(cmd *ImportCommand) error {
	// Load the OpenAPI file
	importer, err := api.NewOpenAPIImporterFromFile(cmd.FilePath)
	if err != nil {
		return handleImportError(cmd, err)
	}

	// Get preview
	preview, err := importer.Preview()
	if err != nil {
		return handleImportError(cmd, err)
	}

	// If dry-run, just show preview and exit
	if cmd.DryRun {
		return outputPreview(cmd, preview)
	}

	// Build import options
	opts := api.ImportOptions{
		Name:            cmd.Name,
		IncludeExamples: true,
	}

	// Convert to collection
	collection, err := importer.ToCollection(opts)
	if err != nil {
		return handleImportError(cmd, err)
	}

	// Determine output path
	outputPath := cmd.Output
	if outputPath == "" {
		workspacePath, err := config.GetWorkspacePath()
		if err != nil {
			return handleImportError(cmd, fmt.Errorf("failed to get workspace path: %w", err))
		}
		collectionsDir := filepath.Join(workspacePath, ".lazycurl", "collections")
		if err := os.MkdirAll(collectionsDir, 0755); err != nil {
			return handleImportError(cmd, fmt.Errorf("failed to create collections directory: %w", err))
		}
		filename := sanitizeFilename(collection.Name) + ".json"
		outputPath = filepath.Join(collectionsDir, filename)
	}

	// Save collection
	collection.FilePath = outputPath
	if err := api.SaveCollection(collection, outputPath); err != nil {
		return handleImportError(cmd, fmt.Errorf("failed to save collection: %w", err))
	}

	// Count requests
	requestCount := countCollectionRequests(collection)

	// Output result
	result := ImportResult{
		Success:        true,
		CollectionName: collection.Name,
		FilePath:       outputPath,
		FolderCount:    len(collection.Folders),
		RequestCount:   requestCount,
		Warnings:       preview.Warnings,
	}

	return outputResult(cmd, result)
}

// handleImportError handles and formats import errors
func handleImportError(cmd *ImportCommand, err error) error {
	result := ImportResult{
		Success: false,
		Error:   err.Error(),
	}

	// Extract line/column info from ImportError if available
	var importErr *api.ImportError
	if errors.As(err, &importErr) {
		result.Error = importErr.Message
		result.ErrorLine = importErr.Line
		result.ErrorColumn = importErr.Column
	}

	if cmd.JSONOutput {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		os.Exit(1)
	}

	if result.ErrorLine > 0 {
		fmt.Fprintf(os.Stderr, "Error at line %d", result.ErrorLine)
		if result.ErrorColumn > 0 {
			fmt.Fprintf(os.Stderr, ", column %d", result.ErrorColumn)
		}
		fmt.Fprintf(os.Stderr, ": %s\n", result.Error)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
	}
	os.Exit(1)
	return nil
}

// outputPostmanCollectionPreview outputs Postman collection preview
func outputPostmanCollectionPreview(cmd *ImportCommand, result *postman.ImportResult) error {
	if cmd.JSONOutput {
		preview := map[string]interface{}{
			"type":           "postman_collection",
			"name":           result.Collection.Name,
			"description":    result.Collection.Description,
			"folders_count":  result.Summary.FoldersCount,
			"requests_count": result.Summary.RequestsCount,
			"warnings":       result.Summary.Warnings,
		}
		data, err := json.MarshalIndent(preview, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Postman Collection Import Preview\n")
	fmt.Printf("==================================\n\n")
	fmt.Printf("Name:        %s\n", result.Collection.Name)
	if result.Collection.Description != "" {
		desc := result.Collection.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("Description: %s\n", desc)
	}
	fmt.Printf("\n")
	fmt.Printf("Folders:     %d\n", result.Summary.FoldersCount)
	fmt.Printf("Requests:    %d\n", result.Summary.RequestsCount)

	if len(result.Summary.Warnings) > 0 {
		fmt.Printf("\nWarnings:\n")
		for _, w := range result.Summary.Warnings {
			fmt.Printf("  ! %s\n", w)
		}
	}

	fmt.Printf("\n(dry-run mode - no files created)\n")
	return nil
}

// outputPostmanEnvironmentPreview outputs Postman environment preview
func outputPostmanEnvironmentPreview(cmd *ImportCommand, result *postman.ImportResult) error {
	if cmd.JSONOutput {
		preview := map[string]interface{}{
			"type":      "postman_environment",
			"name":      result.Environment.Name,
			"variables": len(result.Environment.Variables),
			"warnings":  result.Summary.Warnings,
		}
		data, err := json.MarshalIndent(preview, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Postman Environment Import Preview\n")
	fmt.Printf("===================================\n\n")
	fmt.Printf("Name:      %s\n", result.Environment.Name)
	fmt.Printf("Variables: %d\n", len(result.Environment.Variables))

	if len(result.Summary.Warnings) > 0 {
		fmt.Printf("\nWarnings:\n")
		for _, w := range result.Summary.Warnings {
			fmt.Printf("  ! %s\n", w)
		}
	}

	fmt.Printf("\n(dry-run mode - no files created)\n")
	return nil
}

// outputPreview outputs the import preview
func outputPreview(cmd *ImportCommand, preview *api.ImportPreview) error {
	if cmd.JSONOutput {
		data, err := json.MarshalIndent(preview, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("OpenAPI Import Preview\n")
	fmt.Printf("======================\n\n")
	fmt.Printf("Title:       %s\n", preview.Title)
	fmt.Printf("Version:     OpenAPI %s\n", preview.SpecVersion)
	if preview.Description != "" {
		desc := preview.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		fmt.Printf("Description: %s\n", desc)
	}
	fmt.Printf("\n")
	fmt.Printf("Endpoints:   %d\n", preview.EndpointCount)
	fmt.Printf("Folders:     %d\n", preview.FolderCount)

	if len(preview.Folders) > 0 {
		fmt.Printf("\nFolders:\n")
		for _, f := range preview.Folders {
			fmt.Printf("  - %s (%d requests)\n", f.Name, f.RequestCount)
		}
	}

	if len(preview.Servers) > 0 {
		fmt.Printf("\nServers:\n")
		for _, s := range preview.Servers {
			fmt.Printf("  - %s\n", s)
		}
	}

	if len(preview.Warnings) > 0 {
		fmt.Printf("\nWarnings:\n")
		for _, w := range preview.Warnings {
			fmt.Printf("  ! %s\n", w)
		}
	}

	fmt.Printf("\n(dry-run mode - no files created)\n")
	return nil
}

// outputResult outputs the import result
func outputResult(cmd *ImportCommand, result ImportResult) error {
	if cmd.JSONOutput {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Determine import type from context
	importType := "file"
	if result.FolderCount > 0 || result.RequestCount > 0 {
		importType = "collection"
	}

	fmt.Printf("Successfully imported %s\n\n", importType)
	fmt.Printf("Name: %s\n", result.CollectionName)
	fmt.Printf("File: %s\n", result.FilePath)

	if result.FolderCount > 0 || result.RequestCount > 0 {
		fmt.Printf("Folders:  %d\n", result.FolderCount)
		fmt.Printf("Requests: %d\n", result.RequestCount)
	}

	if len(result.Warnings) > 0 {
		fmt.Printf("\nWarnings:\n")
		for _, w := range result.Warnings {
			fmt.Printf("  ! %s\n", w)
		}
	}

	return nil
}

// sanitizeFilename converts a name to a valid filename
func sanitizeFilename(name string) string {
	// Replace spaces and special characters
	result := make([]rune, 0, len(name))
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result = append(result, r)
		} else if r == ' ' {
			result = append(result, '-')
		}
	}

	filename := string(result)
	if filename == "" {
		filename = "imported-api"
	}

	return filename
}

// countCollectionRequests counts all requests in a collection
func countCollectionRequests(c *api.CollectionFile) int {
	count := len(c.Requests)
	for _, folder := range c.Folders {
		count += countFolderRequests(&folder)
	}
	return count
}

// countFolderRequests counts requests in a folder recursively
func countFolderRequests(f *api.Folder) int {
	count := len(f.Requests)
	for _, subfolder := range f.Folders {
		count += countFolderRequests(&subfolder)
	}
	return count
}
