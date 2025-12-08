# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Lazygit-style multi-panel interface (Collections, Request, Response)
- Vim motions for navigation (h/j/k/l)
- Workspace system with `.lazycurl/` directory structure
- YAML-based configuration (global and workspace)
- Customizable keybindings system
- Customizable themes with Lipgloss
- Environment variables system (local + global)
- Collections panel for browsing requests
- Request builder panel with full HTTP request editing
  - URL editor with variable highlighting
  - Query and Path parameters management
  - Headers editor with enable/disable toggle
  - Body editor (JSON, XML, Form, Raw, None)
  - Authorization (Bearer, Basic, API Key)
  - Pre/Post request scripts
- Response viewer panel with HTTP response display
  - Animated loading indicator during requests
  - Status badge with color coding (2xx/3xx/4xx/5xx)
  - Time and Size metrics with icons
  - Body tab with JSON auto-formatting
  - Headers and Cookies tabs with vim navigation
  - Horizontal scrolling for long lines
- Environments panel with toggle ('e' key)
- Configuration loading from `~/.config/lazycurl/config.yaml`
- Workspace configuration from `.lazycurl/config.yaml`
- CI/CD pipeline with GitHub Actions
  - Automated testing on push/PR
  - Multi-platform builds (Linux, macOS, Windows)
  - golangci-lint integration
  - Security scanning with gosec
  - Automated releases with GoReleaser
- Pre-commit hooks for code quality

### Changed

- **BREAKING**: Removed all Git-related functionality
- **BREAKING**: Switched from JSON to YAML for configuration
- **BREAKING**: Changed project focus from Git+API tool to pure HTTP/API client
- Refactored UI to use Lazygit-inspired layout
- Updated architecture to focus solely on API testing

### Removed

- Git operations and git panel
- Git-related configuration options
- Split view mode (Git + API)

## [0.1.0] - 2025-10-23

### Added

- Initial project setup with Go modules
- Bubble Tea framework integration
- Bubbles components (viewport, textarea, list)
- Lipgloss styling system
- Bubble Zone for mouse support
- Basic TUI structure
- Makefile for build automation
- Project documentation (README, DEVELOPMENT_PLAN)

---

[Unreleased]: https://github.com/kbrdn1/LazyCurl/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/kbrdn1/LazyCurl/releases/tag/v0.1.0
