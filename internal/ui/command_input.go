package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// CommandInput handles command mode text input with history
type CommandInput struct {
	input        string   // Current input text
	cursor       int      // Cursor position
	visible      bool     // Whether input is visible
	history      []string // Command history
	historyIndex int      // Current position in history (-1 = not browsing)
	tempInput    string   // Temporary storage for current input when browsing history
}

// CommandExecuteMsg is sent when a command is submitted
type CommandExecuteMsg struct {
	Command string   // The command name
	Args    []string // Command arguments
	Raw     string   // Raw input string
}

// NewCommandInput creates a new command input
func NewCommandInput() *CommandInput {
	return &CommandInput{
		input:        "",
		cursor:       0,
		visible:      false,
		history:      []string{},
		historyIndex: -1,
		tempInput:    "",
	}
}

// Show makes the command input visible and resets state
func (c *CommandInput) Show() {
	c.visible = true
	c.input = ""
	c.cursor = 0
	c.historyIndex = -1
	c.tempInput = ""
}

// Hide hides the command input
func (c *CommandInput) Hide() {
	c.visible = false
	c.input = ""
	c.cursor = 0
	c.historyIndex = -1
}

// IsVisible returns whether the command input is visible
func (c *CommandInput) IsVisible() bool {
	return c.visible
}

// Update handles messages for the command input
func (c *CommandInput) Update(msg tea.Msg) (*CommandInput, tea.Cmd) {
	if !c.visible {
		return c, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if c.input != "" {
				// Add to history
				c.history = append(c.history, c.input)
				// Parse and execute command
				cmd, args := parseCommand(c.input)
				raw := c.input
				c.Hide()
				return c, func() tea.Msg {
					return CommandExecuteMsg{
						Command: cmd,
						Args:    args,
						Raw:     raw,
					}
				}
			}
			c.Hide()
			return c, nil

		case "esc":
			c.Hide()
			return c, nil

		case "backspace":
			if c.cursor > 0 {
				c.input = c.input[:c.cursor-1] + c.input[c.cursor:]
				c.cursor--
			}

		case "delete":
			if c.cursor < len(c.input) {
				c.input = c.input[:c.cursor] + c.input[c.cursor+1:]
			}

		case "left":
			if c.cursor > 0 {
				c.cursor--
			}

		case "right":
			if c.cursor < len(c.input) {
				c.cursor++
			}

		case "home", "ctrl+a":
			c.cursor = 0

		case "end", "ctrl+e":
			c.cursor = len(c.input)

		case "up":
			// Navigate history backward
			if len(c.history) > 0 {
				if c.historyIndex == -1 {
					// Save current input before browsing history
					c.tempInput = c.input
					c.historyIndex = len(c.history) - 1
				} else if c.historyIndex > 0 {
					c.historyIndex--
				}
				c.input = c.history[c.historyIndex]
				c.cursor = len(c.input)
			}

		case "down":
			// Navigate history forward
			if c.historyIndex >= 0 {
				if c.historyIndex < len(c.history)-1 {
					c.historyIndex++
					c.input = c.history[c.historyIndex]
				} else {
					// Return to original input
					c.historyIndex = -1
					c.input = c.tempInput
				}
				c.cursor = len(c.input)
			}

		case "ctrl+u":
			// Clear line
			c.input = ""
			c.cursor = 0

		case "ctrl+w":
			// Delete word backward
			if c.cursor > 0 {
				// Find word boundary
				pos := c.cursor - 1
				for pos > 0 && c.input[pos] == ' ' {
					pos--
				}
				for pos > 0 && c.input[pos-1] != ' ' {
					pos--
				}
				c.input = c.input[:pos] + c.input[c.cursor:]
				c.cursor = pos
			}

		default:
			// Insert character
			if len(msg.String()) == 1 {
				char := msg.String()
				c.input = c.input[:c.cursor] + char + c.input[c.cursor:]
				c.cursor++
			}
		}
	}

	return c, nil
}

// View renders the command input
func (c *CommandInput) View(width int) string {
	if !c.visible {
		return ""
	}

	prefixStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Lavender)

	inputStyle := lipgloss.NewStyle().
		Foreground(styles.Text).
		Background(styles.Mantle)

	// Show cursor
	var displayInput string
	if c.cursor < len(c.input) {
		cursorStyle := lipgloss.NewStyle().
			Background(styles.Text).
			Foreground(styles.Base)
		displayInput = c.input[:c.cursor] + cursorStyle.Render(string(c.input[c.cursor])) + c.input[c.cursor+1:]
	} else {
		cursorStyle := lipgloss.NewStyle().
			Background(styles.Text).
			Foreground(styles.Base)
		displayInput = c.input + cursorStyle.Render(" ")
	}

	prefix := prefixStyle.Render(":")
	input := inputStyle.Width(width - 1).Render(displayInput)

	return prefix + input
}

// GetInput returns the current input text
func (c *CommandInput) GetInput() string {
	return c.input
}

// parseCommand parses a command string into command name and arguments
func parseCommand(input string) (string, []string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", nil
	}

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", nil
	}

	cmd := strings.ToLower(parts[0])
	args := parts[1:]

	return cmd, args
}

// Common command constants
const (
	CmdQuit             = "q"
	CmdQuitLong         = "quit"
	CmdWrite            = "w"
	CmdWriteLong        = "write"
	CmdWriteQuit        = "wq"
	CmdWorkspace        = "workspace"
	CmdWorkspaceShort   = "ws"
	CmdHelp             = "help"
	CmdSet              = "set"
	CmdEnv              = "env"
	CmdCollections      = "collections"
	CmdCollectionsShort = "col"
)

// Workspace subcommands
const (
	WorkspaceList   = "list"
	WorkspaceSwitch = "switch"
	WorkspaceCreate = "create"
	WorkspaceDelete = "delete"
)
