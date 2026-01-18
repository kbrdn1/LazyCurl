# lc.cookies

Cookie management API for handling HTTP cookies in scripts.

**Availability**: Both (pre-request and post-response)

> For the complete API reference, see [Scripting API Reference](../scripting-api-reference.md#lccookies).

## Table of Contents

- [Overview](#overview)
- [Quick Reference](#quick-reference)
- [Methods](#methods)
- [Cookie Options](#cookie-options)
- [Examples](#examples)
- [See Also](#see-also)

---

## Overview

The `lc.cookies` API provides methods to read, write, and manage HTTP cookies:

- Read cookies from responses
- Set cookies for requests
- Manage cookie lifecycle (delete, clear)
- Generate Cookie header strings

---

## Quick Reference

| Method | Description |
|--------|-------------|
| `get(name)` | Get cookie value by name |
| `getAll()` | Get all cookies as array |
| `set(name, value, options?)` | Set a cookie |
| `delete(name)` | Remove a cookie |
| `clear()` | Remove all cookies |
| `has(name)` | Check if cookie exists |
| `toHeader()` | Generate Cookie header string |

---

## Methods

### get(name)

Returns the value of a cookie by name.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The cookie name |

**Returns**: `string` | `undefined` - The cookie value, or `undefined` if not found

```javascript
var sessionId = lc.cookies.get("session_id");
if (sessionId) {
    console.log("Session: " + sessionId);
}
```

---

### getAll()

Returns all cookies as an array of cookie objects.

**Returns**: `array` - Array of cookie objects with name, value, and options

```javascript
var allCookies = lc.cookies.getAll();
allCookies.forEach(function(cookie) {
    console.log(cookie.name + " = " + cookie.value);
});
```

---

### set(name, value, options?)

Sets a cookie with optional attributes.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The cookie name |
| `value` | string | Yes | The cookie value |
| `options` | object | No | Cookie options (see below) |

**Returns**: `void`

```javascript
// Simple cookie
lc.cookies.set("user_id", "12345");

// Cookie with options
lc.cookies.set("auth_token", "abc123xyz", {
    domain: "api.example.com",
    path: "/",
    secure: true,
    httpOnly: true
});
```

---

### delete(name)

Removes a cookie by name.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The cookie name to delete |

**Returns**: `void`

```javascript
lc.cookies.delete("session_id");
```

---

### clear()

Removes all cookies.

**Returns**: `void`

```javascript
// Clear all cookies (e.g., for logout)
lc.cookies.clear();
console.log("All cookies cleared");
```

---

### has(name)

Checks if a cookie exists.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The cookie name to check |

**Returns**: `boolean` - `true` if the cookie exists

```javascript
if (lc.cookies.has("csrf_token")) {
    lc.request.headers.set("X-CSRF-Token", lc.cookies.get("csrf_token"));
}
```

---

### toHeader()

Generates a Cookie header string from all cookies.

**Returns**: `string` - Cookie header value (e.g., `"name1=value1; name2=value2"`)

```javascript
// Add all cookies to request
var cookieHeader = lc.cookies.toHeader();
if (cookieHeader) {
    lc.request.headers.set("Cookie", cookieHeader);
}
```

---

## Cookie Options

When setting cookies with `lc.cookies.set()`, you can specify options:

| Property | Type | Description |
|----------|------|-------------|
| `domain` | string | Domain scope for the cookie |
| `path` | string | Path scope for the cookie |
| `secure` | boolean | Cookie sent only over HTTPS |
| `httpOnly` | boolean | Cookie inaccessible to client-side JavaScript |
| `expires` | string | Expiration date (RFC1123 format) |

```javascript
lc.cookies.set("session", "abc123", {
    domain: ".example.com",
    path: "/api",
    secure: true,
    httpOnly: true,
    expires: "Wed, 09 Jun 2024 10:18:14 GMT"
});
```

---

## Examples

### CSRF Token Workflow

```javascript
// Post-response: Capture CSRF token from cookies
if (lc.cookies.has("csrf_token")) {
    var csrfToken = lc.cookies.get("csrf_token");
    lc.env.set("csrf_token", csrfToken);
    console.log("CSRF token captured");
}
```

```javascript
// Pre-request: Use CSRF token
if (lc.cookies.has("csrf_token")) {
    // Add token to header
    lc.request.headers.set("X-CSRF-Token", lc.cookies.get("csrf_token"));

    // Include all cookies
    lc.request.headers.set("Cookie", lc.cookies.toHeader());
}
```

### Session Management

```javascript
// Pre-request: Check session validity
if (lc.cookies.has("session_id")) {
    console.log("Session cookie present");
    lc.request.headers.set("Cookie", lc.cookies.toHeader());
} else {
    console.warn("No session - request may fail");
}
```

```javascript
// Post-response: Handle session expiry
if (lc.response.status === 401) {
    console.log("Session expired, clearing cookies");
    lc.cookies.clear();
    lc.env.unset("session_id");
}
```

### Cookie-Based Authentication

```javascript
// Post-response: Store authentication cookies
if (lc.response.status === 200) {
    var allCookies = lc.cookies.getAll();

    allCookies.forEach(function(cookie) {
        console.log("Received cookie: " + cookie.name);

        // Store important cookies in environment
        if (cookie.name === "auth_token" || cookie.name === "refresh_token") {
            lc.env.set("cookie_" + cookie.name, cookie.value);
        }
    });
}
```

### Multiple Cookie Handling

```javascript
// Pre-request: Set multiple cookies
lc.cookies.set("client_id", lc.env.get("client_id"));
lc.cookies.set("tracking_id", lc.variables.uuid());
lc.cookies.set("timestamp", Date.now().toString());

// Add all cookies to request
lc.request.headers.set("Cookie", lc.cookies.toHeader());
```

---

## See Also

- [lc.request](./request.md) - HTTP request manipulation
- [Authentication Examples](../examples/authentication.md)

---

*[← lc.test](./test.md) | [lc.crypto →](./crypto.md)*
