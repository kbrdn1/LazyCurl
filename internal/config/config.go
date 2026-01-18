package config

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Script configuration defaults
const (
	DefaultScriptTimeout = 5 * time.Second
	DefaultScriptEnabled = true
)

// ScriptConfig holds scripting-related configuration
type ScriptConfig struct {
	// Timeout is the maximum execution time for scripts
	Timeout time.Duration `yaml:"timeout"`
	// Enabled controls whether scripting is active
	Enabled bool `yaml:"enabled"`
}

// DefaultScriptConfig returns default script configuration
func DefaultScriptConfig() ScriptConfig {
	return ScriptConfig{
		Timeout: DefaultScriptTimeout,
		Enabled: DefaultScriptEnabled,
	}
}

// GlobalConfig represents the global configuration
type GlobalConfig struct {
	Theme         ThemeConfig             `yaml:"theme"`
	KeyBindings   KeyBindings             `yaml:"keybindings"`
	Editor        string                  `yaml:"editor"`
	Workspaces    []string                `yaml:"workspaces"` // List of recent workspaces
	LastWorkspace string                  `yaml:"last_workspace"`
	Environments  map[string]*Environment `yaml:"global_environments,omitempty"`
	Script        ScriptConfig            `yaml:"script"`
}

// WorkspaceConfig represents a workspace configuration (.lazycurl/config.yaml)
type WorkspaceConfig struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	DefaultEnv  string   `yaml:"default_env,omitempty"`
	Collections []string `yaml:"collections,omitempty"`
}

// ThemeConfig represents theme configuration
type ThemeConfig struct {
	Name           string `yaml:"name"`
	PrimaryColor   string `yaml:"primary_color"`
	SecondaryColor string `yaml:"secondary_color"`
	AccentColor    string `yaml:"accent_color"`
	BorderColor    string `yaml:"border_color"`
	ActiveColor    string `yaml:"active_color"`
}

// KeyBindings represents customizable key bindings
type KeyBindings struct {
	Quit             []string `yaml:"quit"`
	NavigateLeft     []string `yaml:"navigate_left"`
	NavigateRight    []string `yaml:"navigate_right"`
	NavigateUp       []string `yaml:"navigate_up"`
	NavigateDown     []string `yaml:"navigate_down"`
	Select           []string `yaml:"select"`
	Back             []string `yaml:"back"`
	NewRequest       []string `yaml:"new_request"`
	SendRequest      []string `yaml:"send_request"`
	SaveRequest      []string `yaml:"save_request"`
	DeleteRequest    []string `yaml:"delete_request"`
	FocusCollections []string `yaml:"focus_collections"`
	FocusRequest     []string `yaml:"focus_request"`
	FocusResponse    []string `yaml:"focus_response"`
	ToggleEnvs       []string `yaml:"toggle_envs"`
	ImportCurl       []string `yaml:"import_curl"`
	ExportCurl       []string `yaml:"export_curl"`
	ImportOpenAPI    []string `yaml:"import_openapi"`
	RunCollection    []string `yaml:"run_collection"`
}

// Environment represents an environment with variables
type Environment struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Variables   map[string]string `yaml:"variables"`
}

// DefaultGlobalConfig returns default global configuration
func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Theme: ThemeConfig{
			Name:           "dark",
			PrimaryColor:   "#7D56F4",
			SecondaryColor: "#00D9FF",
			AccentColor:    "#FF6B6B",
			BorderColor:    "#3C3C3C",
			ActiveColor:    "#00FF00",
		},
		KeyBindings: DefaultKeyBindings(),
		Editor:      "vim",
		Workspaces:  []string{},
		Script:      DefaultScriptConfig(),
	}
}

// DefaultKeyBindings returns default vim-style key bindings
func DefaultKeyBindings() KeyBindings {
	return KeyBindings{
		Quit:             []string{"q"},
		NavigateLeft:     []string{"h"},
		NavigateRight:    []string{"l"},
		NavigateUp:       []string{"k"},
		NavigateDown:     []string{"j"},
		Select:           []string{"enter"},
		Back:             []string{"esc"},
		NewRequest:       []string{"n"},
		SendRequest:      []string{"ctrl+s"},
		SaveRequest:      []string{"ctrl+w"},
		DeleteRequest:    []string{"d"},
		FocusCollections: []string{},
		FocusRequest:     []string{},
		FocusResponse:    []string{},
		ToggleEnvs:       []string{"e"},
		ImportCurl:       []string{"ctrl+i"},
		ExportCurl:       []string{"ctrl+e"},
		ImportOpenAPI:    []string{"ctrl+o"},
		RunCollection:    []string{"ctrl+r"},
	}
}

// DefaultWorkspaceConfig returns default workspace configuration
func DefaultWorkspaceConfig() *WorkspaceConfig {
	return &WorkspaceConfig{
		Name:        "My Workspace",
		Description: "",
		Collections: []string{},
	}
}

// LoadGlobalConfig loads global configuration from file
func LoadGlobalConfig() (*GlobalConfig, error) {
	path := GetGlobalConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultGlobalConfig(), nil
		}
		return nil, err
	}

	var config GlobalConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveGlobalConfig saves global configuration to file
func (c *GlobalConfig) Save() error {
	path := GetGlobalConfigPath()

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// LoadWorkspaceConfig loads workspace configuration
func LoadWorkspaceConfig(workspacePath string) (*WorkspaceConfig, error) {
	configPath := filepath.Join(workspacePath, ".lazycurl", "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultWorkspaceConfig(), nil
		}
		return nil, err
	}

	var config WorkspaceConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveWorkspaceConfig saves workspace configuration
func (c *WorkspaceConfig) Save(workspacePath string) error {
	lazycurlPath := filepath.Join(workspacePath, ".lazycurl")
	if err := os.MkdirAll(lazycurlPath, 0755); err != nil {
		return err
	}

	// Create subdirectories (errors intentionally ignored - these are optional)
	_ = os.MkdirAll(filepath.Join(lazycurlPath, "collections"), 0755)
	_ = os.MkdirAll(filepath.Join(lazycurlPath, "envs"), 0755)

	configPath := filepath.Join(lazycurlPath, "config.yaml")

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetGlobalConfigPath returns the global config file path
func GetGlobalConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".lazycurl/config.yaml"
	}
	return filepath.Join(home, ".config", "lazycurl", "config.yaml")
}

// GetWorkspacePath returns the workspace path (current directory)
func GetWorkspacePath() (string, error) {
	return os.Getwd()
}

// InitWorkspace initializes a new workspace in the current directory
func InitWorkspace(name string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	config := &WorkspaceConfig{
		Name:        name,
		Description: "",
		Collections: []string{},
	}

	return config.Save(cwd)
}
