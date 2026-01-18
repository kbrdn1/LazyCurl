# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.1](https://github.com/kbrdn1/LazyCurl/compare/v1.3.0...v1.3.1) (2026-01-18)


### Bug Fixes

* **roadmap:** correct Request Chaining status - not complete 🐛 ([2e9e9dc](https://github.com/kbrdn1/LazyCurl/commit/2e9e9dc5612b4599f24b01b3a33385395051b977))

## [1.3.0](https://github.com/kbrdn1/LazyCurl/compare/v1.2.0...v1.3.0) (2026-01-18)

### Features

* **scripting:** add JavaScript scripting engine with Goja runtime 🔧 ([#75](https://github.com/kbrdn1/LazyCurl/pull/75)) ([9804a26](https://github.com/kbrdn1/LazyCurl/commit/9804a26))
  * Pre-request and post-response script execution
  * `lc.request` — Read/modify HTTP request (URL, headers, body, params)
  * `lc.response` — Access response data (status, headers, body)
  * `lc.env` / `lc.globals` — Environment and session variables
  * `lc.test` / `lc.expect` — Test assertions with 16 matchers
  * `lc.cookies`, `lc.crypto`, `lc.base64`, `lc.variables` — Utility modules
  * `lc.sendRequest` — Request chaining within scripts
  * `console` — Logging functions (log, info, warn, error, debug)
* **docs:** add comprehensive JavaScript scripting API reference 📚
  * 12 API module documentation pages in `docs/api/`
  * 3 practical example pages with 59 code examples
  * GitHub Wiki with sidebar navigation (19 pages)
  * Documentation contribution guide with sync process
* **devtools:** add non-interactive worktree creation for Claude Code ([cf095d6](https://github.com/kbrdn1/LazyCurl/commit/cf095d6))

## [1.2.0](https://github.com/kbrdn1/LazyCurl/compare/v1.1.0...v1.2.0) (2026-01-17)

### Features

* **cli:** add Postman import command support ✨ ([#74](https://github.com/kbrdn1/LazyCurl/issues/74)) ([50ec0c7](https://github.com/kbrdn1/LazyCurl/commit/50ec0c7)), closes [#72](https://github.com/kbrdn1/LazyCurl/issues/72)
* **openapi:** add security scheme import support ✨ ([#73](https://github.com/kbrdn1/LazyCurl/issues/73)) ([3fab935](https://github.com/kbrdn1/LazyCurl/commit/3fab935))

### Bug Fixes

* **cli:** address CodeRabbit and Copilot review comments from PR [#74](https://github.com/kbrdn1/LazyCurl/issues/74) 🐛 ([5eaa014](https://github.com/kbrdn1/LazyCurl/commit/5eaa014))
* **pre-commit:** require emoji at end for release-please compatibility ([3f91ed7](https://github.com/kbrdn1/LazyCurl/commit/3f91ed7))

## [1.1.0](https://github.com/kbrdn1/LazyCurl/compare/v1.0.0...v1.1.0) (2026-01-17)

### Features

* **api:** add cURL import/export functionality ✨ ([#63](https://github.com/kbrdn1/LazyCurl/issues/63)) ([1eb5dae](https://github.com/kbrdn1/LazyCurl/commit/1eb5dae))
* **ui:** add jump mode navigation (vim-easymotion style) ✨ ([#64](https://github.com/kbrdn1/LazyCurl/issues/64)) ([a3cf8f7](https://github.com/kbrdn1/LazyCurl/commit/a3cf8f7))
* **tools:** add worktree manager improvements for Claude Code workflows

## [1.0.0](https://github.com/kbrdn1/LazyCurl/releases/tag/v1.0.0) (2026-01-17)

### Features

* **ui:** Lazygit-style multi-panel interface (Collections, Request, Response)
* **ui:** Vim motions for navigation (h/j/k/l) with NORMAL, INSERT, VIEW, COMMAND modes
* **ui:** Variable interpolation highlighting and preview in body editor ✨ ([#58](https://github.com/kbrdn1/LazyCurl/issues/58))
* **core:** Workspace system with `.lazycurl/` directory structure
* **core:** YAML-based configuration (global and workspace)
* **core:** Environment variables system with `{{var}}` syntax
* **core:** Session persistence with auto-save/restore
* **ui:** Collections panel for browsing requests and folders
* **ui:** Request builder panel (URL, headers, body, params, auth)
* **ui:** Response viewer with JSON formatting, headers, cookies tabs
* **ui:** Console tab with request history
* **ui:** WhichKey help system

---
