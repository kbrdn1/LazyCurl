# LazyCurl Roadmap

> **Vision** : Devenir le "Bruno du terminal" avec toutes les fonctionnalités de Yaak

---

## Vision Stratégique

LazyCurl vise à être le **client HTTP TUI de référence** pour les développeurs terminal-first :

- **Single binary Go** — Performance, distribution simple, démarrage instantané
- **Vim-first UX** — Motions authentiques (h/j/k/l), jump mode unique
- **Lazygit-style** — Interface familière pour les développeurs terminal
- **Git-native** — Collections JSON versionnables, 100% offline
- **Multi-protocol** — REST, GraphQL, WebSocket, gRPC, SSE (objectif Yaak)
- **Full-featured** — Scripting, assertions, runner (objectif Bruno)

---

## Completed Releases

### v1.0.0 - Foundation (Sprint 1-2) ✅

Core TUI avec capacités essentielles de test HTTP.

- [x] Interface multi-panneaux style Lazygit
- [x] Modes Vim (NORMAL, INSERT, VIEW, COMMAND)
- [x] Gestion collections et dossiers
- [x] Variables d'environnement avec syntaxe `{{var}}`
- [x] Builder de requêtes HTTP (method, URL, headers, body)
- [x] Viewer de réponses avec formatage JSON
- [x] Persistance de session
- [x] Onglet Console (historique des requêtes)
- [x] Système d'aide WhichKey
- [x] Recherche avec `/`

### v1.1.0 - Import/Export (Sprint 3a) ✅

Interopérabilité avec outils et workflows existants.

- [x] Import/export cURL ([#60](https://github.com/kbrdn1/LazyCurl/issues/60))
- [x] Navigation Jump mode ([#61](https://github.com/kbrdn1/LazyCurl/issues/61))

### v1.2.0 - External Tools (Sprint 3b) ✅

Intégration avec éditeurs externes et specs API.

- [x] Intégration éditeur externe ([#65](https://github.com/kbrdn1/LazyCurl/issues/65))
- [x] Import OpenAPI 3.x avec security schemes ([#66](https://github.com/kbrdn1/LazyCurl/issues/66), [#71](https://github.com/kbrdn1/LazyCurl/issues/71))
- [x] Import collection/environment Postman ([#14](https://github.com/kbrdn1/LazyCurl/issues/14), [#72](https://github.com/kbrdn1/LazyCurl/issues/72))

### v1.3.0 - JavaScript Scripting (Sprint 4a) ✅

Moteur de scripting JavaScript avec Goja runtime.

- [x] JavaScript Scripting via Goja ES5.1+ ([#35](https://github.com/kbrdn1/LazyCurl/issues/35), [#75](https://github.com/kbrdn1/LazyCurl/pull/75))
- [x] Test Assertions avec 16 matchers (`lc.test`, `lc.expect`)
- [x] Basic chaining via `lc.sendRequest()` (scripting-based)
- [x] API Documentation complète (75+ méthodes documentées)

---

## Current Sprint

### Sprint 4b - Collection Runner & Chaining 🔥

**Objectif** : Compléter la parité Bruno avec runner, chaining complet et améliorations UX.

#### Critical Priority 🔴

| Feature | Issue | Description | Concurrent |
|---------|-------|-------------|------------|
| Collection Runner | [#44](https://github.com/kbrdn1/LazyCurl/issues/44) | Exécution séquentielle de toutes les requêtes | Bruno, Yaak |
| Request Chaining | [#42](https://github.com/kbrdn1/LazyCurl/issues/42) | JSONPath/Regex extraction, chain definition UI | Bruno |

#### High Priority 🟡

| Feature | Issue | Description | Concurrent |
|---------|-------|-------------|------------|
| Fuzzy Finder | [#45](https://github.com/kbrdn1/LazyCurl/issues/45) | Recherche rapide style fzf | UX improvement |

#### Medium Priority 🟢

| Feature | Issue | Description |
|---------|-------|-------------|
| Request Diff | [#46](https://github.com/kbrdn1/LazyCurl/issues/46) | Comparaison de deux réponses |
| Request Templates | [#47](https://github.com/kbrdn1/LazyCurl/issues/47) | Patterns de requêtes réutilisables |
| Hot Reload Config | [#41](https://github.com/kbrdn1/LazyCurl/issues/41) | Rechargement auto sur changement config |

---

## Future Sprints

### Sprint 5 - Multi-Protocol (Parité Yaak)

**Objectif** : Supporter tous les protocoles modernes comme Yaak.

| Feature | Issue | Priority | Description | Gap vs Yaak |
|---------|-------|----------|-------------|-------------|
| GraphQL Support | [#18](https://github.com/kbrdn1/LazyCurl/issues/18) | 🔴 Critical | Schema explorer, variables, queries | ✅ Requis |
| WebSocket Testing | [#19](https://github.com/kbrdn1/LazyCurl/issues/19) | 🔴 Critical | Client WS interactif | ✅ Requis |
| SSE Support | [#48](https://github.com/kbrdn1/LazyCurl/issues/48) | 🟡 High | Viewer Server-Sent Events | ✅ Requis |
| gRPC Support | [#20](https://github.com/kbrdn1/LazyCurl/issues/20) | 🟡 High | Proto reflection, streaming | ✅ Requis |

### Sprint 6 - Enterprise Features

**Objectif** : Features pour environnements professionnels et entreprise.

| Feature | Issue | Priority | Description |
|---------|-------|----------|-------------|
| OAuth2 Flows | [#17](https://github.com/kbrdn1/LazyCurl/issues/17) | 🔴 Critical | Auth code, client credentials, refresh |
| AWS Signature v4 | [#17](https://github.com/kbrdn1/LazyCurl/issues/17) | 🟡 High | Authentification API AWS |
| mTLS / Client Certs | [#49](https://github.com/kbrdn1/LazyCurl/issues/49) | 🟡 High | Authentification TLS mutuelle |
| Proxy Support | [#50](https://github.com/kbrdn1/LazyCurl/issues/50) | 🟡 High | Proxy HTTP/SOCKS |
| Request Retry | [#51](https://github.com/kbrdn1/LazyCurl/issues/51) | 🟢 Medium | Auto-retry avec backoff |
| Rate Limiting | [#52](https://github.com/kbrdn1/LazyCurl/issues/52) | 🟢 Medium | Respect des limites API |

### Sprint 7 - CLI & Automation

**Objectif** : Mode headless et intégration CI/CD.

| Feature | Issue | Priority | Description |
|---------|-------|----------|-------------|
| CLI Mode | [#26](https://github.com/kbrdn1/LazyCurl/issues/26) | 🔴 Critical | `lazycurl run collection.json` |
| CI/CD Integration | [#53](https://github.com/kbrdn1/LazyCurl/issues/53) | 🔴 Critical | Exit codes, JSON output |
| Request Export | [#54](https://github.com/kbrdn1/LazyCurl/issues/54) | 🟡 High | Export vers Go, Python, JS |
| Mock Server | [#55](https://github.com/kbrdn1/LazyCurl/issues/55) | 🟢 Medium | Mock local depuis collection |
| API Docs Generator | [#56](https://github.com/kbrdn1/LazyCurl/issues/56) | 🟢 Medium | Génération docs depuis collection |

---

## Backlog

Features non encore planifiées :

| Feature | Issue | Description |
|---------|-------|-------------|
| Animated Dashboard | [#36](https://github.com/kbrdn1/LazyCurl/issues/36) | Sélecteur workspace avec splash screen |
| Settings Panel | [#25](https://github.com/kbrdn1/LazyCurl/issues/25) | UI de configuration in-app |
| Theme System | [#12](https://github.com/kbrdn1/LazyCurl/issues/12), [#13](https://github.com/kbrdn1/LazyCurl/issues/13) | Refactoring et UI des thèmes |

---

## Competitive Analysis

### vs posting (Concurrent TUI direct)

| Critère | posting | LazyCurl | Avantage |
|---------|---------|----------|----------|
| Langage | Python | Go | **LazyCurl** (perf) |
| Startup | ~500ms | ~50ms | **LazyCurl** |
| Jump mode | ❌ | ✅ | **LazyCurl** |
| OpenAPI import | ❌ | ✅ | **LazyCurl** |
| WebSocket | ✅ | ❌ (Sprint 5) | posting |
| SSH tunneling | ✅ | ❌ | posting |

### vs Bruno (Concurrent philosophique)

| Critère | Bruno | LazyCurl | Gap |
|---------|-------|----------|-----|
| Scripting JS | ✅ | ✅ | ✅ Parité |
| Test Assertions | ✅ | ✅ | ✅ Parité |
| Request Chaining | ✅ | ⚠️ Basic | [#42](https://github.com/kbrdn1/LazyCurl/issues/42) Sprint 4b |
| Collection Runner | ✅ | ❌ | [#44](https://github.com/kbrdn1/LazyCurl/issues/44) Sprint 4b |
| GraphQL | ✅ | ❌ | Sprint 5 |
| CLI mode | ✅ | ❌ | Sprint 7 |
| Git-friendly | ✅ | ✅ | ✅ Parité |

### vs Yaak (Référence multi-protocol)

| Protocol | Yaak | LazyCurl | Gap |
|----------|------|----------|-----|
| REST | ✅ | ✅ | ✅ Parité |
| GraphQL | ✅ | ❌ | Sprint 5 |
| WebSocket | ✅ | ❌ | Sprint 5 |
| gRPC | ✅ | ❌ | Sprint 5 |
| SSE | ✅ | ❌ | Sprint 5 |

---

## Timeline

```
2026 Q1: Sprint 4 - Parité Bruno
         ├── v1.3.0 ✅ Scripting + Assertions + Basic Chaining
         └── v1.4.0 🔄 Collection Runner + Full Chaining + UX

2026 Q2: Sprint 5 - Multi-Protocol
         └── GraphQL + WebSocket + SSE + gRPC

2026 Q3: Sprint 6 - Enterprise
         └── OAuth2 + mTLS + Proxy

2026 Q4: Sprint 7 - CI/CD
         └── CLI Mode + Automation
```

---

## Priority Legend

| Symbol | Priority | Description |
|--------|----------|-------------|
| 🔴 | Critical | Fonctionnalité core, bloque d'autres features |
| 🟡 | High | Important pour l'expérience utilisateur |
| 🟢 | Medium | Nice to have, qualité de vie |

---

## Success Metrics

| Milestone | Target | Status |
|-----------|--------|--------|
| Parité posting | v1.3.0 | ✅ Complete |
| Scripting & Assertions | v1.3.0 | ✅ Complete |
| Parité Bruno (core) | v1.4.0 | 🔄 In Progress (Runner + Chaining pending) |
| Parité Yaak protocols | v1.5.0 | ⏳ Sprint 5 |
| Enterprise-ready | v1.6.0 | ⏳ Sprint 6 |
| CI/CD complete | v2.0.0 | ⏳ Sprint 7 |

---

## Contributing

Want to help? Check out:

1. [Contributing Guide](CONTRIBUTING.md)
2. [Good First Issues](https://github.com/kbrdn1/LazyCurl/labels/good%20first%20issue)
3. [Help Wanted](https://github.com/kbrdn1/LazyCurl/labels/help%20wanted)

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for release history.
