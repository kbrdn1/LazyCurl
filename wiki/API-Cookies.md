# lc.cookies

Manage HTTP cookies for requests and responses.

**Availability**: Both (pre-request and post-response)

## Quick Reference

| Method | Description |
|--------|-------------|
| `get(name)` | Get cookie value by name |
| `getAll()` | Get all cookies as array |
| `set(name, value)` | Set a cookie |
| `delete(name)` | Delete a cookie |
| `clear()` | Clear all cookies |
| `has(name)` | Check if cookie exists |
| `toObject()` | Get cookies as object |

## Methods

### get(name)

Returns the value of a cookie.

```javascript
var sessionId = lc.cookies.get("session_id");
var auth = lc.cookies.get("auth_token");

if (sessionId) {
    console.log("Session: " + sessionId);
}
```

### getAll()

Returns all cookies as an array of objects.

```javascript
var cookies = lc.cookies.getAll();
cookies.forEach(function(cookie) {
    console.log(cookie.name + " = " + cookie.value);
});
```

### set(name, value)

Sets a cookie. Optionally provide cookie attributes.

```javascript
// Simple cookie
lc.cookies.set("preference", "dark-mode");

// Cookie with attributes
lc.cookies.set("session", "abc123", {
    domain: "api.example.com",
    path: "/",
    secure: true,
    httpOnly: true,
    maxAge: 3600
});
```

### delete(name)

Deletes a cookie by name.

```javascript
lc.cookies.delete("expired_token");
```

### clear()

Removes all cookies.

```javascript
lc.cookies.clear();
console.log("All cookies cleared");
```

### has(name)

Checks if a cookie exists.

```javascript
if (lc.cookies.has("session_id")) {
    console.log("Session cookie present");
} else {
    console.warn("No session - user may need to login");
}
```

### toObject()

Returns all cookies as key-value object.

```javascript
var cookies = lc.cookies.toObject();
console.log(JSON.stringify(cookies, null, 2));
```

## Examples

### Session Management

```javascript
// Pre-request: Check for session
if (!lc.cookies.has("session_id")) {
    console.warn("No session cookie - request may fail");
}
```

```javascript
// Post-response: Store session from response
if (lc.response.status === 200) {
    var sessionId = lc.response.headers.get("X-Session-ID");
    if (sessionId) {
        lc.cookies.set("session_id", sessionId);
        console.log("Session established");
    }
}
```

### Auth Cookie Handling

```javascript
// Post-response: Clear auth on 401
if (lc.response.status === 401) {
    lc.cookies.delete("auth_token");
    lc.cookies.delete("refresh_token");
    console.warn("Auth cookies cleared");
}
```

### Debug Cookies

```javascript
// Pre-request: Log all cookies
var cookies = lc.cookies.toObject();
console.log("Cookies being sent:");
for (var name in cookies) {
    console.log("  " + name + ": " + cookies[name].substring(0, 20) + "...");
}
```

## See Also

- [lc.request](API-Request) - HTTP request manipulation
- [lc.env](API-Env) - Environment variables
- [Authentication Examples](Examples-Auth)

---

*[lc.test](API-Test) | [lc.crypto](API-Crypto)*
