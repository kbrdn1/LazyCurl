# API Overview

LazyCurl provides a JavaScript scripting API via the `lc` global object.

## Script Types

| Type | When it Runs | Primary Use |
|------|--------------|-------------|
| **Pre-request** | Before HTTP request is sent | Modify requests, add auth |
| **Post-response** | After response is received | Validate, test, extract data |

## Available Objects

### Core Objects

| Object | Description | Pre-request | Post-response |
|--------|-------------|:-----------:|:-------------:|
| [lc.request](API-Request) | HTTP request manipulation | Read/Write | Read-only |
| [lc.response](API-Response) | HTTP response data | - | Read-only |

### Variable Storage

| Object | Description | Persistence | Pre-request | Post-response |
|--------|-------------|-------------|:-----------:|:-------------:|
| [lc.env](API-Env) | Environment variables | File-based | Read/Write | Read/Write |
| [lc.globals](API-Env#lcglobals) | Session variables | In-memory | Read/Write | Read/Write |

### Testing

| Object | Description | Pre-request | Post-response |
|--------|-------------|:-----------:|:-------------:|
| [lc.test](API-Test) | Test case definition | - | Available |
| [lc.expect](API-Test#lcexpect) | Fluent assertions | - | Available |

### Utilities

| Object | Description | Pre-request | Post-response |
|--------|-------------|:-----------:|:-------------:|
| [lc.cookies](API-Cookies) | Cookie jar management | Read/Write | Read/Write |
| [lc.crypto](API-Crypto) | Hash functions, HMAC | Available | Available |
| [lc.base64](API-Base64) | Base64 encoding/decoding | Available | Available |
| [lc.variables](API-Variables) | Dynamic data generation | Available | Available |
| [lc.info](API-Info) | Execution context | Read-only | Read-only |
| [lc.sendRequest](API-SendRequest) | Request chaining | Available | Available |
| [console](API-Console) | Logging to console | Available | Available |

## Quick Reference

### Request Manipulation (Pre-request)

```javascript
// URL and method
lc.request.url = "https://api.example.com/users";
var method = lc.request.method; // Read-only

// Headers
lc.request.headers.set("Authorization", "Bearer " + token);
lc.request.headers.get("Content-Type");
lc.request.headers.remove("X-Custom");

// Body
lc.request.body.set(JSON.stringify({ name: "John" }));
var body = lc.request.body.json();

// Query parameters
lc.request.params.get("page");
lc.request.params.has("limit");
```

### Response Access (Post-response)

```javascript
// Status
var status = lc.response.status;        // 200
var text = lc.response.statusText;      // "200 OK"

// Timing
var ms = lc.response.time;              // Response time in ms

// Headers
var type = lc.response.headers.get("Content-Type");

// Body
var raw = lc.response.body.raw();       // Raw string
var json = lc.response.body.json();     // Parsed JSON
```

### Testing (Post-response)

```javascript
lc.test("Status is 200", function() {
    lc.expect(lc.response.status).toBe(200);
});

lc.test("Has valid data", function() {
    var data = lc.response.body.json();
    lc.expect(data).toHaveProperty("id");
    lc.expect(data.name).not.toBeNull();
});
```

### Variables

```javascript
// Environment (persisted to file)
lc.env.set("token", "abc123");
var token = lc.env.get("token");
lc.env.unset("token");

// Globals (session only)
lc.globals.set("counter", 1);
var count = lc.globals.get("counter");
```

### Utilities

```javascript
// Crypto
var hash = lc.crypto.sha256("data");
var hmac = lc.crypto.hmacSha256("data", "secret");

// Base64
var encoded = lc.base64.encode("hello");
var decoded = lc.base64.decode(encoded);

// Dynamic values
var uuid = lc.variables.uuid();
var timestamp = lc.variables.timestamp();
```

## Examples

- [Authentication Examples](Examples-Auth)
- [Testing Examples](Examples-Testing)
- [Request Chaining Examples](Examples-Chaining)

---

*[Home](Home) | [lc.request](API-Request)*
