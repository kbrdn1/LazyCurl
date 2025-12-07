// Package session provides session persistence for LazyCurl.
// It manages saving and restoring application state to .lazycurl/session.yml.
package session

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// SessionVersion is the current session file format version
	SessionVersion = 1
	// SessionFileName is the name of the session file
	SessionFileName = "session.yml"
	// SessionDir is the directory containing session files
	SessionDir = ".lazycurl"
)

// Session represents the complete application state at a point in time.
type Session struct {
	Version           int         `yaml:"version"`
	LastUpdated       time.Time   `yaml:"last_updated"`
	ActivePanel       string      `yaml:"active_panel"`
	ActiveCollection  string      `yaml:"active_collection,omitempty"`
	ActiveRequest     string      `yaml:"active_request,omitempty"`
	ActiveEnvironment string      `yaml:"active_environment,omitempty"`
	Panels            PanelsState `yaml:"panels"`
}

// PanelsState contains state for all panels.
type PanelsState struct {
	Collections CollectionsPanelState `yaml:"collections"`
	Request     RequestPanelState     `yaml:"request"`
	Response    ResponsePanelState    `yaml:"response"`
}

// CollectionsPanelState represents collections panel state.
type CollectionsPanelState struct {
	ExpandedFolders []string `yaml:"expanded_folders,omitempty"`
	ScrollPosition  int      `yaml:"scroll_position"`
	SelectedIndex   int      `yaml:"selected_index"`
}

// RequestPanelState represents request panel state.
type RequestPanelState struct {
	ActiveTab  string          `yaml:"active_tab"`
	URLCursor  int             `yaml:"url_cursor,omitempty"`
	BodyCursor *CursorPosition `yaml:"body_cursor,omitempty"`
}

// ResponsePanelState represents response panel state.
type ResponsePanelState struct {
	ActiveTab      string `yaml:"active_tab"`
	ScrollPosition int    `yaml:"scroll_position"`
}

// CursorPosition represents cursor in multi-line editor.
type CursorPosition struct {
	Line   int `yaml:"line"`
	Column int `yaml:"column"`
}

// DefaultSession returns a new Session with sensible default values.
func DefaultSession() *Session {
	return &Session{
		Version:     SessionVersion,
		LastUpdated: time.Now(),
		ActivePanel: "collections",
		Panels: PanelsState{
			Collections: CollectionsPanelState{
				ScrollPosition: 0,
				SelectedIndex:  0,
			},
			Request: RequestPanelState{
				ActiveTab: "params",
			},
			Response: ResponsePanelState{
				ActiveTab:      "body",
				ScrollPosition: 0,
			},
		},
	}
}

// GetSessionPath returns the full path to the session file.
func GetSessionPath(workspacePath string) string {
	return filepath.Join(workspacePath, SessionDir, SessionFileName)
}

// LoadSession loads session from .lazycurl/session.yml in the given workspace.
// Returns DefaultSession if file is missing, invalid, or has unsupported version.
func LoadSession(workspacePath string) (*Session, error) {
	sessionPath := GetSessionPath(workspacePath)

	data, err := os.ReadFile(sessionPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File not found - return default session, no error
			return DefaultSession(), nil
		}
		// Other I/O error - return default session, no error (graceful degradation)
		return DefaultSession(), nil
	}

	var session Session
	if err := yaml.Unmarshal(data, &session); err != nil {
		// Parse error - return default session
		return DefaultSession(), nil
	}

	// Check version compatibility
	if session.Version < 1 || session.Version > SessionVersion {
		// Unsupported version - return default session
		return DefaultSession(), nil
	}

	return &session, nil
}

// Save saves session to .lazycurl/session.yml in the given workspace.
// Uses atomic write (temp file + rename) for safety.
func (s *Session) Save(workspacePath string) error {
	sessionDir := filepath.Join(workspacePath, SessionDir)
	sessionPath := GetSessionPath(workspacePath)

	// Create directory if needed
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return err
	}

	// Update timestamp
	s.LastUpdated = time.Now()

	// Marshal to YAML
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	// Write to temp file
	tempPath := sessionPath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}

	// Atomic rename
	if err := os.Rename(tempPath, sessionPath); err != nil {
		// Clean up temp file on failure
		os.Remove(tempPath)
		return err
	}

	return nil
}

// Validate validates session references and clears invalid ones.
// Returns the same session with invalid references cleared.
func (s *Session) Validate(workspacePath string) *Session {
	// Validate ActiveCollection
	if s.ActiveCollection != "" {
		collectionPath := filepath.Join(workspacePath, SessionDir, "collections", s.ActiveCollection)
		if _, err := os.Stat(collectionPath); os.IsNotExist(err) {
			s.ActiveCollection = ""
			s.ActiveRequest = "" // Clear request if collection is invalid
		}
	}

	// Note: ActiveRequest is validated by ID when loading - no need to clear it here
	// The request ID is sufficient to find the request in any collection

	// Note: ActiveEnvironment is validated by name when setting it on EnvironmentsView
	// The display name doesn't match the filename, so we skip file validation here

	// Validate panel values
	if s.ActivePanel != "collections" && s.ActivePanel != "request" && s.ActivePanel != "response" {
		s.ActivePanel = "collections"
	}

	// Ensure scroll positions are non-negative
	if s.Panels.Collections.ScrollPosition < 0 {
		s.Panels.Collections.ScrollPosition = 0
	}
	if s.Panels.Collections.SelectedIndex < 0 {
		s.Panels.Collections.SelectedIndex = 0
	}
	if s.Panels.Response.ScrollPosition < 0 {
		s.Panels.Response.ScrollPosition = 0
	}
	if s.Panels.Request.URLCursor < 0 {
		s.Panels.Request.URLCursor = 0
	}

	// Validate tab values
	validRequestTabs := map[string]bool{"params": true, "headers": true, "body": true, "auth": true, "scripts": true}
	if !validRequestTabs[s.Panels.Request.ActiveTab] {
		s.Panels.Request.ActiveTab = "params"
	}

	validResponseTabs := map[string]bool{"body": true, "headers": true, "cookies": true, "console": true}
	if !validResponseTabs[s.Panels.Response.ActiveTab] {
		s.Panels.Response.ActiveTab = "body"
	}

	return s
}
