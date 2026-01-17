package ui

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"golang.design/x/clipboard"

	"github.com/kbrdn1/LazyCurl/internal/api"
	"github.com/kbrdn1/LazyCurl/internal/config"
	"github.com/kbrdn1/LazyCurl/internal/session"
	"github.com/kbrdn1/LazyCurl/internal/ui/components"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// HTTPResponseMsg is sent when an HTTP request completes
type HTTPResponseMsg struct {
	Response *api.Response
	Error    error
}

// HTTPSendingMsg is sent when an HTTP request starts
type HTTPSendingMsg struct{}

// LoaderTickMsg is sent to animate the loader
type LoaderTickMsg struct{}

// loaderTickCmd returns a command that sends a tick for loader animation
func loaderTickCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return LoaderTickMsg{}
	})
}

// SendHTTPRequestCmd creates a command to send an HTTP request
func SendHTTPRequestCmd(req *api.Request) tea.Cmd {
	return func() tea.Msg {
		client := api.NewClient()
		resp, err := client.Send(req)
		return HTTPResponseMsg{Response: resp, Error: err}
	}
}

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

	width       int
	height      int
	activePanel PanelType
	ready       bool
	zoneManager *zone.Manager
	layoutMode  LayoutMode

	// Panels
	leftPanel     *LeftPanel
	requestPanel  *RequestView
	responsePanel *ResponseView

	// Mode system
	mode         Mode
	jumpMode     *JumpModeState
	statusBar    *StatusBar
	commandInput *CommandInput

	// Dialog and WhichKey
	dialog   *components.Dialog
	whichKey *components.WhichKey

	// HTTP client
	httpClient *api.Client
	isSending  bool

	// Fullscreen mode
	isFullscreen    bool
	fullscreenPanel PanelType

	// Console history
	consoleHistory *api.ConsoleHistory
	lastRequest    *api.Request // Track the last sent request for console logging
	requestStart   time.Time    // Track when request started for duration calculation

	// Session persistence
	session          *session.Session
	sessionDirtyTime time.Time

	// Import modal
	importModal *ImportModalModel
}

// NewModel creates a new application model
func NewModel(globalConfig *config.GlobalConfig, workspaceConfig *config.WorkspaceConfig, workspacePath string) Model {
	zm := zone.New()

	// Load session (returns default if missing/invalid)
	sess, _ := session.LoadSession(workspacePath)
	sess = sess.Validate(workspacePath)

	// Determine active panel from session
	activePanel := CollectionsPanel
	switch sess.ActivePanel {
	case "collections":
		activePanel = CollectionsPanel
	case "request":
		activePanel = RequestPanel
	case "response":
		activePanel = ResponsePanel
	}

	// Create panels
	leftPanel := NewLeftPanel(workspacePath)
	requestPanel := NewRequestView()
	responsePanel := NewResponseView()

	// Apply session state to panels
	leftPanel.SetSessionState(sess.Panels.Collections)
	requestPanel.SetSessionState(sess.Panels.Request)
	responsePanel.SetSessionState(sess.Panels.Response)

	// Restore active environment
	if sess.ActiveEnvironment != "" {
		leftPanel.GetEnvironments().SetActiveEnvironmentName(sess.ActiveEnvironment)
	}

	// Restore active request (find in tree and load FULL request from collection)
	if sess.ActiveRequest != "" {
		collections := leftPanel.GetCollections().GetCollections()
		for _, coll := range collections {
			if req := coll.FindRequest(sess.ActiveRequest); req != nil {
				requestPanel.LoadCollectionRequest(req)
				break
			}
		}
	}

	// Create status bar and set initial state
	statusBar := NewStatusBar("v1.0.0")
	if sess.ActiveEnvironment != "" {
		statusBar.SetEnvironment(sess.ActiveEnvironment)
	}

	return Model{
		globalConfig:    globalConfig,
		workspaceConfig: workspaceConfig,
		workspacePath:   workspacePath,
		activePanel:     activePanel,
		zoneManager:     zm,
		leftPanel:       leftPanel,
		requestPanel:    requestPanel,
		responsePanel:   responsePanel,
		mode:            NormalMode,
		jumpMode:        NewJumpMode(),
		statusBar:       statusBar,
		commandInput:    NewCommandInput(),
		dialog:          components.NewDialog(),
		whichKey:        components.NewWhichKey(),
		httpClient:      api.NewClient(),
		isSending:       false,
		consoleHistory:  api.NewConsoleHistory(1000),
		session:         sess,
		importModal:     NewImportModal(),
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Initialize clipboard (ignore error - clipboard may not be available on all systems)
	_ = clipboard.Init()
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

	// Handle import modal input first if visible
	if m.importModal.IsVisible() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			var cmd tea.Cmd
			m.importModal, cmd = m.importModal.Update(msg)
			return m, cmd
		case tea.WindowSizeMsg:
			m.importModal.SetSize(msg.Width, msg.Height)
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
		// Forward search messages to the appropriate panel
		var cmd tea.Cmd
		switch m.activePanel {
		case ResponsePanel:
			*m.responsePanel, cmd = m.responsePanel.Update(msg, m.globalConfig)
		case RequestPanel:
			*m.requestPanel, cmd = m.requestPanel.Update(msg, m.globalConfig)
		default:
			*m.leftPanel, cmd = m.leftPanel.Update(msg, m.globalConfig)
		}
		return m, cmd

	case components.EditorQuitMsg:
		// Editor requested to quit the application (Q key in NORMAL mode)
		return m.saveSessionAndQuit()

	case SessionSaveTickMsg:
		// Handle debounced session save
		// Only save if this tick matches the current dirty time (debounce)
		if !m.sessionDirtyTime.IsZero() && msg.DirtyTime.Equal(m.sessionDirtyTime) {
			m.saveSession()
			m.sessionDirtyTime = time.Time{} // Reset dirty time
		}
		return m, nil

	case components.DialogResultMsg:
		return m.handleDialogResult(msg)

	case tea.KeyMsg:
		// CTRL+C always quits (save session first)
		if msg.String() == "ctrl+c" {
			return m.saveSessionAndQuit()
		}

		// CTRL+S sends HTTP request from ANY context (global handler)
		if msg.String() == "ctrl+s" {
			return m.sendHTTPRequest()
		}

		// CTRL+I opens import cURL modal (global handler)
		if m.matchKey(msg.String(), m.globalConfig.KeyBindings.ImportCurl) {
			m.importModal.SetSize(m.width, m.height)
			m.importModal.Show()
			return m, nil
		}

		// CTRL+E exports current request as cURL (global handler)
		if m.matchKey(msg.String(), m.globalConfig.KeyBindings.ExportCurl) {
			return m.exportCurlCommand()
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

		// Handle Escape key - exit fullscreen, jump mode, or return to NORMAL mode
		if msg.String() == "esc" {
			// Exit jump mode first if active
			if m.jumpMode.IsActive() {
				m.jumpMode.Deactivate()
				m.mode = NormalMode
				m.statusBar.SetMode(NormalMode)
				return m, nil
			}
			// Exit fullscreen mode if active
			if m.isFullscreen {
				m.isFullscreen = false
				return m, nil
			}
			// Then handle mode changes
			if m.mode != NormalMode {
				oldMode := m.mode
				m.mode = NormalMode
				m.statusBar.SetMode(NormalMode)
				return m, func() tea.Msg {
					return ModeChangeMsg{From: oldMode, To: NormalMode}
				}
			}
		}

		// Handle Jump mode key input
		if m.jumpMode.IsActive() {
			// Tab cycles through panel scope (only in single-panel mode)
			if msg.String() == "tab" {
				newScope := m.jumpMode.CycleScopePanel()
				// Refresh targets for the new scope panel
				m.refreshJumpTargetsForScope(newScope)
				return m, nil
			}

			target, shouldCancel := m.jumpMode.HandleKey(msg.String())
			if target != nil {
				return m.jumpToTarget(*target)
			}
			if shouldCancel {
				m.jumpMode.Deactivate()
				m.mode = NormalMode
				m.statusBar.SetMode(NormalMode)
			}
			return m, nil
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
				return m.saveSessionAndQuit()
			}

			// ? to show WhichKey modal
			if msg.String() == "?" {
				m.whichKey.Show()
				return m, nil
			}

			// f to activate jump mode (current panel only)
			if msg.String() == "f" {
				return m.activateJumpMode(false)
			}

			// F (Shift+f) to activate cross-panel jump mode
			if msg.String() == "F" {
				return m.activateJumpMode(true)
			}

			// Z to toggle fullscreen (zoom) for current panel
			if msg.String() == "Z" {
				m.toggleFullscreen()
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
			// IMPORTANT: In CollectionsPanel, let the tree handle l/h first for expand/collapse
			if m.mode.AllowsNavigation() && !m.leftPanel.IsSearching() {
				// Left navigation (h) - in CollectionsPanel, only navigate if at root level collapsed folder
				if m.matchKey(msg.String(), m.globalConfig.KeyBindings.NavigateLeft) {
					// In CollectionsPanel, h should collapse folders, not navigate panels
					// Only navigate panels from Request or Response panels
					if m.activePanel > CollectionsPanel {
						m.activePanel--
						// Update fullscreen panel if in fullscreen mode
						if m.isFullscreen {
							m.fullscreenPanel = m.activePanel
						}
						return m, m.markSessionDirty()
					}
					// In CollectionsPanel, let tree handle h for collapse
				}
				// Right navigation (l) - in CollectionsPanel, let tree handle it
				if m.matchKey(msg.String(), m.globalConfig.KeyBindings.NavigateRight) {
					// In CollectionsPanel, l should expand folders or select requests
					// Only navigate panels from Request panel
					if m.activePanel == RequestPanel {
						m.activePanel++
						// Update fullscreen panel if in fullscreen mode
						if m.isFullscreen {
							m.fullscreenPanel = m.activePanel
						}
						return m, m.markSessionDirty()
					}
					// In CollectionsPanel, let tree handle l for expand/select
					// In ResponsePanel, we're already at the rightmost panel
				}
			}
		}

		// Handle VIEW mode navigation (read-only browsing)
		if m.mode == ViewMode {
			if m.mode.AllowsNavigation() {
				// Same logic as NORMAL mode - let tree handle h/l in CollectionsPanel
				if m.matchKey(msg.String(), m.globalConfig.KeyBindings.NavigateLeft) {
					if m.activePanel > CollectionsPanel {
						m.activePanel--
						if m.isFullscreen {
							m.fullscreenPanel = m.activePanel
						}
						return m, m.markSessionDirty()
					}
				}
				if m.matchKey(msg.String(), m.globalConfig.KeyBindings.NavigateRight) {
					if m.activePanel == RequestPanel {
						m.activePanel++
						if m.isFullscreen {
							m.fullscreenPanel = m.activePanel
						}
						return m, m.markSessionDirty()
					}
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
			// Find and load the FULL request from the collection
			collections := m.leftPanel.GetCollections().GetCollections()
			found := false
			for _, coll := range collections {
				if req := coll.FindRequest(msg.Node.ID); req != nil {
					m.requestPanel.LoadCollectionRequest(req)
					found = true
					break
				}
			}

			if !found {
				m.statusBar.Error(fmt.Errorf("request not found: %s", msg.Node.ID))
			}

			// Focus the Request Panel
			m.activePanel = RequestPanel

			// Update status bar with method and breadcrumb
			m.statusBar.SetMethod(msg.Node.HTTPMethod)
			breadcrumb := buildBreadcrumb(msg.Node)
			m.statusBar.SetBreadcrumb(breadcrumb...)
		}
		return m, m.markSessionDirty()

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

	case ResendRequestMsg:
		// Resend a request from console history
		if msg.Request != nil {
			m.isSending = true
			m.lastRequest = msg.Request
			m.requestStart = time.Now()
			m.responsePanel.ClearResponse()
			m.responsePanel.SetLoading(true)
			m.statusBar.Info("Resending request...")
			return m, tea.Batch(SendHTTPRequestCmd(msg.Request), loaderTickCmd())
		}
		return m, nil

	case CopyToClipboardMsg:
		// Copy content to clipboard
		if msg.Content != "" {
			clipboard.Write(clipboard.FmtText, []byte(msg.Content))
			// Note: clipboard.Write doesn't return an error in this library version
			m.statusBar.Success("Copied", msg.Label)
		} else {
			m.statusBar.Info("Nothing to copy")
		}
		return m, nil

	case ConsoleStatusMsg:
		// Display status message from console
		switch msg.Type {
		case StatusSuccess:
			m.statusBar.Success("", msg.Message)
		case StatusError:
			m.statusBar.Error(fmt.Errorf("%s", msg.Message))
		default:
			m.statusBar.Info(msg.Message)
		}
		return m, nil

	case SwitchToConsoleTabMsg:
		// Switch response panel to Console tab
		m.responsePanel.tabs.SetActive(3) // Console is tab index 3
		return m, nil

	case SwitchToResponseTabMsg:
		// Switch response panel to Body tab
		m.responsePanel.tabs.SetActive(0) // Body is tab index 0
		return m, nil

	case ShowImportModalMsg:
		// Show the import modal
		m.importModal.SetSize(m.width, m.height)
		m.importModal.Show()
		return m, nil

	case HideImportModalMsg:
		// Hide the import modal
		m.importModal.Hide()
		return m, nil

	case CurlImportedMsg:
		// Handle successful cURL import
		if msg.Request != nil {
			m.requestPanel.LoadCollectionRequest(msg.Request)
			m.statusBar.Success("Imported", msg.Request.Name)
			// Focus the request panel
			m.activePanel = RequestPanel
			// Update status bar with method
			m.statusBar.SetMethod(string(msg.Request.Method))
		}
		return m, nil

	case CurlExportedMsg:
		// Handle cURL export result
		if msg.Error != nil {
			m.statusBar.Error(msg.Error)
		} else if msg.Success {
			m.statusBar.Success("Copied", "cURL command to clipboard")
		}
		return m, nil

	case PostmanImportedMsg:
		// Handle successful Postman import
		if msg.IsEnv {
			// Save imported environment to workspace
			if msg.Environment != nil {
				if err := SaveImportedEnvironment(msg.Environment, m.workspacePath); err != nil {
					m.statusBar.Error(err)
				} else {
					m.statusBar.Success("Imported", msg.Summary)
					// Reload environments
					m.leftPanel.GetEnvironments().ReloadEnvironments()
				}
			}
		} else {
			// Save imported collection to workspace
			if msg.Collection != nil {
				if err := SaveImportedCollection(msg.Collection, m.workspacePath); err != nil {
					m.statusBar.Error(err)
				} else {
					m.statusBar.Success("Imported", msg.Summary)
					// Reload collections
					m.leftPanel.GetCollections().ReloadCollections()
				}
			}
		}
		return m, nil

	case PostmanExportedMsg:
		// Handle Postman export result
		if msg.Error != nil {
			m.statusBar.Error(msg.Error)
		} else if msg.Success {
			m.statusBar.Success("Exported", msg.FilePath)
		}
		return m, nil

	case PostmanImportErrorMsg:
		// Handle Postman import error
		m.statusBar.Error(msg.Error)
		return m, nil

	case HTTPSendingMsg:
		// HTTP request is being sent
		m.isSending = true
		m.statusBar.Info("Sending request...")
		m.responsePanel.ClearResponse()
		m.responsePanel.SetLoading(true)
		return m, loaderTickCmd()

	case LoaderTickMsg:
		// Animate the loader if still loading
		if m.responsePanel.IsLoading() {
			m.responsePanel.TickLoader()
			return m, loaderTickCmd()
		}
		return m, nil

	case HTTPResponseMsg:
		// HTTP response received
		m.isSending = false
		m.responsePanel.SetLoading(false)
		duration := time.Since(m.requestStart)

		// Log to console history
		if m.lastRequest != nil && m.consoleHistory != nil {
			entry := api.NewConsoleEntry(m.lastRequest, msg.Response, msg.Error, duration)
			m.consoleHistory.Add(*entry)
		}

		if msg.Error != nil {
			m.statusBar.Error(msg.Error)
			return m, nil
		}
		if msg.Response != nil {
			// Parse headers into simple map
			headers := make(map[string]string)
			for key, values := range msg.Response.Headers {
				if len(values) > 0 {
					headers[key] = strings.Join(values, ", ")
				}
			}

			// Parse cookies from Set-Cookie headers
			cookies := make(map[string]string)
			if cookieHeaders, ok := msg.Response.Headers["Set-Cookie"]; ok {
				for _, cookie := range cookieHeaders {
					// Parse "name=value; attributes" format
					parts := strings.SplitN(cookie, "=", 2)
					if len(parts) == 2 {
						name := parts[0]
						valueParts := strings.SplitN(parts[1], ";", 2)
						cookies[name] = valueParts[0]
					}
				}
			}

			// Format time and size
			timeStr := formatDuration(msg.Response.Time)
			sizeStr := formatBytes(msg.Response.Size)

			// Update response panel
			m.responsePanel.SetResponse(
				msg.Response.StatusCode,
				msg.Response.Status,
				headers,
				cookies,
				msg.Response.Body,
				timeStr,
				sizeStr,
			)

			// Update status bar with HTTP status
			statusText := ""
			switch {
			case msg.Response.StatusCode >= 200 && msg.Response.StatusCode < 300:
				statusText = "OK"
			case msg.Response.StatusCode >= 300 && msg.Response.StatusCode < 400:
				statusText = "Redirect"
			case msg.Response.StatusCode >= 400 && msg.Response.StatusCode < 500:
				statusText = "Client Error"
			case msg.Response.StatusCode >= 500:
				statusText = "Server Error"
			}
			m.statusBar.SetHTTPStatus(msg.Response.StatusCode, statusText)

			// Focus response panel
			m.activePanel = ResponsePanel
			m.statusBar.Success("Response", fmt.Sprintf("%d %s in %s", msg.Response.StatusCode, statusText, timeStr))
		}
		return m, nil

	case CommandExecuteMsg:
		// Handle command execution
		return m.handleCommand(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		// Update layout mode based on terminal size
		m.layoutMode = m.detectLayoutMode()
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
		// Pass console history to response panel for Console tab
		*m.responsePanel, cmd = m.responsePanel.UpdateWithHistory(msg, m.globalConfig, m.consoleHistory)
	}

	return m, cmd
}

// Minimum terminal size constants
const (
	MinTerminalWidth  = 80
	MinTerminalHeight = 24
)

// LayoutMode represents the panel arrangement mode
type LayoutMode int

const (
	// VerticalLayout is the default 3-panel Lazygit-style layout (side-by-side)
	VerticalLayout LayoutMode = iota
	// HorizontalLayout stacks panels vertically for small terminals
	HorizontalLayout
)

// Responsive breakpoints for layout switching
const (
	// ResponsiveWidthThreshold - below this width, switch to horizontal layout
	ResponsiveWidthThreshold = 100
	// ResponsiveHeightThreshold - below this height, switch to horizontal layout
	ResponsiveHeightThreshold = 30
)

// detectLayoutMode determines the layout mode based on terminal size
func (m Model) detectLayoutMode() LayoutMode {
	if m.width < ResponsiveWidthThreshold || m.height < ResponsiveHeightThreshold {
		return HorizontalLayout
	}
	return VerticalLayout
}

// renderVerticalLayout renders the default Lazygit-style 3-panel layout (side-by-side)
func (m Model) renderVerticalLayout() string {
	// Calculate panel dimensions
	// Reserve 1 line for status bar
	contentHeight := m.height - 1

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
	responseContent := m.responsePanel.ViewWithHistory(
		rightWidth-4,
		bottomRightHeight-2,
		m.activePanel == ResponsePanel,
		m.consoleHistory,
	)
	responsePanel := m.renderPanel("Response", responseContent, rightWidth, bottomRightHeight, m.activePanel == ResponsePanel)

	// Combine right panels vertically - no extra spacing
	rightSide := requestPanel + "\n" + responsePanel

	// Combine left and right horizontally
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanelRendered,
		rightSide,
	)
}

// renderHorizontalLayout renders panels stacked vertically for small terminals
func (m Model) renderHorizontalLayout() string {
	// Calculate panel dimensions
	// Reserve 1 line for status bar
	contentHeight := m.height - 1

	// Horizontal (stacked) layout:
	// +-------------------------------------+
	// |         Collections                 |
	// +-------------------------------------+
	// |           Request                   |
	// +-------------------------------------+
	// |           Response                  |
	// +-------------------------------------+
	// |         Status Bar                  |
	// +-------------------------------------+

	// Full width for all panels
	panelWidth := m.width

	// Distribute height equally among 3 panels
	// Each panel gets 1/3 of the available content height
	collectionsHeight := contentHeight / 3
	requestHeight := contentHeight / 3
	responseHeight := contentHeight - collectionsHeight - requestHeight

	// Collections panel (top)
	collectionsContent := m.leftPanel.View(
		panelWidth-4,
		collectionsHeight-2,
		m.activePanel == CollectionsPanel,
	)
	collectionsPanel := m.renderPanelWithTabs(m.leftPanel, collectionsContent, panelWidth, collectionsHeight, m.activePanel == CollectionsPanel)

	// Request panel (middle)
	requestContent := m.requestPanel.View(
		panelWidth-4,
		requestHeight-2,
		m.activePanel == RequestPanel,
	)
	requestPanel := m.renderPanel("Request", requestContent, panelWidth, requestHeight, m.activePanel == RequestPanel)

	// Response panel (bottom)
	responseContent := m.responsePanel.ViewWithHistory(
		panelWidth-4,
		responseHeight-2,
		m.activePanel == ResponsePanel,
		m.consoleHistory,
	)
	responsePanel := m.renderPanel("Response", responseContent, panelWidth, responseHeight, m.activePanel == ResponsePanel)

	// Stack panels vertically
	return collectionsPanel + "\n" + requestPanel + "\n" + responsePanel
}

// toggleFullscreen toggles fullscreen mode for the current active panel
func (m *Model) toggleFullscreen() {
	if m.isFullscreen {
		// Exit fullscreen
		m.isFullscreen = false
	} else {
		// Enter fullscreen with current panel
		m.isFullscreen = true
		m.fullscreenPanel = m.activePanel
	}
}

// renderFullscreenLayout renders the active panel in fullscreen mode
func (m Model) renderFullscreenLayout() string {
	// Reserve 1 line for status bar
	contentHeight := m.height - 1
	panelWidth := m.width

	var panelContent string
	var panelTitle string

	switch m.fullscreenPanel {
	case CollectionsPanel:
		content := m.leftPanel.View(
			panelWidth-4,
			contentHeight-2,
			true, // Always active in fullscreen
		)
		return m.renderPanelWithTabs(m.leftPanel, content, panelWidth, contentHeight, true)

	case RequestPanel:
		panelTitle = "Request"
		panelContent = m.requestPanel.View(
			panelWidth-4,
			contentHeight-2,
			true,
		)

	case ResponsePanel:
		panelTitle = "Response"
		panelContent = m.responsePanel.ViewWithHistory(
			panelWidth-4,
			contentHeight-2,
			true,
			m.consoleHistory,
		)

	default:
		// Fallback to collections
		content := m.leftPanel.View(
			panelWidth-4,
			contentHeight-2,
			true,
		)
		return m.renderPanelWithTabs(m.leftPanel, content, panelWidth, contentHeight, true)
	}

	return m.renderPanel(panelTitle, panelContent, panelWidth, contentHeight, true)
}

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

	// Render main content based on layout mode
	var mainContent string
	if m.isFullscreen {
		// Fullscreen mode - render only the active panel
		mainContent = m.renderFullscreenLayout()
	} else if m.layoutMode == HorizontalLayout {
		mainContent = m.renderHorizontalLayout()
	} else {
		mainContent = m.renderVerticalLayout()
	}

	// Status bar or command input
	var bottomBar string
	if m.commandInput.IsVisible() {
		bottomBar = m.commandInput.View(m.width)
	} else {
		bottomBar = m.renderStatusBar()
	}

	// Join without extra spacing
	result := mainContent + "\n" + bottomBar

	// Apply jump mode overlay if active
	if m.jumpMode.IsActive() {
		result = m.jumpMode.RenderOverlay(result, m.width, m.height)
	}

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

	// Overlay import modal if visible
	if m.importModal.IsVisible() {
		importView := m.importModal.View()
		result = m.overlayDialog(result, importView)
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
	envName := m.leftPanel.GetEnvironments().GetActiveEnvironmentName()
	m.statusBar.SetEnvironment(envName)

	// Update environment variables in request panel for preview mode
	envVars := m.leftPanel.GetEnvironments().GetActiveEnvironmentVariables()
	m.requestPanel.SetEnvironmentVariables(envVars)

	// Update fullscreen state
	m.statusBar.SetFullscreen(m.isFullscreen)

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
		// :q or :quit - exit application (save session first)
		return m.saveSessionAndQuit()

	case CmdWrite, CmdWriteLong:
		// :w or :write - save current request
		m.statusBar.Success("Saved", "request")
		return m, nil

	case CmdWriteQuit:
		// :wq - save and quit (save session first)
		return m.saveSessionAndQuit()

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

	case CmdImport:
		// :import - import files (postman)
		return m.handleImportCommand(msg.Args)

	case CmdExport:
		// :export - export files (postman)
		return m.handleExportCommand(msg.Args)

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

// handleImportCommand processes import subcommands
func (m Model) handleImportCommand(args []string) (tea.Model, tea.Cmd) {
	if len(args) == 0 {
		m.statusBar.Info("Usage: :import postman <file>")
		return m, nil
	}

	switch args[0] {
	case ImportPostman:
		// :import postman <file> - import Postman collection or environment
		if len(args) < 2 {
			m.statusBar.Info("Usage: :import postman <file>")
			return m, nil
		}
		filePath := args[1]
		m.statusBar.Info("Importing " + filePath + "...")
		return m, ImportPostmanFile(filePath)

	default:
		m.statusBar.Info("Unknown import type: " + args[0] + ". Use: :import postman <file>")
		return m, nil
	}
}

// handleExportCommand processes export subcommands
func (m Model) handleExportCommand(args []string) (tea.Model, tea.Cmd) {
	if len(args) == 0 {
		m.statusBar.Info("Usage: :export postman <file>")
		return m, nil
	}

	switch args[0] {
	case ExportPostman:
		// :export postman <file> - export current collection to Postman format
		if len(args) < 2 {
			m.statusBar.Info("Usage: :export postman <file>")
			return m, nil
		}
		outputPath := args[1]

		// Get the active collection
		collections := m.leftPanel.GetCollections().GetCollections()
		if len(collections) == 0 {
			m.statusBar.Info("No collection to export")
			return m, nil
		}

		// Export the first collection
		if len(collections) > 1 {
			m.statusBar.Info("Warning: Multiple collections found, exporting first one: " + collections[0].Name)
		} else {
			m.statusBar.Info("Exporting to " + outputPath + "...")
		}
		return m, ExportCollectionToPostman(collections[0], outputPath)

	default:
		m.statusBar.Info("Unknown export type: " + args[0] + ". Use: :export postman <file>")
		return m, nil
	}
}

// handleDialogResult processes dialog results
func (m Model) handleDialogResult(msg components.DialogResultMsg) (tea.Model, tea.Cmd) {
	if !msg.Confirmed {
		m.statusBar.Info("Canceled")
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

	// Update URL without clearing other request data
	m.requestPanel.SetURL(newURL)
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
			if m.responsePanel.GetActiveTab() == "Console" {
				m.whichKey.SetContext(components.ContextConsole)
			} else {
				m.whichKey.SetContext(components.ContextNormalResponse)
			}
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

// sendHTTPRequest builds and sends an HTTP request from the current request panel state
func (m Model) sendHTTPRequest() (tea.Model, tea.Cmd) {
	// Check if a request is loaded
	url := m.requestPanel.GetURL()
	if url == "" {
		m.statusBar.Info("No URL to send")
		return m, nil
	}

	// Check if already sending
	if m.isSending {
		m.statusBar.Info("Request already in progress...")
		return m, nil
	}

	// Build the HTTP request
	req := m.buildHTTPRequest()
	if req == nil {
		m.statusBar.Info("Could not build request")
		return m, nil
	}

	// Update state to sending
	m.isSending = true
	m.lastRequest = req         // Track request for console logging
	m.requestStart = time.Now() // Track start time for duration
	m.responsePanel.ClearResponse()
	m.responsePanel.SetLoading(true)
	m.statusBar.Info("Sending request...")

	// Send the request asynchronously with loader tick
	return m, tea.Batch(SendHTTPRequestCmd(req), loaderTickCmd())
}

// buildHTTPRequest constructs an API Request from the current RequestView state
func (m *Model) buildHTTPRequest() *api.Request {
	method := m.requestPanel.GetMethod()
	url := m.requestPanel.GetURL()

	// Replace environment variables in URL
	envVars := m.leftPanel.GetEnvironments().GetActiveEnvironmentVariables()
	url = replaceVariables(url, envVars)

	// Build headers map from headers table
	headers := make(map[string]string)
	headersTable := m.requestPanel.GetHeadersTable()
	if headersTable != nil {
		for _, row := range headersTable.Rows {
			if row.Enabled && row.Key != "" {
				value := replaceVariables(row.Value, envVars)
				headers[row.Key] = value
			}
		}
	}

	// Add auth headers
	authConfig := m.requestPanel.GetAuthConfig()
	if authConfig != nil {
		switch authConfig.Type {
		case "bearer":
			prefix := authConfig.Prefix
			if prefix == "" {
				prefix = "Bearer"
			}
			token := replaceVariables(authConfig.Token, envVars)
			headers["Authorization"] = prefix + " " + token
		case "basic":
			username := replaceVariables(authConfig.Username, envVars)
			password := replaceVariables(authConfig.Password, envVars)
			credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
			headers["Authorization"] = "Basic " + credentials
		case "api_key":
			keyName := replaceVariables(authConfig.APIKeyName, envVars)
			keyValue := replaceVariables(authConfig.APIKeyValue, envVars)
			if authConfig.APIKeyLocation == "header" || authConfig.APIKeyLocation == "" {
				headers[keyName] = keyValue
			} else if authConfig.APIKeyLocation == "query" {
				// Append to URL as query param
				if strings.Contains(url, "?") {
					url += "&" + keyName + "=" + keyValue
				} else {
					url += "?" + keyName + "=" + keyValue
				}
			}
		}
	}

	// Get body content
	var body interface{}
	bodyContent := m.requestPanel.GetBodyContent()
	if bodyContent != "" {
		bodyContent = replaceVariables(bodyContent, envVars)
		// Try to parse as JSON for proper serialization
		var jsonBody interface{}
		if err := json.Unmarshal([]byte(bodyContent), &jsonBody); err == nil {
			body = jsonBody
		} else {
			// Use raw string as body
			body = bodyContent
		}
	}

	return &api.Request{
		Method:  api.HTTPMethod(method),
		URL:     url,
		Headers: headers,
		Body:    body,
		Timeout: 30 * time.Second,
	}
}

// replaceVariables replaces {{variable}} patterns with environment values
func replaceVariables(input string, vars map[string]string) string {
	result := input
	for key, value := range vars {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dμs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// formatBytes formats bytes for display
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
	)
	switch {
	case bytes < KB:
		return fmt.Sprintf("%dB", bytes)
	case bytes < MB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%.1fMB", float64(bytes)/MB)
	}
}

// SessionSaveTickMsg is sent when the debounced save timer fires
type SessionSaveTickMsg struct {
	DirtyTime time.Time
}

// sessionSaveTick returns a command that fires after the debounce delay (500ms)
func sessionSaveTick(dirtyTime time.Time) tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return SessionSaveTickMsg{DirtyTime: dirtyTime}
	})
}

// markSessionDirty marks the session as dirty and returns a command to trigger debounced save
func (m *Model) markSessionDirty() tea.Cmd {
	now := time.Now()
	m.sessionDirtyTime = now
	return sessionSaveTick(now)
}

// saveSession saves the current session state to disk
func (m *Model) saveSession() {
	if m.session == nil {
		return
	}

	// Update session from current state
	switch m.activePanel {
	case CollectionsPanel:
		m.session.ActivePanel = "collections"
	case RequestPanel:
		m.session.ActivePanel = "request"
	case ResponsePanel:
		m.session.ActivePanel = "response"
	}

	// Save active request ID
	m.session.ActiveRequest = m.requestPanel.GetCurrentRequestID()

	// Save active environment
	m.session.ActiveEnvironment = m.leftPanel.GetEnvironments().GetActiveEnvironmentName()

	// Get panel states
	m.session.Panels.Collections = m.leftPanel.GetSessionState()
	m.session.Panels.Request = m.requestPanel.GetSessionState()
	m.session.Panels.Response = m.responsePanel.GetSessionState()

	// Note: LastUpdated is set by session.Save()

	// Save to disk (ignore errors silently)
	_ = m.session.Save(m.workspacePath)
}

// saveSessionAndQuit saves the session and returns the quit command
func (m *Model) saveSessionAndQuit() (Model, tea.Cmd) {
	m.saveSession()
	return *m, tea.Quit
}

// exportCurlCommand exports the current request as a cURL command to clipboard
func (m Model) exportCurlCommand() (tea.Model, tea.Cmd) {
	// Check if a request is loaded
	url := m.requestPanel.GetURL()
	if url == "" {
		m.statusBar.Info("No request to export")
		return m, nil
	}

	// Build a CollectionRequest from current request panel state
	req := m.buildCollectionRequest()
	if req == nil {
		m.statusBar.Info("Could not build request")
		return m, nil
	}

	// Generate cURL command
	curlCmd := api.GenerateCurlCommand(req)
	if curlCmd == "" {
		m.statusBar.Info("Could not generate cURL command")
		return m, nil
	}

	// Copy to clipboard
	clipboard.Write(clipboard.FmtText, []byte(curlCmd))

	return m, func() tea.Msg {
		return CurlExportedMsg{Success: true}
	}
}

// buildCollectionRequest builds a CollectionRequest from the current RequestView state
func (m *Model) buildCollectionRequest() *api.CollectionRequest {
	method := m.requestPanel.GetMethod()
	url := m.requestPanel.GetURL()

	if url == "" {
		return nil
	}

	// Build headers from headers table
	var headers []api.KeyValueEntry
	headersTable := m.requestPanel.GetHeadersTable()
	if headersTable != nil {
		for _, row := range headersTable.Rows {
			if row.Key != "" {
				headers = append(headers, api.KeyValueEntry{
					Key:     row.Key,
					Value:   row.Value,
					Enabled: row.Enabled,
				})
			}
		}
	}

	// Build body config
	var body *api.BodyConfig
	bodyContent := m.requestPanel.GetBodyContent()
	if bodyContent != "" {
		body = &api.BodyConfig{
			Type:    "raw",
			Content: bodyContent,
		}
	}

	// Build auth config
	var auth *api.AuthConfig
	authConfig := m.requestPanel.GetAuthConfig()
	if authConfig != nil && authConfig.Type != "" && authConfig.Type != "none" {
		auth = authConfig
	}

	return &api.CollectionRequest{
		ID:      m.requestPanel.GetCurrentRequestID(),
		Name:    "Exported Request",
		Method:  api.HTTPMethod(method),
		URL:     url,
		Headers: headers,
		Body:    body,
		Auth:    auth,
	}
}

// activateJumpMode activates jump mode and collects targets from panels
func (m *Model) activateJumpMode(allPanels bool) (Model, tea.Cmd) {
	if allPanels {
		m.jumpMode.ActivateAllPanels()
	} else {
		m.jumpMode.Activate(m.activePanel)
	}

	// Collect targets from relevant panels
	var targets []JumpTarget

	if allPanels {
		// Cross-panel mode: collect from all panels
		targets = append(targets, m.collectCollectionTargets()...)
		targets = append(targets, m.collectRequestTargets()...)
		targets = append(targets, m.collectResponseTargets()...)
	} else {
		// Single panel mode: collect from current panel only
		switch m.activePanel {
		case CollectionsPanel:
			targets = m.collectCollectionTargets()
		case RequestPanel:
			targets = m.collectRequestTargets()
		case ResponsePanel:
			targets = m.collectResponseTargets()
		}
	}

	// Handle case with no targets
	if len(targets) == 0 {
		m.jumpMode.Deactivate()
		m.statusBar.Info("No jump targets available")
		return *m, nil
	}

	// Set targets and assign labels
	m.jumpMode.SetTargets(targets)

	// Update mode display
	m.mode = JumpMode
	m.statusBar.SetMode(JumpMode)

	return *m, nil
}

// collectCollectionTargets collects jump targets from collections panel
func (m *Model) collectCollectionTargets() []JumpTarget {
	startRow, startCol := m.getCollectionsPanelPosition()
	return m.leftPanel.GetJumpTargets(startRow, startCol)
}

// collectRequestTargets collects jump targets from request panel
func (m *Model) collectRequestTargets() []JumpTarget {
	startRow, startCol := m.getRequestPanelPosition()
	return m.requestPanel.GetJumpTargets(startRow, startCol)
}

// collectResponseTargets collects jump targets from response panel
func (m *Model) collectResponseTargets() []JumpTarget {
	startRow, startCol := m.getResponsePanelPosition()
	return m.responsePanel.GetJumpTargets(startRow, startCol)
}

// getCollectionsPanelPosition returns the (row, col) offset for the collections panel
// based on the current layout mode (vertical, horizontal, or fullscreen).
func (m *Model) getCollectionsPanelPosition() (startRow, startCol int) {
	if m.isFullscreen {
		// Fullscreen: panel fills entire screen
		return 0, 0
	}

	if m.layoutMode == HorizontalLayout {
		// Horizontal (stacked): collections at top
		return 0, 0
	}

	// Vertical layout (default): collections on the left
	return 0, 0
}

// getRequestPanelPosition returns the (row, col) offset for the request panel
// based on the current layout mode (vertical, horizontal, or fullscreen).
func (m *Model) getRequestPanelPosition() (startRow, startCol int) {
	if m.isFullscreen {
		// Fullscreen: panel fills entire screen
		return 0, 0
	}

	contentHeight := m.height - 1 // Reserve status bar

	if m.layoutMode == HorizontalLayout {
		// Horizontal (stacked): request in middle, below collections
		collectionsHeight := contentHeight / 3
		return collectionsHeight, 0
	}

	// Vertical layout (default): request panel in top-right
	// Left panel takes 1/3 of width
	leftPanelWidth := m.width / 3
	return 0, leftPanelWidth
}

// getResponsePanelPosition returns the (row, col) offset for the response panel
// based on the current layout mode (vertical, horizontal, or fullscreen).
func (m *Model) getResponsePanelPosition() (startRow, startCol int) {
	if m.isFullscreen {
		// Fullscreen: panel fills entire screen
		return 0, 0
	}

	contentHeight := m.height - 1 // Reserve status bar

	if m.layoutMode == HorizontalLayout {
		// Horizontal (stacked): response at bottom
		collectionsHeight := contentHeight / 3
		requestHeight := contentHeight / 3
		return collectionsHeight + requestHeight, 0
	}

	// Vertical layout (default): response panel in bottom-right
	// Left panel takes 1/3 of width
	// Request panel takes 40% of right side height
	leftPanelWidth := m.width / 3
	requestPanelHeight := contentHeight * 40 / 100
	return requestPanelHeight, leftPanelWidth
}

// refreshJumpTargetsForScope refreshes jump targets for a specific panel scope.
// Used when Tab cycles through panels in single-panel jump mode.
func (m *Model) refreshJumpTargetsForScope(panel PanelType) {
	var targets []JumpTarget

	switch panel {
	case CollectionsPanel:
		targets = m.collectCollectionTargets()
	case RequestPanel:
		targets = m.collectRequestTargets()
	case ResponsePanel:
		targets = m.collectResponseTargets()
	}

	if len(targets) == 0 {
		m.statusBar.Info("No jump targets in this panel")
		return
	}

	m.jumpMode.SetTargets(targets)
}

// jumpToTarget executes a jump to the specified target
func (m *Model) jumpToTarget(target JumpTarget) (Model, tea.Cmd) {
	// Deactivate jump mode
	m.jumpMode.Deactivate()
	m.mode = NormalMode
	m.statusBar.SetMode(NormalMode)

	// Switch to target panel if different
	if target.Panel != m.activePanel {
		m.activePanel = target.Panel
	}

	// Execute appropriate action based on target
	switch target.Panel {
	case CollectionsPanel:
		// Select the item in collections tree
		if target.Index >= 0 {
			m.leftPanel.GetCollections().SelectIndex(target.Index)
		}
	case RequestPanel:
		// Handle request panel navigation
		// Will be implemented fully in T034
		if target.ElementID != "" {
			// Switch to tab or focus element
			m.requestPanel.JumpTo(target.ElementID)
		}
	case ResponsePanel:
		// Handle response panel navigation
		// Will be implemented fully in T035
		if target.ElementID != "" {
			m.responsePanel.JumpTo(target.ElementID)
		}
	}

	// Enter insert mode if target action is focus
	if target.Action == JumpFocus {
		m.mode = InsertMode
		m.statusBar.SetMode(InsertMode)
	}

	return *m, m.markSessionDirty()
}
