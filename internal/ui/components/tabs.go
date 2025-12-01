package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// Tab represents a single tab
type Tab struct {
	Name   string
	Active bool
}

// Tabs represents a tab bar component
type Tabs struct {
	Items       []string
	ActiveIndex int
}

// NewTabs creates a new tabs component
func NewTabs(items []string) *Tabs {
	return &Tabs{
		Items:       items,
		ActiveIndex: 0,
	}
}

// Next moves to the next tab
func (t *Tabs) Next() {
	if t.ActiveIndex < len(t.Items)-1 {
		t.ActiveIndex++
	}
}

// Previous moves to the previous tab
func (t *Tabs) Previous() {
	if t.ActiveIndex > 0 {
		t.ActiveIndex--
	}
}

// SetActive sets the active tab by index
func (t *Tabs) SetActive(index int) {
	if index >= 0 && index < len(t.Items) {
		t.ActiveIndex = index
	}
}

// GetActive returns the name of the active tab
func (t *Tabs) GetActive() string {
	if t.ActiveIndex >= 0 && t.ActiveIndex < len(t.Items) {
		return t.Items[t.ActiveIndex]
	}
	return ""
}

// View renders the tabs
func (t *Tabs) View(width int) string {
	var tabs []string

	activeTabStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Lavender).
		Background(styles.Surface1).
		Padding(0, 1)

	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Background(styles.Mantle).
		Padding(0, 1)

	for i, item := range t.Items {
		if i == t.ActiveIndex {
			tabs = append(tabs, activeTabStyle.Render(item))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(item))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

// ViewWithBorder renders the tabs with a border
func (t *Tabs) ViewWithBorder(width int) string {
	tabBar := t.View(width)

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Surface0).
		Width(width - 2)

	return borderStyle.Render(tabBar)
}

// ViewSimple renders tabs without styling (for mockup)
func (t *Tabs) ViewSimple() string {
	var result strings.Builder

	for i, item := range t.Items {
		if i > 0 {
			result.WriteString(" | ")
		}

		if i == t.ActiveIndex {
			result.WriteString("[")
			result.WriteString(item)
			result.WriteString("]")
		} else {
			result.WriteString(item)
		}
	}

	return result.String()
}
