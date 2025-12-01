package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/kbrdn1/LazyCurl/internal/config"
	"github.com/kbrdn1/LazyCurl/internal/ui/components"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// PanelType represents the type of panel
type PanelType int

// requestDialogContext holds context for Request panel dialogs
type requestDialogContext struct {
	Tab   string
	Index int
	Key   string
	Value string
}

const (
	CollectionsPanel PanelType = iota
	RequestPanel
	ResponsePanel
	EnvironmentsPanel
)

// Model represents the main application model
type Model struct {
	globalConfig    *config.GlobalConfig
	workspaceConfig *config.WorkspaceConfig
	workspacePath   string

	width        int
	height       int
	activePanel  PanelType
	ready        bool
	zoneManager  *zone.Manager

	// Panels
	leftPanel     *LeftPanel
	requestPanel  *RequestView
	responsePanel *ResponseView

	// Mode system
	mode         Mode
	statusBar    *StatusBar
	commandInput *CommandInput

	// Dialog and WhichKey
	dialog   *components.Dialog
	whichKey *components.WhichKey
}

// NewModel creates a new application model
func NewModel(globalConfig *config.GlobalConfig, workspaceConfig *config.WorkspaceConfig, workspacePath string) Model {
	zm := zone.New()

	return Model{
		globalConfig:    globalConfig,
		workspaceConfig: workspaceConfig,
		workspacePath:   workspacePath,
		activePanel:     CollectionsPanel,
		zoneManager:     zm,
		leftPanel:       NewLeftPanel(workspacePath),
		requestPanel:    NewRequestView(),
		responsePanel:   NewResponseView(),
		mode:            NormalMode,
		statusBar:       NewStatusBar("v1.0.0"),
		commandInput:    NewCommandInput(),
		dialog:          components.NewDialog(),
		whichKey:        components.NewWhichKey(),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Update WhichKey context based on current state
	m.updateWhichKeyContext()

	// Handle WhichKey modal input first if visible
	if m.whichKey.IsVisible() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			m.whichKey, _ = m.whichKey.Update(msg)
		}
		return m, nil
	}

	// Handle environment modal input first if visible
	if m.leftPanel.GetEnvironments().HasActiveModal() {
		*m.leftPanel.GetEnvironments(), _ = m.leftPanel.GetEnvironments().Update(msg, m.globalConfig)
		return m, nil
	}

	// Handle dialog input first if visible
	if m.dialog.IsVisible() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			dialog, cmd := m.dialog.Update(msg)
			m.dialog = dialog
			return m, cmd
		case components.DialogResultMsg:
			return m.handleDialogResult(msg)
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case components.ModalCloseMsg:
		// Forward modal close messages to environments view
		if m.leftPanel.GetActiveTab() == EnvironmentsTab {
			*m.leftPanel.GetEnvironments(), _ = m.leftPanel.GetEnvironments().Update(msg, m.globalConfig)
		}
		// Force a refresh by sending a nil window size (triggers re-render)
		return m, func() tea.Msg {
			return tea.WindowSizeMsg{Width: m.width, Height: m.height}
		}

	case components.SearchUpdateMsg, components.SearchCloseMsg:
		// Forward search messages to left panel
		var cmd tea.Cmd
		*m.leftPanel, cmd = m.leftPanel.Update(msg, m.globalConfig)
		return m, cmd

	case components.DialogResultMsg:
		return m.handleDialogResult(msg)

	case tea.KeyMsg:
		// CTRL+C always quits
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle COMMAND mode input first (forward all keys except escape)
		if m.mode == CommandMode {
			if msg.String() == "esc" {
				// Exit COMMAND mode
				m.commandInput.Hide()
				m.mode = NormalMode
				m.statusBar.SetMode(NormalMode)
				return m, func() tea.Msg {
					return ModeChangeMsg{From: CommandMode, To: NormalMode}
				}
			}
			// Forward key to command input
			var cmd tea.Cmd
			m.commandInput, cmd = m.commandInput.Update(msg)
			return m, cmd
		}

		// Handle Escape key - return to NORMAL mode from any mode
		if msg.String() == "esc" {
			if m.mode != NormalMode {
				oldMode := m.mode
				m.mode = NormalMode
				m.statusBar.SetMode(NormalMode)
				return m, func() tea.Msg {
					return ModeChangeMsg{From: oldMode, To: NormalMode}
				}
			}
		}

		// Check if request panel is editing URL - if so, forward all keys to it
		if m.activePanel == RequestPanel && m.requestPanel.IsEditingURL() {
			var cmd tea.Cmd
			*m.requestPanel, cmd = m.requestPanel.Update(msg, m.globalConfig)
			return m, cmd
		}

		// Check if request panel Body or Scripts tab is active - forward ALL keys to editor
		// The editor has its own vim-like modes (NORMAL/INSERT) and handles q, h, l, etc.
		// This MUST return to prevent quit handler from catching 'q'
		if m.activePanel == RequestPanel && m.requestPanel.IsEditorActive() {
			var cmd tea.Cmd
			*m.requestPanel, cmd = m.requestPanel.Update(msg, m.globalConfig)
			return m, cmd
		}

		// Handle mode transitions from NORMAL mode
		if m.mode == NormalMode {
			switch msg.String() {
			case "c", "i":
				// In Collections panel, c/i edits the selected request
				// In other panels, i enters INSERT mode
				if m.activePanel == CollectionsPanel {
					// Let the panel handle c/i for editing requests
					// This will be forwarded to tree.go which emits TreeEditRequestMsg
					break
				}
				// Only 'i' enters INSERT mode (not 'c')
				if msg.String() == "i" {
					m.mode = InsertMode
					m.statusBar.SetMode(InsertMode)
					return m, func() tea.Msg {
						return ModeChangeMsg{From: NormalMode, To: InsertMode}
					}
				}
			case "v":
				// Transition to VIEW mode
				m.mode = ViewMode
				m.statusBar.SetMode(ViewMode)
				return m, func() tea.Msg {
					return ModeChangeMsg{From: NormalMode, To: ViewMode}
				}
			case ":":
				// Transition to COMMAND mode and show input
				m.mode = CommandMode
				m.statusBar.SetMode(CommandMode)
				m.commandInput.Show()
				return m, func() tea.Msg {
					return ModeChangeMsg{From: NormalMode, To: CommandMode}
				}
			}

			// Check for quit in NORMAL mode
			if m.matchKey(msg.String(), m.globalConfig.KeyBindings.Quit) {
				return m, tea.Quit
			}

			// ? to show WhichKey modal
			if msg.String() == "?" {
				m.whichKey.Show()
				return m, nil
			}

			// Tab switching with 1/2 (when left panel is active)
			if m.activePanel == CollectionsPanel {
				if msg.String() == "1" {
					m.leftPanel.SetActiveTab(CollectionsTab)
					return m, nil
				}
				if msg.String() == "2" {
					m.leftPanel.SetActiveTab(EnvironmentsTab)
					return m, nil
				}
			}

			// Panel navigation with h/l only in NORMAL mode
			// Skip navigation if search is active in the left panel
			// Note: Body tab is handled earlier and returns before reaching here
			if m.mode.AllowsNavigation() && !m.leftPanel.IsSearching() {
				if m.matchKey(msg.String(), m.globalConfig.KeyBindings.NavigateLeft) {
					if m.activePanel > CollectionsPanel {
						m.activePanel--
					}
					return m, nil
				}
				if m.matchKey(msg.String(), m.globalConfig.KeyBindings.NavigateRight) {
					if m.activePanel < ResponsePanel {
						m.activePanel++
					}
					return m, nil
				}
			}
		}

		// Handle VIEW mode navigation (read-only browsing)
		if m.mode == ViewMode {
			if m.mode.AllowsNavigation() {
				if m.matchKey(msg.String(), m.globalConfig.KeyBindings.NavigateLeft) {
					if m.activePanel > CollectionsPanel {
						m.activePanel--
					}
					return m, nil
				}
				if m.matchKey(msg.String(), m.globalConfig.KeyBindings.NavigateRight) {
					if m.activePanel < ResponsePanel {
						m.activePanel++
					}
					return m, nil
				}
			}
		}

	case ModeChangeMsg:
		// Handle mode change side effects
		m.statusBar.SetMode(msg.To)
		return m, nil

	case components.TreeSelectionMsg:
		// Handle request selection from tree
		if msg.Node != nil && msg.Node.Type == components.RequestNode {
			// Load the selected request into the request panel
			m.requestPanel.LoadRequest(msg.Node.ID, msg.Node.Name, msg.Node.HTTPMethod, msg.Node.URL)

			// Focus the Request Panel
			m.activePanel = RequestPanel

			// Update status bar with method and breadcrumb
			m.statusBar.SetMethod(msg.Node.HTTPMethod)
			breadcrumb := buildBreadcrumb(msg.Node)
			m.statusBar.SetBreadcrumb(breadcrumb...)
		}
		return m, nil

	case components.TreeRenameMsg:
		// Handle rename request - show input dialog
		if msg.Node != nil {
			m.dialog.ShowInput(
				"Rename",
				"Enter new name:",
				msg.Node.Name,
				"rename",
				msg.Node,
			)
		}
		return m, nil

	case components.TreeDeleteMsg:
		// Handle delete request - show confirmation dialog
		if msg.Node != nil {
			nodeType := "item"
			switch msg.Node.Type {
			case components.CollectionNode:
				nodeType = "collection"
			case components.FolderNode:
				nodeType = "folder"
			case components.RequestNode:
				nodeType = "request"
			}
			m.dialog.ShowConfirm(
				"Delete "+nodeType,
				"Are you sure you want to delete '"+msg.Node.Name+"'?",
				"delete",
				msg.Node,
			)
		}
		return m, nil

	case components.TreeNewRequestMsg:
		// Handle new request creation - show new request dialog
		m.dialog.ShowNewRequest("new_request", msg.ParentNode)
		return m, nil

	case components.TreeNewFolderMsg:
		// Handle new folder creation - show input dialog
		m.dialog.ShowInput(
			"New Folder",
			"Enter folder name:",
			"New Folder",
			"new_folder",
			msg.ParentNode,
		)
		return m, nil

	case components.TreeDuplicateMsg:
		// Handle duplicate request
		if msg.Node != nil {
			m.performDuplicate(msg.Node)
		}
		return m, nil

	case components.TreeEditRequestMsg:
		// Handle edit request - show edit request dialog
		if msg.Node != nil && msg.Node.Type == components.RequestNode {
			m.dialog.ShowEditRequest(msg.Node)
		}
		return m, nil

	case components.TreeYankMsg:
		// Handle yank (copy) to clipboard
		if msg.Node != nil {
			m.leftPanel.GetCollections().SetClipboard(msg.Node)
			m.statusBar.Success("Yanked", msg.Node.Name)
		}
		return m, nil

	case components.TreePasteMsg:
		// Handle paste from clipboard with smart targeting
		clipboard := m.leftPanel.GetCollections().GetClipboard()
		if clipboard == nil {
			m.statusBar.Info("Nothing to paste")
			return m, nil
		}

		if err := m.leftPanel.GetCollections().PasteNode(clipboard, msg.TargetNode); err != nil {
			m.statusBar.Error(err)
			return m, nil
		}

		m.statusBar.Success("Pasted", clipboard.Name)
		m.leftPanel.GetCollections().ReloadCollections()
		return m, nil

	// === REQUEST PANEL MESSAGES ===

	case RequestRenameMsg:
		// Handle rename key - show input dialog
		m.dialog.ShowInput(
			"Rename Key",
			"Enter new key name:",
			msg.Key,
			"request_rename",
			&requestDialogContext{Tab: msg.Tab, Index: msg.Index, Key: msg.Key, Value: msg.Value},
		)
		return m, nil

	case RequestDeleteMsg:
		// Handle delete - show confirmation dialog
		m.dialog.ShowConfirm(
			"Delete Entry",
			"Are you sure you want to delete '"+msg.Key+"'?",
			"request_delete",
			&requestDialogContext{Tab: msg.Tab, Index: msg.Index, Key: msg.Key},
		)
		return m, nil

	case RequestEditMsg:
		// Handle edit - show key-value input dialog
		m.dialog.ShowKeyValue(
			"Edit Entry",
			msg.Key,
			msg.Value,
			"request_edit",
			&requestDialogContext{Tab: msg.Tab, Index: msg.Index},
		)
		return m, nil

	case RequestNewMsg:
		// Handle new entry - show key-value input dialog
		m.dialog.ShowKeyValue(
			"New Entry",
			"",
			"",
			"request_new",
			&requestDialogContext{Tab: msg.Tab},
		)
		return m, nil

	case RequestDuplicateMsg:
		// Handle duplicate - directly duplicate without dialog
		m.requestPanel.DuplicateRow(msg.Index)
		m.statusBar.Success("Duplicated", "entry")
		return m, nil

	case RequestYankMsg:
		// Handle yank (copy) to clipboard
		m.requestPanel.SetClipboard(msg.Key, msg.Value)
		m.statusBar.Success("Yanked", msg.Key)
		return m, nil

	case RequestPasteMsg:
		// Handle paste from clipboard
		clipboard := m.requestPanel.GetClipboard()
		if clipboard == nil {
			m.statusBar.Info("Nothing to paste")
			return m, nil
		}
		m.requestPanel.AddRow(clipboard.Key+"_copy", clipboard.Value)
		m.statusBar.Success("Pasted", clipboard.Key)
		return m, nil

	case RequestURLChangedMsg:
		// Handle URL change from request panel
		requestID := m.requestPanel.GetCurrentRequestID()
		if requestID != "" {
			if err := m.leftPanel.GetCollections().UpdateRequestURLByID(requestID, msg.URL); err != nil {
				m.statusBar.Error(err)
			} else {
				m.statusBar.Success("URL saved", "")
				m.leftPanel.GetCollections().ReloadCollections()
			}
		}
		return m, nil

	case RequestParamToggleMsg:
		// Handle param toggle - sync URL and save
		if msg.Tab == "Params" {
			m.syncParamsAndSave()
		}
		return m, nil

	case RequestBodyChangedMsg:
		// Handle body content change - save to collection
		requestID := m.requestPanel.GetCurrentRequestID()
		if requestID != "" {
			if err := m.leftPanel.GetCollections().UpdateRequestBodyByID(requestID, msg.BodyType, msg.Content); err != nil {
				m.statusBar.Error(err)
			}
		}
		return m, nil

	case RequestScriptsChangedMsg:
		// Handle scripts content change - save to collection
		requestID := m.requestPanel.GetCurrentRequestID()
		if requestID != "" {
			if err := m.leftPanel.GetCollections().UpdateRequestScriptsByID(requestID, msg.PreRequest, msg.PostRequest); err != nil {
				m.statusBar.Error(err)
			}
		}
		return m, nil

	case RequestAuthChangedMsg:
		// Handle auth configuration change - save to collection
		requestID := m.requestPanel.GetCurrentRequestID()
		if requestID != "" {
			if err := m.leftPanel.GetCollections().UpdateRequestAuthByID(requestID, msg.Auth); err != nil {
				m.statusBar.Error(err)
			}
		}
		return m, nil

	case CommandExecuteMsg:
		// Handle command execution
		return m.handleCommand(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil
	}

	// Update active panel (pass mode context)
	var cmd tea.Cmd
	switch m.activePanel {
	case CollectionsPanel:
		*m.leftPanel, cmd = m.leftPanel.Update(msg, m.globalConfig)
	case RequestPanel:
		*m.requestPanel, cmd = m.requestPanel.Update(msg, m.globalConfig)
	case ResponsePanel:
		*m.responsePanel, cmd = m.responsePanel.Update(msg, m.globalConfig)
	}

	return m, cmd
}

// Minimum terminal size constants
const (
	MinTerminalWidth  = 80
	MinTerminalHeight = 24
)

// View renders the model
func (m Model) View() string {
	if !m.ready {
		return "Initializing LazyCurl..."
	}

	// Check minimum terminal size
	if m.width < MinTerminalWidth || m.height < MinTerminalHeight {
		warningStyle := lipgloss.NewStyle().
			Foreground(styles.Yellow).
			Bold(true)
		return warningStyle.Render("Terminal too small. Please resize to at least 80x24.")
	}

	// Calculate panel dimensions
	// Reserve 1 line for status bar
	contentHeight := m.height - 4

	// Lazygit-style layout:
	// +------------------+------------------+
	// |   Collections    |    Request       |
	// |   (left 1/3)     |    (right 2/3)   |
	// |                  +------------------+
	// |                  |    Response      |
	// +------------------+------------------+
	// |         Status Bar                  |
	// +-------------------------------------+

	// Main 3-panel layout - simplified borders
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 1 // -1 to prevent overflow

	// Better proportions: 40% request, 60% response
	topRightHeight := (contentHeight * 2) / 5
	bottomRightHeight := contentHeight - topRightHeight

	// Left panel (Collections/Env with tabs)
	leftContent := m.leftPanel.View(
		leftWidth-4,
		contentHeight-2,
		m.activePanel == CollectionsPanel,
	)
	leftPanelRendered := m.renderPanelWithTabs(m.leftPanel, leftContent, leftWidth, contentHeight, m.activePanel == CollectionsPanel)

	// Request panel (top right)
	requestContent := m.requestPanel.View(
		rightWidth-4,
		topRightHeight-2,
		m.activePanel == RequestPanel,
	)
	requestPanel := m.renderPanel("Request", requestContent, rightWidth, topRightHeight, m.activePanel == RequestPanel)

	// Response panel (bottom right)
	responseContent := m.responsePanel.View(
		rightWidth-4,
		bottomRightHeight-2,
		m.activePanel == ResponsePanel,
	)
	responsePanel := m.renderPanel("Response", responseContent, rightWidth, bottomRightHeight, m.activePanel == ResponsePanel)

	// Combine right panels vertically - no extra spacing
	rightSide := requestPanel + "\n" + responsePanel

	// Combine left and right horizontally
	mainContent := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanelRendered,
		rightSide,
	)

	// Status bar or command input
	var bottomBar string
	if m.commandInput.IsVisible() {
		bottomBar = m.commandInput.View(m.width)
	} else {
		bottomBar = m.renderStatusBar()
	}

	// Join without extra spacing
	result := mainContent + "\n" + bottomBar

	// Overlay dialog if visible
	if m.dialog.IsVisible() {
		dialogView := m.dialog.View(m.width, m.height)
		// Place dialog in center of screen
		result = m.overlayDialog(result, dialogView)
	}

	// Overlay environment modal if visible
	if m.leftPanel.GetEnvironments().HasActiveModal() {
		modalView := m.leftPanel.GetEnvironments().RenderModal(m.width, m.height)
		if modalView != "" {
			result = m.overlayDialog(result, modalView)
		}
	}

	// Overlay WhichKey modal if visible
	if m.whichKey.IsVisible() {
		whichKeyView := m.whichKey.View(m.width, m.height)
		result = m.overlayDialog(result, whichKeyView)
	}

	return result
}

// renderPanelWithTabs renders a panel with tab support in the title bar
func (m Model) renderPanelWithTabs(lp *LeftPanel, content string, width, height int, active bool) string {
	var borderColor lipgloss.Color

	if active {
		borderColor = styles.Lavender
	} else {
		borderColor = styles.Surface0
	}

	borderChar := lipgloss.NewStyle().Foreground(borderColor)

	// Build the top border with tabs
	// Format: ╭─Collections─Env───────────────╮
	innerWidth := width - 2 // Account for corners (╭ and ╮)
	tabsContent := lp.RenderTabs(innerWidth, active, borderColor)

	topBorder := borderChar.Render("╭") + tabsContent + borderChar.Render("╮")

	// Build the content area
	contentStyle := lipgloss.NewStyle().
		Width(width - 4).
		Height(height - 2)

	styledContent := contentStyle.Render(content)

	// Split content into lines and add side borders
	contentLines := strings.Split(styledContent, "\n")
	var borderedContent strings.Builder

	for i := 0; i < height-2; i++ {
		line := ""
		if i < len(contentLines) {
			line = contentLines[i]
		}
		// Pad line to width
		lineWidth := lipgloss.Width(line)
		padding := width - 4 - lineWidth
		if padding < 0 {
			padding = 0
		}
		borderedContent.WriteString(borderChar.Render("│") + " " + line + strings.Repeat(" ", padding) + " " + borderChar.Render("│") + "\n")
	}

	// Build bottom border
	bottomBorder := borderChar.Render("╰") +
		borderChar.Render(strings.Repeat("─", width-2)) +
		borderChar.Render("╯")

	return topBorder + "\n" + borderedContent.String() + bottomBorder
}

func (m Model) renderPanel(title string, content string, width, height int, active bool) string {
	var borderColor lipgloss.Color
	var titleFg lipgloss.Color

	if active {
		borderColor = styles.Lavender
		titleFg = styles.Lavender
	} else {
		borderColor = styles.Surface0
		titleFg = styles.Subtext0
	}

	// Build the top border with embedded title
	// Format: ╭─ Title ─────────────────────╮
	titleText := " " + title + " "
	titleStyled := lipgloss.NewStyle().
		Foreground(titleFg).
		Bold(true).
		Render(titleText)

	// Calculate border segments
	innerWidth := width - 2 // Account for corners (╭ and ╮)
	titleWidth := lipgloss.Width(titleStyled)
	leftPadding := 1 // Padding after corner
	rightDashes := innerWidth - leftPadding - titleWidth
	if rightDashes < 0 {
		rightDashes = 0
	}

	borderChar := lipgloss.NewStyle().Foreground(borderColor)

	topBorder := borderChar.Render("╭") +
		borderChar.Render(strings.Repeat("─", leftPadding)) +
		titleStyled +
		borderChar.Render(strings.Repeat("─", rightDashes)) +
		borderChar.Render("╮")

	// Build the content area
	contentStyle := lipgloss.NewStyle().
		Width(width - 4).
		Height(height - 2)

	styledContent := contentStyle.Render(content)

	// Split content into lines and add side borders
	contentLines := strings.Split(styledContent, "\n")
	var borderedContent strings.Builder

	for i := 0; i < height-2; i++ {
		line := ""
		if i < len(contentLines) {
			line = contentLines[i]
		}
		// Pad line to width
		lineWidth := lipgloss.Width(line)
		padding := width - 4 - lineWidth
		if padding < 0 {
			padding = 0
		}
		borderedContent.WriteString(borderChar.Render("│") + " " + line + strings.Repeat(" ", padding) + " " + borderChar.Render("│") + "\n")
	}

	// Build bottom border
	bottomBorder := borderChar.Render("╰") +
		borderChar.Render(strings.Repeat("─", width-2)) +
		borderChar.Render("╯")

	return topBorder + "\n" + borderedContent.String() + bottomBorder
}

func (m Model) renderStatusBar() string {
	// Update environment display
	m.statusBar.SetEnvironment(m.leftPanel.GetEnvironments().GetActiveEnvironmentName())

	// Update breadcrumb based on active tab
	if m.leftPanel.GetActiveTab() == EnvironmentsTab && m.activePanel == CollectionsPanel {
		m.statusBar.SetBreadcrumb(m.leftPanel.GetEnvironments().GetBreadcrumb()...)
	}

	// Update dynamic hints from WhichKey
	m.statusBar.SetHints(m.GetWhichKeyHints())

	// Use the new StatusBar component
	return m.statusBar.View(m.width)
}

// matchKey checks if a key matches any in the binding
func (m Model) matchKey(key string, bindings []string) bool {
	for _, binding := range bindings {
		if key == binding {
			return true
		}
	}
	return false
}

// buildBreadcrumb builds a breadcrumb path from a tree node
func buildBreadcrumb(node *components.TreeNode) []string {
	if node == nil {
		return []string{}
	}

	var parts []string

	// Walk up the tree to build breadcrumb
	current := node
	for current != nil {
		parts = append([]string{current.Name}, parts...)
		current = current.Parent
	}

	return parts
}

// handleCommand processes command input from COMMAND mode
func (m Model) handleCommand(msg CommandExecuteMsg) (tea.Model, tea.Cmd) {
	switch msg.Command {
	case CmdQuit, CmdQuitLong:
		// :q or :quit - exit application
		return m, tea.Quit

	case CmdWrite, CmdWriteLong:
		// :w or :write - save current request
		m.statusBar.Success("Saved", "request")
		return m, nil

	case CmdWriteQuit:
		// :wq - save and quit
		return m, tea.Quit

	case CmdWorkspace, CmdWorkspaceShort:
		// :workspace or :ws - workspace management
		return m.handleWorkspaceCommand(msg.Args)

	case CmdHelp:
		// :help - show help
		m.statusBar.Info(":q quit | :w save | :ws workspace | :env environments")
		return m, nil

	case CmdSet:
		// :set - set configuration
		if len(msg.Args) >= 2 {
			m.statusBar.Success("Set "+msg.Args[0], msg.Args[1])
		}
		return m, nil

	case CmdEnv:
		// :env - switch to environments tab
		m.leftPanel.SetActiveTab(EnvironmentsTab)
		m.activePanel = CollectionsPanel
		return m, nil

	case CmdCollections, CmdCollectionsShort:
		// :collections or :col - switch to collections tab
		m.leftPanel.SetActiveTab(CollectionsTab)
		m.activePanel = CollectionsPanel
		return m, nil

	default:
		// Unknown command
		m.statusBar.Info("Unknown command: " + msg.Command)
		return m, nil
	}
}

// handleWorkspaceCommand processes workspace subcommands
func (m Model) handleWorkspaceCommand(args []string) (tea.Model, tea.Cmd) {
	if len(args) == 0 {
		// Show current workspace
		m.statusBar.Success("Workspace", m.workspaceConfig.Name)
		return m, nil
	}

	switch args[0] {
	case WorkspaceList:
		// :workspace list - list all workspaces
		workspaces := m.globalConfig.Workspaces
		if len(workspaces) == 0 {
			m.statusBar.Info("No recent workspaces")
		} else {
			// Show first few workspaces
			msg := ""
			for i, ws := range workspaces {
				if i > 2 {
					msg += "..."
					break
				}
				if i > 0 {
					msg += ", "
				}
				msg += ws
			}
			m.statusBar.Success("Workspaces", msg)
		}
		return m, nil

	case WorkspaceSwitch:
		// :workspace switch <name> - switch workspace
		if len(args) < 2 {
			m.statusBar.Info("Usage: :ws switch <name>")
			return m, nil
		}
		// TODO: Implement actual workspace switching
		m.statusBar.Success("Switching", args[1])
		return m, nil

	case WorkspaceCreate:
		// :workspace create <name> - create new workspace
		if len(args) < 2 {
			m.statusBar.Info("Usage: :ws create <name>")
			return m, nil
		}
		// TODO: Implement actual workspace creation
		m.statusBar.Success("Created", args[1])
		return m, nil

	case WorkspaceDelete:
		// :workspace delete <name> - delete workspace
		if len(args) < 2 {
			m.statusBar.Info("Usage: :ws delete <name>")
			return m, nil
		}
		// TODO: Implement actual workspace deletion
		m.statusBar.Success("Deleted", args[1])
		return m, nil

	default:
		m.statusBar.Info("Unknown: " + args[0])
		return m, nil
	}
}

// handleDialogResult processes dialog results
func (m Model) handleDialogResult(msg components.DialogResultMsg) (tea.Model, tea.Cmd) {
	if !msg.Confirmed {
		m.statusBar.Info("Cancelled")
		return m, nil
	}

	switch msg.Action {
	case "rename":
		if msg.Node != nil && msg.Value != "" {
			m.performRename(msg.Node, msg.Value)
		}
	case "delete":
		if msg.Node != nil {
			m.performDelete(msg.Node)
		}
	case "new_request":
		if msg.Value != "" {
			m.performNewRequest(msg.Value, msg.Method, msg.URL, msg.Node)
		}
	case "new_folder":
		if msg.Value != "" {
			m.performNewFolder(msg.Value, msg.Node)
		}
	case "edit_request":
		if msg.Node != nil && msg.Value != "" {
			m.performEditRequest(msg.Node, msg.Value, msg.Method, msg.URL)
		}

	// === REQUEST PANEL ACTIONS ===
	case "request_rename":
		if ctx, ok := msg.Context.(*requestDialogContext); ok && msg.Value != "" {
			m.requestPanel.RenameRow(ctx.Index, msg.Value)
			m.statusBar.Success("Renamed", msg.Value)
			// Sync params to URL and save if Params or PathParams tab
			if ctx.Tab == "Params" {
				m.syncParamsAndSave()
			} else if ctx.Tab == "PathParams" {
				m.syncPathParamsAndSave(ctx.Index, msg.Value)
			}
		}
	case "request_delete":
		if ctx, ok := msg.Context.(*requestDialogContext); ok {
			m.requestPanel.DeleteRow(ctx.Index)
			m.statusBar.Success("Deleted", ctx.Key)
			// Sync params to URL and save if Params tab
			if ctx.Tab == "Params" {
				m.syncParamsAndSave()
			} else if ctx.Tab == "PathParams" {
				// Remove path param from URL
				m.removePathParamFromURL(ctx.Key)
			}
		}
	case "request_edit":
		if ctx, ok := msg.Context.(*requestDialogContext); ok && msg.Value != "" {
			// msg.Value = key, msg.URL = value (from key-value dialog)
			m.requestPanel.UpdateRow(ctx.Index, msg.Value, msg.URL)
			m.statusBar.Success("Updated", msg.Value)
			// Sync params to URL and save if Params tab
			if ctx.Tab == "Params" {
				m.syncParamsAndSave()
			}
			// Note: PathParams edit updates the value, not the key (which is in URL)
		}
	case "request_new":
		if ctx, ok := msg.Context.(*requestDialogContext); ok && msg.Value != "" {
			if ctx.Tab == "PathParams" {
				// For path params, add to URL and reload
				m.requestPanel.AddPathParamToURL(msg.Value)
				m.saveURLToCollection()
				m.statusBar.Success("Added path param", ":"+msg.Value)
			} else {
				// msg.Value = key, msg.URL = value (from key-value dialog)
				m.requestPanel.AddRow(msg.Value, msg.URL)
				m.statusBar.Success("Added", msg.Value)
				// Sync params to URL and save if Params tab
				if ctx.Tab == "Params" {
					m.syncParamsAndSave()
				}
			}
		}
	}

	return m, nil
}

// performRename renames a tree node
func (m *Model) performRename(node *components.TreeNode, newName string) {
	if node == nil || newName == "" {
		return
	}

	if err := m.leftPanel.GetCollections().RenameNode(node, newName); err != nil {
		m.statusBar.Error(err)
		return
	}

	m.leftPanel.GetCollections().ReloadCollections()
	m.statusBar.Success("Renamed", newName)
}

// performDelete deletes a tree node
func (m *Model) performDelete(node *components.TreeNode) {
	if node == nil {
		return
	}

	if err := m.leftPanel.GetCollections().DeleteNode(node); err != nil {
		m.statusBar.Error(err)
		return
	}

	m.statusBar.Success("Deleted", node.Name)
	m.leftPanel.GetCollections().ReloadCollections()
}

// performNewRequest creates a new request
func (m *Model) performNewRequest(name, method, url string, parent *components.TreeNode) {
	if err := m.leftPanel.GetCollections().AddRequestToCollection(name, method, url, parent); err != nil {
		m.statusBar.Error(err)
		return
	}

	m.statusBar.Success("Created", method+" "+name)
	m.leftPanel.GetCollections().ReloadCollections()
}

// performNewFolder creates a new folder
func (m *Model) performNewFolder(name string, parent *components.TreeNode) {
	if err := m.leftPanel.GetCollections().AddFolderToCollection(name, parent); err != nil {
		m.statusBar.Error(err)
		return
	}

	m.statusBar.Success("Created", name)
	m.leftPanel.GetCollections().ReloadCollections()
}

// performEditRequest updates a request's name, method, and URL
func (m *Model) performEditRequest(node *components.TreeNode, name, method, url string) {
	if node == nil || name == "" {
		return
	}

	if err := m.leftPanel.GetCollections().UpdateRequest(node, name, method, url); err != nil {
		m.statusBar.Error(err)
		return
	}

	m.statusBar.Success("Updated", method+" "+name)
	m.leftPanel.GetCollections().ReloadCollections()
}

// performDuplicate duplicates a tree node
func (m *Model) performDuplicate(node *components.TreeNode) {
	if node == nil {
		return
	}

	if err := m.leftPanel.GetCollections().DuplicateNode(node); err != nil {
		m.statusBar.Error(err)
		return
	}

	m.statusBar.Success("Duplicated", node.Name)
	m.leftPanel.GetCollections().ReloadCollections()
}

// syncParamsAndSave syncs the params table to URL and saves to collection
func (m *Model) syncParamsAndSave() {
	// Update URL from params
	newURL := m.requestPanel.SyncURLFromParams()

	// Save to collection
	requestID := m.requestPanel.GetCurrentRequestID()
	if requestID != "" {
		if err := m.leftPanel.GetCollections().UpdateRequestURLByID(requestID, newURL); err != nil {
			m.statusBar.Error(err)
			return
		}
		m.leftPanel.GetCollections().ReloadCollections()
	}
}

// syncPathParamsAndSave syncs a renamed path param to the URL and saves
func (m *Model) syncPathParamsAndSave(index int, newKey string) {
	// Get old key from path params table before rename
	pathParams := m.requestPanel.GetPathParamsTable()
	if pathParams == nil || index < 0 || index >= pathParams.RowCount() {
		return
	}

	// The row was already renamed, so we need to save the URL to collection
	m.saveURLToCollection()
}

// removePathParamFromURL removes a path param placeholder from the URL
func (m *Model) removePathParamFromURL(paramKey string) {
	url := m.requestPanel.GetURL()
	// Remove /:paramKey from URL
	newURL := strings.Replace(url, "/:"+paramKey, "", 1)
	// Also try removing just :paramKey (in case it's not prefixed with /)
	if newURL == url {
		newURL = strings.Replace(url, ":"+paramKey, "", 1)
	}

	// Update internal URL and save
	m.requestPanel.LoadRequest(
		m.requestPanel.GetCurrentRequestID(),
		"",   // name doesn't change
		m.requestPanel.GetMethod(),
		newURL,
	)
	m.saveURLToCollection()
}

// saveURLToCollection saves the current URL to the collection file
func (m *Model) saveURLToCollection() {
	requestID := m.requestPanel.GetCurrentRequestID()
	if requestID != "" {
		url := m.requestPanel.GetURL()
		if err := m.leftPanel.GetCollections().UpdateRequestURLByID(requestID, url); err != nil {
			m.statusBar.Error(err)
			return
		}
		m.leftPanel.GetCollections().ReloadCollections()
	}
}

// overlayDialog overlays a dialog centered on the background content
func (m Model) overlayDialog(background, dialog string) string {
	bgLines := strings.Split(background, "\n")
	dialogLines := strings.Split(dialog, "\n")

	dialogHeight := len(dialogLines)
	dialogWidth := lipgloss.Width(dialog)

	// Calculate center position
	startRow := (m.height - dialogHeight) / 2
	startCol := (m.width - dialogWidth) / 2
	if startRow < 0 {
		startRow = 0
	}
	if startCol < 0 {
		startCol = 0
	}

	// Ensure we have enough background lines
	for len(bgLines) < m.height {
		bgLines = append(bgLines, "")
	}

	// Overlay: keep bg visible, only replace where dialog appears
	for i, dialogLine := range dialogLines {
		bgIdx := startRow + i
		if bgIdx >= 0 && bgIdx < len(bgLines) {
			bgLine := bgLines[bgIdx]

			// Pad bg line to have enough width
			for lipgloss.Width(bgLine) < m.width {
				bgLine += " "
			}

			// We can't easily cut ANSI strings, so just use the dialog centered
			// with spaces preserving the visual width
			lineWidth := lipgloss.Width(dialogLine)
			leftPad := strings.Repeat(" ", startCol)
			rightPad := strings.Repeat(" ", m.width-startCol-lineWidth)
			if m.width-startCol-lineWidth < 0 {
				rightPad = ""
			}

			bgLines[bgIdx] = leftPad + dialogLine + rightPad
		}
	}

	return strings.Join(bgLines, "\n")
}

// updateWhichKeyContext updates the WhichKey context based on current state
func (m *Model) updateWhichKeyContext() {
	// Dialog context takes priority
	if m.dialog.IsVisible() {
		m.whichKey.SetContext(components.ContextDialog)
		return
	}

	// Modal context
	if m.leftPanel.GetEnvironments().HasActiveModal() {
		m.whichKey.SetContext(components.ContextModal)
		return
	}

	// Mode-based context
	switch m.mode {
	case InsertMode:
		m.whichKey.SetContext(components.ContextInsert)
	case ViewMode:
		m.whichKey.SetContext(components.ContextView)
	case CommandMode:
		m.whichKey.SetContext(components.ContextCommand)
	case NormalMode:
		// Panel-specific context
		switch m.activePanel {
		case CollectionsPanel:
			// Check for search context first
			if m.leftPanel.HasSearchQuery() {
				if m.leftPanel.GetActiveTab() == EnvironmentsTab {
					m.whichKey.SetContext(components.ContextSearchEnv)
				} else {
					m.whichKey.SetContext(components.ContextSearchCollections)
				}
			} else if m.leftPanel.GetActiveTab() == EnvironmentsTab {
				m.whichKey.SetContext(components.ContextNormalEnv)
			} else {
				m.whichKey.SetContext(components.ContextNormalCollections)
			}
		case RequestPanel:
			// Set context based on active tab
			switch m.requestPanel.GetActiveTab() {
			case "Params":
				m.whichKey.SetContext(components.ContextRequestParams)
			case "Authorization":
				m.whichKey.SetContext(components.ContextRequestAuth)
			case "Headers":
				m.whichKey.SetContext(components.ContextRequestHeaders)
			case "Body":
				m.whichKey.SetContext(components.ContextRequestBody)
			case "Scripts":
				m.whichKey.SetContext(components.ContextRequestScripts)
			default:
				m.whichKey.SetContext(components.ContextNormalRequest)
			}
		case ResponsePanel:
			m.whichKey.SetContext(components.ContextNormalResponse)
		default:
			m.whichKey.SetContext(components.ContextGlobal)
		}
	default:
		m.whichKey.SetContext(components.ContextGlobal)
	}
}

// GetWhichKeyHints returns the current WhichKey hints for the statusbar
func (m *Model) GetWhichKeyHints() string {
	return m.whichKey.GetHintsForStatusBar(m.whichKey.GetContext())
}
