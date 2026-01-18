# lc.crypto

Cryptographic hash functions and HMAC operations.

**Availability**: Both (pre-request and post-response)

## Quick Reference

| Method | Description |
|--------|-------------|
| `md5(data)` | MD5 hash (hex) |
| `sha1(data)` | SHA-1 hash (hex) |
| `sha256(data)` | SHA-256 hash (hex) |
| `sha512(data)` | SHA-512 hash (hex) |
| `hmacSha1(data, key)` | HMAC-SHA1 signature (hex) |
| `hmacSha256(data, key)` | HMAC-SHA256 signature (hex) |
| `hmacSha512(data, key)` | HMAC-SHA512 signature (hex) |

## Hash Functions

### md5(data)

Computes MD5 hash. Returns hex string.

```javascript
var hash = lc.crypto.md5("hello world");
// "5eb63bbbe01eeed093cb22bb8f5acdc3"
```

### sha1(data)

Computes SHA-1 hash. Returns hex string.

```javascript
var hash = lc.crypto.sha1("hello world");
// "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed"
```

### sha256(data)

Computes SHA-256 hash. Returns hex string.

```javascript
var hash = lc.crypto.sha256("hello world");
// "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
```

### sha512(data)

Computes SHA-512 hash. Returns hex string.

```javascript
var hash = lc.crypto.sha512("hello world");
```

## HMAC Functions

### hmacSha1(data, key)

Computes HMAC-SHA1 signature. Returns hex string.

```javascript
var signature = lc.crypto.hmacSha1("message", "secret-key");
```

### hmacSha256(data, key)

Computes HMAC-SHA256 signature. Returns hex string.

```javascript
var signature = lc.crypto.hmacSha256("message", "secret-key");
```

### hmacSha512(data, key)

Computes HMAC-SHA512 signature. Returns hex string.

```javascript
var signature = lc.crypto.hmacSha512("message", "secret-key");
```

## Examples

### API Request Signing

```javascript
// Pre-request: Sign API request
var timestamp = Math.floor(Date.now() / 1000).toString();
var body = lc.request.body.raw() || "";
var secret = lc.env.get("api_secret");

// Build string to sign
var stringToSign = lc.request.method + "\n" +
                   lc.request.url + "\n" +
                   timestamp + "\n" +
                   body;

var signature = lc.crypto.hmacSha256(stringToSign, secret);

lc.request.headers.set("X-Timestamp", timestamp);
lc.request.headers.set("X-Signature", signature);
console.log("Request signed with HMAC-SHA256");
```

### AWS-Style Signature

```javascript
// Pre-request: AWS Signature Version 4 (simplified)
var date = new Date().toISOString().slice(0, 10).replace(/-/g, "");
var region = lc.env.get("aws_region") || "us-east-1";
var service = "execute-api";
var secretKey = lc.env.get("aws_secret_key");

// Create signing key
var kDate = lc.crypto.hmacSha256(date, "AWS4" + secretKey);
var kRegion = lc.crypto.hmacSha256(region, kDate);
var kService = lc.crypto.hmacSha256(service, kRegion);
var signingKey = lc.crypto.hmacSha256("aws4_request", kService);

// Sign request (simplified)
var stringToSign = "AWS4-HMAC-SHA256\n" + date + "\n" + region + "/" + service;
var signature = lc.crypto.hmacSha256(stringToSign, signingKey);

lc.request.headers.set("X-Amz-Date", date);
lc.request.headers.set("Authorization", "AWS4-HMAC-SHA256 Signature=" + signature);
```

### Content Hashing

```javascript
// Pre-request: Add content hash header
var body = lc.request.body.raw() || "";
var contentHash = lc.crypto.sha256(body);

lc.request.headers.set("X-Content-SHA256", contentHash);
console.log("Content hash: " + contentHash.substring(0, 16) + "...");
```

### Webhook Signature Verification

```javascript
// Post-response: Verify webhook signature
var payload = lc.response.body.raw();
var receivedSig = lc.response.headers.get("X-Webhook-Signature");
var secret = lc.env.get("webhook_secret");

var expectedSig = lc.crypto.hmacSha256(payload, secret);

lc.test("Webhook signature is valid", function() {
    lc.expect(receivedSig).toBe(expectedSig);
});
```

### Password Hashing (for testing)

```javascript
// Pre-request: Hash password before sending
var password = lc.env.get("test_password");
var salt = lc.env.get("password_salt");

var hashedPassword = lc.crypto.sha256(salt + password);

var body = lc.request.body.json() || {};
body.password = hashedPassword;
lc.request.body.set(JSON.stringify(body));
```

## See Also

- [lc.base64](API-Base64) - Base64 encoding
- [lc.request](API-Request) - HTTP request manipulation
- [Authentication Examples](Examples-Auth)

---

*[lc.cookies](API-Cookies) | [lc.base64](API-Base64)*
