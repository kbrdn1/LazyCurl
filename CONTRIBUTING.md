# Contributing to LazyCurl 🤝

Thank you for your interest in contributing to LazyCurl! This document provides guidelines and conventions to follow when contributing to this project.

## Table of Contents

- [Getting Started](#getting-started)
- [Working with Claude Code (Parallel Features)](#working-with-claude-code-parallel-features)
- [Development Workflow](#development-workflow)
- [Branch Convention](#branch-convention)
- [Commit Convention](#commit-convention)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [Testing](#testing)
- [Documentation](#documentation)
- [GitHub Labels](#github-labels)
- [Release Process](#release-process)

---

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- A GitHub account

### Setup Development Environment

1. Fork the repository
2. Clone your fork:

   ```bash
   git clone https://github.com/YOUR_USERNAME/LazyCurl.git
   cd LazyCurl
   ```

3. Add upstream remote:

   ```bash
   git remote add upstream https://github.com/kbrdn1/LazyCurl.git
   ```

4. Install dependencies:

   ```bash
   go mod download
   ```

5. Build the project:

   ```bash
   make build
   ```

6. Run the application:

   ```bash
   make run
   ```

---

## Working with Claude Code (Parallel Features)

When working on multiple features simultaneously with Claude Code, use **Git Worktrees** to maintain isolated working directories. This prevents context switching and keeps each Claude Code session focused on its specific task.

### Why Git Worktrees?

- Each worktree has its own working directory with isolated files while sharing the same Git history
- Claude Code maintains deep context understanding for each feature without pollution from other work
- No need for `git stash` or constant branch switching
- Parallel development without merge conflicts with yourself

### gwq - Git Worktree Manager

We use [gwq](https://github.com/d-kuro/gwq) for efficient worktree management with fuzzy finder integration.

**Installation:**

```bash
# Via Homebrew (macOS/Linux)
brew install d-kuro/tap/gwq

# Via Go
go install github.com/d-kuro/gwq/cmd/gwq@latest
```

### gwq Commands Reference

| Command | Description |
|---------|-------------|
| `gwq add -b <branch>` | Create worktree with new branch |
| `gwq add -i` | Interactive worktree creation with fuzzy finder |
| `gwq list` | List all worktrees |
| `gwq list -v` | List with verbose info (uncommitted changes, etc.) |
| `gwq get <pattern>` | Get worktree path (for `cd $(gwq get feat)`) |
| `gwq cd <pattern>` | Change to worktree directory (launches new shell) |
| `gwq exec <pattern> -- <cmd>` | Execute command in worktree |
| `gwq remove <branch>` | Remove worktree |
| `gwq remove -b <branch>` | Remove worktree AND delete branch |
| `gwq status` | Show status of all worktrees |
| `gwq status --watch` | Monitor worktrees in real-time |
| `gwq prune` | Clean up stale worktree references |

### Quick Create Examples

```bash
# Create a feature worktree
gwq add -b feat/#123-user-authentication

# Create a bugfix worktree
gwq add -b fix/#456-http-timeout

# Create a hotfix worktree
gwq add -b hotfix/#789-critical-security-fix

# Interactive creation with fuzzy finder
gwq add -i

# Use the interactive manager
make worktree
```

**Branch naming convention**: `<type>/#<issue>-<description>`

**Branch types available**: `feat`, `fix`, `hotfix`, `docs`, `test`, `refactor`, `chore`, `perf`, `ci`, `build`

### Navigation & Execution

```bash
# Navigate to a worktree (fuzzy match)
cd $(gwq get authentication)
gwq cd feat                          # Opens new shell in matching worktree

# Execute commands in worktrees
gwq exec authentication -- make build
gwq exec -s feat -- make test         # Stay in worktree after execution

# Monitor all worktrees
gwq status --watch
```

### Running Claude Code Sessions

```bash
# Terminal 1 - Working on authentication
cd $(gwq get authentication)
claude

# Terminal 2 - Working on API refactor
cd $(gwq get refactor)
claude

# Terminal 3 - Fixing bug
cd $(gwq get fix)
claude
```

Or using `gwq cd` (launches new shell):

```bash
gwq cd authentication && claude
```

### Cleanup

```bash
# Remove a specific worktree
gwq remove feat/#123-user-authentication

# Remove worktree AND delete the branch
gwq remove -b feat/#123-completed-feature

# Dry run to preview what would be removed
gwq remove --dry-run feat/#123-old-feature

# Clean up stale worktree references
gwq prune
```

### Configuration (Optional)

Create `~/.config/gwq/config.toml` or `.gwq.toml` in the project root:

```toml
[worktree]
basedir = "~/worktrees"

[[repository_settings]]
repository = "~/Projects/Perso/LazyCurl"
setup_commands = ["make deps"]
```

This automatically runs `make deps` when creating worktrees.

### Best Practices

1. **Use gwq**: Prefer `gwq add -b` over manual `git worktree add` for consistency
2. **Bootstrap Each Worktree**: Run `make deps` in each new worktree (or configure in `.gwq.toml`)
3. **Keep Worktrees Updated**: Regularly merge `main` into feature branches to avoid large conflicts
4. **Clean Up**: Use `gwq remove -b` after merging to remove worktrees AND branches
5. **Monitor Status**: Use `gwq status --watch` to track changes across all worktrees

---

## Development Workflow

1. **Sync with upstream**:

   ```bash
   git checkout main
   git pull upstream main
   ```

2. **Create a new branch** following the [Branch Convention](#branch-convention)

3. **Make your changes** following the [Code Style](#code-style)

4. **Test your changes** (see [Testing](#testing))

5. **Commit your changes** following the [Commit Convention](#commit-convention)

6. **Push to your fork** and create a Pull Request

---

## Branch Convention 🌿

Main branches:

- `main`: Production-ready code
- `dev`: Development branch (currently not used, all development on feature branches)

### Naming Convention 📛

```bash
<type>/#<issue-number>-<short-description>
```

**Components**:

- `type`: Type of the branch (see types below)
- `issue-number`: Related GitHub issue number
- `short-description`: Brief description in kebab-case

### Branch Types

- `feat` or `feature`: New feature implementation
- `fix`: Bug fix
- `hotfix`: Critical bug fix in production
- `docs`: Documentation changes
- `test`: Adding or modifying tests
- `refactor`: Code refactoring
- `chore`: Maintenance tasks
- `ci`: CI/CD configuration changes
- `build`: Build system changes
- `perf`: Performance improvements

### Examples

- `feat/#12-add-collection-loader`
- `fix/#25-fix-yaml-parsing`
- `docs/#8-update-contributing-guide`
- `refactor/#33-reorganize-ui-components`
- `perf/#45-optimize-response-rendering`
- `test/#18-add-http-client-tests`

---

## Commit Convention 📝

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification with [Gitmoji](https://gitmoji.dev/) emojis.

### Format

```bash
<emoji> <type>(<scope>)<!>: <subject>
```

### Emojis

Use [Gitmoji](https://gitmoji.dev/) prefixes for commit messages:

| Emoji | Code | Description |
|-------|------|-------------|
| ✨ | `:sparkles:` | New feature |
| 🐛 | `:bug:` | Bug fix |
| 📝 | `:memo:` | Documentation |
| ♻️ | `:recycle:` | Refactor code |
| ⚡️ | `:zap:` | Performance |
| ✅ | `:white_check_mark:` | Tests |
| 🔧 | `:wrench:` | Configuration |
| 🚀 | `:rocket:` | Deployment |
| 🎨 | `:art:` | UI/Style |
| 🔥 | `:fire:` | Remove code/files |
| 🚑️ | `:ambulance:` | Critical hotfix |
| ⬆️ | `:arrow_up:` | Upgrade dependencies |
| ⬇️ | `:arrow_down:` | Downgrade dependencies |
| 🏗️ | `:building_construction:` | Architecture changes |

**Tip**: Install the [Gitmoji VSCode extension](https://marketplace.visualstudio.com/items?itemName=seatonjiang.gitmoji-vscode)

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `style`: Code style changes (formatting, missing semi-colons, etc.)
- `refactor`: Code refactoring (neither fixes a bug nor adds a feature)
- `perf`: Performance improvements
- `test`: Adding or correcting tests
- `chore`: Maintenance tasks (build, dependencies, etc.)
- `ci`: CI/CD changes
- `build`: Build system changes

### Scopes

Choose a scope based on the affected module:

- `ui`: User interface components
- `api`: HTTP client and API logic
- `config`: Configuration management
- `collections`: Collections management
- `environments`: Environment variables
- `styles`: Lipgloss styles
- `cli`: Command-line interface
- `docs`: Documentation
- `tests`: Test files

### Breaking Changes

Indicate breaking changes with `!` after the type/scope:

```bash
✨ feat(api)!: change collection file format to v2
```

### Subject Guidelines

Use imperative mood and follow these patterns:

| Verb | Use Case | Example |
|------|----------|---------|
| `add` | Create capability | `✨ feat(collections): add folder support` |
| `change` | Change behavior | `♻️ refactor(ui): change panel layout logic` |
| `remove` | Delete capability | `🔥 feat(api): remove deprecated methods` |
| `fix` | Fix issue | `🐛 fix(config): fix YAML parsing error` |
| `bump` | Increase version | `⬆️ chore(deps): bump bubbletea to v1.4.0` |
| `optimize` | Performance | `⚡️ perf(ui): optimize viewport rendering` |
| `refactor` | Restructure | `♻️ refactor(api): refactor HTTP client` |
| `update` | Update code | `🔧 chore(config): update default theme colors` |
| `improve` | Enhance code | `✨ feat(ui): improve keyboard navigation` |
| `disable` | Disable code | `🔒 chore(api): disable experimental feature` |

**Rules**:

- Don't capitalize first letter
- No period (.) at the end
- Keep it under 72 characters

### Commit Examples

```bash
✨ feat(collections): add JSON collection loader
🐛 fix(ui): fix panel resize on terminal size change
📝 docs: update installation instructions
♻️ refactor(api): refactor request builder logic
⚡️ perf(ui): optimize large collection rendering
✅ test(api): add HTTP client unit tests
🔧 chore(config): update default keybindings
🚀 ci: add GitHub Actions workflow
🎨 style(ui): improve response viewer colors
🔥 feat(api)!: remove legacy request format
```

---

## Pull Request Process

### Before Creating a PR

1. ✅ Ensure your code compiles: `make build`
2. ✅ Run tests: `make test` (when available)
3. ✅ Format your code: `make fmt`
4. ✅ Update documentation if needed
5. ✅ Ensure your branch is up to date with `main`

### PR Title

Use the same format as commit messages:

```bash
<emoji> <type>(<scope>): <description>
```

Example: `✨ feat(collections): add Postman import support`

### PR Description Template

```markdown
## Description
Brief description of the changes

## Related Issue
Fixes #<issue-number>

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Changes Made
- Change 1
- Change 2
- Change 3

## Testing
Describe how you tested your changes

## Screenshots (if applicable)
Add screenshots for UI changes

## Checklist
- [ ] My code follows the project's code style
- [ ] I have performed a self-review of my code
- [ ] I have commented my code where necessary
- [ ] I have updated the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix/feature works
- [ ] New and existing tests pass locally
```

### Review Process

- At least 1 approval is required
- All CI checks must pass
- Code must be up to date with `main` branch
- Resolve all review comments before merging

---

## Code Style

### Go Code Style

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

**Key points**:

- Use `gofmt` for formatting (automatically done with `make fmt`)
- Use meaningful variable and function names
- Keep functions small and focused
- Add comments for exported functions and types
- Use Go idioms and best practices

### File Organization

```
LazyCurl/
├── cmd/                   # Application entrypoints
│   └── lazycurl/
├── internal/              # Private application code
│   ├── api/              # HTTP client and API logic
│   ├── config/           # Configuration management
│   └── ui/               # TUI components
├── pkg/                   # Public libraries
│   └── styles/           # Lipgloss styles
├── docs/                  # Documentation
├── .github/              # GitHub configuration
└── scripts/              # Build and deployment scripts
```

### Naming Conventions

**Files**:

- Use snake_case: `collections_view.go`
- Test files: `collections_view_test.go`

**Functions/Methods**:

- Exported: `PascalCase` (e.g., `LoadCollection`)
- Private: `camelCase` (e.g., `parseJSON`)

**Constants**:

- Exported: `PascalCase` (e.g., `DefaultTimeout`)
- Private: `camelCase` (e.g., `maxRetries`)

**Variables**:

- Use descriptive names: `collectionPath`, `httpClient`
- Avoid single letters except for short scopes (i, j, k in loops)

### Comments

```go
// LoadCollection loads a collection from the specified path.
// It returns an error if the file doesn't exist or is invalid JSON.
func LoadCollection(path string) (*Collection, error) {
    // Implementation
}
```

---

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test ./internal/api/...
```

### Writing Tests

- Place test files next to the code they test
- Use table-driven tests when possible
- Test both success and error cases
- Mock external dependencies

**Example**:

```go
func TestLoadCollection(t *testing.T) {
    tests := []struct {
        name    string
        path    string
        want    *Collection
        wantErr bool
    }{
        {
            name:    "valid collection",
            path:    "testdata/valid.json",
            want:    &Collection{Name: "Test"},
            wantErr: false,
        },
        {
            name:    "invalid path",
            path:    "nonexistent.json",
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := LoadCollection(tt.path)
            if (err != nil) != tt.wantErr {
                t.Errorf("LoadCollection() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("LoadCollection() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

## Documentation

### Code Documentation

- Document all exported functions, types, and constants
- Use GoDoc format
- Include examples when helpful

### Project Documentation

- Update README.md for user-facing changes
- Update DEVELOPMENT_PLAN.md for roadmap changes
- Add examples in `docs/` directory

### Changelog

Update CHANGELOG.md following [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
## [Unreleased]
### Added
- New feature description

### Changed
- Changed feature description

### Fixed
- Bug fix description
```

---

## GitHub Labels

> **Note**: For complete label documentation including organization-level labels and implementation instructions, see [`.github/LABELS.md`](.github/LABELS.md).

### Type Labels

| Label | Color | Description |
|-------|-------|-------------|
| **feature** | `#0E8A16` | New feature implementation |
| **fix** | `#D73A4A` | Bug fix |
| **hotfix** | `#FF3333` | Critical production bug fix |
| **docs** | `#1D76DB` | Documentation changes |
| **test** | `#87CEEB` | Test additions or modifications |
| **refactor** | `#FBCA04` | Code restructuring |
| **chore** | `#808080` | Maintenance tasks |
| **optimization** | `#FFA500` | Performance improvements |

### Domain Labels

| Label | Color | Description |
|-------|-------|-------------|
| **ui/ux** | `#FF69B4` | User interface/experience |
| **api** | `#0075CA` | HTTP client and API logic |
| **collections** | `#7D56F4` | Collections management |
| **environments** | `#00D9FF` | Environment variables |
| **configuration** | `#26A69A` | Configuration system |
| **ci/cd** | `#26A69A` | CI/CD pipeline |
| **security** | `#B60205` | Security issues |

### Management Labels

| Label | Color | Description |
|-------|-------|-------------|
| **dependencies** | `#8B008B` | Dependency updates |
| **breaking** | `#FF0000` | Breaking changes |
| **good first issue** | `#7057ff` | Good for newcomers |
| **help wanted** | `#008672` | Extra attention needed |
| **urgent** | `#FF1493` | Requires immediate attention |

### Status Labels

| Label | Color | Description |
|-------|-------|-------------|
| **duplicate** | `#CCCCCC` | Duplicate issue/PR |
| **invalid** | `#444444` | Invalid issue |
| **wontfix** | `#FFFFFF` | Will not be fixed |

### Priority Levels

Use GitHub project boards or issue fields for priorities:

- **Critical**: Blocking issue, immediate resolution needed
- **High**: Important, resolve quickly
- **Medium**: Standard priority
- **Low**: Minor issue, can be deferred
- **Trivial**: Cosmetic improvements

---

## Release Process 🚀

Releases are managed using [Semantic Versioning](https://semver.org/):

### Versioning Format

```
vMAJOR.MINOR.PATCH
```

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Examples

- `v1.0.0` - Initial release
- `v1.1.0` - New feature added
- `v1.1.1` - Bug fix
- `v2.0.0` - Breaking changes

### Release Workflow

1. Update CHANGELOG.md
2. Create release branch: `release/vX.Y.Z`
3. Update version in code if applicable
4. Create PR to `main`
5. After merge, create GitHub release with tag `vX.Y.Z`
6. GitHub Actions will automatically build and publish binaries

---

## Community Guidelines

### Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Provide constructive feedback
- Focus on what is best for the community

### Getting Help

- 📖 Read the [documentation](README.md)
- 💬 Open a [Discussion](https://github.com/kbrdn1/LazyCurl/discussions)
- 🐛 Report bugs via [Issues](https://github.com/kbrdn1/LazyCurl/issues)
- 💡 Suggest features via [Issues](https://github.com/kbrdn1/LazyCurl/issues)

---

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Bubble Tea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Gitmoji Guide](https://gitmoji.dev/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)

---

## Thank You! 🙏

Thank you for contributing to LazyCurl! Your contributions help make this project better for everyone.

---

<p align="center">
 Copyright &copy; 2024-present <a href="https://github.com/kbrdn1" target="_blank">@kbrdn1</a>
</p>

<p align="center">
 <a href="https://github.com/kbrdn1/LazyCurl/blob/main/LICENSE"><img src="https://img.shields.io/static/v1.svg?style=for-the-badge&label=License&message=MIT&logoColor=d9e0ee&colorA=363a4f&colorB=b7bdf8" alt="MIT License"/></a>
</p>
