# Plan de Développement - LazyCurl
## TUI HTTP Client (style Lazygit + fonctionnalités Postman)

Ce document détaille le plan de développement complet pour faire de LazyCurl un client HTTP TUI avec une interface inspirée de Lazygit.

---

## Phase 1: Fondations ✅ COMPLÉTÉ

### 1.1 Configuration du projet ✅
- [x] Initialiser le module Go
- [x] Installer les dépendances (Bubble Tea, Bubbles, Lipgloss, Bubble Zone, YAML)
- [x] Créer l'architecture de base
- [x] Point d'entrée principal

### 1.2 Interface de base Lazygit-style ✅
- [x] Layout multi-panneaux (Collections, Request, Response)
- [x] Navigation vim motions (h/j/k/l)
- [x] Styles Lipgloss personnalisés
- [x] Panneaux actifs/inactifs visuels

### 1.3 Système de configuration ✅
- [x] Configuration YAML (globale + workspace)
- [x] Keybindings personnalisables
- [x] Thèmes configurables
- [x] Système de workspace (`.lazycurl/`)

---

## Phase 2: Collections et Workspace 🔥 PRIORITAIRE

### 2.1 Gestion des fichiers de collections
**Fichiers**: `internal/api/collection.go`, `internal/ui/collections_view.go`

**Tâches**:
- [ ] Définir le format JSON des collections
- [ ] Implémenter le chargement des collections depuis `.lazycurl/collections/`
- [ ] Afficher les collections dans le panel gauche
- [ ] Navigation dans l'arborescence (dossiers/requêtes)
- [ ] Expand/collapse des collections
- [ ] Support des icônes HTTP methods (GET, POST, etc.)

**Format de collection**:
```json
{
  "name": "My API",
  "description": "API description",
  "folders": [
    {
      "name": "Users",
      "requests": [...]
    }
  ],
  "requests": [
    {
      "id": "unique-id",
      "name": "Get Users",
      "method": "GET",
      "url": "{{base_url}}/users",
      "headers": {...},
      "body": null,
      "tests": []
    }
  ]
}
```

### 2.2 Sauvegarde et création de collections
**Fichiers**: `internal/api/collection.go`

**Tâches**:
- [ ] Créer une nouvelle collection (touche `n` dans Collections panel)
- [ ] Sauvegarder automatiquement les modifications
- [ ] Ajouter une requête à une collection
- [ ] Créer des dossiers dans les collections
- [ ] Renommer collections/dossiers
- [ ] Supprimer collections/requêtes (touche `d`)
- [ ] Validation du JSON

### 2.3 Import/Export Postman
**Fichiers**: `internal/api/postman.go` (nouveau)

**Tâches**:
- [ ] Parser les collections Postman v2.1
- [ ] Convertir vers le format LazyCurl
- [ ] Exporter depuis LazyCurl vers Postman
- [ ] Support des variables Postman
- [ ] Migration des pre-request scripts

---

## Phase 3: Request Builder (Panel 2) 🔥 PRIORITAIRE

### 3.1 Éditeur de requêtes interactif
**Fichiers**: `internal/ui/request_view.go`, `internal/ui/request_editor.go` (nouveau)

**Tâches**:
- [ ] Mode édition avec `textarea` de Bubbles
- [ ] Sélecteur de méthode HTTP (GET, POST, PUT, etc.)
- [ ] Éditeur d'URL avec autocomplétion
- [ ] Sub-panels: Headers, Body, Query Params, Auth
- [ ] Navigation entre sub-panels (Tab)
- [ ] Validation en temps réel

**Composants Bubbles nécessaires**:
- `github.com/charmbracelet/bubbles/textinput` - URL, headers
- `github.com/charmbracelet/bubbles/textarea` - Body editor
- `github.com/charmbracelet/bubbles/list` - Method selector

### 3.2 Gestion des Headers
**Fichiers**: `internal/ui/headers_editor.go` (nouveau)

**Tâches**:
- [ ] Liste éditable des headers (key/value)
- [ ] Ajouter/supprimer des headers
- [ ] Headers suggérés courants (Content-Type, Authorization, etc.)
- [ ] Toggle enable/disable par header
- [ ] Bulk edit mode

### 3.3 Body Editor
**Fichiers**: `internal/ui/body_editor.go` (nouveau)

**Tâches**:
- [ ] Support de différents formats (JSON, XML, Raw, Form-data)
- [ ] Éditeur JSON avec validation
- [ ] Pretty print JSON automatique
- [ ] Syntax highlighting (si possible avec Lipgloss)
- [ ] Support des fichiers (form-data avec upload)

### 3.4 Query Parameters
**Tâches**:
- [ ] Éditeur de query params (key/value)
- [ ] Génération automatique depuis l'URL
- [ ] URL encoding automatique
- [ ] Bulk edit mode

### 3.5 Authentication
**Fichiers**: `internal/api/auth.go` (nouveau)

**Tâches**:
- [ ] Support Basic Auth
- [ ] Support Bearer Token
- [ ] Support API Key (header/query)
- [ ] Support OAuth 2.0 (optionnel phase ultérieure)
- [ ] Stockage sécurisé dans environnements

---

## Phase 4: HTTP Client et Response Viewer 🔥 PRIORITAIRE

### 4.1 Client HTTP fonctionnel
**Fichiers**: `internal/api/http.go`

**Tâches**:
- [ ] Envoyer des requêtes HTTP réelles
- [ ] Support de tous les verbes HTTP
- [ ] Gestion des timeouts
- [ ] Gestion des redirects
- [ ] Support HTTPS avec certificats
- [ ] Proxy support (optionnel)
- [ ] Progress indicator pendant la requête

### 4.2 Response Viewer (Panel 3)
**Fichiers**: `internal/ui/response_view.go`

**Tâches**:
- [ ] Afficher le status code avec coloration (2xx vert, 4xx orange, 5xx rouge)
- [ ] Afficher temps de réponse et taille
- [ ] Sub-panels: Body, Headers, Cookies
- [ ] Body viewer avec `viewport` de Bubbles
- [ ] Formatage JSON automatique
- [ ] Syntax highlighting JSON
- [ ] Copier la réponse dans le clipboard (optionnel)
- [ ] Recherche dans la réponse (/)

**Composants Bubbles**:
- `github.com/charmbracelet/bubbles/viewport` - Scroll du body

### 4.3 Pretty Printing et Formatting
**Fichiers**: `internal/format/formatter.go` (nouveau)

**Tâches**:
- [ ] JSON formatter avec indentation
- [ ] XML formatter
- [ ] HTML formatter (basique)
- [ ] Détection automatique du format
- [ ] Coloration syntaxique basique

---

## Phase 5: Environnements et Variables

### 5.1 Gestion des environnements
**Fichiers**: `internal/api/environment.go`, `internal/ui/environments_view.go`

**Tâches**:
- [ ] Charger les environnements depuis `.lazycurl/envs/*.json`
- [ ] Charger les environnements globaux depuis `~/.config/lazycurl/`
- [ ] Afficher la liste dans le panel Environments
- [ ] Sélectionner un environnement actif
- [ ] Éditer les variables d'environnement
- [ ] Créer/supprimer des environnements
- [ ] Importer/exporter des environnements

**Format d'environnement**:
```json
{
  "name": "Development",
  "description": "Local dev environment",
  "variables": {
    "base_url": "http://localhost:3000",
    "api_key": "dev_key_123",
    "username": "admin"
  }
}
```

### 5.2 Substitution de variables
**Fichiers**: `internal/api/variables.go` (nouveau)

**Tâches**:
- [ ] Parser les variables `{{variable}}` dans les URLs
- [ ] Parser les variables dans les headers
- [ ] Parser les variables dans le body
- [ ] Substitution au moment de l'envoi
- [ ] Afficher les variables non résolues
- [ ] Support des variables imbriquées
- [ ] Variables système (date, timestamp, random, uuid)

**Variables système**:
```
{{$timestamp}}  - Unix timestamp
{{$datetime}}   - ISO datetime
{{$randomInt}}  - Random integer
{{$uuid}}       - UUID v4
{{$guid}}       - GUID
```

### 5.3 Éditeur d'environnements
**Fichiers**: `internal/ui/env_editor.go` (nouveau)

**Tâches**:
- [ ] Créer un nouvel environnement
- [ ] Éditer les variables (key/value)
- [ ] Dupliquer un environnement
- [ ] Exporter vers JSON
- [ ] Import depuis JSON

---

## Phase 6: Historique et Sessions

### 6.1 Historique des requêtes
**Fichiers**: `internal/api/history.go` (nouveau), `.lazycurl/history.json`

**Tâches**:
- [ ] Sauvegarder automatiquement chaque requête envoyée
- [ ] Stocker dans `.lazycurl/history.json`
- [ ] Panel d'historique (accessible avec un raccourci)
- [ ] Filtrer l'historique par méthode, URL, date
- [ ] Rejouer une requête depuis l'historique
- [ ] Sauvegarder une requête de l'historique vers une collection
- [ ] Effacer l'historique (tout ou sélection)
- [ ] Limite configurable de l'historique

**Format d'historique**:
```json
{
  "requests": [
    {
      "timestamp": "2025-10-23T22:00:00Z",
      "request": {...},
      "response": {...},
      "duration_ms": 142
    }
  ]
}
```

### 6.2 Sessions de travail
**Fichiers**: `.lazycurl/session.json`

**Tâches**:
- [ ] Sauvegarder l'état actuel (requête en cours, panel actif)
- [ ] Restaurer la session au redémarrage
- [ ] Multiple sessions nommées
- [ ] Switch entre sessions

---

## Phase 7: Fonctionnalités avancées

### 7.1 Tests et Assertions
**Fichiers**: `internal/api/tests.go` (nouveau)

**Tâches**:
- [ ] Définir des tests pour les requêtes (JSON)
- [ ] Assertions sur le status code
- [ ] Assertions sur les headers
- [ ] Assertions sur le body (JSONPath)
- [ ] Exécuter les tests automatiquement
- [ ] Afficher les résultats des tests
- [ ] Scripts pre-request et post-response (optionnel)

**Format de test**:
```json
{
  "tests": [
    {
      "name": "Status is 200",
      "assert": "response.status == 200"
    },
    {
      "name": "Has users array",
      "assert": "response.body.users != null"
    }
  ]
}
```

### 7.2 Chaînage de requêtes
**Fichiers**: `internal/api/chain.go` (nouveau)

**Tâches**:
- [ ] Extraire des données de la réponse
- [ ] Utiliser dans la requête suivante
- [ ] Variables de session (scope request)
- [ ] Workflows de requêtes

### 7.3 Mock Server (optionnel)
**Fichiers**: `internal/mock/server.go` (nouveau)

**Tâches**:
- [ ] Créer des mocks de réponses
- [ ] Serveur HTTP local pour les mocks
- [ ] Utile pour les tests

### 7.4 Recherche globale
**Fichiers**: `internal/ui/search.go` (nouveau)

**Tâches**:
- [ ] Recherche fuzzy dans les collections (/)
- [ ] Recherche dans les requêtes par nom/URL
- [ ] Recherche dans l'historique
- [ ] Navigation rapide vers les résultats

### 7.5 Documentation et Notes
**Tâches**:
- [ ] Ajouter des descriptions aux requêtes
- [ ] Markdown support pour les descriptions
- [ ] Documentation au niveau collection
- [ ] Export de documentation

---

## Phase 8: Thèmes et Personnalisation

### 8.1 Système de thèmes
**Fichiers**: `pkg/styles/themes.go` (nouveau)

**Tâches**:
- [ ] Thèmes prédéfinis (dark, light, dracula, gruvbox, etc.)
- [ ] Charger depuis la config YAML
- [ ] Preview des thèmes
- [ ] Thèmes personnalisés par l'utilisateur
- [ ] Export/import de thèmes

### 8.2 Layout personnalisable
**Tâches**:
- [ ] Configurer la largeur des panneaux
- [ ] Toggle visibilité des panneaux
- [ ] Layouts alternatifs (vertical split, etc.)

---

## Phase 9: Performance et Optimisation

### 9.1 Performance
**Tâches**:
- [ ] Lazy loading des collections
- [ ] Virtual scrolling pour grandes listes
- [ ] Cache des réponses (optionnel)
- [ ] Optimisation du rendu Lipgloss

### 9.2 Gestion des gros fichiers
**Tâches**:
- [ ] Streaming des grandes réponses
- [ ] Pagination dans le response viewer
- [ ] Limite de taille configurable

---

## Phase 10: Tests et Documentation

### 10.1 Tests unitaires
**Tâches**:
- [ ] Tests pour le client HTTP
- [ ] Tests pour les collections
- [ ] Tests pour les environnements
- [ ] Tests pour la configuration
- [ ] Coverage > 80%

### 10.2 Tests d'intégration
**Tâches**:
- [ ] Tests E2E avec mock server
- [ ] Tests des flows utilisateur
- [ ] Tests de performance

### 10.3 Documentation
**Tâches**:
- [ ] Documentation GoDoc complète
- [ ] Guide utilisateur détaillé
- [ ] Tutoriels et exemples
- [ ] FAQ
- [ ] Vidéos de démo

---

## Phase 11: Distribution

### 11.1 Packaging
**Tâches**:
- [ ] Binaires multi-platform (Linux, macOS, Windows)
- [ ] Homebrew formula
- [ ] Snap package
- [ ] AUR package (Arch)
- [ ] Chocolatey (Windows)
- [ ] Docker image (optionnel)

### 11.2 CI/CD
**Tâches**:
- [ ] GitHub Actions pour tests
- [ ] GitHub Actions pour releases
- [ ] Versioning sémantique automatique
- [ ] Changelog automatique

---

## Priorités de développement - Sprints

### Sprint 1 (2-3 semaines) - MVP 🔥
**Objectif**: Application fonctionnelle de base

1. **Collections**: Charger et afficher les collections JSON
2. **Request Builder**: Éditer méthode, URL, headers, body
3. **HTTP Client**: Envoyer des requêtes réelles
4. **Response Viewer**: Afficher status, headers, body formaté JSON
5. **Sauvegarde**: Créer et sauvegarder des requêtes

**Deliverable**: Pouvoir créer une collection, ajouter des requêtes, les envoyer et voir les réponses

### Sprint 2 (2-3 semaines) - Environnements
**Objectif**: Variables et environnements fonctionnels

1. **Environnements**: Charger, éditer, sélectionner
2. **Variables**: Substitution dans URL, headers, body
3. **Panel Environments**: Interface complète
4. **Import Postman**: Support basique
5. **Historique**: Sauvegarde et visualisation

**Deliverable**: Utiliser des variables d'environnement et importer des collections Postman

### Sprint 3 (2-3 semaines) - Polish et UX
**Objectif**: Améliorer l'expérience utilisateur

1. **Recherche**: Recherche fuzzy dans collections
2. **Thèmes**: Multiple thèmes prédéfinis
3. **Sessions**: Sauvegarder/restaurer l'état
4. **Tests**: Assertions basiques sur les réponses
5. **Documentation**: README complet et exemples

**Deliverable**: Application polie avec bonne UX et documentation

### Sprint 4 (2-3 semaines) - Fonctionnalités avancées
**Objectif**: Features pro

1. **Tests avancés**: JSONPath assertions
2. **Chaînage**: Variables de session entre requêtes
3. **Auth avancée**: OAuth 2.0
4. **Performance**: Optimisations
5. **Export**: Documentation auto-générée

**Deliverable**: Features avancées type Postman Pro

### Sprint 5 (1-2 semaines) - Release
**Objectif**: Publication v1.0

1. **Tests**: Coverage complète
2. **Documentation**: Complète
3. **Packaging**: Tous les formats
4. **CI/CD**: Automatisation complète
5. **Marketing**: Site web, vidéos

**Deliverable**: Release publique v1.0

---

## Dépendances supplémentaires à considérer

```bash
# Syntax highlighting
go get github.com/alecthomas/chroma

# JSON parsing avancé
go get github.com/tidwall/gjson

# JSONPath pour tests
go get github.com/PaesslerAG/jsonpath

# Fuzzy search
go get github.com/sahilm/fuzzy

# Clipboard support
go get github.com/atotto/clipboard
```

---

## Métriques de succès

- ✅ Application compile et lance sans erreur
- ✅ Interface réactive < 100ms pour toutes les actions
- ✅ Supporté toutes les méthodes HTTP
- ✅ Import Postman collections sans perte de données
- ✅ Documentation claire et complète
- ✅ Coverage tests > 80%
- ✅ 0 bugs critiques

---

## Ressources

- [Bubble Tea Docs](https://github.com/charmbracelet/bubbletea)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Lazygit Source](https://github.com/jesseduffield/lazygit) - Pour inspiration UI
- [Postman API Format](https://schema.postman.com/) - Pour compatibilité
- [HTTP/1.1 Spec](https://tools.ietf.org/html/rfc2616)

---

## Notes importantes

### Architecture
- Séparer logique métier de l'UI
- Interfaces pour faciliter les tests
- Pas de dépendances circulaires

### UX Lazygit-style
- Tout accessible au clavier
- Vim motions partout
- Feedback visuel immédiat
- Pas de dialogue modale (inline editing)

### Performance
- Lazy loading
- Virtual scrolling
- Debouncing
- Profiling régulier

---

**Prochaine étape**: Commencer le Sprint 1 - MVP avec les collections et le request builder basique.
