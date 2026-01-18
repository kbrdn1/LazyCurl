# lc.base64

Base64 encoding and decoding utilities. Browser-compatible global functions (`btoa`, `atob`) are also available.

**Availability**: Both (pre-request and post-response)

> For the complete API reference, see [Scripting API Reference](../scripting-api-reference.md#lcbase64).

## Table of Contents

- [Overview](#overview)
- [Quick Reference](#quick-reference)
- [Methods](#methods)
- [Global Functions](#global-functions)
- [Examples](#examples)
- [See Also](#see-also)

---

## Overview

The `lc.base64` API provides Base64 encoding and decoding:

- Encode strings to Base64 format
- Decode Base64 strings back to original
- Compatible with browser `btoa()`/`atob()` functions

Common use cases:

- HTTP Basic Authentication
- Encoding binary data in JSON
- Data URI schemes

---

## Quick Reference

| Method/Function | Description |
|-----------------|-------------|
| `lc.base64.encode(data)` | Encode string to Base64 |
| `lc.base64.decode(encoded)` | Decode Base64 to string |
| `btoa(data)` | Global function, same as `encode()` |
| `atob(encoded)` | Global function, same as `decode()` |

---

## Methods

### encode(data)

Encodes a string to Base64 format.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `data` | string | Yes | String to encode |

**Returns**: `string` - Base64 encoded string

```javascript
var encoded = lc.base64.encode("Hello, World!");
// Returns: "SGVsbG8sIFdvcmxkIQ=="

var credentials = lc.base64.encode("username:password");
// Returns: "dXNlcm5hbWU6cGFzc3dvcmQ="
```

---

### decode(encoded)

Decodes a Base64 string back to the original string.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `encoded` | string | Yes | Base64 encoded string |

**Returns**: `string` - Decoded string

```javascript
var decoded = lc.base64.decode("SGVsbG8sIFdvcmxkIQ==");
// Returns: "Hello, World!"

var credentials = lc.base64.decode("dXNlcm5hbWU6cGFzc3dvcmQ=");
// Returns: "username:password"
```

---

## Global Functions

For browser-style compatibility, LazyCurl provides global `btoa()` and `atob()` functions.

### btoa(data)

Browser-compatible function to encode to Base64 ("binary to ASCII").

```javascript
var encoded = btoa("Hello, World!");
// Returns: "SGVsbG8sIFdvcmxkIQ=="
```

### atob(encoded)

Browser-compatible function to decode from Base64 ("ASCII to binary").

```javascript
var decoded = atob("SGVsbG8sIFdvcmxkIQ==");
// Returns: "Hello, World!"
```

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| No argument | Returns empty string `""` |
| Empty string | Returns empty string |
| Invalid Base64 | Returns empty string (no exception) |
| Unicode characters | Encodes UTF-8 bytes |

```javascript
// Edge cases
lc.base64.encode();       // ""
lc.base64.encode("");     // ""
lc.base64.decode("!!!"); // "" (invalid Base64)
```

---

## Examples

### HTTP Basic Authentication

```javascript
// Pre-request: Add Basic Auth header
var username = lc.env.get("api_username");
var password = lc.env.get("api_password");

if (username && password) {
    var credentials = btoa(username + ":" + password);
    lc.request.headers.set("Authorization", "Basic " + credentials);
    console.log("Added Basic Auth for user: " + username);
}
```

### Decode JWT Payload

JWT tokens consist of three Base64-encoded parts separated by dots.

```javascript
// Post-response: Decode JWT token to inspect claims
var data = lc.response.body.json();

if (data && data.access_token) {
    // JWT format: header.payload.signature
    var parts = data.access_token.split(".");

    if (parts.length === 3) {
        // Decode payload (middle part)
        // Note: JWT uses Base64URL, but atob handles most cases
        var payload = atob(parts[1]);
        var claims = JSON.parse(payload);

        console.log("Token subject: " + claims.sub);
        console.log("Token expires: " + new Date(claims.exp * 1000).toISOString());

        // Store user info from token
        if (claims.email) {
            lc.globals.set("user_email", claims.email);
        }
    }
}
```

### Encode JSON for Header

```javascript
// Pre-request: Encode JSON metadata in header
var metadata = {
    client: "LazyCurl",
    version: "1.2.0",
    timestamp: Date.now()
};

var encoded = btoa(JSON.stringify(metadata));
lc.request.headers.set("X-Request-Metadata", encoded);
```

### Decode Base64 Response

```javascript
// Post-response: Handle Base64-encoded response content
var data = lc.response.body.json();

if (data && data.content && data.encoding === "base64") {
    var decoded = atob(data.content);
    console.log("Decoded content: " + decoded);

    // If it's JSON
    try {
        var parsedContent = JSON.parse(decoded);
        lc.globals.set("decoded_content", parsedContent);
    } catch (e) {
        // Plain text
        lc.globals.set("decoded_content", decoded);
    }
}
```

### Data URI Creation

```javascript
// Pre-request: Create data URI for embedded content
var jsonData = JSON.stringify({
    name: "test",
    value: 123
});

var dataUri = "data:application/json;base64," + btoa(jsonData);
lc.request.headers.set("X-Embedded-Data", dataUri);
```

### Compare Encoded Values

```javascript
// Post-response: Verify expected encoded value
var received = lc.response.headers.get("X-Encoded-Token");
var expected = btoa(lc.env.get("expected_token"));

lc.test("Token matches expected value", function() {
    lc.expect(received).toBe(expected);
});
```

---

## See Also

- [lc.crypto](./crypto.md) - Cryptographic hash functions
- [lc.request](./request.md) - HTTP request manipulation
- [Authentication Examples](../examples/authentication.md)

---

*[← lc.crypto](./crypto.md) | [lc.variables →](./variables.md)*
