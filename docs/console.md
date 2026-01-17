# Console (Request History)

The Console tab in the Response panel displays a chronological history of all HTTP requests made during your session.

## Overview

The Console provides:

- Request/response logging for all HTTP calls
- Visual status indicators (color-coded badges)
- Quick actions (resend, copy to clipboard)
- Vim-style navigation

## Accessing the Console

| Key | Action |
|-----|--------|
| `4` | Switch to Console tab |
| `Tab` | Cycle through Response tabs |

The Console is the 4th tab in the Response panel:

```
Body │ Headers │ Cookies │ Console
                          ^^^^^^^^
```

## Console View

### List View

Shows all requests in chronological order (newest first):

```
┌─Console──────────────────────────────────────────┐
│ ● 200  GET    /api/users           142ms   1.2KB │
│ ● 201  POST   /api/users           89ms    256B  │
│ ○ ERR  GET    /api/invalid         --      --    │
│ ● 404  GET    /api/missing         45ms    128B  │
└──────────────────────────────────────────────────┘
```

**Entry Components:**

| Component | Description |
|-----------|-------------|
| Status badge | Color-coded circle (●/○) |
| Status code | HTTP status or "ERR" for errors |
| Method | HTTP method |
| Path | Request URL path |
| Duration | Response time |
| Size | Response body size |

### Status Colors

| Status | Color | Indicator |
|--------|-------|-----------|
| 2xx Success | Green | ● |
| 3xx Redirect | Blue | ● |
| 4xx Client Error | Orange | ● |
| 5xx Server Error | Red | ● |
| Network Error | Gray | ○ |

### Expanded View

Press `Enter` or `l` to expand an entry:

```
┌─Console: GET /api/users─────────────────────────┐
│ Status: 200 OK                                  │
│ Time: 142ms                                     │
│ Size: 1.2 KB                                    │
│                                                 │
│ ─── Request ───                                 │
│ GET https://api.example.com/api/users           │
│ Authorization: Bearer ***                        │
│ Content-Type: application/json                  │
│                                                 │
│ ─── Response ───                                │
│ {                                               │
│   "users": [...]                                │
│ }                                               │
│                                                 │
│ [R]esend [H]eaders [B]ody [E]rror [A]ll        │
└─────────────────────────────────────────────────┘
```

## Navigation

### List View Navigation

| Key | Action |
|-----|--------|
| `j` | Move down |
| `k` | Move up |
| `g` | Jump to first entry |
| `G` | Jump to last entry |
| `Enter` / `l` | Expand selected entry |

### Expanded View Navigation

| Key | Action |
|-----|--------|
| `Esc` / `h` / `q` | Collapse back to list |
| `j` / `k` | Scroll content |

## Actions

### Resend Request

| Context | Key | Action |
|---------|-----|--------|
| List view | `R` | Resend selected request |
| Expanded view | `R` | Resend current request |

Resending:

1. Creates a new request with same parameters
2. Sends immediately
3. New entry added to console

### Copy to Clipboard

Available in expanded view:

| Key | Copies |
|-----|--------|
| `H` | Response headers |
| `B` | Response body |
| `E` | Error message (if failed) |
| `C` | Response cookies |
| `I` | Request info (method, URL, headers) |
| `A` | All (request + response) |
| `U` | URL only (also works in list view) |

## Console Entry Details

### Successful Request

```yaml
Request:
  Method: POST
  URL: https://api.example.com/users
  Headers:
    Content-Type: application/json
    Authorization: Bearer {{token}}
  Body: '{"name": "John"}'

Response:
  Status: 201 Created
  Time: 89ms
  Size: 256 bytes
  Headers:
    Content-Type: application/json
    X-Request-Id: abc123
  Body: '{"id": 1, "name": "John"}'
```

### Failed Request

```yaml
Request:
  Method: GET
  URL: https://invalid.example.com/api
  Headers: ...

Error:
  Type: Network Error
  Message: "dial tcp: lookup invalid.example.com: no such host"
```

## Session Behavior

### Persistence

Console history is **session-only**:

- Cleared when LazyCurl exits
- Not saved to disk
- Fresh start each session

### Capacity

Console maintains recent history:

- Default: Last 100 requests
- Oldest entries removed when limit reached

## Use Cases

### Debugging

1. Send request
2. Check Console for details
3. Compare with previous attempts
4. Copy headers/body for inspection

### Replay Testing

1. Find previous successful request
2. Press `R` to resend
3. Compare results

### Documentation

1. Expand successful request
2. Press `A` to copy all
3. Paste into documentation

## Technical Details

### Entry Structure

```go
type ConsoleEntry struct {
    ID        string
    Timestamp time.Time
    Request   *http.Request
    Response  *http.Response
    Duration  time.Duration
    Error     error
    Size      int64
}
```

### Timing

Duration is measured from request send to response complete (including body read).

### Size Calculation

Response size includes body only, not headers.

---

## See Also

- [Keybindings](keybindings.md) - All keyboard shortcuts
- [Getting Started](getting-started.md) - First steps
- [Import/Export](import-export.md) - Export requests as cURL
