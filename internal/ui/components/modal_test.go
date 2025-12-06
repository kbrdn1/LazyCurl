package components

import (
	"testing"
)

// TestModal_EmptyPlaceholder verifies modal handles empty placeholder without panics
func TestModal_EmptyPlaceholder(t *testing.T) {
	tests := []struct {
		name        string
		placeholder string
		value       string
		focused     bool
	}{
		{
			name:        "empty placeholder, empty value, focused",
			placeholder: "",
			value:       "",
			focused:     true,
		},
		{
			name:        "empty placeholder, empty value, not focused",
			placeholder: "",
			value:       "",
			focused:     false,
		},
		{
			name:        "with placeholder, empty value, focused",
			placeholder: "Enter name",
			value:       "",
			focused:     true,
		},
		{
			name:        "single char placeholder, empty value, focused",
			placeholder: "x",
			value:       "",
			focused:     true,
		},
		{
			name:        "empty placeholder, with value, focused",
			placeholder: "",
			value:       "test",
			focused:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic occurred: %v", r)
				}
			}()

			modal := NewInputModal("Test", "Field", tt.placeholder, "test")
			modal.Fields[0].Value = tt.value
			if tt.focused {
				modal.FocusIndex = 0
			} else {
				modal.FocusIndex = 1 // Focus on button
			}
			modal.Show()

			// Render the modal to trigger the placeholder rendering code
			_ = modal.View(80, 24)
		})
	}
}

// TestModal_CursorBounds verifies cursor positioning handles edge cases
func TestModal_CursorBounds(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		cursorPos int
	}{
		{
			name:      "cursor at start",
			value:     "test",
			cursorPos: 0,
		},
		{
			name:      "cursor at end",
			value:     "test",
			cursorPos: 4,
		},
		{
			name:      "cursor beyond end",
			value:     "test",
			cursorPos: 10,
		},
		{
			name:      "empty value, cursor at 0",
			value:     "",
			cursorPos: 0,
		},
		{
			name:      "single char, cursor at end",
			value:     "a",
			cursorPos: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic occurred: %v", r)
				}
			}()

			modal := NewInputModal("Test", "Field", "", "test")
			modal.Fields[0].Value = tt.value
			modal.Fields[0].CursorPos = tt.cursorPos
			modal.FocusIndex = 0
			modal.Show()

			// Render the modal to trigger cursor rendering
			_ = modal.View(80, 24)
		})
	}
}
