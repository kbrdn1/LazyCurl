package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// JumpAction defines what happens when jumping to a target
type JumpAction int

const (
	JumpFocus    JumpAction = iota // Focus the element (inputs, URL field)
	JumpSelect                     // Select in list (tree items, rows)
	JumpActivate                   // Trigger action (tabs, buttons)
)

// String returns the display name for the action
func (a JumpAction) String() string {
	switch a {
	case JumpFocus:
		return "focus"
	case JumpSelect:
		return "select"
	case JumpActivate:
		return "activate"
	default:
		return "unknown"
	}
}

// JumpTarget represents an element that can be jumped to
type JumpTarget struct {
	// Label assignment
	Label string // Assigned label: "a", "b", ..., "aa", "ab", ...

	// Location
	Panel PanelType // Which panel (CollectionsPanel, RequestPanel, ResponsePanel)
	Row   int       // Screen row for label placement (0-indexed from panel start)
	Col   int       // Screen column for label placement

	// Target identification
	Index     int    // Index within panel's item list
	ElementID string // Optional unique ID (e.g., request ID, tab name)

	// Action on jump
	Action JumpAction // What happens when this target is jumped to
}

// JumpModeState manages the jump navigation state machine
type JumpModeState struct {
	// State
	Active    bool // Whether jump mode is currently active
	AllPanels bool // true = cross-panel mode (F), false = single panel (f)

	// Target management
	Targets    []JumpTarget // All available jump targets with assigned labels
	ScopePanel PanelType    // Active panel when jump mode started

	// Input handling
	Filter        string // Current typed filter for multi-key mode
	MatchingCount int    // Number of targets matching current filter

	// Memory
	LastJumpLabel string // Last successful jump for repeat navigation
}

// Home row keys have priority (easier to reach)
var homeRowKeys = []rune{'a', 's', 'd', 'f', 'j', 'k', 'l'}

// Other keys used after home row
var otherKeys = []rune{'g', 'h', 'q', 'w', 'e', 'r', 't', 'y', 'u', 'i', 'o', 'p', 'z', 'x', 'c', 'v', 'b', 'n', 'm'}

// allKeys combines home row and other keys for label assignment
var allKeys = append(homeRowKeys, otherKeys...)

// NewJumpMode creates a new JumpModeState with default values
func NewJumpMode() *JumpModeState {
	return &JumpModeState{
		Active:        false,
		AllPanels:     false,
		Targets:       nil,
		ScopePanel:    CollectionsPanel,
		Filter:        "",
		MatchingCount: 0,
		LastJumpLabel: "",
	}
}

// Activate starts jump mode for a single panel.
// Collects targets from the specified panel only.
func (m *JumpModeState) Activate(currentPanel PanelType) {
	m.Active = true
	m.AllPanels = false
	m.ScopePanel = currentPanel
	m.Filter = ""
	m.MatchingCount = 0
	m.Targets = nil
}

// ActivateAllPanels starts jump mode for all panels.
// Collects targets from Collections, Request, and Response panels.
func (m *JumpModeState) ActivateAllPanels() {
	m.Active = true
	m.AllPanels = true
	m.Filter = ""
	m.MatchingCount = 0
	m.Targets = nil
}

// Deactivate exits jump mode without performing a jump.
func (m *JumpModeState) Deactivate() {
	m.Active = false
	m.Filter = ""
	m.MatchingCount = 0
	m.Targets = nil
}

// IsActive returns whether jump mode is currently active.
func (m *JumpModeState) IsActive() bool {
	return m.Active
}

// GetFilter returns the current typed filter string.
func (m *JumpModeState) GetFilter() string {
	return m.Filter
}

// CycleScopePanel cycles to the next panel scope.
// Only works when not in all-panels mode.
// Returns the new scope panel.
func (m *JumpModeState) CycleScopePanel() PanelType {
	if m.AllPanels {
		return m.ScopePanel
	}

	// Cycle through: Collections -> Request -> Response -> Collections
	switch m.ScopePanel {
	case CollectionsPanel:
		m.ScopePanel = RequestPanel
	case RequestPanel:
		m.ScopePanel = ResponsePanel
	case ResponsePanel:
		m.ScopePanel = CollectionsPanel
	default:
		m.ScopePanel = CollectionsPanel
	}

	// Reset filter when changing scope
	m.Filter = ""
	m.MatchingCount = len(m.GetVisibleTargets())

	return m.ScopePanel
}

// SetTargets sets the jump targets and assigns labels to them.
func (m *JumpModeState) SetTargets(targets []JumpTarget) {
	m.AssignLabels(targets)
	m.Targets = targets
	m.MatchingCount = len(targets)
}

// HandleKey processes a key press during jump mode.
// Returns:
//   - target: The matched target if a complete match is found, nil otherwise
//   - shouldCancel: true if mode should be canceled (invalid key or escape)
func (m *JumpModeState) HandleKey(key string) (target *JumpTarget, shouldCancel bool) {
	// Escape cancels
	if key == "esc" {
		return nil, true
	}

	// Only accept single lowercase letters
	if len(key) != 1 {
		return nil, true
	}

	r := rune(key[0])
	if r < 'a' || r > 'z' {
		return nil, true
	}

	// Add to filter
	m.Filter += key

	// Find matching targets
	matching := m.GetVisibleTargets()
	m.MatchingCount = len(matching)

	// No matches - cancel
	if m.MatchingCount == 0 {
		return nil, true
	}

	// Exact match found
	if m.MatchingCount == 1 && matching[0].Label == m.Filter {
		m.LastJumpLabel = matching[0].Label
		return &matching[0], false
	}

	// Check for exact match among multiple possibilities
	for i := range matching {
		if matching[i].Label == m.Filter {
			m.LastJumpLabel = matching[i].Label
			return &matching[i], false
		}
	}

	// Still filtering, wait for more input
	return nil, false
}

// GetVisibleTargets returns targets that match the current filter and panel scope.
// Used for rendering - only visible targets should display labels.
func (m *JumpModeState) GetVisibleTargets() []JumpTarget {
	var visible []JumpTarget

	for _, t := range m.Targets {
		// Apply panel scope filter if not in all-panels mode
		if !m.AllPanels && t.Panel != m.ScopePanel {
			continue
		}

		// Apply prefix filter if set
		if m.Filter != "" {
			if len(t.Label) < len(m.Filter) || t.Label[:len(m.Filter)] != m.Filter {
				continue
			}
		}

		visible = append(visible, t)
	}
	return visible
}

// AssignLabels assigns unique labels to a list of targets using home-row priority
//
// Algorithm:
//  1. Use home row keys first: a, s, d, f, j, k, l
//  2. Then other keys: g, h, q, w, e, r, t, y, u, i, o, p, z, x, c, v, b, n, m
//  3. For >26 targets, use two-character combinations: aa, as, ad, ...
//
// The targets slice is modified in place with Label field populated.
func (m *JumpModeState) AssignLabels(targets []JumpTarget) {
	numKeys := len(allKeys)

	for i := range targets {
		if i < numKeys {
			// Single character label
			targets[i].Label = string(allKeys[i])
		} else {
			// Two character label for overflow
			firstIdx := (i - numKeys) / numKeys
			secondIdx := (i - numKeys) % numKeys

			// Wrap around if we have very many targets
			if firstIdx >= numKeys {
				firstIdx = firstIdx % numKeys
			}

			targets[i].Label = string(allKeys[firstIdx]) + string(allKeys[secondIdx])
		}
	}
}

// RenderOverlay overlays jump labels onto the base view.
// It places styled labels at the Row/Col positions of visible targets.
//
// Parameters:
//   - baseView: The fully rendered view string to overlay labels on
//   - width: Total width of the view
//   - height: Total height of the view
//
// Returns the view with jump labels overlaid.
func (m *JumpModeState) RenderOverlay(baseView string, width, height int) string {
	if !m.Active {
		return baseView
	}

	visibleTargets := m.GetVisibleTargets()
	if len(visibleTargets) == 0 {
		return baseView
	}

	// Define label styles
	activeLabelStyle := lipgloss.NewStyle().
		Background(styles.JumpLabelBg).
		Foreground(styles.JumpLabelFg).
		Bold(true)

	matchedCharStyle := lipgloss.NewStyle().
		Background(styles.JumpLabelMatchBg).
		Foreground(styles.JumpLabelMatchFg).
		Bold(true)

	// Split view into lines for manipulation
	lines := strings.Split(baseView, "\n")

	// Process each visible target
	for _, target := range visibleTargets {
		row := target.Row
		col := target.Col

		// Bounds check
		if row < 0 || row >= len(lines) {
			continue
		}

		line := lines[row]
		lineRunes := []rune(line)

		// Ensure column is in bounds
		if col < 0 || col >= len(lineRunes) {
			// Try to place at end of line if col is out of bounds
			if col >= len(lineRunes) && len(lineRunes) > 0 {
				col = len(lineRunes) - len(target.Label)
				if col < 0 {
					col = 0
				}
			} else {
				continue
			}
		}

		// Build the styled label
		label := target.Label
		var styledLabel string

		if m.Filter != "" && len(m.Filter) <= len(label) {
			// Highlight matched prefix in green, rest in orange
			matchedPart := label[:len(m.Filter)]
			remainingPart := label[len(m.Filter):]
			styledLabel = matchedCharStyle.Render(matchedPart) + activeLabelStyle.Render(remainingPart)
		} else {
			// Full label in orange
			styledLabel = activeLabelStyle.Render(label)
		}

		// Replace characters at position with styled label
		// We need to handle ANSI codes properly - for simplicity, we'll prepend label
		// A more sophisticated approach would parse ANSI codes

		// Simple approach: insert label at beginning of visible content
		// For MVP, we'll place labels at the start of the line if within panel bounds
		if col == 0 {
			lines[row] = styledLabel + " " + string(lineRunes)
		} else if col+len(label) <= len(lineRunes) {
			// Replace characters at position
			before := string(lineRunes[:col])
			after := string(lineRunes[col+len(label):])
			lines[row] = before + styledLabel + after
		} else {
			// Append at end
			lines[row] = string(lineRunes) + " " + styledLabel
		}
	}

	return strings.Join(lines, "\n")
}

// GetScopePanel returns the current scope panel.
func (m *JumpModeState) GetScopePanel() PanelType {
	return m.ScopePanel
}

// IsAllPanels returns whether jump mode is targeting all panels.
func (m *JumpModeState) IsAllPanels() bool {
	return m.AllPanels
}
