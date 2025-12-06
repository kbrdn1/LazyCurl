package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/internal/api"
	"github.com/kbrdn1/LazyCurl/internal/config"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// ConsoleView renders the console history list within ResponseView
type ConsoleView struct {
	cursor        int     // Currently selected entry index
	scrollOffset  int     // Viewport scroll position
	expandedEntry *string // ID of expanded entry (nil = list view)
	width         int     // Available width
	height        int     // Available height
}

// NewConsoleView creates a new console view
func NewConsoleView() *ConsoleView {
	return &ConsoleView{
		cursor:        0,
		scrollOffset:  0,
		expandedEntry: nil,
	}
}

// Update handles keyboard input for the console view
func (c ConsoleView) Update(msg tea.Msg, history *api.ConsoleHistory, cfg *config.GlobalConfig) (ConsoleView, tea.Cmd) {
	if history == nil || history.IsEmpty() {
		return c, nil
	}

	maxIdx := history.Len() - 1

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If expanded, handle expanded view keys
		if c.expandedEntry != nil {
			switch msg.String() {
			case "esc", "h", "q":
				c.expandedEntry = nil
				return c, nil
			case "R":
				// Resend from expanded view
				if entry, ok := history.GetByIndex(c.cursor); ok && entry.Request != nil {
					c.expandedEntry = nil
					return c, func() tea.Msg {
						return ResendRequestMsg{Request: entry.Request}
					}
				}
			case "H":
				// Copy headers
				if entry, ok := history.GetByIndex(c.cursor); ok {
					return c, func() tea.Msg {
						return CopyToClipboardMsg{
							Content: entry.CopyHeaders(),
							Label:   "Headers",
						}
					}
				}
			case "B":
				// Copy body
				if entry, ok := history.GetByIndex(c.cursor); ok {
					return c, func() tea.Msg {
						return CopyToClipboardMsg{
							Content: entry.CopyBody(),
							Label:   "Body",
						}
					}
				}
			case "E":
				// Copy error
				if entry, ok := history.GetByIndex(c.cursor); ok {
					if entry.HasError() {
						return c, func() tea.Msg {
							return CopyToClipboardMsg{
								Content: entry.CopyError(),
								Label:   "Error",
							}
						}
					}
					return c, func() tea.Msg {
						return ConsoleStatusMsg{
							Message: "No error to copy",
							Type:    StatusInfo,
						}
					}
				}
			case "A":
				// Copy all
				if entry, ok := history.GetByIndex(c.cursor); ok {
					return c, func() tea.Msg {
						return CopyToClipboardMsg{
							Content: entry.CopyAll(),
							Label:   "Request & Response",
						}
					}
				}
			}
			return c, nil
		}

		// List view navigation
		switch msg.String() {
		case "j", "down":
			if c.cursor < maxIdx {
				c.cursor++
			}
		case "k", "up":
			if c.cursor > 0 {
				c.cursor--
			}
		case "g":
			c.cursor = 0
			c.scrollOffset = 0
		case "G":
			c.cursor = maxIdx
		case "enter", "l":
			// Expand selected entry
			if entry, ok := history.GetByIndex(c.cursor); ok {
				c.expandedEntry = &entry.ID
			}
		case "R":
			// Resend selected request
			if entry, ok := history.GetByIndex(c.cursor); ok && entry.Request != nil {
				return c, func() tea.Msg {
					return ResendRequestMsg{Request: entry.Request}
				}
			}
		case "U":
			// Copy URL
			if entry, ok := history.GetByIndex(c.cursor); ok && entry.Request != nil {
				return c, func() tea.Msg {
					return CopyToClipboardMsg{
						Content: entry.Request.URL,
						Label:   "URL",
					}
				}
			}
		case "H":
			// Copy headers
			if entry, ok := history.GetByIndex(c.cursor); ok {
				return c, func() tea.Msg {
					return CopyToClipboardMsg{
						Content: entry.CopyHeaders(),
						Label:   "Headers",
					}
				}
			}
		case "B":
			// Copy body
			if entry, ok := history.GetByIndex(c.cursor); ok {
				return c, func() tea.Msg {
					return CopyToClipboardMsg{
						Content: entry.CopyBody(),
						Label:   "Body",
					}
				}
			}
		case "E":
			// Copy error
			if entry, ok := history.GetByIndex(c.cursor); ok {
				if entry.HasError() {
					return c, func() tea.Msg {
						return CopyToClipboardMsg{
							Content: entry.CopyError(),
							Label:   "Error",
						}
					}
				}
				return c, func() tea.Msg {
					return ConsoleStatusMsg{
						Message: "No error to copy",
						Type:    StatusInfo,
					}
				}
			}
		case "C":
			// Copy cookies
			if entry, ok := history.GetByIndex(c.cursor); ok {
				cookies := entry.CopyCookies()
				if cookies != "" {
					return c, func() tea.Msg {
						return CopyToClipboardMsg{
							Content: cookies,
							Label:   "Cookies",
						}
					}
				}
				return c, func() tea.Msg {
					return ConsoleStatusMsg{
						Message: "No cookies to copy",
						Type:    StatusInfo,
					}
				}
			}
		case "I":
			// Copy info
			if entry, ok := history.GetByIndex(c.cursor); ok {
				return c, func() tea.Msg {
					return CopyToClipboardMsg{
						Content: entry.CopyInfo(),
						Label:   "Info",
					}
				}
			}
		case "A":
			// Copy all
			if entry, ok := history.GetByIndex(c.cursor); ok {
				return c, func() tea.Msg {
					return CopyToClipboardMsg{
						Content: entry.CopyAll(),
						Label:   "Request & Response",
					}
				}
			}
		}
	}

	return c, nil
}

// View renders the console content
func (c ConsoleView) View(width, height int, history *api.ConsoleHistory, active bool) string {
	c.width = width
	c.height = height

	if history == nil || history.IsEmpty() {
		return lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Render("No requests yet. Send a request to see it here.")
	}

	// Check if we're in expanded view
	if c.expandedEntry != nil {
		return c.renderExpandedView(width, height, history)
	}

	return c.renderListView(width, height, history)
}

// renderListView renders the console list
func (c *ConsoleView) renderListView(width, height int, history *api.ConsoleHistory) string {
	var result strings.Builder
	entries := history.GetReversed()

	// Column widths
	const timeCol = 8     // "HH:MM:SS"
	const statusCol = 7   // " 200 " + space
	const methodCol = 6  // " DELETE " + space
	const durSizeCol = 20 // right side for dur + size
	urlWidth := width - timeCol - statusCol - methodCol - durSizeCol
	if urlWidth < 10 {
		urlWidth = 10
	}

	// Render header with consistent spacing
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Subtext0)
	durHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Teal)
	sizeHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Peach)

	header := fmt.Sprintf("%-8s%-7s%-7s%-*s",
		"Time", "Status", "Method", urlWidth, "URL")
	result.WriteString(headerStyle.Render(header))
	result.WriteString(durHeaderStyle.Render("    ◷ Dur"))
	result.WriteString(sizeHeaderStyle.Render(" ◆ Size"))
	result.WriteString("\n")
	result.WriteString(strings.Repeat("─", width))
	result.WriteString("\n")

	// Calculate visible range
	visibleRows := height - 2 // Account for header and separator
	if visibleRows < 1 {
		visibleRows = 1
	}

	// Adjust scroll offset to keep cursor visible
	if c.cursor >= c.scrollOffset+visibleRows {
		c.scrollOffset = c.cursor - visibleRows + 1
	}
	if c.cursor < c.scrollOffset {
		c.scrollOffset = c.cursor
	}

	endIdx := c.scrollOffset + visibleRows
	if endIdx > len(entries) {
		endIdx = len(entries)
	}

	// Render entries
	for i := c.scrollOffset; i < endIdx; i++ {
		entry := entries[i]
		row := c.renderEntryRow(&entry, width, i == c.cursor)
		result.WriteString(row)
		result.WriteString("\n")
	}

	return result.String()
}

// renderEntryRow renders a single console entry row
func (c *ConsoleView) renderEntryRow(entry *api.ConsoleEntry, width int, selected bool) string {
	// Get status code and colors - same badge style as StatusBadge and Collections
	statusCode := entry.GetStatusCode()
	var statusStr string
	var statusBg, statusFg lipgloss.Color

	if entry.HasError() {
		statusStr = "Err"
		statusBg = styles.Red
		statusFg = styles.Base
	} else {
		statusStr = fmt.Sprintf("%d", statusCode)
		switch {
		case statusCode >= 200 && statusCode < 300:
			statusBg = styles.Status2xxBg
			statusFg = styles.Status2xxFg
		case statusCode >= 300 && statusCode < 400:
			statusBg = styles.Status3xxBg
			statusFg = styles.Status3xxFg
		case statusCode >= 400 && statusCode < 500:
			statusBg = styles.Status4xxBg
			statusFg = styles.Status4xxFg
		case statusCode >= 500:
			statusBg = styles.Status5xxBg
			statusFg = styles.Status5xxFg
		default:
			statusBg = styles.Surface1
			statusFg = styles.Text
		}
	}

	// Method badge with same style as Collections panel
	methodStr := string(entry.Request.Method)
	methodBg, methodFg := c.getMethodColors(methodStr)

	// Column widths (same as header)
	const timeCol = 8
	const statusCol = 7
	const methodCol = 6
	const durSizeCol = 20
	urlWidth := width - timeCol - statusCol - methodCol - durSizeCol
	if urlWidth < 12 {
		urlWidth = 12
	}

	// URL (truncated if needed)
	url := entry.Request.URL
	if len(url) > urlWidth {
		url = url[:urlWidth]
	}

	// Time in gray (no icon)
	timeStr := entry.FormatTimestamp()
	timeStyle := lipgloss.NewStyle().Foreground(styles.Subtext0)
	timeText := timeStyle.Render(timeStr)

	// Duration and size with icons and colors
	dur := entry.FormatDuration()
	size := entry.FormatSize()
	durStyle := lipgloss.NewStyle().Foreground(styles.Teal)
	sizeStyle := lipgloss.NewStyle().Foreground(styles.Peach)
	durText := durStyle.Render(fmt.Sprintf("◷ %s", dur))
	sizeText := sizeStyle.Render(fmt.Sprintf("◆ %s", size))

	// Apply badge styles (background + foreground like Collections panel)
	statusStyle := lipgloss.NewStyle().
		Background(statusBg).
		Foreground(statusFg).
		Padding(0, 1)
	methodStyle := lipgloss.NewStyle().
		Background(methodBg).
		Foreground(methodFg).
		Padding(0, 1)

	// Render badges
	statusBadge := statusStyle.Render(statusStr)
	methodBadge := methodStyle.Render(methodStr)

	// Calculate padding for badges to align columns
	statusWidth := lipgloss.Width(statusBadge)
	methodWidth := lipgloss.Width(methodBadge)
	statusPad := statusCol - statusWidth + 1
	methodPad := methodCol - methodWidth
	if statusPad < 1 {
		statusPad = 1
	}
	if methodPad < 1 {
		methodPad = 1
	}

	// Build row with column alignment matching header
	// Column widths: timeCol=8, statusCol=7, methodCol=10, urlWidth, durSizeCol=20
	statusPadding := statusCol - statusWidth
	if statusPadding < 0 {
		statusPadding = 0
	}
	methodPadding := methodCol - methodWidth
	if methodPadding < 0 {
		methodPadding = 0
	}

	var rowBuilder strings.Builder
	rowBuilder.WriteString(timeText)
	rowBuilder.WriteString(strings.Repeat(" ", methodPadding))
	rowBuilder.WriteString(statusBadge)
	rowBuilder.WriteString(methodBadge)
	rowBuilder.WriteString(strings.Repeat(" ", statusPadding))
	rowBuilder.WriteString(fmt.Sprintf("%-*s", urlWidth, url))
	rowBuilder.WriteString(fmt.Sprintf("   %s %s", durText, sizeText))
	row := rowBuilder.String()

	// Apply selection style
	if selected {
		rowStyle := lipgloss.NewStyle().
			Background(styles.Surface1).
			Foreground(styles.Text)
		// Pad to full width
		if len(row) < width {
			row += strings.Repeat(" ", width-lipgloss.Width(row))
		}
		return rowStyle.Render(row)
	}

	return row
}

// renderExpandedView renders the expanded entry details
func (c *ConsoleView) renderExpandedView(width, height int, history *api.ConsoleHistory) string {
	entry, ok := history.GetByIndex(c.cursor)
	if !ok {
		return "Entry not found"
	}

	var result strings.Builder

	// Request section
	requestStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Blue)
	result.WriteString(requestStyle.Render("─ Request "))
	result.WriteString(strings.Repeat("─", width-11))
	result.WriteString("\n")

	if entry.Request != nil {
		// Method badge with same style as Collections panel
		methodBg, methodFg := c.getMethodColors(string(entry.Request.Method))
		methodStyle := lipgloss.NewStyle().
			Background(methodBg).
			Foreground(methodFg).
			Padding(0, 1)
		result.WriteString(methodStyle.Render(string(entry.Request.Method)))
		result.WriteString(" ")
		result.WriteString(entry.Request.URL)
		result.WriteString("\n\n")

		if len(entry.Request.Headers) > 0 {
			headerLabelStyle := lipgloss.NewStyle().Foreground(styles.Subtext1)
			result.WriteString(headerLabelStyle.Render("Headers:"))
			result.WriteString("\n")
			for key, value := range entry.Request.Headers {
				result.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
			}
			result.WriteString("\n")
		}

		if entry.Request.Body != nil {
			headerLabelStyle := lipgloss.NewStyle().Foreground(styles.Subtext1)
			result.WriteString(headerLabelStyle.Render("Body:"))
			result.WriteString("\n")
			result.WriteString(fmt.Sprintf("  %v\n", entry.Request.Body))
		}
	}

	// Response section
	result.WriteString("\n")
	responseStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Green)
	result.WriteString(responseStyle.Render("─ Response "))
	result.WriteString(strings.Repeat("─", width-12))
	result.WriteString("\n")

	if entry.Error != nil {
		errorStyle := lipgloss.NewStyle().Foreground(styles.Red)
		result.WriteString(errorStyle.Render("Error: "))
		result.WriteString(entry.Error.Error())
		result.WriteString("\n")
	} else if entry.Response != nil {
		// Use StatusBadge for consistent styling with response_view
		statusBadge := NewStatusBadge(entry.Response.StatusCode)
		result.WriteString(statusBadge.Render())

		// Time and size with same icons as response_view
		timeStyle := lipgloss.NewStyle().Foreground(styles.Teal)
		sizeStyle := lipgloss.NewStyle().Foreground(styles.Peach)
		timeIcon := "◷"
		sizeIcon := "◆"
		result.WriteString("  ")
		result.WriteString(timeStyle.Render(fmt.Sprintf("%s %s", timeIcon, entry.FormatDuration())))
		result.WriteString("  ")
		result.WriteString(sizeStyle.Render(fmt.Sprintf("%s %s", sizeIcon, entry.FormatSize())))
		result.WriteString("\n\n")

		if len(entry.Response.Headers) > 0 {
			headerLabelStyle := lipgloss.NewStyle().Foreground(styles.Subtext1)
			result.WriteString(headerLabelStyle.Render("Headers:"))
			result.WriteString("\n")
			headerCount := 0
			for key, values := range entry.Response.Headers {
				if headerCount >= 10 {
					result.WriteString("  ...\n")
					break
				}
				for _, value := range values {
					result.WriteString(fmt.Sprintf("  %s: %s\n", key, value))
					headerCount++
				}
			}
			result.WriteString("\n")
		}

		if entry.Response.Body != "" {
			headerLabelStyle := lipgloss.NewStyle().Foreground(styles.Subtext1)
			result.WriteString(headerLabelStyle.Render("Body:"))
			result.WriteString("\n")
			body := entry.Response.Body
			// Truncate body preview
			maxBodyLen := 500
			if len(body) > maxBodyLen {
				body = body[:maxBodyLen] + "..."
			}
			result.WriteString(fmt.Sprintf("  %s\n", body))
		}
	}

	// Action hints
	result.WriteString("\n")
	hintStyle := lipgloss.NewStyle().Foreground(styles.Subtext0)
	result.WriteString(hintStyle.Render("[R]esend  [H]eaders  [B]ody  [E]rror  [A]ll  [Esc]Back"))

	return result.String()
}

// getStatusDisplay returns the icon and color for a status
func (c *ConsoleView) getStatusDisplay(status api.ConsoleEntryStatus) (string, lipgloss.Color) {
	switch status {
	case api.StatusSuccess:
		return "✓", styles.Status2xxBg
	case api.StatusRedirect:
		return "→", styles.Status3xxBg
	case api.StatusClientError:
		return "✗", styles.Status4xxBg
	case api.StatusServerError:
		return "✗", styles.Status5xxBg
	case api.StatusNetworkError:
		return "⚠", styles.Red
	default:
		return "?", styles.Subtext0
	}
}

// getMethodColors returns the background and foreground colors for an HTTP method
func (c *ConsoleView) getMethodColors(method string) (lipgloss.Color, lipgloss.Color) {
	switch method {
	case "GET":
		return styles.MethodGetBg, styles.MethodGetFg
	case "POST":
		return styles.MethodPostBg, styles.MethodPostFg
	case "PUT":
		return styles.MethodPutBg, styles.MethodPutFg
	case "PATCH":
		return styles.MethodPatchBg, styles.MethodPatchFg
	case "DELETE":
		return styles.MethodDeleteBg, styles.MethodDeleteFg
	case "HEAD":
		return styles.MethodHeadBg, styles.MethodHeadFg
	case "OPTIONS":
		return styles.MethodOptionsBg, styles.MethodOptionsFg
	default:
		return styles.Surface1, styles.Text
	}
}

// GetSelectedEntry returns the currently selected entry
func (c *ConsoleView) GetSelectedEntry(history *api.ConsoleHistory) *api.ConsoleEntry {
	if history == nil || history.IsEmpty() {
		return nil
	}
	entry, _ := history.GetByIndex(c.cursor)
	return entry
}

// IsExpanded returns true if viewing entry details
func (c *ConsoleView) IsExpanded() bool {
	return c.expandedEntry != nil
}

// SetDimensions updates the view dimensions
func (c *ConsoleView) SetDimensions(width, height int) {
	c.width = width
	c.height = height
}

// Reset resets the view state
func (c *ConsoleView) Reset() {
	c.cursor = 0
	c.scrollOffset = 0
	c.expandedEntry = nil
}
