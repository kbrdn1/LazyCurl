# Session Persistence

LazyCurl automatically saves and restores your application state, so you can pick up exactly where you left off.

## Overview

Session persistence tracks:

- Active panel and selection
- Expanded folders in Collections
- Active tabs in Request/Response panels
- Scroll positions
- Active environment

## How It Works

### Automatic Saving

State is saved automatically:

- **On change:** When you navigate, expand folders, or switch tabs
- **Debounced:** 500ms delay prevents excessive disk writes
- **On quit:** Final save when exiting LazyCurl

### Automatic Restore

When launching LazyCurl:

1. Session file is loaded (if exists)
2. References are validated (collection exists, etc.)
3. State is applied to panels
4. Invalid references are silently ignored

## Session File

### Location

```
.lazycurl/session.yml
```

### Format

```yaml
version: 1
last_updated: "2025-01-15T10:30:00Z"
active_panel: request
active_collection: "api.json"
active_request: "req_001"
active_environment: "development"
panels:
  collections:
    expanded_folders:
      - "Users"
      - "Products"
    scroll_position: 5
    selected_index: 3
  request:
    active_tab: "body"
  response:
    active_tab: "headers"
    scroll_position: 0
```

### Fields Reference

| Field | Type | Description |
|-------|------|-------------|
| `version` | int | Session format version |
| `last_updated` | timestamp | Last save time |
| `active_panel` | string | Currently focused panel |
| `active_collection` | string | Selected collection filename |
| `active_request` | string | Selected request ID |
| `active_environment` | string | Active environment name |
| `panels.collections.expanded_folders` | array | List of expanded folder names |
| `panels.collections.scroll_position` | int | Scroll offset in list |
| `panels.collections.selected_index` | int | Cursor position |
| `panels.request.active_tab` | string | Active tab (params, auth, headers, body) |
| `panels.response.active_tab` | string | Active tab (body, headers, cookies, console) |
| `panels.response.scroll_position` | int | Scroll offset in response |

## What's Persisted

### Persisted State

| State | Description |
|-------|-------------|
| Active panel | Collections, Request, or Response |
| Collection selection | Which collection is loaded |
| Request selection | Which request is selected |
| Environment | Active environment |
| Folder expansion | Which folders are expanded/collapsed |
| Tab selection | Active tab in Request and Response |
| Scroll position | Vertical scroll in lists and content |
| Cursor position | Selected item index |

### Not Persisted

| State | Reason |
|-------|--------|
| Request content edits | Use `Ctrl+S` to save |
| Response data | Ephemeral, re-send request |
| Console history | Session-only, cleared on restart |
| Search queries | Intentionally transient |

## Graceful Degradation

Session persistence is designed to be resilient:

### Missing Session File

- Fresh start with default state
- No error shown

### Invalid References

- Collection not found → Reset to first collection
- Request not found → Reset to first request
- Environment not found → No active environment

### Corrupted Session

- Parse error → Fresh start
- Invalid version → Fresh start
- Missing fields → Use defaults

## Configuration

Session persistence is always enabled. To reset session:

```bash
# Delete session file
rm .lazycurl/session.yml

# Start fresh
lazycurl
```

## Technical Details

### Atomic Writes

Session saves use atomic write pattern:

1. Write to temp file: `session.yml.tmp`
2. Rename temp to target: `session.yml`

This prevents corruption if process is interrupted.

### Debouncing

Changes are debounced with 500ms delay:

```
Change → Wait 500ms → Save
         ↑
Change → Reset timer
```

Multiple rapid changes result in single write.

### File Permissions

Session file is created with `0644` permissions (owner read/write, others read).

## Troubleshooting

### Session Not Restoring

**Check:** Session file exists and is readable

```bash
cat .lazycurl/session.yml
```

**Verify:** Collection/environment references are valid

### State Not Saving

**Check:** Write permissions on `.lazycurl/` directory

```bash
ls -la .lazycurl/
```

**Verify:** Disk space available

### Wrong Collection/Request on Startup

**Cause:** Referenced items were deleted or renamed

**Solution:** Session will fall back to defaults. Edit session file manually if needed.

## Best Practices

### Version Control

Add session file to `.gitignore`:

```gitignore
# .gitignore
.lazycurl/session.yml
```

This prevents personal state from being committed.

### Multiple Workspaces

Each workspace has independent session:

```
project-a/.lazycurl/session.yml
project-b/.lazycurl/session.yml
```

Switch directories to switch contexts.

### Shared Workspaces

For team environments:

1. Keep `session.yml` in `.gitignore`
2. Commit collections and environments
3. Each team member has their own session

---

## See Also

- [Configuration](configuration.md) - Workspace configuration
- [Collections](collections.md) - Collection management
- [Environments](environments.md) - Environment management
