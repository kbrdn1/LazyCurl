# Authentication Examples

Practical examples for handling authentication in LazyCurl scripts.

## Table of Contents

- [Overview](#overview)
- [Bearer Token](#bearer-token)
- [Basic Authentication](#basic-authentication)
- [API Key Authentication](#api-key-authentication)
- [OAuth2 Client Credentials](#oauth2-client-credentials)
- [Token Storage and Refresh](#token-storage-and-refresh)
- [Common Patterns](#common-patterns)
- [See Also](#see-also)

---

## Overview

LazyCurl scripts can handle various authentication methods:

| Method | Use Case | Script Type |
|--------|----------|-------------|
| Bearer Token | OAuth2, JWT | Pre-request |
| Basic Auth | Legacy APIs | Pre-request |
| API Key | Simple APIs | Pre-request |
| OAuth2 Flow | Token fetch/refresh | Pre-request |
| Token Storage | Save tokens | Post-response |

---

## Bearer Token

### From Environment Variable

**Use Case**: Add a Bearer token stored in environment variables.

**Script Type**: Pre-request

```javascript
// Add Bearer token from environment
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

> **Tip**: Store tokens in `lc.globals` for immediate use within a session, and in `lc.env` for persistence across sessions.

---

## Basic Authentication

**Use Case**: Add HTTP Basic Authentication header using credentials from environment.

**Script Type**: Pre-request

```javascript
// Build Basic Auth header from credentials
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

### With Validation

```javascript
// Pre-request: Basic auth with validation
var username = lc.env.get("api_username");
var password = lc.env.get("api_password");

if (!username) {
    console.error("Missing api_username in environment");
}
if (!password) {
    console.error("Missing api_password in environment");
}

if (username && password) {
    var credentials = lc.base64.encode(username + ":" + password);
    lc.request.headers.set("Authorization", "Basic " + credentials);
    console.info("Basic auth configured for: " + username);
}
```

---

## API Key Authentication

**Use Case**: Add API key in header or query parameter.

**Script Type**: Pre-request

### Header Authentication

```javascript
// Add API key in header (most common)
var apiKey = lc.env.get("api_key");

if (apiKey) {
    lc.request.headers.set("X-API-Key", apiKey);
    console.log("Added API key header");
} else {
    console.error("No api_key found in environment");
}
```

### Query Parameter Authentication

```javascript
// Add API key as query parameter
var apiKey = lc.env.get("api_key");

if (apiKey) {
    var separator = lc.request.url.includes("?") ? "&" : "?";
    lc.request.url = lc.request.url + separator + "api_key=" + apiKey;
    console.log("Added API key to URL");
}
```

### Configurable Location

```javascript
// Pre-request: Flexible API key placement
var apiKey = lc.env.get("api_key");
var keyLocation = lc.env.get("api_key_location") || "header";
var keyName = lc.env.get("api_key_name") || "X-API-Key";

if (apiKey) {
    if (keyLocation === "header") {
        lc.request.headers.set(keyName, apiKey);
    } else if (keyLocation === "query") {
        var separator = lc.request.url.includes("?") ? "&" : "?";
        lc.request.url = lc.request.url + separator + keyName + "=" + apiKey;
    }
    console.log("Added API key via " + keyLocation);
}
```

---

## OAuth2 Client Credentials

**Use Case**: Fetch OAuth2 token using client credentials grant.

**Script Type**: Pre-request

```javascript
// Check if we have a valid token
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
            console.error("Failed to fetch token:", err || response.status);
            return;
        }

        var data = response.body.json();
        if (data && data.access_token) {
            lc.globals.set("oauth_token", data.access_token);
            // Set expiry with 60 second buffer
            lc.globals.set("oauth_expiry", now + (data.expires_in * 1000) - 60000);
            console.log("New token acquired, expires in " + data.expires_in + "s");
        }
    });
}

// Use the token
var currentToken = lc.globals.get("oauth_token");
if (currentToken) {
    lc.request.headers.set("Authorization", "Bearer " + currentToken);
}
```

### With Refresh Token

```javascript
// Pre-request: OAuth2 with refresh token support
var accessToken = lc.env.get("access_token");
var refreshToken = lc.env.get("refresh_token");
var expiry = lc.globals.get("token_expiry") || 0;
var now = Date.now();

// Check if token needs refresh
if (now >= expiry && refreshToken) {
    console.log("Access token expired, using refresh token...");

    lc.sendRequest({
        url: lc.env.get("oauth_url") + "/oauth/token",
        method: "POST",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded"
        },
        body: "grant_type=refresh_token" +
              "&refresh_token=" + refreshToken +
              "&client_id=" + lc.env.get("client_id")
    }, function(err, response) {
        if (err || response.status !== 200) {
            console.error("Token refresh failed");
            return;
        }

        var data = response.body.json();
        if (data) {
            lc.env.set("access_token", data.access_token);
            if (data.refresh_token) {
                lc.env.set("refresh_token", data.refresh_token);
            }
            lc.globals.set("token_expiry", now + (data.expires_in * 1000) - 60000);
            console.log("Token refreshed successfully");
        }
    });
}

// Add token to request
var token = lc.env.get("access_token");
if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
}
```

---

## Token Storage and Refresh

### Store Token from Login Response

**Use Case**: Save authentication tokens from a login response.

**Script Type**: Post-response (on login endpoint)

```javascript
// Store authentication tokens from login response
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
            console.log("Refresh token stored");
        }

        // Calculate and store expiry time
        if (data.expires_in) {
            var expiresAt = Date.now() + (data.expires_in * 1000);
            lc.globals.set("token_expires_at", expiresAt);
            console.log("Token expires: " + new Date(expiresAt).toISOString());
        }

        // Store user info if provided
        if (data.user) {
            lc.globals.set("current_user", data.user);
            lc.env.set("user_id", data.user.id.toString());
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

### Clear Auth on 401

```javascript
// Post-response: Clear auth state on unauthorized response
if (lc.response.status === 401) {
    console.warn("Unauthorized - clearing auth state");

    // Clear tokens
    lc.env.unset("access_token");
    lc.env.unset("refresh_token");
    lc.globals.unset("access_token");
    lc.globals.unset("token_expires_at");
    lc.globals.unset("current_user");

    console.log("Please login again");
}
```

---

## Common Patterns

### Ensure Authentication Exists

```javascript
// Pre-request: Ensure auth is configured
function ensureAuth() {
    if (!lc.env.has("access_token") && !lc.globals.has("access_token")) {
        console.error("No authentication configured!");
        console.error("Please run the Login request first.");
        return false;
    }
    return true;
}

if (ensureAuth()) {
    var token = lc.globals.get("access_token") || lc.env.get("access_token");
    lc.request.headers.set("Authorization", "Bearer " + token);
}
```

### Environment-Based Auth Selection

```javascript
// Pre-request: Different auth for different environments
var env = lc.info.environmentName || "development";

switch (env) {
    case "production":
        // Use OAuth2 for production
        var prodToken = lc.env.get("prod_token");
        if (prodToken) {
            lc.request.headers.set("Authorization", "Bearer " + prodToken);
        }
        break;

    case "staging":
        // Use API key for staging
        var stagingKey = lc.env.get("staging_api_key");
        if (stagingKey) {
            lc.request.headers.set("X-API-Key", stagingKey);
        }
        break;

    default:
        // Use Basic Auth for development
        var devUser = lc.env.get("dev_username") || "dev";
        var devPass = lc.env.get("dev_password") || "dev123";
        lc.request.headers.set("Authorization", "Basic " + btoa(devUser + ":" + devPass));
}

console.log("Auth configured for environment: " + env);
```

### Skip Auth for Public Endpoints

```javascript
// Pre-request: Skip auth for public endpoints
var publicPaths = ["/health", "/status", "/public"];
var url = lc.request.url;

var isPublic = publicPaths.some(function(path) {
    return url.includes(path);
});

if (!isPublic) {
    var token = lc.env.get("access_token");
    if (token) {
        lc.request.headers.set("Authorization", "Bearer " + token);
    }
} else {
    console.log("Skipping auth for public endpoint");
}
```

### Request Signing

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

---

## See Also

- [lc.env](../api/env.md) - Environment variables
- [lc.globals](../api/env.md#lcglobals) - Session variables
- [lc.crypto](../api/crypto.md) - Cryptographic functions
- [lc.base64](../api/base64.md) - Base64 encoding
- [lc.sendRequest](../api/sendrequest.md) - Request chaining
- [Request Chaining Examples](./request-chaining.md)

---

*[← Examples Overview](../api/overview.md#api-modules) | [Testing Examples →](./testing.md)*
