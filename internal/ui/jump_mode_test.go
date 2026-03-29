package ui

import (
	"strings"
	"testing"
)

func TestInsertAtVisualPosition(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		col        int
		insertText string
		replaceLen int
		wantSame   bool // if true, expect line returned unchanged
	}{
		{
			name:       "position found - simple insertion",
			line:       "Hello World",
			col:        6,
			insertText: "[a]",
			replaceLen: 1,
			wantSame:   false,
		},
		{
			name:       "position not found - col beyond line width",
			line:       "Short",
			col:        50,
			insertText: "[a]",
			replaceLen: 1,
			wantSame:   true, // should return line unchanged
		},
		{
			name:       "position at start",
			line:       "Hello",
			col:        0,
			insertText: "[x]",
			replaceLen: 1,
			wantSame:   false,
		},
		{
			name:       "empty line",
			line:       "",
			col:        0,
			insertText: "[a]",
			replaceLen: 1,
			wantSame:   true, // empty line, no position to insert
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := insertAtVisualPosition(tt.line, tt.col, tt.insertText, tt.replaceLen)
			if tt.wantSame && got != tt.line {
				t.Errorf("expected line unchanged, got %q", got)
			}
			if !tt.wantSame && got == tt.line {
				t.Errorf("expected line modified, got same %q", got)
			}
		})
	}
}

func TestAssignLabels(t *testing.T) {
	tests := []struct {
		name       string
		numTargets int
		wantSingle bool // true = single-char labels, false = two-char
	}{
		{
			name:       "5 targets get single-char labels",
			numTargets: 5,
			wantSingle: true,
		},
		{
			name:       "26 targets get single-char labels",
			numTargets: 26,
			wantSingle: true,
		},
		{
			name:       "27 targets get two-char labels",
			numTargets: 27,
			wantSingle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jm := NewJumpMode()
			targets := make([]JumpTarget, tt.numTargets)
			for i := range targets {
				targets[i] = JumpTarget{
					Row:    i,
					Col:    0,
					Label:  "",
					Action: JumpSelect,
					Panel:  CollectionsPanel,
				}
			}
			jm.SetTargets(targets)

			visible := jm.GetVisibleTargets()
			if len(visible) != tt.numTargets {
				t.Fatalf("expected %d visible targets, got %d", tt.numTargets, len(visible))
			}

			for _, target := range visible {
				if tt.wantSingle && len(target.Label) != 1 {
					t.Errorf("expected single-char label, got %q", target.Label)
					break
				}
				if !tt.wantSingle && len(target.Label) != 2 {
					t.Errorf("expected two-char label, got %q", target.Label)
					break
				}
			}
		})
	}
}

func TestRenderOverlayPreservesLineCount(t *testing.T) {
	jm := NewJumpMode()

	// Create a simple background
	bg := "Line 1 content here\nLine 2 content here\nLine 3 content here\nLine 4 content here\nLine 5 content here"
	bgLineCount := len(strings.Split(bg, "\n"))

	// Create targets
	targets := []JumpTarget{
		{Row: 1, Col: 0, Label: "a", Action: JumpSelect, Panel: CollectionsPanel},
		{Row: 3, Col: 5, Label: "s", Action: JumpSelect, Panel: CollectionsPanel},
	}

	jm.Activate(CollectionsPanel)
	jm.SetTargets(targets)

	result := jm.RenderOverlay(bg, 80, 5)
	resultLines := strings.Split(result, "\n")

	if len(resultLines) != bgLineCount {
		t.Errorf("RenderOverlay changed line count: got %d, want %d", len(resultLines), bgLineCount)
	}
}
