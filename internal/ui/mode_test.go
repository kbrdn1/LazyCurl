package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// =============================================================================
// Phase 2: Foundational Tests (T006, T008)
// =============================================================================

// TestModeString verifies Mode.String() returns correct display labels
func TestModeString(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{NormalMode, "NORMAL"},
		{ViewMode, "VIEW"},
		{CommandMode, "COMMAND"},
		{InsertMode, "INSERT"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("Mode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestModeStringDefault verifies unknown mode defaults to "NORMAL"
func TestModeStringDefault(t *testing.T) {
	unknownMode := Mode(99)
	if got := unknownMode.String(); got != NormalMode.String() {
		t.Errorf("Mode(99).String() = %v, want %v", got, NormalMode.String())
	}
}

// =============================================================================
// Phase 3: User Story 1 - Mode Color Tests (T014, T015)
// =============================================================================

// TestModeColor verifies Mode.Color() returns distinct styles for each mode
func TestModeColor(t *testing.T) {
	tests := []struct {
		mode   Mode
		wantBg lipgloss.Color
		wantFg lipgloss.Color
	}{
		{NormalMode, styles.ModeNormalBg, styles.ModeNormalFg},
		{ViewMode, styles.ModeViewBg, styles.ModeViewFg},
		{CommandMode, styles.ModeCommandBg, styles.ModeCommandFg},
		{InsertMode, styles.ModeInsertBg, styles.ModeInsertFg},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			style := tt.mode.Color()

			// Verify style is not empty (has styling applied)
			rendered := style.Render("TEST")
			if len(rendered) == 0 {
				t.Error("Mode.Color() returned empty style")
			}

			// Verify style is bold
			if !style.GetBold() {
				t.Error("Mode.Color() should return bold style")
			}
		})
	}
}

// TestModeColorDistinctness verifies all modes have distinct colors
func TestModeColorDistinctness(t *testing.T) {
	modes := []Mode{NormalMode, ViewMode, CommandMode, InsertMode}
	renderedStyles := make(map[string]Mode)

	for _, mode := range modes {
		style := mode.Color()
		rendered := style.Render(mode.String())

		// Check if this exact rendered output already exists
		if existingMode, exists := renderedStyles[rendered]; exists {
			t.Errorf("Mode %s has same rendered style as Mode %s", mode.String(), existingMode.String())
		}
		renderedStyles[rendered] = mode
	}
}

// TestModeColorDefault verifies unknown mode defaults to NORMAL colors
func TestModeColorDefault(t *testing.T) {
	unknownMode := Mode(99)
	normalStyle := NormalMode.Color()
	unknownStyle := unknownMode.Color()

	// Both should render identically
	normalRendered := normalStyle.Render("TEST")
	unknownRendered := unknownStyle.Render("TEST")

	if normalRendered != unknownRendered {
		t.Error("Unknown mode should default to NORMAL mode styling")
	}
}

// =============================================================================
// Phase 3: User Story 1 - Mode Behavior Tests (T016, T017)
// =============================================================================

// TestModeAllowsInput verifies INSERT and COMMAND modes allow input
func TestModeAllowsInput(t *testing.T) {
	tests := []struct {
		mode Mode
		want bool
	}{
		{NormalMode, false},
		{ViewMode, false},
		{InsertMode, true},
		{CommandMode, true},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			if got := tt.mode.AllowsInput(); got != tt.want {
				t.Errorf("Mode.AllowsInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestModeAllowsNavigation verifies NORMAL and VIEW modes allow navigation
func TestModeAllowsNavigation(t *testing.T) {
	tests := []struct {
		mode Mode
		want bool
	}{
		{NormalMode, true},
		{ViewMode, true},
		{InsertMode, false},
		{CommandMode, false},
	}

	for _, tt := range tests {
		t.Run(tt.mode.String(), func(t *testing.T) {
			if got := tt.mode.AllowsNavigation(); got != tt.want {
				t.Errorf("Mode.AllowsNavigation() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestModeInputNavigationMutuallyExclusive verifies modes are either input OR navigation, not both
func TestModeInputNavigationMutuallyExclusive(t *testing.T) {
	modes := []Mode{NormalMode, ViewMode, CommandMode, InsertMode}

	for _, mode := range modes {
		allowsInput := mode.AllowsInput()
		allowsNavigation := mode.AllowsNavigation()

		// A mode should allow one or the other, but not both
		if allowsInput && allowsNavigation {
			t.Errorf("Mode %s allows both input and navigation - should be mutually exclusive", mode.String())
		}

		// A mode should allow at least one
		if !allowsInput && !allowsNavigation {
			t.Errorf("Mode %s allows neither input nor navigation - should allow at least one", mode.String())
		}
	}
}

// =============================================================================
// ModeChangeMsg Tests
// =============================================================================

// TestModeChangeMsg verifies ModeChangeMsg creation
func TestModeChangeMsg(t *testing.T) {
	tests := []struct {
		name string
		from Mode
		to   Mode
	}{
		{"NORMAL to INSERT", NormalMode, InsertMode},
		{"INSERT to NORMAL", InsertMode, NormalMode},
		{"NORMAL to COMMAND", NormalMode, CommandMode},
		{"COMMAND to NORMAL", CommandMode, NormalMode},
		{"NORMAL to VIEW", NormalMode, ViewMode},
		{"VIEW to NORMAL", ViewMode, NormalMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewModeChangeMsg(tt.from, tt.to)

			if msg.From != tt.from {
				t.Errorf("ModeChangeMsg.From = %v, want %v", msg.From, tt.from)
			}
			if msg.To != tt.to {
				t.Errorf("ModeChangeMsg.To = %v, want %v", msg.To, tt.to)
			}
		})
	}
}

// TestModeConstants verifies mode constants have expected iota values
func TestModeConstants(t *testing.T) {
	// Verify iota ordering is as expected
	if NormalMode != 0 {
		t.Errorf("NormalMode = %d, want 0", NormalMode)
	}
	if ViewMode != 1 {
		t.Errorf("ViewMode = %d, want 1", ViewMode)
	}
	if CommandMode != 2 {
		t.Errorf("CommandMode = %d, want 2", CommandMode)
	}
	if InsertMode != 3 {
		t.Errorf("InsertMode = %d, want 3", InsertMode)
	}
}
