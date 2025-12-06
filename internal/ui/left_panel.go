package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/internal/config"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// LeftPanelTab represents the active tab in the left panel
type LeftPanelTab int

const (
	CollectionsTab LeftPanelTab = iota
	EnvironmentsTab
)

// LeftPanel wraps Collections and Environments views with tabs
type LeftPanel struct {
	activeTab    LeftPanelTab
	collections  *CollectionsView
	environments *EnvironmentsView
}

// NewLeftPanel creates a new left panel
func NewLeftPanel(workspacePath string) *LeftPanel {
	return &LeftPanel{
		activeTab:    CollectionsTab,
		collections:  NewCollectionsView(workspacePath),
		environments: NewEnvironmentsView(workspacePath),
	}
}

// GetActiveTab returns the currently active tab
func (l *LeftPanel) GetActiveTab() LeftPanelTab {
	return l.activeTab
}

// SetActiveTab sets the active tab
func (l *LeftPanel) SetActiveTab(tab LeftPanelTab) {
	l.activeTab = tab
}

// GetCollections returns the collections view
func (l *LeftPanel) GetCollections() *CollectionsView {
	return l.collections
}

// GetEnvironments returns the environments view
func (l *LeftPanel) GetEnvironments() *EnvironmentsView {
	return l.environments
}

// Update handles messages for the left panel
func (l LeftPanel) Update(msg tea.Msg, cfg *config.GlobalConfig) (LeftPanel, tea.Cmd) {
	var cmd tea.Cmd

	switch l.activeTab {
	case CollectionsTab:
		*l.collections, cmd = l.collections.Update(msg, cfg)
	case EnvironmentsTab:
		*l.environments, cmd = l.environments.Update(msg, cfg)
	}

	return l, cmd
}

// View renders the left panel content (without title bar - that's in renderPanel)
func (l LeftPanel) View(width, height int, active bool) string {
	switch l.activeTab {
	case CollectionsTab:
		return l.collections.View(width, height, active)
	case EnvironmentsTab:
		return l.environments.View(width, height, active)
	default:
		return l.collections.View(width, height, active)
	}
}

// IsSearching returns true if the active tab has search input visible
func (l *LeftPanel) IsSearching() bool {
	switch l.activeTab {
	case CollectionsTab:
		return l.collections.GetTree().IsSearching()
	case EnvironmentsTab:
		return l.environments.IsSearching()
	default:
		return false
	}
}

// HasSearchQuery returns true if the active tab has an active search query (not input)
func (l *LeftPanel) HasSearchQuery() bool {
	switch l.activeTab {
	case CollectionsTab:
		return l.collections.GetTree().HasSearchQuery()
	case EnvironmentsTab:
		return l.environments.HasSearchQuery()
	default:
		return false
	}
}

// RenderTabs renders the tab bar for the panel title
func (l LeftPanel) RenderTabs(width int, active bool, borderColor lipgloss.Color) string {
	var activeColor, inactiveColor lipgloss.Color
	if active {
		activeColor = styles.Lavender
	} else {
		activeColor = styles.Subtext0
	}
	inactiveColor = styles.Subtext0

	// Tab styles
	activeTabStyle := lipgloss.NewStyle().
		Foreground(activeColor).
		Bold(true)

	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(inactiveColor)

	borderStyle := lipgloss.NewStyle().
		Foreground(borderColor)

	// Render tabs
	var collectionsTab, envTab string
	if l.activeTab == CollectionsTab {
		collectionsTab = activeTabStyle.Render("Collections")
		envTab = inactiveTabStyle.Render("Envs")
	} else {
		collectionsTab = inactiveTabStyle.Render("Collections")
		envTab = activeTabStyle.Render("Envs")
	}

	// Format: "─Collections─Env─────────"
	// Calculate actual text widths (without ANSI codes)
	collectionsWidth := lipgloss.Width(collectionsTab)
	envWidth := lipgloss.Width(envTab)

	// Total used: 1 (prefix ─) + collectionsWidth + 1 (separator ─) + envWidth
	usedWidth := 1 + collectionsWidth + 1 + envWidth
	remainingWidth := width - usedWidth
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	return borderStyle.Render("─") + collectionsTab + borderStyle.Render("─") + envTab + borderStyle.Render(strings.Repeat("─", remainingWidth))
}
