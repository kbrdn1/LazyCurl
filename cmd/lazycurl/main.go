package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kbrdn1/LazyCurl/internal/config"
	"github.com/kbrdn1/LazyCurl/internal/ui"
)

// Version information set by goreleaser ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Handle --version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("lazycurl %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	// Handle --help flag
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printHelp()
		os.Exit(0)
	}

	// Handle import subcommand
	if len(os.Args) > 1 && os.Args[1] == "import" {
		cmd, err := ParseImportArgs(os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		if err := RunImportCommand(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "Import failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Load global config
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		globalConfig = config.DefaultGlobalConfig()
	}

	// Get workspace path
	workspacePath, err := config.GetWorkspacePath()
	if err != nil {
		fmt.Printf("Error getting workspace path: %v\n", err)
		os.Exit(1)
	}

	// Load workspace config
	workspaceConfig, err := config.LoadWorkspaceConfig(workspacePath)
	if err != nil {
		fmt.Printf("Error loading workspace config: %v\n", err)
		workspaceConfig = config.DefaultWorkspaceConfig()
	}

	// Initialize the Bubble Tea program
	p := tea.NewProgram(
		ui.NewModel(globalConfig, workspaceConfig, workspacePath),
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// printHelp prints the help message
func printHelp() {
	fmt.Printf(`LazyCurl - A TUI HTTP client

Usage:
  lazycurl                         Start the TUI application
  lazycurl import <format> <file>  Import API specification
  lazycurl --version               Show version information
  lazycurl --help                  Show this help message

Commands:
  import    Import API specifications into collections

Import Formats:
  openapi   Import OpenAPI 3.x specification (JSON/YAML)

Import Options:
  --name NAME      Override collection name
  --output PATH    Custom output path for collection file
  --dry-run        Preview import without saving
  --json           Output results as JSON

Examples:
  lazycurl import openapi api.yaml
  lazycurl import openapi api.json --name "My API"
  lazycurl import openapi spec.yaml --dry-run
  lazycurl import openapi spec.yaml --json

Keyboard Shortcuts (TUI):
  Ctrl+O    Import OpenAPI specification
  Ctrl+I    Import cURL command
  Ctrl+S    Send request
  q         Quit

For more information, visit: https://github.com/kbrdn1/LazyCurl
`)
}
