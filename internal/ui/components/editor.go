package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// Editor is a simple text editor component with line numbers
type Editor struct {
	content    []string // Lines of content
	cursorRow  int      // Current row
	cursorCol  int      // Current column
	scrollY    int      // Vertical scroll offset
	height     int      // Available height
	width      int      // Available width
	readOnly   bool     // Whether the editor is read-only
	syntaxType string   // "json", "javascript", "text"
}

// NewEditor creates a new editor component
func NewEditor(content string, syntaxType string) *Editor {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	return &Editor{
		content:    lines,
		cursorRow:  0,
		cursorCol:  0,
		scrollY:    0,
		syntaxType: syntaxType,
	}
}

// SetContent sets the editor content
func (e *Editor) SetContent(content string) {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	e.content = lines
	e.cursorRow = 0
	e.cursorCol = 0
	e.scrollY = 0
}

// GetContent returns the editor content as a single string
func (e *Editor) GetContent() string {
	return strings.Join(e.content, "\n")
}

// SetReadOnly sets whether the editor is read-only
func (e *Editor) SetReadOnly(readOnly bool) {
	e.readOnly = readOnly
}

// Update handles editor messages
func (e *Editor) Update(msg tea.Msg, allowInput bool) (*Editor, tea.Cmd) {
	if e.readOnly || !allowInput {
		// Still allow navigation in read-only mode
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "j", "down":
				if e.cursorRow < len(e.content)-1 {
					e.cursorRow++
					e.ensureCursorInBounds()
					e.scrollIntoView()
				}
			case "k", "up":
				if e.cursorRow > 0 {
					e.cursorRow--
					e.ensureCursorInBounds()
					e.scrollIntoView()
				}
			case "g":
				e.cursorRow = 0
				e.cursorCol = 0
				e.scrollIntoView()
			case "G":
				e.cursorRow = len(e.content) - 1
				e.cursorCol = 0
				e.scrollIntoView()
			}
		}
		return e, nil
	}

	// Handle edit mode input
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left":
			if e.cursorCol > 0 {
				e.cursorCol--
			}
		case "right":
			if e.cursorCol < len(e.content[e.cursorRow]) {
				e.cursorCol++
			}
		case "up":
			if e.cursorRow > 0 {
				e.cursorRow--
				e.ensureCursorInBounds()
				e.scrollIntoView()
			}
		case "down":
			if e.cursorRow < len(e.content)-1 {
				e.cursorRow++
				e.ensureCursorInBounds()
				e.scrollIntoView()
			}
		case "home":
			e.cursorCol = 0
		case "end":
			e.cursorCol = len(e.content[e.cursorRow])
		case "enter":
			// Split line at cursor
			line := e.content[e.cursorRow]
			before := line[:e.cursorCol]
			after := line[e.cursorCol:]
			e.content[e.cursorRow] = before
			// Insert new line after current
			newContent := make([]string, 0, len(e.content)+1)
			newContent = append(newContent, e.content[:e.cursorRow+1]...)
			newContent = append(newContent, after)
			newContent = append(newContent, e.content[e.cursorRow+1:]...)
			e.content = newContent
			e.cursorRow++
			e.cursorCol = 0
			e.scrollIntoView()
		case "backspace":
			if e.cursorCol > 0 {
				// Delete character before cursor
				line := e.content[e.cursorRow]
				e.content[e.cursorRow] = line[:e.cursorCol-1] + line[e.cursorCol:]
				e.cursorCol--
			} else if e.cursorRow > 0 {
				// Join with previous line
				prevLine := e.content[e.cursorRow-1]
				currLine := e.content[e.cursorRow]
				e.cursorCol = len(prevLine)
				e.content[e.cursorRow-1] = prevLine + currLine
				// Remove current line
				e.content = append(e.content[:e.cursorRow], e.content[e.cursorRow+1:]...)
				e.cursorRow--
				e.scrollIntoView()
			}
		case "delete":
			line := e.content[e.cursorRow]
			if e.cursorCol < len(line) {
				// Delete character at cursor
				e.content[e.cursorRow] = line[:e.cursorCol] + line[e.cursorCol+1:]
			} else if e.cursorRow < len(e.content)-1 {
				// Join with next line
				nextLine := e.content[e.cursorRow+1]
				e.content[e.cursorRow] = line + nextLine
				// Remove next line
				e.content = append(e.content[:e.cursorRow+1], e.content[e.cursorRow+2:]...)
			}
		default:
			// Insert character
			if len(msg.String()) == 1 {
				char := msg.String()
				line := e.content[e.cursorRow]
				e.content[e.cursorRow] = line[:e.cursorCol] + char + line[e.cursorCol:]
				e.cursorCol++
			}
		}
	}

	return e, nil
}

// ensureCursorInBounds ensures cursor column is within line bounds
func (e *Editor) ensureCursorInBounds() {
	if e.cursorRow >= len(e.content) {
		e.cursorRow = len(e.content) - 1
	}
	if e.cursorRow < 0 {
		e.cursorRow = 0
	}
	lineLen := len(e.content[e.cursorRow])
	if e.cursorCol > lineLen {
		e.cursorCol = lineLen
	}
}

// scrollIntoView ensures cursor is visible
func (e *Editor) scrollIntoView() {
	if e.cursorRow < e.scrollY {
		e.scrollY = e.cursorRow
	}
	if e.height > 0 && e.cursorRow >= e.scrollY+e.height {
		e.scrollY = e.cursorRow - e.height + 1
	}
}

// View renders the editor
func (e *Editor) View(width, height int, active bool) string {
	e.width = width
	e.height = height

	var lines []string
	lineNumWidth := 3 // Width for line numbers

	lineNumStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Width(lineNumWidth).
		Align(lipgloss.Right)

	textStyle := lipgloss.NewStyle().
		Foreground(styles.Text)

	cursorLineStyle := lipgloss.NewStyle().
		Background(styles.Surface0)

	// Calculate visible range
	start := e.scrollY
	end := e.scrollY + height
	if end > len(e.content) {
		end = len(e.content)
	}

	for i := start; i < end; i++ {
		lineNum := lineNumStyle.Render(string(rune('0'+((i+1)/10%10))) + string(rune('0'+((i+1)%10))))
		content := e.content[i]

		// Apply syntax highlighting
		if e.syntaxType == "json" {
			content = e.highlightJSON(content)
		} else if e.syntaxType == "javascript" {
			content = e.highlightJS(content)
		} else {
			content = textStyle.Render(content)
		}

		line := lineNum + " │ " + content

		// Highlight cursor line when active
		if active && i == e.cursorRow {
			line = cursorLineStyle.Width(width).Render(line)
		}

		lines = append(lines, line)
	}

	// Fill remaining space with empty lines
	for i := len(lines); i < height; i++ {
		lineNum := lineNumStyle.Render("  ")
		lines = append(lines, lineNum+" │ ")
	}

	return strings.Join(lines, "\n")
}

// highlightJSON applies basic JSON syntax highlighting
func (e *Editor) highlightJSON(line string) string {
	var result strings.Builder
	inString := false
	inKey := false

	keyStyle := lipgloss.NewStyle().Foreground(styles.Blue)
	stringStyle := lipgloss.NewStyle().Foreground(styles.Green)
	numberStyle := lipgloss.NewStyle().Foreground(styles.Peach)
	boolStyle := lipgloss.NewStyle().Foreground(styles.Mauve)
	punctStyle := lipgloss.NewStyle().Foreground(styles.Subtext1)

	i := 0
	for i < len(line) {
		char := line[i]

		switch {
		case char == '"':
			if !inString {
				inString = true
				// Check if this is a key (colon follows after string)
				// Simple heuristic: if we're at the start or after whitespace/punctuation
				inKey = true
				result.WriteString(keyStyle.Render("\""))
			} else {
				if inKey {
					result.WriteString(keyStyle.Render("\""))
				} else {
					result.WriteString(stringStyle.Render("\""))
				}
				inString = false
				inKey = false
			}
		case inString:
			if inKey {
				result.WriteString(keyStyle.Render(string(char)))
			} else {
				result.WriteString(stringStyle.Render(string(char)))
			}
		case char == ':':
			inKey = false // After colon, it's a value
			result.WriteString(punctStyle.Render(":"))
		case char >= '0' && char <= '9' || char == '-' || char == '.':
			// Number
			numStr := ""
			for i < len(line) && (line[i] >= '0' && line[i] <= '9' || line[i] == '-' || line[i] == '.' || line[i] == 'e' || line[i] == 'E') {
				numStr += string(line[i])
				i++
			}
			i-- // Back up one since loop will increment
			result.WriteString(numberStyle.Render(numStr))
		case char == '{' || char == '}' || char == '[' || char == ']' || char == ',':
			result.WriteString(punctStyle.Render(string(char)))
		default:
			// Check for true/false/null
			remaining := line[i:]
			if strings.HasPrefix(remaining, "true") {
				result.WriteString(boolStyle.Render("true"))
				i += 3
			} else if strings.HasPrefix(remaining, "false") {
				result.WriteString(boolStyle.Render("false"))
				i += 4
			} else if strings.HasPrefix(remaining, "null") {
				result.WriteString(boolStyle.Render("null"))
				i += 3
			} else {
				result.WriteString(string(char))
			}
		}
		i++
	}

	return result.String()
}

// highlightJS applies basic JavaScript syntax highlighting
func (e *Editor) highlightJS(line string) string {
	// Simple keyword highlighting
	keywordStyle := lipgloss.NewStyle().Foreground(styles.Mauve).Bold(true)
	commentStyle := lipgloss.NewStyle().Foreground(styles.Subtext0).Italic(true)
	funcStyle := lipgloss.NewStyle().Foreground(styles.Blue)

	// Check for comments
	if strings.HasPrefix(strings.TrimSpace(line), "//") {
		return commentStyle.Render(line)
	}

	result := line

	// Highlight keywords
	keywords := []string{"const", "let", "var", "function", "return", "if", "else", "for", "while", "async", "await"}
	for _, kw := range keywords {
		result = strings.ReplaceAll(result, kw+" ", keywordStyle.Render(kw)+" ")
	}

	// Highlight console.log
	result = strings.ReplaceAll(result, "console.log", funcStyle.Render("console.log"))

	return result
}

// SetHeight sets the visible height
func (e *Editor) SetHeight(h int) {
	e.height = h
	e.scrollIntoView()
}
