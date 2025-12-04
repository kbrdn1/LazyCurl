package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// ModalType represents the type of modal
type ModalType int

const (
	ModalConfirm ModalType = iota
	ModalInput
	ModalForm
)

// ModalResult represents the result of a modal action
type ModalResult struct {
	Confirmed bool
	Values    map[string]interface{}
}

// ModalCloseMsg is sent when a modal is closed
type ModalCloseMsg struct {
	Result ModalResult
	Tag    string // Identifier for which modal closed
}

// FormField represents a field in a form modal
type FormField struct {
	Name        string
	Label       string
	Value       string
	Type        string   // "text", "checkbox", "radio"
	Options     []string // For radio buttons
	Placeholder string
	CursorPos   int // Cursor position for text fields
}

// Modal represents a modal dialog
type Modal struct {
	Title       string
	Message     string
	Type        ModalType
	Tag         string // Identifier for this modal
	Visible     bool
	Fields      []FormField
	FocusIndex  int
	ConfirmText string
	CancelText  string
	Width       int
	Height      int
}

// NewConfirmModal creates a confirmation modal
func NewConfirmModal(title, message, tag string) *Modal {
	return &Modal{
		Title:       title,
		Message:     message,
		Type:        ModalConfirm,
		Tag:         tag,
		Visible:     false,
		ConfirmText: "Yes",
		CancelText:  "No",
		FocusIndex:  1, // Focus on "No" by default for safety
		Width:       50,
		Height:      8,
	}
}

// NewInputModal creates a single input modal
func NewInputModal(title, label, placeholder, tag string) *Modal {
	return &Modal{
		Title:   title,
		Type:    ModalInput,
		Tag:     tag,
		Visible: false,
		Fields: []FormField{
			{Name: "input", Label: label, Type: "text", Placeholder: placeholder},
		},
		ConfirmText: "OK",
		CancelText:  "Cancel",
		FocusIndex:  0,
		Width:       50,
		Height:      10,
	}
}

// NewFormModal creates a form modal with multiple fields
func NewFormModal(title, tag string, fields []FormField) *Modal {
	return &Modal{
		Title:       title,
		Type:        ModalForm,
		Tag:         tag,
		Visible:     false,
		Fields:      fields,
		ConfirmText: "Save",
		CancelText:  "Cancel",
		FocusIndex:  0,
		Width:       60,
		Height:      12 + len(fields)*2,
	}
}

// Show displays the modal
func (m *Modal) Show() {
	m.Visible = true
	m.FocusIndex = 0
}

// Hide hides the modal
func (m *Modal) Hide() {
	m.Visible = false
}

// IsVisible returns whether the modal is visible
func (m *Modal) IsVisible() bool {
	return m.Visible
}

// SetFieldValue sets a field value by name
func (m *Modal) SetFieldValue(name, value string) {
	for i := range m.Fields {
		if m.Fields[i].Name == name {
			m.Fields[i].Value = value
			m.Fields[i].CursorPos = len(value) // Position cursor at end
			return
		}
	}
}

// GetFieldValue gets a field value by name
func (m *Modal) GetFieldValue(name string) string {
	for _, f := range m.Fields {
		if f.Name == name {
			return f.Value
		}
	}
	return ""
}

// Update handles messages for the modal
func (m *Modal) Update(msg tea.Msg) (*Modal, tea.Cmd) {
	if !m.Visible {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Hide()
			return m, func() tea.Msg {
				return ModalCloseMsg{Result: ModalResult{Confirmed: false}, Tag: m.Tag}
			}

		case "enter":
			if m.Type == ModalConfirm {
				confirmed := m.FocusIndex == 0
				m.Hide()
				return m, func() tea.Msg {
					return ModalCloseMsg{Result: ModalResult{Confirmed: confirmed}, Tag: m.Tag}
				}
			}
			// For form/input modals, Enter always submits with Confirmed: true
			// (unless focused on Cancel button)
			confirmed := m.FocusIndex != len(m.Fields)+1 // Cancel is at index len(Fields)+1
			m.Hide()
			values := make(map[string]interface{})
			for _, f := range m.Fields {
				if f.Type == "checkbox" {
					values[f.Name] = f.Value == "true"
				} else {
					values[f.Name] = f.Value
				}
			}
			return m, func() tea.Msg {
				return ModalCloseMsg{Result: ModalResult{Confirmed: confirmed, Values: values}, Tag: m.Tag}
			}

		case "tab", "down":
			if m.Type == ModalConfirm {
				m.FocusIndex = (m.FocusIndex + 1) % 2
			} else {
				// Fields + 2 buttons
				m.FocusIndex = (m.FocusIndex + 1) % (len(m.Fields) + 2)
			}

		case "j":
			// When on buttons or confirm modal, navigate down; when on text field, type 'j'
			if m.Type == ModalConfirm {
				m.FocusIndex = (m.FocusIndex + 1) % 2
			} else if m.FocusIndex >= len(m.Fields) {
				m.FocusIndex = (m.FocusIndex + 1) % (len(m.Fields) + 2)
			} else if m.FocusIndex < len(m.Fields) && m.Fields[m.FocusIndex].Type == "text" {
				m.insertCharAtCursor(m.FocusIndex, "j")
			}

		case "shift+tab", "up":
			if m.Type == ModalConfirm {
				m.FocusIndex = (m.FocusIndex + 1) % 2
			} else {
				m.FocusIndex--
				if m.FocusIndex < 0 {
					m.FocusIndex = len(m.Fields) + 1
				}
			}

		case "k":
			// When on buttons or confirm modal, navigate up; when on text field, type 'k'
			if m.Type == ModalConfirm {
				m.FocusIndex = (m.FocusIndex + 1) % 2
			} else if m.FocusIndex >= len(m.Fields) {
				m.FocusIndex--
				if m.FocusIndex < 0 {
					m.FocusIndex = len(m.Fields) + 1
				}
			} else if m.FocusIndex < len(m.Fields) && m.Fields[m.FocusIndex].Type == "text" {
				m.insertCharAtCursor(m.FocusIndex, "k")
			}

		case "left":
			// Arrow left moves cursor in text field
			if m.FocusIndex < len(m.Fields) && m.Fields[m.FocusIndex].Type == "text" {
				if m.Fields[m.FocusIndex].CursorPos > 0 {
					m.Fields[m.FocusIndex].CursorPos--
				}
			} else if m.Type == ModalConfirm || m.FocusIndex >= len(m.Fields) {
				// Navigate buttons only with arrows when on buttons
				if m.FocusIndex > 0 {
					m.FocusIndex--
				}
			}

		case "h":
			// When on buttons, navigate left; when on text field, type 'h'
			if m.Type == ModalConfirm || m.FocusIndex >= len(m.Fields) {
				if m.FocusIndex > 0 {
					m.FocusIndex--
				}
			} else if m.FocusIndex < len(m.Fields) && m.Fields[m.FocusIndex].Type == "text" {
				m.insertCharAtCursor(m.FocusIndex, "h")
			}

		case "right":
			// Arrow right moves cursor in text field
			if m.FocusIndex < len(m.Fields) && m.Fields[m.FocusIndex].Type == "text" {
				if m.Fields[m.FocusIndex].CursorPos < len(m.Fields[m.FocusIndex].Value) {
					m.Fields[m.FocusIndex].CursorPos++
				}
			} else if m.Type == ModalConfirm {
				m.FocusIndex = (m.FocusIndex + 1) % 2
			} else if m.FocusIndex >= len(m.Fields) {
				if m.FocusIndex < len(m.Fields)+1 {
					m.FocusIndex++
				}
			}

		case "l":
			// When on buttons, navigate right; when on text field, type 'l'
			if m.Type == ModalConfirm {
				m.FocusIndex = (m.FocusIndex + 1) % 2
			} else if m.FocusIndex >= len(m.Fields) {
				if m.FocusIndex < len(m.Fields)+1 {
					m.FocusIndex++
				}
			} else if m.FocusIndex < len(m.Fields) && m.Fields[m.FocusIndex].Type == "text" {
				m.insertCharAtCursor(m.FocusIndex, "l")
			}

		case " ":
			// Toggle checkbox
			if m.FocusIndex < len(m.Fields) && m.Fields[m.FocusIndex].Type == "checkbox" {
				if m.Fields[m.FocusIndex].Value == "true" {
					m.Fields[m.FocusIndex].Value = "false"
				} else {
					m.Fields[m.FocusIndex].Value = "true"
				}
			}

		case "backspace":
			if m.FocusIndex < len(m.Fields) && m.Fields[m.FocusIndex].Type == "text" {
				v := m.Fields[m.FocusIndex].Value
				pos := m.Fields[m.FocusIndex].CursorPos
				if pos > 0 && len(v) > 0 {
					m.Fields[m.FocusIndex].Value = v[:pos-1] + v[pos:]
					m.Fields[m.FocusIndex].CursorPos--
				}
			}

		default:
			// Text input
			if m.FocusIndex < len(m.Fields) && m.Fields[m.FocusIndex].Type == "text" {
				if len(msg.String()) == 1 {
					m.insertCharAtCursor(m.FocusIndex, msg.String())
				}
			}
		}
	}

	return m, nil
}

// insertCharAtCursor inserts a character at the cursor position in the specified field
func (m *Modal) insertCharAtCursor(fieldIndex int, char string) {
	if fieldIndex >= len(m.Fields) {
		return
	}
	field := &m.Fields[fieldIndex]
	pos := field.CursorPos
	field.Value = field.Value[:pos] + char + field.Value[pos:]
	field.CursorPos++
}

// View renders the modal
func (m *Modal) View(screenWidth, screenHeight int) string {
	if !m.Visible {
		return ""
	}

	// Calculate modal dimensions
	width := m.Width
	if width > screenWidth-4 {
		width = screenWidth - 4
	}

	// Build modal content
	var content strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Lavender).
		Width(width - 4).
		Align(lipgloss.Center)
	content.WriteString(titleStyle.Render(m.Title))
	content.WriteString("\n\n")

	// Message (for confirm modals)
	if m.Message != "" {
		msgStyle := lipgloss.NewStyle().
			Foreground(styles.Text).
			Width(width - 4).
			Align(lipgloss.Center)
		content.WriteString(msgStyle.Render(m.Message))
		content.WriteString("\n\n")
	}

	// Fields
	for i, field := range m.Fields {
		focused := i == m.FocusIndex

		labelStyle := lipgloss.NewStyle().Foreground(styles.Subtext1)
		content.WriteString(labelStyle.Render(field.Label + ": "))

		switch field.Type {
		case "text":
			inputStyle := lipgloss.NewStyle().
				Foreground(styles.Text).
				Background(styles.Surface0).
				Width(width - len(field.Label) - 8)
			if focused {
				inputStyle = inputStyle.Background(styles.Surface1)
			}
			displayVal := field.Value
			if displayVal == "" && field.Placeholder != "" {
				displayVal = field.Placeholder
				inputStyle = inputStyle.Foreground(styles.Subtext0)
				if focused {
					displayVal = "▌" + displayVal[1:] // Cursor at start for placeholder
					if len(field.Placeholder) == 0 {
						displayVal = "▌"
					}
				}
			} else if focused {
				// Insert cursor at cursor position
				pos := m.Fields[i].CursorPos
				if pos >= len(displayVal) {
					displayVal += "▌"
				} else {
					displayVal = displayVal[:pos] + "▌" + displayVal[pos+1:]
				}
			}
			content.WriteString(inputStyle.Render(displayVal))

		case "checkbox":
			checkbox := "[ ]"
			checkStyle := lipgloss.NewStyle().Foreground(styles.CheckboxOff)
			if field.Value == "true" {
				checkbox = "[x]"
				checkStyle = checkStyle.Foreground(styles.CheckboxOn)
			}
			if focused {
				checkStyle = checkStyle.Background(styles.Surface1).Bold(true)
			}
			content.WriteString(checkStyle.Render(checkbox))
		}
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Buttons
	var buttons []string

	confirmStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Foreground(styles.Text).
		Background(styles.Surface0)
	cancelStyle := confirmStyle.Copy()

	confirmIdx := len(m.Fields)
	cancelIdx := len(m.Fields) + 1
	if m.Type == ModalConfirm {
		confirmIdx = 0
		cancelIdx = 1
	}

	if m.FocusIndex == confirmIdx {
		confirmStyle = confirmStyle.
			Background(styles.Lavender).
			Foreground(styles.Base).
			Bold(true)
	}
	if m.FocusIndex == cancelIdx {
		cancelStyle = cancelStyle.
			Background(styles.Red).
			Foreground(styles.Base).
			Bold(true)
	}

	buttons = append(buttons, confirmStyle.Render(m.ConfirmText))
	buttons = append(buttons, cancelStyle.Render(m.CancelText))

	buttonRow := lipgloss.JoinHorizontal(lipgloss.Center, buttons[0], "  ", buttons[1])
	buttonRowStyle := lipgloss.NewStyle().Width(width - 4).Align(lipgloss.Center)
	content.WriteString(buttonRowStyle.Render(buttonRow))

	// Modal box - transparent background, only border
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Lavender).
		Padding(1, 2).
		Width(width)

	// Return just the modal box, centering is handled by caller
	return modalStyle.Render(content.String())
}
