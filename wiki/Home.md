# LazyCurl Wiki

Welcome to the LazyCurl Wiki! This is your comprehensive guide to using LazyCurl, the terminal-based HTTP client with vim-style navigation.

## Quick Navigation

| Section | Description |
|---------|-------------|
| [Getting Started](Getting-Started) | Installation and your first script |
| [API Reference](API-Overview) | Complete JavaScript scripting API |
| [Examples](Examples-Auth) | Practical script examples |
| [Contributing](Contributing) | How to contribute to documentation |

## API Reference

### Core Objects

| Object | Description | Docs |
|--------|-------------|------|
| `lc.request` | Access and modify HTTP requests | [Request](API-Request) |
| `lc.response` | Read HTTP response data | [Response](API-Response) |
| `lc.env` | Environment variables (persisted) | [Environment](API-Env) |
| `lc.globals` | Session variables (in-memory) | [Environment](API-Env) |

### Testing & Assertions

| Object | Description | Docs |
|--------|-------------|------|
| `lc.test` | Define test cases | [Testing](API-Test) |
| `lc.expect` | Fluent assertions | [Testing](API-Test) |

### Utilities

| Object | Description | Docs |
|--------|-------------|------|
| `lc.cookies` | Cookie management | [Cookies](API-Cookies) |
| `lc.crypto` | Hash functions & HMAC | [Crypto](API-Crypto) |
| `lc.base64` | Base64 encoding/decoding | [Base64](API-Base64) |
| `lc.variables` | Dynamic data generation | [Variables](API-Variables) |
| `lc.info` | Execution context info | [Info](API-Info) |
| `lc.sendRequest` | Request chaining | [SendRequest](API-SendRequest) |
| `console` | Logging | [Console](API-Console) |

## Quick Example

```javascript
// Pre-request: Add authentication
var token = lc.env.get("auth_token");
if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
}

// Post-response: Validate and store
lc.test("Status is 200", function() {
    lc.expect(lc.response.status).toBe(200);
});

var data = lc.response.body.json();
if (data && data.token) {
    lc.env.set("auth_token", data.token);
}
```

## Resources

- [Full Documentation](https://github.com/kbrdn1/LazyCurl/tree/main/docs)
- [Report Issues](https://github.com/kbrdn1/LazyCurl/issues)
- [Discussions](https://github.com/kbrdn1/LazyCurl/discussions)

---

*This wiki covers LazyCurl v1.2.0+*
