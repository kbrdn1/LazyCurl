# lc.base64

Base64 encoding and decoding utilities. Also available as global `btoa()` and `atob()`.

**Availability**: Both (pre-request and post-response)

## Quick Reference

| Method | Description |
|--------|-------------|
| `lc.base64.encode(data)` | Encode string to Base64 |
| `lc.base64.decode(data)` | Decode Base64 to string |
| `btoa(data)` | Global alias for encode |
| `atob(data)` | Global alias for decode |

## Methods

### encode(data)

Encodes a string to Base64.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `data` | string | Yes | The string to encode |

**Returns**: `string` - Base64 encoded string

```javascript
var encoded = lc.base64.encode("hello world");
// "aGVsbG8gd29ybGQ="

var jsonEncoded = lc.base64.encode(JSON.stringify({ user: "admin" }));
```

### decode(data)

Decodes a Base64 string.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `data` | string | Yes | The Base64 string to decode |

**Returns**: `string` - Decoded string

```javascript
var decoded = lc.base64.decode("aGVsbG8gd29ybGQ=");
// "hello world"
```

## Global Aliases

### btoa(data)

Browser-compatible alias for `lc.base64.encode()`.

```javascript
var encoded = btoa("hello world");
// "aGVsbG8gd29ybGQ="
```

### atob(data)

Browser-compatible alias for `lc.base64.decode()`.

```javascript
var decoded = atob("aGVsbG8gd29ybGQ=");
// "hello world"
```

## Examples

### Basic Authentication

```javascript
// Pre-request: Add Basic Auth header
var username = lc.env.get("api_username");
var password = lc.env.get("api_password");

if (username && password) {
    var credentials = lc.base64.encode(username + ":" + password);
    lc.request.headers.set("Authorization", "Basic " + credentials);
    console.log("Added Basic Auth for user: " + username);
}
```

### Encode Request Data

```javascript
// Pre-request: Encode payload
var payload = {
    timestamp: Date.now(),
    data: "sensitive information"
};

var encodedPayload = lc.base64.encode(JSON.stringify(payload));
lc.request.body.set(JSON.stringify({ encoded: encodedPayload }));
```

### Decode Response Data

```javascript
// Post-response: Decode Base64 field
var data = lc.response.body.json();

if (data && data.encodedMessage) {
    var decoded = lc.base64.decode(data.encodedMessage);
    console.log("Decoded message: " + decoded);

    // Parse if it's JSON
    try {
        var parsedMessage = JSON.parse(decoded);
        lc.globals.set("decoded_message", parsedMessage);
    } catch (e) {
        lc.globals.set("decoded_message", decoded);
    }
}
```

### JWT Token Inspection

```javascript
// Post-response: Decode JWT payload (for debugging)
var data = lc.response.body.json();

if (data && data.access_token) {
    var parts = data.access_token.split(".");
    if (parts.length === 3) {
        // Decode payload (middle part)
        var payload = lc.base64.decode(parts[1]);
        var claims = JSON.parse(payload);

        console.log("Token subject: " + claims.sub);
        console.log("Token expires: " + new Date(claims.exp * 1000).toISOString());

        // Store expiry for later use
        lc.globals.set("token_expires_at", claims.exp * 1000);
    }
}
```

### Using Browser-Compatible Aliases

```javascript
// Works like in browser JavaScript
var encoded = btoa("username:password");
var decoded = atob(encoded);

lc.request.headers.set("Authorization", "Basic " + btoa("admin:secret"));
```

## See Also

- [lc.crypto](API-Crypto) - Cryptographic functions
- [Authentication Examples](Examples-Auth)

---

*[lc.crypto](API-Crypto) | [lc.variables](API-Variables)*
