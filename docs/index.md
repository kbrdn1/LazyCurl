# LazyCurl Documentation

> A powerful Terminal User Interface (TUI) HTTP client combining Lazygit's elegant interface with Postman's API testing capabilities.

## Quick Links

| Getting Started | Core Features | Advanced |
|-----------------|---------------|----------|
| [Installation](installation.md) | [Collections](collections.md) | [Import/Export](import-export.md) |
| [Quick Start](getting-started.md) | [Environments](environments.md) | [External Editor](external-editor.md) |
| [Keybindings](keybindings.md) | [Configuration](configuration.md) | [Session Persistence](session.md) |

## Documentation Structure

### Getting Started

- **[Installation](installation.md)** - Install LazyCurl on your system
- **[Quick Start](getting-started.md)** - Create your first request in minutes
- **[Keybindings](keybindings.md)** - Complete keyboard shortcuts reference

### User Guide

- **[Collections](collections.md)** - Organize requests in folders and collections
- **[Environments](environments.md)** - Manage variables across environments
- **[Configuration](configuration.md)** - Customize themes, keybindings, and settings

### Features

- **[Import/Export](import-export.md)** - Import from cURL, OpenAPI, Postman; Export to cURL
- **[External Editor](external-editor.md)** - Edit request bodies in vim, VS Code, etc.
- **[Session Persistence](session.md)** - Automatic state save and restore
- **[Console](console.md)** - Request history and logging
- **[CLI Reference](cli.md)** - Command-line interface and automation

### Scripting API Reference

- **[API Overview](api/overview.md)** - JavaScript scripting API introduction
- **[lc.request](api/request.md)** - HTTP request manipulation
- **[lc.response](api/response.md)** - HTTP response access
- **[lc.env & lc.globals](api/env.md)** - Environment and session variables
- **[lc.test & lc.expect](api/test.md)** - Testing and assertions
- **[Full API Reference](scripting-api-reference.md)** - Complete single-page reference

### Reference

- **[Statusbar](statusbar.md)** - Statusbar component API reference
- **[Architecture](architecture.md)** - Technical architecture and design

---

## Feature Overview

### Interface

```
┌─Collections────────┬─Request──────────────────────────┐
│                    │ POST    {{base_url}}/api/users   │
│ ▼ My API           │──────────────────────────────────│
│   ▶ Users          │ Params │ Auth │ Headers │ Body   │
│     GET  /users    │──────────────────────────────────│
│     POST /users    │ {                                │
│   ▶ Products       │   "name": "{{user_name}}"        │
│                    │ }                                │
│                    ├─Response─────────────────────────┤
│                    │ 201 Created  │ 142ms │ 1.2 KB    │
│                    │──────────────────────────────────│
│                    │ Body │ Headers │ Cookies │Console│
└────────────────────┴──────────────────────────────────┘
 NORMAL │ POST │ My API > Users │ dev │ ?:help
```

### Vim Modes

| Mode | Description | Activation |
|------|-------------|------------|
| **NORMAL** | Navigate, execute commands | `Esc` |
| **INSERT** | Edit text fields | `i`, `a`, `Enter` |
| **VIEW** | Browse responses (read-only) | Auto in Response |
| **COMMAND** | Execute commands | `:` |
| **JUMP** | Quick navigation | `f` |

### Key Features

| Feature | Description |
|---------|-------------|
| **Multi-Panel Layout** | Collections, Request, Response in one view |
| **Vim Navigation** | `h/j/k/l` navigation, modal editing |
| **Collections** | Organize requests in folders (JSON files) |
| **Environments** | Variable substitution with `{{var}}` syntax |
| **Import/Export** | cURL, OpenAPI 3.x, Postman collections |
| **External Editor** | Edit in vim, VS Code, nano, etc. |
| **Session Persistence** | Auto-save/restore application state |
| **Jump Mode** | vim-easymotion style navigation |

---

## Version

This documentation covers **LazyCurl v1.2.0**.

### Changelog

- **v1.2.0** - Postman import/export, OpenAPI security schemes
- **v1.1.0** - cURL import/export, Jump mode navigation
- **v1.0.0** - Initial stable release

---

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines.

## License

LazyCurl is [MIT licensed](../LICENSE).
