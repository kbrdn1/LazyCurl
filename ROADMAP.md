# LazyCurl Roadmap

> Development roadmap for LazyCurl - a powerful TUI HTTP client

## Vision

LazyCurl aims to be the **go-to terminal HTTP client** for developers who prefer keyboard-driven workflows. Our goals:

- **Single binary** — No runtime dependencies, easy installation
- **Vim-first UX** — Authentic modal editing with familiar keybindings
- **File-based storage** — Git-friendly JSON collections and environments
- **Full-featured** — Match desktop tools like Postman and Insomnia

---

## Completed Releases

### v1.0.0 - Foundation (Sprint 1-2)

Core TUI with essential HTTP testing capabilities.

- [x] Lazygit-style multi-panel interface
- [x] Vim modes (NORMAL, INSERT, VIEW, COMMAND)
- [x] Collections and folders management
- [x] Environment variables with `{{var}}` syntax
- [x] HTTP request builder (method, URL, headers, body)
- [x] Response viewer with JSON formatting
- [x] Session persistence
- [x] Console tab (request history)
- [x] WhichKey help system
- [x] Search with `/`

### v1.1.0 - Import/Export (Sprint 3a)

Interoperability with existing tools and workflows.

- [x] cURL import/export ([#60](https://github.com/kbrdn1/LazyCurl/issues/60))
- [x] Jump mode navigation ([#61](https://github.com/kbrdn1/LazyCurl/issues/61))

### v1.2.0 - External Tools (Sprint 3b)

Integration with external editors and API specs.

- [x] External editor integration ([#65](https://github.com/kbrdn1/LazyCurl/issues/65))
- [x] OpenAPI 3.x import with security schemes ([#66](https://github.com/kbrdn1/LazyCurl/issues/66), [#71](https://github.com/kbrdn1/LazyCurl/issues/71))
- [x] Postman collection/environment import ([#14](https://github.com/kbrdn1/LazyCurl/issues/14), [#72](https://github.com/kbrdn1/LazyCurl/issues/72))

---

## Current Sprint

### Sprint 4 - Advanced Features

**Goal**: Power user features for advanced API testing workflows.

#### Critical Priority 🔴

| Feature | Issue | Description |
|---------|-------|-------------|
| JavaScript Scripting | [#35](https://github.com/kbrdn1/LazyCurl/issues/35) | Pre/post-request scripts using Goja JS engine |
| Request Chaining | [#42](https://github.com/kbrdn1/LazyCurl/issues/42) | Extract values from responses, chain requests |

#### High Priority 🟡

| Feature | Issue | Description |
|---------|-------|-------------|
| Test Assertions | [#43](https://github.com/kbrdn1/LazyCurl/issues/43) | Assert response status, body, headers |
| Collection Runner | [#44](https://github.com/kbrdn1/LazyCurl/issues/44) | Run all requests in sequence |
| Fuzzy Finder | [#45](https://github.com/kbrdn1/LazyCurl/issues/45) | fzf-style quick search |

#### Medium Priority 🟢

| Feature | Issue | Description |
|---------|-------|-------------|
| Request Diff | [#46](https://github.com/kbrdn1/LazyCurl/issues/46) | Compare two responses |
| Request Templates | [#47](https://github.com/kbrdn1/LazyCurl/issues/47) | Reusable request patterns |
| Settings Panel | [#25](https://github.com/kbrdn1/LazyCurl/issues/25) | In-app configuration UI |
| Theme System | [#12](https://github.com/kbrdn1/LazyCurl/issues/12), [#13](https://github.com/kbrdn1/LazyCurl/issues/13) | Theme refactoring and UI |
| Hot Reload Config | [#41](https://github.com/kbrdn1/LazyCurl/issues/41) | Auto-reload on config change |

---

## Future Sprints

### Sprint 5 - Protocol Expansion

**Goal**: Support for modern API protocols beyond REST.

| Feature | Issue | Priority | Description |
|---------|-------|----------|-------------|
| GraphQL Support | [#18](https://github.com/kbrdn1/LazyCurl/issues/18) | 🔴 Critical | Schema explorer, variables, queries |
| WebSocket Testing | [#19](https://github.com/kbrdn1/LazyCurl/issues/19) | 🟡 High | Interactive WS client |
| gRPC Support | [#20](https://github.com/kbrdn1/LazyCurl/issues/20) | 🟡 High | Proto reflection, streaming |
| SSE Support | [#48](https://github.com/kbrdn1/LazyCurl/issues/48) | 🟢 Medium | Server-Sent Events viewer |

### Sprint 6 - Enterprise Features

**Goal**: Features for enterprise and production API testing.

| Feature | Issue | Priority | Description |
|---------|-------|----------|-------------|
| OAuth2 Flows | [#17](https://github.com/kbrdn1/LazyCurl/issues/17) | 🔴 Critical | Auth code, client credentials, refresh |
| AWS Signature v4 | [#17](https://github.com/kbrdn1/LazyCurl/issues/17) | 🟡 High | AWS API authentication |
| mTLS / Client Certs | [#49](https://github.com/kbrdn1/LazyCurl/issues/49) | 🟡 High | Mutual TLS authentication |
| Proxy Support | [#50](https://github.com/kbrdn1/LazyCurl/issues/50) | 🟡 High | HTTP/SOCKS proxy |
| Request Retry | [#51](https://github.com/kbrdn1/LazyCurl/issues/51) | 🟢 Medium | Auto-retry with backoff |
| Rate Limiting | [#52](https://github.com/kbrdn1/LazyCurl/issues/52) | 🟢 Medium | Respect API rate limits |

### Sprint 7 - CLI & Automation

**Goal**: Headless operation and CI/CD integration.

| Feature | Issue | Priority | Description |
|---------|-------|----------|-------------|
| CLI Mode | [#26](https://github.com/kbrdn1/LazyCurl/issues/26) | 🔴 Critical | `lazycurl run collection.json` |
| CI/CD Integration | [#53](https://github.com/kbrdn1/LazyCurl/issues/53) | 🔴 Critical | Exit codes, JSON output |
| Request Export | [#54](https://github.com/kbrdn1/LazyCurl/issues/54) | 🟡 High | Export to Go, Python, JS code |
| Mock Server | [#55](https://github.com/kbrdn1/LazyCurl/issues/55) | 🟢 Medium | Local mock from collection |
| API Docs Generator | [#56](https://github.com/kbrdn1/LazyCurl/issues/56) | 🟢 Medium | Generate docs from collection |

---

## Backlog

Features not yet scheduled:

| Feature | Issue | Description |
|---------|-------|-------------|
| Animated Dashboard | [#36](https://github.com/kbrdn1/LazyCurl/issues/36) | Workspace selector with splash screen |

---

## Priority Legend

| Symbol | Priority | Description |
|--------|----------|-------------|
| 🔴 | Critical | Core functionality, blocks other features |
| 🟡 | High | Important for user experience |
| 🟢 | Medium | Nice to have, quality of life |

---

## Contributing

Want to help? Check out:

1. [Contributing Guide](CONTRIBUTING.md)
2. [Good First Issues](https://github.com/kbrdn1/LazyCurl/labels/good%20first%20issue)
3. [Help Wanted](https://github.com/kbrdn1/LazyCurl/labels/help%20wanted)

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for release history.
