# Contributing to Documentation

Guide for contributing to LazyCurl documentation.

## Documentation Structure

LazyCurl documentation exists in two places:

| Location | Purpose | Format |
|----------|---------|--------|
| `docs/` | Primary documentation in repository | Markdown with relative links |
| Wiki | GitHub Wiki for quick navigation | Markdown with wiki links |

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

The wiki is a separate repository. To edit:

1. Clone the wiki: `git clone https://github.com/kbrdn1/LazyCurl.wiki.git`
2. Make changes
3. Push directly (if you have access) or submit an issue

## File Naming Conventions

### Repository (`docs/`)

- Use **lowercase with hyphens**: `request-chaining.md`
- API docs go in `docs/api/`: `request.md`, `response.md`
- Examples go in `docs/examples/`: `authentication.md`

### Wiki

- Use **PascalCase with hyphens**: `API-Request.md`, `Examples-Auth.md`
- Sidebar file: `_Sidebar.md`
- Home page: `Home.md`

## Link Formats

### Repository Links

```markdown
[lc.request](./request.md)
[Authentication Examples](../examples/authentication.md)
```

### Wiki Links

```markdown
[lc.request](API-Request)
[Authentication Examples](Examples-Auth)
```

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

```

## Style Guidelines

### Code Examples

- Use **ES5 JavaScript** (no arrow functions, let/const)
- Keep examples **concise and practical**
- Add **comments** explaining key steps
- Test all examples before submitting

```javascript
// Good
var token = lc.env.get("token");
if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
}

// Avoid (ES6 syntax)
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

## Syncing Wiki with Repository

After updating `docs/`, the wiki should be updated to match:

1. API reference pages in `docs/api/` → Wiki `API-*.md` pages
2. Example pages in `docs/examples/` → Wiki `Examples-*.md` pages
3. Update `_Sidebar.md` if new pages added

### Key Differences

| Aspect | Repository | Wiki |
|--------|------------|------|
| Links | `[text](./file.md)` | `[text](Page-Name)` |
| Navigation | Footer prev/next links | Sidebar |
| Images | `![alt](./images/img.png)` | Upload to wiki |

## Reporting Issues

Found an error or want to suggest improvements?

1. [Open an issue](https://github.com/kbrdn1/LazyCurl/issues)
2. Use the `documentation` label
3. Describe the problem or suggestion clearly

## Questions?

- [Open a discussion](https://github.com/kbrdn1/LazyCurl/discussions)
- Check existing issues and discussions first

---

Thank you for contributing to LazyCurl documentation!

*[Home](Home)*
