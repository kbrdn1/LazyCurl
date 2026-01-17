package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseImportArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		wantCmd *ImportCommand
	}{
		{
			name:    "no args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "only format",
			args:    []string{"openapi"},
			wantErr: true,
		},
		{
			name:    "basic openapi import",
			args:    []string{"openapi", "spec.yaml"},
			wantErr: false,
			wantCmd: &ImportCommand{
				Format:   "openapi",
				FilePath: "spec.yaml",
			},
		},
		{
			name:    "with --name flag",
			args:    []string{"openapi", "spec.yaml", "--name", "My API"},
			wantErr: false,
			wantCmd: &ImportCommand{
				Format:   "openapi",
				FilePath: "spec.yaml",
				Name:     "My API",
			},
		},
		{
			name:    "with --output flag",
			args:    []string{"openapi", "spec.yaml", "--output", "/tmp/api.json"},
			wantErr: false,
			wantCmd: &ImportCommand{
				Format:   "openapi",
				FilePath: "spec.yaml",
				Output:   "/tmp/api.json",
			},
		},
		{
			name:    "with --dry-run flag",
			args:    []string{"openapi", "spec.yaml", "--dry-run"},
			wantErr: false,
			wantCmd: &ImportCommand{
				Format:   "openapi",
				FilePath: "spec.yaml",
				DryRun:   true,
			},
		},
		{
			name:    "with --json flag",
			args:    []string{"openapi", "spec.yaml", "--json"},
			wantErr: false,
			wantCmd: &ImportCommand{
				Format:     "openapi",
				FilePath:   "spec.yaml",
				JSONOutput: true,
			},
		},
		{
			name:    "all flags combined",
			args:    []string{"openapi", "spec.yaml", "--name", "API", "--output", "/tmp/out.json", "--dry-run", "--json"},
			wantErr: false,
			wantCmd: &ImportCommand{
				Format:     "openapi",
				FilePath:   "spec.yaml",
				Name:       "API",
				Output:     "/tmp/out.json",
				DryRun:     true,
				JSONOutput: true,
			},
		},
		{
			name:    "unknown flag",
			args:    []string{"openapi", "spec.yaml", "--unknown"},
			wantErr: true,
		},
		{
			name:    "--name without value",
			args:    []string{"openapi", "spec.yaml", "--name"},
			wantErr: true,
		},
		{
			name:    "--output without value",
			args:    []string{"openapi", "spec.yaml", "--output"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseImportArgs(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if cmd.Format != tt.wantCmd.Format {
				t.Errorf("Format = %q, want %q", cmd.Format, tt.wantCmd.Format)
			}
			if cmd.FilePath != tt.wantCmd.FilePath {
				t.Errorf("FilePath = %q, want %q", cmd.FilePath, tt.wantCmd.FilePath)
			}
			if cmd.Name != tt.wantCmd.Name {
				t.Errorf("Name = %q, want %q", cmd.Name, tt.wantCmd.Name)
			}
			if cmd.Output != tt.wantCmd.Output {
				t.Errorf("Output = %q, want %q", cmd.Output, tt.wantCmd.Output)
			}
			if cmd.DryRun != tt.wantCmd.DryRun {
				t.Errorf("DryRun = %v, want %v", cmd.DryRun, tt.wantCmd.DryRun)
			}
			if cmd.JSONOutput != tt.wantCmd.JSONOutput {
				t.Errorf("JSONOutput = %v, want %v", cmd.JSONOutput, tt.wantCmd.JSONOutput)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple name",
			input: "MyAPI",
			want:  "MyAPI",
		},
		{
			name:  "with spaces",
			input: "My API Name",
			want:  "My-API-Name",
		},
		{
			name:  "with special characters",
			input: "API@v1.0!",
			want:  "APIv10",
		},
		{
			name:  "empty string",
			input: "",
			want:  "imported-api",
		},
		{
			name:  "only special characters",
			input: "@#$%",
			want:  "imported-api",
		},
		{
			name:  "with underscores and dashes",
			input: "my-api_v2",
			want:  "my-api_v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRunOpenAPIImport_DryRun(t *testing.T) {
	// Get the fixture path
	fixturePath := filepath.Join("..", "..", "testdata", "openapi", "minimal-3.0.json")

	// Check if fixture exists
	if _, err := os.Stat(fixturePath); os.IsNotExist(err) {
		t.Skip("test fixture not found")
	}

	cmd := &ImportCommand{
		Format:   "openapi",
		FilePath: fixturePath,
		DryRun:   true,
	}

	// Dry-run should not return an error for a valid file
	err := RunImportCommand(cmd)
	if err != nil {
		t.Errorf("dry-run import failed: %v", err)
	}
}

func TestRunImportCommand_UnsupportedFormat(t *testing.T) {
	cmd := &ImportCommand{
		Format:   "unknown",
		FilePath: "test.txt",
	}

	err := RunImportCommand(cmd)
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}
