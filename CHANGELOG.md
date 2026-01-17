# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0](https://github.com/kbrdn1/LazyCurl/compare/v1.1.0...v1.2.0) (2026-01-17)


### Features

* **cli:** add Postman import command support ✨ ([#74](https://github.com/kbrdn1/LazyCurl/issues/74)) ([50ec0c7](https://github.com/kbrdn1/LazyCurl/commit/50ec0c75b4d89e95c601d91c83584e99b1d1a18b)), closes [#72](https://github.com/kbrdn1/LazyCurl/issues/72)
* **openapi:** add security scheme import support ✨ ([#73](https://github.com/kbrdn1/LazyCurl/issues/73)) ([3fab935](https://github.com/kbrdn1/LazyCurl/commit/3fab935e7ac63f2cf6a71e6677bbeb5580a50b31))


### Bug Fixes

* **cli:** address CodeRabbit and Copilot review comments from PR [#74](https://github.com/kbrdn1/LazyCurl/issues/74) 🐛 ([5eaa014](https://github.com/kbrdn1/LazyCurl/commit/5eaa014aa2e0c783f0c955586f69ac9db5272e2e))
* **pre-commit:** require emoji at end for release-please compatibility ([3f91ed7](https://github.com/kbrdn1/LazyCurl/commit/3f91ed7c2086ca9df05465b1d32a28501ce1d1b8))

## [1.1.0](https://github.com/kbrdn1/LazyCurl/compare/v1.0.0...v1.1.0) (2026-01-17)


### Features

* **api:** add cURL import/export functionality ✨ ([#63](https://github.com/kbrdn1/LazyCurl/issues/63)) ([1eb5dae](https://github.com/kbrdn1/LazyCurl/commit/1eb5daeff9f0d9f259920c277069f32ac6a34958))
* **tools:** add CLAUDE.md copy to worktree setup ([92ac991](https://github.com/kbrdn1/LazyCurl/commit/92ac991ae9960c8104a766b775f7708ebd3eb3eb))
* **tools:** add speckit integration and Claude launch to worktree manager ([0f0f027](https://github.com/kbrdn1/LazyCurl/commit/0f0f027c7d578ca03eb7e862468f60f5058f68d8))
* **tools:** auto-copy .claude/commands directory in worktree setup ([4c332c8](https://github.com/kbrdn1/LazyCurl/commit/4c332c8b897cd21fadf839168fa28793281cfeda))
* **tools:** auto-copy .specify directory in worktree setup ([10854e0](https://github.com/kbrdn1/LazyCurl/commit/10854e06f37f1a5713dd1871d97011247252c82b))
* **ui:** add jump mode navigation (vim-easymotion style) ✨ ([#64](https://github.com/kbrdn1/LazyCurl/issues/64)) ([a3cf8f7](https://github.com/kbrdn1/LazyCurl/commit/a3cf8f7c66458dd7f7caca2f3a1a162197d7884a))

## 1.0.0 (2026-01-17)


### Features

* **ui:** add variable interpolation highlighting and preview in body editor ✨ ([#58](https://github.com/kbrdn1/LazyCurl/issues/58)) ([46d81b9](https://github.com/kbrdn1/LazyCurl/commit/46d81b912d04a2b60cb9c218b5dd1e84958070d4))

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
