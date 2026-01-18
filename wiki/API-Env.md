# lc.env & lc.globals

Two mechanisms for managing variables: environment variables (persisted) and global variables (session-only).

**Availability**: Both (pre-request and post-response)

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

## Comparison

| Feature | lc.env | lc.globals |
|---------|--------|------------|
| **Persistence** | Saved to file | In-memory only |
| **Survives restart** | Yes | No |
| **Value types** | Strings only | Any JavaScript value |
| **Best for** | API keys, config | Request chaining, state |

## lc.env

Environment variables persist to your active environment file.

### get(name)

```javascript
var baseUrl = lc.env.get("base_url");
var apiKey = lc.env.get("api_key");
```

### set(name, value)

```javascript
lc.env.set("auth_token", "eyJhbGciOiJIUzI1NiIs...");
lc.env.set("last_run", new Date().toISOString());
```

### unset(name)

```javascript
lc.env.unset("auth_token");
lc.env.unset("refresh_token");
```

### has(name)

```javascript
if (lc.env.has("api_key")) {
    lc.request.headers.set("X-API-Key", lc.env.get("api_key"));
}
```

### toObject()

```javascript
var allVars = lc.env.toObject();
console.log(JSON.stringify(allVars, null, 2));
```

## lc.globals

Global variables exist only in memory during the current session. They can store **any JavaScript value**.

### get(name)

```javascript
var user = lc.globals.get("current_user");
if (user) {
    console.log("Logged in as: " + user.name);
}
```

### set(name, value)

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
```

### unset(name)

```javascript
lc.globals.unset("temp_data");
```

### has(name)

```javascript
if (lc.globals.has("access_token")) {
    lc.request.headers.set("Authorization", "Bearer " + lc.globals.get("access_token"));
}
```

### clear()

```javascript
lc.globals.clear();
console.log("Session state cleared");
```

### toObject()

```javascript
var allGlobals = lc.globals.toObject();
console.log("Session variables: " + JSON.stringify(allGlobals));
```

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
}
```

### Request Counter

```javascript
var count = lc.globals.get("request_count") || 0;
count++;
lc.globals.set("request_count", count);

console.log("Request #" + count);
lc.request.headers.set("X-Request-Number", count.toString());
```

### Store User Context

```javascript
if (lc.response.status === 200) {
    var data = lc.response.body.json();

    // Store complex user object (only possible with globals)
    lc.globals.set("current_user", {
        id: data.user.id,
        name: data.user.name,
        email: data.user.email,
        roles: data.user.roles
    });

    // Store simple ID in env for persistence
    lc.env.set("user_id", data.user.id.toString());
}
```

## See Also

- [lc.request](API-Request) - HTTP request manipulation
- [lc.response](API-Response) - HTTP response access
- [Request Chaining Examples](Examples-Chaining)
- [Authentication Examples](Examples-Auth)

---

*[lc.response](API-Response) | [lc.test](API-Test)*
