# lc.env & lc.globals

Two mechanisms for managing variables: environment variables (persisted) and global variables (session-only).

**Availability**: Both (pre-request and post-response)

> For the complete API reference, see [Scripting API Reference](../scripting-api-reference.md#lcenv--lcglobals).

## Table of Contents

- [Overview](#overview)
- [Quick Reference](#quick-reference)
- [lc.env](#lcenv)
- [lc.globals](#lcglobals)
- [Comparison](#comparison)
- [Examples](#examples)
- [See Also](#see-also)

---

## Overview

LazyCurl provides two variable storage mechanisms:

| Storage | Persistence | Value Types | Use Case |
|---------|-------------|-------------|----------|
| **lc.env** | Saved to environment file | Strings only | API keys, base URLs, configuration |
| **lc.globals** | In-memory (session only) | Any JavaScript value | Request chaining, temporary state |

---

## Quick Reference

### lc.env Methods

| Method | Description |
|--------|-------------|
| `get(name)` | Get environment variable value |
| `set(name, value)` | Set environment variable (persisted) |
| `unset(name)` | Remove environment variable |
| `has(name)` | Check if variable exists |
| `toObject()` | Get all variables as object |

### lc.globals Methods

| Method | Description |
|--------|-------------|
| `get(name)` | Get global variable value |
| `set(name, value)` | Set global variable (any type) |
| `unset(name)` | Remove global variable |
| `has(name)` | Check if variable exists |
| `clear()` | Remove all global variables |
| `toObject()` | Get all variables as object |

---

## lc.env

Environment variables are tied to your active environment file and **persist to disk**. Changes made via `lc.env.set()` are saved after script execution.

### get(name)

Retrieves the value of an environment variable.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The variable name |

**Returns**: `string` - The variable value, or empty string if not found

```javascript
var baseUrl = lc.env.get("base_url");
var apiKey = lc.env.get("api_key");

console.log("Base URL: " + baseUrl);
```

---

### set(name, value)

Sets an environment variable. The value is **persisted to the environment file**.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The variable name |
| `value` | string | Yes | The variable value |

**Returns**: `void`

```javascript
// Store a token (persisted to file)
lc.env.set("auth_token", "eyJhbGciOiJIUzI1NiIs...");

// Store timestamp
lc.env.set("last_run", new Date().toISOString());
```

---

### unset(name)

Removes an environment variable.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The variable name to remove |

**Returns**: `void`

```javascript
// Clear expired token
lc.env.unset("auth_token");
lc.env.unset("refresh_token");
```

---

### has(name)

Checks if an environment variable exists.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The variable name to check |

**Returns**: `boolean` - `true` if the variable exists

```javascript
if (lc.env.has("api_key")) {
    lc.request.headers.set("X-API-Key", lc.env.get("api_key"));
} else {
    console.warn("No API key configured");
}
```

---

### toObject()

Returns all environment variables as a JavaScript object.

**Returns**: `object` - All variables as key-value pairs

```javascript
var allVars = lc.env.toObject();
console.log(JSON.stringify(allVars, null, 2));

// Check which variables are set
for (var name in allVars) {
    console.log(name + " = " + allVars[name]);
}
```

---

## lc.globals

Global variables exist **only in memory** during the current LazyCurl session. They can store **any JavaScript value** (objects, arrays, numbers, booleans), not just strings.

### get(name)

Retrieves a global variable.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The variable name |

**Returns**: `any` - The variable value, or `null` if not found

```javascript
var user = lc.globals.get("current_user");
if (user) {
    console.log("Logged in as: " + user.name);
}

// Get with default value
var count = lc.globals.get("request_count") || 0;
```

---

### set(name, value)

Sets a global variable. Can store **any JavaScript value**.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The variable name |
| `value` | any | Yes | The variable value (any type) |

**Returns**: `void`

```javascript
// Store complex objects
lc.globals.set("user", {
    id: 123,
    name: "John Doe",
    roles: ["admin", "user"]
});

// Store arrays
lc.globals.set("request_history", []);

// Store numbers
lc.globals.set("request_count", 0);

// Store booleans
lc.globals.set("is_authenticated", true);
```

---

### unset(name)

Removes a global variable.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The variable name to remove |

**Returns**: `void`

```javascript
lc.globals.unset("temp_data");
```

---

### has(name)

Checks if a global variable exists.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The variable name to check |

**Returns**: `boolean` - `true` if the variable exists

```javascript
if (lc.globals.has("access_token")) {
    lc.request.headers.set("Authorization", "Bearer " + lc.globals.get("access_token"));
}
```

---

### clear()

Removes **all** global variables.

**Returns**: `void`

```javascript
// Reset session state
lc.globals.clear();
console.log("Session state cleared");
```

---

### toObject()

Returns all global variables as a JavaScript object.

**Returns**: `object` - All variables as key-value pairs

```javascript
var allGlobals = lc.globals.toObject();
console.log("Session variables: " + JSON.stringify(allGlobals));
```

---

## Comparison

| Feature | lc.env | lc.globals |
|---------|--------|------------|
| **Persistence** | Saved to environment file | In-memory only |
| **Survives restart** | Yes | No |
| **Value types** | Strings only | Any JavaScript value |
| **Scope** | Tied to active environment | Session-wide |
| **Best for** | Configuration, API keys | Request chaining, state |

### When to Use Each

**Use `lc.env` for:**

- API keys and secrets
- Base URLs
- Configuration that should persist
- Values shared across sessions

**Use `lc.globals` for:**

- Tokens obtained during request chaining
- Temporary computation results
- Complex objects (user data, arrays)
- Counters and state within a session

---

## Examples

### Token Management with Both

```javascript
// Post-response: Store token from login
if (lc.response.status === 200) {
    var data = lc.response.body.json();

    if (data && data.access_token) {
        // Store in globals for immediate use (fast, in-memory)
        lc.globals.set("access_token", data.access_token);

        // Also persist to env for future sessions
        lc.env.set("access_token", data.access_token);

        // Store expiry in globals (complex data)
        lc.globals.set("token_expires_at", Date.now() + (data.expires_in * 1000));

        console.log("Token stored");
    }
}
```

```javascript
// Pre-request: Use token with fallback
var token = lc.globals.get("access_token") || lc.env.get("access_token");

if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
} else {
    console.warn("No access token available");
}
```

### Request Counter

```javascript
// Pre-request: Track request count
var count = lc.globals.get("request_count") || 0;
count++;
lc.globals.set("request_count", count);

console.log("Request #" + count);
lc.request.headers.set("X-Request-Number", count.toString());
```

### Store User Context

```javascript
// Post-response: Store user data after login
if (lc.response.status === 200) {
    var data = lc.response.body.json();

    // Store complex user object (only possible with globals)
    lc.globals.set("current_user", {
        id: data.user.id,
        name: data.user.name,
        email: data.user.email,
        roles: data.user.roles,
        loginTime: new Date().toISOString()
    });

    // Store simple ID in env for persistence
    lc.env.set("user_id", data.user.id.toString());
}
```

### Environment-Specific Configuration

```javascript
// Pre-request: Configure based on environment
var env = lc.env.get("environment") || "development";

switch (env) {
    case "production":
        lc.request.headers.set("X-Environment", "prod");
        break;
    case "staging":
        lc.request.headers.set("X-Environment", "staging");
        lc.request.headers.set("X-Debug", "true");
        break;
    default:
        lc.request.headers.set("X-Environment", "dev");
        lc.request.headers.set("X-Debug", "true");
}
```

---

## See Also

- [lc.request](./request.md) - HTTP request manipulation
- [lc.response](./response.md) - HTTP response access
- [Request Chaining Examples](../examples/request-chaining.md)
- [Authentication Examples](../examples/authentication.md)

---

*[← lc.response](./response.md) | [lc.test →](./test.md)*
