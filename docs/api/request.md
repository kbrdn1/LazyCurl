# lc.request

Access and manipulate the current HTTP request. In pre-request scripts, most properties are mutable. In post-response scripts, all properties are read-only.

**Availability**: Pre-request (mutable) | Post-response (read-only)

> For the complete API reference, see [Scripting API Reference](../scripting-api-reference.md#lcrequest).

## Table of Contents

- [Overview](#overview)
- [Quick Reference](#quick-reference)
- [Properties](#properties)
- [lc.request.headers](#lcrequestheaders)
- [lc.request.body](#lcrequestbody)
- [lc.request.params](#lcrequestparams)
- [Examples](#examples)
- [See Also](#see-also)

---

## Overview

The `lc.request` object provides access to the current HTTP request. Use it to:

- Read request method, URL, headers, body, and query parameters
- Modify URL, headers, and body before the request is sent (pre-request only)
- Access query parameters for conditional logic

---

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

---

## Properties

### method

The HTTP method of the request. This property is read-only.

**Type**: `string`
**Read-only**: Yes

```javascript
var method = lc.request.method;
console.log("Method: " + method); // "GET", "POST", "PUT", "DELETE", etc.

// Conditional logic based on method
if (lc.request.method === "POST" || lc.request.method === "PUT") {
    // Ensure content-type is set for requests with body
    if (!lc.request.headers.get("Content-Type")) {
        lc.request.headers.set("Content-Type", "application/json");
    }
}
```

---

### url

The request URL including any query string. In pre-request scripts, you can modify this property.

**Type**: `string`
**Read-only**: No (pre-request) | Yes (post-response)

```javascript
// Read the URL
var url = lc.request.url;
console.log("URL: " + url);

// Modify the URL (pre-request only)
lc.request.url = "https://api.example.com/v2/users";

// Add query parameters dynamically
lc.request.url = lc.request.url + "?timestamp=" + Date.now();

// Replace environment-specific base URL
var env = lc.env.get("environment") || "dev";
lc.request.url = lc.request.url.replace("{{base_url}}", "https://" + env + ".api.example.com");
```

---

## lc.request.headers

Object providing methods for HTTP header manipulation.

### headers.get(name)

Returns the value of a header. Header lookup is **case-insensitive**.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The header name to retrieve |

**Returns**: `string` | `undefined` - The header value, or `undefined` if not found

```javascript
var contentType = lc.request.headers.get("Content-Type");
var auth = lc.request.headers.get("authorization"); // Case-insensitive

if (contentType) {
    console.log("Content-Type: " + contentType);
}
```

---

### headers.set(name, value)

Sets or updates a header value. **Pre-request scripts only**.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The header name |
| `value` | string | Yes | The header value |

**Returns**: `void`

```javascript
// Set a custom header
lc.request.headers.set("X-Custom-Header", "custom-value");

// Add authentication
var token = lc.env.get("auth_token");
lc.request.headers.set("Authorization", "Bearer " + token);

// Add timestamp header
lc.request.headers.set("X-Timestamp", Date.now().toString());
```

---

### headers.remove(name)

Removes a header from the request. **Pre-request scripts only**. Header lookup is case-insensitive.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The header name to remove |

**Returns**: `void`

```javascript
// Remove a header
lc.request.headers.remove("X-Deprecated-Header");

// Remove default content-type for form uploads
lc.request.headers.remove("Content-Type");
```

---

### headers.all()

Returns a copy of all headers as a key-value object.

**Returns**: `object` - Object containing all headers

```javascript
var allHeaders = lc.request.headers.all();
console.log(JSON.stringify(allHeaders, null, 2));

// Iterate over headers
for (var name in allHeaders) {
    console.log(name + ": " + allHeaders[name]);
}
```

---

## lc.request.body

Object providing methods for request body access and manipulation.

### body.raw()

Returns the raw request body as a string.

**Returns**: `string` - The raw body content

```javascript
var rawBody = lc.request.body.raw();
console.log("Body length: " + rawBody.length + " characters");
```

---

### body.json()

Parses the request body as JSON. Returns `null` if parsing fails or body is empty.

**Returns**: `object` | `null` - Parsed JSON object, or `null` if invalid/empty

```javascript
var data = lc.request.body.json();
if (data) {
    console.log("User ID: " + data.userId);
    console.log("Name: " + data.name);
}
```

---

### body.set(content)

Sets the request body content. **Pre-request scripts only**.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `content` | string | Yes | The new body content |

**Returns**: `void`

```javascript
// Set body as JSON string
lc.request.body.set('{"name": "John", "email": "john@example.com"}');

// Build body programmatically
var payload = {
    timestamp: Date.now(),
    requestId: lc.variables.uuid(),
    data: lc.request.body.json()
};
lc.request.body.set(JSON.stringify(payload));
```

---

## lc.request.params

Object providing **read-only** access to URL query parameters.

### params.get(name)

Returns the first value of a query parameter.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The parameter name |

**Returns**: `string` | `undefined` - The parameter value, or `undefined` if not found

```javascript
// URL: https://api.example.com/users?page=1&limit=10
var page = lc.request.params.get("page");   // "1"
var limit = lc.request.params.get("limit"); // "10"
var missing = lc.request.params.get("sort"); // undefined
```

---

### params.getAll(name)

Returns all values for a query parameter (for parameters with multiple values).

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The parameter name |

**Returns**: `string[]` - Array of all values (empty array if not found)

```javascript
// URL: https://api.example.com/search?tag=javascript&tag=golang&tag=rust
var tags = lc.request.params.getAll("tag");
// ["javascript", "golang", "rust"]

tags.forEach(function(tag) {
    console.log("Tag: " + tag);
});
```

---

### params.has(name)

Checks if a query parameter exists.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The parameter name to check |

**Returns**: `boolean` - `true` if the parameter exists

```javascript
// URL: https://api.example.com/users?active=true
if (lc.request.params.has("active")) {
    console.log("Filtering by active status");
}
```

---

### params.keys()

Returns an array of all query parameter names.

**Returns**: `string[]` - Array of parameter names

```javascript
// URL: https://api.example.com/search?q=test&page=1&limit=20
var keys = lc.request.params.keys();
// ["q", "page", "limit"]
```

---

### params.all()

Returns all query parameters as a key-value object. For parameters with multiple values, only the first value is included.

**Returns**: `object` - Object containing all parameters

```javascript
// URL: https://api.example.com/users?page=1&limit=10&sort=name
var params = lc.request.params.all();
// { "page": "1", "limit": "10", "sort": "name" }
```

---

## Examples

### Add Authentication Header

```javascript
// Pre-request: Add Bearer token from environment
var token = lc.env.get("auth_token");
if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
    console.log("Added Bearer token");
} else {
    console.warn("No auth_token found in environment");
}
```

### Dynamic Request Body

```javascript
// Pre-request: Add metadata to request body
var existingData = lc.request.body.json() || {};

var enrichedData = {
    ...existingData,
    metadata: {
        requestId: lc.variables.uuid(),
        timestamp: lc.variables.isoTimestamp(),
        client: "LazyCurl"
    }
};

lc.request.body.set(JSON.stringify(enrichedData));
```

### API Signature

```javascript
// Pre-request: Sign the request
var timestamp = Math.floor(Date.now() / 1000).toString();
var body = lc.request.body.raw() || "";
var secret = lc.env.get("api_secret");

var signature = lc.crypto.hmacSha256(timestamp + body, secret);

lc.request.headers.set("X-Timestamp", timestamp);
lc.request.headers.set("X-Signature", signature);
```

---

## See Also

- [lc.response](./response.md) - Access HTTP response data
- [lc.env](./env.md) - Environment variables
- [Authentication Examples](../examples/authentication.md)

---

*[← API Overview](./overview.md) | [lc.response →](./response.md)*
