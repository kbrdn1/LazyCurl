# lc.response

Read-only access to HTTP response data. Only available in post-response scripts.

**Availability**: Post-response only

> For the complete API reference, see [Scripting API Reference](../scripting-api-reference.md#lcresponse).

## Table of Contents

- [Overview](#overview)
- [Quick Reference](#quick-reference)
- [Properties](#properties)
- [lc.response.headers](#lcresponseheaders)
- [lc.response.body](#lcresponsebody)
- [Examples](#examples)
- [See Also](#see-also)

---

## Overview

The `lc.response` object provides read-only access to the HTTP response after the request completes. Use it to:

- Check response status codes
- Access response headers
- Parse response body (JSON, raw text)
- Measure response time

> **Note**: This object is only available in post-response scripts. Attempting to access it in pre-request scripts will result in `undefined`.

---

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

---

## Properties

### status

The HTTP status code of the response.

**Type**: `number`
**Read-only**: Yes

```javascript
// Check status code
if (lc.response.status === 200) {
    console.log("Request successful");
} else if (lc.response.status === 401) {
    console.error("Unauthorized - check credentials");
} else if (lc.response.status >= 500) {
    console.error("Server error: " + lc.response.status);
}
```

---

### statusText

The full status text including the code.

**Type**: `string`
**Read-only**: Yes

```javascript
console.log("Status: " + lc.response.statusText);
// Output: "Status: 200 OK" or "Status: 404 Not Found"
```

---

### time

The response time in milliseconds (from request sent to response received).

**Type**: `number`
**Read-only**: Yes

```javascript
console.log("Response received in " + lc.response.time + "ms");

// Performance test
lc.test("Response time is acceptable", function() {
    lc.expect(lc.response.time).toBeLessThan(2000);
});
```

---

## lc.response.headers

Object providing methods to access response headers.

### headers.get(name)

Returns the value of a response header. Header lookup is **case-insensitive**.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | The header name to retrieve |

**Returns**: `string` | `undefined` - The header value, or `undefined` if not found

```javascript
// Get content type
var contentType = lc.response.headers.get("Content-Type");
console.log("Content-Type: " + contentType);

// Case-insensitive lookup
var authToken = lc.response.headers.get("x-auth-token");
var authTokenAlt = lc.response.headers.get("X-Auth-Token"); // Same result

// Check rate limit headers
var remaining = lc.response.headers.get("X-RateLimit-Remaining");
if (remaining && parseInt(remaining) < 10) {
    console.warn("Rate limit almost reached: " + remaining + " requests remaining");
}
```

---

### headers.all()

Returns a copy of all response headers as a key-value object.

**Returns**: `object` - Object containing all headers

```javascript
// Get all headers
var headers = lc.response.headers.all();

// Log all headers
for (var key in headers) {
    console.log(key + ": " + headers[key]);
}

// Check for specific patterns
var cacheControl = headers["Cache-Control"];
if (cacheControl && cacheControl.includes("no-cache")) {
    console.log("Response is not cached");
}
```

---

## lc.response.body

Object providing methods to access the response body.

### body.raw()

Returns the raw response body as a string.

**Returns**: `string` - The response body content

```javascript
var rawBody = lc.response.body.raw();
console.log("Body length: " + rawBody.length + " characters");

// Check if body contains specific text
if (rawBody.includes("error")) {
    console.warn("Response contains error message");
}
```

---

### body.json()

Parses the response body as JSON. Returns `null` if the body is empty or cannot be parsed.

**Returns**: `object` | `array` | `null` - Parsed JSON data, or `null` on parse failure

```javascript
var data = lc.response.body.json();

if (data !== null) {
    console.log("User ID: " + data.id);
    console.log("Username: " + data.username);

    // Store data for next request
    if (data.accessToken) {
        lc.env.set("access_token", data.accessToken);
    }
} else {
    console.error("Failed to parse response as JSON");
}
```

---

## Examples

### Basic Response Validation

```javascript
// Post-response: Log and validate response
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
// Post-response: Store authentication token
if (lc.response.status === 200) {
    var data = lc.response.body.json();

    if (data && data.accessToken) {
        // Store in environment (persisted)
        lc.env.set("access_token", data.accessToken);
        console.log("Access token saved to environment");

        // Store refresh token if present
        if (data.refreshToken) {
            lc.env.set("refresh_token", data.refreshToken);
        }

        // Calculate and store expiry
        if (data.expiresIn) {
            var expiresAt = Date.now() + (data.expiresIn * 1000);
            lc.globals.set("token_expires_at", expiresAt);
            console.log("Token expires: " + new Date(expiresAt).toISOString());
        }
    }
} else if (lc.response.status === 401) {
    console.error("Authentication failed");
    lc.env.unset("access_token");
}
```

### Comprehensive API Response Tests

```javascript
// Post-response: Full API response validation
lc.test("Status code is 200", function() {
    lc.expect(lc.response.status).toBe(200);
});

lc.test("Response time under 2 seconds", function() {
    lc.expect(lc.response.time).toBeLessThan(2000);
});

lc.test("Content-Type is JSON", function() {
    var contentType = lc.response.headers.get("Content-Type");
    lc.expect(contentType).toContain("application/json");
});

var data = lc.response.body.json();

lc.test("Response has data array", function() {
    lc.expect(data).not.toBeNull();
    lc.expect(data.items).toBeDefined();
    lc.expect(Array.isArray(data.items)).toBe(true);
});

lc.test("Pagination info present", function() {
    lc.expect(data.total).toBeDefined();
    lc.expect(data.page).toBeDefined();
    lc.expect(data.pageSize).toBeDefined();
});

// Log summary
console.log("Retrieved " + data.items.length + " of " + data.total + " items");
```

### Handle Different Status Codes

```javascript
// Post-response: Handle various status codes
var status = lc.response.status;

if (status >= 200 && status < 300) {
    // Success
    console.log("Request successful");
    var data = lc.response.body.json();
    if (data) {
        lc.globals.set("last_response", data);
    }
} else if (status === 400) {
    // Bad Request
    var error = lc.response.body.json();
    console.error("Bad request: " + (error ? error.message : "Unknown error"));
} else if (status === 401) {
    // Unauthorized
    console.error("Unauthorized - please login again");
    lc.env.unset("access_token");
} else if (status === 403) {
    // Forbidden
    console.error("Access forbidden");
} else if (status === 404) {
    // Not Found
    console.warn("Resource not found");
} else if (status === 429) {
    // Rate Limited
    var retryAfter = lc.response.headers.get("Retry-After");
    console.warn("Rate limited. Retry after: " + (retryAfter || "unknown") + " seconds");
} else if (status >= 500) {
    // Server Error
    console.error("Server error: " + lc.response.statusText);
}
```

---

## See Also

- [lc.request](./request.md) - HTTP request manipulation
- [lc.test](./test.md) - Testing and assertions
- [Testing Examples](../examples/testing.md)

---

*[← lc.request](./request.md) | [lc.env →](./env.md)*
