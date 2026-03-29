package ui

import "strings"

// ansiSlice extracts a visual substring from an ANSI-colored string,
// preserving active style state. startCol and endCol are 0-based visual columns.
// Returns the substring with proper ANSI code handling.
func ansiSlice(s string, startCol, endCol int) string {
	if startCol >= endCol {
		return ""
	}

	var result strings.Builder
	var pendingCodes strings.Builder // ANSI codes active at startCol
	visualPos := 0
	i := 0
	started := false

	for i < len(s) {
		// Check for ANSI escape sequence
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			// Find end of escape sequence
			j := i + 2
			for j < len(s) && !isAnsiTerminator(s[j]) {
				j++
			}
			if j < len(s) {
				j++ // include terminator
			}
			escSeq := s[i:j]

			if !started {
				// Before startCol: accumulate active codes
				pendingCodes.WriteString(escSeq)
			} else if visualPos < endCol {
				// Between startCol and endCol: include in result
				result.WriteString(escSeq)
			}
			i = j
			continue
		}

		if visualPos >= endCol {
			break
		}

		if visualPos >= startCol {
			if !started {
				started = true
				// Emit accumulated ANSI codes
				result.WriteString(pendingCodes.String())
			}
			result.WriteByte(s[i])
		}

		visualPos++
		i++
	}

	// Pad if the string is shorter than endCol
	// First skip up to startCol if we haven't started yet
	if !started && visualPos < startCol {
		visualPos = startCol
	}
	for visualPos < endCol {
		if !started {
			started = true
		}
		result.WriteByte(' ')
		visualPos++
	}

	// Reset styles at the end if we emitted any ANSI codes
	if started && pendingCodes.Len() > 0 {
		result.WriteString("\x1b[0m")
	}

	return result.String()
}

func isAnsiTerminator(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
