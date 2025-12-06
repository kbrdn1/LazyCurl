package ui

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/internal/api"
	"github.com/kbrdn1/LazyCurl/internal/config"
	"github.com/kbrdn1/LazyCurl/internal/ui/components"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// BodyType represents the type of request body
type BodyType int

const (
	NoneBody BodyType = iota
	JSONBody
	FormDataBody
	RawBody
	BinaryBody
)

// String returns the display name for the body type
func (b BodyType) String() string {
	switch b {
	case NoneBody:
		return "none"
	case JSONBody:
		return "JSON"
	case FormDataBody:
		return "form-data"
	case RawBody:
		return "raw"
	case BinaryBody:
		return "binary"
	default:
		return "none"
	}
}

// === REQUEST ACTION MESSAGES ===
// These are sent to the parent model to handle dialogs

// RequestRenameMsg is sent when rename is requested
type RequestRenameMsg struct {
	Tab   string // "Params" or "Headers"
	Index int
	Key   string
	Value string
}

// RequestDeleteMsg is sent when delete is requested
type RequestDeleteMsg struct {
	Tab   string
	Index int
	Key   string
}

// RequestEditMsg is sent when edit is requested
type RequestEditMsg struct {
	Tab   string
	Index int
	Key   string
	Value string
}

// RequestNewMsg is sent when new entry is requested
type RequestNewMsg struct {
	Tab string
}

// RequestDuplicateMsg is sent when duplicate is requested
type RequestDuplicateMsg struct {
	Tab   string
	Index int
}

// RequestYankMsg is sent when yank is requested
type RequestYankMsg struct {
	Tab   string
	Index int
	Key   string
	Value string
}

// RequestPasteMsg is sent when paste is requested
type RequestPasteMsg struct {
	Tab string
}

// RequestURLChangedMsg is sent when the URL is modified
type RequestURLChangedMsg struct {
	URL string
}

// RequestParamsChangedMsg is sent when params are modified (to sync with URL)
type RequestParamsChangedMsg struct {
	URL string
}

// RequestParamToggleMsg is sent when a param is toggled (enabled/disabled)
type RequestParamToggleMsg struct {
	Tab string
}

// RequestBodyChangedMsg is sent when body content is modified
type RequestBodyChangedMsg struct {
	BodyType string // "json", "raw", "none", etc.
	Content  string
}

// RequestScriptsChangedMsg is sent when scripts content is modified
type RequestScriptsChangedMsg struct {
	PreRequest  string
	PostRequest string
}

// RequestAuthChangedMsg is sent when auth configuration is modified
type RequestAuthChangedMsg struct {
	Auth *api.AuthConfig
}

// ParamsSection represents which section is active in Params tab
type ParamsSection int

const (
	PathParamsSection ParamsSection = iota
	QueryParamsSection
)

// ScriptsSection represents which section is active in Scripts tab
type ScriptsSection int

const (
	PreRequestSection ScriptsSection = iota
	PostRequestSection
)

// AuthType represents the type of authentication
type AuthType int

const (
	AuthNone AuthType = iota
	AuthBearer
	AuthBasic
	AuthAPIKey
)

// String returns the display name for the auth type
func (a AuthType) String() string {
	switch a {
	case AuthNone:
		return "No Auth"
	case AuthBearer:
		return "Bearer Token"
	case AuthBasic:
		return "Basic Auth"
	case AuthAPIKey:
		return "API Key"
	default:
		return "No Auth"
	}
}

// AuthField represents which field is selected in Authorization tab
type AuthField int

const (
	AuthFieldType AuthField = iota
	AuthFieldToken
	AuthFieldPrefix
	AuthFieldUsername
	AuthFieldPassword
	AuthFieldAPIKeyName
	AuthFieldAPIKeyValue
	AuthFieldAPIKeyLocation
)

// RequestView represents the request builder panel
type RequestView struct {
	method       api.HTTPMethod
	url          string
	tabs         *components.Tabs
	paramsTable  *components.Table // Query params
	pathParams   *components.Table // Path params (:id, :slug, etc.)
	headersTable *components.Table
	bodyEditor   *components.Editor
	bodyType     BodyType

	// Authorization tab
	authType           AuthType
	authToken          string
	authPrefix         string // Bearer prefix (default: "Bearer")
	authUsername       string
	authPassword       string
	authAPIKeyName     string
	authAPIKeyValue    string
	authAPIKeyLocation string // "header" or "query"
	authField          AuthField
	authEditing        bool // Whether we're editing a field

	// Scripts tab editors
	preRequestEditor  *components.Editor
	postRequestEditor *components.Editor
	scriptsSection    ScriptsSection

	// Params tab section (Query or Path)
	paramsSection ParamsSection

	// Current request tracking (for saving changes)
	currentRequestID   string
	currentRequestName string

	// URL editing state
	editingURL bool
	urlCursor  int

	// Clipboard for yank/paste
	clipboard *KeyValueClipboard
}

// KeyValueClipboard holds copied key-value data
type KeyValueClipboard struct {
	Key   string
	Value string
}

// NewRequestView creates a new request view
func NewRequestView() *RequestView {
	// Create tabs (shortcuts not displayed)
	tabs := components.NewTabs([]string{
		"Params",
		"Authorization",
		"Headers",
		"Body",
		"Scripts",
	})

	paramsTable := components.NewTable([]string{"", "Key", "Value"})
	pathParams := components.NewTable([]string{"", "Key", "Value"})
	headersTable := components.NewTable([]string{"", "Key", "Value"})

	// Initialize body editor with sample JSON
	bodyEditor := components.NewEditor(`{
  "name": "John Doe",
  "email": "john@example.com"
}`, "json")

	// Initialize script editors
	preRequestEditor := components.NewEditor(`// Pre-request script
// Runs before the request is sent

console.log('Request about to be sent');

// Access environment variables
const baseUrl = pm.environment.get('base_url');`, "javascript")

	postRequestEditor := components.NewEditor(`// Post-request script
// Runs after the response is received

console.log('Response received');

// Access response data
const response = pm.response.json();`, "javascript")

	rv := &RequestView{
		method:             api.GET,
		url:                "{{base_url}}/admin/users/:id",
		tabs:               tabs,
		paramsTable:        paramsTable,
		pathParams:         pathParams,
		headersTable:       headersTable,
		bodyEditor:         bodyEditor,
		bodyType:           JSONBody,
		authType:           AuthNone,
		authToken:          "",
		authPrefix:         "Bearer",
		authUsername:       "",
		authPassword:       "",
		authAPIKeyName:     "",
		authAPIKeyValue:    "",
		authAPIKeyLocation: "header",
		authField:          AuthFieldType,
		paramsSection:      QueryParamsSection,
		preRequestEditor:   preRequestEditor,
		postRequestEditor:  postRequestEditor,
		scriptsSection:     PreRequestSection,
	}

	// Add default headers like Postman
	rv.addDefaultHeaders()

	return rv
}

// addDefaultHeaders adds default HTTP headers like Postman
func (r *RequestView) addDefaultHeaders() {
	r.headersTable.AddRow("Content-Type", "application/json")
	r.headersTable.AddRow("Accept", "*/*")
	r.headersTable.AddRow("User-Agent", "LazyCurl/1.0")
	r.headersTable.AddRow("Accept-Encoding", "gzip, deflate, br")
	r.headersTable.AddRow("Connection", "keep-alive")
}

// getCurrentTable returns the table for the current tab/section
func (r *RequestView) getCurrentTable() *components.Table {
	switch r.tabs.GetActive() {
	case "Params":
		if r.paramsSection == PathParamsSection {
			return r.pathParams
		}
		return r.paramsTable
	case "Headers":
		return r.headersTable
	default:
		return nil
	}
}

// getTabName returns the tab name including section for Params tab
func (r *RequestView) getTabName() string {
	if r.tabs.GetActive() == "Params" {
		if r.paramsSection == PathParamsSection {
			return "PathParams"
		}
		return "Params"
	}
	return r.tabs.GetActive()
}

// GetActiveTab returns the active tab name
func (r *RequestView) GetActiveTab() string {
	return r.tabs.GetActive()
}

// GetClipboard returns the clipboard
func (r *RequestView) GetClipboard() *KeyValueClipboard {
	return r.clipboard
}

// SetClipboard sets the clipboard
func (r *RequestView) SetClipboard(key, value string) {
	r.clipboard = &KeyValueClipboard{Key: key, Value: value}
}

// AddRow adds a row to the current table
func (r *RequestView) AddRow(key, value string) {
	table := r.getCurrentTable()
	if table != nil {
		table.AddRow(key, value)
	}
}

// UpdateRow updates a row in the current table
func (r *RequestView) UpdateRow(index int, key, value string) {
	table := r.getCurrentTable()
	if table != nil && index >= 0 && index < len(table.Rows) {
		table.Rows[index].Key = key
		table.Rows[index].Value = value
	}
}

// RenameRow renames only the key of a row
func (r *RequestView) RenameRow(index int, newKey string) {
	table := r.getCurrentTable()
	if table != nil && index >= 0 && index < len(table.Rows) {
		table.Rows[index].Key = newKey
	}
}

// DeleteRow deletes a row from the current table
func (r *RequestView) DeleteRow(index int) {
	table := r.getCurrentTable()
	if table != nil {
		table.DeleteRow(index)
	}
}

// DuplicateRow duplicates a row
func (r *RequestView) DuplicateRow(index int) {
	table := r.getCurrentTable()
	if table != nil && index >= 0 && index < len(table.Rows) {
		row := table.Rows[index]
		newKey := row.Key + "_copy"
		table.AddRow(newKey, row.Value)
	}
}

// IsEditorActive returns true if an editor tab (Body or Scripts) is active
func (r *RequestView) IsEditorActive() bool {
	tab := r.tabs.GetActive()
	return tab == "Body" || tab == "Scripts"
}

// IsEditorInInsertMode returns true if the body editor is in INSERT mode
func (r *RequestView) IsEditorInInsertMode() bool {
	return r.bodyEditor.GetMode() == components.EditorInsertMode
}

// IsScriptsEditorInInsertMode returns true if the active scripts editor is in INSERT mode
func (r *RequestView) IsScriptsEditorInInsertMode() bool {
	if r.scriptsSection == PreRequestSection {
		return r.preRequestEditor.GetMode() == components.EditorInsertMode
	}
	return r.postRequestEditor.GetMode() == components.EditorInsertMode
}

// IsAuthEditing returns true if editing a field in Authorization tab
func (r *RequestView) IsAuthEditing() bool {
	return r.authEditing
}

// GetAuthConfig returns the current auth configuration
func (r *RequestView) GetAuthConfig() *api.AuthConfig {
	switch r.authType {
	case AuthNone:
		return nil
	case AuthBearer:
		prefix := r.authPrefix
		if prefix == "" {
			prefix = "Bearer"
		}
		return &api.AuthConfig{
			Type:   "bearer",
			Token:  r.authToken,
			Prefix: prefix,
		}
	case AuthBasic:
		return &api.AuthConfig{
			Type:     "basic",
			Username: r.authUsername,
			Password: r.authPassword,
		}
	case AuthAPIKey:
		location := r.authAPIKeyLocation
		if location == "" {
			location = "header"
		}
		return &api.AuthConfig{
			Type:           "api_key",
			APIKeyName:     r.authAPIKeyName,
			APIKeyValue:    r.authAPIKeyValue,
			APIKeyLocation: location,
		}
	}
	return nil
}

// getVisibleAuthFields returns the list of visible fields for current auth type
func (r *RequestView) getVisibleAuthFields() []AuthField {
	switch r.authType {
	case AuthNone:
		return []AuthField{AuthFieldType}
	case AuthBearer:
		return []AuthField{AuthFieldType, AuthFieldToken, AuthFieldPrefix}
	case AuthBasic:
		return []AuthField{AuthFieldType, AuthFieldUsername, AuthFieldPassword}
	case AuthAPIKey:
		return []AuthField{AuthFieldType, AuthFieldAPIKeyName, AuthFieldAPIKeyValue, AuthFieldAPIKeyLocation}
	}
	return []AuthField{AuthFieldType}
}

// getAuthFieldIndex returns the index of the current field in visible fields
func (r *RequestView) getAuthFieldIndex() int {
	fields := r.getVisibleAuthFields()
	for i, f := range fields {
		if f == r.authField {
			return i
		}
	}
	return 0
}

// GetActiveScriptsEditor returns the currently active scripts editor
func (r *RequestView) GetActiveScriptsEditor() *components.Editor {
	if r.scriptsSection == PreRequestSection {
		return r.preRequestEditor
	}
	return r.postRequestEditor
}

// Update handles messages for the request view
func (r RequestView) Update(msg tea.Msg, cfg *config.GlobalConfig) (RequestView, tea.Cmd) {
	switch msg := msg.(type) {
	case components.SearchUpdateMsg, components.SearchCloseMsg:
		// Forward search messages to the active editor
		if r.tabs.GetActive() == "Body" && r.bodyType == JSONBody {
			editor, cmd := r.bodyEditor.Update(msg, true)
			r.bodyEditor = editor
			return r, cmd
		}
		if r.tabs.GetActive() == "Scripts" {
			activeEditor := r.GetActiveScriptsEditor()
			editor, cmd := activeEditor.Update(msg, true)
			if r.scriptsSection == PreRequestSection {
				r.preRequestEditor = editor
			} else {
				r.postRequestEditor = editor
			}
			return r, cmd
		}
		return r, nil

	case components.EditorFormatMsg:
		// Handle format result from editor - also emit body changed
		if msg.Success && r.tabs.GetActive() == "Body" {
			bodyType := r.bodyType.String()
			content := r.bodyEditor.GetContent()
			return r, func() tea.Msg {
				return RequestBodyChangedMsg{BodyType: bodyType, Content: content}
			}
		}
		return r, nil

	case components.EditorContentChangedMsg:
		// Handle content changes from body editor
		if r.tabs.GetActive() == "Body" && r.bodyType == JSONBody {
			bodyType := r.bodyType.String()
			return r, func() tea.Msg {
				return RequestBodyChangedMsg{BodyType: bodyType, Content: msg.Content}
			}
		}
		// Handle scripts content changes
		if r.tabs.GetActive() == "Scripts" {
			return r, func() tea.Msg {
				return RequestScriptsChangedMsg{
					PreRequest:  r.preRequestEditor.GetContent(),
					PostRequest: r.postRequestEditor.GetContent(),
				}
			}
		}
		return r, nil

	case tea.KeyMsg:
		// If editing URL, handle URL input
		if r.editingURL {
			return r.handleURLInput(msg)
		}

		// If in Body tab with JSON body type, forward to editor
		if r.tabs.GetActive() == "Body" && r.bodyType == JSONBody {
			// Only intercept tab switching and send request when in NORMAL mode and not searching
			if r.bodyEditor.GetMode() == components.EditorInsertMode || r.bodyEditor.IsSearching() {
				// In INSERT mode or searching, forward everything to editor
				editor, cmd := r.bodyEditor.Update(msg, true)
				r.bodyEditor = editor
				return r, cmd
			}

			// In NORMAL mode (not searching), check for tab switching first
			switch msg.String() {
			case "tab":
				r.tabs.Next()
				return r, nil
			case "shift+tab":
				r.tabs.Previous()
				return r, nil
			case "1", "2", "3", "4", "5":
				// Allow number-based tab switching
				switch msg.String() {
				case "1":
					r.tabs.SetActive(0)
				case "2":
					r.tabs.SetActive(1)
				case "3":
					r.tabs.SetActive(2)
				case "4":
					r.tabs.SetActive(3)
				case "5":
					r.tabs.SetActive(4)
				}
				return r, nil
			case "ctrl+s":
				// TODO: Send HTTP request
				return r, nil
			default:
				// Forward to editor for NORMAL mode commands
				editor, cmd := r.bodyEditor.Update(msg, true)
				r.bodyEditor = editor
				return r, cmd
			}
		}

		// If in Scripts tab, forward to the active script editor
		if r.tabs.GetActive() == "Scripts" {
			activeEditor := r.GetActiveScriptsEditor()

			// In INSERT mode or searching, forward everything to editor
			if activeEditor.GetMode() == components.EditorInsertMode || activeEditor.IsSearching() {
				editor, cmd := activeEditor.Update(msg, true)
				if r.scriptsSection == PreRequestSection {
					r.preRequestEditor = editor
				} else {
					r.postRequestEditor = editor
				}
				return r, cmd
			}

			// In NORMAL mode (not searching), check for special keys first
			switch msg.String() {
			case "tab":
				r.tabs.Next()
				return r, nil
			case "shift+tab":
				r.tabs.Previous()
				return r, nil
			case "1", "2", "3", "4", "5":
				// Allow number-based tab switching
				switch msg.String() {
				case "1":
					r.tabs.SetActive(0)
				case "2":
					r.tabs.SetActive(1)
				case "3":
					r.tabs.SetActive(2)
				case "4":
					r.tabs.SetActive(3)
				case "5":
					r.tabs.SetActive(4)
				}
				return r, nil
			case "h":
				// Switch to Pre-request section (left)
				if r.scriptsSection != PreRequestSection {
					r.scriptsSection = PreRequestSection
					return r, nil
				}
				// Forward h to editor for cursor movement
				editor, cmd := activeEditor.Update(msg, true)
				if r.scriptsSection == PreRequestSection {
					r.preRequestEditor = editor
				} else {
					r.postRequestEditor = editor
				}
				return r, cmd
			case "l":
				// Switch to Post-request section (right)
				if r.scriptsSection != PostRequestSection {
					r.scriptsSection = PostRequestSection
					return r, nil
				}
				// Forward l to editor for cursor movement
				editor, cmd := activeEditor.Update(msg, true)
				if r.scriptsSection == PreRequestSection {
					r.preRequestEditor = editor
				} else {
					r.postRequestEditor = editor
				}
				return r, cmd
			case "ctrl+s":
				// TODO: Send HTTP request
				return r, nil
			default:
				// Forward to editor for NORMAL mode commands
				editor, cmd := activeEditor.Update(msg, true)
				if r.scriptsSection == PreRequestSection {
					r.preRequestEditor = editor
				} else {
					r.postRequestEditor = editor
				}
				return r, cmd
			}
		}

		// If in Authorization tab, handle auth-specific keys
		if r.tabs.GetActive() == "Authorization" {
			return r.handleAuthInput(msg)
		}

		// Handle send request
		if msg.String() == "ctrl+s" {
			// TODO: Send HTTP request
			return r, nil
		}

		// "I" to edit URL input (uppercase I)
		if msg.String() == "I" {
			r.editingURL = true
			r.urlCursor = len(r.url)
			return r, nil
		}

		// Tab navigation with numbers 1-5 (NORMAL mode)
		switch msg.String() {
		case "tab":
			r.tabs.Next()
		case "shift+tab":
			r.tabs.Previous()
		case "1":
			r.tabs.SetActive(0) // Params
		case "2":
			r.tabs.SetActive(1) // Authorization
		case "3":
			r.tabs.SetActive(2) // Headers
		case "4":
			r.tabs.SetActive(3) // Body
		case "5":
			r.tabs.SetActive(4) // Scripts
		}

		// Handle Params tab section switching with h/l when in Params tab
		if r.tabs.GetActive() == "Params" {
			switch msg.String() {
			case "h":
				// Switch to Path Params section (left)
				if r.paramsSection != PathParamsSection {
					r.paramsSection = PathParamsSection
					return r, nil
				}
			case "l":
				// Switch to Query Params section (right)
				if r.paramsSection != QueryParamsSection {
					r.paramsSection = QueryParamsSection
					return r, nil
				}
			case "N":
				// New path param - switch to path params section and request new entry
				r.paramsSection = PathParamsSection
				return r, func() tea.Msg {
					return RequestNewMsg{
						Tab: "PathParams",
					}
				}
			}
		}

		// Navigation and actions for table tabs (like Collections)
		table := r.getCurrentTable()
		if table != nil {
			switch msg.String() {
			// Navigation
			case "j", "down":
				table.MoveDown()
			case "k", "up":
				table.MoveUp()
			case "g":
				table.Cursor = 0
			case "G":
				if table.RowCount() > 0 {
					table.Cursor = table.RowCount() - 1
				}

			// Actions - send messages to parent model
			case "c", "i":
				// Edit current row
				if table.Cursor >= 0 && table.Cursor < table.RowCount() {
					row := table.Rows[table.Cursor]
					return r, func() tea.Msg {
						return RequestEditMsg{
							Tab:   r.getTabName(),
							Index: table.Cursor,
							Key:   row.Key,
							Value: row.Value,
						}
					}
				}

			case "R":
				// Rename key
				if table.Cursor >= 0 && table.Cursor < table.RowCount() {
					row := table.Rows[table.Cursor]
					return r, func() tea.Msg {
						return RequestRenameMsg{
							Tab:   r.getTabName(),
							Index: table.Cursor,
							Key:   row.Key,
							Value: row.Value,
						}
					}
				}

			case "d":
				// Delete current row
				if table.Cursor >= 0 && table.Cursor < table.RowCount() {
					row := table.Rows[table.Cursor]
					return r, func() tea.Msg {
						return RequestDeleteMsg{
							Tab:   r.getTabName(),
							Index: table.Cursor,
							Key:   row.Key,
						}
					}
				}

			case "D":
				// Duplicate current row
				if table.Cursor >= 0 && table.Cursor < table.RowCount() {
					return r, func() tea.Msg {
						return RequestDuplicateMsg{
							Tab:   r.getTabName(),
							Index: table.Cursor,
						}
					}
				}

			case "y":
				// Yank (copy) current row
				if table.Cursor >= 0 && table.Cursor < table.RowCount() {
					row := table.Rows[table.Cursor]
					return r, func() tea.Msg {
						return RequestYankMsg{
							Tab:   r.getTabName(),
							Index: table.Cursor,
							Key:   row.Key,
							Value: row.Value,
						}
					}
				}

			case "p":
				// Paste
				return r, func() tea.Msg {
					return RequestPasteMsg{
						Tab: r.getTabName(),
					}
				}

			case "n":
				// New query param (in Params tab) or new header
				return r, func() tea.Msg {
					return RequestNewMsg{
						Tab: r.getTabName(),
					}
				}

			case "s", "S":
				// Toggle enabled state of current row
				if table.Cursor >= 0 && table.Cursor < table.RowCount() {
					table.ToggleCurrentEnabled()
					// Send message to sync params if in Params tab
					if r.tabs.GetActive() == "Params" {
						return r, func() tea.Msg {
							return RequestParamToggleMsg{Tab: r.getTabName()}
						}
					}
				}
			}
		}
	}

	return r, nil
}

// handleURLInput handles keyboard input when editing the URL
func (r RequestView) handleURLInput(msg tea.KeyMsg) (RequestView, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyEnter:
		// Exit URL editing mode
		r.editingURL = false
		// Parse URL params after editing
		r.ParseURLParams()
		// Send message to update collection with new URL
		newURL := r.url
		return r, func() tea.Msg {
			return RequestURLChangedMsg{URL: newURL}
		}

	case tea.KeyBackspace:
		if r.urlCursor > 0 && len(r.url) > 0 {
			// Delete character before cursor
			r.url = r.url[:r.urlCursor-1] + r.url[r.urlCursor:]
			r.urlCursor--
		}
		return r, nil

	case tea.KeyDelete:
		if r.urlCursor < len(r.url) {
			// Delete character at cursor
			r.url = r.url[:r.urlCursor] + r.url[r.urlCursor+1:]
		}
		return r, nil

	case tea.KeyLeft:
		if r.urlCursor > 0 {
			r.urlCursor--
		}
		return r, nil

	case tea.KeyRight:
		if r.urlCursor < len(r.url) {
			r.urlCursor++
		}
		return r, nil

	case tea.KeyHome, tea.KeyCtrlA:
		r.urlCursor = 0
		return r, nil

	case tea.KeyEnd, tea.KeyCtrlE:
		r.urlCursor = len(r.url)
		return r, nil

	case tea.KeyRunes:
		// Insert character at cursor
		char := string(msg.Runes)
		r.url = r.url[:r.urlCursor] + char + r.url[r.urlCursor:]
		r.urlCursor += len(char)
		return r, nil

	case tea.KeySpace:
		// Insert space
		r.url = r.url[:r.urlCursor] + " " + r.url[r.urlCursor:]
		r.urlCursor++
		return r, nil
	}

	return r, nil
}

// handleAuthInput handles keyboard input in Authorization tab
func (r RequestView) handleAuthInput(msg tea.KeyMsg) (RequestView, tea.Cmd) {
	// If editing a field, handle text input
	if r.authEditing {
		return r.handleAuthFieldEdit(msg)
	}

	// Navigation mode
	switch msg.String() {
	case "tab":
		r.tabs.Next()
		return r, nil
	case "shift+tab":
		r.tabs.Previous()
		return r, nil
	case "1", "2", "3", "4", "5":
		// Allow number-based tab switching
		switch msg.String() {
		case "1":
			r.tabs.SetActive(0)
		case "2":
			r.tabs.SetActive(1)
		case "3":
			r.tabs.SetActive(2)
		case "4":
			r.tabs.SetActive(3)
		case "5":
			r.tabs.SetActive(4)
		}
		return r, nil
	case "j", "down":
		// Move to next field
		fields := r.getVisibleAuthFields()
		idx := r.getAuthFieldIndex()
		if idx < len(fields)-1 {
			r.authField = fields[idx+1]
		}
		return r, nil
	case "k", "up":
		// Move to previous field
		fields := r.getVisibleAuthFields()
		idx := r.getAuthFieldIndex()
		if idx > 0 {
			r.authField = fields[idx-1]
		}
		return r, nil
	case "h", "left":
		// For type field, cycle auth types backward
		if r.authField == AuthFieldType {
			if r.authType > AuthNone {
				r.authType--
			} else {
				r.authType = AuthAPIKey
			}
			// Reset field to type when changing auth type
			r.authField = AuthFieldType
			// Emit auth change
			return r, r.emitAuthChanged()
		}
		// For API key location, toggle
		if r.authField == AuthFieldAPIKeyLocation {
			if r.authAPIKeyLocation == "header" {
				r.authAPIKeyLocation = "query"
			} else {
				r.authAPIKeyLocation = "header"
			}
			return r, r.emitAuthChanged()
		}
		return r, nil
	case "l", "right":
		// For type field, cycle auth types forward
		if r.authField == AuthFieldType {
			if r.authType < AuthAPIKey {
				r.authType++
			} else {
				r.authType = AuthNone
			}
			// Reset field to type when changing auth type
			r.authField = AuthFieldType
			// Emit auth change
			return r, r.emitAuthChanged()
		}
		// For API key location, toggle
		if r.authField == AuthFieldAPIKeyLocation {
			if r.authAPIKeyLocation == "header" {
				r.authAPIKeyLocation = "query"
			} else {
				r.authAPIKeyLocation = "header"
			}
			return r, r.emitAuthChanged()
		}
		return r, nil
	case "enter", "i", "c":
		// Enter edit mode for editable fields (not type or location)
		if r.authField != AuthFieldType && r.authField != AuthFieldAPIKeyLocation {
			r.authEditing = true
		}
		return r, nil
	case "ctrl+s":
		// TODO: Send HTTP request
		return r, nil
	}

	return r, nil
}

// handleAuthFieldEdit handles text input when editing an auth field
func (r RequestView) handleAuthFieldEdit(msg tea.KeyMsg) (RequestView, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyEnter:
		// Exit editing mode
		r.authEditing = false
		// Emit auth change
		return r, r.emitAuthChanged()

	case tea.KeyBackspace:
		// Delete last character from current field
		switch r.authField {
		case AuthFieldToken:
			if len(r.authToken) > 0 {
				r.authToken = r.authToken[:len(r.authToken)-1]
			}
		case AuthFieldPrefix:
			if len(r.authPrefix) > 0 {
				r.authPrefix = r.authPrefix[:len(r.authPrefix)-1]
			}
		case AuthFieldUsername:
			if len(r.authUsername) > 0 {
				r.authUsername = r.authUsername[:len(r.authUsername)-1]
			}
		case AuthFieldPassword:
			if len(r.authPassword) > 0 {
				r.authPassword = r.authPassword[:len(r.authPassword)-1]
			}
		case AuthFieldAPIKeyName:
			if len(r.authAPIKeyName) > 0 {
				r.authAPIKeyName = r.authAPIKeyName[:len(r.authAPIKeyName)-1]
			}
		case AuthFieldAPIKeyValue:
			if len(r.authAPIKeyValue) > 0 {
				r.authAPIKeyValue = r.authAPIKeyValue[:len(r.authAPIKeyValue)-1]
			}
		}
		return r, nil

	case tea.KeyRunes:
		// Append character to current field
		char := string(msg.Runes)
		switch r.authField {
		case AuthFieldToken:
			r.authToken += char
		case AuthFieldPrefix:
			r.authPrefix += char
		case AuthFieldUsername:
			r.authUsername += char
		case AuthFieldPassword:
			r.authPassword += char
		case AuthFieldAPIKeyName:
			r.authAPIKeyName += char
		case AuthFieldAPIKeyValue:
			r.authAPIKeyValue += char
		}
		return r, nil

	case tea.KeySpace:
		// Append space to current field
		switch r.authField {
		case AuthFieldToken:
			r.authToken += " "
		case AuthFieldPrefix:
			r.authPrefix += " "
		case AuthFieldUsername:
			r.authUsername += " "
		case AuthFieldPassword:
			r.authPassword += " "
		case AuthFieldAPIKeyName:
			r.authAPIKeyName += " "
		case AuthFieldAPIKeyValue:
			r.authAPIKeyValue += " "
		}
		return r, nil
	}

	return r, nil
}

// emitAuthChanged returns a command to emit auth changed message
func (r *RequestView) emitAuthChanged() tea.Cmd {
	auth := r.GetAuthConfig()
	return func() tea.Msg {
		return RequestAuthChangedMsg{Auth: auth}
	}
}

// ParseURLParams extracts query parameters from the URL and adds them to the params table
func (r *RequestView) ParseURLParams() {
	// Parse path parameters first
	r.ParsePathParams()

	// Find query string in URL
	urlParts := strings.SplitN(r.url, "?", 2)
	if len(urlParts) < 2 {
		return // No query string
	}

	queryString := urlParts[1]
	// Don't parse if empty
	if queryString == "" {
		return
	}

	// Parse query parameters
	pairs := strings.Split(queryString, "&")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			key := kv[0]
			value := kv[1]
			// Check if param already exists
			exists := false
			for _, row := range r.paramsTable.Rows {
				if row.Key == key {
					exists = true
					break
				}
			}
			if !exists {
				r.paramsTable.AddRow(key, value)
			}
		} else if len(kv) == 1 && kv[0] != "" {
			// Parameter without value
			exists := false
			for _, row := range r.paramsTable.Rows {
				if row.Key == kv[0] {
					exists = true
					break
				}
			}
			if !exists {
				r.paramsTable.AddRow(kv[0], "")
			}
		}
	}

	// Update cursor if needed
	if r.paramsTable.Cursor < 0 && r.paramsTable.RowCount() > 0 {
		r.paramsTable.Cursor = 0
	}
}

// ParsePathParams extracts path parameters (:param) from the URL and adds them to the pathParams table
func (r *RequestView) ParsePathParams() {
	pathParamNames := r.ExtractPathParamsFromURL()

	for _, paramName := range pathParamNames {
		// Check if param already exists
		exists := false
		for _, row := range r.pathParams.Rows {
			if row.Key == paramName {
				exists = true
				break
			}
		}
		if !exists {
			// Add with empty value - user will fill in the actual value
			r.pathParams.AddRow(paramName, "")
		}
	}

	// Remove params that no longer exist in URL
	var rowsToKeep []components.KeyValuePair
	for _, row := range r.pathParams.Rows {
		found := false
		for _, paramName := range pathParamNames {
			if row.Key == paramName {
				found = true
				break
			}
		}
		if found {
			rowsToKeep = append(rowsToKeep, row)
		}
	}
	r.pathParams.Rows = rowsToKeep

	// Update cursor if needed
	if r.pathParams.Cursor >= r.pathParams.RowCount() {
		r.pathParams.Cursor = r.pathParams.RowCount() - 1
	}
	if r.pathParams.Cursor < 0 && r.pathParams.RowCount() > 0 {
		r.pathParams.Cursor = 0
	}
}

// ExtractVariablesFromURL finds all {{variable}} patterns in the URL
func (r *RequestView) ExtractVariablesFromURL() []string {
	variablePattern := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := variablePattern.FindAllStringSubmatch(r.url, -1)

	var variables []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			variables = append(variables, match[1])
			seen[match[1]] = true
		}
	}
	return variables
}

// ExtractPathParamsFromURL finds all :param patterns in the URL
func (r *RequestView) ExtractPathParamsFromURL() []string {
	paramPattern := regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)`)
	matches := paramPattern.FindAllStringSubmatch(r.url, -1)

	var params []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			params = append(params, match[1])
			seen[match[1]] = true
		}
	}
	return params
}

// BuildURLFromParams reconstructs the URL with current query params
func (r *RequestView) BuildURLFromParams() string {
	// Get base URL (without query string)
	baseURL := r.url
	if idx := strings.Index(r.url, "?"); idx != -1 {
		baseURL = r.url[:idx]
	}

	// Build query string from enabled params
	var params []string
	for _, row := range r.paramsTable.Rows {
		if row.Enabled {
			if row.Value != "" {
				params = append(params, row.Key+"="+row.Value)
			} else {
				params = append(params, row.Key)
			}
		}
	}

	if len(params) == 0 {
		return baseURL
	}
	return baseURL + "?" + strings.Join(params, "&")
}

// SyncURLFromParams updates the internal URL from the params table
// and returns the new URL
func (r *RequestView) SyncURLFromParams() string {
	r.url = r.BuildURLFromParams()
	return r.url
}

// BuildURLWithPathParams returns the URL with path params substituted with their values
// This is used when sending requests
func (r *RequestView) BuildURLWithPathParams() string {
	url := r.url

	// Replace each :param with its value from pathParams table
	for _, row := range r.pathParams.Rows {
		if row.Enabled && row.Value != "" {
			url = strings.Replace(url, ":"+row.Key, row.Value, 1)
		}
	}

	return url
}

// AddPathParamToURL adds a new path param placeholder to the URL
func (r *RequestView) AddPathParamToURL(paramName string) {
	// Find the query string position if any
	queryIndex := strings.Index(r.url, "?")

	if queryIndex == -1 {
		// No query string, append at end
		r.url = r.url + "/:" + paramName
	} else {
		// Insert before query string
		r.url = r.url[:queryIndex] + "/:" + paramName + r.url[queryIndex:]
	}

	// Parse path params to update the table
	r.ParsePathParams()
}

// GetPathParamsTable returns the path params table for external access
func (r *RequestView) GetPathParamsTable() *components.Table {
	return r.pathParams
}

// GetParamsSection returns the current params section
func (r *RequestView) GetParamsSection() ParamsSection {
	return r.paramsSection
}

// renderURLWithCursor renders URL in editing mode with cursor
func (r *RequestView) renderURLWithCursor() string {
	// Style for cursor
	cursorStyle := lipgloss.NewStyle().
		Background(styles.Text).
		Foreground(styles.Base)

	// Style for editing background
	editStyle := lipgloss.NewStyle().
		Foreground(styles.Text).
		Background(styles.Surface0).
		Padding(0, 1)

	// Build URL with cursor
	var result strings.Builder

	// Ensure cursor is within bounds
	cursor := r.urlCursor
	if cursor > len(r.url) {
		cursor = len(r.url)
	}
	if cursor < 0 {
		cursor = 0
	}

	// Text before cursor
	if cursor > 0 {
		result.WriteString(r.url[:cursor])
	}

	// Cursor character
	if cursor < len(r.url) {
		result.WriteString(cursorStyle.Render(string(r.url[cursor])))
		// Text after cursor
		if cursor+1 < len(r.url) {
			result.WriteString(r.url[cursor+1:])
		}
	} else {
		// Cursor at end - show block cursor
		result.WriteString(cursorStyle.Render(" "))
	}

	return editStyle.Render(result.String())
}

// renderURLWithHighlight renders URL with syntax highlighting for variables and params
func (r *RequestView) renderURLWithHighlight(url string) string {
	// Patterns for highlighting
	variablePattern := regexp.MustCompile(`\{\{[^}]+\}\}`)
	paramPattern := regexp.MustCompile(`:([a-zA-Z_][a-zA-Z0-9_]*)`)

	// Style definitions
	variableStyle := lipgloss.NewStyle().Foreground(styles.URLVariable)
	paramStyle := lipgloss.NewStyle().Foreground(styles.URLParam)
	baseStyle := lipgloss.NewStyle().Foreground(styles.URLBase)

	// Find all matches and their positions
	type match struct {
		start int
		end   int
		text  string
		style lipgloss.Style
	}

	var matches []match

	// Find variables {{...}}
	for _, loc := range variablePattern.FindAllStringIndex(url, -1) {
		matches = append(matches, match{
			start: loc[0],
			end:   loc[1],
			text:  url[loc[0]:loc[1]],
			style: variableStyle,
		})
	}

	// Find params :name
	for _, loc := range paramPattern.FindAllStringIndex(url, -1) {
		matches = append(matches, match{
			start: loc[0],
			end:   loc[1],
			text:  url[loc[0]:loc[1]],
			style: paramStyle,
		})
	}

	// Sort matches by start position
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].start < matches[i].start {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Build highlighted string
	var result strings.Builder
	lastEnd := 0

	for _, m := range matches {
		// Add unstyled text before this match
		if m.start > lastEnd {
			result.WriteString(baseStyle.Render(url[lastEnd:m.start]))
		}
		// Add styled match
		result.WriteString(m.style.Render(m.text))
		lastEnd = m.end
	}

	// Add remaining text
	if lastEnd < len(url) {
		result.WriteString(baseStyle.Render(url[lastEnd:]))
	}

	return result.String()
}

// renderTextWithVariables renders text with variable highlighting ({{var}} in special color)
// If maskNonVariables is true, non-variable text will be masked with bullets
func renderTextWithVariables(text string, baseStyle, variableStyle lipgloss.Style, maskNonVariables bool) string {
	if text == "" {
		return ""
	}

	variablePattern := regexp.MustCompile(`\{\{[^}]+\}\}`)

	// Find all variable matches
	matches := variablePattern.FindAllStringIndex(text, -1)

	if len(matches) == 0 {
		// No variables, just render the text (masked or not)
		if maskNonVariables {
			return baseStyle.Render(strings.Repeat("•", min(len(text), 20)))
		}
		return baseStyle.Render(text)
	}

	// Build highlighted string
	var result strings.Builder
	lastEnd := 0

	for _, loc := range matches {
		// Add text before this match
		if loc[0] > lastEnd {
			segment := text[lastEnd:loc[0]]
			if maskNonVariables {
				result.WriteString(baseStyle.Render(strings.Repeat("•", min(len(segment), 10))))
			} else {
				result.WriteString(baseStyle.Render(segment))
			}
		}
		// Add variable with special style
		result.WriteString(variableStyle.Render(text[loc[0]:loc[1]]))
		lastEnd = loc[1]
	}

	// Add remaining text
	if lastEnd < len(text) {
		segment := text[lastEnd:]
		if maskNonVariables {
			result.WriteString(baseStyle.Render(strings.Repeat("•", min(len(segment), 10))))
		} else {
			result.WriteString(baseStyle.Render(segment))
		}
	}

	return result.String()
}

// getMethodStyle returns the style for the HTTP method badge
func (r *RequestView) getMethodStyle() (lipgloss.Color, lipgloss.Color) {
	switch r.method {
	case api.GET:
		return styles.MethodGetBg, styles.MethodGetFg
	case api.POST:
		return styles.MethodPostBg, styles.MethodPostFg
	case api.PUT:
		return styles.MethodPutBg, styles.MethodPutFg
	case api.PATCH:
		return styles.MethodPatchBg, styles.MethodPatchFg
	case api.DELETE:
		return styles.MethodDeleteBg, styles.MethodDeleteFg
	case api.HEAD:
		return styles.MethodHeadBg, styles.MethodHeadFg
	case api.OPTIONS:
		return styles.MethodOptionsBg, styles.MethodOptionsFg
	default:
		return styles.MethodGetBg, styles.MethodGetFg
	}
}

// View renders the request view
func (r RequestView) View(width, height int, active bool) string {
	var result strings.Builder

	// === REQUEST URL LINE ===
	// Method badge (same style as Collections panel badges)
	bg, fg := r.getMethodStyle()
	methodStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(fg).
		Background(bg).
		Padding(0, 1)

	// URL input with syntax highlighting or editing mode
	var urlContent string
	if r.editingURL {
		urlContent = r.renderURLWithCursor()
	} else {
		urlContent = r.renderURLWithHighlight(r.url)
	}

	// Combine method and URL on one line
	result.WriteString(methodStyle.Render(string(r.method)))
	result.WriteString("  ")
	result.WriteString(urlContent)
	result.WriteString("\n")

	// Separator line ABOVE tabs
	separatorStyle := lipgloss.NewStyle().Foreground(styles.Surface0)
	result.WriteString(separatorStyle.Render(strings.Repeat("─", width)))
	result.WriteString("\n")

	// === TABS BAR (no shortcuts displayed) ===
	tabBar := r.tabs.View(width)
	result.WriteString(tabBar)
	result.WriteString("\n")

	// Separator line below tabs
	result.WriteString(separatorStyle.Render(strings.Repeat("─", width)))
	result.WriteString("\n")

	// === TAB CONTENT AREA ===
	contentHeight := height - 4 // Account for URL line, 2 separators, tabs bar
	var tabContent string

	switch r.tabs.GetActive() {
	case "Params":
		tabContent = r.renderParamsTab(width, contentHeight, active)
	case "Authorization":
		tabContent = r.renderAuthTab(width, contentHeight)
	case "Headers":
		tabContent = r.renderHeadersTab(width, contentHeight, active)
	case "Body":
		tabContent = r.renderBodyTab(width, contentHeight)
	case "Scripts":
		tabContent = r.renderScriptsTab(width, contentHeight)
	default:
		tabContent = "Select a tab to configure the request"
	}

	result.WriteString(tabContent)

	return result.String()
}

// renderParamsTab renders the Query Parameters and Path Parameters tab
func (r *RequestView) renderParamsTab(width, height int, active bool) string {
	var result strings.Builder

	// Section headers
	sectionHeaderActive := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Lavender).
		Background(styles.Surface0).
		Padding(0, 1)

	sectionHeaderInactive := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Padding(0, 1)

	separatorStyle := lipgloss.NewStyle().Foreground(styles.Surface0)

	// Section tabs: Path Params | Query Params (Path first like Postman)
	if r.paramsSection == PathParamsSection {
		result.WriteString(sectionHeaderActive.Render("Path Params"))
	} else {
		result.WriteString(sectionHeaderInactive.Render("Path Params"))
	}
	result.WriteString(separatorStyle.Render("  │  "))
	if r.paramsSection == QueryParamsSection {
		result.WriteString(sectionHeaderActive.Render("Query Params"))
	} else {
		result.WriteString(sectionHeaderInactive.Render("Query Params"))
	}
	result.WriteString("\n")

	result.WriteString(separatorStyle.Render(strings.Repeat("─", width)))
	result.WriteString("\n")

	// Render the active section
	// Subtract 2 for section tabs line and separator line
	contentHeight := height - 2

	if r.paramsSection == PathParamsSection {
		if r.pathParams.RowCount() == 0 {
			emptyStyle := lipgloss.NewStyle().
				Foreground(styles.Subtext0).
				Width(width).
				Align(lipgloss.Center).
				Padding(2, 0)
			result.WriteString(emptyStyle.Render("No path parameters\n\nPath params use :name syntax in URL (e.g., /users/:id)\nPress n to add"))
		} else {
			result.WriteString(r.renderTableEnvStyle(r.pathParams, width, contentHeight, active))
		}
	} else {
		if r.paramsTable.RowCount() == 0 {
			emptyStyle := lipgloss.NewStyle().
				Foreground(styles.Subtext0).
				Width(width).
				Align(lipgloss.Center).
				Padding(2, 0)
			result.WriteString(emptyStyle.Render("No query parameters\n\nPress n to add a parameter"))
		} else {
			result.WriteString(r.renderTableEnvStyle(r.paramsTable, width, contentHeight, active))
		}
	}

	return result.String()
}

// renderAuthTab renders the Authorization tab (Envs style)
func (r *RequestView) renderAuthTab(width, height int) string {
	var result strings.Builder

	// Styles
	labelStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Width(16)

	valueStyle := lipgloss.NewStyle().
		Foreground(styles.Text)

	selectedStyle := lipgloss.NewStyle().
		Background(styles.Surface0).
		Foreground(styles.Lavender).
		Bold(true)

	editingStyle := lipgloss.NewStyle().
		Background(styles.Surface1).
		Foreground(styles.Green)

	arrowStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender)

	emptyStyle := lipgloss.NewStyle().Foreground(styles.Subtext0)

	variableStyle := lipgloss.NewStyle().Foreground(styles.URLVariable)

	separatorStyle := lipgloss.NewStyle().Foreground(styles.Surface0)

	// Helper to render auth field value with variable highlighting
	renderAuthValue := func(value string, isSelected, isEditing, maskValue bool) string {
		if value == "" {
			if isSelected && isEditing {
				return editingStyle.Render("█")
			}
			if isSelected {
				return selectedStyle.Render("(empty)")
			}
			return emptyStyle.Render("(empty)")
		}

		if isEditing {
			// In editing mode, show raw text with cursor
			return editingStyle.Render(value + "█")
		}

		if isSelected {
			// When selected, show full value with variable highlighting
			return renderTextWithVariables(value, selectedStyle, variableStyle, false)
		}

		// Not selected: show with variable highlighting (and optionally mask non-variables)
		return renderTextWithVariables(value, valueStyle, variableStyle, maskValue)
	}

	// Get visible fields for current auth type
	visibleFields := r.getVisibleAuthFields()

	// Render each visible field
	for _, field := range visibleFields {
		isSelected := r.authField == field
		var line strings.Builder

		// Selection indicator
		if isSelected {
			line.WriteString(arrowStyle.Render("▸ "))
		} else {
			line.WriteString("  ")
		}

		// Render based on field type
		switch field {
		case AuthFieldType:
			line.WriteString(labelStyle.Render("Type"))
			typeText := fmt.Sprintf("◀ %s ▶", r.authType.String())
			if isSelected {
				line.WriteString(selectedStyle.Render(typeText))
			} else {
				line.WriteString(valueStyle.Render(r.authType.String()))
			}

		case AuthFieldToken:
			line.WriteString(labelStyle.Render("Token"))
			line.WriteString(renderAuthValue(r.authToken, isSelected, r.authEditing, true))

		case AuthFieldPrefix:
			line.WriteString(labelStyle.Render("Prefix"))
			displayVal := r.authPrefix
			if displayVal == "" && !isSelected {
				displayVal = "Bearer"
			}
			if displayVal == "" {
				line.WriteString(renderAuthValue("", isSelected, r.authEditing, false))
			} else {
				line.WriteString(renderAuthValue(displayVal, isSelected, r.authEditing, false))
			}

		case AuthFieldUsername:
			line.WriteString(labelStyle.Render("Username"))
			line.WriteString(renderAuthValue(r.authUsername, isSelected, r.authEditing, false))

		case AuthFieldPassword:
			line.WriteString(labelStyle.Render("Password"))
			line.WriteString(renderAuthValue(r.authPassword, isSelected, r.authEditing, true))

		case AuthFieldAPIKeyName:
			line.WriteString(labelStyle.Render("Key Name"))
			line.WriteString(renderAuthValue(r.authAPIKeyName, isSelected, r.authEditing, false))

		case AuthFieldAPIKeyValue:
			line.WriteString(labelStyle.Render("Key Value"))
			line.WriteString(renderAuthValue(r.authAPIKeyValue, isSelected, r.authEditing, true))

		case AuthFieldAPIKeyLocation:
			line.WriteString(labelStyle.Render("Add to"))
			location := r.authAPIKeyLocation
			if location == "" {
				location = "header"
			}
			locationText := fmt.Sprintf("◀ %s ▶", location)
			if isSelected {
				line.WriteString(selectedStyle.Render(locationText))
			} else {
				line.WriteString(valueStyle.Render(location))
			}
		}

		result.WriteString(line.String())
		result.WriteString("\n")
	}

	// Separator
	result.WriteString("\n")
	result.WriteString(separatorStyle.Render(strings.Repeat("─", width)))
	result.WriteString("\n\n")

	// Help text based on auth type
	helpStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Italic(true)

	switch r.authType {
	case AuthNone:
		result.WriteString(helpStyle.Render("No authentication will be applied to this request"))
	case AuthBearer:
		prefix := r.authPrefix
		if prefix == "" {
			prefix = "Bearer"
		}
		result.WriteString(helpStyle.Render(fmt.Sprintf("Header: Authorization: %s <token>", prefix)))
	case AuthBasic:
		result.WriteString(helpStyle.Render("Header: Authorization: Basic <base64(username:password)>"))
	case AuthAPIKey:
		location := r.authAPIKeyLocation
		if location == "" {
			location = "header"
		}
		keyName := r.authAPIKeyName
		if keyName == "" {
			keyName = "X-API-Key"
		}
		if location == "header" {
			result.WriteString(helpStyle.Render(fmt.Sprintf("Header: %s: <value>", keyName)))
		} else {
			result.WriteString(helpStyle.Render(fmt.Sprintf("Query: ?%s=<value>", keyName)))
		}
	}

	return result.String()
}

// renderHeadersTab renders the HTTP Headers tab (Envs style)
func (r *RequestView) renderHeadersTab(width, height int, active bool) string {
	if r.headersTable.RowCount() == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Width(width).
			Align(lipgloss.Center).
			Padding(2, 0)

		return emptyStyle.Render("No custom headers\n\nPress n to add a header")
	}

	return r.renderTableEnvStyle(r.headersTable, width, height, active)
}

// renderBodyTab renders the Request Body tab
func (r *RequestView) renderBodyTab(width, height int) string {
	// Body content based on type - use full height for editor
	if r.bodyType == NoneBody {
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Width(width).
			Align(lipgloss.Center).
			Padding(2, 0)
		return emptyStyle.Render("No body content for this request")
	} else if r.bodyType == JSONBody {
		// Use full available height for the editor
		return r.bodyEditor.View(width, height, true)
	}

	// Other body types placeholder
	placeholderStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Italic(true).
		Padding(1, 0)
	return placeholderStyle.Render(fmt.Sprintf("%s editor not yet implemented", r.bodyType.String()))
}

// renderScriptsTab renders the Scripts tab
func (r *RequestView) renderScriptsTab(width, height int) string {
	var result strings.Builder

	// Section headers (Pre-request / Post-request)
	sectionHeaderActive := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Lavender).
		Background(styles.Surface0).
		Padding(0, 1)

	sectionHeaderInactive := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Padding(0, 1)

	separatorStyle := lipgloss.NewStyle().Foreground(styles.Surface0)

	// Section tabs: Pre-request | Post-request
	if r.scriptsSection == PreRequestSection {
		result.WriteString(sectionHeaderActive.Render("Pre-request"))
	} else {
		result.WriteString(sectionHeaderInactive.Render("Pre-request"))
	}
	result.WriteString(separatorStyle.Render("  │  "))
	if r.scriptsSection == PostRequestSection {
		result.WriteString(sectionHeaderActive.Render("Post-request"))
	} else {
		result.WriteString(sectionHeaderInactive.Render("Post-request"))
	}
	result.WriteString("\n")

	result.WriteString(separatorStyle.Render(strings.Repeat("─", width)))
	result.WriteString("\n")

	// Render the active section's editor
	// Subtract 2 for section tabs line and separator line
	editorHeight := height - 2

	if r.scriptsSection == PreRequestSection {
		result.WriteString(r.preRequestEditor.View(width, editorHeight, true))
	} else {
		result.WriteString(r.postRequestEditor.View(width, editorHeight, true))
	}

	return result.String()
}

// renderTableEnvStyle renders a table in Envs panel style (like Collections tree)
func (r *RequestView) renderTableEnvStyle(table *components.Table, width, height int, active bool) string {
	var lines []string

	for i, row := range table.Rows {
		isSelected := i == table.Cursor

		// Build row: > [] key   value (like Envs panel)
		var line strings.Builder

		// Checkbox based on enabled state
		if row.Enabled {
			checkStyle := lipgloss.NewStyle().Foreground(styles.CheckboxOn)
			line.WriteString(checkStyle.Render("☑"))
		} else {
			checkStyle := lipgloss.NewStyle().Foreground(styles.CheckboxOff)
			line.WriteString(checkStyle.Render("☐"))
		}
		line.WriteString(" ")

		// Key (dimmed if disabled)
		keyStyle := lipgloss.NewStyle().Foreground(styles.Subtext1)
		if !row.Enabled {
			keyStyle = keyStyle.Foreground(styles.Subtext0)
		}
		keyWidth := 20
		key := row.Key
		// Truncate key to fit (no ellipsis - just cut)
		if len(key) > keyWidth {
			key = key[:keyWidth]
		}
		// Pad key to align values
		keyPadded := key + strings.Repeat(" ", keyWidth-len(key))
		line.WriteString(keyStyle.Render(keyPadded))

		line.WriteString("   ")

		// Calculate available width for value: width - checkbox(2) - key(20) - separator(3)
		valueWidth := width - 2 - keyWidth - 3
		if valueWidth < 3 {
			valueWidth = 3
		}

		// Value (highlight variables, dimmed if disabled)
		value := row.Value
		// Truncate value to fit (no ellipsis - just cut)
		if len(value) > valueWidth {
			value = value[:valueWidth]
		}
		if strings.Contains(row.Value, "{{") {
			valueStyle := lipgloss.NewStyle().Foreground(styles.URLVariable)
			if !row.Enabled {
				valueStyle = valueStyle.Foreground(styles.Subtext0)
			}
			line.WriteString(valueStyle.Render(value))
		} else {
			valueStyle := lipgloss.NewStyle().Foreground(styles.Text)
			if !row.Enabled {
				valueStyle = valueStyle.Foreground(styles.Subtext0)
			}
			line.WriteString(valueStyle.Render(value))
		}

		// Apply selection styling (like Collections tree)
		lineStr := line.String()
		style := lipgloss.NewStyle().Width(width)
		if isSelected {
			if active {
				// Active panel selection
				style = style.Background(styles.SelectedPanelBg).Foreground(styles.SelectedPanelFg).Bold(true)
			} else {
				// Inactive panel selection
				style = style.Background(styles.SelectedRequestBg).Foreground(styles.SelectedRequestFg)
			}
		}

		lines = append(lines, style.Render(lineStr))
	}

	return strings.Join(lines, "\n")
}

// GetMethod returns the current HTTP method
func (r *RequestView) GetMethod() string {
	return string(r.method)
}

// GetURL returns the current URL
func (r *RequestView) GetURL() string {
	return r.url
}

// IsEditingURL returns whether the URL input is being edited
func (r *RequestView) IsEditingURL() bool {
	return r.editingURL
}

// SetEditingURL sets the URL editing state
func (r *RequestView) SetEditingURL(editing bool) {
	r.editingURL = editing
	if editing {
		r.urlCursor = len(r.url)
	}
}

// GetTitle returns a formatted title for the panel header
func (r *RequestView) GetTitle() string {
	return fmt.Sprintf("%s  %s", r.method, r.url)
}

// LoadRequest loads a request from the tree selection
func (r *RequestView) LoadRequest(id, name, method, url string) {
	// Store current request info for saving changes
	r.currentRequestID = id
	r.currentRequestName = name

	// Set HTTP method
	switch method {
	case "GET":
		r.method = api.GET
	case "POST":
		r.method = api.POST
	case "PUT":
		r.method = api.PUT
	case "DELETE":
		r.method = api.DELETE
	case "PATCH":
		r.method = api.PATCH
	case "HEAD":
		r.method = api.HEAD
	case "OPTIONS":
		r.method = api.OPTIONS
	default:
		r.method = api.GET
	}

	// Set URL from request
	r.url = url

	// Clear existing params and headers
	r.paramsTable.Rows = nil
	r.headersTable.Rows = nil

	// Parse URL to extract query params
	r.ParseURLParams()
}

// GetCurrentRequestID returns the ID of the currently loaded request
func (r *RequestView) GetCurrentRequestID() string {
	return r.currentRequestID
}

// GetHeadersTable returns the headers table for HTTP request building
func (r *RequestView) GetHeadersTable() *components.Table {
	return r.headersTable
}

// GetBodyContent returns the body content from the body editor
func (r *RequestView) GetBodyContent() string {
	if r.bodyType == NoneBody {
		return ""
	}
	return r.bodyEditor.GetContent()
}

// LoadCollectionRequest loads a full CollectionRequest with all its data
func (r *RequestView) LoadCollectionRequest(req *api.CollectionRequest) {
	if req == nil {
		return
	}

	// Store current request info for saving changes
	r.currentRequestID = req.ID
	r.currentRequestName = req.Name

	// Set HTTP method
	r.method = req.Method

	// Set URL
	r.url = req.URL

	// Clear and load params
	r.paramsTable.Rows = nil
	for _, param := range req.Params {
		r.paramsTable.AddRowWithState(param.Key, param.Value, param.Enabled)
	}
	// Also parse URL for any params not already in the list
	r.ParseURLParams()

	// Clear and load headers
	r.headersTable.Rows = nil
	for _, header := range req.Headers {
		r.headersTable.AddRowWithState(header.Key, header.Value, header.Enabled)
	}

	// Add default headers if none exist (like Postman)
	if r.headersTable.RowCount() == 0 {
		r.addDefaultHeaders()
	}

	// Reset cursors
	if r.paramsTable.RowCount() > 0 {
		r.paramsTable.Cursor = 0
	} else {
		r.paramsTable.Cursor = -1
	}

	if r.headersTable.RowCount() > 0 {
		r.headersTable.Cursor = 0
	} else {
		r.headersTable.Cursor = -1
	}

	// Load body content
	if req.Body != nil {
		r.bodyType = JSONBody // Default to JSON
		switch req.Body.Type {
		case "json":
			r.bodyType = JSONBody
		case "raw":
			r.bodyType = RawBody
		case "form-data":
			r.bodyType = FormDataBody
		case "binary":
			r.bodyType = BinaryBody
		case "none":
			r.bodyType = NoneBody
		}

		// Convert body content to string for editor
		var bodyContent string
		switch content := req.Body.Content.(type) {
		case string:
			bodyContent = content
		case map[string]interface{}, []interface{}:
			// Re-encode JSON content
			if jsonBytes, err := json.MarshalIndent(content, "", "  "); err == nil {
				bodyContent = string(jsonBytes)
			}
		}

		if bodyContent != "" {
			r.bodyEditor = components.NewEditor(bodyContent, "json")
		}
	} else {
		// No body - set empty editor
		r.bodyType = JSONBody
		r.bodyEditor = components.NewEditor(`{

}`, "json")
	}

	// Load scripts content
	if req.Scripts != nil {
		if req.Scripts.PreRequest != "" {
			r.preRequestEditor = components.NewEditor(req.Scripts.PreRequest, "javascript")
		} else {
			r.preRequestEditor = components.NewEditor(`// Pre-request script
// Runs before the request is sent

`, "javascript")
		}

		if req.Scripts.PostRequest != "" {
			r.postRequestEditor = components.NewEditor(req.Scripts.PostRequest, "javascript")
		} else {
			r.postRequestEditor = components.NewEditor(`// Post-request script
// Runs after the response is received

`, "javascript")
		}
	} else {
		// No scripts - set default editors
		r.preRequestEditor = components.NewEditor(`// Pre-request script
// Runs before the request is sent

`, "javascript")
		r.postRequestEditor = components.NewEditor(`// Post-request script
// Runs after the response is received

`, "javascript")
	}

	// Load auth configuration
	r.loadAuthFromRequest(req)
}

// loadAuthFromRequest loads authentication configuration from a CollectionRequest
func (r *RequestView) loadAuthFromRequest(req *api.CollectionRequest) {
	// Reset auth fields to defaults
	r.authType = AuthNone
	r.authToken = ""
	r.authPrefix = "Bearer"
	r.authUsername = ""
	r.authPassword = ""
	r.authAPIKeyName = ""
	r.authAPIKeyValue = ""
	r.authAPIKeyLocation = "header"
	r.authField = AuthFieldType
	r.authEditing = false

	if req == nil || req.Auth == nil {
		return
	}

	auth := req.Auth

	// Parse auth type
	switch auth.Type {
	case "bearer":
		r.authType = AuthBearer
		r.authToken = auth.Token
		if auth.Prefix != "" {
			r.authPrefix = auth.Prefix
		}
	case "basic":
		r.authType = AuthBasic
		r.authUsername = auth.Username
		r.authPassword = auth.Password
	case "apikey":
		r.authType = AuthAPIKey
		r.authAPIKeyName = auth.APIKeyName
		r.authAPIKeyValue = auth.APIKeyValue
		if auth.APIKeyLocation != "" {
			r.authAPIKeyLocation = auth.APIKeyLocation
		}
	default:
		r.authType = AuthNone
	}
}
