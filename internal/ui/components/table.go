package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// KeyValuePair represents a key-value pair
type KeyValuePair struct {
	Key   string
	Value string
}

// Table represents an editable table component
type Table struct {
	Headers []string
	Rows    []KeyValuePair
	Cursor  int
	Editing bool
	EditCol int // 0 for key, 1 for value
}

// NewTable creates a new table
func NewTable(headers []string) *Table {
	return &Table{
		Headers: headers,
		Rows:    []KeyValuePair{},
		Cursor:  -1,
		Editing: false,
		EditCol: 0,
	}
}

// AddRow adds a new row to the table
func (t *Table) AddRow(key, value string) {
	t.Rows = append(t.Rows, KeyValuePair{Key: key, Value: value})
}

// DeleteRow removes a row from the table
func (t *Table) DeleteRow(index int) {
	if index >= 0 && index < len(t.Rows) {
		t.Rows = append(t.Rows[:index], t.Rows[index+1:]...)
		if t.Cursor >= len(t.Rows) {
			t.Cursor = len(t.Rows) - 1
		}
	}
}

// MoveUp moves cursor up
func (t *Table) MoveUp() {
	if t.Cursor > 0 {
		t.Cursor--
	}
}

// MoveDown moves cursor down
func (t *Table) MoveDown() {
	if t.Cursor < len(t.Rows)-1 {
		t.Cursor++
	}
}

// ToggleEdit toggles edit mode
func (t *Table) ToggleEdit() {
	t.Editing = !t.Editing
}

// View renders the table
func (t *Table) View(width, height int) string {
	if len(t.Rows) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Render("No entries. Press 'a' to add.")
	}

	var result strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Width(width / 2)

	result.WriteString(headerStyle.Render(t.Headers[0]))
	result.WriteString(" ")
	result.WriteString(headerStyle.Render(t.Headers[1]))
	result.WriteString("\n")

	// Separator
	result.WriteString(strings.Repeat("─", width))
	result.WriteString("\n")

	// Rows
	rowStyle := lipgloss.NewStyle().Width(width / 2)
	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#3C3C3C")).
		Width(width / 2)

	for i, row := range t.Rows {
		cursor := "  "
		if i == t.Cursor {
			cursor = "▶ "
		}

		keyStyle := rowStyle
		valueStyle := rowStyle

		if i == t.Cursor {
			keyStyle = selectedStyle
			valueStyle = selectedStyle
		}

		result.WriteString(cursor)
		result.WriteString(keyStyle.Render(row.Key))
		result.WriteString(" ")
		result.WriteString(valueStyle.Render(row.Value))
		result.WriteString("\n")
	}

	return result.String()
}

// ViewSimple renders a simple table view
func (t *Table) ViewSimple() string {
	if len(t.Rows) == 0 {
		return "No entries"
	}

	var result strings.Builder
	for _, row := range t.Rows {
		result.WriteString(fmt.Sprintf("%s: %s\n", row.Key, row.Value))
	}
	return result.String()
}

// ToMap converts table rows to a map
func (t *Table) ToMap() map[string]string {
	result := make(map[string]string)
	for _, row := range t.Rows {
		result[row.Key] = row.Value
	}
	return result
}

// FromMap populates table from a map
func (t *Table) FromMap(data map[string]string) {
	t.Rows = []KeyValuePair{}
	for key, value := range data {
		t.Rows = append(t.Rows, KeyValuePair{Key: key, Value: value})
	}
}
