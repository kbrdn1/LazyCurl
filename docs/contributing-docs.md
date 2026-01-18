# Contributing to LazyCurl Documentation

Guide for contributors maintaining and improving LazyCurl documentation.

## Documentation Structure

LazyCurl documentation exists in two synchronized locations:

| Location | Purpose | Format | URL |
|----------|---------|--------|-----|
| `docs/` | Primary source in repository | Markdown with relative links | GitHub repo |
| Wiki | GitHub Wiki for quick navigation | Markdown with wiki links | github.com/kbrdn1/LazyCurl/wiki |

### Repository Structure (`docs/`)

```
docs/
├── index.md                    # Documentation home
├── api/                        # API Reference (source of truth)
│   ├── overview.md             # API overview and quick reference
│   ├── request.md              # lc.request module
│   ├── response.md             # lc.response module
│   ├── env.md                  # lc.env and lc.globals
│   ├── test.md                 # lc.test and lc.expect
│   ├── cookies.md              # lc.cookies module
│   ├── crypto.md               # lc.crypto module
│   ├── base64.md               # lc.base64 and globals
│   ├── variables.md            # lc.variables module
│   ├── info.md                 # lc.info module
│   ├── sendrequest.md          # lc.sendRequest()
│   └── console.md              # console module
└── examples/                   # Practical examples
    ├── authentication.md       # Auth patterns
    ├── testing.md              # Testing patterns
    └── request-chaining.md     # Chaining patterns
```

### Wiki Structure

```
wiki/
├── Home.md                     # Wiki landing page
├── _Sidebar.md                 # Navigation sidebar
├── Getting-Started.md          # Quick start guide
├── API-Overview.md             # API overview
├── API-Request.md              # lc.request (adapted)
├── API-Response.md             # lc.response (adapted)
├── API-Env.md                  # lc.env/globals (adapted)
├── API-Test.md                 # lc.test/expect (adapted)
├── API-Cookies.md              # lc.cookies (adapted)
├── API-Crypto.md               # lc.crypto (adapted)
├── API-Base64.md               # lc.base64 (adapted)
├── API-Variables.md            # lc.variables (adapted)
├── API-Info.md                 # lc.info (adapted)
├── API-SendRequest.md          # sendRequest (adapted)
├── API-Console.md              # console (adapted)
├── Examples-Auth.md            # Auth examples (adapted)
├── Examples-Testing.md         # Testing examples (adapted)
├── Examples-Chaining.md        # Chaining examples (adapted)
└── Contributing.md             # Contribution guide
```

## Making Changes

### Repository Documentation (`docs/`)

1. **Fork and clone** the repository
2. Make changes to files in `docs/`
3. **Test links** locally
4. Submit a **pull request**

```bash
git clone https://github.com/kbrdn1/LazyCurl.git
cd LazyCurl
# Edit docs files
git checkout -b docs/your-change
git add docs/
git commit -m "docs: describe your change"
git push origin docs/your-change
```

### Wiki Documentation

The wiki is a separate repository:

```bash
# Clone the wiki
git clone https://github.com/kbrdn1/LazyCurl.wiki.git
cd LazyCurl.wiki

# Make changes
# Push directly (if you have access) or submit an issue
git add .
git commit -m "docs: describe your change"
git push origin master
```

## File Naming Conventions

### Repository (`docs/`)

- Use **lowercase with hyphens**: `request-chaining.md`
- API docs in `docs/api/`: `request.md`, `response.md`
- Examples in `docs/examples/`: `authentication.md`

### Wiki

- Use **PascalCase with hyphens**: `API-Request.md`, `Examples-Auth.md`
- Sidebar file: `_Sidebar.md`
- Home page: `Home.md`

## Link Format Differences

### Repository Links (relative paths)

```markdown
[lc.request](./request.md)
[Authentication Examples](../examples/authentication.md)
[API Overview](./overview.md)
```

### Wiki Links (page names only)

```markdown
[lc.request](API-Request)
[Authentication Examples](Examples-Auth)
[API Overview](API-Overview)
```

## Synchronization Process

**Source of Truth**: Repository `docs/` is the primary source. Wiki adapts content.

### When to Sync

Sync the Wiki after:

- Adding new API methods
- Updating method signatures
- Adding new examples
- Fixing documentation errors
- Major restructuring

### Sync Workflow

1. **Make changes in `docs/`** (repository)
2. **Submit PR** and get it merged
3. **Update Wiki** by adapting changes:
   - Convert relative links to wiki links
   - Update `_Sidebar.md` if adding new pages
   - Verify all links work

### Link Conversion Table

| Repo Link | Wiki Link |
|-----------|-----------|
| `[text](./file.md)` | `[text](Page-Name)` |
| `[text](../examples/file.md)` | `[text](Examples-File)` |
| `[text](#anchor)` | `[text](#anchor)` |

### Page Mapping

| Repository File | Wiki Page |
|-----------------|-----------|
| `docs/api/request.md` | `API-Request.md` |
| `docs/api/response.md` | `API-Response.md` |
| `docs/api/env.md` | `API-Env.md` |
| `docs/api/test.md` | `API-Test.md` |
| `docs/api/cookies.md` | `API-Cookies.md` |
| `docs/api/crypto.md` | `API-Crypto.md` |
| `docs/api/base64.md` | `API-Base64.md` |
| `docs/api/variables.md` | `API-Variables.md` |
| `docs/api/info.md` | `API-Info.md` |
| `docs/api/sendrequest.md` | `API-SendRequest.md` |
| `docs/api/console.md` | `API-Console.md` |
| `docs/examples/authentication.md` | `Examples-Auth.md` |
| `docs/examples/testing.md` | `Examples-Testing.md` |
| `docs/examples/request-chaining.md` | `Examples-Chaining.md` |

## Documentation Template

Use this structure for API module pages:

```markdown
# Module Name

Brief description of the module.

**Availability**: Pre-request | Post-response | Both

## Quick Reference

| Method | Description |
|--------|-------------|
| `method1()` | Description |
| `method2()` | Description |

## Methods

### method1()

Description of what it does.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `param` | type | Yes/No | Description |

**Returns**: `type` - Description

```javascript
// Example code
```

## Examples

### Common Use Case

```javascript
// Practical example
```

## See Also

- [Related Module](Link)

---

*[Previous Page](Link) | [Next Page](Link)*

```

## Style Guidelines

### Code Examples

- Use **ES5 JavaScript** (no arrow functions, let/const)
- Keep examples **concise and practical**
- Add **comments** explaining key steps
- Test all examples before submitting

```javascript
// Good - ES5 syntax
var token = lc.env.get("token");
if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
}

// Avoid - ES6 syntax
const token = lc.env.get("token");
token && lc.request.headers.set("Authorization", `Bearer ${token}`);
```

### Writing Style

- Use **active voice**: "This method returns..." not "is returned by..."
- Be **concise**: Get to the point quickly
- Use **consistent terminology**: "method" not "function", "property" not "field"
- Include **practical examples** for every method

### Tables

Use tables for quick references:

```markdown
| Method | Description |
|--------|-------------|
| `get()` | Returns the value |
| `set()` | Sets the value |
```

## Adding New API Methods

When a new method is added to the scripting API:

1. **Update `docs/api/<module>.md`**:
   - Add to Quick Reference table
   - Add full method documentation
   - Include at least one example

2. **Update `docs/api/overview.md`**:
   - Add to the quick reference table
   - Update method count if needed

3. **Update Wiki pages**:
   - Sync changes to `wiki/API-<Module>.md`
   - Update `wiki/API-Overview.md`

4. **Update examples if relevant**:
   - Add usage to appropriate example file
   - Sync to Wiki example pages

## Validation Checklist

Before submitting documentation changes:

- [ ] All links work (relative or wiki format)
- [ ] Code examples are tested and work
- [ ] ES5 JavaScript syntax used
- [ ] Tables are properly formatted
- [ ] Navigation links are correct
- [ ] Spelling and grammar checked
- [ ] Consistent formatting with existing docs

## Reporting Issues

Found an error or want to suggest improvements?

1. [Open an issue](https://github.com/kbrdn1/LazyCurl/issues)
2. Use the `documentation` label
3. Describe the problem or suggestion clearly
4. Reference specific file paths if applicable

## Questions?

- [Open a discussion](https://github.com/kbrdn1/LazyCurl/discussions)
- Check existing issues and discussions first

---

*[Back to Documentation Index](./index.md)*
