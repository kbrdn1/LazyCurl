package components

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// EditorQuitMsg is sent when user presses Q in NORMAL mode to quit the app
type EditorQuitMsg struct{}

// SearchMatch represents a match position in the editor
type SearchMatch struct {
	Row      int // Line number (0-indexed)
	ColStart int // Start column
	ColEnd   int // End column (exclusive)
}

// EditorState represents a snapshot of editor state for undo/redo
type EditorState struct {
	content   []string
	cursorRow int
	cursorCol int
}

// maxUndoHistory is the maximum number of undo states to keep
const maxUndoHistory = 100

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

	// Search state
	search          *SearchInput  // Search input component
	searchQuery     string        // Current search query
	searchMatches   []SearchMatch // All matches in content
	currentMatchIdx int           // Index of current match (-1 if none)

	// Undo/Redo state
	undoStack []EditorState // Stack of previous states for undo
	redoStack []EditorState // Stack of undone states for redo
}

// NewEditor creates a new editor component
func NewEditor(content string, syntaxType string) *Editor {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		lines = []string{""}
	}
	return &Editor{
		content:         lines,
		cursorRow:       0,
		cursorCol:       0,
		scrollY:         0,
		syntaxType:      syntaxType,
		search:          NewSearchInput(),
		currentMatchIdx: -1,
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
	// Handle search messages first (they come from the search input component)
	switch msg := msg.(type) {
	case SearchUpdateMsg:
		e.searchQuery = msg.Query
		e.findMatches()
		e.goToCurrentMatch()
		return e, nil

	case SearchCloseMsg:
		if msg.Canceled {
			e.clearSearch()
		}
		// Keep matches for n/N navigation when not cancelled
		return e, nil
	}

	// Handle search input when visible
	if e.IsSearching() {
		var cmd tea.Cmd
		e.search, cmd = e.search.Update(msg)
		return e, cmd
	}

	if e.readOnly || !allowInput {
		// Still allow navigation and search in read-only mode (NORMAL mode commands)
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
			case "/":
				// Open search
				e.search.Show()
				return e, nil
			case "n":
				// Next match
				if e.HasSearchQuery() {
					e.nextMatch()
				}
				return e, nil
			case "N":
				// Previous match
				if e.HasSearchQuery() {
					e.prevMatch()
				}
				return e, nil
			case "esc":
				// Clear search if active
				if e.searchQuery != "" {
					e.clearSearch()
					return e, nil
				}
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
		e.saveState() // Save state before INSERT mode
		e.mode = EditorInsertMode
	case "I":
		// Enter INSERT mode at line start
		e.saveState() // Save state before INSERT mode
		e.cursorCol = 0
		e.mode = EditorInsertMode
	case "a":
		// Enter INSERT mode after cursor
		e.saveState() // Save state before INSERT mode
		if e.cursorCol < len(e.content[e.cursorRow]) {
			e.cursorCol++
		}
		e.mode = EditorInsertMode
	case "A":
		// Enter INSERT mode at line end
		e.saveState() // Save state before INSERT mode
		e.cursorCol = len(e.content[e.cursorRow])
		e.mode = EditorInsertMode
	case "o":
		// Open new line below and enter INSERT mode
		e.saveState() // Save state before INSERT mode
		e.insertLineBelow()
		e.mode = EditorInsertMode
		contentModified = true
	case "O":
		// Open new line above and enter INSERT mode
		e.saveState() // Save state before INSERT mode
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
			e.saveState() // Save state before deletion
			e.content[e.cursorRow] = line[:e.cursorCol] + line[e.cursorCol+1:]
			e.ensureCursorInBounds()
			contentModified = true
		}
	case "d":
		// Delete line (dd would require tracking previous key)
		e.saveState() // Save state before deletion
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

	// Undo/Redo
	case "u":
		// Undo
		if e.undo() {
			contentModified = true
		}
	case "ctrl+r":
		// Redo
		if e.redo() {
			contentModified = true
		}

	// Quit application
	case "Q":
		return e, func() tea.Msg { return EditorQuitMsg{} }

	// Format JSON (key feature!)
	case "F":
		if e.syntaxType == "json" {
			return e, e.FormatJSON()
		}

	// Search commands
	case "/":
		e.search.Show()
		return e, nil
	case "n":
		if e.HasSearchQuery() {
			e.nextMatch()
		}
		return e, nil
	case "N":
		if e.HasSearchQuery() {
			e.prevMatch()
		}
		return e, nil
	case "esc":
		// Clear search if active
		if e.searchQuery != "" {
			e.clearSearch()
			return e, nil
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
	availableHeight := height - 1 // Reserve 1 line for mode indicator

	var output []string

	// Render search box if visible
	if e.IsSearching() {
		current, total := e.GetMatchCount()
		searchBox := e.search.ViewCompact(width, current, total)
		output = append(output, searchBox)
		availableHeight--
	} else if e.searchQuery != "" {
		// Show compact filter indicator with count
		current, total := e.GetMatchCount()
		filterStyle := lipgloss.NewStyle().Foreground(styles.Yellow)
		countStyle := lipgloss.NewStyle().Foreground(styles.Subtext0)
		escStyle := lipgloss.NewStyle().Foreground(styles.Subtext0).Italic(true)
		filterText := filterStyle.Render("/"+e.searchQuery) + countStyle.Render(fmt.Sprintf(" %d/%d", current, total)) + escStyle.Render(" esc")
		output = append(output, filterText)
		availableHeight--
	}

	e.height = availableHeight

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
		displayStart := e.scrollX
		if e.scrollX > 0 && e.scrollX < len(rawContent) {
			displayContent = rawContent[e.scrollX:]
		} else if e.scrollX >= len(rawContent) {
			displayContent = ""
			displayStart = len(rawContent)
		}

		// Truncate to fit
		if len(displayContent) > contentWidth && contentWidth > 0 {
			displayContent = displayContent[:contentWidth]
		}

		adjustedCursorCol := e.cursorCol - e.scrollX

		var content string

		// Handle cursor rendering on the current line
		if active && i == e.cursorRow {
			content = e.renderLineWithCursorAndMatches(displayContent, i, displayStart, adjustedCursorCol, normalCursorStyle, insertCursorStyle)
		} else {
			// Render with search highlights
			content = e.renderLineWithMatches(displayContent, i, displayStart, textStyle)
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

	output = append(output, strings.Join(lines, "\n"))
	return strings.Join(output, "\n")
}

// renderLineWithMatches renders a line with search match highlighting
func (e *Editor) renderLineWithMatches(displayContent string, row int, displayStart int, textStyle lipgloss.Style) string {
	if e.searchQuery == "" || len(e.searchMatches) == 0 {
		// No search, use normal rendering
		if e.syntaxType == "json" {
			return e.highlightJSON(displayContent)
		} else if e.syntaxType == "javascript" {
			return e.highlightJS(displayContent)
		}
		return textStyle.Render(displayContent)
	}

	// Find matches on this line
	var lineMatches []SearchMatch
	for idx, match := range e.searchMatches {
		if match.Row == row {
			// Adjust for horizontal scroll
			adjustedMatch := SearchMatch{
				Row:      match.Row,
				ColStart: match.ColStart - displayStart,
				ColEnd:   match.ColEnd - displayStart,
			}
			// Check if match is visible
			if adjustedMatch.ColEnd > 0 && adjustedMatch.ColStart < len(displayContent) {
				// Clamp to visible area
				if adjustedMatch.ColStart < 0 {
					adjustedMatch.ColStart = 0
				}
				if adjustedMatch.ColEnd > len(displayContent) {
					adjustedMatch.ColEnd = len(displayContent)
				}
				// Store if this is current match
				adjustedMatch.Row = idx // Reuse Row field to store original index
				lineMatches = append(lineMatches, adjustedMatch)
			}
		}
	}

	if len(lineMatches) == 0 {
		// No visible matches, use normal rendering
		if e.syntaxType == "json" {
			return e.highlightJSON(displayContent)
		} else if e.syntaxType == "javascript" {
			return e.highlightJS(displayContent)
		}
		return textStyle.Render(displayContent)
	}

	// Render with highlights
	matchStyle := lipgloss.NewStyle().
		Background(styles.Yellow).
		Foreground(styles.Base)

	currentMatchStyle := lipgloss.NewStyle().
		Background(styles.Peach).
		Foreground(styles.Base).
		Bold(true)

	var result strings.Builder
	pos := 0

	for _, match := range lineMatches {
		// Render text before match
		if pos < match.ColStart {
			beforeText := displayContent[pos:match.ColStart]
			if e.syntaxType == "json" {
				result.WriteString(e.highlightJSON(beforeText))
			} else {
				result.WriteString(textStyle.Render(beforeText))
			}
		}

		// Render match with highlight
		matchText := displayContent[match.ColStart:match.ColEnd]
		if match.Row == e.currentMatchIdx { // Row field stores original match index
			result.WriteString(currentMatchStyle.Render(matchText))
		} else {
			result.WriteString(matchStyle.Render(matchText))
		}

		pos = match.ColEnd
	}

	// Render remaining text
	if pos < len(displayContent) {
		afterText := displayContent[pos:]
		if e.syntaxType == "json" {
			result.WriteString(e.highlightJSON(afterText))
		} else {
			result.WriteString(textStyle.Render(afterText))
		}
	}

	return result.String()
}

// renderLineWithCursorAndMatches renders a line with both cursor and search highlights
func (e *Editor) renderLineWithCursorAndMatches(displayContent string, row int, displayStart int, cursorPos int, normalCursorStyle, insertCursorStyle lipgloss.Style) string {
	if e.searchQuery == "" || len(e.searchMatches) == 0 {
		// No search, use normal cursor rendering
		return e.renderLineWithCursorAtPos(displayContent, cursorPos, normalCursorStyle, insertCursorStyle)
	}

	cursorStyle := normalCursorStyle
	if e.mode == EditorInsertMode {
		cursorStyle = insertCursorStyle
	}

	// Find matches on this line
	var lineMatches []SearchMatch
	for idx, match := range e.searchMatches {
		if match.Row == row {
			adjustedMatch := SearchMatch{
				Row:      idx, // Store original index
				ColStart: match.ColStart - displayStart,
				ColEnd:   match.ColEnd - displayStart,
			}
			if adjustedMatch.ColEnd > 0 && adjustedMatch.ColStart < len(displayContent) {
				if adjustedMatch.ColStart < 0 {
					adjustedMatch.ColStart = 0
				}
				if adjustedMatch.ColEnd > len(displayContent) {
					adjustedMatch.ColEnd = len(displayContent)
				}
				lineMatches = append(lineMatches, adjustedMatch)
			}
		}
	}

	matchStyle := lipgloss.NewStyle().
		Background(styles.Yellow).
		Foreground(styles.Base)

	currentMatchStyle := lipgloss.NewStyle().
		Background(styles.Peach).
		Foreground(styles.Base).
		Bold(true)

	textStyle := lipgloss.NewStyle().Foreground(styles.Text)

	var result strings.Builder
	pos := 0

	// Helper to check if position is in a match
	isInMatch := func(p int) (bool, int, bool) {
		for _, m := range lineMatches {
			if p >= m.ColStart && p < m.ColEnd {
				return true, m.Row, m.Row == e.currentMatchIdx
			}
		}
		return false, -1, false
	}

	for i := 0; i <= len(displayContent); i++ {
		inMatch, _, isCurrent := isInMatch(i)

		if i == cursorPos {
			// Render cursor
			var char string
			if i < len(displayContent) {
				char = string(displayContent[i])
			} else {
				char = " "
			}
			result.WriteString(cursorStyle.Render(char))
			pos = i + 1
		} else if i < len(displayContent) {
			// Check for match transitions
			nextInMatch, _, _ := isInMatch(i + 1)

			if inMatch && !nextInMatch {
				// End of match
				text := displayContent[pos : i+1]
				if isCurrent {
					result.WriteString(currentMatchStyle.Render(text))
				} else {
					result.WriteString(matchStyle.Render(text))
				}
				pos = i + 1
			} else if !inMatch && nextInMatch {
				// Start of new match
				if pos <= i {
					text := displayContent[pos : i+1]
					if e.syntaxType == "json" {
						result.WriteString(e.highlightJSON(text))
					} else {
						result.WriteString(textStyle.Render(text))
					}
				}
				pos = i + 1
			}
		}
	}

	// Render any remaining text
	if pos < len(displayContent) {
		text := displayContent[pos:]
		inMatch, _, isCurrent := isInMatch(pos)
		if inMatch {
			if isCurrent {
				result.WriteString(currentMatchStyle.Render(text))
			} else {
				result.WriteString(matchStyle.Render(text))
			}
		} else {
			if e.syntaxType == "json" {
				result.WriteString(e.highlightJSON(text))
			} else {
				result.WriteString(textStyle.Render(text))
			}
		}
	}

	// Cursor at end
	if cursorPos >= len(displayContent) {
		result.WriteString(cursorStyle.Render(" "))
	}

	return result.String()
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
		if e.HasSearchQuery() {
			helpText = " n:next  N:prev  esc:clear  /:search "
		} else {
			helpText = " i:insert  /:search  F:format  u:undo  ^R:redo "
		}
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

// IsSearching returns true if search input is visible
func (e *Editor) IsSearching() bool {
	return e.search != nil && e.search.IsVisible()
}

// HasSearchQuery returns true if there's an active search query
func (e *Editor) HasSearchQuery() bool {
	return e.searchQuery != "" && !e.IsSearching()
}

// findMatches searches for all occurrences of the query in the content
func (e *Editor) findMatches() {
	e.searchMatches = nil
	e.currentMatchIdx = -1

	if e.searchQuery == "" {
		return
	}

	query := strings.ToLower(e.searchQuery)
	for row, line := range e.content {
		lineLower := strings.ToLower(line)
		start := 0
		for {
			idx := strings.Index(lineLower[start:], query)
			if idx == -1 {
				break
			}
			e.searchMatches = append(e.searchMatches, SearchMatch{
				Row:      row,
				ColStart: start + idx,
				ColEnd:   start + idx + len(query),
			})
			start = start + idx + 1
		}
	}

	// Set current match to the first one at or after cursor
	if len(e.searchMatches) > 0 {
		for i, match := range e.searchMatches {
			if match.Row > e.cursorRow || (match.Row == e.cursorRow && match.ColStart >= e.cursorCol) {
				e.currentMatchIdx = i
				return
			}
		}
		// If no match after cursor, wrap to first
		e.currentMatchIdx = 0
	}
}

// goToCurrentMatch moves cursor to the current match
func (e *Editor) goToCurrentMatch() {
	if e.currentMatchIdx < 0 || e.currentMatchIdx >= len(e.searchMatches) {
		return
	}
	match := e.searchMatches[e.currentMatchIdx]
	e.cursorRow = match.Row
	e.cursorCol = match.ColStart
	e.scrollIntoView()
}

// nextMatch moves to the next search match
func (e *Editor) nextMatch() {
	if len(e.searchMatches) == 0 {
		return
	}
	e.currentMatchIdx = (e.currentMatchIdx + 1) % len(e.searchMatches)
	e.goToCurrentMatch()
}

// prevMatch moves to the previous search match
func (e *Editor) prevMatch() {
	if len(e.searchMatches) == 0 {
		return
	}
	e.currentMatchIdx--
	if e.currentMatchIdx < 0 {
		e.currentMatchIdx = len(e.searchMatches) - 1
	}
	e.goToCurrentMatch()
}

// clearSearch clears the search state
func (e *Editor) clearSearch() {
	e.searchQuery = ""
	e.searchMatches = nil
	e.currentMatchIdx = -1
}

// GetMatchCount returns (currentMatch, totalMatches)
func (e *Editor) GetMatchCount() (int, int) {
	if len(e.searchMatches) == 0 {
		return 0, 0
	}
	return e.currentMatchIdx + 1, len(e.searchMatches)
}

// saveState saves the current editor state to the undo stack
func (e *Editor) saveState() {
	// Create a deep copy of content
	contentCopy := make([]string, len(e.content))
	copy(contentCopy, e.content)

	state := EditorState{
		content:   contentCopy,
		cursorRow: e.cursorRow,
		cursorCol: e.cursorCol,
	}

	// Add to undo stack
	e.undoStack = append(e.undoStack, state)

	// Limit undo stack size
	if len(e.undoStack) > maxUndoHistory {
		e.undoStack = e.undoStack[1:]
	}

	// Clear redo stack when new changes are made
	e.redoStack = nil
}

// undo restores the previous state
func (e *Editor) undo() bool {
	if len(e.undoStack) == 0 {
		return false
	}

	// Save current state to redo stack before undoing
	contentCopy := make([]string, len(e.content))
	copy(contentCopy, e.content)
	e.redoStack = append(e.redoStack, EditorState{
		content:   contentCopy,
		cursorRow: e.cursorRow,
		cursorCol: e.cursorCol,
	})

	// Pop from undo stack
	state := e.undoStack[len(e.undoStack)-1]
	e.undoStack = e.undoStack[:len(e.undoStack)-1]

	// Restore state
	e.content = state.content
	e.cursorRow = state.cursorRow
	e.cursorCol = state.cursorCol
	e.ensureCursorInBounds()
	e.scrollIntoView()

	return true
}

// redo restores a previously undone state
func (e *Editor) redo() bool {
	if len(e.redoStack) == 0 {
		return false
	}

	// Save current state to undo stack before redoing
	contentCopy := make([]string, len(e.content))
	copy(contentCopy, e.content)
	e.undoStack = append(e.undoStack, EditorState{
		content:   contentCopy,
		cursorRow: e.cursorRow,
		cursorCol: e.cursorCol,
	})

	// Pop from redo stack
	state := e.redoStack[len(e.redoStack)-1]
	e.redoStack = e.redoStack[:len(e.redoStack)-1]

	// Restore state
	e.content = state.content
	e.cursorRow = state.cursorRow
	e.cursorCol = state.cursorCol
	e.ensureCursorInBounds()
	e.scrollIntoView()

	return true
}

// CanUndo returns true if there are states to undo
func (e *Editor) CanUndo() bool {
	return len(e.undoStack) > 0
}

// CanRedo returns true if there are states to redo
func (e *Editor) CanRedo() bool {
	return len(e.redoStack) > 0
}
