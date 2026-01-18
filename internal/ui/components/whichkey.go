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
	// Request panel tab contexts
	ContextRequestParams  KeyContext = "request_params"
	ContextRequestAuth    KeyContext = "request_auth"
	ContextRequestHeaders KeyContext = "request_headers"
	ContextRequestBody    KeyContext = "request_body"
	ContextRequestScripts KeyContext = "request_scripts"
	// Response panel tab contexts
	ContextConsole KeyContext = "console"
	// Jump mode context
	ContextJump KeyContext = "jump"
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
			Name: "Navigation",
			Bindings: []KeyBinding{
				{Key: "H/L", Desc: "Panel"},
				{Key: "tab", Desc: "Next tab"},
				{Key: "1-5", Desc: "Direct tab"},
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
			Name: "Navigation",
			Bindings: []KeyBinding{
				{Key: "j/k", Desc: "Up/Down"},
				{Key: "g/G", Desc: "Top/Bottom"},
				{Key: "H/L", Desc: "Panel"},
				{Key: "w/b", Desc: "Next/Prev word"},
			},
		},
		{
			Name: "Tabs",
			Bindings: []KeyBinding{
				{Key: "tab", Desc: "Next tab"},
				{Key: "1-5", Desc: "Direct tab"},
			},
		},
		{
			Name: "Clipboard",
			Bindings: []KeyBinding{
				{Key: "y", Desc: "Yank line"},
				{Key: "Y", Desc: "Yank all"},
				{Key: "Ctrl+C", Desc: "Copy all"},
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

	// Request panel tab contexts
	w.bindings[ContextRequestParams] = []KeyGroup{
		{
			Name: "Params",
			Bindings: []KeyBinding{
				{Key: "j/k", Desc: "Up/Down"},
				{Key: "h/l", Desc: "Section"},
				{Key: "n", Desc: "New param"},
				{Key: "c/i", Desc: "Edit"},
				{Key: "d", Desc: "Delete"},
				{Key: "space", Desc: "Toggle"},
				{Key: "H/L", Desc: "Panel"},
				{Key: "tab", Desc: "Next tab"},
			},
		},
	}

	w.bindings[ContextRequestAuth] = []KeyGroup{
		{
			Name: "Authorization",
			Bindings: []KeyBinding{
				{Key: "j/k", Desc: "Navigate"},
				{Key: "h/l", Desc: "Change type"},
				{Key: "i/c/Enter", Desc: "Edit"},
				{Key: "H/L", Desc: "Panel"},
				{Key: "tab", Desc: "Next tab"},
				{Key: "ctrl+s", Desc: "Send"},
			},
		},
	}

	w.bindings[ContextRequestHeaders] = []KeyGroup{
		{
			Name: "Headers",
			Bindings: []KeyBinding{
				{Key: "j/k", Desc: "Up/Down"},
				{Key: "n", Desc: "New header"},
				{Key: "c/i", Desc: "Edit"},
				{Key: "d", Desc: "Delete"},
				{Key: "space", Desc: "Toggle"},
				{Key: "H/L", Desc: "Panel"},
				{Key: "tab", Desc: "Next tab"},
			},
		},
	}

	w.bindings[ContextRequestBody] = []KeyGroup{
		{
			Name: "Body",
			Bindings: []KeyBinding{
				{Key: "h/l", Desc: "Cursor"},
				{Key: "j/k", Desc: "Up/Down"},
				{Key: "i", Desc: "Insert mode"},
				{Key: "ctrl+f", Desc: "Format"},
				{Key: "H/L", Desc: "Panel"},
				{Key: "tab", Desc: "Next tab"},
			},
		},
	}

	w.bindings[ContextRequestScripts] = []KeyGroup{
		{
			Name: "Scripts",
			Bindings: []KeyBinding{
				{Key: "h/l", Desc: "Cursor"},
				{Key: "j/k", Desc: "Up/Down"},
				{Key: "[/]", Desc: "Section"},
				{Key: "i", Desc: "Insert mode"},
				{Key: "H/L", Desc: "Panel"},
				{Key: "tab", Desc: "Next tab"},
			},
		},
	}

	// Console tab context
	w.bindings[ContextConsole] = []KeyGroup{
		{
			Name: "Navigation",
			Bindings: []KeyBinding{
				{Key: "j/k", Desc: "Up/Down"},
				{Key: "g/G", Desc: "Top/Bottom"},
				{Key: "enter", Desc: "Expand/Collapse"},
			},
		},
		{
			Name: "Actions",
			Bindings: []KeyBinding{
				{Key: "R", Desc: "Resend request"},
			},
		},
		{
			Name: "Copy",
			Bindings: []KeyBinding{
				{Key: "U", Desc: "Copy URL"},
				{Key: "H", Desc: "Copy headers"},
				{Key: "B", Desc: "Copy body"},
				{Key: "C", Desc: "Copy cookies"},
				{Key: "I", Desc: "Copy info"},
				{Key: "E", Desc: "Copy error"},
				{Key: "A", Desc: "Copy all"},
			},
		},
		{
			Name: "Help",
			Bindings: []KeyBinding{
				{Key: "?", Desc: "Show all keys"},
			},
		},
	}

	// Jump mode context
	w.bindings[ContextJump] = []KeyGroup{
		{
			Name: "Navigation",
			Bindings: []KeyBinding{
				{Key: "char", Desc: "Type label"},
				{Key: "esc", Desc: "Cancel"},
			},
		},
		{
			Name: "Help",
			Bindings: []KeyBinding{
				{Key: "?", Desc: "Show all keys"},
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
