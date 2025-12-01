# Keybindings Reference

Complete keyboard shortcut reference for LazyCurl.

## Table of Contents

- [Vim-Style Modes](#vim-style-modes)
- [Global Keybindings](#global-keybindings)
- [Navigation](#navigation)
- [Collections Panel](#collections-panel)
- [Environments Panel](#environments-panel)
- [Request Panel](#request-panel)
- [Response Panel](#response-panel)
- [Search Mode](#search-mode)
- [Command Mode](#command-mode)
- [Dialogs & Modals](#dialogs--modals)
- [WhichKey](#whichkey)

---

## Vim-Style Modes

LazyCurl uses vim-style modes for different interaction contexts:

| Mode | Status Bar | Color | Description |
|------|------------|-------|-------------|
| **NORMAL** | `NORMAL` | Blue | Default mode for navigation and commands |
| **INSERT** | `INSERT` | Gray | Text input mode for editing fields |
| **VIEW** | `VIEW` | Green | Read-only browsing mode |
| **COMMAND** | `COMMAND` | Orange | Command line mode (`:` prefix) |

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
                       │           │             │
                       └─────┬─────┘             │
                             │ Esc               │ Esc/Enter
                             ▼                   ▼
                       ┌─────────────┐     ┌─────────────┐
                       │   NORMAL    │     │   NORMAL    │
                       └─────────────┘     └─────────────┘
```

---

## Global Keybindings

These work in most contexts:

| Key | Action | Mode |
|-----|--------|------|
| `Esc` | Return to NORMAL mode | Any |
| `q` | Quit application | NORMAL |
| `Ctrl+C` | Force quit | Any |
| `:` | Enter COMMAND mode | NORMAL |
| `?` | Show WhichKey (keybinding hints) | NORMAL |
| `Ctrl+S` | Send HTTP request | NORMAL |

---

## Navigation

### Panel Navigation

| Key | Action |
|-----|--------|
| `h` | Navigate to left panel |
| `l` | Navigate to right panel |

Panel order: **Collections** ← → **Request** ← → **Response**

### Tab Navigation (Left Panel)

| Key | Action |
|-----|--------|
| `1` | Switch to Collections tab |
| `2` | Switch to Environments tab |

### Tab Navigation (Request/Response Panels)

| Key | Action |
|-----|--------|
| `Tab` | Next tab |
| `Shift+Tab` | Previous tab |
| `1-6` | Jump to specific tab (Request: Params/Auth/Headers/Body/Scripts/Settings) |
| `1-3` | Jump to specific tab (Response: Body/Headers/Cookies) |

### List Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` | Jump to top |
| `G` | Jump to bottom |

### Tree Navigation (Collections/Environments)

| Key | Action |
|-----|--------|
| `h` | Collapse folder / Go to parent |
| `l` | Expand folder / Enter folder |
| `Enter` / `Space` | Select item / Open request |

---

## Collections Panel

### Actions

| Key | Action |
|-----|--------|
| `n` | Create new request |
| `N` | Create new folder |
| `c` / `i` | Edit selected request |
| `R` | Rename item |
| `d` | Delete item |
| `D` | Duplicate item |

### Clipboard Operations

| Key | Action |
|-----|--------|
| `y` | Yank (copy) to clipboard |
| `p` | Paste from clipboard |

### Search

| Key | Action |
|-----|--------|
| `/` | Open search |

---

## Environments Panel

### Navigation

| Key | Action |
|-----|--------|
| `j` / `k` | Move up/down |
| `h` | Collapse environment |
| `l` | Expand environment |
| `g` / `G` | Jump to top/bottom |

### Environment Actions

| Key | Action |
|-----|--------|
| `N` | Create new environment |
| `S` / `Enter` | Select/activate environment |
| `d` | Delete environment |
| `D` | Duplicate environment |
| `R` | Rename environment |

### Variable Actions

| Key | Action |
|-----|--------|
| `n` | Create new variable |
| `c` / `i` | Edit variable value |
| `d` | Delete variable |
| `D` | Duplicate variable |
| `R` | Rename variable |

### Toggle States

| Key | Action |
|-----|--------|
| `a` / `A` | Toggle variable active/inactive |
| `s` | Toggle variable secret/visible |

### Search

| Key | Action |
|-----|--------|
| `/` | Open search |

---

## Request Panel

### Tab Shortcuts

| Key | Tab |
|-----|-----|
| `1` | Params |
| `2` | Authorization |
| `3` | Headers |
| `4` | Body |
| `5` | Scripts |
| `6` | Settings |

### Actions

| Key | Action |
|-----|--------|
| `Tab` | Next tab |
| `Shift+Tab` | Previous tab |
| `i` | Enter INSERT mode (edit fields) |
| `Ctrl+S` | Send request |

### In INSERT Mode

| Key | Action |
|-----|--------|
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `Enter` | Confirm input (in single-line fields) |
| `Esc` | Exit INSERT mode |

---

## Response Panel

### Tab Shortcuts

| Key | Tab |
|-----|-----|
| `1` | Body |
| `2` | Headers |
| `3` | Cookies |

### Navigation

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll down/up |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `v` | Enter VIEW mode (focused reading) |

### VIEW Mode

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll |
| `g` / `G` | Top/Bottom |
| `h` / `l` | Switch panels |
| `Esc` | Exit VIEW mode |

---

## Search Mode

When search is active (after pressing `/`):

### While Typing (INSERT)

| Key | Action |
|-----|--------|
| *Type* | Enter search query |
| `Enter` | Confirm search |
| `Esc` | Cancel search |

### After Confirming (NORMAL with active search)

| Key | Action |
|-----|--------|
| `n` | Jump to next match |
| `N` | Jump to previous match |
| `j` / `k` | Move up/down (all items) |
| `i` | Edit search query |
| `Enter` / `Space` | Open selected item |
| `Esc` | Clear search filter |

### Visual Feedback

- **Matching items**: Normal color with highlighted match
- **Non-matching items**: Dimmed (gray)
- **Current selection**: Highlighted with primary color

---

## Command Mode

Enter with `:` from NORMAL mode.

### Available Commands

| Command | Aliases | Action |
|---------|---------|--------|
| `:q` | `:quit` | Quit application |
| `:w` | `:write`, `:save` | Save current request |
| `:wq` | | Save and quit |
| `:help` | `:h` | Show help |
| `:e` | `:env` | Switch to environments |
| `:col` | `:collections` | Switch to collections |

### Workspace Commands

| Command | Action |
|---------|--------|
| `:ws list` | List recent workspaces |
| `:ws switch <name>` | Switch to workspace |
| `:ws create <name>` | Create new workspace |
| `:ws delete <name>` | Delete workspace |

### Navigation

| Key | Action |
|-----|--------|
| `Enter` | Execute command |
| `Esc` | Cancel and return to NORMAL |
| `Backspace` | Delete character |
| `Ctrl+U` | Clear entire line |

---

## Dialogs & Modals

### Input Dialogs (New Request, Rename, etc.)

| Key | Action |
|-----|--------|
| `Tab` / `↓` | Next field |
| `Shift+Tab` / `↑` | Previous field |
| `Enter` | Confirm dialog |
| `Esc` | Cancel dialog |

### Confirmation Dialogs

| Key | Action |
|-----|--------|
| `Enter` / `y` | Confirm action |
| `Esc` / `n` | Cancel action |
| `Tab` | Switch between Yes/No buttons |

### Modal Navigation

| Key | Action |
|-----|--------|
| `Tab` / `j` | Next option |
| `Shift+Tab` / `k` | Previous option |
| `Enter` | Select option |
| `Esc` | Close modal |

---

## WhichKey

Press `?` to show context-aware keybinding hints.

### WhichKey Modal

| Key | Action |
|-----|--------|
| `?` | Toggle WhichKey |
| `Esc` | Close WhichKey |
| `q` | Close WhichKey |

### Context Indicators

WhichKey shows different hints based on your current context:
- `normal_collections` - Collections panel in NORMAL mode
- `normal_env` - Environments panel in NORMAL mode
- `normal_request` - Request panel in NORMAL mode
- `normal_response` - Response panel in NORMAL mode
- `search_collections` - Search active in Collections
- `search_env` - Search active in Environments
- `insert` - INSERT mode (any panel)
- `view` - VIEW mode
- `command` - COMMAND mode
- `dialog` - Dialog is open
- `modal` - Modal is open

---

## Quick Reference

### Most Used Shortcuts

| Action | Keys |
|--------|------|
| Navigate between panels | `h` / `l` |
| Move in list | `j` / `k` |
| Create new item | `n` |
| Edit item | `c` or `i` |
| Delete item | `d` |
| Duplicate | `D` |
| Search | `/` |
| Send request | `Ctrl+S` |
| Show help | `?` |
| Quit | `q` |

### Vim Users Cheat Sheet

| Vim Command | LazyCurl Equivalent |
|-------------|---------------------|
| `:q` | `:q` - Quit |
| `:w` | `:w` - Save |
| `/pattern` | `/` - Search |
| `n` / `N` | `n` / `N` - Next/Previous match |
| `dd` | `d` - Delete |
| `yy` | `y` - Yank |
| `p` | `p` - Paste |
| `i` | `i` - Insert mode |
| `Esc` | `Esc` - Normal mode |
