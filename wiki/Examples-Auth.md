# Authentication Examples

Practical examples for handling authentication in LazyCurl scripts.

## Overview

| Method | Use Case | Script Type |
|--------|----------|-------------|
| Bearer Token | OAuth2, JWT | Pre-request |
| Basic Auth | Legacy APIs | Pre-request |
| API Key | Simple APIs | Pre-request |
| OAuth2 Flow | Token fetch/refresh | Pre-request |
| Token Storage | Save tokens | Post-response |

## Bearer Token

### From Environment Variable

```javascript
// Pre-request: Add Bearer token from environment
var token = lc.env.get("access_token");

if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
    console.log("Added Bearer token to request");
} else {
    console.warn("No access_token found in environment");
}
```

### From Session (Globals)

```javascript
// Pre-request: Try globals first, then environment
var token = lc.globals.get("access_token") || lc.env.get("access_token");

if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
} else {
    console.error("No token available - please login first");
}
```

## Basic Authentication

```javascript
// Pre-request: Build Basic Auth header
var username = lc.env.get("api_username");
var password = lc.env.get("api_password");

if (username && password) {
    var credentials = btoa(username + ":" + password);
    lc.request.headers.set("Authorization", "Basic " + credentials);
    console.log("Added Basic Auth for user: " + username);
} else {
    console.error("Missing credentials in environment");
}
```

## API Key Authentication

### Header Authentication

```javascript
// Pre-request: Add API key in header
var apiKey = lc.env.get("api_key");

if (apiKey) {
    lc.request.headers.set("X-API-Key", apiKey);
    console.log("Added API key header");
}
```

### Query Parameter

```javascript
// Pre-request: Add API key as query parameter
var apiKey = lc.env.get("api_key");

if (apiKey) {
    var separator = lc.request.url.includes("?") ? "&" : "?";
    lc.request.url = lc.request.url + separator + "api_key=" + apiKey;
}
```

## OAuth2 Client Credentials

```javascript
// Pre-request: Fetch OAuth2 token
var token = lc.globals.get("oauth_token");
var expiry = lc.globals.get("oauth_expiry") || 0;
var now = Date.now();

if (!token || now >= expiry) {
    console.log("Token missing or expired, fetching new token...");

    lc.sendRequest({
        url: lc.env.get("oauth_url") + "/oauth/token",
        method: "POST",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded"
        },
        body: "grant_type=client_credentials" +
              "&client_id=" + lc.env.get("client_id") +
              "&client_secret=" + lc.env.get("client_secret")
    }, function(err, response) {
        if (err || response.status !== 200) {
            console.error("Failed to fetch token");
            return;
        }

        var data = response.body.json();
        if (data && data.access_token) {
            lc.globals.set("oauth_token", data.access_token);
            lc.globals.set("oauth_expiry", now + (data.expires_in * 1000) - 60000);
            console.log("New token acquired");
        }
    });
}

// Use the token
var currentToken = lc.globals.get("oauth_token");
if (currentToken) {
    lc.request.headers.set("Authorization", "Bearer " + currentToken);
}
```

## Token Storage from Login

```javascript
// Post-response: Store authentication tokens
if (lc.response.status === 200) {
    var data = lc.response.body.json();

    if (data && data.access_token) {
        // Store access token
        lc.env.set("access_token", data.access_token);
        lc.globals.set("access_token", data.access_token);
        console.log("Access token stored");

        // Store refresh token if provided
        if (data.refresh_token) {
            lc.env.set("refresh_token", data.refresh_token);
        }

        // Calculate and store expiry time
        if (data.expires_in) {
            var expiresAt = Date.now() + (data.expires_in * 1000);
            lc.globals.set("token_expires_at", expiresAt);
        }
    }
}

// Validate login response
lc.test("Login successful", function() {
    lc.expect(lc.response.status).toBe(200);
});

lc.test("Response contains access_token", function() {
    var data = lc.response.body.json();
    lc.expect(data).toHaveProperty("access_token");
});
```

## Clear Auth on 401

```javascript
// Post-response: Clear auth state on unauthorized
if (lc.response.status === 401) {
    console.warn("Unauthorized - clearing auth state");

    lc.env.unset("access_token");
    lc.env.unset("refresh_token");
    lc.globals.unset("access_token");
    lc.globals.unset("token_expires_at");

    console.log("Please login again");
}
```

## Environment-Based Auth

```javascript
// Pre-request: Different auth per environment
var env = lc.info.environmentName || "development";

switch (env) {
    case "production":
        var prodToken = lc.env.get("prod_token");
        if (prodToken) {
            lc.request.headers.set("Authorization", "Bearer " + prodToken);
        }
        break;

    case "staging":
        var stagingKey = lc.env.get("staging_api_key");
        if (stagingKey) {
            lc.request.headers.set("X-API-Key", stagingKey);
        }
        break;

    default:
        var devUser = lc.env.get("dev_username") || "dev";
        var devPass = lc.env.get("dev_password") || "dev123";
        lc.request.headers.set("Authorization", "Basic " + btoa(devUser + ":" + devPass));
}

console.log("Auth configured for: " + env);
```

## Request Signing

```javascript
// Pre-request: Sign requests with HMAC
var timestamp = Math.floor(Date.now() / 1000).toString();
var body = lc.request.body.raw() || "";
var secret = lc.env.get("api_secret");

var dataToSign = lc.request.method + "\n" +
                 lc.request.url + "\n" +
                 timestamp + "\n" +
                 body;

var signature = lc.crypto.hmacSha256(dataToSign, secret);

lc.request.headers.set("X-Timestamp", timestamp);
lc.request.headers.set("X-Signature", signature);
console.log("Request signed with HMAC-SHA256");
```

## See Also

- [lc.env](API-Env) - Environment variables
- [lc.crypto](API-Crypto) - Cryptographic functions
- [lc.base64](API-Base64) - Base64 encoding
- [lc.sendRequest](API-SendRequest) - Request chaining
- [Request Chaining Examples](Examples-Chaining)

---

*[API Overview](API-Overview) | [Testing Examples](Examples-Testing)*
