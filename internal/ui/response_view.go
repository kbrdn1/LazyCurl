package ui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/internal/api"
	"github.com/kbrdn1/LazyCurl/internal/config"
	"github.com/kbrdn1/LazyCurl/internal/session"
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
	isLoading    bool // Whether a request is in progress
	loaderFrame  int  // Animation frame for loader

	// Cursor tracking for vim-like navigation
	headersCursor int
	cookiesCursor int
	headersKeys   []string // Sorted header keys for stable iteration
	cookiesKeys   []string // Sorted cookie keys for stable iteration

	// Console view
	consoleView *ConsoleView

	// Test results from script assertions
	testResults       []api.AssertionResult
	testResultsCursor int // Cursor for navigating test results
}

// NewResponseView creates a new response view
func NewResponseView() *ResponseView {
	tabs := components.NewTabs([]string{
		"Body",
		"Cookies",
		"Headers",
		"Tests",
		"Console",
	})

	// Initialize body editor for viewing response
	bodyEditor := components.NewEditor("", "json")
	bodyEditor.SetReadOnly(true)

	return &ResponseView{
		statusCode:        0,
		status:            "No response yet",
		headers:           make(map[string]string),
		cookies:           make(map[string]string),
		body:              "",
		time:              "0ms",
		size:              "0B",
		tabs:              tabs,
		bodyEditor:        bodyEditor,
		statusBadge:       NewStatusBadge(0),
		scrollOffset:      0,
		headersCursor:     0,
		cookiesCursor:     0,
		headersKeys:       []string{},
		cookiesKeys:       []string{},
		consoleView:       NewConsoleView(),
		testResults:       []api.AssertionResult{},
		testResultsCursor: 0,
	}
}

// Update handles messages for the response view
func (r ResponseView) Update(msg tea.Msg, cfg *config.GlobalConfig) (ResponseView, tea.Cmd) {
	return r.UpdateWithHistory(msg, cfg, nil)
}

// UpdateWithHistory handles messages for the response view with console history
func (r ResponseView) UpdateWithHistory(msg tea.Msg, cfg *config.GlobalConfig, history *api.ConsoleHistory) (ResponseView, tea.Cmd) {
	switch msg := msg.(type) {
	case components.SearchUpdateMsg, components.SearchCloseMsg:
		// Forward search messages to body editor
		if r.tabs.GetActive() == "Body" {
			editor, cmd := r.bodyEditor.Update(msg, false)
			r.bodyEditor = editor
			return r, cmd
		}
		return r, nil

	case tea.KeyMsg:
		activeTab := r.tabs.GetActive()

		// Tab navigation with Tab key - but not when searching
		if !r.bodyEditor.IsSearching() {
			switch msg.String() {
			case "tab":
				r.tabs.Next()
				return r, nil
			case "shift+tab":
				r.tabs.Previous()
				return r, nil
			case "1":
				r.tabs.SetActive(0) // Body
				return r, nil
			case "2":
				r.tabs.SetActive(1) // Cookies
				return r, nil
			case "3":
				r.tabs.SetActive(2) // Headers
				return r, nil
			case "4":
				r.tabs.SetActive(3) // Tests
				return r, nil
			case "5":
				r.tabs.SetActive(4) // Console
				return r, nil
			}
		}

		// Tab-specific navigation
		switch activeTab {
		case "Body":
			// Forward all keys to body editor for vim-like navigation
			editor, cmd := r.bodyEditor.Update(msg, false) // Read-only navigation
			r.bodyEditor = editor
			return r, cmd

		case "Cookies":
			// Vim-like navigation in cookies list
			switch msg.String() {
			case "j", "down":
				if r.cookiesCursor < len(r.cookiesKeys)-1 {
					r.cookiesCursor++
				}
			case "k", "up":
				if r.cookiesCursor > 0 {
					r.cookiesCursor--
				}
			case "g":
				r.cookiesCursor = 0
			case "G":
				if len(r.cookiesKeys) > 0 {
					r.cookiesCursor = len(r.cookiesKeys) - 1
				}
			}

		case "Headers":
			// Vim-like navigation in headers list
			switch msg.String() {
			case "j", "down":
				if r.headersCursor < len(r.headersKeys)-1 {
					r.headersCursor++
				}
			case "k", "up":
				if r.headersCursor > 0 {
					r.headersCursor--
				}
			case "g":
				r.headersCursor = 0
			case "G":
				if len(r.headersKeys) > 0 {
					r.headersCursor = len(r.headersKeys) - 1
				}
			}

		case "Tests":
			// Vim-like navigation in test results list
			switch msg.String() {
			case "j", "down":
				if r.testResultsCursor < len(r.testResults)-1 {
					r.testResultsCursor++
				}
			case "k", "up":
				if r.testResultsCursor > 0 {
					r.testResultsCursor--
				}
			case "g":
				r.testResultsCursor = 0
			case "G":
				if len(r.testResults) > 0 {
					r.testResultsCursor = len(r.testResults) - 1
				}
			}

		case "Console":
			// Forward keys to console view
			if history != nil {
				consoleView, cmd := r.consoleView.Update(msg, history, cfg)
				r.consoleView = &consoleView
				return r, cmd
			}
		}
	}

	return r, nil
}

// GetActiveTab returns the currently active tab name
func (r *ResponseView) GetActiveTab() string {
	return r.tabs.GetActive()
}

// View renders the response view
func (r ResponseView) View(width, height int, active bool) string {
	return r.ViewWithHistory(width, height, active, nil)
}

// ViewWithHistory renders the response view with console history support
func (r ResponseView) ViewWithHistory(width, height int, active bool, history *api.ConsoleHistory) string {
	var result strings.Builder

	// Show loading bar if request is in progress
	if r.isLoading {
		// Use the horizontal loader from components
		loaderLine := components.HorizontalLoader(width, r.loaderFrame, "Sending request")
		result.WriteString(loaderLine)
		result.WriteString("\n")
	} else if r.statusCode > 0 && r.tabs.GetActive() != "Console" {
		// Status bar with badge and icons for time/size aligned to right (not shown for Console tab)
		statusPart := r.statusBadge.Render()

		// Right-aligned time and size with Nerd Font / Unicode icons
		// Using:  (nf-fa-clock) or ◷ for time,  (nf-fa-database) or ◊ for size
		timeStyle := lipgloss.NewStyle().Foreground(styles.Teal)
		sizeStyle := lipgloss.NewStyle().Foreground(styles.Peach)

		// Unicode icons that work in most terminals
		timeIcon := "◷" // Clock icon
		sizeIcon := "◆" // Size/data icon

		timeText := timeStyle.Render(fmt.Sprintf("%s %s", timeIcon, r.time))
		sizeText := sizeStyle.Render(fmt.Sprintf("%s %s", sizeIcon, r.size))
		rightPart := timeText + "  " + sizeText

		// Calculate padding to align right part to the right
		statusLen := lipgloss.Width(statusPart)
		rightLen := lipgloss.Width(rightPart)
		padding := width - statusLen - rightLen - 2
		if padding < 1 {
			padding = 1
		}

		result.WriteString(statusPart)
		result.WriteString(strings.Repeat(" ", padding))
		result.WriteString(rightPart)
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
	// Calculate content height: total height minus tab bar (1), separator (1)
	contentHeight := height - 2
	if (r.statusCode > 0 || r.isLoading) && r.tabs.GetActive() != "Console" {
		contentHeight-- // Account for status bar or loading bar
	}
	if contentHeight < 3 {
		contentHeight = 3
	}

	activeTab := r.tabs.GetActive()

	// Console and Tests tabs are always available regardless of response status
	if activeTab == "Console" {
		tabContent = r.consoleView.View(width, contentHeight, history, active)
	} else if activeTab == "Tests" {
		tabContent = r.renderTestsTab(width, contentHeight)
	} else if r.isLoading {
		// Show loading message in content area
		loadingStyle := lipgloss.NewStyle().
			Foreground(styles.Blue).
			Italic(true)
		tabContent = loadingStyle.Render("Waiting for response...")
	} else if r.statusCode == 0 {
		tabContent = lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Render("No response yet. Send a request with Ctrl+S")
	} else {
		switch activeTab {
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

	if len(r.cookiesKeys) == 0 {
		result.WriteString(lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Render("No cookies in response"))
	} else {
		// Calculate how many rows we can show
		visibleRows := height - 2 // Account for header and separator
		startIdx := 0
		if r.cookiesCursor >= visibleRows {
			startIdx = r.cookiesCursor - visibleRows + 1
		}

		for i := startIdx; i < len(r.cookiesKeys) && i < startIdx+visibleRows; i++ {
			key := r.cookiesKeys[i]
			value := r.cookies[key]

			// Truncate key and value to fit width
			keyWidth := 25
			valueWidth := width - keyWidth - 1
			if len(key) > keyWidth {
				key = key[:keyWidth]
			}
			if len(value) > valueWidth && valueWidth > 0 {
				value = value[:valueWidth]
			}

			// Highlight selected row
			if i == r.cookiesCursor {
				rowStyle := lipgloss.NewStyle().
					Background(styles.Surface1).
					Foreground(styles.Text)
				row := fmt.Sprintf("%-25s %s", key, value)
				// Pad to full width
				if len(row) < width {
					row += strings.Repeat(" ", width-len(row))
				}
				result.WriteString(rowStyle.Render(row))
			} else {
				keyStyle := lipgloss.NewStyle().Foreground(styles.Text)
				valueStyle := lipgloss.NewStyle().Foreground(styles.Subtext1)
				result.WriteString(keyStyle.Render(fmt.Sprintf("%-25s", key)))
				result.WriteString(valueStyle.Render(value))
			}
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

	if len(r.headersKeys) == 0 {
		result.WriteString(lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Render("No headers in response"))
	} else {
		// Calculate how many rows we can show
		visibleRows := height - 2 // Account for header and separator
		startIdx := 0
		if r.headersCursor >= visibleRows {
			startIdx = r.headersCursor - visibleRows + 1
		}

		for i := startIdx; i < len(r.headersKeys) && i < startIdx+visibleRows; i++ {
			key := r.headersKeys[i]
			value := r.headers[key]

			// Truncate key and value to fit width
			keyWidth := 25
			valueWidth := width - keyWidth - 1
			if len(key) > keyWidth {
				key = key[:keyWidth]
			}
			if len(value) > valueWidth && valueWidth > 0 {
				value = value[:valueWidth]
			}

			// Highlight selected row
			if i == r.headersCursor {
				rowStyle := lipgloss.NewStyle().
					Background(styles.Surface1).
					Foreground(styles.Text)
				row := fmt.Sprintf("%-25s %s", key, value)
				// Pad to full width
				if len(row) < width {
					row += strings.Repeat(" ", width-len(row))
				}
				result.WriteString(rowStyle.Render(row))
			} else {
				keyStyle := lipgloss.NewStyle().Foreground(styles.Text)
				valueStyle := lipgloss.NewStyle().Foreground(styles.Subtext1)
				result.WriteString(keyStyle.Render(fmt.Sprintf("%-25s", key)))
				result.WriteString(valueStyle.Render(value))
			}
			result.WriteString("\n")
		}
	}

	return result.String()
}

func (r *ResponseView) renderTestsTab(width, height int) string {
	var result strings.Builder

	// Summary header
	passed := 0
	failed := 0
	for _, test := range r.testResults {
		if test.Passed {
			passed++
		} else {
			failed++
		}
	}

	// Summary line with styled counts
	passedStyle := lipgloss.NewStyle().Foreground(styles.Green).Bold(true)
	failedStyle := lipgloss.NewStyle().Foreground(styles.Red).Bold(true)
	summaryStyle := lipgloss.NewStyle().Foreground(styles.Subtext0)

	if len(r.testResults) > 0 {
		summary := fmt.Sprintf("Tests: %s passed, %s failed",
			passedStyle.Render(fmt.Sprintf("%d", passed)),
			failedStyle.Render(fmt.Sprintf("%d", failed)))
		result.WriteString(summary)
		result.WriteString("\n")
		result.WriteString(strings.Repeat("─", width))
		result.WriteString("\n")
	}

	if len(r.testResults) == 0 {
		result.WriteString(summaryStyle.Render("No test assertions in scripts."))
		result.WriteString("\n")
		result.WriteString(summaryStyle.Render("Use lc.test.assert(name, condition) in your scripts to add tests."))
		return result.String()
	}

	// Calculate how many rows we can show
	visibleRows := height - 3 // Account for summary and separator
	if visibleRows < 1 {
		visibleRows = 1
	}
	startIdx := 0
	if r.testResultsCursor >= visibleRows {
		startIdx = r.testResultsCursor - visibleRows + 1
	}

	// Render test results
	passIcon := lipgloss.NewStyle().Foreground(styles.Green).Render("✓")
	failIcon := lipgloss.NewStyle().Foreground(styles.Red).Render("✗")

	for i := startIdx; i < len(r.testResults) && i < startIdx+visibleRows; i++ {
		test := r.testResults[i]

		// Icon based on pass/fail
		icon := passIcon
		if !test.Passed {
			icon = failIcon
		}

		// Test name
		nameStyle := lipgloss.NewStyle().Foreground(styles.Text)
		name := test.Name
		maxNameWidth := width - 4 // Icon + space + padding
		if len(name) > maxNameWidth && maxNameWidth > 0 {
			name = name[:maxNameWidth-3] + "..."
		}

		// Highlight selected row
		if i == r.testResultsCursor {
			rowStyle := lipgloss.NewStyle().
				Background(styles.Surface1).
				Foreground(styles.Text)
			row := fmt.Sprintf("%s %s", icon, name)
			// Pad to full width
			if lipgloss.Width(row) < width {
				row += strings.Repeat(" ", width-lipgloss.Width(row))
			}
			result.WriteString(rowStyle.Render(row))
		} else {
			result.WriteString(fmt.Sprintf("%s %s", icon, nameStyle.Render(name)))
		}
		result.WriteString("\n")

		// Show message for failed tests on selected row
		if i == r.testResultsCursor && !test.Passed && test.Message != "" {
			messageStyle := lipgloss.NewStyle().
				Foreground(styles.Red).
				Italic(true).
				PaddingLeft(3)
			msg := test.Message
			maxMsgWidth := width - 6
			if len(msg) > maxMsgWidth && maxMsgWidth > 0 {
				msg = msg[:maxMsgWidth-3] + "..."
			}
			result.WriteString(messageStyle.Render(msg))
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
	r.isLoading = false // Clear loading state when response is received

	// Update body editor with response body and auto-format JSON
	r.bodyEditor.SetContent(body)

	// Check if content type is JSON and auto-format
	contentType := ""
	for k, v := range headers {
		if strings.ToLower(k) == "content-type" {
			contentType = strings.ToLower(v)
			break
		}
	}
	if strings.Contains(contentType, "json") || strings.HasPrefix(strings.TrimSpace(body), "{") || strings.HasPrefix(strings.TrimSpace(body), "[") {
		// Auto-format JSON for better readability
		r.bodyEditor.FormatJSON()
	}

	// Sort header and cookie keys for stable iteration
	r.headersKeys = make([]string, 0, len(headers))
	for k := range headers {
		r.headersKeys = append(r.headersKeys, k)
	}
	sort.Strings(r.headersKeys)

	r.cookiesKeys = make([]string, 0, len(cookies))
	for k := range cookies {
		r.cookiesKeys = append(r.cookiesKeys, k)
	}
	sort.Strings(r.cookiesKeys)

	// Reset cursors
	r.headersCursor = 0
	r.cookiesCursor = 0
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
	r.headersKeys = []string{}
	r.cookiesKeys = []string{}
	r.headersCursor = 0
	r.cookiesCursor = 0
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

// SetLoading sets the loading state
func (r *ResponseView) SetLoading(loading bool) {
	r.isLoading = loading
	if loading {
		r.loaderFrame = 0
	}
}

// SetTestResults sets the test assertion results from script execution
func (r *ResponseView) SetTestResults(results []api.AssertionResult) {
	r.testResults = results
	r.testResultsCursor = 0
}

// ClearTestResults clears the test results
func (r *ResponseView) ClearTestResults() {
	r.testResults = []api.AssertionResult{}
	r.testResultsCursor = 0
}

// GetTestResults returns the current test results
func (r *ResponseView) GetTestResults() []api.AssertionResult {
	return r.testResults
}

// IsLoading returns whether a request is in progress
func (r *ResponseView) IsLoading() bool {
	return r.isLoading
}

// TickLoader advances the loader animation frame
func (r *ResponseView) TickLoader() {
	if r.isLoading {
		r.loaderFrame++
	}
}

// SetSessionState applies session state to the response panel
func (r *ResponseView) SetSessionState(state session.ResponsePanelState) {
	// Set active tab
	tabIndex := 0
	switch state.ActiveTab {
	case "body":
		tabIndex = 0
	case "cookies":
		tabIndex = 1
	case "headers":
		tabIndex = 2
	case "tests":
		tabIndex = 3
	case "console":
		tabIndex = 4
	}
	r.tabs.SetActive(tabIndex)

	// Restore scroll position
	if state.ScrollPosition >= 0 {
		r.scrollOffset = state.ScrollPosition
	}
}

// GetSessionState returns the current session state for the response panel
func (r *ResponseView) GetSessionState() session.ResponsePanelState {
	state := session.ResponsePanelState{
		ScrollPosition: r.scrollOffset,
	}

	// Get active tab name
	switch r.tabs.ActiveIndex {
	case 0:
		state.ActiveTab = "body"
	case 1:
		state.ActiveTab = "cookies"
	case 2:
		state.ActiveTab = "headers"
	case 3:
		state.ActiveTab = "tests"
	case 4:
		state.ActiveTab = "console"
	default:
		state.ActiveTab = "body"
	}

	return state
}

// JumpTo jumps to a specific element by its ID (tab name, field, etc.)
func (r *ResponseView) JumpTo(elementID string) {
	// Handle tab navigation (uses tabs component)
	switch elementID {
	case "tab-body":
		r.tabs.SetActive(0)
	case "tab-cookies":
		r.tabs.SetActive(1)
	case "tab-headers":
		r.tabs.SetActive(2)
	case "tab-tests":
		r.tabs.SetActive(3)
	case "tab-console":
		r.tabs.SetActive(4)
	}
}

// GetJumpTargets returns jump targets for the response view.
// Includes tabs (Body, Cookies, Headers, Tests, Console).
func (r *ResponseView) GetJumpTargets(startRow, startCol int) []JumpTarget {
	var targets []JumpTarget

	// Tab targets - Row 1 is the tabs row (after panel header)
	tabNames := []string{"tab-body", "tab-cookies", "tab-headers", "tab-tests", "tab-console"}
	tabLabels := []string{"Body", "Cookies", "Headers", "Tests", "Console"}
	tabCol := startCol + 1 // Start after border

	// Tab separator width: " | " = 3 characters between tabs
	const tabSeparatorWidth = 3

	for i, tabID := range tabNames {
		targets = append(targets, JumpTarget{
			Panel:     ResponsePanel,
			Row:       startRow + 1, // First content row (tabs)
			Col:       tabCol,
			Index:     i,
			ElementID: tabID,
			Action:    JumpActivate,
		})
		tabCol += len(tabLabels[i]) + tabSeparatorWidth
	}

	return targets
}
