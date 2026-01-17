<h1 align="center">
  <img src="LazyCurl.png" alt="LazyCurl Logo" width="128" height="128" />
  <br />
  LazyCurl
</h1>

<p align="center">
  <strong>A powerful Terminal User Interface (TUI) HTTP client</strong><br />
  Combining <strong>Lazygit</strong>'s elegant interface with <strong>Postman</strong>'s API testing capabilities
</p>

<p align="center">
  <a href="#-features">Features</a> •
  <a href="#-installation">Installation</a> •
  <a href="#-quick-start">Quick Start</a> •
  <a href="#%EF%B8%8F-keyboard-shortcuts">Shortcuts</a> •
  <a href="#-documentation">Documentation</a> •
  <a href="#-contributing">Contributing</a>
</p>

<p align="center">
  <a href="https://github.com/kbrdn1/LazyCurl/actions/workflows/ci.yml"><img src="https://github.com/kbrdn1/LazyCurl/actions/workflows/ci.yml/badge.svg" alt="CI" /></a>
  <a href="https://github.com/kbrdn1/LazyCurl/releases"><img src="https://img.shields.io/github/v/release/kbrdn1/LazyCurl?style=flat-square" alt="Release" /></a>
  <a href="https://codecov.io/gh/kbrdn1/LazyCurl"><img src="https://codecov.io/gh/kbrdn1/LazyCurl/branch/main/graph/badge.svg" alt="Coverage" /></a>
  <a href="https://goreportcard.com/report/github.com/kbrdn1/LazyCurl"><img src="https://goreportcard.com/badge/github.com/kbrdn1/LazyCurl" alt="Go Report Card" /></a>
  <a href="https://coderabbit.ai"><img src="https://img.shields.io/coderabbit/prs/github/kbrdn1/LazyCurl?utm_source=oss&utm_medium=github&utm_campaign=kbrdn1%2FLazyCurl&labelColor=171717&color=FF570A&link=https%3A%2F%2Fcoderabbit.ai&label=CodeRabbit+Reviews" alt="CodeRabbit Reviews" /></a>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version" />
  <img src="https://img.shields.io/badge/TUI-Bubble_Tea-b4befe?style=for-the-badge" alt="Bubble Tea" />
  <img src="https://img.shields.io/badge/Theme-Catppuccin_Mocha-f5c2e7?style=for-the-badge" alt="Catppuccin" />
  <img src="https://img.shields.io/badge/License-MIT-a6e3a1?style=for-the-badge" alt="License" />
</p>

---

## 🎯 Philosophy

LazyCurl is designed for developers who live in the terminal. It combines:

- **Elegant Interface** — Lazygit's proven multi-panel layout with vim motions
- **Powerful Features** — Postman's comprehensive HTTP testing capabilities
- **File-Based Storage** — Simple, versionable JSON/YAML files you can commit to git
- **Keyboard-First** — Full vim-style navigation, no mouse required

---

## ✨ Features

### 🎨 Lazygit-Style Interface

```
┌─Collections────────┬─Request──────────────────────────┐
│                    │ POST    {{base_url}}/api/users   │
│ ▼ My API           │──────────────────────────────────│
│   ▶ Users          │ Params │ Auth │ Headers │ Body   │
│     GET  /users    │──────────────────────────────────│
│     POST /users    │ {                                │
│     GET  /users/:id│   "name": "John Doe",            │
│   ▶ Products       │   "email": "{{user_email}}"      │
│                    │ }                                │
│                    ├─Response────────────────────────-┤
│                    │ 201 Created  │ 142ms │ 1.2 KB    │
│                    │──────────────────────────────────│
│                    │ {                                │
│                    │   "id": 123,                     │
│                    │   "name": "John Doe"             │
│                    │ }                                │
└────────────────────┴──────────────────────────────────┘
 NORMAL │ POST │ My API > Users > Create User │ dev │ ?:help
```

### 🚀 Core Features

| Feature | Description |
|---------|-------------|
| **Multi-Panel Layout** | Collections, Request Builder, Response Viewer in one view |
| **Vim Motions** | Navigate with `h/j/k/l`, modes (NORMAL, INSERT, VIEW, COMMAND) |
| **HTTP Methods** | GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS |
| **Collections** | Organize requests in folders, stored as JSON files |
| **Environments** | Multiple environments with variable substitution `{{var}}` |
| **Variable System** | Environment variables + built-in system variables |
| **Search** | Filter collections and environments with `/` |
| **WhichKey** | Press `?` for context-aware keybinding hints |
| **Command Mode** | Vim-style commands with `:` prefix |
| **Mouse Support** | Click to select, scroll to navigate |
| **Session Persistence** | Automatic save/restore of application state |

### 🔧 Request Builder

- **URL** with variable substitution (`{{base_url}}/api/users`)
- **Headers** with key-value editor
- **Body Types**: JSON, Form Data, Raw Text, Binary
- **Query Parameters** with interactive table
- **Authorization**: Bearer Token, API Key, Basic Auth

### 📊 Response Viewer

- **Formatted Output**: JSON syntax highlighting
- **Status Badges**: Color-coded (2xx green, 4xx orange, 5xx red)
- **Metadata**: Response time, size, headers
- **Tabs**: Body, Headers, Cookies, Console

### 📟 Console (Request History)

- **Request/Response Logging**: View chronological history of all HTTP requests made during session
- **Visual Status Indicators**: Color-coded badges for 2xx/3xx/4xx/5xx responses and network errors
- **Quick Actions**: Resend requests, copy headers/body/error to clipboard
- **Vim Navigation**: Browse history with j/k, expand entries with Enter/l

### 📋 cURL Import/Export

- **Import cURL Commands** (`Ctrl+I`): Paste cURL commands to create requests
  - Supports multiline commands (backslash or backtick continuation)
  - Parses method, URL, headers, body, and authentication
  - Converts shell variables (`$VAR`, `${VAR}`) to LazyCurl syntax (`{{VAR}}`)
  - Handles `-u`/`--user` for Basic authentication
- **Export as cURL** (`Ctrl+E`): Copy current request as a cURL command to clipboard
  - Generates valid cURL with proper quoting
  - Includes headers, body, and authentication

### 💾 Session Persistence

- **Automatic State Saving**: Application state is saved automatically on changes
- **Seamless Restore**: Resume exactly where you left off when reopening
- **What's Saved**: Active panel, selected request, active environment, expanded folders, scroll positions, active tabs
- **Smart Debouncing**: 500ms delay prevents excessive writes during rapid changes
- **Atomic Writes**: Safe file operations using temp file + rename pattern
- **Graceful Degradation**: Missing or invalid session files are handled silently

### 🌍 Environment System

```json
{
  "name": "Development",
  "variables": {
    "base_url": { "value": "http://localhost:3000", "active": true },
    "api_token": { "value": "secret-token", "secret": true, "active": true }
  }
}
```

**Built-in System Variables:**

- `{{$timestamp}}` — Unix timestamp
- `{{$datetime}}` — RFC3339 datetime
- `{{$uuid}}` — UUID v4
- `{{$randomInt}}` — Random integer
- `{{$random}}` — Random alphanumeric string

---

## 📦 Installation

### Prerequisites

- Go 1.21 or higher
- Terminal with Unicode support

### From Source

```bash
# Clone the repository
git clone https://github.com/kbrdn1/LazyCurl.git
cd LazyCurl

# Build
make build

# Run
./bin/lazycurl

# Or install globally
make install
```

### Using Go Install

```bash
go install github.com/kbrdn1/LazyCurl/cmd/lazycurl@latest
```

### Development Mode

```bash
# Live reload during development
make dev
```

---

## 🚀 Quick Start

### 1. Initialize a Workspace

```bash
cd your-api-project
lazycurl
```

LazyCurl automatically creates a `.lazycurl/` directory:

```
.lazycurl/
├── config.yaml           # Workspace settings
├── collections/          # Your API collections
│   └── example.json
└── environments/         # Environment files
    └── development.json
```

### 2. Create Your First Request

1. Press `n` to create a new request
2. Enter request details (name, method, URL)
3. Press `Enter` to confirm
4. Press `Ctrl+S` to send the request

### 3. Use Variables

1. Press `2` or `Tab` to switch to Environments tab
2. Press `n` to create a variable
3. Use `{{variable_name}}` in your URLs and body

---

## ⌨️ Keyboard Shortcuts

### Vim-Style Modes

| Mode | Indicator | Description |
|------|-----------|-------------|
| **NORMAL** | `NORMAL` (blue) | Default mode, navigate and execute commands |
| **INSERT** | `INSERT` (gray) | Text input mode for editing fields |
| **VIEW** | `VIEW` (green) | Read-only browsing of responses |
| **COMMAND** | `COMMAND` (orange) | Execute commands with `:` prefix |

### Navigation

| Key | Action |
|-----|--------|
| `h` / `l` | Switch panels (left/right) |
| `j` / `k` | Move up/down in lists |
| `g` / `G` | Jump to top/bottom |
| `1` / `2` | Switch tabs (Collections/Environments) |
| `Tab` | Next tab in request builder |

### Collections Panel

| Key | Action |
|-----|--------|
| `n` | New request |
| `N` | New folder |
| `c` / `i` | Edit request |
| `R` | Rename |
| `d` | Delete |
| `D` | Duplicate |
| `y` | Yank (copy) |
| `p` | Paste |
| `/` | Search |
| `Enter` / `Space` | Open request |

### Environments Panel

| Key | Action |
|-----|--------|
| `n` | New variable |
| `N` | New environment |
| `c` / `i` | Edit value |
| `R` | Rename |
| `d` | Delete |
| `D` | Duplicate |
| `a` / `A` | Toggle active |
| `s` | Toggle secret |
| `S` / `Enter` | Select environment |
| `/` | Search |

### Response Panel

| Key | Action |
|-----|--------|
| `1` | Body tab |
| `2` | Cookies tab |
| `3` | Headers tab |
| `4` | Console tab (request history) |
| `Tab` / `Shift+Tab` | Next/previous tab |
| `j` / `k` | Scroll content / navigate list |
| `g` / `G` | Jump to top/bottom |

### Console Tab (List View)

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate up/down in history |
| `g` / `G` | Jump to first/last entry |
| `Enter` / `l` | Expand selected entry |
| `R` | Resend selected request |
| `U` | Copy URL to clipboard |

### Console Tab (Expanded View)

| Key | Action |
|-----|--------|
| `Esc` / `h` / `q` | Collapse back to list |
| `R` | Resend request |
| `H` | Copy headers |
| `B` | Copy body |
| `E` | Copy error message |
| `C` | Copy cookies |
| `I` | Copy request info |
| `A` | Copy all (request & response) |

### Search Mode

| Key | Action |
|-----|--------|
| `n` | Next match |
| `N` | Previous match |
| `i` | Edit search query |
| `Esc` | Clear search |
| `Enter` / `Space` | Open selected item |

### Global

| Key | Action |
|-----|--------|
| `Ctrl+S` | Send request |
| `Ctrl+I` | Import cURL command |
| `Ctrl+E` | Export request as cURL (copy to clipboard) |
| `?` | Show keybinding help (WhichKey) |
| `:` | Enter command mode |
| `Esc` | Return to NORMAL mode |
| `q` | Quit |

### Command Mode

| Command | Action |
|---------|--------|
| `:q` | Quit |
| `:w` | Save |
| `:wq` | Save and quit |
| `:help` | Show help |
| `:ws list` | List workspaces |

---

## 📁 File Structure

### Workspace Layout

```
your-project/
├── .lazycurl/
│   ├── config.yaml              # Workspace configuration
│   ├── session.yml              # Session state (auto-generated)
│   ├── collections/
│   │   ├── api.json             # Collection file
│   │   └── admin.json
│   └── environments/
│       ├── development.json     # Environment file
│       ├── staging.json
│       └── production.json
└── ...
```

### Global Configuration

```
~/.config/lazycurl/
└── config.yaml                  # Global settings, themes, keybindings
```

### Collection Format

```json
{
  "name": "My API",
  "description": "API collection description",
  "folders": [
    {
      "name": "Users",
      "requests": [
        {
          "id": "req_001",
          "name": "Get All Users",
          "method": "GET",
          "url": "{{base_url}}/api/users",
          "headers": {
            "Authorization": "Bearer {{token}}"
          }
        },
        {
          "id": "req_002",
          "name": "Create User",
          "method": "POST",
          "url": "{{base_url}}/api/users",
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "name": "{{user_name}}",
            "email": "{{user_email}}"
          }
        }
      ]
    }
  ]
}
```

### Environment Format

```json
{
  "name": "Development",
  "description": "Local development environment",
  "variables": {
    "base_url": {
      "value": "http://localhost:3000",
      "active": true
    },
    "token": {
      "value": "dev-secret-token",
      "secret": true,
      "active": true
    },
    "user_name": {
      "value": "Test User",
      "active": true
    }
  }
}
```

---

## 🎨 Theme

LazyCurl uses the **Catppuccin Mocha** color scheme:

| Element | Color |
|---------|-------|
| Primary (Selection) | Lavender `#b4befe` |
| Secondary | Blue `#89b4fa` |
| Success/Active | Green `#a6e3a1` |
| Warning | Peach `#fab387` |
| Error | Red `#f38ba8` |
| Text | White `#cdd6f4` |
| Background | Base `#1e1e2e` |

**HTTP Method Colors:**

- GET/HEAD: Green
- POST: Orange
- PUT: Blue
- PATCH: Purple
- DELETE: Red
- OPTIONS: Brown

---

## 📚 Documentation

Full documentation is available in the `docs/` directory:

| Document | Description |
|----------|-------------|
| [Installation](docs/installation.md) | Complete installation guide |
| [Getting Started](docs/getting-started.md) | First steps with LazyCurl |
| [Keybindings](docs/keybindings.md) | Complete keyboard reference |
| [Configuration](docs/configuration.md) | Config files and options |
| [Collections](docs/collections.md) | Managing request collections |
| [Environments](docs/environments.md) | Variables and environments |
| [StatusBar](docs/statusbar.md) | StatusBar component and modes |
| [Architecture](docs/architecture.md) | Technical architecture |

---

## 🛠️ Development

### Make Commands

```bash
make build          # Build binary to bin/lazycurl
make run            # Build and run
make dev            # Live reload with air
make test           # Run tests
make test-coverage  # Generate coverage report
make fmt            # Format code
make lint           # Run linter
make clean          # Clean build artifacts
make build-all      # Cross-compile for all platforms
```

### Project Structure

```
LazyCurl/
├── cmd/lazycurl/           # Application entry point
├── internal/
│   ├── api/                # HTTP client, collections, environments
│   ├── config/             # Configuration management
│   ├── format/             # Response formatting
│   └── ui/                 # User interface
│       ├── components/     # Reusable UI components
│       └── *.go            # Panel implementations
├── pkg/styles/             # Catppuccin theme and styles
├── docs/                   # Documentation
└── Makefile
```

---

## 🤝 Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

### Quick Start

```bash
# Fork and clone
git clone https://github.com/YOUR_USERNAME/LazyCurl.git
cd LazyCurl

# Create feature branch
git checkout -b feat/#123-your-feature

# Make changes and commit (using Gitmoji)
git commit -m "✨ feat(ui): add new feature"

# Push and create PR
git push origin feat/#123-your-feature
```

### Commit Convention

We use **Gitmoji + Conventional Commits**:

```
✨ feat(scope): add new feature
🐛 fix(scope): fix bug
📝 docs(scope): update documentation
♻️ refactor(scope): refactor code
🎨 style(scope): improve styling
✅ test(scope): add tests
```

---

## 🗺️ Roadmap

### ✅ Phase 1 - Foundation (Complete)

- Lazygit-style multi-panel interface
- Vim-style navigation and modes
- Workspace system with file-based storage
- Configuration system (global + workspace)
- Collections and environments management

### 🔥 Sprint 1 - MVP (Complete)

- [x] Collection tree view with folders
- [x] Environment management with variables
- [x] Request builder UI
- [x] Search functionality
- [x] WhichKey keybinding hints
- [x] HTTP request execution
- [x] Response formatting
- [x] CI/CD pipeline

### 📋 Sprint 2 - UX Improvements

- [x] Responsive panel layout (horizontal on 80x24 terminals) [#7](https://github.com/kbrdn1/LazyCurl/issues/7)
- [x] Fullscreen panel toggle (`F` in NORMAL mode, `h/l` to navigate) [#8](https://github.com/kbrdn1/LazyCurl/issues/8)
- [x] Find in editors (`/` in NORMAL mode) [#24](https://github.com/kbrdn1/LazyCurl/issues/24)
- [ ] Settings Panel (`Ctrl+;` fullscreen with Global/Workspace tabs) [#25](https://github.com/kbrdn1/LazyCurl/issues/25)
- [x] Console tab in Response Panel (request/response history) [#9](https://github.com/kbrdn1/LazyCurl/issues/9)
- [x] Improved statusbar rendering and display [#10](https://github.com/kbrdn1/LazyCurl/issues/10)
- [x] Session persistence (`.lazycurl/session.yml`) [#11](https://github.com/kbrdn1/LazyCurl/issues/11)
- [ ] Theme system refactoring [#12](https://github.com/kbrdn1/LazyCurl/issues/12)
- [ ] Theme management and custom themes [#13](https://github.com/kbrdn1/LazyCurl/issues/13)

### 🔮 Future

- Import/export Postman collections [#14](https://github.com/kbrdn1/LazyCurl/issues/14)
- Request history [#15](https://github.com/kbrdn1/LazyCurl/issues/15)
- Pre-request & post-response scripting (JavaScript via Goja) [#35](https://github.com/kbrdn1/LazyCurl/issues/35)
- Authentication helpers (OAuth2, AWS Sig) [#17](https://github.com/kbrdn1/LazyCurl/issues/17)
- GraphQL support [#18](https://github.com/kbrdn1/LazyCurl/issues/18)
- WebSocket testing [#19](https://github.com/kbrdn1/LazyCurl/issues/19)
- gRPC support [#20](https://github.com/kbrdn1/LazyCurl/issues/20)
- CLI commands architecture [#26](https://github.com/kbrdn1/LazyCurl/issues/26)

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

LazyCurl is built on the shoulders of giants:

- **[Lazygit](https://github.com/jesseduffield/lazygit)** — Inspiration for the TUI design
- **[Postman](https://www.postman.com/)** — Inspiration for API testing features
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** — TUI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — Terminal styling
- **[Catppuccin](https://github.com/catppuccin)** — Color scheme

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/kbrdn1">@kbrdn1</a>
</p>

<p align="center">
  <a href="https://github.com/kbrdn1/LazyCurl/stargazers">⭐ Star us on GitHub</a>
</p>
