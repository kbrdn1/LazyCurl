package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kbrdn1/LazyCurl/internal/config"
	"github.com/kbrdn1/LazyCurl/internal/ui"
)

func main() {
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
