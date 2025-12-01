package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// Mode represents the current interaction context for the TUI
type Mode int

const (
	NormalMode  Mode = iota // Default navigation mode
	ViewMode                // Read-only viewing mode
	CommandMode             // Command input mode
	InsertMode              // Text input mode
)

// String returns the display name for the mode
func (m Mode) String() string {
	switch m {
	case NormalMode:
		return "NORMAL"
	case ViewMode:
		return "VIEW"
	case CommandMode:
		return "COMMAND"
	case InsertMode:
		return "INSERT"
	default:
		return "NORMAL"
	}
}

// Color returns the Lipgloss style for the mode indicator
func (m Mode) Color() lipgloss.Style {
	base := lipgloss.NewStyle().Padding(0, 1).Bold(true)

	switch m {
	case NormalMode:
		return base.Background(styles.ModeNormalBg).Foreground(styles.ModeNormalFg)
	case ViewMode:
		return base.Background(styles.ModeViewBg).Foreground(styles.ModeViewFg)
	case CommandMode:
		return base.Background(styles.ModeCommandBg).Foreground(styles.ModeCommandFg)
	case InsertMode:
		return base.Background(styles.ModeInsertBg).Foreground(styles.ModeInsertFg)
	default:
		return base.Background(styles.ModeNormalBg).Foreground(styles.ModeNormalFg)
	}
}

// AllowsInput returns true if mode accepts text input
func (m Mode) AllowsInput() bool {
	return m == InsertMode || m == CommandMode
}

// AllowsNavigation returns true if mode allows panel navigation
func (m Mode) AllowsNavigation() bool {
	return m == NormalMode || m == ViewMode
}

// ModeChangeMsg signals a mode transition
type ModeChangeMsg struct {
	From Mode
	To   Mode
}

// NewModeChangeMsg creates a mode change message
func NewModeChangeMsg(from, to Mode) ModeChangeMsg {
	return ModeChangeMsg{From: from, To: to}
}
