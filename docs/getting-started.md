# Getting Started with LazyCurl

Welcome to LazyCurl! This guide will help you get up and running with your first API requests.

## Table of Contents

- [Installation](#installation)
- [Your First Workspace](#your-first-workspace)
- [Creating Your First Request](#creating-your-first-request)
- [Using Environments](#using-environments)
- [Sending Requests](#sending-requests)
- [Next Steps](#next-steps)

---

## Installation

### Prerequisites

- **Go 1.21+** - [Download Go](https://go.dev/dl/)
- **Terminal** with Unicode support (most modern terminals)
- **Git** (optional, for cloning)

### Install from Source

```bash
# Clone the repository
git clone https://github.com/kbrdn1/LazyCurl.git
cd LazyCurl

# Build the application
make build

# The binary is now at ./bin/lazycurl
./bin/lazycurl --version
```

### Install Globally

```bash
# Install to your GOPATH/bin
make install

# Now you can run from anywhere
lazycurl
```

### Using Go Install

```bash
go install github.com/kbrdn1/LazyCurl/cmd/lazycurl@latest
```

---

## Your First Workspace

LazyCurl uses a workspace system to organize your API collections and environments.

### Automatic Initialization

Simply navigate to your project directory and run LazyCurl:

```bash
cd my-api-project
lazycurl
```

LazyCurl will automatically create the `.lazycurl/` directory structure:

```
my-api-project/
└── .lazycurl/
    ├── config.yaml           # Workspace configuration
    ├── collections/          # Your API collections
    │   └── example.json      # Sample collection
    └── environments/         # Environment files
        └── development.json  # Sample environment
```

### Manual Initialization

If you prefer to set up manually:

```bash
mkdir -p .lazycurl/{collections,environments}
```

### Workspace Configuration

The `config.yaml` file contains workspace settings:

```yaml
name: "My API Project"
description: "REST API for my application"
default_env: "development"
```

---

## Creating Your First Request

### Understanding the Interface

When you launch LazyCurl, you'll see three main panels:

```
┌─Collections────────┬─Request────────────────────────┐
│                    │                                │
│ Your collections   │   Request builder              │
│ and folders        │   (method, URL, headers, body) │
│                    │                                │
│                    ├─Response───────────────────────┤
│                    │                                │
│                    │   Response viewer              │
│                    │   (status, body, headers)      │
│                    │                                │
└────────────────────┴────────────────────────────────┘
```

### Step 1: Create a New Request

1. Make sure you're in the **Collections** panel (press `h` to navigate left if needed)
2. Press `n` to create a new request
3. Fill in the dialog:
   - **Name**: "Get Users"
   - **Method**: GET
   - **URL**: `https://jsonplaceholder.typicode.com/users`
4. Press `Enter` to confirm

### Step 2: View Your Request

Your new request appears in the Collections tree. Press `Enter` or `Space` to select it.

The Request panel (top-right) now shows:

- The HTTP method (GET)
- The URL
- Tabs for Headers, Body, Params, etc.

### Step 3: Add Headers (Optional)

1. Press `l` to navigate to the Request panel
2. Press `3` or navigate to the "Headers" tab
3. Press `n` to add a new header
4. Enter header details (e.g., `Accept: application/json`)

---

## Using Environments

Environments let you use variables like `{{base_url}}` that change based on your context (development, staging, production).

### Switch to Environments Tab

1. Navigate to the left panel with `h`
2. Press `2` to switch to the Environments tab

### Create a Variable

1. Press `N` to create a new environment (e.g., "Development")
2. Press `n` to add a variable:
   - **Name**: `base_url`
   - **Value**: `https://jsonplaceholder.typicode.com`
3. Press `Enter` to confirm

### Use Variables in Requests

Now update your request URL to use the variable:

1. Navigate to the Request panel
2. Edit the URL to: `{{base_url}}/users`

The variable will be replaced when you send the request.

### Variable Types

| Type | Description | Example |
|------|-------------|---------|
| **Regular** | Normal variable | `base_url` |
| **Secret** | Hidden in UI, for sensitive data | `api_token` |
| **Inactive** | Defined but not used | Toggle with `a` |

### System Variables

LazyCurl provides built-in system variables:

| Variable | Description |
|----------|-------------|
| `{{$timestamp}}` | Current Unix timestamp |
| `{{$datetime}}` | Current datetime (RFC3339) |
| `{{$date}}` | Current date (YYYY-MM-DD) |
| `{{$time}}` | Current time (HH:MM:SS) |
| `{{$uuid}}` | Random UUID v4 |
| `{{$randomInt}}` | Random integer (0-999999) |
| `{{$random}}` | Random 10-char string |

---

## Sending Requests

### Send a Request

1. Select a request in the Collections panel
2. Press `Ctrl+S` to send the request
3. The response appears in the Response panel (bottom-right)

### View the Response

The Response panel shows:

- **Status Badge**: Color-coded status code (200 OK = green)
- **Metadata**: Response time, size
- **Body Tab**: Formatted response body
- **Headers Tab**: Response headers
- **Cookies Tab**: Cookies set by the server

### Navigate Responses

- Press `j`/`k` to scroll through long responses
- Press `v` to enter VIEW mode for focused reading
- Press `Tab` to switch between Body/Headers/Cookies tabs

---

## Next Steps

Now that you're familiar with the basics, explore more features:

### Learn Keyboard Shortcuts

Press `?` at any time to see context-aware keybinding hints.

See the full [Keybindings Reference](keybindings.md).

### Organize with Folders

- Press `N` to create folders in your collections
- Drag requests between folders
- Use search (`/`) to find requests quickly

See [Collections Guide](collections.md).

### Multiple Environments

Create separate environments for different contexts:

- `development.json` - Local development
- `staging.json` - Staging server
- `production.json` - Production API

See [Environments Guide](environments.md).

### Customize Configuration

Edit `~/.config/lazycurl/config.yaml` to:

- Customize keybindings
- Change theme colors
- Set default editor

See [Configuration Guide](configuration.md).

---

## Troubleshooting

### LazyCurl Won't Start

1. Check Go version: `go version` (needs 1.21+)
2. Rebuild: `make clean && make build`
3. Check terminal supports Unicode

### Collections Not Loading

1. Check `.lazycurl/collections/` directory exists
2. Verify JSON files are valid: `cat .lazycurl/collections/*.json | jq .`
3. Check file permissions

### Variables Not Replaced

1. Ensure the environment is selected (green badge in status bar)
2. Check variable is marked as "active"
3. Verify syntax: `{{variable_name}}` (double curly braces)

### Getting Help

- Press `?` for keybinding help
- Press `:help` for command help
- Visit [GitHub Issues](https://github.com/kbrdn1/LazyCurl/issues)

---

## Quick Reference Card

| Action | Keys |
|--------|------|
| Navigate panels | `h` / `l` |
| Move in list | `j` / `k` |
| New request | `n` |
| New folder | `N` |
| Edit | `c` or `i` |
| Delete | `d` |
| Duplicate | `D` |
| Search | `/` |
| Send request | `Ctrl+S` |
| Switch to Environments | `2` |
| Show help | `?` |
| Quit | `q` |
