# lc.request

Access and manipulate the current HTTP request. In pre-request scripts, most properties are mutable. In post-response scripts, all properties are read-only.

**Availability**: Pre-request (mutable) | Post-response (read-only)

## Quick Reference

| Method/Property | Description |
|-----------------|-------------|
| `method` | HTTP method (GET, POST, etc.) - read-only |
| `url` | Request URL - read/write in pre-request |
| `headers.get(name)` | Get header value |
| `headers.set(name, value)` | Set header (pre-request only) |
| `headers.remove(name)` | Remove header (pre-request only) |
| `headers.all()` | Get all headers as object |
| `body.raw()` | Get raw body as string |
| `body.json()` | Parse body as JSON |
| `body.set(content)` | Set body content (pre-request only) |
| `params.get(name)` | Get query parameter value |
| `params.getAll(name)` | Get all values for parameter |
| `params.has(name)` | Check if parameter exists |
| `params.keys()` | Get all parameter names |
| `params.all()` | Get all parameters as object |

## Properties

### method

The HTTP method of the request (read-only).

```javascript
var method = lc.request.method;
console.log("Method: " + method); // "GET", "POST", etc.

if (lc.request.method === "POST" || lc.request.method === "PUT") {
    if (!lc.request.headers.get("Content-Type")) {
        lc.request.headers.set("Content-Type", "application/json");
    }
}
```

### url

The request URL including any query string.

```javascript
// Read the URL
var url = lc.request.url;

// Modify the URL (pre-request only)
lc.request.url = "https://api.example.com/v2/users";

// Add query parameters dynamically
lc.request.url = lc.request.url + "?timestamp=" + Date.now();
```

## lc.request.headers

### headers.get(name)

Returns header value. Header lookup is **case-insensitive**.

```javascript
var contentType = lc.request.headers.get("Content-Type");
var auth = lc.request.headers.get("authorization"); // Case-insensitive
```

### headers.set(name, value)

Sets or updates a header. **Pre-request only**.

```javascript
lc.request.headers.set("X-Custom-Header", "custom-value");
lc.request.headers.set("Authorization", "Bearer " + token);
```

### headers.remove(name)

Removes a header. **Pre-request only**.

```javascript
lc.request.headers.remove("X-Deprecated-Header");
```

### headers.all()

Returns all headers as object.

```javascript
var allHeaders = lc.request.headers.all();
console.log(JSON.stringify(allHeaders, null, 2));
```

## lc.request.body

### body.raw()

Returns the raw request body as string.

```javascript
var rawBody = lc.request.body.raw();
```

### body.json()

Parses body as JSON. Returns `null` if parsing fails.

```javascript
var data = lc.request.body.json();
if (data) {
    console.log("User ID: " + data.userId);
}
```

### body.set(content)

Sets body content. **Pre-request only**.

```javascript
lc.request.body.set('{"name": "John"}');

// Build body programmatically
var payload = {
    timestamp: Date.now(),
    requestId: lc.variables.uuid()
};
lc.request.body.set(JSON.stringify(payload));
```

## lc.request.params

Query parameters are **read-only**.

### params.get(name)

```javascript
// URL: https://api.example.com/users?page=1&limit=10
var page = lc.request.params.get("page");   // "1"
var limit = lc.request.params.get("limit"); // "10"
```

### params.getAll(name)

```javascript
// URL: ?tag=javascript&tag=golang
var tags = lc.request.params.getAll("tag");
// ["javascript", "golang"]
```

### params.has(name)

```javascript
if (lc.request.params.has("active")) {
    console.log("Filtering by active status");
}
```

### params.keys()

```javascript
var keys = lc.request.params.keys();
// ["q", "page", "limit"]
```

### params.all()

```javascript
var params = lc.request.params.all();
// { "page": "1", "limit": "10" }
```

## Examples

### Add Authentication Header

```javascript
var token = lc.env.get("auth_token");
if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
    console.log("Added Bearer token");
}
```

### Dynamic Request Body

```javascript
var existingData = lc.request.body.json() || {};

var enrichedData = {
    ...existingData,
    metadata: {
        requestId: lc.variables.uuid(),
        timestamp: lc.variables.isoTimestamp()
    }
};

lc.request.body.set(JSON.stringify(enrichedData));
```

### API Signature

```javascript
var timestamp = Math.floor(Date.now() / 1000).toString();
var body = lc.request.body.raw() || "";
var secret = lc.env.get("api_secret");

var signature = lc.crypto.hmacSha256(timestamp + body, secret);

lc.request.headers.set("X-Timestamp", timestamp);
lc.request.headers.set("X-Signature", signature);
```

## See Also

- [lc.response](API-Response) - HTTP response access
- [lc.env](API-Env) - Environment variables
- [Authentication Examples](Examples-Auth)

---

*[API Overview](API-Overview) | [lc.response](API-Response)*
