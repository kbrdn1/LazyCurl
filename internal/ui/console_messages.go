package ui

import (
	"time"

	"github.com/kbrdn1/LazyCurl/internal/api"
)

// ResendRequestMsg signals that a request should be resent
type ResendRequestMsg struct {
	Request *api.Request // Request to resend
}

// CopyToClipboardMsg signals content should be copied to clipboard
type CopyToClipboardMsg struct {
	Content string // Text to copy
	Label   string // What was copied (for status message)
}

// ConsoleStatusMsg displays a temporary status in the console
type ConsoleStatusMsg struct {
	Message  string
	Type     StatusType
	Duration time.Duration // How long to show (default 2s)
}

// StatusType represents the type of status message
type StatusType int

const (
	StatusInfo StatusType = iota
	StatusSuccess
	StatusError
)

// SwitchToConsoleTabMsg switches ResponseView to Console tab
type SwitchToConsoleTabMsg struct{}

// SwitchToResponseTabMsg switches ResponseView to Response (Body) tab
type SwitchToResponseTabMsg struct{}
