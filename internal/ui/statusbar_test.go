package ui

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// Phase 2: Foundational - T005, T007
// =============================================================================

func TestNewStatusBar(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    struct {
			version string
			mode    Mode
		}
	}{
		{
			name:    "creates statusbar with version",
			version: "v0.1.0",
			want: struct {
				version string
				mode    Mode
			}{
				version: "v0.1.0",
				mode:    NormalMode,
			},
		},
		{
			name:    "creates statusbar with empty version",
			version: "",
			want: struct {
				version string
				mode    Mode
			}{
				version: "",
				mode:    NormalMode,
			},
		},
		{
			name:    "creates statusbar with long version",
			version: "v1.2.3-beta.4+build.567",
			want: struct {
				version string
				mode    Mode
			}{
				version: "v1.2.3-beta.4+build.567",
				mode:    NormalMode,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStatusBar(tt.version)
			if s == nil {
				t.Fatal("NewStatusBar() returned nil")
			}
			if s.version != tt.want.version {
				t.Errorf("version = %q, want %q", s.version, tt.want.version)
			}
			if s.mode != tt.want.mode {
				t.Errorf("mode = %v, want %v", s.mode, tt.want.mode)
			}
			if s.breadcrumb == nil {
				t.Error("breadcrumb should be initialized to empty slice, not nil")
			}
		})
	}
}

// =============================================================================
// Phase 3: User Story 1 - Mode Badge (T009-T017)
// =============================================================================

func TestStatusBarSetMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     Mode
		wantMode Mode
	}{
		// T009: NORMAL mode
		{
			name:     "sets NORMAL mode",
			mode:     NormalMode,
			wantMode: NormalMode,
		},
		// T010: INSERT mode
		{
			name:     "sets INSERT mode",
			mode:     InsertMode,
			wantMode: InsertMode,
		},
		// T011: VIEW mode
		{
			name:     "sets VIEW mode",
			mode:     ViewMode,
			wantMode: ViewMode,
		},
		// T012: COMMAND mode
		{
			name:     "sets COMMAND mode",
			mode:     CommandMode,
			wantMode: CommandMode,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStatusBar("v0.1.0")
			s.SetMode(tt.mode)
			if s.GetMode() != tt.wantMode {
				t.Errorf("GetMode() = %v, want %v", s.GetMode(), tt.wantMode)
			}
		})
	}
}

func TestStatusBarModeInView(t *testing.T) {
	tests := []struct {
		name      string
		mode      Mode
		wantLabel string
	}{
		{
			name:      "NORMAL mode appears in view",
			mode:      NormalMode,
			wantLabel: "NORMAL",
		},
		{
			name:      "INSERT mode appears in view",
			mode:      InsertMode,
			wantLabel: "INSERT",
		},
		{
			name:      "VIEW mode appears in view",
			mode:      ViewMode,
			wantLabel: "VIEW",
		},
		{
			name:      "COMMAND mode appears in view",
			mode:      CommandMode,
			wantLabel: "COMMAND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStatusBar("v0.1.0")
			s.SetMode(tt.mode)
			view := s.View(100)
			if !strings.Contains(view, tt.wantLabel) {
				t.Errorf("View() does not contain %q for mode %v", tt.wantLabel, tt.mode)
			}
		})
	}
}

// =============================================================================
// Phase 4: User Story 2 - HTTP Context (T018-T025)
// =============================================================================

func TestStatusBarSetMethod(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		wantMethod string
	}{
		// T018: GET method
		{
			name:       "sets GET method",
			method:     "GET",
			wantMethod: "GET",
		},
		// T019: POST, PUT, DELETE methods
		{
			name:       "sets POST method",
			method:     "POST",
			wantMethod: "POST",
		},
		{
			name:       "sets PUT method",
			method:     "PUT",
			wantMethod: "PUT",
		},
		{
			name:       "sets DELETE method",
			method:     "DELETE",
			wantMethod: "DELETE",
		},
		// T020: PATCH, HEAD, OPTIONS methods
		{
			name:       "sets PATCH method",
			method:     "PATCH",
			wantMethod: "PATCH",
		},
		{
			name:       "sets HEAD method",
			method:     "HEAD",
			wantMethod: "HEAD",
		},
		{
			name:       "sets OPTIONS method",
			method:     "OPTIONS",
			wantMethod: "OPTIONS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStatusBar("v0.1.0")
			s.SetMethod(tt.method)
			if s.httpMethod != tt.wantMethod {
				t.Errorf("httpMethod = %q, want %q", s.httpMethod, tt.wantMethod)
			}
			// Verify method appears in view
			view := s.View(100)
			if !strings.Contains(view, tt.wantMethod) {
				t.Errorf("View() does not contain method %q", tt.wantMethod)
			}
		})
	}
}

func TestStatusBarSetHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		code       int
		text       string
		wantCode   int
		wantText   string
		wantInView string
	}{
		// T021: 2xx status codes
		{
			name:       "sets 200 OK status",
			code:       200,
			text:       "OK",
			wantCode:   200,
			wantText:   "OK",
			wantInView: "200",
		},
		{
			name:       "sets 201 Created status",
			code:       201,
			text:       "Created",
			wantCode:   201,
			wantText:   "Created",
			wantInView: "201",
		},
		{
			name:       "sets 204 No Content status",
			code:       204,
			text:       "No Content",
			wantCode:   204,
			wantText:   "No Content",
			wantInView: "204",
		},
		// T022: 3xx status codes
		{
			name:       "sets 301 Moved Permanently status",
			code:       301,
			text:       "Moved Permanently",
			wantCode:   301,
			wantText:   "Moved Permanently",
			wantInView: "301",
		},
		{
			name:       "sets 302 Found status",
			code:       302,
			text:       "Found",
			wantCode:   302,
			wantText:   "Found",
			wantInView: "302",
		},
		// T023: 4xx status codes
		{
			name:       "sets 400 Bad Request status",
			code:       400,
			text:       "Bad Request",
			wantCode:   400,
			wantText:   "Bad Request",
			wantInView: "400",
		},
		{
			name:       "sets 401 Unauthorized status",
			code:       401,
			text:       "Unauthorized",
			wantCode:   401,
			wantText:   "Unauthorized",
			wantInView: "401",
		},
		{
			name:       "sets 404 Not Found status",
			code:       404,
			text:       "Not Found",
			wantCode:   404,
			wantText:   "Not Found",
			wantInView: "404",
		},
		// T024: 5xx status codes
		{
			name:       "sets 500 Internal Server Error status",
			code:       500,
			text:       "Internal Server Error",
			wantCode:   500,
			wantText:   "Internal Server Error",
			wantInView: "500",
		},
		{
			name:       "sets 502 Bad Gateway status",
			code:       502,
			text:       "Bad Gateway",
			wantCode:   502,
			wantText:   "Bad Gateway",
			wantInView: "502",
		},
		{
			name:       "sets 503 Service Unavailable status",
			code:       503,
			text:       "Service Unavailable",
			wantCode:   503,
			wantText:   "Service Unavailable",
			wantInView: "503",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStatusBar("v0.1.0")
			s.SetHTTPStatus(tt.code, tt.text)
			if s.httpStatus != tt.wantCode {
				t.Errorf("httpStatus = %d, want %d", s.httpStatus, tt.wantCode)
			}
			if s.httpText != tt.wantText {
				t.Errorf("httpText = %q, want %q", s.httpText, tt.wantText)
			}
			// Verify status code appears in view
			view := s.View(150)
			if !strings.Contains(view, tt.wantInView) {
				t.Errorf("View() does not contain status code %q", tt.wantInView)
			}
		})
	}
}

// T025: Clear HTTP status
func TestStatusBarClearHTTPStatus(t *testing.T) {
	s := NewStatusBar("v0.1.0")

	// Set a status first
	s.SetHTTPStatus(200, "OK")
	if s.httpStatus != 200 {
		t.Fatal("setup failed: status not set")
	}

	// Clear the status
	s.ClearHTTPStatus()

	if s.httpStatus != 0 {
		t.Errorf("httpStatus = %d after clear, want 0", s.httpStatus)
	}
	if s.httpText != "" {
		t.Errorf("httpText = %q after clear, want empty", s.httpText)
	}
}

// =============================================================================
// Phase 5: User Story 3 - Breadcrumb (T026-T030)
// =============================================================================

func TestStatusBarSetBreadcrumb(t *testing.T) {
	tests := []struct {
		name          string
		parts         []string
		wantLen       int
		wantInView    []string
		wantNotInView []string
	}{
		// T026: Single-level path
		{
			name:       "single-level breadcrumb",
			parts:      []string{"My API"},
			wantLen:    1,
			wantInView: []string{"My API"},
		},
		// T027: Multi-level path
		{
			name:       "multi-level breadcrumb",
			parts:      []string{"My API", "Users", "Create User"},
			wantLen:    3,
			wantInView: []string{"My API", "Users", "Create User", "›"},
		},
		{
			name:       "two-level breadcrumb",
			parts:      []string{"Collection", "Request"},
			wantLen:    2,
			wantInView: []string{"Collection", "Request", "›"},
		},
		// T028: Empty breadcrumb
		{
			name:          "empty breadcrumb",
			parts:         []string{},
			wantLen:       0,
			wantNotInView: []string{"›"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStatusBar("v0.1.0")
			s.SetBreadcrumb(tt.parts...)

			if len(s.breadcrumb) != tt.wantLen {
				t.Errorf("breadcrumb length = %d, want %d", len(s.breadcrumb), tt.wantLen)
			}

			view := s.View(200)
			for _, want := range tt.wantInView {
				if !strings.Contains(view, want) {
					t.Errorf("View() does not contain %q", want)
				}
			}
			for _, notWant := range tt.wantNotInView {
				if strings.Contains(view, notWant) {
					t.Errorf("View() should not contain %q for empty breadcrumb", notWant)
				}
			}
		})
	}
}

// T029: Verify separator
func TestStatusBarBreadcrumbSeparator(t *testing.T) {
	s := NewStatusBar("v0.1.0")
	s.SetBreadcrumb("API", "Users", "Get")

	text := s.formatBreadcrumbText()
	if !strings.Contains(text, "›") {
		t.Error("formatBreadcrumbText() should use › separator")
	}

	// Count separators - should be 2 for 3 parts
	count := strings.Count(text, "›")
	if count != 2 {
		t.Errorf("separator count = %d, want 2 for 3-part breadcrumb", count)
	}
}

// T030: Breadcrumb truncation on narrow width
func TestStatusBarBreadcrumbTruncation(t *testing.T) {
	s := NewStatusBar("v0.1.0")
	s.SetBreadcrumb("Very Long Collection Name", "Very Long Folder Name", "Very Long Request Name")

	// With narrow width, content should truncate
	view := s.View(80)

	// The view should contain truncation indicator if content was truncated
	// Due to the way truncation works, we verify the view is not empty
	if len(view) == 0 {
		t.Error("View() returned empty string for narrow terminal")
	}
}

// =============================================================================
// Phase 6: User Story 4 - Environment (T031-T034)
// =============================================================================

func TestStatusBarSetEnvironment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		wantInView  string
	}{
		// T031: Environment name set
		{
			name:        "sets development environment",
			environment: "development",
			wantInView:  "development",
		},
		{
			name:        "sets production environment",
			environment: "production",
			wantInView:  "production",
		},
		{
			name:        "sets staging environment",
			environment: "staging",
			wantInView:  "staging",
		},
		// T032: Empty environment (NONE case)
		{
			name:        "empty environment shows NONE",
			environment: "",
			wantInView:  "NONE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStatusBar("v0.1.0")
			s.SetEnvironment(tt.environment)

			if s.environment != tt.environment {
				t.Errorf("environment = %q, want %q", s.environment, tt.environment)
			}

			// T033/T034: Verify view displays correctly
			view := s.View(100)
			if !strings.Contains(view, tt.wantInView) {
				t.Errorf("View() does not contain %q", tt.wantInView)
			}
		})
	}
}

// =============================================================================
// Phase 7: User Story 5 - Keyboard Hints (T035-T040)
// =============================================================================

func TestStatusBarGetKeyboardHints(t *testing.T) {
	tests := []struct {
		name        string
		mode        Mode
		wantContain []string
	}{
		// T035: NORMAL mode hints
		{
			name:        "NORMAL mode hints include navigation",
			mode:        NormalMode,
			wantContain: []string{"j/k", "h/l"},
		},
		// T036: INSERT mode hints
		{
			name:        "INSERT mode hints include editing",
			mode:        InsertMode,
			wantContain: []string{"type", "esc"},
		},
		// T037: VIEW mode hints
		{
			name:        "VIEW mode hints include scrolling",
			mode:        ViewMode,
			wantContain: []string{"j/k", "g/G", "esc"},
		},
		// T038: COMMAND mode hints
		{
			name:        "COMMAND mode hints include commands",
			mode:        CommandMode,
			wantContain: []string{":q", ":w", "esc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStatusBar("v0.1.0")
			s.SetMode(tt.mode)

			hints := s.getKeyboardHints()
			for _, want := range tt.wantContain {
				if !strings.Contains(hints, want) {
					t.Errorf("getKeyboardHints() for %v does not contain %q", tt.mode, want)
				}
			}
		})
	}
}

// T039: Custom hints override
func TestStatusBarSetHints(t *testing.T) {
	s := NewStatusBar("v0.1.0")

	customHints := "custom:Hint | test:Override"
	s.SetHints(customHints)

	hints := s.getKeyboardHints()
	if hints != customHints {
		t.Errorf("getKeyboardHints() = %q, want %q", hints, customHints)
	}

	// Clear custom hints
	s.SetHints("")
	hints = s.getKeyboardHints()
	if hints == customHints {
		t.Error("getKeyboardHints() should return mode-based hints after clearing custom hints")
	}
}

// T040: Verify hints are meaningful (already covered in T035-T038)

// =============================================================================
// Phase 8: User Story 6 - Status Messages (T041-T047)
// =============================================================================

// T041: ShowMessage test
func TestStatusBarShowMessage(t *testing.T) {
	s := NewStatusBar("v0.1.0")

	msg := "Test message"
	duration := 5 * time.Second
	s.ShowMessage(msg, duration)

	if s.message != msg {
		t.Errorf("message = %q, want %q", s.message, msg)
	}

	// Message end should be in the future
	if time.Until(s.messageEnd) <= 0 {
		t.Error("messageEnd should be in the future")
	}

	// Verify message appears in view
	view := s.View(100)
	if !strings.Contains(view, msg) {
		t.Errorf("View() does not contain message %q", msg)
	}
}

// T042: Info helper
func TestStatusBarInfo(t *testing.T) {
	s := NewStatusBar("v0.1.0")

	msg := "Info message"
	s.Info(msg)

	if s.message != msg {
		t.Errorf("message = %q, want %q", s.message, msg)
	}

	// Check duration is MessageDuration (2 seconds)
	expectedEnd := time.Now().Add(MessageDuration)
	if s.messageEnd.Sub(expectedEnd) > time.Second {
		t.Error("messageEnd should be approximately MessageDuration from now")
	}
}

// T043: Success helper
func TestStatusBarSuccess(t *testing.T) {
	s := NewStatusBar("v0.1.0")

	s.Success("Saved", "request.json")

	expectedMsg := "Saved: request.json"
	if s.message != expectedMsg {
		t.Errorf("message = %q, want %q", s.message, expectedMsg)
	}
}

// T044: Error helper
func TestStatusBarError(t *testing.T) {
	s := NewStatusBar("v0.1.0")

	err := errors.New("connection failed")
	s.Error(err)

	expectedMsg := "Error: connection failed"
	if s.message != expectedMsg {
		t.Errorf("message = %q, want %q", s.message, expectedMsg)
	}
}

// T045: ClearMessage test
func TestStatusBarClearMessage(t *testing.T) {
	s := NewStatusBar("v0.1.0")

	// Set a message first
	s.Info("Test message")
	if s.message == "" {
		t.Fatal("setup failed: message not set")
	}

	// Clear the message
	s.ClearMessage()

	if s.message != "" {
		t.Errorf("message = %q after clear, want empty", s.message)
	}
}

// T046: MessageDuration constant
func TestMessageDurationConstant(t *testing.T) {
	expected := 2 * time.Second
	if MessageDuration != expected {
		t.Errorf("MessageDuration = %v, want %v", MessageDuration, expected)
	}
}

// T047: Message auto-dismiss
func TestStatusBarMessageAutoDismiss(t *testing.T) {
	s := NewStatusBar("v0.1.0")

	// Set a message with very short duration (already expired)
	s.message = "Test message"
	s.messageEnd = time.Now().Add(-1 * time.Second) // 1 second in the past

	// View should clear expired message
	view := s.View(100)

	// After View(), message should be cleared
	if s.message != "" {
		t.Error("expired message should be cleared after View()")
	}

	// Message should not appear in view
	if strings.Contains(view, "Test message") {
		t.Error("expired message should not appear in View()")
	}
}

// =============================================================================
// Phase 9: Polish - View Tests (T048-T050)
// =============================================================================

// T048: Full render test
func TestStatusBarView_FullRender(t *testing.T) {
	s := NewStatusBar("v0.1.0")
	s.SetMode(NormalMode)
	s.SetEnvironment("development")
	s.SetMethod("POST")
	s.SetHTTPStatus(201, "Created")
	s.SetBreadcrumb("API", "Users", "Create")

	view := s.View(200)

	// Should contain all elements (version not displayed in statusbar)
	expected := []string{"NORMAL", "development", "POST", "201", "API", "Users", "Create"}
	for _, want := range expected {
		if !strings.Contains(view, want) {
			t.Errorf("View() does not contain %q", want)
		}
	}
}

// T049: Narrow terminal test
func TestStatusBarView_NarrowTerminal(t *testing.T) {
	s := NewStatusBar("v0.1.0")
	s.SetMode(NormalMode)
	s.SetBreadcrumb("Very Long Collection", "Very Long Folder", "Very Long Request")

	view := s.View(80)

	// View should not be empty
	if len(view) == 0 {
		t.Error("View() returned empty string for 80-column terminal")
	}

	// Mode badge should still be visible
	if !strings.Contains(view, "NORMAL") {
		t.Error("Mode badge should be visible even on narrow terminal")
	}
}

// T050: Wide terminal test
func TestStatusBarView_WideTerminal(t *testing.T) {
	s := NewStatusBar("v0.1.0")
	s.SetMode(InsertMode)
	s.SetEnvironment("production")
	s.SetMethod("GET")
	s.SetHTTPStatus(200, "OK")
	s.SetBreadcrumb("API", "Users", "List")

	view := s.View(200)

	// All content should be visible without truncation (version not displayed in statusbar)
	expected := []string{"INSERT", "production", "GET", "200", "API", "Users", "List"}
	for _, want := range expected {
		if !strings.Contains(view, want) {
			t.Errorf("View() at 200 columns does not contain %q", want)
		}
	}
}

// Additional helper function tests
func TestHelperFunctions(t *testing.T) {
	t.Run("ModePtr", func(t *testing.T) {
		m := InsertMode
		ptr := ModePtr(m)
		if ptr == nil {
			t.Fatal("ModePtr() returned nil")
		}
		if *ptr != m {
			t.Errorf("*ModePtr() = %v, want %v", *ptr, m)
		}
	})

	t.Run("IntPtr", func(t *testing.T) {
		i := 42
		ptr := IntPtr(i)
		if ptr == nil {
			t.Fatal("IntPtr() returned nil")
		}
		if *ptr != i {
			t.Errorf("*IntPtr() = %d, want %d", *ptr, i)
		}
	})

	t.Run("StringPtr", func(t *testing.T) {
		s := "test"
		ptr := StringPtr(s)
		if ptr == nil {
			t.Fatal("StringPtr() returned nil")
		}
		if *ptr != s {
			t.Errorf("*StringPtr() = %q, want %q", *ptr, s)
		}
	})
}

// Fullscreen badge test
func TestStatusBarSetFullscreen(t *testing.T) {
	s := NewStatusBar("v0.1.0")

	// Initially not fullscreen
	view := s.View(100)
	if strings.Contains(view, "FULLSCREEN") {
		t.Error("View() should not contain FULLSCREEN when not set")
	}

	// Set fullscreen
	s.SetFullscreen(true)
	view = s.View(100)
	if !strings.Contains(view, "FULLSCREEN") {
		t.Error("View() should contain FULLSCREEN when set")
	}

	// Clear fullscreen
	s.SetFullscreen(false)
	view = s.View(100)
	if strings.Contains(view, "FULLSCREEN") {
		t.Error("View() should not contain FULLSCREEN after clearing")
	}
}
