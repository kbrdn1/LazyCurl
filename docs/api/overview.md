# JavaScript Scripting API Overview

LazyCurl provides a powerful JavaScript scripting API for automating request workflows, testing responses, and managing data dynamically.

**Runtime**: [Goja](https://github.com/dop251/goja) JavaScript engine (ES5.1+)

> For the complete single-page reference, see [Scripting API Reference](../scripting-api-reference.md).

## Table of Contents

- [Script Types](#script-types)
- [Quick Reference](#quick-reference)
- [The `lc` Global Object](#the-lc-global-object)
- [Getting Started](#getting-started)
- [API Modules](#api-modules)

---

## Script Types

| Type | Execution | Use Case |
|------|-----------|----------|
| **Pre-request** | Before HTTP request is sent | Add auth headers, modify URL, generate dynamic data |
| **Post-response** | After HTTP response received | Validate response, extract data, run tests |

---

## Quick Reference

### Core Modules

| Module | Description | Docs |
|--------|-------------|------|
| [`lc.request`](./request.md) | HTTP request manipulation (URL, headers, body, params) | [14 methods/properties](./request.md) |
| [`lc.response`](./response.md) | HTTP response access (status, headers, body) | [7 methods/properties](./response.md) |
| [`lc.env`](./env.md) | Environment variables (persisted to file) | [5 methods](./env.md) |
| [`lc.globals`](./env.md#lcglobals) | Session variables (in-memory) | [6 methods](./env.md#lcglobals) |
| [`lc.test`](./test.md) | Test definitions and assertions | [16 matchers](./test.md) |

### Utility Modules

| Module | Description | Docs |
|--------|-------------|------|
| [`lc.cookies`](./cookies.md) | Cookie management | [7 methods](./cookies.md) |
| [`lc.crypto`](./crypto.md) | Hash functions and HMAC | [7 methods](./crypto.md) |
| [`lc.base64`](./base64.md) | Base64 encoding/decoding | [4 methods](./base64.md) |
| [`lc.variables`](./variables.md) | Dynamic data generation | [12 methods](./variables.md) |
| [`lc.info`](./info.md) | Execution context info | [6 properties](./info.md) |
| [`lc.sendRequest`](./sendrequest.md) | Request chaining | [1 method](./sendrequest.md) |
| [`console`](./console.md) | Logging | [5 methods](./console.md) |

---

## The `lc` Global Object

All scripting APIs are accessible through the `lc` global object:

```javascript
// Pre-request: Add authentication header
var token = lc.env.get("auth_token");
if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
}

// Post-response: Validate and store data
lc.test("Status is 200", function() {
    lc.expect(lc.response.status).toBe(200);
});

var data = lc.response.body.json();
if (data && data.token) {
    lc.env.set("auth_token", data.token);
}
```

---

## Getting Started

### 1. Add a Pre-request Script

Navigate to the **Pre-request** tab in the Request panel and add:

```javascript
// Add timestamp header
lc.request.headers.set("X-Request-Time", Date.now().toString());

// Add API key from environment
var apiKey = lc.env.get("api_key");
if (apiKey) {
    lc.request.headers.set("X-API-Key", apiKey);
}
```

### 2. Add a Post-response Script

Navigate to the **Post-response** tab and add:

```javascript
// Log response info
console.log("Status: " + lc.response.status);
console.log("Time: " + lc.response.time + "ms");

// Run tests
lc.test("Response is successful", function() {
    lc.expect(lc.response.status).toBe(200);
});

lc.test("Response has data", function() {
    var data = lc.response.body.json();
    lc.expect(data).not.toBeNull();
});
```

### 3. Send the Request

Press `Ctrl+S` to send. View results in:

- **Console** tab: Script logs
- **Tests** tab: Assertion results

---

## API Modules

### Request & Response

- **[lc.request](./request.md)** - Read and modify the HTTP request
  - `lc.request.method` - HTTP method (read-only)
  - `lc.request.url` - Request URL
  - `lc.request.headers.*` - Header manipulation
  - `lc.request.body.*` - Body access/modification
  - `lc.request.params.*` - Query parameter access

- **[lc.response](./response.md)** - Access HTTP response data
  - `lc.response.status` - HTTP status code
  - `lc.response.headers.*` - Response headers
  - `lc.response.body.*` - Response body

### Variables

- **[lc.env](./env.md)** - Environment variables persisted to disk
- **[lc.globals](./env.md#lcglobals)** - Session variables in memory

### Testing

- **[lc.test](./test.md)** - Define tests with `lc.test(name, fn)`
- **[lc.expect](./test.md#lcexpectvalue)** - Fluent assertions

### Utilities

- **[lc.cookies](./cookies.md)** - Cookie management
- **[lc.crypto](./crypto.md)** - Cryptographic hash functions
- **[lc.base64](./base64.md)** - Base64 encoding/decoding
- **[lc.variables](./variables.md)** - Generate random test data
- **[lc.info](./info.md)** - Script execution context
- **[lc.sendRequest](./sendrequest.md)** - Chain multiple requests
- **[console](./console.md)** - Logging functions

---

## See Also

- [Authentication Examples](../examples/authentication.md)
- [Testing Examples](../examples/testing.md)
- [Request Chaining Examples](../examples/request-chaining.md)
- [Full API Reference](../scripting-api-reference.md)

---

*[← Back to Documentation](../index.md) | [lc.request →](./request.md)*
