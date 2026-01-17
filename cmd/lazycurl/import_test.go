package main

import (
	"testing"
)

func TestParseImportArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantFormat string
		wantFile   string
		wantErr    bool
	}{
		{
			name:       "file only defaults to auto",
			args:       []string{"collection.json"},
			wantFormat: "auto",
			wantFile:   "collection.json",
			wantErr:    false,
		},
		{
			name:       "explicit auto format",
			args:       []string{"auto", "collection.json"},
			wantFormat: "auto",
			wantFile:   "collection.json",
			wantErr:    false,
		},
		{
			name:       "explicit postman format",
			args:       []string{"postman", "collection.json"},
			wantFormat: "postman",
			wantFile:   "collection.json",
			wantErr:    false,
		},
		{
			name:       "explicit openapi format",
			args:       []string{"openapi", "spec.yaml"},
			wantFormat: "openapi",
			wantFile:   "spec.yaml",
			wantErr:    false,
		},
		{
			name:       "format flag overrides",
			args:       []string{"collection.json", "--format", "postman"},
			wantFormat: "postman",
			wantFile:   "collection.json",
			wantErr:    false,
		},
		{
			name:    "missing file path",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "format without file",
			args:    []string{"postman"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseImportArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseImportArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if cmd.Format != tt.wantFormat {
				t.Errorf("ParseImportArgs() format = %v, want %v", cmd.Format, tt.wantFormat)
			}
			if cmd.FilePath != tt.wantFile {
				t.Errorf("ParseImportArgs() file = %v, want %v", cmd.FilePath, tt.wantFile)
			}
		})
	}
}

func TestParseImportArgsWithFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantName   string
		wantOutput string
		wantDryRun bool
		wantJSON   bool
	}{
		{
			name:       "name flag",
			args:       []string{"file.json", "--name", "MyCollection"},
			wantName:   "MyCollection",
			wantOutput: "",
			wantDryRun: false,
			wantJSON:   false,
		},
		{
			name:       "output flag",
			args:       []string{"file.json", "--output", "/tmp/out.json"},
			wantName:   "",
			wantOutput: "/tmp/out.json",
			wantDryRun: false,
			wantJSON:   false,
		},
		{
			name:       "dry-run flag",
			args:       []string{"file.json", "--dry-run"},
			wantName:   "",
			wantOutput: "",
			wantDryRun: true,
			wantJSON:   false,
		},
		{
			name:       "json flag",
			args:       []string{"file.json", "--json"},
			wantName:   "",
			wantOutput: "",
			wantDryRun: false,
			wantJSON:   true,
		},
		{
			name:       "all flags",
			args:       []string{"postman", "file.json", "--name", "Test", "--output", "out.json", "--dry-run", "--json"},
			wantName:   "Test",
			wantOutput: "out.json",
			wantDryRun: true,
			wantJSON:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := ParseImportArgs(tt.args)
			if err != nil {
				t.Fatalf("ParseImportArgs() unexpected error = %v", err)
			}
			if cmd.Name != tt.wantName {
				t.Errorf("ParseImportArgs() name = %v, want %v", cmd.Name, tt.wantName)
			}
			if cmd.Output != tt.wantOutput {
				t.Errorf("ParseImportArgs() output = %v, want %v", cmd.Output, tt.wantOutput)
			}
			if cmd.DryRun != tt.wantDryRun {
				t.Errorf("ParseImportArgs() dryRun = %v, want %v", cmd.DryRun, tt.wantDryRun)
			}
			if cmd.JSONOutput != tt.wantJSON {
				t.Errorf("ParseImportArgs() json = %v, want %v", cmd.JSONOutput, tt.wantJSON)
			}
		})
	}
}
