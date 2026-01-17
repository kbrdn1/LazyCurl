# External Editor Integration

Edit request bodies and headers in your preferred text editor with full syntax highlighting and editor features.

## Overview

LazyCurl can launch external editors (vim, VS Code, nano, etc.) to edit request content. This is useful for:

- Editing large JSON payloads
- Using editor-specific features (snippets, formatting)
- Familiar editing environment

## Quick Start

1. Set your preferred editor:

   ```bash
   export VISUAL="vim"
   # or
   export EDITOR="nano"
   ```

2. In INSERT mode, press `Ctrl+E`
3. Edit content in your editor
4. Save and exit
5. Content is updated in LazyCurl

## Configuration

### Environment Variables

LazyCurl checks these variables in order:

| Variable | Priority | Description |
|----------|----------|-------------|
| `$VISUAL` | 1st | Preferred editor (supports GUI) |
| `$EDITOR` | 2nd | Fallback editor |
| Built-in | 3rd | `nano` → `vi` (first found) |

### Editor Examples

**Terminal Editors:**

```bash
export VISUAL="vim"
export VISUAL="nvim"
export VISUAL="nano"
export VISUAL="emacs"
export VISUAL="micro"
```

**GUI Editors:**

```bash
# VS Code (--wait is required)
export VISUAL="code --wait"

# Sublime Text
export VISUAL="subl --wait"

# Atom (deprecated)
export VISUAL="atom --wait"

# gedit
export VISUAL="gedit"
```

> **Important:** GUI editors must use `--wait` flag to block until file is closed.

## Usage

### Keybinding

| Context | Key | Action |
|---------|-----|--------|
| INSERT mode (Body) | `Ctrl+E` | Open body in external editor |
| INSERT mode (Headers) | `Ctrl+E` | Open headers in external editor |

### Workflow

```
┌─────────────────────────────────────────────────────┐
│  1. Press Ctrl+E in INSERT mode                     │
│                    ↓                                │
│  2. LazyCurl creates temp file with content         │
│                    ↓                                │
│  3. Editor opens with appropriate extension         │
│     (.json, .xml, .txt based on content)            │
│                    ↓                                │
│  4. Edit content, save, and exit editor             │
│                    ↓                                │
│  5. LazyCurl reads updated content                  │
│                    ↓                                │
│  6. Temp file is cleaned up                         │
└─────────────────────────────────────────────────────┘
```

## Content Type Detection

LazyCurl automatically detects content type and uses appropriate file extension:

| Content Pattern | Extension | Example |
|-----------------|-----------|---------|
| Starts with `{` or `[` | `.json` | JSON objects/arrays |
| Starts with `<?xml` | `.xml` | XML documents |
| Starts with `<!doctype` or `<html>` | `.html` | HTML documents |
| Other | `.txt` | Plain text |

This enables:

- Syntax highlighting in editors
- Format-specific plugins/extensions
- Auto-formatting on save

## Headers Editing

When editing headers, they are serialized as text:

```
Content-Type: application/json
Authorization: Bearer {{token}}
X-Custom-Header: value
```

Edit as plain text, one header per line in `Name: Value` format.

## Error Handling

### Common Errors

| Error | Cause | Solution |
|-------|-------|----------|
| "No editor configured" | `$VISUAL` and `$EDITOR` not set | Set environment variable |
| "Editor not found" | Editor binary not in PATH | Install editor or fix PATH |
| "Editor exited with error" | Editor crashed or error | Check editor logs |

### Error Messages

LazyCurl shows clear error messages in the statusbar:

- `Editor not found: vim` - Editor binary not installed
- `No editor configured` - Set `$VISUAL` or `$EDITOR`
- `Failed to create temp file` - Disk/permission issue

## Advanced Usage

### Using Different Editors per Content Type

While not directly supported, you can use wrapper scripts:

```bash
#!/bin/bash
# ~/bin/smart-editor
case "$1" in
  *.json) code --wait "$1" ;;
  *.xml)  vim "$1" ;;
  *)      nano "$1" ;;
esac
```

```bash
export VISUAL="~/bin/smart-editor"
```

### SSH Remote Editing

For remote development, ensure your editor supports remote files:

```bash
# VS Code Remote
export VISUAL="code --wait --remote ssh-remote+myserver"
```

## Troubleshooting

### Editor Opens But Content Not Updated

**Cause:** Editor exited before file was saved.

**Solution:** Ensure you save (`:w` in vim, `Ctrl+S` in others) before exiting.

### GUI Editor Opens New Window

**Cause:** Missing `--wait` flag.

**Solution:** Add `--wait` to your editor command:

```bash
export VISUAL="code --wait"
```

### Content Lost After Edit

**Cause:** Editor crashed or force-quit.

**Solution:** Temp files are in system temp directory. Check for recovery.

### Wrong Syntax Highlighting

**Cause:** Content type detection failed.

**Solution:** Content type is detected from content start. Ensure valid JSON/XML prefix.

## Technical Details

### Temp File Location

- macOS: `/var/folders/.../lazycurl-*.ext`
- Linux: `/tmp/lazycurl-*.ext`
- Windows: `%TEMP%\lazycurl-*.ext`

### File Lifecycle

1. **Create:** Content written to temp file
2. **Edit:** Editor process launched (TUI suspended)
3. **Read:** Content read back after editor exits
4. **Cleanup:** Temp file deleted

### Process Handling

- TUI is suspended during editing (`tea.ExecProcess`)
- Editor runs in foreground with full terminal control
- Exit code 0 indicates success
- Non-zero exit code shows error message

---

## See Also

- [Keybindings](keybindings.md) - All keyboard shortcuts
- [Configuration](configuration.md) - General configuration
- [Getting Started](getting-started.md) - First steps with LazyCurl
