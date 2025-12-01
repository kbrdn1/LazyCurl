package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kbrdn1/LazyCurl/internal/config"
	"github.com/kbrdn1/LazyCurl/internal/ui/components"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// StatusBadge represents a styled HTTP status badge
type StatusBadge struct {
	Code    int
	Text    string
	BgColor lipgloss.Color
	FgColor lipgloss.Color
}

// NewStatusBadge creates a status badge with appropriate colors
func NewStatusBadge(code int) StatusBadge {
	badge := StatusBadge{Code: code}

	switch {
	case code >= 200 && code < 300:
		badge.Text = fmt.Sprintf("%d OK", code)
		badge.BgColor = styles.Status2xxBg
		badge.FgColor = styles.Status2xxFg
	case code >= 300 && code < 400:
		badge.Text = fmt.Sprintf("%d Redirect", code)
		badge.BgColor = styles.Status3xxBg
		badge.FgColor = styles.Status3xxFg
	case code >= 400 && code < 500:
		badge.Text = fmt.Sprintf("%d Client Error", code)
		badge.BgColor = styles.Status4xxBg
		badge.FgColor = styles.Status4xxFg
	case code >= 500:
		badge.Text = fmt.Sprintf("%d Server Error", code)
		badge.BgColor = styles.Status5xxBg
		badge.FgColor = styles.Status5xxFg
	default:
		badge.Text = "No Response"
		badge.BgColor = styles.Surface1
		badge.FgColor = styles.Text
	}

	return badge
}

// Render returns the styled status badge string
func (s StatusBadge) Render() string {
	style := lipgloss.NewStyle().
		Bold(true).
		Background(s.BgColor).
		Foreground(s.FgColor).
		Padding(0, 1)
	return style.Render(s.Text)
}

// ResponseView represents the response viewer panel
type ResponseView struct {
	statusCode   int
	status       string
	headers      map[string]string
	cookies      map[string]string
	body         string
	time         string
	size         string
	tabs         *components.Tabs
	bodyEditor   *components.Editor
	statusBadge  StatusBadge
	scrollOffset int
}

// NewResponseView creates a new response view
func NewResponseView() *ResponseView {
	tabs := components.NewTabs([]string{
		"Body",
		"Cookies",
		"Headers",
	})

	// Initialize body editor for viewing response
	bodyEditor := components.NewEditor("", "json")
	bodyEditor.SetReadOnly(true)

	return &ResponseView{
		statusCode:   0,
		status:       "No response yet",
		headers:      make(map[string]string),
		cookies:      make(map[string]string),
		body:         "",
		time:         "0ms",
		size:         "0B",
		tabs:         tabs,
		bodyEditor:   bodyEditor,
		statusBadge:  NewStatusBadge(0),
		scrollOffset: 0,
	}
}

// Update handles messages for the response view
func (r ResponseView) Update(msg tea.Msg, cfg *config.GlobalConfig) (ResponseView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Tab navigation
		switch msg.String() {
		case "tab":
			r.tabs.Next()
		case "shift+tab":
			r.tabs.Previous()
		case "1":
			r.tabs.SetActive(0) // Body
		case "2":
			r.tabs.SetActive(1) // Cookies
		case "3":
			r.tabs.SetActive(2) // Headers
		}

		// VIEW mode scrolling in body
		activeTab := r.tabs.GetActive()
		if activeTab == "Body" {
			editor, _ := r.bodyEditor.Update(msg, false) // Read-only navigation
			r.bodyEditor = editor
		}
	}

	return r, nil
}

// View renders the response view
func (r ResponseView) View(width, height int, active bool) string {
	var result strings.Builder

	// Status bar with badge, time, and size
	if r.statusCode > 0 {
		result.WriteString(r.statusBadge.Render())
		result.WriteString(" ")

		timeStyle := lipgloss.NewStyle().
			Foreground(styles.Teal)
		result.WriteString(timeStyle.Render(fmt.Sprintf("Time: %s", r.time)))
		result.WriteString(" ")

		sizeStyle := lipgloss.NewStyle().
			Foreground(styles.Peach)
		result.WriteString(sizeStyle.Render(fmt.Sprintf("Size: %s", r.size)))
		result.WriteString("\n")
	}

	// Tabs
	tabBar := r.tabs.View(width)
	result.WriteString(tabBar)
	result.WriteString("\n")

	// Separator
	result.WriteString(strings.Repeat("─", width))
	result.WriteString("\n")

	// Tab content
	var tabContent string
	contentHeight := height - 4
	if r.statusCode > 0 {
		contentHeight-- // Account for status bar
	}

	if r.statusCode == 0 {
		tabContent = lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Render("No response yet. Send a request with Ctrl+S")
	} else {
		switch r.tabs.GetActive() {
		case "Body":
			tabContent = r.renderBodyTab(width, contentHeight)
		case "Cookies":
			tabContent = r.renderCookiesTab(width, contentHeight)
		case "Headers":
			tabContent = r.renderHeadersTab(width, contentHeight)
		default:
			tabContent = "Select a tab to view response details"
		}
	}

	result.WriteString(tabContent)

	return result.String()
}

func (r *ResponseView) renderBodyTab(width, height int) string {
	if r.body == "" {
		return lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Render("No body content")
	}

	return r.bodyEditor.View(width, height, true)
}

func (r *ResponseView) renderCookiesTab(width, height int) string {
	var result strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Blue)

	result.WriteString(headerStyle.Render(fmt.Sprintf("%-25s %s", "Name", "Value")))
	result.WriteString("\n")
	result.WriteString(strings.Repeat("─", width))
	result.WriteString("\n")

	if len(r.cookies) == 0 {
		result.WriteString(lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Render("No cookies in response"))
	} else {
		for key, value := range r.cookies {
			keyStyle := lipgloss.NewStyle().Foreground(styles.Text)
			valueStyle := lipgloss.NewStyle().Foreground(styles.Subtext1)
			result.WriteString(keyStyle.Render(fmt.Sprintf("%-25s", key)))
			result.WriteString(valueStyle.Render(truncateString(value, width-26)))
			result.WriteString("\n")
		}
	}

	return result.String()
}

func (r *ResponseView) renderHeadersTab(width, height int) string {
	var result strings.Builder

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Blue)

	result.WriteString(headerStyle.Render(fmt.Sprintf("%-25s %s", "Header", "Value")))
	result.WriteString("\n")
	result.WriteString(strings.Repeat("─", width))
	result.WriteString("\n")

	if len(r.headers) == 0 {
		result.WriteString(lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Render("No headers in response"))
	} else {
		for key, value := range r.headers {
			keyStyle := lipgloss.NewStyle().Foreground(styles.Text)
			valueStyle := lipgloss.NewStyle().Foreground(styles.Subtext1)
			result.WriteString(keyStyle.Render(fmt.Sprintf("%-25s", key)))
			result.WriteString(valueStyle.Render(truncateString(value, width-26)))
			result.WriteString("\n")
		}
	}

	return result.String()
}

// SetResponse updates the response view with new data
func (r *ResponseView) SetResponse(statusCode int, status string, headers map[string]string, cookies map[string]string, body string, time string, size string) {
	r.statusCode = statusCode
	r.status = status
	r.headers = headers
	r.cookies = cookies
	r.body = body
	r.time = time
	r.size = size
	r.statusBadge = NewStatusBadge(statusCode)

	// Update body editor with response body
	r.bodyEditor.SetContent(body)
}

// ClearResponse clears the response view
func (r *ResponseView) ClearResponse() {
	r.statusCode = 0
	r.status = "No response yet"
	r.headers = make(map[string]string)
	r.cookies = make(map[string]string)
	r.body = ""
	r.time = "0ms"
	r.size = "0B"
	r.statusBadge = NewStatusBadge(0)
	r.bodyEditor.SetContent("")
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// GetStatusCode returns the current status code
func (r *ResponseView) GetStatusCode() int {
	return r.statusCode
}

// GetResponseTime returns the response time
func (r *ResponseView) GetResponseTime() string {
	return r.time
}

// GetResponseSize returns the response size
func (r *ResponseView) GetResponseSize() string {
	return r.size
}
