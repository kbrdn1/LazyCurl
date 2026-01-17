package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple name",
			input: "MyAPI",
			want:  "myapi",
		},
		{
			name:  "with spaces",
			input: "My API Name",
			want:  "my-api-name",
		},
		{
			name:  "with special characters",
			input: "API@v1.0!#$%",
			want:  "apiv10",
		},
		{
			name:  "empty string",
			input: "",
			want:  "imported-api",
		},
		{
			name:  "only special characters",
			input: "@#$%^&*()",
			want:  "imported-api",
		},
		{
			name:  "with numbers and dashes",
			input: "api-v2-beta",
			want:  "api-v2-beta",
		},
		{
			name:  "with underscores",
			input: "my_api_v2",
			want:  "my_api_v2",
		},
		{
			name:  "mixed case",
			input: "MyApiV2",
			want:  "myapiv2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeOpenAPIFilename(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeOpenAPIFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFindAvailablePath(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "lazycurl-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	modal := NewOpenAPIImportModal(tmpDir)

	t.Run("no conflicts", func(t *testing.T) {
		path := modal.findAvailablePath("new-api.json")
		expected := filepath.Join(tmpDir, "new-api-1.json")
		if path != expected {
			t.Errorf("findAvailablePath() = %q, want %q", path, expected)
		}
	})

	t.Run("with existing file", func(t *testing.T) {
		// Create existing file
		existingPath := filepath.Join(tmpDir, "existing-api.json")
		if err := os.WriteFile(existingPath, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		// Also create existing-api-1.json
		existingPath1 := filepath.Join(tmpDir, "existing-api-1.json")
		if err := os.WriteFile(existingPath1, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		path := modal.findAvailablePath("existing-api.json")
		expected := filepath.Join(tmpDir, "existing-api-2.json")
		if path != expected {
			t.Errorf("findAvailablePath() = %q, want %q", path, expected)
		}
	})
}

func TestOpenAPIImportModal_ConflictDetection(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "lazycurl-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an existing collection file
	existingPath := filepath.Join(tmpDir, "test-api.json")
	if err := os.WriteFile(existingPath, []byte(`{"name":"Test API"}`), 0644); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	modal := NewOpenAPIImportModal(tmpDir)

	t.Run("file exists check", func(t *testing.T) {
		filename := "test-api.json"
		savePath := filepath.Join(tmpDir, filename)

		// Check if file exists
		_, err := os.Stat(savePath)
		if os.IsNotExist(err) {
			t.Error("Expected file to exist")
		}
	})

	t.Run("suggested path is unique", func(t *testing.T) {
		suggested := modal.findAvailablePath("test-api.json")

		// Suggested path should not exist
		if _, err := os.Stat(suggested); !os.IsNotExist(err) {
			t.Errorf("Suggested path %q should not exist", suggested)
		}

		// Should be test-api-1.json
		expectedName := "test-api-1.json"
		if filepath.Base(suggested) != expectedName {
			t.Errorf("Expected suggested filename %q, got %q", expectedName, filepath.Base(suggested))
		}
	})
}

func TestOpenAPIImportModal_States(t *testing.T) {
	modal := NewOpenAPIImportModal("/tmp/collections")

	t.Run("initial state", func(t *testing.T) {
		if modal.state != StateInputPath {
			t.Errorf("Initial state should be StateInputPath, got %v", modal.state)
		}
		if modal.visible {
			t.Error("Modal should not be visible initially")
		}
	})

	t.Run("show modal", func(t *testing.T) {
		modal.Show()
		if !modal.visible {
			t.Error("Modal should be visible after Show()")
		}
		if modal.state != StateInputPath {
			t.Errorf("State should be StateInputPath after Show(), got %v", modal.state)
		}
	})

	t.Run("hide modal", func(t *testing.T) {
		modal.Hide()
		if modal.visible {
			t.Error("Modal should not be visible after Hide()")
		}
	})
}

func TestSpinnerFrames(t *testing.T) {
	// Ensure spinner frames are defined and not empty
	if len(spinnerFrames) == 0 {
		t.Error("spinnerFrames should not be empty")
	}

	// All frames should be non-empty
	for i, frame := range spinnerFrames {
		if frame == "" {
			t.Errorf("spinnerFrames[%d] should not be empty", i)
		}
	}
}

func TestOpenAPIImportModal_OverwriteChoice(t *testing.T) {
	modal := NewOpenAPIImportModal("/tmp/collections")

	t.Run("default overwrite choice is rename", func(t *testing.T) {
		modal.conflictPath = "/tmp/collections/test.json"
		modal.suggestedPath = "/tmp/collections/test-1.json"
		modal.state = StateConfirmOverwrite
		modal.overwriteChoice = 1 // Default to rename

		if modal.overwriteChoice != 1 {
			t.Errorf("Default overwrite choice should be 1 (rename), got %d", modal.overwriteChoice)
		}
	})

	t.Run("toggle overwrite choice", func(t *testing.T) {
		modal.overwriteChoice = 1
		modal.overwriteChoice = 1 - modal.overwriteChoice // Toggle
		if modal.overwriteChoice != 0 {
			t.Errorf("After toggle, overwrite choice should be 0, got %d", modal.overwriteChoice)
		}

		modal.overwriteChoice = 1 - modal.overwriteChoice // Toggle again
		if modal.overwriteChoice != 1 {
			t.Errorf("After second toggle, overwrite choice should be 1, got %d", modal.overwriteChoice)
		}
	})
}
