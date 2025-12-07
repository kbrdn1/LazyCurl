# StatusBar Component

The StatusBar provides real-time context about the application state, displaying mode indicators, HTTP context, navigation breadcrumbs, and keyboard hints in a compact, information-rich footer.

## Table of Contents

- [Overview](#overview)
- [Layout](#layout)
- [Components](#components)
  - [Mode Badge](#mode-badge)
  - [HTTP Method Badge](#http-method-badge)
  - [Fullscreen Badge](#fullscreen-badge)
  - [Middle Content](#middle-content)
  - [Environment Badge](#environment-badge)
  - [HTTP Status Badge](#http-status-badge)
- [Messages](#messages)
- [Keyboard Hints](#keyboard-hints)
- [Color Reference](#color-reference)
- [API Reference](#api-reference)
- [Integration](#integration)

---

## Overview

The StatusBar is a single-line component rendered at the bottom of the terminal. It provides:

- **Mode awareness**: Colored badge showing NORMAL/INSERT/VIEW/COMMAND
- **HTTP context**: Current method and last response status
- **Navigation**: Breadcrumb path through collections/folders/requests
- **Feedback**: Temporary status messages for user actions
- **Discoverability**: Context-sensitive keyboard hints

The StatusBar automatically adapts to terminal width, truncating content intelligently while preserving critical information.

---

## Layout

```text
+-----------------------------------------------------------------------+
| NORMAL | POST | FULLSCREEN | My API > Users > Create | dev | 200 OK  |
+-----------------------------------------------------------------------+
    |       |         |                |                  |       |
   Mode   Method  Fullscreen      Breadcrumb            Env   Status
```

### Element Priority (left to right)

1. **Mode Badge** - Always visible, indicates current interaction mode
2. **Method Badge** - Visible when HTTP method is set
3. **Fullscreen Badge** - Visible when fullscreen mode is active
4. **Middle Content** - Flexible width, shows breadcrumb/message/hints
5. **Environment Badge** - Always visible, shows active environment or "NONE"
6. **HTTP Status Badge** - Visible after request completes

---

## Components

### Mode Badge

Displays the current vim-style interaction mode with distinctive coloring.

| Mode | Display | Background | Foreground | Description |
|------|---------|------------|------------|-------------|
| NORMAL | `NORMAL` | Blue (#6798da) | White | Default navigation mode |
| INSERT | `INSERT` | Light Gray (#b8bcc2) | Black | Text input mode |
| VIEW | `VIEW` | Green (#4c8c49) | White | Read-only browsing |
| COMMAND | `COMMAND` | Orange (#a45e0e) | White | Command line input |

**Behavior:**

- Updates immediately on mode transition
- Always positioned at the far left
- Includes horizontal padding for visual separation

### HTTP Method Badge

Shows the HTTP method of the current/selected request.

| Method | Background | Foreground |
|--------|------------|------------|
| GET | Green (#4c8c49) | White |
| POST | Orange (#a45e0e) | White |
| PUT | Blue (#6798da) | White |
| DELETE | Red/Coral (#fa827c) | White |
| PATCH | Purple (#d48cee) | White |
| HEAD | Green (#4c8c49) | White |
| OPTIONS | Brown/Taupe (#a48e85) | White |

**Behavior:**

- Only visible when a method is set
- Appears after the mode badge
- Cleared when no request is selected

### Fullscreen Badge

Indicates when fullscreen mode is active for a panel.

| State | Display | Background | Foreground |
|-------|---------|------------|------------|
| Active | `FULLSCREEN` | Mauve (#cba6f7) | Dark (#11111b) |

**Behavior:**

- Only visible when fullscreen is enabled
- Positioned after the method badge

### Middle Content

Flexible-width area displaying contextual information in priority order:

1. **Status Message** - Temporary feedback messages (highest priority)
2. **Breadcrumb** - Navigation path when a request is selected
3. **Keyboard Hints** - Context-aware shortcuts (fallback)

**Truncation Behavior:**

- Content truncates with `...` when exceeding available width
- Uses Unicode-aware width calculation for proper handling
- Minimum 3 characters before truncation applies

### Environment Badge

Shows the active environment name for variable substitution.

| State | Display | Color | Style |
|-------|---------|-------|-------|
| Active | Environment name | Green (#a6e3a1) | Bold |
| None | `NONE` | Gray (#a6adc8) | Normal |

**Behavior:**

- Always visible on the right side
- Updates when environment selection changes

### HTTP Status Badge

Displays the status code and text from the last HTTP response.

| Status Range | Meaning | Background | Foreground |
|--------------|---------|------------|------------|
| 2xx | Success | Green (#4c8c49) | White |
| 3xx | Redirect | Blue (#6798da) | White |
| 4xx | Client Error | Orange (#a45e0e) | White |
| 5xx | Server Error | Red/Coral (#fa827c) | White |

**Display Format:**

- With text: `200 OK`, `404 Not Found`
- Code only: `201` (if no text provided)

**Behavior:**

- Only visible after a request completes
- Can be cleared with `ClearHTTPStatus()`

---

## Messages

The StatusBar supports temporary status messages that auto-dismiss after 2 seconds.

### Message Types

| Type | Method | Duration | Style |
|------|--------|----------|-------|
| Info | `Info(msg)` | 2 seconds | Yellow, Bold |
| Success | `Success(action, target)` | 2 seconds | Yellow, Bold |
| Error | `Error(err)` | 2 seconds | Yellow, Bold |
| Custom | `ShowMessage(msg, duration)` | Custom | Yellow, Bold |

### Message Behavior

- Messages appear immediately in the middle content area
- Messages override breadcrumb and hints while active
- Auto-dismissal occurs after the specified duration
- Manual clearing available via `ClearMessage()`

### Message Format Examples

```
Info:    "Request saved"
Success: "Saved: Create User"
Error:   "Error: Connection refused"
```

---

## Keyboard Hints

When no message or breadcrumb is displayed, context-sensitive keyboard hints appear.

### Mode-Based Hints

| Mode | Default Hints |
|------|---------------|
| NORMAL | `j/k:Up/Down | h/l:Nav | n:New | R:Rename | d:Delete | ?:Help` |
| INSERT | `type:Edit | tab:Next | esc:Normal` |
| VIEW | `j/k:Scroll | g/G:Top/End | h/l:Panel | esc:Normal` |
| COMMAND | `:q:quit | :w:save | :ws:workspace | esc:Cancel` |

### Custom Hints

Panels can set custom hints for specific contexts:

```go
m.statusBar.SetHints("j/k:Scroll | Enter:Select | n:New")
```

Custom hints override mode-based defaults until cleared.

---

## Color Reference

### Mode Colors

```go
ModeNormalBg  = "#6798da" // Blue
ModeNormalFg  = "#FFFFFF" // White
ModeViewBg    = "#4c8c49" // Green
ModeViewFg    = "#FFFFFF" // White
ModeCommandBg = "#a45e0e" // Orange
ModeCommandFg = "#FFFFFF" // White
ModeInsertBg  = "#b8bcc2" // Light Gray
ModeInsertFg  = "#000000" // Black
```

### HTTP Method Colors

```go
MethodGetBg     = "#4c8c49" // Green
MethodPostBg    = "#a45e0e" // Orange
MethodPutBg     = "#6798da" // Blue
MethodDeleteBg  = "#fa827c" // Red/Coral
MethodPatchBg   = "#d48cee" // Purple
MethodHeadBg    = "#4c8c49" // Green
MethodOptionsBg = "#a48e85" // Brown/Taupe
```

### HTTP Status Colors

```go
Status2xxBg = "#4c8c49" // Green (Success)
Status3xxBg = "#6798da" // Blue (Redirect)
Status4xxBg = "#a45e0e" // Orange (Client Error)
Status5xxBg = "#fa827c" // Red/Coral (Server Error)
```

---

## API Reference

### StatusBar Struct

```go
type StatusBar struct {
    mode         Mode      // Current interaction mode
    version      string    // Application version
    width        int       // Available rendering width
    httpStatus   int       // HTTP status code (0 = no response)
    httpText     string    // HTTP status text
    httpMethod   string    // Current HTTP method
    breadcrumb   []string  // Navigation breadcrumb parts
    message      string    // Temporary status message
    messageEnd   time.Time // When to clear the message
    environment  string    // Active environment name
    hints        string    // Custom keyboard hints
    isFullscreen bool      // Fullscreen mode indicator
}
```

### Constructor

```go
func NewStatusBar(version string) *StatusBar
```

Creates a new StatusBar instance with default values.

### Mode Methods

```go
// SetMode updates the mode indicator
func (s *StatusBar) SetMode(mode Mode)

// GetMode returns the current mode
func (s *StatusBar) GetMode() Mode
```

### HTTP Methods

```go
// SetMethod sets the current HTTP method display
func (s *StatusBar) SetMethod(method string)

// SetHTTPStatus sets the HTTP status display
func (s *StatusBar) SetHTTPStatus(code int, text string)

// ClearHTTPStatus clears the HTTP status display
func (s *StatusBar) ClearHTTPStatus()
```

### Context Methods

```go
// SetBreadcrumb sets the navigation breadcrumb
func (s *StatusBar) SetBreadcrumb(parts ...string)

// SetEnvironment sets the active environment name
func (s *StatusBar) SetEnvironment(name string)

// SetHints sets custom keyboard hints
func (s *StatusBar) SetHints(hints string)

// SetFullscreen sets the fullscreen mode indicator
func (s *StatusBar) SetFullscreen(fullscreen bool)
```

### Message Methods

```go
// ShowMessage displays a temporary status message
func (s *StatusBar) ShowMessage(msg string, duration time.Duration)

// Info displays an info message (2s duration)
func (s *StatusBar) Info(msg string)

// Success displays a success message (2s duration)
func (s *StatusBar) Success(action, target string)

// Error displays an error message (2s duration)
func (s *StatusBar) Error(err error)

// ClearMessage clears the status message immediately
func (s *StatusBar) ClearMessage()
```

### Rendering

```go
// View renders the status bar to the given width
func (s *StatusBar) View(width int) string
```

### Message Types (Bubble Tea)

```go
// StatusUpdateMsg signals a status bar update
type StatusUpdateMsg struct {
    Mode       *Mode
    HTTPStatus *int
    HTTPText   *string
    Method     *string
    Breadcrumb []string
    Message    *string
    Duration   time.Duration
}

// Helper functions for creating pointers
func ModePtr(m Mode) *Mode
func IntPtr(i int) *int
func StringPtr(s string) *string
```

---

## Integration

### Basic Setup

```go
// Initialize in model
func NewModel() Model {
    return Model{
        statusBar: NewStatusBar("1.0.0"),
    }
}

// Render in View
func (m Model) View() string {
    statusLine := m.statusBar.View(m.width)
    return lipgloss.JoinVertical(lipgloss.Top, content, statusLine)
}
```

### Updating Mode

```go
// In Update function
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "i" {
            m.statusBar.SetMode(InsertMode)
        }
    }
    return m, nil
}
```

### Setting HTTP Context

```go
// When selecting a request
m.statusBar.SetMethod(request.Method)

// After receiving response
m.statusBar.SetHTTPStatus(resp.StatusCode, http.StatusText(resp.StatusCode))
```

### Setting Breadcrumb

```go
// When navigating collections
m.statusBar.SetBreadcrumb("My API", "Users", "Create User")

// Clear breadcrumb
m.statusBar.SetBreadcrumb()
```

### Displaying Messages

```go
// Info message
m.statusBar.Info("Request duplicated")

// Success with context
m.statusBar.Success("Saved", "Create User")

// Error handling
if err != nil {
    m.statusBar.Error(err)
}
```

### Custom Panel Hints

```go
// In panel's Update method
m.statusBar.SetHints("j/k:Navigate | Enter:Select | /: Search")
```

### Handling StatusUpdateMsg

```go
case StatusUpdateMsg:
    if msg.Mode != nil {
        m.statusBar.SetMode(*msg.Mode)
    }
    if msg.HTTPStatus != nil {
        m.statusBar.SetHTTPStatus(*msg.HTTPStatus, *msg.HTTPText)
    }
    if msg.Method != nil {
        m.statusBar.SetMethod(*msg.Method)
    }
    if msg.Breadcrumb != nil {
        m.statusBar.SetBreadcrumb(msg.Breadcrumb...)
    }
    if msg.Message != nil {
        m.statusBar.ShowMessage(*msg.Message, msg.Duration)
    }
```

---

## Design Decisions

### Transparent Middle Content

The middle content area (breadcrumb/hints/messages) uses a transparent background to blend seamlessly with the terminal. This provides visual breathing room between the colored badges.

### 2-Second Message Duration

All action messages display for exactly 2 seconds, providing:

- Sufficient time to read short confirmations
- Quick return to normal status bar content
- Consistent, predictable behavior

### Unicode-Aware Truncation

Content truncation uses `lipgloss.Width()` which correctly handles:

- Multi-byte UTF-8 characters
- Wide characters (CJK)
- Emoji and special symbols

### Badge Priority

Left-side badges (Mode, Method, Fullscreen) take priority over right-side content because:

- Mode awareness is critical for vim-style navigation
- HTTP method context aids request identification
- Environment and status can be viewed in other panels if truncated

---

## Related Documentation

- [Keybindings Reference](keybindings.md) - Complete keyboard shortcut guide
- [Architecture](architecture.md) - System architecture and patterns
- [Configuration](configuration.md) - Theme and customization options
