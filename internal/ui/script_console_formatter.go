package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/internal/api"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// ScriptConsoleFormatter formats script console output for display
type ScriptConsoleFormatter struct{}

// NewScriptConsoleFormatter creates a new formatter
func NewScriptConsoleFormatter() *ScriptConsoleFormatter {
	return &ScriptConsoleFormatter{}
}

// FormatEntry formats a single console log entry for display
func (f *ScriptConsoleFormatter) FormatEntry(entry api.ConsoleLogEntry, width int) string {
	// Format timestamp as HH:MM:SS
	timestamp := entry.Timestamp.Format("15:04:05")

	// Get level icon and color
	icon, color := f.getLevelStyle(entry.Level)

	// Style components
	timeStyle := lipgloss.NewStyle().Foreground(styles.Subtext0)
	iconStyle := lipgloss.NewStyle().Foreground(color)
	msgStyle := lipgloss.NewStyle().Foreground(f.getMessageColor(entry.Level))

	// Calculate available width for message
	// Format: "HH:MM:SS [icon] message"
	prefixWidth := 8 + 1 + 2 + 1 // timestamp + space + icon + space
	msgWidth := width - prefixWidth
	if msgWidth < 10 {
		msgWidth = 10
	}

	// Truncate message if needed
	message := entry.Message
	if len(message) > msgWidth {
		message = message[:msgWidth-3] + "..."
	}

	return fmt.Sprintf("%s %s %s",
		timeStyle.Render(timestamp),
		iconStyle.Render(icon),
		msgStyle.Render(message),
	)
}

// FormatEntries formats multiple console entries
func (f *ScriptConsoleFormatter) FormatEntries(entries []api.ConsoleLogEntry, width, maxLines int) string {
	if len(entries) == 0 {
		return lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Italic(true).
			Render("No console output")
	}

	var result strings.Builder

	// Show last N entries if we have too many
	startIdx := 0
	if len(entries) > maxLines && maxLines > 0 {
		startIdx = len(entries) - maxLines
		result.WriteString(lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Render(fmt.Sprintf("... (%d earlier entries hidden)\n", startIdx)))
	}

	for i := startIdx; i < len(entries); i++ {
		result.WriteString(f.FormatEntry(entries[i], width))
		if i < len(entries)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// FormatHeader returns a styled header for script console section
func (f *ScriptConsoleFormatter) FormatHeader(scriptType ScriptType, width int) string {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Mauve)

	var title string
	switch scriptType {
	case ScriptTypePreRequest:
		title = "Pre-Request Script Output"
	case ScriptTypePostResponse:
		title = "Post-Response Script Output"
	default:
		title = "Script Output"
	}

	titleLen := len(title) + 4 // For "─ " prefix and " " suffix
	lineLen := width - titleLen
	if lineLen < 0 {
		lineLen = 0
	}

	return fmt.Sprintf("%s %s",
		headerStyle.Render("─ "+title),
		strings.Repeat("─", lineLen))
}

// getLevelStyle returns icon and color for a log level
func (f *ScriptConsoleFormatter) getLevelStyle(level api.ConsoleLogLevel) (string, lipgloss.Color) {
	switch level {
	case api.LogLevelLog:
		return "●", styles.Text
	case api.LogLevelInfo:
		return "ℹ", styles.Blue
	case api.LogLevelWarn:
		return "⚠", styles.Yellow
	case api.LogLevelError:
		return "✖", styles.Red
	case api.LogLevelDebug:
		return "◌", styles.Subtext0
	default:
		return "●", styles.Text
	}
}

// getMessageColor returns text color for a log level
func (f *ScriptConsoleFormatter) getMessageColor(level api.ConsoleLogLevel) lipgloss.Color {
	switch level {
	case api.LogLevelWarn:
		return styles.Yellow
	case api.LogLevelError:
		return styles.Red
	case api.LogLevelDebug:
		return styles.Subtext0
	default:
		return styles.Text
	}
}

// FormatAssertionResults formats assertion results for display
func (f *ScriptConsoleFormatter) FormatAssertionResults(results []api.AssertionResult, width int) string {
	if len(results) == 0 {
		return lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Italic(true).
			Render("No assertions")
	}

	var buf strings.Builder

	// Count pass/fail
	passed := 0
	failed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}

	// Summary header
	summaryStyle := lipgloss.NewStyle().Bold(true)
	if failed > 0 {
		buf.WriteString(summaryStyle.Foreground(styles.Red).
			Render(fmt.Sprintf("Tests: %d passed, %d failed", passed, failed)))
	} else {
		buf.WriteString(summaryStyle.Foreground(styles.Green).
			Render(fmt.Sprintf("Tests: %d passed", passed)))
	}
	buf.WriteString("\n\n")

	// Individual results
	passIcon := lipgloss.NewStyle().Foreground(styles.Green).Render("✓")
	failIcon := lipgloss.NewStyle().Foreground(styles.Red).Render("✗")

	for _, r := range results {
		icon := passIcon
		nameStyle := lipgloss.NewStyle().Foreground(styles.Text)
		if !r.Passed {
			icon = failIcon
			nameStyle = nameStyle.Foreground(styles.Red)
		}

		// Format: ✓ Test name
		buf.WriteString(fmt.Sprintf("%s %s", icon, nameStyle.Render(r.Name)))

		// Show expected/actual for failures
		if !r.Passed && (r.Expected != nil || r.Actual != nil) {
			buf.WriteString("\n")
			expectedStyle := lipgloss.NewStyle().Foreground(styles.Green)
			actualStyle := lipgloss.NewStyle().Foreground(styles.Red)
			buf.WriteString(fmt.Sprintf("    Expected: %s\n", expectedStyle.Render(fmt.Sprintf("%v", r.Expected))))
			buf.WriteString(fmt.Sprintf("    Actual:   %s", actualStyle.Render(fmt.Sprintf("%v", r.Actual))))
		}

		// Show message if present
		if r.Message != "" && !r.Passed {
			msgStyle := lipgloss.NewStyle().Foreground(styles.Subtext0).Italic(true)
			buf.WriteString(fmt.Sprintf("\n    %s", msgStyle.Render(r.Message)))
		}

		buf.WriteString("\n")
	}

	return buf.String()
}

// FormatEnvChanges formats environment variable changes for display
func (f *ScriptConsoleFormatter) FormatEnvChanges(changes []api.EnvChange, width int) string {
	if len(changes) == 0 {
		return lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Italic(true).
			Render("No environment changes")
	}

	var buf strings.Builder

	for _, c := range changes {
		var icon, action string
		var color lipgloss.Color

		switch c.Type {
		case api.EnvChangeSet:
			if c.Previous == "" {
				icon = "+"
				action = "set"
				color = styles.Green
			} else {
				icon = "~"
				action = "updated"
				color = styles.Yellow
			}
		case api.EnvChangeUnset:
			icon = "-"
			action = "unset"
			color = styles.Red
		}

		iconStyle := lipgloss.NewStyle().Foreground(color).Bold(true)
		nameStyle := lipgloss.NewStyle().Foreground(styles.Blue)
		valueStyle := lipgloss.NewStyle().Foreground(styles.Text)

		buf.WriteString(fmt.Sprintf("%s %s %s",
			iconStyle.Render(icon),
			nameStyle.Render(c.Name),
			lipgloss.NewStyle().Foreground(styles.Subtext0).Render(action)))

		if c.Type == api.EnvChangeSet {
			buf.WriteString(fmt.Sprintf(": %s", valueStyle.Render(c.Value)))
		}

		buf.WriteString("\n")
	}

	return buf.String()
}

// FormatScriptError formats a script error for display
func (f *ScriptConsoleFormatter) FormatScriptError(err *api.ScriptErrorInfo, width int) string {
	if err == nil {
		return ""
	}

	var buf strings.Builder

	// Error type header
	typeStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Red)
	buf.WriteString(typeStyle.Render(fmt.Sprintf("%s:", err.Type)))
	buf.WriteString("\n")

	// Error message
	msgStyle := lipgloss.NewStyle().Foreground(styles.Text)
	buf.WriteString(msgStyle.Render(err.Message))

	// Line/column info if available
	if err.Line > 0 {
		locStyle := lipgloss.NewStyle().Foreground(styles.Subtext0)
		if err.Column > 0 {
			buf.WriteString(fmt.Sprintf("\n%s", locStyle.Render(fmt.Sprintf("at line %d, column %d", err.Line, err.Column))))
		} else {
			buf.WriteString(fmt.Sprintf("\n%s", locStyle.Render(fmt.Sprintf("at line %d", err.Line))))
		}
	}

	// Stack trace if available
	if err.StackTrace != "" {
		buf.WriteString("\n\n")
		stackStyle := lipgloss.NewStyle().Foreground(styles.Subtext0)
		buf.WriteString(stackStyle.Render("Stack trace:"))
		buf.WriteString("\n")
		buf.WriteString(stackStyle.Render(err.StackTrace))
	}

	return buf.String()
}
