# Release Process

This document describes the release workflow for LazyCurl.

## Overview

LazyCurl uses [Release Please](https://github.com/googleapis/release-please) to automate the release process:

1. **Automatic PR**: When commits are pushed to `main`, Release Please creates/updates a "Release PR"
2. **Changelog Generation**: The PR includes auto-generated changelog based on conventional commits
3. **Version Bump**: Merging the Release PR creates a GitHub release and git tag
4. **GoReleaser**: The tag triggers GoReleaser to build binaries and publish to Homebrew

## Workflow Diagram

```text
Push to main → Release Please PR created/updated
                         ↓
              Merge Release PR
                         ↓
              GitHub Release + Tag created
                         ↓
              GoReleaser builds binaries
                         ↓
              Homebrew formula updated
```

## GitHub Configuration Required

### 1. Create Homebrew Tap Repository

Create a new repository: `kbrdn1/homebrew-tap`

```bash
# On GitHub, create a new public repository named "homebrew-tap"
# No need to initialize with README
```

### 2. Create Personal Access Token (PAT)

1. Go to **GitHub Settings** → **Developer settings** → **Personal access tokens** → **Fine-grained tokens**
2. Click **Generate new token**
3. Configure:
   - **Token name**: `HOMEBREW_TAP_TOKEN`
   - **Expiration**: 1 year (or custom)
   - **Repository access**: Select `homebrew-tap` repository only
   - **Permissions**:
     - **Contents**: Read and write
     - **Metadata**: Read-only
4. Copy the token

### 3. Add Repository Secret

1. Go to **LazyCurl repository** → **Settings** → **Secrets and variables** → **Actions**
2. Click **New repository secret**
3. Add:
   - **Name**: `HOMEBREW_TAP_TOKEN`
   - **Value**: Paste the PAT from step 2

## Commit Convention

Release Please uses [Conventional Commits](https://www.conventionalcommits.org/) to determine version bumps:

| Commit Type | Version Bump | Example |
|-------------|--------------|---------|
| `feat:` | Minor (0.x.0) | `feat(ui): add dark mode toggle` |
| `fix:` | Patch (0.0.x) | `fix(api): handle timeout errors` |
| `feat!:` or `BREAKING CHANGE:` | Major (x.0.0) | `feat(api)!: change response format` |
| `docs:`, `chore:`, `test:` | No bump | `docs: update README` |

### Examples

```bash
# Feature (minor bump: 0.1.0 → 0.2.0)
git commit -m "feat(collections): add folder support"

# Bug fix (patch bump: 0.1.0 → 0.1.1)
git commit -m "fix(ui): fix panel resize issue"

# Breaking change (major bump: 0.1.0 → 1.0.0)
git commit -m "feat(api)!: change collection file format"

# Or with footer:
git commit -m "feat(api): change collection format

BREAKING CHANGE: Collection files now use v2 schema"
```

## Release Types

### Standard Release

1. Push commits to `main` with conventional commit messages
2. Release Please automatically creates/updates a PR titled "chore(main): release X.Y.Z"
3. Review the PR (changelog, version bump)
4. Merge the PR
5. GitHub Release is created automatically
6. GoReleaser builds and publishes binaries
7. Homebrew formula is updated

### Pre-release (Alpha/Beta)

For pre-releases, add the prerelease type to commit:

```bash
git commit -m "feat(ui): add experimental feature" --trailer "Release-As: 1.0.0-alpha.1"
```

### Manual Release (Emergency)

If needed, you can manually trigger a release:

```bash
# Create and push a tag
git tag v1.2.3
git push origin v1.2.3
```

This will skip Release Please and directly trigger GoReleaser.

## Homebrew Installation

After the first release, users can install via Homebrew:

```bash
# Add the tap
brew tap kbrdn1/tap

# Install lazycurl
brew install lazycurl

# Or in one command
brew install kbrdn1/tap/lazycurl
```

## Files Involved

| File | Purpose |
|------|---------|
| `.github/workflows/release-please.yml` | Main release workflow |
| `release-please-config.json` | Release Please configuration |
| `.release-please-manifest.json` | Current version tracking |
| `.goreleaser.yml` | Build and distribution configuration |
| `CHANGELOG.md` | Auto-updated changelog |

## Troubleshooting

### Release Please PR not created

- Ensure commits follow conventional commit format
- Check Actions tab for workflow errors
- Verify `GITHUB_TOKEN` has write permissions

### Homebrew formula not updated

- Check `HOMEBREW_TAP_TOKEN` secret is configured
- Verify the `homebrew-tap` repository exists
- Check GoReleaser logs in Actions

### Version not bumping correctly

- Review commit messages for proper conventional commit format
- Check `release-please-config.json` for changelog section configuration
- Verify `.release-please-manifest.json` has correct current version

## Local Testing

### Test GoReleaser locally

```bash
# Install goreleaser
brew install goreleaser

# Test build without publishing
goreleaser build --snapshot --clean

# Test full release (dry run)
goreleaser release --snapshot --clean
```

### Test version flag

```bash
# Build and test version
go build -o lazycurl ./cmd/lazycurl
./lazycurl --version
# Output: lazycurl dev (commit: none, built: unknown)
```
