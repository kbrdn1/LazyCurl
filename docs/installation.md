# Installation Guide

Complete guide to install LazyCurl on your system.

## Table of Contents

- [Requirements](#requirements)
- [Quick Install](#quick-install)
- [Installation Methods](#installation-methods)
  - [From Source](#from-source)
  - [Go Install](#go-install)
  - [Pre-built Binaries](#pre-built-binaries)
  - [Homebrew (macOS)](#homebrew-macos)
- [Post-Installation](#post-installation)
- [Verify Installation](#verify-installation)
- [Uninstallation](#uninstallation)
- [Troubleshooting](#troubleshooting)

---

## Requirements

### System Requirements

| Requirement | Minimum |
|-------------|---------|
| Operating System | Linux, macOS, Windows |
| Terminal | UTF-8 support, 256 colors recommended |
| Terminal Size | 80x24 minimum, 120x40 recommended |

### Build Requirements (from source)

| Requirement | Version |
|-------------|---------|
| Go | 1.21 or higher |
| Git | Any recent version |
| Make | GNU Make (optional) |

---

## Quick Install

### One-liner (Go required)

```bash
go install github.com/kbrdn1/LazyCurl/cmd/lazycurl@latest
```

### One-liner (from source)

```bash
git clone https://github.com/kbrdn1/LazyCurl.git && cd LazyCurl && make build && ./bin/lazycurl
```

---

## Installation Methods

### From Source

This is the recommended method for development or getting the latest features.

#### 1. Clone the repository

```bash
git clone https://github.com/kbrdn1/LazyCurl.git
cd LazyCurl
```

#### 2. Build the binary

Using Make:
```bash
make build
```

Or using Go directly:
```bash
go build -o bin/lazycurl ./cmd/lazycurl
```

#### 3. Install globally (optional)

**Option A: Copy to /usr/local/bin (requires sudo)**
```bash
sudo cp bin/lazycurl /usr/local/bin/
```

**Option B: Copy to ~/bin (user-local)**
```bash
mkdir -p ~/bin
cp bin/lazycurl ~/bin/
# Add to PATH if not already (add to ~/.zshrc or ~/.bashrc)
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**Option C: Use Go install**
```bash
make install
# or
go install ./cmd/lazycurl
```

This installs to `$GOPATH/bin` (usually `~/go/bin`).

---

### Go Install

If you have Go installed, this is the simplest method:

```bash
go install github.com/kbrdn1/LazyCurl/cmd/lazycurl@latest
```

The binary will be installed to `$GOPATH/bin/lazycurl`.

#### Ensure GOPATH/bin is in PATH

Add to your shell configuration (`~/.zshrc`, `~/.bashrc`, or `~/.profile`):

```bash
export PATH="$HOME/go/bin:$PATH"
```

Then reload:
```bash
source ~/.zshrc  # or ~/.bashrc
```

---

### Pre-built Binaries

Download pre-built binaries from the [Releases](https://github.com/kbrdn1/LazyCurl/releases) page.

#### Linux (AMD64)

```bash
curl -LO https://github.com/kbrdn1/LazyCurl/releases/latest/download/lazycurl_linux_amd64.tar.gz
tar -xzf lazycurl_linux_amd64.tar.gz
sudo mv lazycurl /usr/local/bin/
```

#### Linux (ARM64)

```bash
curl -LO https://github.com/kbrdn1/LazyCurl/releases/latest/download/lazycurl_linux_arm64.tar.gz
tar -xzf lazycurl_linux_arm64.tar.gz
sudo mv lazycurl /usr/local/bin/
```

#### macOS (Intel)

```bash
curl -LO https://github.com/kbrdn1/LazyCurl/releases/latest/download/lazycurl_darwin_amd64.tar.gz
tar -xzf lazycurl_darwin_amd64.tar.gz
sudo mv lazycurl /usr/local/bin/
```

#### macOS (Apple Silicon)

```bash
curl -LO https://github.com/kbrdn1/LazyCurl/releases/latest/download/lazycurl_darwin_arm64.tar.gz
tar -xzf lazycurl_darwin_arm64.tar.gz
sudo mv lazycurl /usr/local/bin/
```

#### Windows

1. Download `lazycurl_windows_amd64.zip` from [Releases](https://github.com/kbrdn1/LazyCurl/releases)
2. Extract the ZIP file
3. Move `lazycurl.exe` to a directory in your PATH
4. Or add the extraction directory to your PATH

Using PowerShell:
```powershell
# Download
Invoke-WebRequest -Uri "https://github.com/kbrdn1/LazyCurl/releases/latest/download/lazycurl_windows_amd64.zip" -OutFile "lazycurl.zip"

# Extract
Expand-Archive -Path "lazycurl.zip" -DestinationPath "C:\Program Files\LazyCurl"

# Add to PATH (run as Administrator)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Program Files\LazyCurl", "Machine")
```

---

### Homebrew (macOS)

> Coming soon! Homebrew tap is planned for future releases.

```bash
# Future installation method
brew tap kbrdn1/tap
brew install lazycurl
```

---

## Post-Installation

### Create Global Configuration

LazyCurl stores global configuration in `~/.config/lazycurl/`:

```bash
mkdir -p ~/.config/lazycurl
```

The configuration file will be created automatically on first run, or you can create it manually:

```yaml
# ~/.config/lazycurl/config.yaml
theme: "catppuccin-mocha"
editor: "vim"
keybindings:
  quit: ["q", "ctrl+c"]
  send: ["ctrl+s"]
  help: ["?"]
```

### Initialize a Workspace

Navigate to your API project directory and run:

```bash
cd your-api-project
lazycurl
```

This creates a `.lazycurl/` directory with:

```
.lazycurl/
├── config.yaml           # Workspace settings
├── collections/          # Your API collections
└── environments/         # Environment files
```

---

## Verify Installation

### Check binary location

```bash
which lazycurl
# Expected: /usr/local/bin/lazycurl or ~/go/bin/lazycurl
```

### Check version

```bash
lazycurl version
```

### Run the application

```bash
lazycurl
```

You should see the TUI interface with three panels.

### Quick test

```bash
# Create a test directory
mkdir -p /tmp/lazycurl-test
cd /tmp/lazycurl-test

# Run LazyCurl
lazycurl

# Press 'q' to quit
```

---

## Uninstallation

### If installed via Go

```bash
rm $(go env GOPATH)/bin/lazycurl
```

### If installed to /usr/local/bin

```bash
sudo rm /usr/local/bin/lazycurl
```

### If installed to ~/bin

```bash
rm ~/bin/lazycurl
```

### Remove configuration (optional)

```bash
# Remove global config
rm -rf ~/.config/lazycurl

# Remove workspace configs (in each project)
rm -rf .lazycurl/
```

---

## Troubleshooting

### "command not found: lazycurl"

**Cause**: Binary not in PATH.

**Solution**: Check installation location and add to PATH:

```bash
# Find the binary
find ~ -name "lazycurl" -type f 2>/dev/null

# Add to PATH (example for ~/go/bin)
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

### "permission denied"

**Cause**: Binary not executable or PATH directory not writable.

**Solution**:
```bash
chmod +x /path/to/lazycurl
```

### "could not open TTY" or display issues

**Cause**: Running in non-interactive terminal or terminal doesn't support TUI.

**Solution**:
- Run in a proper terminal emulator (not in IDE terminal or pipe)
- Ensure terminal supports UTF-8: `echo $LANG` should contain "UTF-8"
- Try a different terminal (iTerm2, Alacritty, WezTerm)

### Build fails with Go errors

**Cause**: Go version too old or dependencies issue.

**Solution**:
```bash
# Check Go version (need 1.21+)
go version

# Update dependencies
go mod download
go mod tidy

# Rebuild
make clean build
```

### Colors look wrong

**Cause**: Terminal doesn't support 256 colors or true color.

**Solution**:
- Use a modern terminal (iTerm2, Alacritty, WezTerm, Windows Terminal)
- Set TERM environment: `export TERM=xterm-256color`
- Enable true color in terminal settings

### Keybindings don't work

**Cause**: Terminal intercepting keys or different key codes.

**Solution**:
- Check if terminal has conflicting shortcuts
- Try different terminal emulator
- Check `~/.config/lazycurl/config.yaml` for custom keybindings

---

## Development Installation

For contributing to LazyCurl:

```bash
# Clone with SSH (for contributors)
git clone git@github.com:kbrdn1/LazyCurl.git
cd LazyCurl

# Install dependencies
make deps

# Install pre-commit hooks
make setup-hooks

# Run in development mode (live reload)
make dev
```

See [CONTRIBUTING.md](../CONTRIBUTING.md) for more details.

---

## Next Steps

- [Getting Started](getting-started.md) - First steps with LazyCurl
- [Configuration](configuration.md) - Customize your setup
- [Keybindings](keybindings.md) - Learn keyboard shortcuts
- [Collections](collections.md) - Organize your API requests
