package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultSession(t *testing.T) {
	tests := []struct {
		name string
		want func(*Session) bool
	}{
		{
			name: "returns non-nil session",
			want: func(s *Session) bool { return s != nil },
		},
		{
			name: "version is current",
			want: func(s *Session) bool { return s.Version == SessionVersion },
		},
		{
			name: "active panel is collections",
			want: func(s *Session) bool { return s.ActivePanel == "collections" },
		},
		{
			name: "request tab is params",
			want: func(s *Session) bool { return s.Panels.Request.ActiveTab == "params" },
		},
		{
			name: "response tab is body",
			want: func(s *Session) bool { return s.Panels.Response.ActiveTab == "body" },
		},
		{
			name: "collections scroll position is 0",
			want: func(s *Session) bool { return s.Panels.Collections.ScrollPosition == 0 },
		},
		{
			name: "last updated is recent",
			want: func(s *Session) bool {
				return time.Since(s.LastUpdated) < time.Second
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultSession()
			if !tt.want(got) {
				t.Errorf("DefaultSession() check failed for %s", tt.name)
			}
		})
	}
}

func TestGetSessionPath(t *testing.T) {
	tests := []struct {
		name          string
		workspacePath string
		want          string
	}{
		{
			name:          "simple path",
			workspacePath: "/home/user/project",
			want:          "/home/user/project/.lazycurl/session.yml",
		},
		{
			name:          "path with trailing slash",
			workspacePath: "/home/user/project/",
			want:          "/home/user/project/.lazycurl/session.yml",
		},
		{
			name:          "current directory",
			workspacePath: ".",
			want:          ".lazycurl/session.yml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSessionPath(tt.workspacePath)
			// Use filepath.Clean for comparison to handle trailing slashes
			if filepath.Clean(got) != filepath.Clean(tt.want) {
				t.Errorf("GetSessionPath(%q) = %q, want %q", tt.workspacePath, got, tt.want)
			}
		})
	}
}

func TestLoadSession(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(dir string) error
		wantDefault bool
		wantErr     bool
	}{
		{
			name: "missing file returns default",
			setup: func(dir string) error {
				// No setup - file doesn't exist
				return nil
			},
			wantDefault: true,
			wantErr:     false,
		},
		{
			name: "valid session file",
			setup: func(dir string) error {
				sessionDir := filepath.Join(dir, SessionDir)
				if err := os.MkdirAll(sessionDir, 0755); err != nil {
					return err
				}
				content := `version: 1
last_updated: "2025-01-01T00:00:00Z"
active_panel: request
active_collection: "api.json"
panels:
  collections:
    scroll_position: 5
    selected_index: 3
  request:
    active_tab: body
  response:
    active_tab: headers
    scroll_position: 10
`
				return os.WriteFile(filepath.Join(sessionDir, SessionFileName), []byte(content), 0644)
			},
			wantDefault: false,
			wantErr:     false,
		},
		{
			name: "invalid YAML returns default",
			setup: func(dir string) error {
				sessionDir := filepath.Join(dir, SessionDir)
				if err := os.MkdirAll(sessionDir, 0755); err != nil {
					return err
				}
				content := `invalid: yaml: content: [[[`
				return os.WriteFile(filepath.Join(sessionDir, SessionFileName), []byte(content), 0644)
			},
			wantDefault: true,
			wantErr:     false,
		},
		{
			name: "unsupported version returns default",
			setup: func(dir string) error {
				sessionDir := filepath.Join(dir, SessionDir)
				if err := os.MkdirAll(sessionDir, 0755); err != nil {
					return err
				}
				content := `version: 999
active_panel: collections
`
				return os.WriteFile(filepath.Join(sessionDir, SessionFileName), []byte(content), 0644)
			},
			wantDefault: true,
			wantErr:     false,
		},
		{
			name: "version 0 returns default",
			setup: func(dir string) error {
				sessionDir := filepath.Join(dir, SessionDir)
				if err := os.MkdirAll(sessionDir, 0755); err != nil {
					return err
				}
				content := `version: 0
active_panel: collections
`
				return os.WriteFile(filepath.Join(sessionDir, SessionFileName), []byte(content), 0644)
			},
			wantDefault: true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			dir, err := os.MkdirTemp("", "session_test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(dir)

			// Setup test
			if err := tt.setup(dir); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			// Load session
			got, err := LoadSession(dir)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got == nil {
				t.Error("LoadSession() returned nil session")
				return
			}

			if tt.wantDefault {
				if got.ActivePanel != "collections" {
					t.Errorf("Expected default session with collections panel, got %s", got.ActivePanel)
				}
			} else {
				if got.ActivePanel != "request" {
					t.Errorf("Expected loaded session with request panel, got %s", got.ActivePanel)
				}
				if got.ActiveCollection != "api.json" {
					t.Errorf("Expected active_collection 'api.json', got %s", got.ActiveCollection)
				}
			}
		})
	}
}

func TestSessionSave(t *testing.T) {
	tests := []struct {
		name    string
		session *Session
		wantErr bool
	}{
		{
			name:    "save default session",
			session: DefaultSession(),
			wantErr: false,
		},
		{
			name: "save session with all fields",
			session: &Session{
				Version:           1,
				LastUpdated:       time.Now(),
				ActivePanel:       "request",
				ActiveCollection:  "my-api.json",
				ActiveRequest:     "req_123",
				ActiveEnvironment: "development",
				Panels: PanelsState{
					Collections: CollectionsPanelState{
						ExpandedFolders: []string{"Users", "Products"},
						ScrollPosition:  10,
						SelectedIndex:   5,
					},
					Request: RequestPanelState{
						ActiveTab: "body",
						URLCursor: 25,
						BodyCursor: &CursorPosition{
							Line:   10,
							Column: 5,
						},
					},
					Response: ResponsePanelState{
						ActiveTab:      "headers",
						ScrollPosition: 50,
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			dir, err := os.MkdirTemp("", "session_test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(dir)

			// Save session
			err = tt.session.Save(dir)

			if (err != nil) != tt.wantErr {
				t.Errorf("Session.Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify file exists
				sessionPath := GetSessionPath(dir)
				if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
					t.Error("Session file was not created")
					return
				}

				// Verify content by loading
				loaded, err := LoadSession(dir)
				if err != nil {
					t.Errorf("Failed to load saved session: %v", err)
					return
				}

				if loaded.ActivePanel != tt.session.ActivePanel {
					t.Errorf("Loaded ActivePanel = %s, want %s", loaded.ActivePanel, tt.session.ActivePanel)
				}
				if loaded.ActiveCollection != tt.session.ActiveCollection {
					t.Errorf("Loaded ActiveCollection = %s, want %s", loaded.ActiveCollection, tt.session.ActiveCollection)
				}
			}
		})
	}
}

func TestSessionRoundTrip(t *testing.T) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "session_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create session with specific values
	original := &Session{
		Version:           1,
		LastUpdated:       time.Now(),
		ActivePanel:       "response",
		ActiveCollection:  "test-api.json",
		ActiveRequest:     "req_456",
		ActiveEnvironment: "staging",
		Panels: PanelsState{
			Collections: CollectionsPanelState{
				ExpandedFolders: []string{"Auth", "Users"},
				ScrollPosition:  15,
				SelectedIndex:   7,
			},
			Request: RequestPanelState{
				ActiveTab: "headers",
				URLCursor: 30,
			},
			Response: ResponsePanelState{
				ActiveTab:      "cookies",
				ScrollPosition: 25,
			},
		},
	}

	// Save
	if err := original.Save(dir); err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Load
	loaded, err := LoadSession(dir)
	if err != nil {
		t.Fatalf("Failed to load session: %v", err)
	}

	// Compare
	if loaded.ActivePanel != original.ActivePanel {
		t.Errorf("ActivePanel: got %s, want %s", loaded.ActivePanel, original.ActivePanel)
	}
	if loaded.ActiveCollection != original.ActiveCollection {
		t.Errorf("ActiveCollection: got %s, want %s", loaded.ActiveCollection, original.ActiveCollection)
	}
	if loaded.ActiveRequest != original.ActiveRequest {
		t.Errorf("ActiveRequest: got %s, want %s", loaded.ActiveRequest, original.ActiveRequest)
	}
	if loaded.ActiveEnvironment != original.ActiveEnvironment {
		t.Errorf("ActiveEnvironment: got %s, want %s", loaded.ActiveEnvironment, original.ActiveEnvironment)
	}
	if loaded.Panels.Collections.ScrollPosition != original.Panels.Collections.ScrollPosition {
		t.Errorf("Collections.ScrollPosition: got %d, want %d",
			loaded.Panels.Collections.ScrollPosition, original.Panels.Collections.ScrollPosition)
	}
	if loaded.Panels.Request.ActiveTab != original.Panels.Request.ActiveTab {
		t.Errorf("Request.ActiveTab: got %s, want %s",
			loaded.Panels.Request.ActiveTab, original.Panels.Request.ActiveTab)
	}
	if loaded.Panels.Response.ActiveTab != original.Panels.Response.ActiveTab {
		t.Errorf("Response.ActiveTab: got %s, want %s",
			loaded.Panels.Response.ActiveTab, original.Panels.Response.ActiveTab)
	}
}

func TestSessionValidate(t *testing.T) {
	tests := []struct {
		name    string
		session *Session
		check   func(*Session) bool
	}{
		{
			name: "invalid active panel is corrected",
			session: &Session{
				Version:     1,
				ActivePanel: "invalid",
				Panels: PanelsState{
					Request:  RequestPanelState{ActiveTab: "params"},
					Response: ResponsePanelState{ActiveTab: "body"},
				},
			},
			check: func(s *Session) bool { return s.ActivePanel == "collections" },
		},
		{
			name: "invalid request tab is corrected",
			session: &Session{
				Version:     1,
				ActivePanel: "request",
				Panels: PanelsState{
					Request:  RequestPanelState{ActiveTab: "invalid"},
					Response: ResponsePanelState{ActiveTab: "body"},
				},
			},
			check: func(s *Session) bool { return s.Panels.Request.ActiveTab == "params" },
		},
		{
			name: "invalid response tab is corrected",
			session: &Session{
				Version:     1,
				ActivePanel: "collections",
				Panels: PanelsState{
					Request:  RequestPanelState{ActiveTab: "params"},
					Response: ResponsePanelState{ActiveTab: "invalid"},
				},
			},
			check: func(s *Session) bool { return s.Panels.Response.ActiveTab == "body" },
		},
		{
			name: "negative scroll position is corrected",
			session: &Session{
				Version:     1,
				ActivePanel: "collections",
				Panels: PanelsState{
					Collections: CollectionsPanelState{ScrollPosition: -5, SelectedIndex: -3},
					Request:     RequestPanelState{ActiveTab: "params"},
					Response:    ResponsePanelState{ActiveTab: "body", ScrollPosition: -10},
				},
			},
			check: func(s *Session) bool {
				return s.Panels.Collections.ScrollPosition == 0 &&
					s.Panels.Collections.SelectedIndex == 0 &&
					s.Panels.Response.ScrollPosition == 0
			},
		},
		{
			name: "valid session unchanged",
			session: &Session{
				Version:     1,
				ActivePanel: "request",
				Panels: PanelsState{
					Collections: CollectionsPanelState{ScrollPosition: 5, SelectedIndex: 2},
					Request:     RequestPanelState{ActiveTab: "body"},
					Response:    ResponsePanelState{ActiveTab: "headers", ScrollPosition: 10},
				},
			},
			check: func(s *Session) bool {
				return s.ActivePanel == "request" &&
					s.Panels.Collections.ScrollPosition == 5 &&
					s.Panels.Request.ActiveTab == "body" &&
					s.Panels.Response.ActiveTab == "headers"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory (for validation that checks file existence)
			dir, err := os.MkdirTemp("", "session_test")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(dir)

			result := tt.session.Validate(dir)
			if !tt.check(result) {
				t.Errorf("Validate() check failed for %s", tt.name)
			}
		})
	}
}

func TestValidateCollectionReference(t *testing.T) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "session_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	// Create collections directory with a file
	collectionsDir := filepath.Join(dir, SessionDir, "collections")
	if err := os.MkdirAll(collectionsDir, 0755); err != nil {
		t.Fatalf("Failed to create collections dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(collectionsDir, "existing.json"), []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to create collection file: %v", err)
	}

	tests := []struct {
		name             string
		activeCollection string
		wantCleared      bool
	}{
		{
			name:             "existing collection is kept",
			activeCollection: "existing.json",
			wantCleared:      false,
		},
		{
			name:             "non-existing collection is cleared",
			activeCollection: "missing.json",
			wantCleared:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Session{
				Version:          1,
				ActivePanel:      "collections",
				ActiveCollection: tt.activeCollection,
				Panels: PanelsState{
					Request:  RequestPanelState{ActiveTab: "params"},
					Response: ResponsePanelState{ActiveTab: "body"},
				},
			}

			result := s.Validate(dir)

			if tt.wantCleared && result.ActiveCollection != "" {
				t.Errorf("Expected ActiveCollection to be cleared, got %s", result.ActiveCollection)
			}
			if !tt.wantCleared && result.ActiveCollection == "" {
				t.Error("Expected ActiveCollection to be kept, but it was cleared")
			}
		})
	}
}
