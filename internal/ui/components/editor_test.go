package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestEditor_EmptyContent verifies editor handles empty content without panics
func TestEditor_EmptyContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		action  func(*Editor) (*Editor, tea.Cmd)
	}{
		{
			name:    "empty string initializes correctly",
			content: "",
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e, nil
			},
		},
		{
			name:    "navigation in empty content - j key",
			content: "",
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, true)
			},
		},
		{
			name:    "navigation in empty content - k key",
			content: "",
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}, true)
			},
		},
		{
			name:    "navigation in empty content - h key",
			content: "",
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}, true)
			},
		},
		{
			name:    "navigation in empty content - l key",
			content: "",
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}, true)
			},
		},
		{
			name:    "delete in empty content - x key",
			content: "",
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, true)
			},
		},
		{
			name:    "delete line in empty content - d key",
			content: "",
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}, true)
			},
		},
		{
			name:    "undo in empty content - u key",
			content: "",
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}, true)
			},
		},
		{
			name:    "new line below in empty content - o key",
			content: "",
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}, true)
			},
		},
		{
			name:    "new line above in empty content - O key",
			content: "",
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'O'}}, true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic occurred: %v", r)
				}
			}()

			editor := NewEditor(tt.content, "text")
			_, _ = tt.action(editor)
			// Test passes if no panic occurs
		})
	}
}

// TestEditor_CursorAtBoundaries verifies cursor operations at content boundaries
func TestEditor_CursorAtBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		content string
		setup   func(*Editor)
		action  func(*Editor) (*Editor, tea.Cmd)
		wantRow int
		wantCol int
	}{
		{
			name:    "single character - move right at end",
			content: "a",
			setup: func(e *Editor) {
				e.cursorCol = 1 // At end of line
			},
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}, true)
			},
			wantRow: 0,
			wantCol: 1, // Should stay at end
		},
		{
			name:    "single character - move left at start",
			content: "a",
			setup: func(e *Editor) {
				e.cursorCol = 0
			},
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}}, true)
			},
			wantRow: 0,
			wantCol: 0, // Should stay at start
		},
		{
			name:    "single line - end of line command",
			content: "hello",
			setup:   func(e *Editor) {},
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'$'}}, true)
			},
			wantRow: 0,
			wantCol: 5,
		},
		{
			name:    "single line - start of line command",
			content: "hello",
			setup: func(e *Editor) {
				e.cursorCol = 3
			},
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}}, true)
			},
			wantRow: 0,
			wantCol: 0,
		},
		{
			name:    "delete at end of line",
			content: "ab",
			setup: func(e *Editor) {
				e.cursorCol = 2 // At end
			},
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, true)
			},
			wantRow: 0,
			wantCol: 2, // Should stay, nothing to delete
		},
		{
			name:    "multiline - move down at last line",
			content: "line1\nline2",
			setup: func(e *Editor) {
				e.cursorRow = 1
			},
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}, true)
			},
			wantRow: 1, // Should stay at last line
			wantCol: 0,
		},
		{
			name:    "multiline - move up at first line",
			content: "line1\nline2",
			setup: func(e *Editor) {
				e.cursorRow = 0
			},
			action: func(e *Editor) (*Editor, tea.Cmd) {
				return e.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}, true)
			},
			wantRow: 0, // Should stay at first line
			wantCol: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic occurred: %v", r)
				}
			}()

			editor := NewEditor(tt.content, "text")
			if tt.setup != nil {
				tt.setup(editor)
			}
			editor, _ = tt.action(editor)

			if editor.cursorRow != tt.wantRow {
				t.Errorf("cursorRow = %d, want %d", editor.cursorRow, tt.wantRow)
			}
			if editor.cursorCol != tt.wantCol {
				t.Errorf("cursorCol = %d, want %d", editor.cursorCol, tt.wantCol)
			}
		})
	}
}

// TestUndo_EmitsContentChangedMsg verifies undo emits EditorContentChangedMsg
func TestUndo_EmitsContentChangedMsg(t *testing.T) {
	editor := NewEditor("initial", "text")

	// Make a change to have something to undo
	editor.mode = EditorInsertMode
	editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}, true)
	editor.mode = EditorNormalMode

	// Save state and make another change
	editor.saveState()
	editor.content[0] = "modified"

	// Undo should emit content changed message
	_, cmd := editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}, true)

	if cmd == nil {
		t.Error("undo did not return a command, expected EditorContentChangedMsg")
		return
	}

	// Execute the command to get the message
	msg := cmd()
	if _, ok := msg.(EditorContentChangedMsg); !ok {
		t.Errorf("expected EditorContentChangedMsg, got %T", msg)
	}
}

// TestRedo_EmitsContentChangedMsg verifies redo emits EditorContentChangedMsg
func TestRedo_EmitsContentChangedMsg(t *testing.T) {
	editor := NewEditor("initial", "text")

	// Make changes and undo them to have something to redo
	editor.saveState()
	editor.content[0] = "modified"
	editor.undo()

	// Test with ctrl+r (the actual redo keybinding)
	editor.saveState()
	editor.content[0] = "modified2"
	editor.undo()

	editor, cmd := editor.handleNormalMode(tea.KeyMsg{Type: tea.KeyCtrlR})

	if cmd == nil {
		t.Error("redo did not return a command, expected EditorContentChangedMsg")
		return
	}

	msg := cmd()
	if _, ok := msg.(EditorContentChangedMsg); !ok {
		t.Errorf("expected EditorContentChangedMsg, got %T", msg)
	}
}

// TestSearch_MatchAtEndOfContent verifies search finds matches at content end
func TestSearch_MatchAtEndOfContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		query     string
		wantFound bool
		wantRow   int
		wantCol   int
	}{
		{
			name:      "match at end of single line",
			content:   "hello world test",
			query:     "test",
			wantFound: true,
			wantRow:   0,
			wantCol:   12,
		},
		{
			name:      "match at end of last line",
			content:   "line1\nline2\nfinal test",
			query:     "test",
			wantFound: true,
			wantRow:   2,
			wantCol:   6,
		},
		{
			name:      "match is entire content",
			content:   "test",
			query:     "test",
			wantFound: true,
			wantRow:   0,
			wantCol:   0,
		},
		{
			name:      "no match",
			content:   "hello world",
			query:     "xyz",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor := NewEditor(tt.content, "text")
			editor.searchQuery = tt.query
			editor.findMatches()

			if tt.wantFound {
				if len(editor.searchMatches) == 0 {
					t.Error("expected matches but found none")
					return
				}

				// Check that we can find the match at the expected position
				found := false
				for _, match := range editor.searchMatches {
					if match.Row == tt.wantRow && match.ColStart == tt.wantCol {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected match at row %d, col %d, but not found in matches: %+v",
						tt.wantRow, tt.wantCol, editor.searchMatches)
				}
			} else {
				if len(editor.searchMatches) > 0 {
					t.Errorf("expected no matches but found %d", len(editor.searchMatches))
				}
			}
		})
	}
}

// TestSearch_LoopBoundary verifies search loop handles boundary conditions
func TestSearch_LoopBoundary(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "empty content search",
			content: "",
		},
		{
			name:    "single character content",
			content: "a",
		},
		{
			name:    "single line no match",
			content: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic in search rendering: %v", r)
				}
			}()

			editor := NewEditor(tt.content, "text")
			editor.searchQuery = "xyz" // Search for something that won't match
			editor.findMatches()

			// Render the view to trigger the search loop
			_ = editor.View(80, 24, true)
		})
	}
}

// TestExtractVariableName verifies variable name extraction from {{variable}} patterns
func TestExtractVariableName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple variable",
			input:    "{{test}}",
			expected: "test",
		},
		{
			name:     "variable with underscore",
			input:    "{{my_var}}",
			expected: "my_var",
		},
		{
			name:     "variable with dollar sign",
			input:    "{{$timestamp}}",
			expected: "$timestamp",
		},
		{
			name:     "variable with numbers",
			input:    "{{var123}}",
			expected: "var123",
		},
		{
			name:     "empty variable",
			input:    "{{}}",
			expected: "",
		},
		{
			name:     "too short input",
			input:    "{{",
			expected: "",
		},
		{
			name:     "minimal valid",
			input:    "{{a}}",
			expected: "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVariableName(tt.input)
			if result != tt.expected {
				t.Errorf("extractVariableName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestEditorVariablePattern verifies the regex pattern matches valid variables
func TestEditorVariablePattern(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantMatch  bool
		wantGroups []string
	}{
		{
			name:       "simple variable",
			input:      "{{test}}",
			wantMatch:  true,
			wantGroups: []string{"{{test}}", "test"},
		},
		{
			name:       "variable with underscore",
			input:      "{{my_var}}",
			wantMatch:  true,
			wantGroups: []string{"{{my_var}}", "my_var"},
		},
		{
			name:       "variable with dollar",
			input:      "{{$uuid}}",
			wantMatch:  true,
			wantGroups: []string{"{{$uuid}}", "$uuid"},
		},
		{
			name:       "variable with numbers",
			input:      "{{var123}}",
			wantMatch:  true,
			wantGroups: []string{"{{var123}}", "var123"},
		},
		{
			name:      "invalid - spaces inside",
			input:     "{{ test }}",
			wantMatch: false,
		},
		{
			name:      "invalid - special chars",
			input:     "{{test-var}}",
			wantMatch: false,
		},
		{
			name:      "invalid - empty",
			input:     "{{}}",
			wantMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := editorVariablePattern.FindStringSubmatch(tt.input)
			if tt.wantMatch {
				if matches == nil {
					t.Errorf("expected match for %q but got none", tt.input)
					return
				}
				for i, want := range tt.wantGroups {
					if i >= len(matches) || matches[i] != want {
						t.Errorf("group %d: got %q, want %q", i, matches[i], want)
					}
				}
			} else {
				if matches != nil {
					t.Errorf("expected no match for %q but got %v", tt.input, matches)
				}
			}
		})
	}
}

// TestEditorPreviewMode verifies preview mode toggle and content resolution
func TestEditorPreviewMode(t *testing.T) {
	editor := NewEditor(`{"url": "{{base_url}}/api"}`, "json")
	editor.SetVariableValues(map[string]string{
		"base_url": "https://api.test.com",
	})

	// Initial state - preview off
	if editor.IsPreviewMode() {
		t.Error("preview mode should be off by default")
	}

	// Content should be unchanged
	content := editor.GetContent()
	if content != `{"url": "{{base_url}}/api"}` {
		t.Errorf("unexpected content: %s", content)
	}

	// Toggle preview on
	editor.TogglePreviewMode()
	if !editor.IsPreviewMode() {
		t.Error("preview mode should be on after toggle")
	}

	// Preview content should have resolved values
	previewContent := editor.GetPreviewContent()
	expected := `{"url": "https://api.test.com/api"}`
	if previewContent != expected {
		t.Errorf("preview content = %q, want %q", previewContent, expected)
	}

	// Original content should still be unchanged
	content = editor.GetContent()
	if content != `{"url": "{{base_url}}/api"}` {
		t.Errorf("original content was modified: %s", content)
	}

	// Toggle preview off
	editor.TogglePreviewMode()
	if editor.IsPreviewMode() {
		t.Error("preview mode should be off after second toggle")
	}
}

// TestEditorPreviewMode_UnresolvedVariables verifies unresolved variables remain unchanged
func TestEditorPreviewMode_UnresolvedVariables(t *testing.T) {
	editor := NewEditor(`{"token": "{{auth_token}}"}`, "json")
	editor.SetVariableValues(map[string]string{
		"other_var": "value", // Different variable
	})

	editor.TogglePreviewMode()
	previewContent := editor.GetPreviewContent()

	// Unresolved variable should remain as-is
	if previewContent != `{"token": "{{auth_token}}"}` {
		t.Errorf("unresolved variable was incorrectly replaced: %s", previewContent)
	}
}

// TestEditorPreviewMode_NoVariables verifies preview mode works with no variables set
func TestEditorPreviewMode_NoVariables(t *testing.T) {
	editor := NewEditor(`{"key": "value"}`, "json")

	editor.TogglePreviewMode()
	previewContent := editor.GetPreviewContent()

	if previewContent != `{"key": "value"}` {
		t.Errorf("content with no variables was modified: %s", previewContent)
	}
}

// TestEditorPreviewMode_KeybindingP verifies P key toggles preview mode
func TestEditorPreviewMode_KeybindingP(t *testing.T) {
	editor := NewEditor(`{"url": "{{base_url}}"}`, "json")
	editor.SetVariableValues(map[string]string{"base_url": "https://api.com"})

	// Initial state
	if editor.IsPreviewMode() {
		t.Error("preview mode should be off initially")
	}

	// Press P key in normal mode
	editor, _ = editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}}, true)

	if !editor.IsPreviewMode() {
		t.Error("preview mode should be on after P key")
	}

	// Press P again
	editor, _ = editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}}, true)

	if editor.IsPreviewMode() {
		t.Error("preview mode should be off after second P key")
	}
}
