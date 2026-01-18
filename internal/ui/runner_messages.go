package ui

import (
	"github.com/kbrdn1/LazyCurl/internal/api"
)

// ============================================================================
// RUNNER MESSAGES
// ============================================================================

// RunnerStartMsg signals the runner should start executing.
type RunnerStartMsg struct {
	Collection *api.CollectionFile
	FolderPath []string
	Config     api.RunConfig
}

// RunnerStartedMsg signals the run has begun.
type RunnerStartedMsg struct {
	Session  *api.RunSession
	Requests []*api.CollectionRequest
}

// RunnerProgressMsg signals progress during execution.
type RunnerProgressMsg struct {
	CurrentIndex  int
	TotalRequests int
	CurrentName   string
	Status        string // "executing", "waiting", "scripting"
}

// RunnerRequestCompleteMsg signals a single request has completed.
type RunnerRequestCompleteMsg struct {
	Result  api.RequestResult
	Session *api.RunSession
}

// RunnerCompleteMsg signals the entire run has finished.
type RunnerCompleteMsg struct {
	Session *api.RunSession
	Report  *api.RunReport
}

// RunnerErrorMsg signals a fatal runner error.
type RunnerErrorMsg struct {
	Error error
}

// RunnerCancelMsg signals the user wants to cancel.
type RunnerCancelMsg struct{}

// RunnerCancelledMsg signals the run was canceled.
type RunnerCancelledMsg struct {
	Session *api.RunSession
	Report  *api.RunReport
}

// RunnerExportMsg signals the user wants to export results.
type RunnerExportMsg struct{}

// RunnerExportedMsg signals the export completed.
type RunnerExportedMsg struct {
	FilePath string
	Error    error
}

// RunnerShowModalMsg signals the modal should be shown.
type RunnerShowModalMsg struct{}

// RunnerHideModalMsg signals the modal should be hidden.
type RunnerHideModalMsg struct{}

// ============================================================================
// TICK MESSAGES
// ============================================================================

// RunnerTickMsg is sent periodically during execution for animation.
type RunnerTickMsg struct{}

// ============================================================================
// EXECUTION MESSAGES
// ============================================================================

// RunnerExecuteNextMsg signals that the next request should be executed.
type RunnerExecuteNextMsg struct{}

// RunnerDelayCompleteMsg signals that the delay between requests is complete.
type RunnerDelayCompleteMsg struct{}
