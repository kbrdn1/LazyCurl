# Request Panel Implementation - Technical Documentation

## Overview

This document details the implementation of the Request Panel for LazyCurl, transforming the basic request builder into a fully-featured HTTP request editor with vim-like navigation, variable support, and persistent state management.

## Architecture

### Component Structure

```
RequestView (internal/ui/request_view.go)
├── Method Selector (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS)
├── URL Input (with variable highlighting and path params extraction)
├── Tabs Component
│   ├── Params Tab (Query + Path Parameters)
│   ├── Authorization Tab (Bearer, Basic, API Key)
│   ├── Headers Tab (Key-Value pairs)
│   ├── Body Tab (JSON, XML, Raw, Form, None)
│   └── Scripts Tab (Pre-request, Post-request)
└── Supporting Components
    ├── Editor (Vim-like text editor)
    ├── Table (Key-value management)
    └── WhichKey (Keybinding hints)
```

### Message Flow

```
User Input → RequestView.Update() → RequestXxxMsg → Model.Update() → CollectionsView.UpdateRequestXxx() → Save to disk
```

## Features Implemented

### 1. URL Editor with Variable Support

**File**: `internal/ui/request_view.go`

- Inline URL editing with cursor navigation
- Syntax highlighting for `{{variables}}`
- Automatic path parameter extraction (`:id`, `:slug`)
- Query parameter synchronization with URL

**Key Functions**:
- `renderURLEditor()` - Renders URL with cursor and variable highlighting
- `parsePathParams()` - Extracts `:param` patterns from URL
- `parseQueryParams()` - Parses `?key=value&...` into table
- `buildURLFromParams()` - Rebuilds URL from params table

### 2. Params Tab (Query + Path Parameters)

**Features**:
- Dual-section layout: Path Parameters / Query Parameters
- h/l navigation between sections
- Automatic sync with URL bar
- Toggle enable/disable with space
- CRUD operations (n/c/d/D/y/p)

**Data Structure**:
```go
type ParamsSection int
const (
    PathParamsSection ParamsSection = iota
    QueryParamsSection
)
```

### 3. Authorization Tab

**Supported Auth Types**:
- **No Auth**: No authentication headers
- **Bearer Token**: `Authorization: Bearer <token>`
- **Basic Auth**: `Authorization: Basic base64(user:pass)`
- **API Key**: Header or Query parameter

**File**: `internal/ui/request_view.go`

**Key Functions**:
- `renderAuthTab()` - Renders auth type selector and fields
- `handleAuthTabInput()` - Handles auth-specific keybindings
- `GetAuthHeaders()` - Generates auth headers for HTTP request

**Navigation**:
- h/l: Change auth type
- j/k: Navigate between fields
- i/c/Enter: Edit selected field
- Esc: Exit edit mode

### 4. Headers Tab

**Features**:
- Default headers (Content-Type, Accept, User-Agent)
- Custom header management
- Enable/disable toggle
- CRUD operations

**Default Headers**:
```go
[]api.KeyValueEntry{
    {Key: "Content-Type", Value: "application/json", Enabled: true},
    {Key: "Accept", Value: "*/*", Enabled: true},
    {Key: "User-Agent", Value: "LazyCurl/1.0", Enabled: true},
}
```

### 5. Body Tab with Vim Editor

**Supported Body Types**:
- JSON (with syntax highlighting)
- XML
- Raw Text
- Form Data (urlencoded)
- None

**Editor Features** (`internal/ui/components/editor.go`):
- NORMAL/INSERT modes
- Vim motions: h/j/k/l, w/b, 0/$, gg/G
- Line operations: dd, yy, p
- Insert modes: i, I, a, A, o, O
- Auto-indentation
- Ctrl+F: Format JSON

### 6. Scripts Tab

**Sections**:
- Pre-request: Executed before sending
- Post-request: Executed after response

**Features**:
- Dual editor layout
- Section switching with h/l
- Full vim-like editing

### 7. WhichKey Integration

**File**: `internal/ui/components/whichkey.go`

**Contexts Added**:
- `ContextRequestParams`
- `ContextRequestAuth`
- `ContextRequestHeaders`
- `ContextRequestBody`
- `ContextRequestScripts`

**Status Bar Hints**:
Dynamic hints based on active tab showing available keybindings.

## API Layer Changes

### Collection Operations

**File**: `internal/api/collection.go`

**New Methods**:
```go
// Update request URL by ID
func (c *CollectionFile) UpdateRequestURL(requestID, newURL string) bool

// Update request body
func (c *CollectionFile) UpdateRequestBody(requestID, bodyType, content string) bool

// Update request scripts
func (c *CollectionFile) UpdateRequestScripts(requestID, preRequest, postRequest string) bool

// Update request auth
func (c *CollectionFile) UpdateRequestAuth(requestID string, auth *AuthConfig) bool
```

**Data Structures Added**:
```go
type AuthConfig struct {
    Type           string `json:"type"`
    Token          string `json:"token,omitempty"`
    Prefix         string `json:"prefix,omitempty"`
    Username       string `json:"username,omitempty"`
    Password       string `json:"password,omitempty"`
    APIKeyName     string `json:"apiKeyName,omitempty"`
    APIKeyValue    string `json:"apiKeyValue,omitempty"`
    APIKeyLocation string `json:"apiKeyLocation,omitempty"`
}

type CollectionRequest struct {
    // ... existing fields
    Auth        *AuthConfig `json:"auth,omitempty"`
    PreRequest  string      `json:"preRequest,omitempty"`
    PostRequest string      `json:"postRequest,omitempty"`
}
```

## UI Improvements

### Single-Line Display (No Ellipsis)

**Problem**: Tree nodes and table rows were wrapping to multiple lines.

**Solution**: Truncate content to fit width without ellipsis.

**Files Modified**:
- `internal/ui/components/tree.go` - Node name truncation
- `internal/ui/environments_view.go` - Variable key/value truncation
- `internal/ui/request_view.go` - Table row truncation

### Height Calculation Fixes

**Problem**: Empty space at bottom of Body and Scripts editors.

**Solution**: Corrected height deductions in view rendering chain.

**Changes**:
- `View()`: `height - 5` → `height - 4`
- `renderScriptsTab()`: `height - 3` → `height - 2`
- `renderParamsTab()`: `height - 3` → `height - 2`

### Editor Focus Bug Fix

**Problem**: Pressing `q` in editor NORMAL mode quit the application.

**Solution**: Added `IsEditorActive()` method to properly forward all keys to editor when Body/Scripts tab is active.

```go
func (r *RequestView) IsEditorActive() bool {
    tab := r.tabs.GetActive()
    return tab == "Body" || tab == "Scripts"
}
```

### Request Selection Focus

**Problem**: Selecting a request in Collections tree didn't focus Request Panel.

**Solution**: Added panel focus in `TreeSelectionMsg` handler.

```go
case components.TreeSelectionMsg:
    if msg.Node != nil && msg.Node.Type == components.RequestNode {
        m.requestPanel.LoadRequest(...)
        m.activePanel = RequestPanel  // Added
        // ...
    }
```

## Keybindings Summary

### Global (Request Panel)
| Key | Action |
|-----|--------|
| Tab | Next tab |
| 1-6 | Direct tab switch |
| Ctrl+S | Send request |

### Params/Headers Tab
| Key | Action |
|-----|--------|
| j/k | Navigate rows |
| h/l | Switch section (Params) |
| n | New entry |
| c/i | Edit selected |
| d | Delete |
| D | Duplicate |
| y | Yank |
| p | Paste |
| Space | Toggle enabled |

### Authorization Tab
| Key | Action |
|-----|--------|
| h/l | Change auth type |
| j/k | Navigate fields |
| i/c/Enter | Edit field |
| Esc | Exit edit |

### Body/Scripts Tab
| Key | Action |
|-----|--------|
| i | Enter INSERT mode |
| Esc | Exit to NORMAL mode |
| h/l | Switch body type / section |
| Ctrl+F | Format JSON |
| (Vim motions) | Standard vim navigation |

## Testing

**File**: `internal/api/collection_test.go`

**Tests Added**:
- `TestCollectionFile_UpdateRequestURL`
- `TestCollectionFile_UpdateRequestBody`
- `TestCollectionFile_UpdateRequestScripts`
- `TestCollectionFile_UpdateRequestAuth`

## Files Modified

| File | Lines Changed | Description |
|------|---------------|-------------|
| `internal/ui/request_view.go` | +2187 | Complete Request Panel implementation |
| `internal/api/collection.go` | +301 | Request update methods, AuthConfig |
| `internal/ui/model.go` | +276 | Request message handlers, editor focus |
| `internal/ui/components/editor.go` | +610 | Vim-like editor enhancements |
| `internal/ui/components/whichkey.go` | +74 | Request panel contexts |
| `internal/ui/components/tabs.go` | +107 | Tab navigation improvements |
| `internal/ui/components/tree.go` | +20 | Single-line truncation |
| `internal/ui/environments_view.go` | +22 | Single-line truncation |
| `internal/ui/collections_view.go` | +92 | Request update methods |
| `internal/api/collection_test.go` | +165 | Unit tests |
| `internal/ui/components/table.go` | +47 | Table enhancements |
| `internal/ui/components/dialog.go` | +143 | Dialog improvements |
| `pkg/styles/styles.go` | +5 | New style colors |

## Future Considerations

1. **HTTP Request Execution**: Wire up Ctrl+S to actually send requests
2. **Response Panel**: Display formatted responses
3. **Variable Resolution**: Replace `{{var}}` with environment values
4. **Request History**: Track sent requests
5. **cURL Import/Export**: Convert to/from cURL commands
