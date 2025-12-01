package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// KeyBinding represents a single keybinding with its description
type KeyBinding struct {
	Key  string
	Desc string
}

// KeyGroup represents a group of related keybindings
type KeyGroup struct {
	Name     string
	Bindings []KeyBinding
}

// KeyContext represents a context with its keybindings
type KeyContext string

const (
	ContextGlobal            KeyContext = "global"
	ContextNormalCollections KeyContext = "normal_collections"
	ContextNormalEnv         KeyContext = "normal_env"
	ContextNormalRequest     KeyContext = "normal_request"
	ContextNormalResponse    KeyContext = "normal_response"
	ContextSearchCollections KeyContext = "search_collections"
	ContextSearchEnv         KeyContext = "search_env"
	ContextInsert            KeyContext = "insert"
	ContextView              KeyContext = "view"
	ContextCommand           KeyContext = "command"
	ContextDialog            KeyContext = "dialog"
	ContextModal             KeyContext = "modal"
)

// WhichKey manages keybinding hints display
type WhichKey struct {
	visible  bool
	context  KeyContext
	bindings map[KeyContext][]KeyGroup
}

// NewWhichKey creates a new WhichKey component
func NewWhichKey() *WhichKey {
	w := &WhichKey{
		bindings: make(map[KeyContext][]KeyGroup),
	}
	w.initBindings()
	return w
}

// initBindings initializes all keybindings by context
func (w *WhichKey) initBindings() {
	// Global bindings (always available)
	w.bindings[ContextGlobal] = []KeyGroup{
		{
			Name: "Mode",
			Bindings: []KeyBinding{
				{Key: "esc", Desc: "Normal"},
				{Key: ":", Desc: "Command"},
				{Key: "q", Desc: "Quit"},
			},
		},
	}

	// Normal mode - Collections panel
	w.bindings[ContextNormalCollections] = []KeyGroup{
		{
			Name: "Navigation",
			Bindings: []KeyBinding{
				{Key: "j/k", Desc: "Up/Down"},
				{Key: "h/l", Desc: "Collapse/Expand"},
				{Key: "g/G", Desc: "Top/Bottom"},
				{Key: "/", Desc: "Search"},
			},
		},
		{
			Name: "Tabs",
			Bindings: []KeyBinding{
				{Key: "1", Desc: "Collections"},
				{Key: "2", Desc: "Environments"},
			},
		},
		{
			Name: "Actions",
			Bindings: []KeyBinding{
				{Key: "n", Desc: "New Request"},
				{Key: "N", Desc: "New Folder"},
				{Key: "c/i", Desc: "Edit"},
				{Key: "R", Desc: "Rename"},
				{Key: "d", Desc: "Delete"},
				{Key: "D", Desc: "Duplicate"},
			},
		},
		{
			Name: "Clipboard",
			Bindings: []KeyBinding{
				{Key: "y", Desc: "Yank"},
				{Key: "p", Desc: "Paste"},
			},
		},
		{
			Name: "Help",
			Bindings: []KeyBinding{
				{Key: "?", Desc: "Show all keys"},
			},
		},
	}

	// Normal mode - Environments panel
	w.bindings[ContextNormalEnv] = []KeyGroup{
		{
			Name: "Navigation",
			Bindings: []KeyBinding{
				{Key: "j/k", Desc: "Up/Down"},
				{Key: "h/l", Desc: "Collapse/Expand"},
				{Key: "g/G", Desc: "Top/Bottom"},
				{Key: "/", Desc: "Search"},
			},
		},
		{
			Name: "Tabs",
			Bindings: []KeyBinding{
				{Key: "1", Desc: "Collections"},
				{Key: "2", Desc: "Environments"},
			},
		},
		{
			Name: "Actions",
			Bindings: []KeyBinding{
				{Key: "n", Desc: "New Variable"},
				{Key: "N", Desc: "New Environment"},
				{Key: "c/i", Desc: "Edit Value"},
				{Key: "R", Desc: "Rename"},
				{Key: "d", Desc: "Delete"},
				{Key: "D", Desc: "Duplicate"},
			},
		},
		{
			Name: "Toggle",
			Bindings: []KeyBinding{
				{Key: "a/A", Desc: "Active"},
				{Key: "s", Desc: "Secret"},
				{Key: "S/enter", Desc: "Select Env"},
			},
		},
		{
			Name: "Help",
			Bindings: []KeyBinding{
				{Key: "?", Desc: "Show all keys"},
			},
		},
	}

	// Normal mode - Request panel
	w.bindings[ContextNormalRequest] = []KeyGroup{
		{
			Name: "Tabs",
			Bindings: []KeyBinding{
				{Key: "tab", Desc: "Next tab"},
				{Key: "1-6", Desc: "Direct tab"},
			},
		},
		{
			Name: "Actions",
			Bindings: []KeyBinding{
				{Key: "ctrl+s", Desc: "Send"},
				{Key: "i", Desc: "Insert mode"},
			},
		},
		{
			Name: "Help",
			Bindings: []KeyBinding{
				{Key: "?", Desc: "Show all keys"},
			},
		},
	}

	// Normal mode - Response panel
	w.bindings[ContextNormalResponse] = []KeyGroup{
		{
			Name: "Tabs",
			Bindings: []KeyBinding{
				{Key: "tab", Desc: "Next tab"},
				{Key: "1-3", Desc: "Direct tab"},
			},
		},
		{
			Name: "Navigation",
			Bindings: []KeyBinding{
				{Key: "j/k", Desc: "Scroll"},
				{Key: "v", Desc: "View mode"},
			},
		},
		{
			Name: "Help",
			Bindings: []KeyBinding{
				{Key: "?", Desc: "Show all keys"},
			},
		},
	}

	// Search mode - Collections
	w.bindings[ContextSearchCollections] = []KeyGroup{
		{
			Name: "Navigation",
			Bindings: []KeyBinding{
				{Key: "n", Desc: "Next match"},
				{Key: "N", Desc: "Prev match"},
				{Key: "j/k", Desc: "Up/Down"},
			},
		},
		{
			Name: "Actions",
			Bindings: []KeyBinding{
				{Key: "enter/space", Desc: "Open request"},
				{Key: "i", Desc: "Edit search"},
				{Key: "esc", Desc: "Clear search"},
			},
		},
	}

	// Search mode - Environments
	w.bindings[ContextSearchEnv] = []KeyGroup{
		{
			Name: "Navigation",
			Bindings: []KeyBinding{
				{Key: "n", Desc: "Next match"},
				{Key: "N", Desc: "Prev match"},
				{Key: "j/k", Desc: "Up/Down"},
			},
		},
		{
			Name: "Actions",
			Bindings: []KeyBinding{
				{Key: "enter/space", Desc: "Open value"},
				{Key: "i", Desc: "Edit search"},
				{Key: "esc", Desc: "Clear search"},
			},
		},
	}

	// Insert mode
	w.bindings[ContextInsert] = []KeyGroup{
		{
			Name: "Input",
			Bindings: []KeyBinding{
				{Key: "type", Desc: "Enter text"},
				{Key: "tab", Desc: "Next field"},
				{Key: "esc", Desc: "Normal mode"},
			},
		},
	}

	// View mode
	w.bindings[ContextView] = []KeyGroup{
		{
			Name: "Navigation",
			Bindings: []KeyBinding{
				{Key: "j/k", Desc: "Scroll"},
				{Key: "g/G", Desc: "Top/Bottom"},
				{Key: "h/l", Desc: "Panel"},
				{Key: "esc", Desc: "Normal mode"},
			},
		},
	}

	// Command mode
	w.bindings[ContextCommand] = []KeyGroup{
		{
			Name: "Commands",
			Bindings: []KeyBinding{
				{Key: ":q", Desc: "Quit"},
				{Key: ":w", Desc: "Save"},
				{Key: ":ws", Desc: "Workspace"},
				{Key: ":help", Desc: "Help"},
				{Key: "esc", Desc: "Cancel"},
			},
		},
	}

	// Dialog context
	w.bindings[ContextDialog] = []KeyGroup{
		{
			Name: "Dialog",
			Bindings: []KeyBinding{
				{Key: "tab/↓", Desc: "Next field"},
				{Key: "enter", Desc: "Confirm"},
				{Key: "esc", Desc: "Cancel"},
			},
		},
	}

	// Modal context
	w.bindings[ContextModal] = []KeyGroup{
		{
			Name: "Modal",
			Bindings: []KeyBinding{
				{Key: "tab/j", Desc: "Next"},
				{Key: "enter", Desc: "Confirm"},
				{Key: "esc", Desc: "Cancel"},
			},
		},
	}
}

// Show displays the WhichKey modal
func (w *WhichKey) Show() {
	w.visible = true
}

// Hide hides the WhichKey modal
func (w *WhichKey) Hide() {
	w.visible = false
}

// Toggle toggles visibility
func (w *WhichKey) Toggle() {
	w.visible = !w.visible
}

// IsVisible returns visibility state
func (w *WhichKey) IsVisible() bool {
	return w.visible
}

// SetContext sets the current context
func (w *WhichKey) SetContext(ctx KeyContext) {
	w.context = ctx
}

// GetContext returns the current context
func (w *WhichKey) GetContext() KeyContext {
	return w.context
}

// Update handles messages
func (w *WhichKey) Update(msg tea.Msg) (*WhichKey, tea.Cmd) {
	if !w.visible {
		return w, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "?", "q":
			w.Hide()
		}
	}

	return w, nil
}

// GetHintsForStatusBar returns a compact hint string for the statusbar
func (w *WhichKey) GetHintsForStatusBar(ctx KeyContext) string {
	groups, ok := w.bindings[ctx]
	if !ok {
		return ""
	}

	var hints []string
	for _, group := range groups {
		for _, binding := range group.Bindings {
			// Skip "Show all keys" hint for statusbar
			if binding.Key == "?" {
				continue
			}
			hints = append(hints, binding.Key+":"+binding.Desc)
		}
	}

	// Limit to reasonable length
	result := strings.Join(hints, " │ ")
	if len(result) > 100 {
		// Take first few hints
		shortHints := hints[:min(5, len(hints))]
		result = strings.Join(shortHints, " │ ") + " │ ?:More"
	}

	return " " + result
}

// View renders the WhichKey modal
func (w *WhichKey) View(screenWidth, screenHeight int) string {
	if !w.visible {
		return ""
	}

	// Modal dimensions
	modalWidth := 60
	if modalWidth > screenWidth-4 {
		modalWidth = screenWidth - 4
	}

	// Build content
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Lavender).
		Width(modalWidth - 4).
		Align(lipgloss.Center)
	content.WriteString(titleStyle.Render("Keybindings"))
	content.WriteString("\n")

	// Context indicator
	ctxStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Italic(true).
		Width(modalWidth - 4).
		Align(lipgloss.Center)
	content.WriteString(ctxStyle.Render(string(w.context)))
	content.WriteString("\n\n")

	// Get bindings for current context
	groups, ok := w.bindings[w.context]
	if !ok {
		groups = w.bindings[ContextGlobal]
	}

	// Render groups
	groupStyle := lipgloss.NewStyle().
		Foreground(styles.Mauve).
		Bold(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(styles.Yellow).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(styles.Text)

	for _, group := range groups {
		content.WriteString(groupStyle.Render(group.Name))
		content.WriteString("\n")

		for _, binding := range group.Bindings {
			content.WriteString("  ")
			content.WriteString(keyStyle.Render(binding.Key))
			content.WriteString(" ")
			content.WriteString(descStyle.Render(binding.Desc))
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Footer
	footerStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Italic(true).
		Width(modalWidth - 4).
		Align(lipgloss.Center)
	content.WriteString(footerStyle.Render("Press ? or esc to close"))

	// Modal box style
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Lavender).
		Padding(1, 2).
		Width(modalWidth)

	return modalStyle.Render(content.String())
}

// min returns the minimum of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
