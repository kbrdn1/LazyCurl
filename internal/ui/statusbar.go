package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// Message duration constant - 2 seconds for all action messages
const MessageDuration = 2 * time.Second

// StatusBar renders the bottom status bar with full context
type StatusBar struct {
	mode         Mode      // Current mode
	version      string    // Application version
	width        int       // Available width
	httpStatus   int       // HTTP status code (0 = no response)
	httpText     string    // HTTP status text
	httpMethod   string    // Current HTTP method
	breadcrumb   []string  // Navigation breadcrumb parts
	message      string    // Temporary status message
	messageEnd   time.Time // When to clear the message
	environment  string    // Active environment name
	hints        string    // Dynamic keybinding hints
	isFullscreen bool      // Whether fullscreen mode is active
}

// NewStatusBar creates a new status bar
func NewStatusBar(version string) *StatusBar {
	return &StatusBar{
		mode:       NormalMode,
		version:    version,
		breadcrumb: []string{},
	}
}

// SetMode updates the mode indicator
func (s *StatusBar) SetMode(mode Mode) {
	s.mode = mode
}

// SetHTTPStatus sets the HTTP status display
func (s *StatusBar) SetHTTPStatus(code int, text string) {
	s.httpStatus = code
	s.httpText = text
}

// ClearHTTPStatus clears the HTTP status display
func (s *StatusBar) ClearHTTPStatus() {
	s.httpStatus = 0
	s.httpText = ""
}

// SetMethod sets the current HTTP method display
func (s *StatusBar) SetMethod(method string) {
	s.httpMethod = method
}

// SetBreadcrumb sets the navigation breadcrumb
func (s *StatusBar) SetBreadcrumb(parts ...string) {
	s.breadcrumb = parts
}

// SetEnvironment sets the active environment name
func (s *StatusBar) SetEnvironment(name string) {
	s.environment = name
}

// SetHints sets the dynamic keybinding hints
func (s *StatusBar) SetHints(hints string) {
	s.hints = hints
}

// SetFullscreen sets the fullscreen mode indicator
func (s *StatusBar) SetFullscreen(fullscreen bool) {
	s.isFullscreen = fullscreen
}

// ShowMessage displays a temporary status message
func (s *StatusBar) ShowMessage(msg string, duration time.Duration) {
	s.message = msg
	s.messageEnd = time.Now().Add(duration)
}

// Info displays an info message (2s)
func (s *StatusBar) Info(msg string) {
	s.ShowMessage(msg, MessageDuration)
}

// Success displays a success message (2s)
func (s *StatusBar) Success(action, target string) {
	s.ShowMessage(fmt.Sprintf("%s: %s", action, target), MessageDuration)
}

// Error displays an error message (2s)
func (s *StatusBar) Error(err error) {
	s.ShowMessage(fmt.Sprintf("Error: %s", err.Error()), MessageDuration)
}

// ClearMessage clears the status message
func (s *StatusBar) ClearMessage() {
	s.message = ""
}

// View renders the status bar
func (s *StatusBar) View(width int) string {
	s.width = width

	// Clear expired message
	if s.message != "" && time.Now().After(s.messageEnd) {
		s.message = ""
	}

	// Mode badge (always first)
	modeBadge := s.mode.Color().Render(s.mode.String())
	modeWidth := lipgloss.Width(modeBadge)

	// HTTP method badge (if present, after mode)
	var methodBadge string
	methodWidth := 0
	if s.httpMethod != "" {
		methodBadge = s.renderMethodBadge()
		methodWidth = lipgloss.Width(methodBadge)
	}

	// Fullscreen badge (if active)
	var fullscreenBadge string
	fullscreenWidth := 0
	if s.isFullscreen {
		fullscreenStyle := lipgloss.NewStyle().
			Foreground(styles.Crust).
			Background(styles.Mauve).
			Bold(true).
			Padding(0, 1)
		fullscreenBadge = fullscreenStyle.Render("FULLSCREEN")
		fullscreenWidth = lipgloss.Width(fullscreenBadge)
	}

	// Environment badge (right side)
	var envBadge string
	envWidth := 0
	if s.environment != "" {
		envStyle := lipgloss.NewStyle().
			Foreground(styles.Green).
			Bold(true).
			Padding(0, 1)
		envBadge = envStyle.Render(s.environment)
		envWidth = lipgloss.Width(envBadge)
	} else {
		envStyle := lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Padding(0, 1)
		envBadge = envStyle.Render("NONE")
		envWidth = lipgloss.Width(envBadge)
	}

	// HTTP status badge (if present, right side)
	var statusBadge string
	statusWidth := 0
	if s.httpStatus > 0 {
		statusBadge = s.renderHTTPStatus()
		statusWidth = lipgloss.Width(statusBadge)
	}

	// Calculate middle content width
	usedWidth := modeWidth + methodWidth + fullscreenWidth + envWidth + statusWidth
	middleWidth := width - usedWidth
	if middleWidth < 0 {
		middleWidth = 0
	}

	// Middle content: message, breadcrumb, or hints (truncated to fit)
	var middleText string
	if s.message != "" {
		middleText = " " + s.message
	} else if len(s.breadcrumb) > 0 {
		middleText = s.formatBreadcrumbText()
	} else {
		middleText = s.getKeyboardHints()
	}

	// Truncate middle text to fit available width
	middleTextWidth := lipgloss.Width(middleText)
	if middleTextWidth > middleWidth {
		if middleWidth > 3 {
			// Truncate by runes to handle unicode properly
			runes := []rune(middleText)
			for lipgloss.Width(string(runes)) > middleWidth-3 && len(runes) > 0 {
				runes = runes[:len(runes)-1]
			}
			middleText = string(runes) + "..."
		} else if middleWidth > 0 {
			runes := []rune(middleText)
			for lipgloss.Width(string(runes)) > middleWidth && len(runes) > 0 {
				runes = runes[:len(runes)-1]
			}
			middleText = string(runes)
		} else {
			middleText = ""
		}
		middleTextWidth = lipgloss.Width(middleText)
	}

	// Pad middle content to exact width
	padding := middleWidth - middleTextWidth
	if padding > 0 {
		middleText = middleText + strings.Repeat(" ", padding)
	}

	// Style middle content (no background for transparency)
	var middleStyle lipgloss.Style
	if s.message != "" {
		middleStyle = lipgloss.NewStyle().
			Foreground(styles.Yellow).
			Bold(true)
	} else {
		middleStyle = lipgloss.NewStyle().
			Foreground(styles.Subtext0)
	}
	middleContent := middleStyle.Render(middleText)

	// Join all parts: Mode | Method | Fullscreen | Middle | Env | Status
	var parts []string
	parts = append(parts, modeBadge)
	if methodBadge != "" {
		parts = append(parts, methodBadge)
	}
	if fullscreenBadge != "" {
		parts = append(parts, fullscreenBadge)
	}
	parts = append(parts, middleContent)
	parts = append(parts, envBadge)
	if statusBadge != "" {
		parts = append(parts, statusBadge)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

// renderHTTPStatus renders the HTTP status badge with color coding
func (s *StatusBar) renderHTTPStatus() string {
	var bgColor, fgColor lipgloss.Color

	switch {
	case s.httpStatus >= 200 && s.httpStatus < 300:
		bgColor = styles.Status2xxBg
		fgColor = styles.Status2xxFg
	case s.httpStatus >= 300 && s.httpStatus < 400:
		bgColor = styles.Status3xxBg
		fgColor = styles.Status3xxFg
	case s.httpStatus >= 400 && s.httpStatus < 500:
		bgColor = styles.Status4xxBg
		fgColor = styles.Status4xxFg
	case s.httpStatus >= 500:
		bgColor = styles.Status5xxBg
		fgColor = styles.Status5xxFg
	default:
		bgColor = styles.Surface1
		fgColor = styles.Text
	}

	style := lipgloss.NewStyle().
		Background(bgColor).
		Foreground(fgColor).
		Bold(true).
		Padding(0, 1)

	text := fmt.Sprintf("%d", s.httpStatus)
	if s.httpText != "" {
		text = fmt.Sprintf("%d %s", s.httpStatus, s.httpText)
	}

	return style.Render(text)
}

// renderMethodBadge renders the HTTP method badge
func (s *StatusBar) renderMethodBadge() string {
	var bgColor, fgColor lipgloss.Color

	switch s.httpMethod {
	case "GET":
		bgColor = styles.MethodGetBg
		fgColor = styles.MethodGetFg
	case "POST":
		bgColor = styles.MethodPostBg
		fgColor = styles.MethodPostFg
	case "PUT":
		bgColor = styles.MethodPutBg
		fgColor = styles.MethodPutFg
	case "DELETE":
		bgColor = styles.MethodDeleteBg
		fgColor = styles.MethodDeleteFg
	case "PATCH":
		bgColor = styles.MethodPatchBg
		fgColor = styles.MethodPatchFg
	case "HEAD":
		bgColor = styles.MethodHeadBg
		fgColor = styles.MethodHeadFg
	case "OPTIONS":
		bgColor = styles.MethodOptionsBg
		fgColor = styles.MethodOptionsFg
	default:
		bgColor = styles.Surface1
		fgColor = styles.Text
	}

	style := lipgloss.NewStyle().
		Background(bgColor).
		Foreground(fgColor).
		Bold(true).
		Padding(0, 1)

	return style.Render(s.httpMethod)
}

// GetMode returns the current mode
func (s *StatusBar) GetMode() Mode {
	return s.mode
}

// formatBreadcrumbText returns breadcrumb as plain text with separators
func (s *StatusBar) formatBreadcrumbText() string {
	if len(s.breadcrumb) == 0 {
		return ""
	}
	return " " + strings.Join(s.breadcrumb, " › ")
}

// getKeyboardHints returns context-sensitive keyboard hints
func (s *StatusBar) getKeyboardHints() string {
	// Use dynamic hints if set
	if s.hints != "" {
		return s.hints
	}

	// Fallback to mode-based hints
	switch s.mode {
	case NormalMode:
		return " j/k:Up/Down │ h/l:Nav │ n:New │ R:Rename │ d:Delete │ ?:Help"
	case InsertMode:
		return " type:Edit │ tab:Next │ esc:Normal"
	case ViewMode:
		return " j/k:Scroll │ g/G:Top/End │ h/l:Panel │ esc:Normal"
	case CommandMode:
		return " :q:quit │ :w:save │ :ws:workspace │ esc:Cancel"
	default:
		return ""
	}
}

// StatusUpdateMsg signals a status bar update
type StatusUpdateMsg struct {
	Mode       *Mode
	HTTPStatus *int
	HTTPText   *string
	Method     *string
	Breadcrumb []string
	Message    *string
	Duration   time.Duration
}

// Helper to create a mode pointer for StatusUpdateMsg
func ModePtr(m Mode) *Mode {
	return &m
}

// IntPtr creates an int pointer
func IntPtr(i int) *int {
	return &i
}

// StringPtr creates a string pointer
func StringPtr(s string) *string {
	return &s
}
