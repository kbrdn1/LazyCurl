# lc.response

Read-only access to HTTP response data. Only available in post-response scripts.

**Availability**: Post-response only

## Quick Reference

| Property/Method | Description |
|-----------------|-------------|
| `status` | HTTP status code (e.g., 200, 404) |
| `statusText` | Full status text (e.g., "200 OK") |
| `time` | Response time in milliseconds |
| `headers.get(name)` | Get response header value |
| `headers.all()` | Get all response headers |
| `body.raw()` | Get raw body as string |
| `body.json()` | Parse body as JSON |

## Properties

### status

The HTTP status code.

```javascript
if (lc.response.status === 200) {
    console.log("Request successful");
} else if (lc.response.status === 401) {
    console.error("Unauthorized - check credentials");
} else if (lc.response.status >= 500) {
    console.error("Server error: " + lc.response.status);
}
```

### statusText

Full status text including the code.

```javascript
console.log("Status: " + lc.response.statusText);
// Output: "Status: 200 OK"
```

### time

Response time in milliseconds.

```javascript
console.log("Response received in " + lc.response.time + "ms");

lc.test("Response time is acceptable", function() {
    lc.expect(lc.response.time).toBeLessThan(2000);
});
```

## lc.response.headers

### headers.get(name)

Returns header value. **Case-insensitive**.

```javascript
var contentType = lc.response.headers.get("Content-Type");

var remaining = lc.response.headers.get("X-RateLimit-Remaining");
if (remaining && parseInt(remaining) < 10) {
    console.warn("Rate limit almost reached");
}
```

### headers.all()

Returns all headers as object.

```javascript
var headers = lc.response.headers.all();
for (var key in headers) {
    console.log(key + ": " + headers[key]);
}
```

## lc.response.body

### body.raw()

Returns raw body as string.

```javascript
var rawBody = lc.response.body.raw();
if (rawBody.includes("error")) {
    console.warn("Response contains error message");
}
```

### body.json()

Parses body as JSON. Returns `null` on failure.

```javascript
var data = lc.response.body.json();

if (data !== null) {
    console.log("User ID: " + data.id);

    if (data.accessToken) {
        lc.env.set("access_token", data.accessToken);
    }
}
```

## Examples

### Basic Response Validation

```javascript
console.log("Status: " + lc.response.statusText);
console.log("Time: " + lc.response.time + "ms");

lc.test("Response is successful", function() {
    lc.expect(lc.response.status).toBe(200);
});

lc.test("Response is JSON", function() {
    var contentType = lc.response.headers.get("Content-Type");
    lc.expect(contentType).toContain("application/json");
});
```

### Extract and Store Token

```javascript
if (lc.response.status === 200) {
    var data = lc.response.body.json();

    if (data && data.accessToken) {
        lc.env.set("access_token", data.accessToken);
        console.log("Access token saved");

        if (data.refreshToken) {
            lc.env.set("refresh_token", data.refreshToken);
        }

        if (data.expiresIn) {
            var expiresAt = Date.now() + (data.expiresIn * 1000);
            lc.globals.set("token_expires_at", expiresAt);
        }
    }
}
```

### Handle Different Status Codes

```javascript
var status = lc.response.status;

if (status >= 200 && status < 300) {
    console.log("Request successful");
    var data = lc.response.body.json();
    if (data) {
        lc.globals.set("last_response", data);
    }
} else if (status === 400) {
    var error = lc.response.body.json();
    console.error("Bad request: " + (error ? error.message : "Unknown"));
} else if (status === 401) {
    console.error("Unauthorized - please login again");
    lc.env.unset("access_token");
} else if (status === 429) {
    var retryAfter = lc.response.headers.get("Retry-After");
    console.warn("Rate limited. Retry after: " + retryAfter + "s");
} else if (status >= 500) {
    console.error("Server error: " + lc.response.statusText);
}
```

## See Also

- [lc.request](API-Request) - HTTP request manipulation
- [lc.test](API-Test) - Testing and assertions
- [Testing Examples](Examples-Testing)

---

*[lc.request](API-Request) | [lc.env](API-Env)*
