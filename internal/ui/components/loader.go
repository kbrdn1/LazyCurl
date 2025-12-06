package components

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// LoaderStyle represents different loader animation styles
type LoaderStyle int

const (
	LoaderSpinner LoaderStyle = iota
	LoaderDots
	LoaderBar
	LoaderPulse
)

// Loader is an animated loading indicator component
type Loader struct {
	style    LoaderStyle
	frame    int
	width    int
	message  string
	active   bool
	lastTick time.Time
}

// LoaderTickMsg is sent to animate the loader
type LoaderTickMsg time.Time

// Spinner frames (Nerd Font compatible)
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Dots frames
var dotsFrames = []string{"   ", ".  ", ".. ", "...", " ..", "  .", "   "}

// Bar frames (progress bar animation)
var barChars = []string{"░", "▒", "▓", "█"}

// Pulse frames
var pulseFrames = []string{"○", "◔", "◑", "◕", "●", "◕", "◑", "◔"}

// NewLoader creates a new loader component
func NewLoader(style LoaderStyle, message string) *Loader {
	return &Loader{
		style:   style,
		message: message,
		frame:   0,
		width:   40,
		active:  false,
	}
}

// Start begins the loader animation
func (l *Loader) Start() tea.Cmd {
	l.active = true
	l.frame = 0
	l.lastTick = time.Now()
	return l.tick()
}

// Stop stops the loader animation
func (l *Loader) Stop() {
	l.active = false
}

// IsActive returns whether the loader is active
func (l *Loader) IsActive() bool {
	return l.active
}

// SetMessage updates the loader message
func (l *Loader) SetMessage(msg string) {
	l.message = msg
}

// SetWidth sets the width for bar-style loaders
func (l *Loader) SetWidth(w int) {
	l.width = w
}

// tick returns a command that sends a tick message
func (l *Loader) tick() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return LoaderTickMsg(t)
	})
}

// Update handles loader animation
func (l *Loader) Update(msg tea.Msg) (*Loader, tea.Cmd) {
	if !l.active {
		return l, nil
	}

	switch msg.(type) {
	case LoaderTickMsg:
		l.frame++
		return l, l.tick()
	}

	return l, nil
}

// View renders the loader
func (l *Loader) View() string {
	if !l.active {
		return ""
	}

	var animation string

	switch l.style {
	case LoaderSpinner:
		animation = l.renderSpinner()
	case LoaderDots:
		animation = l.renderDots()
	case LoaderBar:
		animation = l.renderBar()
	case LoaderPulse:
		animation = l.renderPulse()
	default:
		animation = l.renderSpinner()
	}

	return animation
}

// renderSpinner renders a spinner animation
func (l *Loader) renderSpinner() string {
	spinnerStyle := lipgloss.NewStyle().
		Foreground(styles.Blue).
		Bold(true)

	msgStyle := lipgloss.NewStyle().
		Foreground(styles.Text)

	frame := spinnerFrames[l.frame%len(spinnerFrames)]
	return spinnerStyle.Render(frame) + " " + msgStyle.Render(l.message)
}

// renderDots renders a dots animation
func (l *Loader) renderDots() string {
	msgStyle := lipgloss.NewStyle().
		Foreground(styles.Blue).
		Bold(true)

	dotsStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender)

	frame := dotsFrames[l.frame%len(dotsFrames)]
	return msgStyle.Render(l.message) + dotsStyle.Render(frame)
}

// renderBar renders an animated progress bar
func (l *Loader) renderBar() string {
	var result strings.Builder

	// Message
	msgStyle := lipgloss.NewStyle().
		Foreground(styles.Blue).
		Bold(true)
	result.WriteString(msgStyle.Render(l.message))
	result.WriteString(" ")

	// Animated bar
	barWidth := l.width - len(l.message) - 4
	if barWidth < 10 {
		barWidth = 10
	}

	// Create wave effect
	barStyle := lipgloss.NewStyle().Foreground(styles.Blue)
	dimStyle := lipgloss.NewStyle().Foreground(styles.Surface1)

	result.WriteString("│")
	for i := 0; i < barWidth; i++ {
		// Calculate wave position
		wavePos := (l.frame + i) % (barWidth * 2)
		if wavePos > barWidth {
			wavePos = barWidth*2 - wavePos
		}

		// Distance from wave center
		dist := abs(i - wavePos)
		if dist < 3 {
			charIdx := 3 - dist
			if charIdx >= len(barChars) {
				charIdx = len(barChars) - 1
			}
			result.WriteString(barStyle.Render(barChars[charIdx]))
		} else {
			result.WriteString(dimStyle.Render("░"))
		}
	}
	result.WriteString("│")

	return result.String()
}

// renderPulse renders a pulse animation
func (l *Loader) renderPulse() string {
	pulseStyle := lipgloss.NewStyle().
		Foreground(styles.Lavender).
		Bold(true)

	msgStyle := lipgloss.NewStyle().
		Foreground(styles.Text)

	frame := pulseFrames[l.frame%len(pulseFrames)]
	return pulseStyle.Render(frame) + " " + msgStyle.Render(l.message)
}

// abs returns absolute value of an int
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// HorizontalLoader creates a simple horizontal loading bar for inline use
func HorizontalLoader(width int, frame int, message string) string {
	var result strings.Builder

	// Icon (Nerd Font clock or Unicode)
	iconStyle := lipgloss.NewStyle().
		Foreground(styles.Blue).
		Bold(true)

	msgStyle := lipgloss.NewStyle().
		Foreground(styles.Text)

	// Use rotating spinner
	spinner := spinnerFrames[frame%len(spinnerFrames)]
	result.WriteString(iconStyle.Render(spinner))
	result.WriteString(" ")
	result.WriteString(msgStyle.Render(message))
	result.WriteString(" ")

	// Progress indicator
	barWidth := width - len(message) - 6
	if barWidth < 5 {
		barWidth = 5
	}

	barStyle := lipgloss.NewStyle().Foreground(styles.Blue)
	dimStyle := lipgloss.NewStyle().Foreground(styles.Surface1)

	// Animated sliding block
	pos := frame % (barWidth * 2)
	if pos >= barWidth {
		pos = barWidth*2 - pos - 1
	}

	for i := 0; i < barWidth; i++ {
		dist := abs(i - pos)
		if dist == 0 {
			result.WriteString(barStyle.Render("█"))
		} else if dist == 1 {
			result.WriteString(barStyle.Render("▓"))
		} else if dist == 2 {
			result.WriteString(barStyle.Render("▒"))
		} else {
			result.WriteString(dimStyle.Render("─"))
		}
	}

	return result.String()
}
