# Configuration Guide

Complete reference for configuring LazyCurl.

## Table of Contents

- [Configuration Files](#configuration-files)
- [Global Configuration](#global-configuration)
- [Workspace Configuration](#workspace-configuration)
- [Theme Configuration](#theme-configuration)
- [Keybindings Configuration](#keybindings-configuration)
- [Environment Variables](#environment-variables)

---

## Configuration Files

LazyCurl uses a two-tier configuration system:

### File Locations

| Type | Location | Purpose |
|------|----------|---------|
| **Global** | `~/.config/lazycurl/config.yaml` | User preferences, theme, keybindings |
| **Workspace** | `.lazycurl/config.yaml` | Project-specific settings |

### Priority

Workspace settings override global settings when both are defined.

---

## Global Configuration

The global configuration file (`~/.config/lazycurl/config.yaml`) contains user-wide settings.

### Full Example

```yaml
# Theme configuration
theme:
  name: "catppuccin-mocha"
  primary_color: "#b4befe"
  secondary_color: "#89b4fa"
  accent_color: "#f5c2e7"
  border_color: "#45475a"
  active_color: "#a6e3a1"

# Keybindings (vim-style)
keybindings:
  quit: ["q"]
  navigate_left: ["h"]
  navigate_right: ["l"]
  navigate_up: ["k"]
  navigate_down: ["j"]
  select: ["enter", "space"]
  back: ["esc"]
  new_request: ["n"]
  send_request: ["ctrl+s"]
  save_request: ["ctrl+w"]
  delete_request: ["d"]
  toggle_envs: ["e"]

# Default editor for external editing
editor: "vim"

# Recent workspaces list
workspaces:
  - "/home/user/projects/api-project"
  - "/home/user/projects/backend"

# Last opened workspace
last_workspace: "/home/user/projects/api-project"

# Global environments (available in all workspaces)
global_environments:
  common:
    name: "Common"
    description: "Shared variables"
    variables:
      api_version: "v1"
      timeout: "30"
```

### Configuration Options

#### Theme Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `name` | string | `"dark"` | Theme name identifier |
| `primary_color` | hex | `"#b4befe"` | Primary UI color (Lavender) |
| `secondary_color` | hex | `"#89b4fa"` | Secondary UI color (Blue) |
| `accent_color` | hex | `"#f5c2e7"` | Accent highlights (Pink) |
| `border_color` | hex | `"#45475a"` | Border color (Surface0) |
| `active_color` | hex | `"#a6e3a1"` | Active state color (Green) |

#### Editor Option

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `editor` | string | `"vim"` | External editor command |

Supported editors:

- `vim`, `nvim` - Vim/Neovim
- `nano` - GNU Nano
- `code` - VS Code
- `emacs` - Emacs
- Any executable in PATH

---

## Workspace Configuration

The workspace configuration file (`.lazycurl/config.yaml`) contains project-specific settings.

### Full Example

```yaml
# Workspace name (displayed in status bar)
name: "My API Project"

# Optional description
description: "REST API for e-commerce platform"

# Default environment to activate on startup
default_env: "development"

# Collection files to load (optional, loads all if empty)
collections:
  - "api.json"
  - "admin.json"
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `name` | string | `"My Workspace"` | Workspace display name |
| `description` | string | `""` | Optional description |
| `default_env` | string | `""` | Environment to activate on startup |
| `collections` | []string | `[]` | Specific collections to load |

### Workspace Directory Structure

```
your-project/
└── .lazycurl/
    ├── config.yaml           # Workspace configuration
    ├── collections/          # Request collections
    │   ├── api.json
    │   └── admin.json
    └── environments/         # Environment files
        ├── development.json
        ├── staging.json
        └── production.json
```

---

## Theme Configuration

LazyCurl uses the **Catppuccin Mocha** color palette by default.

### Default Colors

| Element | Color Name | Hex Code | Usage |
|---------|------------|----------|-------|
| Primary | Lavender | `#b4befe` | Selection, active elements |
| Secondary | Blue | `#89b4fa` | Selected text, links |
| Success | Green | `#a6e3a1` | Success states, GET method |
| Warning | Peach | `#fab387` | Warnings, POST method |
| Error | Red | `#f38ba8` | Errors, DELETE method |
| Text | White | `#cdd6f4` | Primary text |
| Subtext | Gray | `#a6adc8` | Secondary text |
| Surface | Dark | `#313244` | Panel backgrounds |
| Base | Darkest | `#1e1e2e` | Main background |

### HTTP Method Colors

| Method | Color | Hex Code |
|--------|-------|----------|
| GET | Green | `#a6e3a1` |
| POST | Peach | `#fab387` |
| PUT | Blue | `#89b4fa` |
| PATCH | Mauve | `#cba6f7` |
| DELETE | Red | `#f38ba8` |
| HEAD | Green | `#a6e3a1` |
| OPTIONS | Yellow | `#f9e2af` |

### Custom Theme Example

```yaml
theme:
  name: "custom-dark"
  primary_color: "#7c3aed"    # Purple
  secondary_color: "#06b6d4"  # Cyan
  accent_color: "#f59e0b"     # Amber
  border_color: "#374151"     # Gray
  active_color: "#10b981"     # Emerald
```

---

## Keybindings Configuration

All keybindings are fully customizable. Each binding accepts an array of keys.

### Default Keybindings

```yaml
keybindings:
  # Application control
  quit: ["q"]

  # Navigation
  navigate_left: ["h"]
  navigate_right: ["l"]
  navigate_up: ["k"]
  navigate_down: ["j"]

  # Selection
  select: ["enter"]
  back: ["esc"]

  # Actions
  new_request: ["n"]
  send_request: ["ctrl+s"]
  save_request: ["ctrl+w"]
  delete_request: ["d"]

  # Panel focus
  focus_collections: []
  focus_request: []
  focus_response: []

  # Toggles
  toggle_envs: ["e"]
```

### Key Format

| Format | Example | Description |
|--------|---------|-------------|
| Single key | `"a"`, `"n"`, `"1"` | Single character |
| Control combo | `"ctrl+s"`, `"ctrl+c"` | Ctrl + key |
| Shift combo | `"shift+tab"` | Shift + key |
| Special keys | `"enter"`, `"esc"`, `"tab"` | Named special keys |
| Function keys | `"f1"`, `"f12"` | Function keys |
| Arrow keys | `"up"`, `"down"`, `"left"`, `"right"` | Arrow keys |

### Multiple Keys Example

```yaml
keybindings:
  # Multiple keys for same action
  select: ["enter", "space", "l"]
  navigate_up: ["k", "up"]
  navigate_down: ["j", "down"]
  quit: ["q", "ctrl+c"]
```

### Emacs-Style Example

```yaml
keybindings:
  quit: ["ctrl+x ctrl+c"]
  navigate_left: ["ctrl+b"]
  navigate_right: ["ctrl+f"]
  navigate_up: ["ctrl+p"]
  navigate_down: ["ctrl+n"]
  select: ["enter"]
  back: ["ctrl+g"]
  new_request: ["ctrl+x n"]
  send_request: ["ctrl+c ctrl+c"]
  save_request: ["ctrl+x ctrl+s"]
```

---

## Environment Variables

### System Environment Variables

LazyCurl respects these system environment variables:

| Variable | Description |
|----------|-------------|
| `LAZYCURL_CONFIG` | Override global config path |
| `LAZYCURL_WORKSPACE` | Override workspace path |
| `EDITOR` | Fallback editor if not configured |
| `HOME` | User home directory for config location |

### Setting Environment Variables

```bash
# Linux/macOS
export LAZYCURL_CONFIG="/custom/path/config.yaml"
export EDITOR="nvim"

# Windows PowerShell
$env:LAZYCURL_CONFIG = "C:\custom\path\config.yaml"
$env:EDITOR = "notepad"
```

---

## Configuration Tips

### 1. Start with Defaults

LazyCurl works out of the box. Only customize what you need:

```yaml
# Minimal config - just change the editor
editor: "nvim"
```

### 2. Version Control Workspace Config

Add `.lazycurl/` to your git repository to share workspace settings:

```bash
# .gitignore - keep environments private
.lazycurl/environments/*.json
!.lazycurl/environments/example.json
```

### 3. Use Global Environments for Common Variables

Define shared variables in global config:

```yaml
global_environments:
  common:
    variables:
      api_version: "v1"
      user_agent: "LazyCurl/1.0"
```

### 4. Per-Project Settings

Keep project-specific settings in workspace config:

```yaml
name: "Production API"
default_env: "production"
```

---

## Troubleshooting

### Config Not Loading

1. Check file permissions: `ls -la ~/.config/lazycurl/`
2. Validate YAML syntax: `cat ~/.config/lazycurl/config.yaml | yaml`
3. Check for typos in key names

### Keybindings Not Working

1. Ensure correct format: `["key"]` not `"key"`
2. Check for conflicts with terminal shortcuts
3. Verify key names are lowercase

### Theme Colors Not Applying

1. Ensure hex format: `"#RRGGBB"` with quotes
2. Check terminal supports 256 colors
3. Verify terminal theme doesn't override

### Reset to Defaults

```bash
# Backup and remove global config
mv ~/.config/lazycurl/config.yaml ~/.config/lazycurl/config.yaml.bak

# Remove workspace config
mv .lazycurl/config.yaml .lazycurl/config.yaml.bak

# LazyCurl will recreate defaults on next run
```
