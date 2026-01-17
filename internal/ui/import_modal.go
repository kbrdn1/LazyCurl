package ui

import (
	"errors"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/internal/api"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// ImportModalModel handles the cURL import modal
type ImportModalModel struct {
	textarea textarea.Model
	error    string
	visible  bool
	width    int
	height   int
}

// NewImportModal creates a new import modal
func NewImportModal() *ImportModalModel {
	ta := textarea.New()
	ta.Placeholder = "Paste your cURL command here...\n\nExample:\ncurl -X POST -H 'Content-Type: application/json' \\\n  -d '{\"name\": \"test\"}' \\\n  https://api.example.com/users"
	ta.ShowLineNumbers = false
	ta.CharLimit = 10000
	ta.SetWidth(60)
	ta.SetHeight(10)

	// Style the textarea
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(styles.Lavender).
		Padding(0, 1)
	ta.BlurredStyle.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(styles.Surface0).
		Padding(0, 1)

	return &ImportModalModel{
		textarea: ta,
		error:    "",
		visible:  false,
		width:    80,
		height:   20,
	}
}

// Show makes the modal visible and focuses the textarea
func (m *ImportModalModel) Show() {
	m.visible = true
	m.error = ""
	m.textarea.Reset()
	m.textarea.Focus()
}

// Hide hides the modal
func (m *ImportModalModel) Hide() {
	m.visible = false
	m.error = ""
	m.textarea.Blur()
}

// IsVisible returns whether the modal is visible
func (m *ImportModalModel) IsVisible() bool {
	return m.visible
}

// SetSize updates the modal dimensions
func (m *ImportModalModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate textarea size (modal content area)
	modalWidth := min(80, width-10)
	textareaWidth := modalWidth - 6 // Account for padding and borders
	textareaHeight := min(12, height-12)

	m.textarea.SetWidth(textareaWidth)
	m.textarea.SetHeight(textareaHeight)
}

// Update handles messages for the import modal
func (m *ImportModalModel) Update(msg tea.Msg) (*ImportModalModel, tea.Cmd) {
	if !m.visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Hide()
			return m, func() tea.Msg {
				return HideImportModalMsg{}
			}

		case "ctrl+enter", "ctrl+s":
			// Submit the cURL command
			input := strings.TrimSpace(m.textarea.Value())
			if input == "" {
				m.error = "Please enter a cURL command"
				return m, nil
			}

			// Parse the cURL command
			request, err := api.ParseCurlCommand(input)
			if err != nil {
				var parseErr *api.ParseError
				if errors.As(err, &parseErr) {
					m.error = parseErr.FormatWithContext()
				} else {
					m.error = err.Error()
				}
				return m, nil
			}

			// Success - hide modal and send import message
			m.Hide()
			return m, func() tea.Msg {
				return CurlImportedMsg{Request: request}
			}
		}
	}

	// Pass other messages to textarea
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

// View renders the import modal
func (m *ImportModalModel) View() string {
	if !m.visible {
		return ""
	}

	// Modal container styles
	modalWidth := min(80, m.width-10)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Lavender).
		MarginBottom(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		MarginTop(1)

	errorStyle := lipgloss.NewStyle().
		Foreground(styles.Red).
		Bold(true).
		MarginTop(1)

	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(styles.Lavender).
		Background(styles.Base)

	// Build content
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("📥 Import cURL Command"))
	content.WriteString("\n\n")

	// Textarea
	content.WriteString(m.textarea.View())

	// Error message if any
	if m.error != "" {
		content.WriteString("\n")
		content.WriteString(errorStyle.Render("⚠ " + m.error))
	}

	// Help text
	content.WriteString("\n")
	content.WriteString(helpStyle.Render("Ctrl+Enter: Import • Esc: Cancel"))

	return modalStyle.Render(content.String())
}

// GetValue returns the current textarea value
func (m *ImportModalModel) GetValue() string {
	return m.textarea.Value()
}

// SetError sets an error message to display
func (m *ImportModalModel) SetError(err string) {
	m.error = err
}

// ClearError clears the error message
func (m *ImportModalModel) ClearError() {
	m.error = ""
}
