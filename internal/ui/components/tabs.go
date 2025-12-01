package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// Tab represents a single tab
type Tab struct {
	Name   string
	Active bool
}

// TabItem represents a tab with optional shortcut
type TabItem struct {
	Name     string
	Shortcut string // e.g., "Shift+1"
}

// Tabs represents a tab bar component
type Tabs struct {
	Items       []string
	Shortcuts   []string // Keyboard shortcuts for each tab
	ActiveIndex int
}

// NewTabs creates a new tabs component
func NewTabs(items []string) *Tabs {
	return &Tabs{
		Items:       items,
		Shortcuts:   make([]string, len(items)),
		ActiveIndex: 0,
	}
}

// NewTabsWithShortcuts creates a new tabs component with keyboard shortcuts
func NewTabsWithShortcuts(items []TabItem) *Tabs {
	names := make([]string, len(items))
	shortcuts := make([]string, len(items))
	for i, item := range items {
		names[i] = item.Name
		shortcuts[i] = item.Shortcut
	}
	return &Tabs{
		Items:       names,
		Shortcuts:   shortcuts,
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

// ViewWithShortcuts renders tabs with keyboard shortcuts displayed
func (t *Tabs) ViewWithShortcuts(width int) string {
	var tabs []string

	activeNameStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Lavender)

	inactiveNameStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0)

	shortcutStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Italic(true)

	activeShortcutStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Italic(true)

	separatorStyle := lipgloss.NewStyle().
		Foreground(styles.Surface0)

	for i, item := range t.Items {
		var tabContent string
		shortcut := ""
		if i < len(t.Shortcuts) && t.Shortcuts[i] != "" {
			shortcut = t.Shortcuts[i]
		}

		if i == t.ActiveIndex {
			if shortcut != "" {
				tabContent = activeNameStyle.Render(item) + " " + activeShortcutStyle.Render(shortcut)
			} else {
				tabContent = activeNameStyle.Render(item)
			}
		} else {
			if shortcut != "" {
				tabContent = inactiveNameStyle.Render(item) + " " + shortcutStyle.Render(shortcut)
			} else {
				tabContent = inactiveNameStyle.Render(item)
			}
		}

		tabs = append(tabs, tabContent)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		strings.Join(tabs, separatorStyle.Render("  │  ")),
	)
}

// ViewCompact renders tabs in a compact format with shortcuts
func (t *Tabs) ViewCompact(width int) string {
	var tabs []string

	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Lavender).
		Background(styles.Surface0).
		Padding(0, 1)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Padding(0, 1)

	for i, item := range t.Items {
		// Create tab label with shortcut number
		label := item
		if i < len(t.Shortcuts) && t.Shortcuts[i] != "" {
			// Extract just the number from shortcut (e.g., "!" from "Shift+1")
			label = fmt.Sprintf("%s %s", item, t.Shortcuts[i])
		}

		if i == t.ActiveIndex {
			tabs = append(tabs, activeStyle.Render(label))
		} else {
			tabs = append(tabs, inactiveStyle.Render(label))
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
