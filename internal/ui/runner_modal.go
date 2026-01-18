package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/internal/api"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// Loader animation frames
var loaderFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// RunnerModal displays the collection runner progress and results.
type RunnerModal struct {
	visible      bool
	session      *api.RunSession
	requests     []*api.CollectionRequest
	width        int
	height       int
	scrollOffset int
	loaderFrame  int
	exported     bool
	exportPath   string
	exportError  string
}

// NewRunnerModal creates a new runner modal.
func NewRunnerModal() *RunnerModal {
	return &RunnerModal{
		visible:      false,
		session:      nil,
		requests:     nil,
		width:        80,
		height:       24,
		scrollOffset: 0,
		loaderFrame:  0,
	}
}

// Show makes the modal visible with the given session.
func (m *RunnerModal) Show(session *api.RunSession, requests []*api.CollectionRequest) {
	m.visible = true
	m.session = session
	m.requests = requests
	m.scrollOffset = 0
	m.loaderFrame = 0
	m.exported = false
	m.exportPath = ""
	m.exportError = ""
}

// Hide hides the modal.
func (m *RunnerModal) Hide() {
	m.visible = false
	m.session = nil
	m.requests = nil
	m.scrollOffset = 0
}

// IsVisible returns whether the modal is visible.
func (m *RunnerModal) IsVisible() bool {
	return m.visible
}

// SetSize updates the modal dimensions.
func (m *RunnerModal) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// UpdateSession updates the session data.
func (m *RunnerModal) UpdateSession(session *api.RunSession) {
	m.session = session
}

// SetExported marks the export as complete.
func (m *RunnerModal) SetExported(path string, err error) {
	m.exported = true
	if err != nil {
		m.exportError = err.Error()
		m.exportPath = ""
	} else {
		m.exportPath = path
		m.exportError = ""
	}
}

// AdvanceLoader advances the loader animation frame.
func (m *RunnerModal) AdvanceLoader() {
	m.loaderFrame = (m.loaderFrame + 1) % len(loaderFrames)
}

// Update handles messages for the runner modal.
func (m *RunnerModal) Update(msg tea.Msg) (*RunnerModal, tea.Cmd) {
	if !m.visible || m.session == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			if m.session.IsTerminal() {
				// Close modal when run is finished
				m.Hide()
				return m, func() tea.Msg {
					return RunnerHideModalMsg{}
				}
			}
			// Cancel running execution
			return m, func() tea.Msg {
				return RunnerCancelMsg{}
			}

		case "e":
			if m.session.IsTerminal() && !m.exported {
				return m, func() tea.Msg {
					return RunnerExportMsg{}
				}
			}

		case "enter":
			if m.session.IsTerminal() {
				m.Hide()
				return m, func() tea.Msg {
					return RunnerHideModalMsg{}
				}
			}

		case "j", "down":
			m.scrollDown()

		case "k", "up":
			m.scrollUp()

		case "g":
			m.scrollOffset = 0

		case "G":
			m.scrollToBottom()
		}

	case RunnerTickMsg:
		m.AdvanceLoader()
	}

	return m, nil
}

// scrollDown scrolls the result list down.
func (m *RunnerModal) scrollDown() {
	maxScroll := m.maxScrollOffset()
	if m.scrollOffset < maxScroll {
		m.scrollOffset++
	}
}

// scrollUp scrolls the result list up.
func (m *RunnerModal) scrollUp() {
	if m.scrollOffset > 0 {
		m.scrollOffset--
	}
}

// scrollToBottom scrolls to the bottom of the list.
func (m *RunnerModal) scrollToBottom() {
	m.scrollOffset = m.maxScrollOffset()
}

// maxScrollOffset returns the maximum scroll offset.
func (m *RunnerModal) maxScrollOffset() int {
	if m.session == nil {
		return 0
	}
	visibleRows := m.visibleResultRows()
	total := len(m.session.Results)
	if total <= visibleRows {
		return 0
	}
	return total - visibleRows
}

// visibleResultRows returns the number of result rows that can be displayed.
func (m *RunnerModal) visibleResultRows() int {
	// Use effective modal height (capped) instead of terminal height
	effectiveHeight := min(m.height-4, 30)
	// Subtract header, progress bar, summary, footer, borders
	available := effectiveHeight - 12
	if available < 3 {
		return 3
	}
	return available
}

// View renders the runner modal.
func (m *RunnerModal) View() string {
	if !m.visible || m.session == nil {
		return ""
	}

	// Calculate modal dimensions
	modalWidth := min(90, m.width-6)
	modalHeight := min(m.height-4, 30)

	// Build content
	var content strings.Builder

	// Header
	content.WriteString(m.renderHeader(modalWidth))
	content.WriteString("\n\n")

	// Progress or status
	content.WriteString(m.renderProgress(modalWidth))
	content.WriteString("\n\n")

	// Summary stats
	content.WriteString(m.renderSummary(modalWidth))
	content.WriteString("\n\n")

	// Results list
	content.WriteString(m.renderResults(modalWidth))

	// Export status
	if m.exported {
		content.WriteString("\n")
		content.WriteString(m.renderExportStatus())
	}

	// Footer
	content.WriteString("\n\n")
	content.WriteString(m.renderFooter())

	// Modal container style
	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(styles.Lavender).
		Background(styles.Base)

	modalContent := modalStyle.Render(content.String())

	// Center the modal
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		modalContent,
	)
}

// renderHeader renders the modal header.
func (m *RunnerModal) renderHeader(width int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Lavender)

	collectionStyle := lipgloss.NewStyle().
		Foreground(styles.Text)

	title := titleStyle.Render("▶ Collection Runner")

	collectionName := m.session.Collection
	if len(m.session.FolderPath) > 0 {
		collectionName = strings.Join(m.session.FolderPath, " / ")
	}

	return title + " " + collectionStyle.Render(collectionName)
}

// renderProgress renders the progress bar or final status.
func (m *RunnerModal) renderProgress(width int) string {
	if m.session.IsTerminal() {
		return m.renderFinalStatus()
	}

	// Animated loader
	loader := loaderFrames[m.loaderFrame]

	// Progress text
	progress := m.session.Progress()

	progressStyle := lipgloss.NewStyle().
		Foreground(styles.Blue)

	loaderStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender)

	// Progress bar
	barWidth := width - 20
	if barWidth < 10 {
		barWidth = 10
	}

	// Guard against divide by zero when TotalRequests is 0
	percent := 0.0
	if m.session.TotalRequests > 0 {
		percent = float64(m.session.CurrentIndex) / float64(m.session.TotalRequests)
	}
	if percent < 0 {
		percent = 0
	} else if percent > 1 {
		percent = 1
	}
	filled := int(float64(barWidth) * percent)
	empty := barWidth - filled

	barStyle := lipgloss.NewStyle().Foreground(styles.Green)
	emptyStyle := lipgloss.NewStyle().Foreground(styles.Surface1)

	bar := barStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty))

	return fmt.Sprintf("%s %s [%s]",
		loaderStyle.Render(loader),
		bar,
		progressStyle.Render(progress))
}

// renderFinalStatus renders the final run status.
func (m *RunnerModal) renderFinalStatus() string {
	var icon string
	var statusStyle lipgloss.Style
	var statusText string

	switch m.session.Status {
	case api.RunStatusCompleted:
		icon = "✓"
		statusStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.Green)
		statusText = "Completed"
	case api.RunStatusCancelled:
		icon = "○"
		statusStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.Yellow)
		statusText = "Canceled"
	case api.RunStatusStopped:
		icon = "✗"
		statusStyle = lipgloss.NewStyle().Bold(true).Foreground(styles.Red)
		statusText = "Stopped (failure)"
	default:
		icon = "?"
		statusStyle = lipgloss.NewStyle().Foreground(styles.Subtext0)
		statusText = string(m.session.Status)
	}

	duration := m.session.EndTime.Sub(m.session.StartTime).Round(time.Millisecond)

	return fmt.Sprintf("%s %s in %s",
		statusStyle.Render(icon),
		statusStyle.Render(statusText),
		lipgloss.NewStyle().Foreground(styles.Subtext0).Render(duration.String()))
}

// renderSummary renders the summary statistics.
func (m *RunnerModal) renderSummary(width int) string {
	passedStyle := lipgloss.NewStyle().Foreground(styles.Green)
	failedStyle := lipgloss.NewStyle().Foreground(styles.Red)
	errorStyle := lipgloss.NewStyle().Foreground(styles.Peach)
	skippedStyle := lipgloss.NewStyle().Foreground(styles.Subtext0)
	labelStyle := lipgloss.NewStyle().Foreground(styles.Subtext1)

	var passed, failed, errors, skipped int
	var totalAssertions, passedAssertions, failedAssertions int

	for _, r := range m.session.Results {
		switch r.Status {
		case api.ResultStatusPassed:
			passed++
			p, f := r.AssertionCount()
			passedAssertions += p
			failedAssertions += f
			totalAssertions += p + f
		case api.ResultStatusFailed:
			failed++
			p, f := r.AssertionCount()
			passedAssertions += p
			failedAssertions += f
			totalAssertions += p + f
		case api.ResultStatusError:
			errors++
		case api.ResultStatusSkipped:
			skipped++
		}
	}

	// Requests summary
	requestsSummary := fmt.Sprintf("%s %s  %s %s  %s %s  %s %s",
		passedStyle.Render(fmt.Sprintf("%d", passed)),
		labelStyle.Render("passed"),
		failedStyle.Render(fmt.Sprintf("%d", failed)),
		labelStyle.Render("failed"),
		errorStyle.Render(fmt.Sprintf("%d", errors)),
		labelStyle.Render("errors"),
		skippedStyle.Render(fmt.Sprintf("%d", skipped)),
		labelStyle.Render("skipped"))

	// Assertions summary (if any)
	var assertionsSummary string
	if totalAssertions > 0 {
		assertionsSummary = fmt.Sprintf("\n%s: %s %s  %s %s",
			labelStyle.Render("Assertions"),
			passedStyle.Render(fmt.Sprintf("%d", passedAssertions)),
			labelStyle.Render("passed"),
			failedStyle.Render(fmt.Sprintf("%d", failedAssertions)),
			labelStyle.Render("failed"))
	}

	return requestsSummary + assertionsSummary
}

// renderResults renders the list of request results.
func (m *RunnerModal) renderResults(width int) string {
	if len(m.session.Results) == 0 && m.session.TotalRequests > 0 {
		// Show pending requests
		return m.renderPendingList(width)
	}

	visibleRows := m.visibleResultRows()
	var lines []string

	// Add completed results
	start := m.scrollOffset
	end := min(start+visibleRows, len(m.session.Results))

	for i := start; i < end; i++ {
		lines = append(lines, m.renderResultRow(m.session.Results[i], width))
	}

	// Show current executing request if not all complete
	if m.session.CurrentIndex < m.session.TotalRequests && len(lines) < visibleRows {
		if m.session.CurrentIndex < len(m.requests) {
			req := m.requests[m.session.CurrentIndex]
			runningStyle := lipgloss.NewStyle().Foreground(styles.Blue)
			loader := loaderFrames[m.loaderFrame]
			lines = append(lines, fmt.Sprintf("%s %s %s",
				runningStyle.Render(loader),
				runningStyle.Render(string(req.Method)),
				runningStyle.Render(req.Name)))
		}
	}

	// Scroll indicator
	if m.maxScrollOffset() > 0 {
		scrollInfo := lipgloss.NewStyle().Foreground(styles.Subtext0).
			Render(fmt.Sprintf(" [%d-%d of %d]", start+1, end, len(m.session.Results)))
		if len(lines) > 0 {
			lines[len(lines)-1] += scrollInfo
		}
	}

	return strings.Join(lines, "\n")
}

// renderPendingList renders the list of pending requests before run starts.
func (m *RunnerModal) renderPendingList(width int) string {
	pendingStyle := lipgloss.NewStyle().Foreground(styles.Subtext0)
	var lines []string

	visibleRows := m.visibleResultRows()
	end := min(visibleRows, len(m.requests))

	for i := 0; i < end; i++ {
		req := m.requests[i]
		lines = append(lines, fmt.Sprintf("%s ○ %s %s",
			pendingStyle.Render(""),
			pendingStyle.Render(string(req.Method)),
			pendingStyle.Render(req.Name)))
	}

	if len(m.requests) > visibleRows {
		lines = append(lines, pendingStyle.Render(
			fmt.Sprintf("  ... and %d more", len(m.requests)-visibleRows)))
	}

	return strings.Join(lines, "\n")
}

// renderResultRow renders a single result row.
func (m *RunnerModal) renderResultRow(r api.RequestResult, width int) string {
	var icon string
	var nameStyle lipgloss.Style

	switch r.Status {
	case api.ResultStatusPassed:
		icon = "✓"
		nameStyle = lipgloss.NewStyle().Foreground(styles.Green)
	case api.ResultStatusFailed:
		icon = "✗"
		nameStyle = lipgloss.NewStyle().Foreground(styles.Red)
	case api.ResultStatusError:
		icon = "⚠"
		nameStyle = lipgloss.NewStyle().Foreground(styles.Peach)
	case api.ResultStatusSkipped:
		icon = "○"
		nameStyle = lipgloss.NewStyle().Foreground(styles.Subtext0)
	case api.ResultStatusRunning:
		icon = loaderFrames[m.loaderFrame]
		nameStyle = lipgloss.NewStyle().Foreground(styles.Blue)
	default:
		icon = "○"
		nameStyle = lipgloss.NewStyle().Foreground(styles.Subtext0)
	}

	iconStyle := nameStyle

	// Build the line
	method := r.Request.Method
	name := r.Request.Name

	// Status code if available
	var statusInfo string
	if r.Response != nil {
		statusStyle := httpStatusStyle(r.Response.Status)
		statusInfo = statusStyle.Render(fmt.Sprintf(" %d", r.Response.Status))
		if r.Response.TimeMs > 0 {
			statusInfo += lipgloss.NewStyle().Foreground(styles.Subtext0).
				Render(fmt.Sprintf(" %dms", r.Response.TimeMs))
		}
	}

	// Error info if present
	var errorInfo string
	if r.Error != nil {
		errorInfo = lipgloss.NewStyle().Foreground(styles.Red).
			Render(fmt.Sprintf(" (%s)", r.Error.Message))
	}

	// Assertion info
	var assertionInfo string
	if r.PostScriptResult != nil {
		passed, failed := r.AssertionCount()
		if passed+failed > 0 {
			if failed > 0 {
				assertionInfo = lipgloss.NewStyle().Foreground(styles.Red).
					Render(fmt.Sprintf(" [%d/%d tests]", passed, passed+failed))
			} else {
				assertionInfo = lipgloss.NewStyle().Foreground(styles.Green).
					Render(fmt.Sprintf(" [%d tests]", passed))
			}
		}
	}

	return fmt.Sprintf("%s %s %s%s%s%s",
		iconStyle.Render(icon),
		nameStyle.Render(method),
		nameStyle.Render(name),
		statusInfo,
		errorInfo,
		assertionInfo)
}

// renderExportStatus renders the export status message.
func (m *RunnerModal) renderExportStatus() string {
	if m.exportError != "" {
		return lipgloss.NewStyle().Foreground(styles.Red).
			Render("⚠ Export failed: " + m.exportError)
	}
	return lipgloss.NewStyle().Foreground(styles.Green).
		Render("✓ Exported to: " + m.exportPath)
}

// renderFooter renders the modal footer with keybindings.
func (m *RunnerModal) renderFooter() string {
	helpStyle := lipgloss.NewStyle().Foreground(styles.Subtext0)

	if m.session.IsTerminal() {
		var hints []string
		if !m.exported {
			hints = append(hints, "e: Export")
		}
		hints = append(hints, "Enter/Esc: Close", "j/k: Scroll")
		return helpStyle.Render(strings.Join(hints, " • "))
	}

	return helpStyle.Render("Esc: Cancel • j/k: Scroll")
}

// LoaderTickCmd returns a command that sends a tick after the given duration.
func LoaderTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return RunnerTickMsg{}
	})
}

// httpStatusStyle returns a style for the given HTTP status code.
func httpStatusStyle(status int) lipgloss.Style {
	switch {
	case status >= 200 && status < 300:
		return lipgloss.NewStyle().Foreground(styles.Green)
	case status >= 300 && status < 400:
		return lipgloss.NewStyle().Foreground(styles.Blue)
	case status >= 400 && status < 500:
		return lipgloss.NewStyle().Foreground(styles.Peach)
	case status >= 500:
		return lipgloss.NewStyle().Foreground(styles.Red)
	default:
		return lipgloss.NewStyle().Foreground(styles.Subtext0)
	}
}
