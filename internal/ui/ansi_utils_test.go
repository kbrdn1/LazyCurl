package ui

import (
	"testing"
)

func TestAnsiSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		startCol int
		endCol   int
		wantLen  int // visual length of result
	}{
		{
			name:     "plain string full",
			input:    "Hello World",
			startCol: 0,
			endCol:   11,
			wantLen:  11,
		},
		{
			name:     "plain string subset",
			input:    "Hello World",
			startCol: 6,
			endCol:   11,
			wantLen:  5,
		},
		{
			name:     "empty range",
			input:    "Hello",
			startCol: 3,
			endCol:   3,
			wantLen:  0,
		},
		{
			name:     "beyond string length pads with spaces",
			input:    "Hi",
			startCol: 0,
			endCol:   5,
			wantLen:  5,
		},
		{
			name:     "with ANSI codes preserves style",
			input:    "\x1b[31mRed Text\x1b[0m Normal",
			startCol: 0,
			endCol:   8,
			wantLen:  8,
		},
		{
			name:     "slice starts mid-styled region",
			input:    "\x1b[31mRed Text\x1b[0m",
			startCol: 4,
			endCol:   8,
			wantLen:  4,
		},
		{
			name:     "startCol past end of string",
			input:    "Short",
			startCol: 10,
			endCol:   15,
			wantLen:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ansiSlice(tt.input, tt.startCol, tt.endCol)
			// Verify visual length (strip ANSI codes to count)
			gotVisual := stripAnsi(got)
			if len(gotVisual) != tt.wantLen {
				t.Errorf("ansiSlice() visual length = %d, want %d, got string = %q", len(gotVisual), tt.wantLen, got)
			}
		})
	}
}

// stripAnsi removes ANSI escape codes for visual length measurement
func stripAnsi(s string) string {
	var result []byte
	i := 0
	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			j := i + 2
			for j < len(s) && !isAnsiTerminator(s[j]) {
				j++
			}
			if j < len(s) {
				j++
			}
			i = j
			continue
		}
		result = append(result, s[i])
		i++
	}
	return string(result)
}
