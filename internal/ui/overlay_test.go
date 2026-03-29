package ui

import (
	"strings"
	"testing"
)

func TestOverlayDialogPreservesBackground(t *testing.T) {
	tests := []struct {
		name            string
		bgLine          string
		dialogLine      string
		startCol        int
		totalWidth      int
		wantLeftPrefix  string
		wantRightSuffix string
	}{
		{
			name:            "plain background preserved on both sides",
			bgLine:          "AAAAAAAAAA",
			dialogLine:      "DIALOG",
			startCol:        2,
			totalWidth:      10,
			wantLeftPrefix:  "AA",
			wantRightSuffix: "AA",
		},
		{
			name:            "dialog at start preserves right side",
			bgLine:          "BBBBBBBBBB",
			dialogLine:      "DLG",
			startCol:        0,
			totalWidth:      10,
			wantLeftPrefix:  "",
			wantRightSuffix: "BBBBBBB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lineWidth := len(tt.dialogLine)
			leftBg := ansiSlice(tt.bgLine, 0, tt.startCol)
			rightBg := ansiSlice(tt.bgLine, tt.startCol+lineWidth, tt.totalWidth)
			result := leftBg + tt.dialogLine + rightBg

			// Verify left side is preserved
			if !strings.HasPrefix(result, tt.wantLeftPrefix) {
				t.Errorf("left side: got %q, want prefix %q", result, tt.wantLeftPrefix)
			}

			// Verify right side is preserved
			if !strings.HasSuffix(result, tt.wantRightSuffix) {
				t.Errorf("right side: got %q, want suffix %q", result, tt.wantRightSuffix)
			}

			// Verify dialog content is present
			if !strings.Contains(result, tt.dialogLine) {
				t.Errorf("dialog content missing: got %q, want to contain %q", result, tt.dialogLine)
			}
		})
	}
}
