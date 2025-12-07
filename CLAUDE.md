# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LazyCurl is a Terminal User Interface (TUI) HTTP client that combines Lazygit's multi-panel interface with Postman's API testing capabilities. Built with Go and Bubble Tea framework, it provides a keyboard-driven, vim-motion interface for managing HTTP requests, collections, and environments.

## Development Commands

### Build & Run
```bash
make build          # Compile to bin/lazycurl
make run            # Build and launch the application
make clean          # Remove binaries and caches
```

### Testing
```bash
make test           # Run all tests
make test-coverage  # Run tests with coverage report (generates coverage.html)
go test ./internal/api/...  # Test specific package
```

### Code Quality
```bash
make fmt            # Format code with gofmt
make lint           # Run golangci-lint (if installed)
```

### Development
```bash
make dev            # Live reload with air (auto-installs if missing)
make deps           # Download and tidy dependencies
```

### Multi-Platform Builds
```bash
make build-all      # Cross-compile for Linux/macOS/Windows (AMD64 & ARM64)
```

## Architecture

### Application Structure

**Bubble Tea Pattern**: LazyCurl follows the Elm architecture via Bubble Tea:
- **Model**: `internal/ui/model.go` - Central application state with 4 panels
- **Update**: Message-driven state updates with keybinding dispatch
- **View**: Lipgloss-styled rendering with Lazygit-inspired 3-panel layout

**Panel System**:
```
┌─────────────────┬──────────────────┐
│  Collections    │   Request        │
│  (left 1/3)     │   (top 40%)      │
│                 ├──────────────────┤
│                 │   Response       │
│                 │   (bottom 60%)   │
└─────────────────┴──────────────────┘
│          Status Bar                │
└────────────────────────────────────┘
```

- **CollectionsPanel**: Tree view of JSON collections (`.lazycurl/collections/*.json`)
- **RequestPanel**: Interactive request builder (method, URL, headers, body)
- **ResponsePanel**: Formatted HTTP response viewer
- **EnvironmentsPanel**: Overlay for variable management (toggled with 'e')

### Key Architectural Patterns

**Configuration System** (`internal/config/`):
- **Two-tier config**: Global (`~/.config/lazycurl/config.yaml`) + Workspace (`.lazycurl/config.yaml`)
- Global: Theme, keybindings, editor preference, workspace history
- Workspace: Project name, default environment, collection references
- All configs use YAML serialization via `gopkg.in/yaml.v3`

**Data Layer** (`internal/api/`):
- **Collections**: Hierarchical structure with folders, requests stored as JSON
- **Environments**: Variable substitution with `{{variable}}` syntax
- **HTTP Client**: Request execution with variable interpolation
- **Response Formatting**: JSON/XML/HTML formatting via `internal/format/`

**Session Layer** (`internal/session/`):
- **Session Persistence**: Auto-save/restore of application state to `.lazycurl/session.yml`
- **Debounced Saves**: 500ms delay prevents excessive disk writes during rapid changes
- **Atomic Writes**: Uses temp file + rename pattern for safe file operations
- **Graceful Degradation**: Missing/invalid session files silently fall back to defaults

**UI Components** (`internal/ui/`):
- Each panel is a self-contained view with Update/View pattern
- Active panel receives keyboard input via central dispatcher
- Zone manager (`bubblezone`) enables mouse interactions
- Styles centralized in `pkg/styles/`

### State Flow

1. **Initialization** (`cmd/lazycurl/main.go`):
   - Load global config from `~/.config/lazycurl/`
   - Load workspace config from current dir `.lazycurl/`
   - Initialize Bubble Tea model with configs
   - Enable alt-screen + mouse support

2. **Message Dispatch** (`internal/ui/model.go:Update`):
   - Keybindings checked via `matchKey()` helper
   - Panel navigation: h/l switches active panel
   - Active panel receives messages for further processing
   - Special keys (quit, toggle envs) handled at model level

3. **Panel Updates**:
   - Each panel implements `Update(msg, globalConfig)` pattern
   - Returns updated view state + optional Bubble Tea commands
   - Uses global config for keybinding consistency

### Critical Code Patterns

**Keybinding Matching**:
```go
// All keybindings stored as string arrays for multi-key support
if m.matchKey(msg.String(), m.globalConfig.KeyBindings.Quit) {
    return m, tea.Quit
}
```

**Collection Loading**:
```go
// Collections loaded from .lazycurl/collections/*.json
// Structure: CollectionFile → Folders (recursive) → CollectionRequest
collection, err := api.LoadCollection(path)
```

**Variable Substitution**:
```go
// {{variable}} patterns replaced from environment.Variables map
// Implemented in internal/api/variables.go
resolvedURL := ReplaceVariables(url, env.Variables)
```

## File Structure Conventions

**Package Organization**:
- `cmd/`: Application entry points (only `lazycurl/main.go`)
- `internal/`: Private application code (api, config, ui, format, session)
- `pkg/`: Reusable libraries (currently only styles)
- Root: Configuration files, documentation, Makefile

**Naming**:
- Files: `snake_case.go` (e.g., `collections_view.go`)
- Tests: `*_test.go` suffix alongside source files
- Exported symbols: `PascalCase`
- Private symbols: `camelCase`

## Testing Standards

**Table-Driven Tests**:
All test files use table-driven approach with struct slices:
```go
tests := []struct {
    name    string
    input   X
    want    Y
    wantErr bool
}{...}
```

**Test Coverage**:
- Run `make test-coverage` to generate HTML report
- Target: All public API functions in `internal/api/` must have tests
- UI components tested via Bubble Tea message simulation

## Git Workflow

**Branch Convention**: `<type>/#<issue>-<description>`
- Types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `perf`
- Example: `feat/#12-add-collection-loader`

**Commit Convention**: Gitmoji + Conventional Commits
```
<emoji> <type>(<scope>): <description>

Examples:
✨ feat(collections): add JSON collection loader
🐛 fix(ui): fix panel resize on terminal size change
♻️ refactor(api): refactor request builder logic
```

**Common Scopes**: `ui`, `api`, `config`, `collections`, `environments`, `styles`

## Current State & Next Steps

**Phase 1 - Foundation** ✅ Complete:
- Lazygit-style multi-panel TUI
- Vim motions (h/j/k/l) navigation
- Workspace system (`.lazycurl/` directory)
- YAML configuration (global + workspace)
- Customizable keybindings and themes

**Sprint 1 - MVP** 🔥 In Progress (see DEVELOPMENT_PLAN.md):
- Load/display collections from JSON files
- Interactive request builder (method, URL, headers, body)
- Send real HTTP requests with variable substitution
- Format and display JSON/XML responses
- Save requests to collections

## Key Dependencies

- **Bubble Tea** (`charmbracelet/bubbletea`): TUI framework (Elm architecture)
- **Bubbles** (`charmbracelet/bubbles`): Pre-built TUI components
- **Lipgloss** (`charmbracelet/lipgloss`): Terminal styling
- **Bubble Zone** (`lrstanley/bubblezone`): Mouse interaction support
- **yaml.v3** (`gopkg.in/yaml.v3`): YAML parsing for configs

## Development Notes

**Panel Dimension Calculation**:
The 3-panel layout uses dynamic sizing based on terminal dimensions:
- Left panel: 1/3 width, full height
- Top-right: 2/3 width, 40% height (request builder)
- Bottom-right: 2/3 width, 60% height (response viewer)

**Status Bar Format**:
`Workspace: <name> | <panel> | h/l: Switch Panel | n: New | Ctrl+S: Send | e: Envs | q: Quit`

**Environment Variable Syntax**:
URLs, headers, and body fields support `{{variable_name}}` interpolation from active environment's variable map.

## Common Patterns

**Adding a New Keybinding**:
1. Add to `KeyBindings` struct in `internal/config/config.go`
2. Add default value in `DefaultKeyBindings()`
3. Check in relevant panel's `Update()` method
4. Document in README keyboard shortcuts table

**Adding a New Panel**:
1. Create `*_view.go` in `internal/ui/`
2. Implement struct with `Update(msg, globalConfig)` and `View(width, height, active)` methods
3. Add to `Model` struct and `NewModel()` initialization
4. Add panel type constant and switching logic in `model.go`

**Collection File Format**:
```json
{
  "name": "My API",
  "description": "...",
  "folders": [{"name": "...", "requests": [...]}],
  "requests": [
    {
      "id": "req_123",
      "name": "Get Users",
      "method": "GET",
      "url": "{{base_url}}/users",
      "headers": {"Authorization": "Bearer {{token}}"},
      "body": null
    }
  ]
}
```

## Active Technologies
- Go 1.21+ + Bubble Tea (TUI), Lipgloss (styling), Bubble Zone (mouse), yaml.v3 (config) (001-vim-mode-workspace)
- File-based (YAML for config, JSON for collections/environments) in `.lazycurl/` workspace (001-vim-mode-workspace)

## Recent Changes
- 001-vim-mode-workspace: Added Go 1.21+ + Bubble Tea (TUI), Lipgloss (styling), Bubble Zone (mouse), yaml.v3 (config)

## Current Feature: Console Tab in Response Panel (Issue #9)

**Branch**: `feat/#9-console-tab-in-response-panel`
**Spec**: `specs/009-console-tab-in-response-panel/`

### Feature Summary
Add Console tab to Response Panel for HTTP request/response history logging with keyboard actions.

### Key Implementation Points
- **Data Layer**: `internal/api/console.go` - ConsoleEntry, ConsoleHistory types
- **UI Component**: `internal/ui/console_view.go` - Console list view with vim navigation
- **Integration**: Add "Console" as 4th tab in ResponseView
- **Clipboard**: Use `golang.design/x/clipboard` package

### Keybindings
- `Ctrl+C`: Switch to Console tab
- `Ctrl+R`: Switch to Response tab
- `j/k/g/G`: Navigate console list
- `R`: Resend selected request
- `H/B/E/A`: Copy headers/body/error/all to clipboard

### Architecture Pattern
```
Request sent → RequestCompleteMsg → Add to ConsoleHistory → ConsoleView updates
```

See `specs/009-console-tab-in-response-panel/quickstart.md` for implementation guide.

## Completed Feature: Session Persistence (Issue #11)

### Overview
Session persistence automatically saves and restores application state across sessions.

### Key Files
- `internal/session/session.go` - Session types and Load/Save functions
- `internal/session/session_test.go` - Comprehensive tests

### Session File Format (`.lazycurl/session.yml`)
```yaml
version: 1
last_updated: "2025-12-06T10:30:00Z"
active_panel: request
active_collection: "api.json"
active_request: "req_001"
active_environment: "development"
panels:
  collections:
    expanded_folders: ["Users"]
    scroll_position: 5
    selected_index: 3
  request:
    active_tab: "body"
  response:
    active_tab: "body"
    scroll_position: 0
```

### Architecture Pattern
```
State change → Mark dirty → 500ms debounce → Save to YAML (atomic write)
Startup → LoadSession() → Validate references → Apply to panels
Quit → Final save
```

### What's Persisted
- Active panel (Collections, Request, Response)
- Active collection and request
- Active environment
- Expanded folders in Collections tree
- Scroll positions and cursor positions
- Active tabs in each panel
