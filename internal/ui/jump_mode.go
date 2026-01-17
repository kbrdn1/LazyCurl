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
//  1. If ≤26 targets: Use single letters (a, s, d, f, j, k, l, g, h, ...)
//  2. If >26 targets: Use ALL two-letter combinations (aa, as, ad, ..., sa, ss, ...)
//     This ensures the first keypress only filters, never jumps.
//
// The targets slice is modified in place with Label field populated.
func (m *JumpModeState) AssignLabels(targets []JumpTarget) {
	numKeys := len(allKeys)
	numTargets := len(targets)

	// If we have more than 26 targets, use all two-letter combinations
	// This way, typing the first letter only filters, never jumps immediately
	useTwoLetters := numTargets > numKeys

	for i := range targets {
		if useTwoLetters {
			// Two character labels for all targets when >26
			firstIdx := i / numKeys
			secondIdx := i % numKeys

			// Wrap around if we have very many targets (>676)
			if firstIdx >= numKeys {
				firstIdx = firstIdx % numKeys
			}

			targets[i].Label = string(allKeys[firstIdx]) + string(allKeys[secondIdx])
		} else {
			// Single character label when ≤26 targets
			targets[i].Label = string(allKeys[i])
		}
	}
}

// insertAtVisualPosition inserts text at a visual column position in an ANSI-colored string.
// It properly handles ANSI escape sequences by tracking visual vs byte positions.
func insertAtVisualPosition(line string, visualCol int, insertText string, replaceLen int) string {
	visualPos := 0
	bytePos := 0
	lineBytes := []byte(line)
	inEscape := false

	// Find the byte position corresponding to the visual column
	startBytePos := -1
	endBytePos := -1

	for bytePos < len(lineBytes) {
		b := lineBytes[bytePos]

		// Check for start of ANSI escape sequence
		if b == '\x1b' && bytePos+1 < len(lineBytes) && lineBytes[bytePos+1] == '[' {
			inEscape = true
			bytePos++
			continue
		}

		if inEscape {
			// Continue until we hit 'm' which ends the escape sequence
			if b == 'm' {
				inEscape = false
			}
			bytePos++
			continue
		}

		// This is a visible character
		if visualPos == visualCol && startBytePos == -1 {
			startBytePos = bytePos
		}

		visualPos++

		// Advance byte position, handling UTF-8
		charLen := 1
		if b >= 0xF0 {
			charLen = 4
		} else if b >= 0xE0 {
			charLen = 3
		} else if b >= 0xC0 {
			charLen = 2
		}
		bytePos += charLen

		if startBytePos != -1 && visualPos == visualCol+replaceLen {
			endBytePos = bytePos
			break
		}
	}

	// If we couldn't find the start position, append at end
	if startBytePos == -1 {
		return line + " " + insertText
	}

	if endBytePos == -1 {
		endBytePos = len(lineBytes)
	}

	// Build new line: before + insert + after
	return string(lineBytes[:startBytePos]) + insertText + string(lineBytes[endBytePos:])
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

		// Check visual width of the line (without ANSI codes)
		visualWidth := lipgloss.Width(line)
		if col < 0 || col >= visualWidth {
			// Adjust column if out of bounds
			if col >= visualWidth && visualWidth > 0 {
				col = visualWidth - len(target.Label)
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

		// Insert the label at the visual position, handling ANSI codes
		lines[row] = insertAtVisualPosition(line, col, styledLabel, len(label))
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
