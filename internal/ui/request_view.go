package ui

import (
	"fmt"
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

// RequestView represents the request builder panel
type RequestView struct {
	method       api.HTTPMethod
	url          string
	tabs         *components.Tabs
	paramsTable  *components.Table
	headersTable *components.Table
	bodyEditor   *components.Editor
	bodyType     BodyType
	authType     string
	authToken    string
	focused      string // "method", "url", "tabs"
}

// NewRequestView creates a new request view
func NewRequestView() *RequestView {
	tabs := components.NewTabs([]string{
		"Params",
		"Authorization",
		"Headers",
		"Body",
		"Scripts",
		"Settings",
	})

	paramsTable := components.NewTable([]string{"Key", "Value"})
	headersTable := components.NewTable([]string{"Key", "Value"})

	// Add default headers
	headersTable.AddRow("Content-Type", "application/json")
	headersTable.AddRow("Accept", "application/json")

	// Initialize body editor with sample JSON
	bodyEditor := components.NewEditor(`{
  "name": "John Doe",
  "email": "john@example.com"
}`, "json")

	return &RequestView{
		method:       api.GET,
		url:          "{{base_url}}/admin/users/:id",
		tabs:         tabs,
		paramsTable:  paramsTable,
		headersTable: headersTable,
		bodyEditor:   bodyEditor,
		bodyType:     JSONBody,
		authType:     "Bearer Token",
		authToken:    "",
		focused:      "url",
	}
}

// Update handles messages for the request view
func (r RequestView) Update(msg tea.Msg, cfg *config.GlobalConfig) (RequestView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle send request
		if msg.String() == "ctrl+s" {
			// TODO: Send HTTP request
			return r, nil
		}

		// Tab navigation (Tab key or numbers 1-6)
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
		case "6":
			r.tabs.SetActive(5) // Settings
		}

		// Handle table navigation for Params and Headers tabs
		activeTab := r.tabs.GetActive()
		if activeTab == "Params" {
			switch msg.String() {
			case "up", "k":
				r.paramsTable.MoveUp()
			case "down", "j":
				r.paramsTable.MoveDown()
			}
		} else if activeTab == "Headers" {
			switch msg.String() {
			case "up", "k":
				r.headersTable.MoveUp()
			case "down", "j":
				r.headersTable.MoveDown()
			}
		}
	}

	return r, nil
}

// View renders the request view
func (r RequestView) View(width, height int, active bool) string {
	var result strings.Builder

	// Method and URL line - simple without borders
	methodStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Green).
		Background(styles.Mantle).
		Padding(0, 1)

	urlStyle := lipgloss.NewStyle().
		Foreground(styles.Text).
		Background(styles.Mantle).
		Padding(0, 1)

	methodButton := methodStyle.Render(string(r.method))
	urlInput := urlStyle.Render(r.url)

	result.WriteString(methodButton)
	result.WriteString(" ")
	result.WriteString(urlInput)
	result.WriteString("\n")

	// Separator
	result.WriteString(strings.Repeat("─", width))
	result.WriteString("\n")

	// Tabs - compact
	tabBar := r.tabs.View(width)
	result.WriteString(tabBar)
	result.WriteString("\n")

	// Separator
	result.WriteString(strings.Repeat("─", width))
	result.WriteString("\n")

	// Tab content area - simple, no borders
	var tabContent string

	switch r.tabs.GetActive() {
	case "Params":
		tabContent = r.renderParamsTab(width, height-6)
	case "Authorization":
		tabContent = r.renderAuthTab(width, height-6)
	case "Headers":
		tabContent = r.renderHeadersTab(width, height-6)
	case "Body":
		tabContent = r.renderBodyTab(width, height-6)
	case "Scripts":
		tabContent = r.renderScriptsTab(width, height-6)
	case "Settings":
		tabContent = r.renderSettingsTab(width, height-6)
	default:
		tabContent = "Select a tab to configure the request"
	}

	result.WriteString(tabContent)

	return result.String()
}

func (r *RequestView) renderParamsTab(width, height int) string {
	var result strings.Builder
	result.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Mauve).
		Render("Query Parameters"))
	result.WriteString("\n")

	result.WriteString(lipgloss.NewStyle().
		Foreground(styles.Subtext1).
		Render("Editable table for query parameters"))
	result.WriteString("\n")
	result.WriteString(lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Render("Press 'a' to add, 'd' to delete, Enter to edit"))
	result.WriteString("\n\n")
	result.WriteString(r.paramsTable.View(width, height-5))

	return result.String()
}

func (r *RequestView) renderAuthTab(width, height int) string {
	var result strings.Builder
	result.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Mauve).
		Render("Authorization"))
	result.WriteString("\n")

	result.WriteString(fmt.Sprintf("Type: %s\n", r.authType))
	result.WriteString("Token: {{api_token}}\n")
	result.WriteString(lipgloss.NewStyle().
		Foreground(styles.Subtext1).
		Render("\nConfigure authorization headers and tokens"))

	return result.String()
}

func (r *RequestView) renderHeadersTab(width, height int) string {
	var result strings.Builder
	result.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Mauve).
		Render("HTTP Headers"))
	result.WriteString("\n")

	result.WriteString(lipgloss.NewStyle().
		Foreground(styles.Subtext1).
		Render("Editable table for headers"))
	result.WriteString("\n")
	result.WriteString(lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Render("Press 'a' to add, 'd' to delete, Enter to edit"))
	result.WriteString("\n\n")
	result.WriteString(r.headersTable.View(width, height-5))

	return result.String()
}

func (r *RequestView) renderBodyTab(width, height int) string {
	var result strings.Builder

	// Body type selector
	typeStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0)

	activeTypeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Lavender).
		Background(styles.Surface0).
		Padding(0, 1)

	inactiveTypeStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext1).
		Padding(0, 1)

	result.WriteString(typeStyle.Render("Type: "))

	bodyTypes := []BodyType{NoneBody, JSONBody, FormDataBody, RawBody, BinaryBody}
	for i, bt := range bodyTypes {
		if bt == r.bodyType {
			result.WriteString(activeTypeStyle.Render(bt.String()))
		} else {
			result.WriteString(inactiveTypeStyle.Render(bt.String()))
		}
		if i < len(bodyTypes)-1 {
			result.WriteString(" ")
		}
	}
	result.WriteString("\n")
	result.WriteString(strings.Repeat("─", width))
	result.WriteString("\n")

	// Body content based on type
	if r.bodyType == NoneBody {
		result.WriteString(lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Render("No body content for this request"))
	} else if r.bodyType == JSONBody {
		result.WriteString(r.bodyEditor.View(width, height-4, true))
	} else {
		result.WriteString(lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Render(fmt.Sprintf("%s editor not yet implemented", r.bodyType.String())))
	}

	return result.String()
}

func (r *RequestView) renderScriptsTab(width, height int) string {
	var result strings.Builder
	result.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Mauve).
		Render("Pre/Post Request Scripts"))
	result.WriteString("\n")

	result.WriteString(lipgloss.NewStyle().
		Foreground(styles.Subtext1).
		Render("JavaScript Editor"))
	result.WriteString("\n")
	result.WriteString(strings.Repeat("─", width))
	result.WriteString("\n")
	result.WriteString("// Pre-request script\n")
	result.WriteString("console.log('Request about to be sent');\n\n")
	result.WriteString("// Post-request script\n")
	result.WriteString("console.log('Response received:', response);")

	return result.String()
}

func (r *RequestView) renderSettingsTab(width, height int) string {
	var result strings.Builder
	result.WriteString(lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Mauve).
		Render("Request Settings"))
	result.WriteString("\n\n")

	result.WriteString("☐ Follow redirects\n")
	result.WriteString("☐ Verify SSL certificates\n")
	result.WriteString("☑ Send cookies\n")
	result.WriteString("Timeout: 30s\n")
	result.WriteString("Max redirects: 5")

	return result.String()
}

// GetMethod returns the current HTTP method
func (r *RequestView) GetMethod() string {
	return string(r.method)
}

// GetURL returns the current URL
func (r *RequestView) GetURL() string {
	return r.url
}

// GetTitle returns a formatted title for the panel header
func (r *RequestView) GetTitle() string {
	return fmt.Sprintf("%s  %s", r.method, r.url)
}

// LoadRequest loads a request from the tree selection
func (r *RequestView) LoadRequest(id, name, method, url string) {
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
}
