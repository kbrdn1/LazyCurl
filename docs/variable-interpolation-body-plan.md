# Variable Interpolation in Request Body - Implementation Plan

## Objective

Add `{{variable}}` interpretation and highlighting in the request body editor, with support for:

1. Visual highlighting of variables in the editor
2. Preview mode showing resolved values
3. Tooltip/inline display of resolved values

## Current State Analysis

### Existing Components

1. **Variable System** (`internal/api/variables.go`):
   - `ReplaceVariables(text, env)` - Replaces `{{var}}` with values
   - `FindVariables(text)` - Extracts all variable names from text
   - `PreviewVariableReplacement(text, env)` - Shows resolved text
   - System variables: `$timestamp`, `$uuid`, `$datetime`, etc.

2. **Editor Component** (`internal/ui/components/editor.go`):
   - `highlightJSON()` - JSON syntax highlighting
   - `highlightJS()` - JavaScript highlighting
   - Support for search match highlighting
   - Line-by-line rendering with styles

3. **Request View** (`internal/ui/request_view.go`):
   - `renderTextWithVariables()` - Already highlights `{{var}}` in Auth tab
   - `renderURLWithHighlight()` - Highlights variables in URL
   - `GetBodyContent()` - Returns raw body content

4. **Environments View** (`internal/ui/environments_view.go`):
   - `GetActiveEnvironment()` - Returns current environment
   - `GetActiveEnvironmentVariables()` - Returns variables map

## Implementation Plan

### Phase 1: Variable Highlighting in Editor

**File: `internal/ui/components/editor.go`**

1. Add variable pattern constant:

```go
var variablePattern = regexp.MustCompile(`\{\{([^}]+)\}\}`)
```

2. Modify `highlightJSON()` to detect and style `{{variables}}`:
   - Style variables with `styles.URLVariable` color
   - Maintain JSON syntax highlighting for non-variable parts

3. Add new method `highlightVariables(line string) string`:
   - Find all `{{var}}` patterns in line
   - Apply variable style to each match
   - Return highlighted line

### Phase 2: Preview Mode (Ctrl+P toggle)

**File: `internal/ui/components/editor.go`**

1. Add preview state to Editor:

```go
type Editor struct {
    // ... existing fields
    previewMode     bool                   // Toggle preview mode
    variableValues  map[string]string      // Current variable values
}
```

2. Add methods:
   - `SetVariableValues(vars map[string]string)` - Set values for preview
   - `TogglePreviewMode()` - Toggle preview on/off
   - `IsPreviewMode() bool` - Check current state

3. Modify rendering:
   - In preview mode, replace `{{var}}` with resolved values
   - Show resolved values in a different style (e.g., green background)

4. Add keybinding:
   - `P` in NORMAL mode to toggle preview

### Phase 3: Integration with Request Panel

**File: `internal/ui/request_view.go`**

1. Pass environment variables to body editor:
   - Add method `SetEnvironment(env *api.EnvironmentFile)`
   - Update variable values when environment changes

2. Update help text in Body tab to show preview shortcut

### Phase 4: Inline Value Display (Optional Enhancement)

Show resolved value as a "ghost" text after the variable:

```
"token": "{{auth_token}}" → eyJhbGciOiJIUzI1NiI...
```

## File Changes Summary

| File | Changes |
|------|---------|
| `internal/ui/components/editor.go` | Add variable highlighting, preview mode |
| `internal/ui/request_view.go` | Pass environment to editor |
| `internal/ui/model.go` | Update environment when it changes |
| `pkg/styles/styles.go` | Add preview-related styles (if needed) |

## Keybindings

| Key | Mode | Action |
|-----|------|--------|
| `P` | NORMAL | Toggle preview mode (show resolved values) |

## Visual Design

### Variable Highlighting

```
{
  "url": "{{base_url}}/api/users",  <- {{base_url}} in orange
  "token": "{{auth_token}}"          <- {{auth_token}} in orange
}
```

### Preview Mode (P toggle)

```
{
  "url": "https://api.example.com/api/users",  <- resolved, green bg
  "token": "Bearer eyJhbG..."                   <- resolved, green bg
}
```

## Testing Checklist

- [ ] Variables highlighted in JSON body
- [ ] Variables highlighted in raw body
- [ ] Preview mode shows resolved values
- [ ] Unresolved variables remain as `{{var}}`
- [ ] System variables (`$timestamp`, `$uuid`) resolved correctly
- [ ] Preview mode toggle works with P key
- [ ] Mode indicator shows "PREVIEW" when active
