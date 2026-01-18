# lc.crypto

Cryptographic hash functions and HMAC operations. All functions return **lowercase hex-encoded strings**.

**Availability**: Both (pre-request and post-response)

> For the complete API reference, see [Scripting API Reference](../scripting-api-reference.md#lccrypto).

## Table of Contents

- [Overview](#overview)
- [Quick Reference](#quick-reference)
- [Hash Functions](#hash-functions)
- [HMAC Functions](#hmac-functions)
- [Examples](#examples)
- [See Also](#see-also)

---

## Overview

The `lc.crypto` API provides cryptographic functions for:

- Hashing data (MD5, SHA-1, SHA-256, SHA-512)
- Creating HMAC signatures for API authentication
- Webhook signature verification

All functions return **lowercase hexadecimal strings**.

---

## Quick Reference

### Hash Functions

| Function | Output Length | Description |
|----------|---------------|-------------|
| `md5(data)` | 32 chars | MD5 hash (legacy, not secure) |
| `sha1(data)` | 40 chars | SHA-1 hash (legacy) |
| `sha256(data)` | 64 chars | SHA-256 hash (recommended) |
| `sha512(data)` | 128 chars | SHA-512 hash |

### HMAC Functions

| Function | Output Length | Description |
|----------|---------------|-------------|
| `hmacSha1(data, secret)` | 40 chars | HMAC-SHA1 (OAuth 1.0) |
| `hmacSha256(data, secret)` | 64 chars | HMAC-SHA256 (recommended) |
| `hmacSha512(data, secret)` | 128 chars | HMAC-SHA512 |

---

## Hash Functions

### md5(data)

Computes MD5 hash. **Note**: MD5 is considered cryptographically broken and should only be used for legacy compatibility.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `data` | string | Yes | Data to hash |

**Returns**: `string` - 32-character lowercase hex string

```javascript
var hash = lc.crypto.md5("hello world");
// Returns: "5eb63bbbe01eeed093cb22bb8f5acdc3"
```

---

### sha1(data)

Computes SHA-1 hash. **Note**: SHA-1 is deprecated for security purposes but still used in some APIs.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `data` | string | Yes | Data to hash |

**Returns**: `string` - 40-character lowercase hex string

```javascript
var hash = lc.crypto.sha1("hello world");
// Returns: "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"
```

---

### sha256(data)

Computes SHA-256 hash. **Recommended** for general-purpose hashing.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `data` | string | Yes | Data to hash |

**Returns**: `string` - 64-character lowercase hex string

```javascript
var hash = lc.crypto.sha256("hello world");
// Returns: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
```

---

### sha512(data)

Computes SHA-512 hash. Use for applications requiring maximum security.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `data` | string | Yes | Data to hash |

**Returns**: `string` - 128-character lowercase hex string

```javascript
var hash = lc.crypto.sha512("hello world");
// Returns: "309ecc489c12d6eb4cc40f50c902f2b4d0ed77ee511a7c7a9bcd3ca86d4cd86f989dd35bc5ff499670da34255b45b0cfd830e81f605dcf7dc5542e93ae9cd76f"
```

---

## HMAC Functions

HMAC (Hash-based Message Authentication Code) functions combine a secret key with the hash function for message authentication.

### hmacSha1(data, secret)

Computes HMAC-SHA1. Used in OAuth 1.0 and some legacy APIs.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `data` | string | Yes | Data to authenticate |
| `secret` | string | Yes | Secret key |

**Returns**: `string` - 40-character lowercase hex string

```javascript
var signature = lc.crypto.hmacSha1("message", "secret");
// Returns: "0caf649feee4953d87bf903ac1176c45e028df16"
```

---

### hmacSha256(data, secret)

Computes HMAC-SHA256. **Recommended** for API request signing.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `data` | string | Yes | Data to authenticate |
| `secret` | string | Yes | Secret key |

**Returns**: `string` - 64-character lowercase hex string

```javascript
var signature = lc.crypto.hmacSha256("message", "secret");
// Returns: "6e9ef29b75fffc5b7abae527d58fdadb2fe42e7219011976917343065f58ed4a"
```

---

### hmacSha512(data, secret)

Computes HMAC-SHA512. Maximum security HMAC.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `data` | string | Yes | Data to authenticate |
| `secret` | string | Yes | Secret key |

**Returns**: `string` - 128-character lowercase hex string

```javascript
var signature = lc.crypto.hmacSha512("message", "secret");
```

---

## Examples

### API Request Signing

Many APIs require signed requests for authentication.

```javascript
// Pre-request: Sign the request
var timestamp = Math.floor(Date.now() / 1000).toString();
var body = lc.request.body.raw() || "";
var secret = lc.env.get("api_secret");

// Create signature from timestamp + body
var signature = lc.crypto.hmacSha256(timestamp + body, secret);

lc.request.headers.set("X-Timestamp", timestamp);
lc.request.headers.set("X-Signature", signature);

console.log("Request signed with timestamp: " + timestamp);
```

### AWS-Style Signature

```javascript
// Pre-request: AWS-style request signing (simplified)
var method = lc.request.method;
var path = new URL(lc.request.url).pathname;
var timestamp = new Date().toISOString().replace(/[:-]/g, "").split(".")[0] + "Z";
var date = timestamp.substring(0, 8);

// Create canonical request
var canonicalRequest = method + "\n" + path + "\n" + timestamp;
var stringToSign = "AWS4-HMAC-SHA256\n" + timestamp + "\n" + lc.crypto.sha256(canonicalRequest);

// Sign
var secret = lc.env.get("aws_secret_key");
var kDate = lc.crypto.hmacSha256(date, "AWS4" + secret);
var signature = lc.crypto.hmacSha256(stringToSign, kDate);

lc.request.headers.set("X-Amz-Date", timestamp);
lc.request.headers.set("Authorization", "AWS4-HMAC-SHA256 Signature=" + signature);
```

### Webhook Signature Verification

```javascript
// Post-response: Verify webhook signature
var payload = lc.response.body.raw();
var secret = lc.env.get("webhook_secret");

// Calculate expected signature
var expected = "sha256=" + lc.crypto.hmacSha256(payload, secret);

// Get received signature
var received = lc.response.headers.get("X-Hub-Signature-256");

lc.test("Webhook signature is valid", function() {
    lc.expect(received).toBe(expected);
});

if (received === expected) {
    console.log("Webhook signature verified");
} else {
    console.error("Signature mismatch!");
    console.log("Expected: " + expected);
    console.log("Received: " + received);
}
```

### Content Hash Verification

```javascript
// Post-response: Verify content integrity
var body = lc.response.body.raw();
var expectedHash = lc.response.headers.get("X-Content-SHA256");

if (expectedHash) {
    var actualHash = lc.crypto.sha256(body);

    lc.test("Content integrity verified", function() {
        lc.expect(actualHash).toBe(expectedHash);
    });
}
```

### Generate Idempotency Key

```javascript
// Pre-request: Generate unique request identifier
var requestData = lc.request.method + lc.request.url + (lc.request.body.raw() || "");
var idempotencyKey = lc.crypto.sha256(requestData + Date.now());

lc.request.headers.set("Idempotency-Key", idempotencyKey);
console.log("Idempotency key: " + idempotencyKey.substring(0, 16) + "...");
```

### OAuth 1.0 Signature (Legacy)

```javascript
// Pre-request: OAuth 1.0 HMAC-SHA1 signature
var consumerSecret = lc.env.get("consumer_secret");
var tokenSecret = lc.env.get("token_secret") || "";
var signingKey = consumerSecret + "&" + tokenSecret;

// Base string (simplified)
var timestamp = Math.floor(Date.now() / 1000).toString();
var nonce = lc.variables.randomHex(16);
var baseString = lc.request.method + "&" + encodeURIComponent(lc.request.url) + "&" + timestamp;

var signature = lc.crypto.hmacSha1(baseString, signingKey);

lc.request.headers.set("Authorization",
    'OAuth oauth_signature="' + encodeURIComponent(signature) + '",' +
    'oauth_timestamp="' + timestamp + '",' +
    'oauth_nonce="' + nonce + '"'
);
```

---

## See Also

- [lc.base64](./base64.md) - Base64 encoding/decoding
- [lc.request](./request.md) - HTTP request manipulation
- [Authentication Examples](../examples/authentication.md)

---

*[← lc.cookies](./cookies.md) | [lc.base64 →](./base64.md)*
