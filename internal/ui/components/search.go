package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// SearchInput represents a search input field
type SearchInput struct {
	visible   bool
	value     string
	cursorPos int
}

// SearchUpdateMsg is sent when search query changes
type SearchUpdateMsg struct {
	Query string
}

// SearchCloseMsg is sent when search is closed
type SearchCloseMsg struct {
	Cancelled bool
}

// NewSearchInput creates a new search input
func NewSearchInput() *SearchInput {
	return &SearchInput{}
}

// Show displays the search input
func (s *SearchInput) Show() {
	s.visible = true
	s.value = ""
	s.cursorPos = 0
}

// Hide hides the search input
func (s *SearchInput) Hide() {
	s.visible = false
}

// IsVisible returns visibility state
func (s *SearchInput) IsVisible() bool {
	return s.visible
}

// GetQuery returns current search query
func (s *SearchInput) GetQuery() string {
	return s.value
}

// Clear clears the search query
func (s *SearchInput) Clear() {
	s.value = ""
	s.cursorPos = 0
}

// Update handles input messages
func (s *SearchInput) Update(msg tea.Msg) (*SearchInput, tea.Cmd) {
	if !s.visible {
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.Hide()
			s.Clear()
			return s, func() tea.Msg {
				return SearchCloseMsg{Cancelled: true}
			}

		case "enter":
			s.Hide()
			return s, func() tea.Msg {
				return SearchCloseMsg{Cancelled: false}
			}

		case "backspace":
			if len(s.value) > 0 && s.cursorPos > 0 {
				s.value = s.value[:s.cursorPos-1] + s.value[s.cursorPos:]
				s.cursorPos--
				return s, func() tea.Msg {
					return SearchUpdateMsg{Query: s.value}
				}
			}

		case "left":
			if s.cursorPos > 0 {
				s.cursorPos--
			}

		case "right":
			if s.cursorPos < len(s.value) {
				s.cursorPos++
			}

		case "home", "ctrl+a":
			s.cursorPos = 0

		case "end", "ctrl+e":
			s.cursorPos = len(s.value)

		case "ctrl+u":
			// Clear line
			s.value = ""
			s.cursorPos = 0
			return s, func() tea.Msg {
				return SearchUpdateMsg{Query: s.value}
			}

		default:
			// Insert character
			if len(msg.String()) == 1 {
				char := msg.String()
				s.value = s.value[:s.cursorPos] + char + s.value[s.cursorPos:]
				s.cursorPos++
				return s, func() tea.Msg {
					return SearchUpdateMsg{Query: s.value}
				}
			}
		}
	}

	return s, nil
}

// View renders the search input inline (for panel header)
func (s *SearchInput) View(width int) string {
	if !s.visible {
		return ""
	}

	// Style
	prefixStyle := lipgloss.NewStyle().
		Foreground(styles.Yellow).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(styles.Text)

	cursorStyle := lipgloss.NewStyle().
		Foreground(styles.Green).
		Bold(true)

	// Render with cursor
	prefix := prefixStyle.Render("/")

	var inputContent string
	if s.cursorPos >= len(s.value) {
		inputContent = inputStyle.Render(s.value) + cursorStyle.Render("█")
	} else {
		before := s.value[:s.cursorPos]
		after := s.value[s.cursorPos:]
		inputContent = inputStyle.Render(before) + cursorStyle.Render("█") + inputStyle.Render(after)
	}

	// Limit width
	maxInputWidth := width - 2 // account for "/" prefix
	if lipgloss.Width(inputContent) > maxInputWidth {
		// Truncate from left to keep cursor visible
		inputContent = "…" + inputContent[len(inputContent)-maxInputWidth+1:]
	}

	return prefix + inputContent
}

// ViewBox renders the search as a box at the top of content (legacy - use ViewCompact instead)
func (s *SearchInput) ViewBox(width int) string {
	return s.ViewCompact(width, 0, 0)
}

// ViewCompact renders a compact inline search bar with match count (no border)
func (s *SearchInput) ViewCompact(width int, matchCount int, totalCount int) string {
	if !s.visible {
		return ""
	}

	prefixStyle := lipgloss.NewStyle().
		Foreground(styles.Yellow).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(styles.Text)

	cursorStyle := lipgloss.NewStyle().
		Foreground(styles.Green).
		Bold(true)

	countStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0)

	prefix := prefixStyle.Render("/")

	var inputContent string
	if s.cursorPos >= len(s.value) {
		inputContent = inputStyle.Render(s.value) + cursorStyle.Render("█")
	} else {
		before := s.value[:s.cursorPos]
		after := s.value[s.cursorPos:]
		inputContent = inputStyle.Render(before) + cursorStyle.Render("█") + inputStyle.Render(after)
	}

	// Count display
	countText := ""
	if totalCount > 0 || matchCount > 0 {
		countText = countStyle.Render(fmt.Sprintf(" %d/%d", matchCount, totalCount))
	}

	// Calculate available width for input (no border, just prefix and count)
	countWidth := lipgloss.Width(countText)
	availableWidth := width - countWidth - 1 // -1 for prefix

	// Build content with input and count aligned right
	inputWidth := lipgloss.Width(inputContent)
	spacing := availableWidth - inputWidth - 1 // -1 for prefix
	if spacing < 0 {
		spacing = 0
	}

	return prefix + inputContent + strings.Repeat(" ", spacing) + countText
}

// MatchesQuery checks if a string matches the search query (case-insensitive)
func MatchesQuery(text, query string) bool {
	if query == "" {
		return true
	}
	return strings.Contains(strings.ToLower(text), strings.ToLower(query))
}
