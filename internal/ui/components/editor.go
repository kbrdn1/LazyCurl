package components

import (
	"bytes"
	"encoding/json"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// EditorMode represents the current editing mode
type EditorMode int

const (
	EditorNormalMode EditorMode = iota
	EditorInsertMode
)

// EditorFormatMsg is sent when JSON is formatted
type EditorFormatMsg struct {
	Success bool
	Error   string
}

// EditorContentChangedMsg is sent when editor content is modified
type EditorContentChangedMsg struct {
	Content string
}

// Editor is a vim-like text editor component with line numbers
type Editor struct {
	content    []string   // Lines of content
	cursorRow  int        // Current row
	cursorCol  int        // Current column
	scrollY    int        // Vertical scroll offset
	scrollX    int        // Horizontal scroll offset
	height     int        // Available height
	width      int        // Available width
	readOnly   bool       // Whether the editor is read-only
	syntaxType string     // "json", "javascript", "text"
	mode       EditorMode // Current vim mode (NORMAL/INSERT)
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

// GetMode returns the current editor mode
func (e *Editor) GetMode() EditorMode {
	return e.mode
}

// FormatJSON formats the content as JSON with proper indentation
func (e *Editor) FormatJSON() tea.Cmd {
	content := e.GetContent()
	if content == "" {
		return nil
	}

	// Try to parse and format JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return func() tea.Msg {
			return EditorFormatMsg{Success: false, Error: err.Error()}
		}
	}

	var formatted bytes.Buffer
	encoder := json.NewEncoder(&formatted)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(parsed); err != nil {
		return func() tea.Msg {
			return EditorFormatMsg{Success: false, Error: err.Error()}
		}
	}

	// Remove trailing newline from encoder
	formattedStr := strings.TrimSuffix(formatted.String(), "\n")

	// Update content
	e.SetContent(formattedStr)

	return func() tea.Msg {
		return EditorFormatMsg{Success: true}
	}
}

// Update handles editor messages
func (e *Editor) Update(msg tea.Msg, allowInput bool) (*Editor, tea.Cmd) {
	if e.readOnly || !allowInput {
		// Still allow navigation in read-only mode (NORMAL mode commands)
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
			case "h", "left":
				if e.cursorCol > 0 {
					e.cursorCol--
					e.scrollIntoView()
				}
			case "l", "right":
				if e.cursorCol < len(e.content[e.cursorRow]) {
					e.cursorCol++
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
			case "0":
				e.cursorCol = 0
				e.scrollIntoView()
			case "$":
				e.cursorCol = len(e.content[e.cursorRow])
				e.scrollIntoView()
			case "w":
				// Move to next word
				e.moveToNextWord()
				e.scrollIntoView()
			case "b":
				// Move to previous word
				e.moveToPrevWord()
				e.scrollIntoView()
			}
		}
		return e, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle based on current mode
		if e.mode == EditorNormalMode {
			return e.handleNormalMode(msg)
		}
		return e.handleInsertMode(msg)
	}

	return e, nil
}

// handleNormalMode handles keyboard input in NORMAL mode
func (e *Editor) handleNormalMode(msg tea.KeyMsg) (*Editor, tea.Cmd) {
	contentModified := false

	switch msg.String() {
	// Mode switching
	case "i":
		// Enter INSERT mode at cursor
		e.mode = EditorInsertMode
	case "I":
		// Enter INSERT mode at line start
		e.cursorCol = 0
		e.mode = EditorInsertMode
	case "a":
		// Enter INSERT mode after cursor
		if e.cursorCol < len(e.content[e.cursorRow]) {
			e.cursorCol++
		}
		e.mode = EditorInsertMode
	case "A":
		// Enter INSERT mode at line end
		e.cursorCol = len(e.content[e.cursorRow])
		e.mode = EditorInsertMode
	case "o":
		// Open new line below and enter INSERT mode
		e.insertLineBelow()
		e.mode = EditorInsertMode
		contentModified = true
	case "O":
		// Open new line above and enter INSERT mode
		e.insertLineAbove()
		e.mode = EditorInsertMode
		contentModified = true

	// Navigation
	case "h", "left":
		if e.cursorCol > 0 {
			e.cursorCol--
			e.scrollIntoView()
		}
	case "l", "right":
		if e.cursorCol < len(e.content[e.cursorRow]) {
			e.cursorCol++
			e.scrollIntoView()
		}
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
	case "0":
		e.cursorCol = 0
		e.scrollIntoView()
	case "$":
		e.cursorCol = len(e.content[e.cursorRow])
		e.scrollIntoView()
	case "g":
		e.cursorRow = 0
		e.cursorCol = 0
		e.scrollIntoView()
	case "G":
		e.cursorRow = len(e.content) - 1
		e.cursorCol = 0
		e.scrollIntoView()
	case "w":
		e.moveToNextWord()
	case "b":
		e.moveToPrevWord()

	// Editing commands
	case "x":
		// Delete character at cursor
		line := e.content[e.cursorRow]
		if e.cursorCol < len(line) {
			e.content[e.cursorRow] = line[:e.cursorCol] + line[e.cursorCol+1:]
			e.ensureCursorInBounds()
			contentModified = true
		}
	case "d":
		// Delete line (dd would require tracking previous key)
		if len(e.content) > 1 {
			e.content = append(e.content[:e.cursorRow], e.content[e.cursorRow+1:]...)
			if e.cursorRow >= len(e.content) {
				e.cursorRow = len(e.content) - 1
			}
			e.ensureCursorInBounds()
			e.scrollIntoView()
		} else {
			e.content[0] = ""
			e.cursorCol = 0
		}
		contentModified = true

	// Format JSON (key feature!)
	case "F":
		if e.syntaxType == "json" {
			return e, e.FormatJSON()
		}
	}

	// Return content changed message if modified
	if contentModified {
		content := e.GetContent()
		return e, func() tea.Msg {
			return EditorContentChangedMsg{Content: content}
		}
	}

	return e, nil
}

// handleInsertMode handles keyboard input in INSERT mode
func (e *Editor) handleInsertMode(msg tea.KeyMsg) (*Editor, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Exit INSERT mode, go to NORMAL mode
		e.mode = EditorNormalMode
		// Move cursor back one if not at start
		if e.cursorCol > 0 {
			e.cursorCol--
		}
		// Emit content changed message when exiting INSERT mode
		content := e.GetContent()
		return e, func() tea.Msg {
			return EditorContentChangedMsg{Content: content}
		}

	case tea.KeyLeft:
		if e.cursorCol > 0 {
			e.cursorCol--
			e.scrollIntoView()
		}
	case tea.KeyRight:
		if e.cursorCol < len(e.content[e.cursorRow]) {
			e.cursorCol++
			e.scrollIntoView()
		}
	case tea.KeyUp:
		if e.cursorRow > 0 {
			e.cursorRow--
			e.ensureCursorInBounds()
			e.scrollIntoView()
		}
	case tea.KeyDown:
		if e.cursorRow < len(e.content)-1 {
			e.cursorRow++
			e.ensureCursorInBounds()
			e.scrollIntoView()
		}
	case tea.KeyHome:
		e.cursorCol = 0
	case tea.KeyEnd:
		e.cursorCol = len(e.content[e.cursorRow])

	case tea.KeyEnter:
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

	case tea.KeyBackspace:
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

	case tea.KeyDelete:
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

	case tea.KeyTab:
		// Insert 2 spaces for tab
		line := e.content[e.cursorRow]
		e.content[e.cursorRow] = line[:e.cursorCol] + "  " + line[e.cursorCol:]
		e.cursorCol += 2

	case tea.KeyRunes:
		// Insert characters
		char := string(msg.Runes)
		line := e.content[e.cursorRow]
		e.content[e.cursorRow] = line[:e.cursorCol] + char + line[e.cursorCol:]
		e.cursorCol += len(char)
		e.scrollIntoView()

	case tea.KeySpace:
		// Insert space
		line := e.content[e.cursorRow]
		e.content[e.cursorRow] = line[:e.cursorCol] + " " + line[e.cursorCol:]
		e.cursorCol++
		e.scrollIntoView()
	}

	return e, nil
}

// moveToNextWord moves cursor to the start of the next word
func (e *Editor) moveToNextWord() {
	line := e.content[e.cursorRow]

	// Skip current word
	for e.cursorCol < len(line) && !isWordSeparator(line[e.cursorCol]) {
		e.cursorCol++
	}
	// Skip separators
	for e.cursorCol < len(line) && isWordSeparator(line[e.cursorCol]) {
		e.cursorCol++
	}

	// If at end of line, go to next line
	if e.cursorCol >= len(line) && e.cursorRow < len(e.content)-1 {
		e.cursorRow++
		e.cursorCol = 0
		e.scrollIntoView()
	}
}

// moveToPrevWord moves cursor to the start of the previous word
func (e *Editor) moveToPrevWord() {
	// If at start of line, go to previous line
	if e.cursorCol == 0 && e.cursorRow > 0 {
		e.cursorRow--
		e.cursorCol = len(e.content[e.cursorRow])
		e.scrollIntoView()
	}

	line := e.content[e.cursorRow]

	// Skip separators
	for e.cursorCol > 0 && isWordSeparator(line[e.cursorCol-1]) {
		e.cursorCol--
	}
	// Skip current word
	for e.cursorCol > 0 && !isWordSeparator(line[e.cursorCol-1]) {
		e.cursorCol--
	}
}

// isWordSeparator returns true if the character is a word separator
func isWordSeparator(c byte) bool {
	return c == ' ' || c == '\t' || c == '{' || c == '}' ||
		c == '[' || c == ']' || c == '(' || c == ')' ||
		c == ':' || c == ',' || c == '"' || c == '\''
}

// insertLineBelow inserts a new empty line below the current line
func (e *Editor) insertLineBelow() {
	newContent := make([]string, 0, len(e.content)+1)
	newContent = append(newContent, e.content[:e.cursorRow+1]...)
	newContent = append(newContent, "")
	newContent = append(newContent, e.content[e.cursorRow+1:]...)
	e.content = newContent
	e.cursorRow++
	e.cursorCol = 0
	e.scrollIntoView()
}

// insertLineAbove inserts a new empty line above the current line
func (e *Editor) insertLineAbove() {
	newContent := make([]string, 0, len(e.content)+1)
	newContent = append(newContent, e.content[:e.cursorRow]...)
	newContent = append(newContent, "")
	newContent = append(newContent, e.content[e.cursorRow:]...)
	e.content = newContent
	e.cursorCol = 0
	e.scrollIntoView()
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

// scrollIntoView ensures cursor is visible (vertical and horizontal)
func (e *Editor) scrollIntoView() {
	// Vertical scrolling
	if e.cursorRow < e.scrollY {
		e.scrollY = e.cursorRow
	}
	if e.height > 0 && e.cursorRow >= e.scrollY+e.height {
		e.scrollY = e.cursorRow - e.height + 1
	}

	// Horizontal scrolling
	lineNumWidth := 3
	separatorWidth := 3
	contentWidth := e.width - lineNumWidth - separatorWidth - 2
	if contentWidth < 10 {
		contentWidth = 10
	}

	margin := 5
	if margin > contentWidth/4 {
		margin = contentWidth / 4
	}

	if e.cursorCol < e.scrollX {
		e.scrollX = e.cursorCol - margin
		if e.scrollX < 0 {
			e.scrollX = 0
		}
	}
	if e.cursorCol >= e.scrollX+contentWidth-margin {
		e.scrollX = e.cursorCol - contentWidth + margin + 1
		if e.scrollX < 0 {
			e.scrollX = 0
		}
	}
}

// View renders the editor
func (e *Editor) View(width, height int, active bool) string {
	e.width = width
	e.height = height - 1 // Reserve 1 line for mode indicator

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

	// Cursor styles
	normalCursorStyle := lipgloss.NewStyle().
		Background(styles.Text).
		Foreground(styles.Base)

	insertCursorStyle := lipgloss.NewStyle().
		Background(styles.Green).
		Foreground(styles.Base)

	// Calculate visible range
	start := e.scrollY
	end := e.scrollY + e.height
	if end > len(e.content) {
		end = len(e.content)
	}

	// Calculate content width for horizontal scrolling
	lineNumWidthCalc := 3
	separatorWidthCalc := 3
	contentWidth := width - lineNumWidthCalc - separatorWidthCalc - 2

	for i := start; i < end; i++ {
		lineNum := lineNumStyle.Render(string(rune('0'+((i+1)/10%10))) + string(rune('0'+((i+1)%10))))
		rawContent := e.content[i]

		// Apply horizontal scrolling
		displayContent := rawContent
		if e.scrollX > 0 && e.scrollX < len(rawContent) {
			displayContent = rawContent[e.scrollX:]
		} else if e.scrollX >= len(rawContent) {
			displayContent = ""
		}

		// Truncate to fit
		if len(displayContent) > contentWidth && contentWidth > 0 {
			displayContent = displayContent[:contentWidth]
		}

		adjustedCursorCol := e.cursorCol - e.scrollX

		var content string

		// Handle cursor rendering on the current line
		if active && i == e.cursorRow {
			content = e.renderLineWithCursorAtPos(displayContent, adjustedCursorCol, normalCursorStyle, insertCursorStyle)
		} else {
			if e.syntaxType == "json" {
				content = e.highlightJSON(displayContent)
			} else if e.syntaxType == "javascript" {
				content = e.highlightJS(displayContent)
			} else {
				content = textStyle.Render(displayContent)
			}
		}

		// Scroll indicators
		leftInd := ""
		if e.scrollX > 0 {
			leftInd = "◀"
		}
		rightInd := ""
		if len(rawContent) > e.scrollX+contentWidth {
			rightInd = "▶"
		}

		line := leftInd + lineNum + " │ " + content + rightInd

		if active && i == e.cursorRow {
			line = cursorLineStyle.Width(width).Render(line)
		}

		lines = append(lines, line)
	}

	// Fill remaining space with empty lines
	for i := len(lines); i < e.height; i++ {
		lineNum := lineNumStyle.Render("  ")
		lines = append(lines, lineNum+" │ ")
	}

	// Add mode indicator line
	modeIndicator := e.renderModeIndicator(width, active)
	lines = append(lines, modeIndicator)

	return strings.Join(lines, "\n")
}

// renderLineWithCursorAtPos renders a line with the cursor at a specific position
func (e *Editor) renderLineWithCursorAtPos(line string, cursorPos int, normalStyle, insertStyle lipgloss.Style) string {
	var result strings.Builder

	cursorStyle := normalStyle
	if e.mode == EditorInsertMode {
		cursorStyle = insertStyle
	}

	if cursorPos < 0 {
		cursorPos = 0
	}

	if e.syntaxType == "json" {
		if cursorPos < len(line) {
			before := line[:cursorPos]
			cursorChar := string(line[cursorPos])
			after := line[cursorPos+1:]

			result.WriteString(e.highlightJSON(before))
			result.WriteString(cursorStyle.Render(cursorChar))
			if len(after) > 0 {
				result.WriteString(e.highlightJSON(after))
			}
		} else {
			result.WriteString(e.highlightJSON(line))
			result.WriteString(cursorStyle.Render(" "))
		}
	} else {
		if cursorPos < len(line) {
			result.WriteString(line[:cursorPos])
			result.WriteString(cursorStyle.Render(string(line[cursorPos])))
			if cursorPos+1 < len(line) {
				result.WriteString(line[cursorPos+1:])
			}
		} else {
			result.WriteString(line)
			result.WriteString(cursorStyle.Render(" "))
		}
	}

	return result.String()
}

// renderModeIndicator renders the mode indicator bar at the bottom
func (e *Editor) renderModeIndicator(width int, active bool) string {
	var modeText string
	var modeStyle lipgloss.Style

	if e.mode == EditorNormalMode {
		modeText = " NORMAL "
		modeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.Base).
			Background(styles.Blue)
	} else {
		modeText = " INSERT "
		modeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(styles.Base).
			Background(styles.Green)
	}

	// Help text based on mode
	helpStyle := lipgloss.NewStyle().
		Foreground(styles.Subtext0).
		Background(styles.Surface0)

	var helpText string
	if e.mode == EditorNormalMode {
		helpText = " i:insert  F:format  hjkl:nav  x:del  d:del line "
	} else {
		helpText = " Esc:normal  Type to insert "
	}

	// Build the bar
	barStyle := lipgloss.NewStyle().
		Background(styles.Surface0).
		Width(width)

	content := modeStyle.Render(modeText) + helpStyle.Render(helpText)

	if !active {
		// Dimmed when not active
		content = lipgloss.NewStyle().Foreground(styles.Subtext0).Render("── JSON Editor ──")
	}

	return barStyle.Render(content)
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
