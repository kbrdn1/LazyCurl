# lc.sendRequest

Make HTTP requests from within scripts for request chaining.

**Availability**: Both (pre-request and post-response)

## Syntax

```javascript
lc.sendRequest(options, callback)
```

### Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `options` | object | Yes | Request configuration |
| `callback` | function | Yes | Callback with (error, response) |

### Options Object

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `url` | string | Yes | Request URL |
| `method` | string | No | HTTP method (default: "GET") |
| `headers` | object | No | Request headers |
| `body` | string | No | Request body |

### Callback Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `error` | Error\|null | Error object if request failed |
| `response` | object | Response with status, headers, body |

## Basic Usage

```javascript
lc.sendRequest({
    url: "https://api.example.com/data",
    method: "GET",
    headers: {
        "Authorization": "Bearer " + lc.env.get("token")
    }
}, function(err, response) {
    if (err) {
        console.error("Request failed: " + err.message);
        return;
    }

    console.log("Status: " + response.status);
    var data = response.body.json();
    console.log("Data: " + JSON.stringify(data));
});
```

## Examples

### Pre-fetch Authentication Token

```javascript
// Pre-request: Get fresh token before main request
var token = lc.globals.get("access_token");
var expiry = lc.globals.get("token_expiry") || 0;

if (!token || Date.now() >= expiry) {
    console.log("Fetching new token...");

    lc.sendRequest({
        url: lc.env.get("auth_url") + "/oauth/token",
        method: "POST",
        headers: {
            "Content-Type": "application/x-www-form-urlencoded"
        },
        body: "grant_type=client_credentials" +
              "&client_id=" + lc.env.get("client_id") +
              "&client_secret=" + lc.env.get("client_secret")
    }, function(err, response) {
        if (err || response.status !== 200) {
            console.error("Token fetch failed");
            return;
        }

        var data = response.body.json();
        if (data && data.access_token) {
            lc.globals.set("access_token", data.access_token);
            lc.globals.set("token_expiry", Date.now() + (data.expires_in * 1000));
            console.log("Token acquired");
        }
    });
}

// Use the token
token = lc.globals.get("access_token");
if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
}
```

### Create Resource Before Main Request

```javascript
// Pre-request: Ensure folder exists before uploading
var folderId = lc.globals.get("upload_folder_id");

if (!folderId) {
    console.log("Creating upload folder...");

    lc.sendRequest({
        url: lc.env.get("base_url") + "/folders",
        method: "POST",
        headers: {
            "Authorization": "Bearer " + lc.env.get("token"),
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            name: "Uploads " + lc.variables.timestamp()
        })
    }, function(err, response) {
        if (err || response.status !== 201) {
            console.error("Failed to create folder");
            return;
        }

        var folder = response.body.json();
        lc.globals.set("upload_folder_id", folder.id);
        console.log("Created folder: " + folder.id);
    });
}

// Add folder ID to request body
folderId = lc.globals.get("upload_folder_id");
if (folderId) {
    var body = lc.request.body.json() || {};
    body.folderId = folderId;
    lc.request.body.set(JSON.stringify(body));
}
```

### Fetch Configuration

```javascript
// Pre-request: Load API config before request
lc.sendRequest({
    url: lc.env.get("base_url") + "/config",
    method: "GET",
    headers: {
        "Authorization": "Bearer " + lc.env.get("token")
    }
}, function(err, response) {
    if (err || response.status !== 200) {
        console.error("Failed to fetch config");
        return;
    }

    var config = response.body.json();

    // Apply config to current request
    if (config.apiVersion) {
        lc.request.headers.set("X-API-Version", config.apiVersion);
    }

    if (config.features && config.features.useV2) {
        lc.request.url = lc.request.url.replace("/v1/", "/v2/");
    }

    lc.globals.set("api_config", config);
    console.log("Config loaded: v" + config.apiVersion);
});
```

### Token Refresh

```javascript
// Pre-request: Auto-refresh expired token
var token = lc.env.get("access_token");
var expiry = lc.globals.get("token_expires") || 0;
var now = Date.now();

// Refresh if expired or expiring in 60 seconds
if (!token || now >= expiry - 60000) {
    var refreshToken = lc.env.get("refresh_token");

    if (refreshToken) {
        console.log("Refreshing token...");

        lc.sendRequest({
            url: lc.env.get("auth_url") + "/token/refresh",
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({
                refresh_token: refreshToken
            })
        }, function(err, response) {
            if (err || response.status !== 200) {
                console.error("Token refresh failed");
                return;
            }

            var data = response.body.json();
            if (data && data.access_token) {
                lc.env.set("access_token", data.access_token);
                lc.globals.set("token_expires", now + (data.expires_in * 1000));
                console.log("Token refreshed");
            }
        });
    }
}

// Use token
token = lc.env.get("access_token");
if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
}
```

## Important Notes

- `lc.sendRequest()` is **synchronous** in effect - the callback completes before the main request
- Use for setup tasks like authentication, not for parallel requests
- Keep chained requests minimal to avoid slow script execution
- Store results in `lc.globals` for immediate use or `lc.env` for persistence

## See Also

- [lc.env](API-Env) - Environment variables
- [lc.globals](API-Env#lcglobals) - Session variables
- [Request Chaining Examples](Examples-Chaining)
- [Authentication Examples](Examples-Auth)

---

*[lc.info](API-Info) | [console](API-Console)*
