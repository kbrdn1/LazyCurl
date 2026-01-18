# lc.sendRequest

Send HTTP requests from within scripts for request chaining and dynamic workflows.

**Availability**: Both (pre-request and post-response)

> For the complete API reference, see [Scripting API Reference](../scripting-api-reference.md#console--lcsendrequest).

## Table of Contents

- [Overview](#overview)
- [Quick Reference](#quick-reference)
- [Syntax](#syntax)
- [Options Object](#options-object)
- [Callback Response](#callback-response)
- [Examples](#examples)
- [See Also](#see-also)

---

## Overview

The `lc.sendRequest()` function enables request chaining by sending HTTP requests from within scripts. Use it to:

- Fetch authentication tokens before the main request
- Chain multiple API calls in sequence
- Refresh expired tokens automatically
- Retrieve data needed for the current request

---

## Quick Reference

```javascript
lc.sendRequest(options, callback);
```

- **options**: Object with `url`, `method`, `headers`, `body`
- **callback**: Function receiving `(err, response)`

---

## Syntax

```javascript
lc.sendRequest({
    url: "https://api.example.com/endpoint",
    method: "GET",  // optional, defaults to "GET"
    headers: {},    // optional
    body: ""        // optional
}, function(err, response) {
    if (err) {
        console.error("Request failed:", err);
        return;
    }
    // Process response
});
```

---

## Options Object

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `url` | string | Yes | Target URL (supports `{{variable}}` substitution) |
| `method` | string | No | HTTP method (default: `"GET"`) |
| `headers` | object | No | Key-value pairs for request headers |
| `body` | any | No | Request body (objects are JSON-stringified) |

```javascript
// GET request (minimal)
lc.sendRequest({
    url: "{{base_url}}/users"
}, callback);

// POST request with body
lc.sendRequest({
    url: "{{base_url}}/auth/login",
    method: "POST",
    headers: {
        "Content-Type": "application/json"
    },
    body: JSON.stringify({
        username: "admin",
        password: "secret"
    })
}, callback);
```

---

## Callback Response

The callback receives `(err, response)`:

### Error (`err`)

- `null` if request succeeded
- Error object if request failed

### Response Object

| Property | Type | Description |
|----------|------|-------------|
| `status` | number | HTTP status code |
| `statusText` | string | HTTP status text |
| `time` | number | Response time in milliseconds |
| `headers` | object | Response headers |
| `body.raw` | string | Raw response body |
| `body.json()` | function | Parse body as JSON |

```javascript
lc.sendRequest(options, function(err, response) {
    if (err) {
        console.error("Error:", err);
        return;
    }

    console.log("Status:", response.status);
    console.log("Time:", response.time + "ms");

    var data = response.body.json();
    if (data) {
        console.log("Data:", JSON.stringify(data));
    }
});
```

---

## Examples

### Basic GET Request

```javascript
// Pre-request: Fetch configuration
lc.sendRequest({
    url: "{{base_url}}/config",
    method: "GET"
}, function(err, response) {
    if (err) {
        console.error("Failed to fetch config:", err);
        return;
    }

    if (response.status === 200) {
        var config = response.body.json();
        lc.globals.set("api_config", config);
        console.log("Config loaded successfully");
    }
});
```

### Authentication Token

```javascript
// Pre-request: Get auth token before main request
lc.sendRequest({
    url: "{{auth_url}}/oauth/token",
    method: "POST",
    headers: {
        "Content-Type": "application/x-www-form-urlencoded"
    },
    body: "grant_type=client_credentials" +
          "&client_id=" + lc.env.get("client_id") +
          "&client_secret=" + lc.env.get("client_secret")
}, function(err, response) {
    if (err || response.status !== 200) {
        console.error("Auth failed");
        return;
    }

    var data = response.body.json();
    if (data && data.access_token) {
        lc.env.set("access_token", data.access_token);
        lc.request.headers.set("Authorization", "Bearer " + data.access_token);
        console.log("Token acquired");
    }
});
```

### OAuth2 Token Refresh

```javascript
// Pre-request: Auto-refresh expired token
var tokenExpiry = lc.globals.get("token_expiry") || 0;
var now = Date.now();

if (now >= tokenExpiry) {
    console.log("Token expired, refreshing...");

    lc.sendRequest({
        url: "{{auth_url}}/oauth/token",
        method: "POST",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded"
        },
        body: "grant_type=refresh_token" +
              "&refresh_token=" + lc.env.get("refresh_token") +
              "&client_id=" + lc.env.get("client_id")
    }, function(err, response) {
        if (err || response.status !== 200) {
            console.error("Token refresh failed");
            return;
        }

        var data = response.body.json();
        if (data) {
            lc.env.set("access_token", data.access_token);

            // Store expiry (with 60 second buffer)
            var expiry = now + (data.expires_in * 1000) - 60000;
            lc.globals.set("token_expiry", expiry);

            console.log("Token refreshed, expires in " + data.expires_in + "s");
        }
    });
}

// Add current token to request
var token = lc.env.get("access_token");
if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
}
```

### Request Chaining (Get User, Then Profile)

```javascript
// Post-response: After getting user, fetch their profile
if (lc.response.status === 200) {
    var user = lc.response.body.json();

    if (user && user.profileId) {
        lc.sendRequest({
            url: "{{base_url}}/profiles/" + user.profileId,
            method: "GET",
            headers: {
                "Authorization": "Bearer " + lc.env.get("access_token")
            }
        }, function(err, response) {
            if (err) {
                console.error("Failed to fetch profile");
                return;
            }

            if (response.status === 200) {
                var profile = response.body.json();
                lc.globals.set("user_profile", profile);
                console.log("Profile loaded: " + profile.displayName);
            }
        });
    }
}
```

### Pre-flight Check

```javascript
// Pre-request: Check if API is available
lc.sendRequest({
    url: "{{base_url}}/health",
    method: "GET"
}, function(err, response) {
    if (err || response.status !== 200) {
        console.error("API is not available!");
        console.error("Health check failed:", err || response.status);
    } else {
        console.log("API is healthy (" + response.time + "ms)");
    }
});
```

### Create Resource and Use ID

```javascript
// Pre-request: Create a resource and use its ID
lc.sendRequest({
    url: "{{base_url}}/folders",
    method: "POST",
    headers: {
        "Content-Type": "application/json",
        "Authorization": "Bearer " + lc.env.get("access_token")
    },
    body: JSON.stringify({
        name: "Test Folder " + lc.variables.timestamp()
    })
}, function(err, response) {
    if (err || response.status !== 201) {
        console.error("Failed to create folder");
        return;
    }

    var folder = response.body.json();
    if (folder && folder.id) {
        // Use the folder ID in the current request
        lc.request.url = lc.request.url.replace("{folderId}", folder.id);
        lc.globals.set("test_folder_id", folder.id);
        console.log("Created folder: " + folder.id);
    }
});
```

### Fetch Dynamic Values

```javascript
// Pre-request: Get current user for request body
lc.sendRequest({
    url: "{{base_url}}/me",
    method: "GET",
    headers: {
        "Authorization": "Bearer " + lc.env.get("access_token")
    }
}, function(err, response) {
    if (err || response.status !== 200) {
        console.error("Failed to get current user");
        return;
    }

    var me = response.body.json();
    if (me) {
        // Add current user info to request body
        var body = lc.request.body.json() || {};
        body.createdBy = me.id;
        body.teamId = me.teamId;
        lc.request.body.set(JSON.stringify(body));
        console.log("Added user context to request");
    }
});
```

---

## See Also

- [lc.env](./env.md) - Environment variables
- [lc.globals](./env.md#lcglobals) - Session variables
- [Request Chaining Examples](../examples/request-chaining.md)
- [Authentication Examples](../examples/authentication.md)

---

*[← lc.info](./info.md) | [console →](./console.md)*
