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

### Git Worktrees (Feature Branches)

**For humans**: Use the interactive menu:

```bash
make worktree       # Interactive worktree manager with gwq
```

**For Claude Code**: Use non-interactive `make worktree-create`:

```bash
# Create worktree for a feature (copies all necessary files + runs make deps)
make worktree-create TYPE=feat ISSUE=35 DESC=js-scripting

# This automatically:
# - Creates branch: feat/#35-js-scripting
# - Creates worktree at: ~/cc-worktree/LazyCurl/feat-35-js-scripting
# - Copies: CLAUDE.md, .lazycurl/, .claude/commands/, .specify/
# - Creates: specs/035-js-scripting/ folder
# - Runs: make deps

# List worktrees
make worktree-list

# Cleanup stale references
make worktree-cleanup
```

**After creation**: The user launches a new Claude Code session in the worktree:

```bash
cd ~/cc-worktree/LazyCurl/feat-35-js-scripting
claude
```

**Branch naming convention**: `<type>/#<issue>-<description>`

- Types: `feat`, `fix`, `hotfix`, `docs`, `test`, `refactor`, `chore`, `perf`, `ci`, `build`
- Example: `feat/#35-js-scripting`, `fix/#42-request-chaining`

**Remove worktree when done**:

```bash
gwq remove -b feat/#35-js-scripting
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

**External Editor Integration** (`internal/api/`, `internal/ui/components/`):

- **Editor Detection** (`external_editor.go`): Auto-detect from `$VISUAL` → `$EDITOR` → fallback (`nano`, `vi`)
- **Temp File Management** (`temp_file.go`): Create temp files with smart extensions (`.json`, `.xml`, `.html`, `.txt`)
- **Content Type Detection**: Heuristic analysis based on content prefix (`{`/`[` → JSON, `<?xml` → XML, `<!doctype` → HTML)
- **Message Types** (`editor_messages.go`): `ExternalEditorRequestMsg`, `ExternalEditorFinishedMsg`, `ExternalEditorErrorMsg`
- **Error Categorization**: Typed errors (`EditorErrorNoEditor`, `EditorErrorNotFound`, `EditorErrorTempFile`, `EditorErrorReadContent`)

**External Editor Message Flow**:

```text
User presses Ctrl+E (INSERT mode)
    ↓
Editor.handleInsertMode() → ExternalEditorRequestMsg{Field, Content, ContentType}
    ↓
RequestView.Update() → forwards message unchanged
    ↓
Model.openExternalEditor():
  1. GetEditorConfig() → detect editor from env vars
  2. EditorConfig.Validate() → verify binary exists
  3. CreateTempFile() → write content with smart extension
  4. tea.ExecProcess() → suspend TUI, launch editor
    ↓
External editor opens (vim, code --wait, nano...)
User edits and saves
    ↓
Callback:
  1. ReadTempFile() → read modified content
  2. Compare with original
  3. Return ExternalEditorFinishedMsg{Content, Changed, Duration}
    ↓
Model.Update():
  1. CleanupTempFile() → delete temp file
  2. Forward to RequestView
    ↓
RequestView.Update() → bodyEditor.SetContent(newContent)
```

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

**StatusBar** (`internal/ui/statusbar.go`):

- Displays mode (NORMAL/INSERT/VIEW/COMMAND), HTTP method, breadcrumb, environment, HTTP status
- Mode colors defined in `pkg/styles/styles.go` (ModeNormalBg, ModeInsertBg, etc.)
- HTTP method colors: GET=green, POST=orange, PUT=blue, DELETE=red, PATCH=purple
- HTTP status colors: 2xx=green, 3xx=blue, 4xx=orange, 5xx=red
- Temporary messages auto-dismiss after 2 seconds (MessageDuration constant)
- Middle content priority: message > breadcrumb > keyboard hints
- See `docs/statusbar.md` for complete API reference

**Mode System** (`internal/ui/mode.go`):

- Mode enum: NormalMode, ViewMode, CommandMode, InsertMode
- Mode.String() returns display name ("NORMAL", "INSERT", etc.)
- Mode.Color() returns Lipgloss style with background/foreground colors
- Mode.AllowsInput() returns true for INSERT, COMMAND
- Mode.AllowsNavigation() returns true for NORMAL, VIEW

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

**Sprint 1 - MVP** ✅ Complete:

- Load/display collections from JSON files
- Interactive request builder (method, URL, headers, body)
- Send real HTTP requests with variable substitution
- Format and display JSON/XML responses
- Save requests to collections

**Sprint 2 - UX Improvements** ✅ Complete:

- Responsive panel layout, fullscreen toggle
- Find in editors, console tab (request history)
- Improved statusbar, session persistence

**Sprint 3 - Competitive Parity** ✅ Complete (v1.2.0):

- cURL import/export
- Jump mode navigation (vim-easymotion)
- External editor integration
- OpenAPI 3.x import with security schemes
- Postman import/export with CLI support

**Sprint 4 - Competitive Advantage** 🔥 In Progress:

- ✅ JavaScript scripting (Goja) - Pre/Post-request scripts with `lc` API
- ✅ Test assertions - Built-in assertion API with Tests tab
- Request chaining (planned)
- Collection runner (planned)

## Key Dependencies

- **Bubble Tea** (`charmbracelet/bubbletea`): TUI framework (Elm architecture)
- **Bubbles** (`charmbracelet/bubbles`): Pre-built TUI components
- **Lipgloss** (`charmbracelet/lipgloss`): Terminal styling
- **Bubble Zone** (`lrstanley/bubblezone`): Mouse interaction support
- **yaml.v3** (`gopkg.in/yaml.v3`): YAML parsing for configs
- **libopenapi** (`github.com/pb33f/libopenapi`): OpenAPI 3.x parsing and validation
- **Goja** (`github.com/dop251/goja`): JavaScript runtime for scripting

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

- Markdown (GitHub Flavored) + N/A (documentation-only feature) (001-api-documentation)
- Git repository + GitHub Wiki (001-api-documentation)

- Go 1.25 (035-js-scripting)
- JSON files (collections in `.lazycurl/collections/`) (035-js-scripting)

- Go 1.21+ + Bubble Tea (TUI), Lipgloss (styling), Bubble Zone (mouse), yaml.v3 (config) (001-vim-mode-workspace)
- File-based (YAML for config, JSON for collections/environments) in `.lazycurl/` workspace (001-vim-mode-workspace)

## Recent Changes

- 035-js-scripting: Added JavaScript scripting with Goja runtime (pre/post-request scripts, `lc` API, test assertions)
- v1.2.0: Added Postman import/export with CLI support (`lazycurl import postman`)
- v1.2.0: Added OpenAPI security scheme import (Bearer, Basic, API Key)
- v1.1.0: Added cURL import/export, jump mode navigation
- 066-openapi-import: Added OpenAPI 3.x import via TUI (Ctrl+O) and CLI (`lazycurl import openapi`)

## Completed Feature: Console Tab (Issue #9)

### Overview

Console tab in Response Panel for HTTP request/response history logging with keyboard actions.

### Key Files

- `internal/api/console.go` - ConsoleEntry, ConsoleHistory types
- `internal/ui/console_view.go` - Console list view with vim navigation

### Keybindings

- `Tab` / `4`: Switch to Console tab
- `j/k/g/G`: Navigate console list
- `R`: Resend selected request
- `H/B/E/A`: Copy headers/body/error/all to clipboard

## Completed Feature: External Editor Integration (Issue #65)

### Overview

External editor integration allows users to edit request body/headers in their preferred text editor (vim, VS Code, etc.) by pressing `Ctrl+E` in INSERT mode.

### Key Files

- `internal/api/external_editor.go` - Editor detection, content type analysis, validation
- `internal/api/temp_file.go` - Temp file lifecycle management
- `internal/api/headers.go` - Header text serialization for editor
- `internal/ui/components/editor.go` - Editor component with external edit trigger
- `internal/ui/components/editor_messages.go` - Message type definitions
- `internal/ui/model.go` - Process orchestration via `tea.ExecProcess`

### Configuration

External editor is configured via environment variables (not config file):

```bash
export VISUAL="vim"           # Primary (preferred)
export EDITOR="nano"          # Fallback
export VISUAL="code --wait"   # GUI editors need --wait flag
```

### Keybindings

- `Ctrl+E` (INSERT mode): Open current field in external editor

### Architecture Pattern

```text
Ctrl+E → ExternalEditorRequestMsg → CreateTempFile → tea.ExecProcess
       → Editor process → ExternalEditorFinishedMsg → CleanupTempFile → Update content
```

### Content Type Detection

| Content | Extension | Detection |
|---------|-----------|-----------|
| JSON | `.json` | Starts with `{` or `[` |
| XML | `.xml` | Starts with `<?xml` or `<tag>` |
| HTML | `.html` | Starts with `<!doctype` or `<html>` |
| Text | `.txt` | Default fallback |

---

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

```text
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

## Completed Feature: OpenAPI 3.x Import (Issue #66)

### Overview

Import OpenAPI 3.x specifications (JSON/YAML) into LazyCurl collections via TUI modal or CLI.

### Key Files

- `internal/api/openapi.go` - OpenAPIImporter, version detection, error handling
- `internal/api/openapi_converter.go` - Conversion to LazyCurl collection format
- `internal/ui/openapi_import_modal.go` - TUI import modal component
- `cmd/lazycurl/import.go` - CLI import subcommand

### Supported Versions

- OpenAPI 3.0.x (full support)
- OpenAPI 3.1.x (full support)
- Swagger 2.0 (rejected with upgrade guidance)

### UI Import (Ctrl+O)

1. Opens import modal with file path input
2. Validates OpenAPI spec and shows preview (endpoints, tags)
3. Creates collection organized by tags (folders)
4. Untagged operations go to "Untagged" folder

### CLI Import

```bash
lazycurl import openapi <file>           # Basic import
lazycurl import openapi spec.yaml --name "My API"  # Custom name
lazycurl import openapi spec.yaml --output /path/to/collection.json
lazycurl import openapi spec.yaml --dry-run   # Preview without saving
lazycurl import openapi spec.yaml --json      # JSON output for scripting
```

### Architecture Pattern

```text
OpenAPI File → OpenAPIImporter.Parse() → BuildV3Model() → ToCollection()
                    ↓                         ↓              ↓
             Version check            $ref resolution    Tag-to-folder mapping
```

### Conversion Features

- **Tag Organization**: Operations grouped by tags into folders
- **Path Parameters**: Extracted to URL with `{param}` syntax
- **Query Parameters**: Added to request params with examples
- **Headers**: Extracted from parameter definitions
- **Request Bodies**: JSON body with schema-generated examples
- **$ref Resolution**: Automatic via libopenapi's BuildV3Model()
- **Circular Refs**: Handled with depth limit (> 5 levels)
- **Security Schemes**: Extracts Bearer, Basic, API Key auth from security definitions

### Security Scheme Support (v1.2.0)

| Scheme Type | OpenAPI Config | LazyCurl AuthConfig |
|-------------|----------------|---------------------|
| Bearer | `type: http, scheme: bearer` | `Type: "bearer", Token: ""` |
| Basic | `type: http, scheme: basic` | `Type: "basic", Username: "", Password: ""` |
| API Key Header | `type: apiKey, in: header` | `Type: "apikey", APIKeyLocation: "header"` |
| API Key Query | `type: apiKey, in: query` | `Type: "apikey", APIKeyLocation: "query"` |

### Dependencies

- `github.com/pb33f/libopenapi` - OpenAPI parsing and validation

## Completed Feature: Postman Import/Export (Issue #14, #72)

### Overview

Import Postman collections and environments into LazyCurl via TUI modal (`Ctrl+P`) or CLI.

### Key Files

- `internal/import/postman/` - Postman import/export package
  - `types.go` - Postman collection/environment types
  - `importer.go` - Import logic with auto-detection
  - `converter.go` - Conversion to LazyCurl format
  - `exporter.go` - Export to Postman format
- `internal/ui/postman_commands.go` - TUI import commands
- `cmd/lazycurl/import.go` - CLI import subcommand

### Supported Formats

- Postman Collection v2.0 and v2.1
- Postman Environment files
- Auto-detection of file type

### TUI Import (`Ctrl+P`)

1. Opens import modal with file path input
2. Auto-detects collection vs environment
3. Preserves folder structure and request organization
4. Converts variables to LazyCurl format

### CLI Import

```bash
lazycurl import postman collection.json         # Import collection
lazycurl import postman environment.json        # Import environment (auto-detected)
lazycurl import collection.json                 # Auto-detect format (postman/openapi)
lazycurl import postman collection.json --dry-run  # Preview without saving
lazycurl import postman collection.json --json  # JSON output for scripting
```

### Conversion Features

- **Folder Structure**: Preserves Postman folder hierarchy
- **Variables**: Converts `{{var}}` syntax (same format)
- **Authentication**: Imports Bearer, Basic, API Key settings
- **Request Bodies**: Converts raw, form-data, urlencoded
- **Headers**: Preserves all headers with enabled/disabled state

---

## Feature: JavaScript Scripting (Issue #35)

### Overview

JavaScript scripting support via the Goja runtime allows users to write pre-request and post-response scripts for dynamic request manipulation, environment variable management, and test assertions.

### Key Files

- `internal/api/script_executor.go` - ScriptExecutor interface and Goja implementation
- `internal/api/script_runtime.go` - Runtime setup with `lc` global object
- `internal/api/script_request.go` - ScriptRequest for request manipulation
- `internal/api/script_response.go` - ScriptResponse for response access
- `internal/api/script_env.go` - ScriptEnvironment for variable management
- `internal/api/script_console.go` - Console logging API
- `internal/api/script_assertions.go` - Test assertions API
- `internal/ui/request_view.go` - Script editor tabs (Pre-request, Post-response)
- `internal/ui/response_view.go` - Tests tab for assertion results
- `internal/ui/model.go` - Script execution integration with HTTP flow

### Script Editor UI

The Request Panel now includes two new tabs:

- **Pre-request** (Tab 5): Script runs before the HTTP request is sent
- **Post-response** (Tab 6): Script runs after the response is received

### JavaScript API (`lc` global object)

```javascript
// Request object (pre-request only, read-only in post-response)
lc.request.method      // HTTP method (readonly)
lc.request.url         // Request URL
lc.request.getHeader(name)
lc.request.setHeader(name, value)
lc.request.removeHeader(name)
lc.request.body        // Request body

// Response object (post-response only)
lc.response.status     // Status code (e.g., 200)
lc.response.statusText // Status text (e.g., "200 OK")
lc.response.getHeader(name)
lc.response.body       // Response body
lc.response.json()     // Parse body as JSON
lc.response.time       // Response time in ms

// Environment variables
lc.env.get(name)       // Get variable value
lc.env.set(name, value) // Set variable (persists after script)
lc.env.unset(name)     // Remove variable
lc.env.has(name)       // Check if variable exists

// Console logging
lc.console.log(msg)    // Log message
lc.console.info(msg)   // Info message
lc.console.warn(msg)   // Warning message
lc.console.error(msg)  // Error message

// Test assertions
lc.test.assert(name, condition)           // Basic assertion
lc.test.assertEqual(name, actual, expected)
lc.test.assertNotEqual(name, actual, expected)
lc.test.assertTrue(name, condition)
lc.test.assertFalse(name, condition)
lc.test.assertContains(name, haystack, needle)
lc.test.assertStatus(expectedStatus)      // Assert response status
lc.test.assertHeader(name, expectedValue) // Assert header value
```

### Script Execution Flow

```text
User presses Ctrl+S (Send)
    ↓
sendHTTPRequest() checks for pre-request script
    ↓
If pre-request script exists:
  1. ExecutePreRequestScriptCmd runs script with Goja
  2. Script can modify request (URL, headers, body)
  3. Script can set environment variables
  4. PreRequestScriptResultMsg returns with modified request
    ↓
HTTP request sent with modifications applied
    ↓
HTTPResponseMsg received
    ↓
If post-response script exists:
  1. ExecutePostResponseScriptCmd runs script
  2. Script can access response, run assertions
  3. Script can modify environment variables
  4. PostResponseScriptResultMsg returns with results
    ↓
Assertions displayed in Tests tab
Environment changes persisted to file
```

### Tests Tab (Response Panel)

The Response Panel now includes a "Tests" tab (Tab 4) showing:

- Summary: "Tests: X passed, Y failed"
- List of all assertion results with ✓/✗ icons
- Detailed error messages for failed assertions
- Vim navigation (j/k/g/G) through results

### Example Scripts

**Pre-request: Add timestamp header**

```javascript
lc.request.setHeader("X-Timestamp", Date.now().toString());
```

**Pre-request: Dynamic authentication**

```javascript
var token = lc.env.get("auth_token");
if (token) {
    lc.request.setHeader("Authorization", "Bearer " + token);
}
```

**Post-response: Store token from response**

```javascript
var data = lc.response.json();
if (data && data.token) {
    lc.env.set("auth_token", data.token);
    lc.console.log("Token saved: " + data.token.substring(0, 10) + "...");
}
```

**Post-response: Test assertions**

```javascript
lc.test.assertStatus(200);
lc.test.assertHeader("Content-Type", "application/json");

var data = lc.response.json();
lc.test.assert("Response has data", data !== null);
lc.test.assert("User ID is present", data.id !== undefined);
lc.test.assertEqual("Status is active", data.status, "active");
```

### Keybindings

- `Tab 5`: Switch to Pre-request script editor (Request Panel)
- `Tab 6`: Switch to Post-response script editor (Request Panel)
- `Tab 4` / `4`: Switch to Tests tab (Response Panel)
- `Tab 5` / `5`: Switch to Console tab (Response Panel)
- `j/k/g/G`: Navigate test results in Tests tab

### Architecture Patterns

**Script Executor Interface**

```go
type ScriptExecutor interface {
    ExecutePreRequest(script string, req *ScriptRequest, env *Environment) (*ScriptResult, error)
    ExecutePostResponse(script string, req *ScriptRequest, resp *ScriptResponse, env *Environment) (*ScriptResult, error)
    SetTimeout(timeout time.Duration)
}
```

**Message-Based Integration**

```go
PreRequestScriptResultMsg{
    Result      *api.ScriptResult
    ModifiedReq *api.ScriptRequest
    Error       error
    OriginalReq *api.Request
}

PostResponseScriptResultMsg{
    Result *api.ScriptResult
    Error  error
}
```

### Dependencies

- `github.com/dop251/goja` - JavaScript runtime for Go
