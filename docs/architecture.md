# Architecture Documentation

Technical architecture and design patterns for LazyCurl.

## Table of Contents

- [Overview](#overview)
- [Project Structure](#project-structure)
- [Core Architecture](#core-architecture)
- [UI Layer](#ui-layer)
- [Data Layer](#data-layer)
- [State Management](#state-management)
- [Key Patterns](#key-patterns)
- [Extension Points](#extension-points)

---

## Overview

LazyCurl is built on the **Elm Architecture** (Model-View-Update) via the Bubble Tea framework. This provides predictable state management and clean separation of concerns.

### Technology Stack

| Layer | Technology | Purpose |
|-------|------------|---------|
| Framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) | TUI framework (Elm arch) |
| Styling | [Lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling |
| Components | [Bubbles](https://github.com/charmbracelet/bubbles) | Pre-built TUI components |
| Mouse Support | [Bubble Zone](https://github.com/lrstanley/bubblezone) | Mouse interaction |
| Config | [yaml.v3](https://gopkg.in/yaml.v3) | YAML parsing |
| Language | Go 1.21+ | Core implementation |

### Design Principles

1. **Keyboard-First**: Full functionality without mouse
2. **Vim-Motion Inspired**: `h/j/k/l` navigation, modal editing
3. **File-Based Storage**: JSON/YAML files, git-friendly
4. **Modular Panels**: Independent, composable UI components
5. **Two-Tier Config**: Global preferences + workspace settings

---

## Project Structure

```
LazyCurl/
├── cmd/
│   └── lazycurl/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/                     # HTTP client and data models
│   │   ├── collection.go        # Collection file handling
│   │   ├── console.go           # Request/response history
│   │   ├── environment.go       # Environment file handling
│   │   ├── http.go              # HTTP request execution
│   │   └── variables.go         # Variable substitution
│   ├── config/                  # Configuration management
│   │   └── config.go            # Global & workspace config
│   ├── format/                  # Response formatting
│   │   └── formatter.go         # JSON/XML/HTML formatting
│   ├── session/                 # Session persistence
│   │   └── session.go           # Session save/load
│   └── ui/                      # User interface
│       ├── model.go             # Main Bubble Tea model
│       ├── mode.go              # Vim-style modes
│       ├── left_panel.go        # Collections/Environments tabs
│       ├── collections_view.go  # Collection tree view
│       ├── environments_view.go # Environment management
│       ├── request_view.go      # Request builder
│       ├── response_view.go     # Response viewer
│       ├── statusbar.go         # Status bar
│       ├── command_input.go     # Command mode input
│       └── components/          # Reusable UI components
│           ├── dialog.go        # Input dialogs
│           ├── modal.go         # Modal windows
│           ├── search.go        # Search functionality
│           ├── tabs.go          # Tab navigation
│           ├── table.go         # Key-value tables
│           ├── tree.go          # Tree view component
│           ├── editor.go        # Text editor
│           └── whichkey.go      # Keybinding hints
├── pkg/
│   └── styles/
│       └── styles.go            # Catppuccin theme & styles
├── docs/                        # Documentation
├── Makefile                     # Build commands
└── go.mod                       # Dependencies
```

### Package Responsibilities

| Package | Responsibility |
|---------|----------------|
| `cmd/lazycurl` | Entry point, initialization |
| `internal/api` | HTTP client, data models, file I/O |
| `internal/config` | Configuration loading/saving |
| `internal/format` | Response body formatting |
| `internal/session` | Session state persistence |
| `internal/ui` | User interface, Bubble Tea models |
| `pkg/styles` | Theme colors, reusable styles |

---

## Core Architecture

### Elm Architecture (MVU)

```
┌─────────────────────────────────────────────────────────┐
│                      Bubble Tea                          │
├─────────────────────────────────────────────────────────┤
│                                                          │
│   ┌──────────┐     ┌──────────┐     ┌──────────┐        │
│   │  Model   │────▶│  Update  │────▶│   View   │        │
│   │ (State)  │     │ (Logic)  │     │  (UI)    │        │
│   └──────────┘     └──────────┘     └──────────┘        │
│        ▲                │                               │
│        │                │                               │
│        └────────────────┘                               │
│              Messages (tea.Msg)                         │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

### Main Model Structure

```go
type Model struct {
    // Configuration
    globalConfig    *config.GlobalConfig
    workspaceConfig *config.WorkspaceConfig
    workspacePath   string

    // Dimensions
    width  int
    height int

    // Panel system
    activePanel  PanelType
    leftPanel    *LeftPanel
    requestPanel *RequestView
    responsePanel *ResponseView

    // Mode system
    mode         Mode
    statusBar    *StatusBar
    commandInput *CommandInput

    // Components
    dialog   *components.Dialog
    whichKey *components.WhichKey

    // State
    ready       bool
    zoneManager *zone.Manager
}
```

### Panel Types

```go
const (
    CollectionsPanel PanelType = iota
    RequestPanel
    ResponsePanel
    EnvironmentsPanel
)
```

---

## UI Layer

### Layout Architecture

```
┌─────────────────────┬──────────────────────────────────┐
│                     │                                   │
│   Left Panel        │   Request Panel                  │
│   (1/3 width)       │   (2/3 width, 40% height)        │
│                     │                                   │
│   ┌─────────────┐   ├──────────────────────────────────┤
│   │ Collections │   │                                   │
│   │ Environments│   │   Response Panel                 │
│   └─────────────┘   │   (2/3 width, 60% height)        │
│                     │                                   │
│                     │                                   │
├─────────────────────┴──────────────────────────────────┤
│                     Status Bar                          │
└────────────────────────────────────────────────────────┘
```

### Panel Dimensions

```go
// Layout calculation
leftWidth := m.width / 3
rightWidth := m.width - leftWidth - 1

// Subtract borders/padding
contentHeight := m.height - 2  // Status bar

// Request/Response split (40/60)
requestHeight := (contentHeight * 40) / 100
responseHeight := contentHeight - requestHeight
```

### Component Hierarchy

```
Model
├── LeftPanel
│   ├── CollectionsView
│   │   ├── TreeComponent
│   │   └── SearchComponent
│   └── EnvironmentsView
│       ├── TreeComponent
│       └── SearchComponent
├── RequestView
│   ├── TabsComponent
│   ├── TableComponent (params, headers)
│   └── EditorComponent (body)
├── ResponseView
│   ├── TabsComponent
│   └── EditorComponent (body)
├── StatusBar                    # Mode, method, breadcrumb, env, status
├── CommandInput
├── Dialog
└── WhichKey
```

### StatusBar Architecture

The StatusBar provides real-time context through a structured layout:

```text
+-----------------------------------------------------------------------+
| NORMAL | POST | FULLSCREEN | My API > Users > Create | dev | 200 OK  |
+-----------------------------------------------------------------------+
    |       |         |                |                  |       |
   Mode   Method  Fullscreen      Breadcrumb            Env   Status
```

**Key Features:**

- **Mode Badge**: Colored indicator for NORMAL/INSERT/VIEW/COMMAND
- **HTTP Method Badge**: Color-coded method (GET=green, POST=orange, etc.)
- **Fullscreen Badge**: Optional indicator when a panel is fullscreen
- **Middle Content**: Breadcrumb path, status messages, or keyboard hints
- **Environment Badge**: Active environment name or "NONE"
- **HTTP Status Badge**: Last response status with semantic coloring

**Content Priority (middle area):**

1. Status messages (temporary, auto-dismiss after 2s)
2. Breadcrumb navigation (Collection > Folder > Request)
3. Context-aware keyboard hints (fallback)

For detailed API documentation, see [StatusBar Documentation](statusbar.md).

---

## Data Layer

### Collection Model

```go
// CollectionFile represents a collection file
type CollectionFile struct {
    Name        string              `json:"name"`
    Description string              `json:"description,omitempty"`
    Folders     []Folder            `json:"folders,omitempty"`
    Requests    []CollectionRequest `json:"requests,omitempty"`
    FilePath    string              `json:"-"`
}

// Folder represents a folder in a collection
type Folder struct {
    Name        string               `json:"name"`
    Description string               `json:"description,omitempty"`
    Folders     []Folder             `json:"folders,omitempty"`
    Requests    []CollectionRequest  `json:"requests,omitempty"`
}

// CollectionRequest represents an HTTP request
type CollectionRequest struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description,omitempty"`
    Method      HTTPMethod        `json:"method"`
    URL         string            `json:"url"`
    Headers     map[string]string `json:"headers,omitempty"`
    Body        interface{}       `json:"body,omitempty"`
    Tests       []Test            `json:"tests,omitempty"`
}
```

### Environment Model

```go
// EnvironmentFile represents an environment
type EnvironmentFile struct {
    Name        string                          `json:"name"`
    Description string                          `json:"description,omitempty"`
    Variables   map[string]*EnvironmentVariable `json:"variables"`
    FilePath    string                          `json:"-"`
}

// EnvironmentVariable represents a variable
type EnvironmentVariable struct {
    Value    string `json:"value"`
    Secret   bool   `json:"secret,omitempty"`
    Active   bool   `json:"active"`
}
```

### Configuration Model

```go
// GlobalConfig - user-wide settings
type GlobalConfig struct {
    Theme         ThemeConfig
    KeyBindings   KeyBindings
    Editor        string
    Workspaces    []string
    LastWorkspace string
}

// WorkspaceConfig - project-specific settings
type WorkspaceConfig struct {
    Name        string
    Description string
    DefaultEnv  string
    Collections []string
}
```

---

## State Management

### Mode System

The mode system controls user interaction context and is displayed in the StatusBar.

```go
type Mode int

const (
    NormalMode  Mode = iota // Default navigation mode
    InsertMode              // Text input mode
    ViewMode                // Read-only browsing
    CommandMode             // Command line input
)

// Mode methods
func (m Mode) String() string        // Returns display name ("NORMAL", etc.)
func (m Mode) Color() lipgloss.Style // Returns mode-specific colored style
func (m Mode) AllowsInput() bool     // True for INSERT, COMMAND
func (m Mode) AllowsNavigation() bool // True for NORMAL, VIEW
```

See [StatusBar Documentation](statusbar.md) for mode badge styling details.

### Mode Transitions

```
                    ┌─────────────┐
         Esc        │   NORMAL    │        :
    ┌───────────────│   (default) │───────────────┐
    │               └─────────────┘               │
    │                   │     │                   │
    │              i    │     │    v              │
    │                   ▼     ▼                   ▼
┌───┴────┐        ┌─────────┐ ┌─────────┐   ┌─────────┐
│ INSERT │        │ INSERT  │ │  VIEW   │   │ COMMAND │
│        │◄───────│         │ │         │   │         │
└────────┘   Esc  └─────────┘ └─────────┘   └─────────┘
```

### Message Flow

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 1. Update WhichKey context
    m.updateWhichKeyContext()

    // 2. Handle WhichKey if visible
    if m.whichKey.IsVisible() {
        // Handle WhichKey messages
    }

    // 3. Handle Dialog if visible
    if m.dialog.IsVisible() {
        // Handle Dialog messages
    }

    // 4. Handle Command mode
    if m.mode == CommandMode {
        // Handle Command input
    }

    // 5. Dispatch to active panel
    switch m.activePanel {
    case CollectionsPanel:
        m.leftPanel.Update(msg)
    case RequestPanel:
        m.requestPanel.Update(msg)
    case ResponsePanel:
        m.responsePanel.Update(msg)
    }

    return m, cmd
}
```

### Session Persistence

Session state is managed by the `internal/session` package and persisted to `.lazycurl/session.yml`.

```go
// Session represents the complete application state
type Session struct {
    Version           int         `yaml:"version"`
    LastUpdated       time.Time   `yaml:"last_updated"`
    ActivePanel       string      `yaml:"active_panel"`
    ActiveCollection  string      `yaml:"active_collection,omitempty"`
    ActiveRequest     string      `yaml:"active_request,omitempty"`
    ActiveEnvironment string      `yaml:"active_environment,omitempty"`
    Panels            PanelsState `yaml:"panels"`
}
```

**Key Features:**

- **Auto-save**: State changes trigger debounced saves (500ms delay)
- **Atomic writes**: Uses temp file + rename for safe file operations
- **Version control**: Session format version for future migrations
- **Graceful degradation**: Invalid/missing sessions fall back to defaults

**Save/Load Flow:**

```
┌─────────────────────────────────────────────────────────────┐
│ Startup                                                      │
│   LoadSession() → Validate() → Apply to Model               │
├─────────────────────────────────────────────────────────────┤
│ Runtime                                                      │
│   State change → Mark dirty → 500ms debounce → Save()       │
├─────────────────────────────────────────────────────────────┤
│ Shutdown                                                     │
│   Final Save() before exit                                  │
└─────────────────────────────────────────────────────────────┘
```

---

## Key Patterns

### 1. Keybinding Matching

```go
// matchKey checks if input matches any configured keybinding
func (m *Model) matchKey(input string, bindings []string) bool {
    for _, binding := range bindings {
        if input == binding {
            return true
        }
    }
    return false
}

// Usage
if m.matchKey(msg.String(), m.globalConfig.KeyBindings.Quit) {
    return m, tea.Quit
}
```

### 2. Variable Substitution

```go
// ReplaceVariables substitutes {{var}} patterns
func ReplaceVariables(text string, vars map[string]*EnvironmentVariable) string {
    re := regexp.MustCompile(`\{\{(\$?\w+)\}\}`)
    return re.ReplaceAllStringFunc(text, func(match string) string {
        name := match[2 : len(match)-2]

        // Check system variables first
        if strings.HasPrefix(name, "$") {
            return getSystemVariable(name)
        }

        // Check environment variables
        if v, ok := vars[name]; ok && v.Active {
            return v.Value
        }

        return match // Leave unchanged if not found
    })
}
```

### 3. Panel Rendering

```go
// renderPanel creates a styled panel box
func (m Model) renderPanel(title string, content string, active bool) string {
    var borderColor, titleFg lipgloss.Color

    if active {
        borderColor = styles.Lavender
        titleFg = styles.Lavender
    } else {
        borderColor = styles.Surface0
        titleFg = styles.Subtext0
    }

    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(borderColor).
        Render(content)
}
```

### 4. Tree Navigation

```go
// TreeNode represents a navigable tree item
type TreeNode struct {
    Name     string
    Type     NodeType  // Folder or Request
    Expanded bool
    Children []*TreeNode
    Parent   *TreeNode
    Depth    int
}

// Navigate implements j/k/h/l navigation
func (t *Tree) Navigate(direction string) {
    switch direction {
    case "j": t.moveDown()
    case "k": t.moveUp()
    case "h": t.collapseOrParent()
    case "l": t.expandOrEnter()
    }
}
```

### 5. Component Composition

```go
// Components implement a standard interface
type Component interface {
    Update(msg tea.Msg) (Component, tea.Cmd)
    View(width, height int) string
}

// Panels compose multiple components
type LeftPanel struct {
    tabs           *components.Tabs
    collectionsView *CollectionsView
    environmentsView *EnvironmentsView
    activeTab      int
}
```

---

## Extension Points

### Adding a New Panel

1. **Create the panel view** (`internal/ui/new_panel.go`):

```go
type NewPanel struct {
    width, height int
    // ... panel-specific state
}

func NewNewPanel() *NewPanel {
    return &NewPanel{}
}

func (p *NewPanel) Update(msg tea.Msg, globalConfig *config.GlobalConfig) (*NewPanel, tea.Cmd) {
    // Handle messages
    return p, nil
}

func (p *NewPanel) View(width, height int, active bool) string {
    // Render panel
    return ""
}
```

2. **Add panel type** (`internal/ui/model.go`):

```go
const (
    CollectionsPanel PanelType = iota
    RequestPanel
    ResponsePanel
    NewPanel  // Add new panel type
)
```

3. **Initialize in Model**:

```go
func NewModel(...) Model {
    return Model{
        // ...
        newPanel: NewNewPanel(),
    }
}
```

4. **Add routing in Update**:

```go
case NewPanel:
    m.newPanel, cmd = m.newPanel.Update(msg, m.globalConfig)
```

### Adding a New Component

1. **Create component** (`internal/ui/components/new_component.go`):

```go
type NewComponent struct {
    // Component state
}

func NewNewComponent() *NewComponent {
    return &NewComponent{}
}

func (c *NewComponent) Update(msg tea.Msg) (*NewComponent, tea.Cmd) {
    return c, nil
}

func (c *NewComponent) View(width, height int) string {
    return ""
}
```

2. **Use in panels**:

```go
type SomePanel struct {
    myComponent *components.NewComponent
}
```

### Adding a New Keybinding

1. **Add to KeyBindings struct** (`internal/config/config.go`):

```go
type KeyBindings struct {
    // ...
    NewAction []string `yaml:"new_action"`
}
```

2. **Add default** (`internal/config/config.go`):

```go
func DefaultKeyBindings() KeyBindings {
    return KeyBindings{
        // ...
        NewAction: []string{"ctrl+n"},
    }
}
```

3. **Handle in Update**:

```go
if m.matchKey(msg.String(), m.globalConfig.KeyBindings.NewAction) {
    // Handle action
}
```

### Adding a New HTTP Feature

1. **Extend Request model** (`internal/api/collection.go`):

```go
type CollectionRequest struct {
    // ...
    NewField string `json:"new_field,omitempty"`
}
```

2. **Add UI controls** in `request_view.go`

3. **Handle in HTTP execution** (`internal/api/http.go`)

---

## Testing Strategy

### Unit Tests

Located alongside source files (`*_test.go`):

```go
// internal/api/collection_test.go
func TestLoadCollection(t *testing.T) {
    tests := []struct {
        name    string
        path    string
        want    *CollectionFile
        wantErr bool
    }{
        // Test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := LoadCollection(tt.path)
            // Assertions
        })
    }
}
```

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Specific package
go test ./internal/api/...

# Verbose
go test -v ./...
```

---

## Performance Considerations

### Rendering Optimization

- **Lazy rendering**: Only render visible content
- **String builders**: Use `strings.Builder` for concatenation
- **Style caching**: Cache computed styles

### Memory Management

- **Pointer receivers**: Use `*Struct` for large structs
- **Slice pre-allocation**: `make([]T, 0, capacity)`
- **Clear unused references**: Nil out large objects

### File I/O

- **Lazy loading**: Load collections on demand
- **Caching**: Cache parsed files in memory
- **Async saves**: Don't block UI on file writes

---

## Contributing Architecture Changes

When modifying architecture:

1. **Document changes** in this file
2. **Update CLAUDE.md** with patterns
3. **Add tests** for new components
4. **Follow existing patterns** for consistency
5. **Consider backwards compatibility**
