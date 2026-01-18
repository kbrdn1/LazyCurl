# Request Chaining Examples

Practical examples for chaining multiple requests and passing data between them.

## Table of Contents

- [Overview](#overview)
- [Basic Chaining Patterns](#basic-chaining-patterns)
- [Authentication Flows](#authentication-flows)
- [CRUD Operations](#crud-operations)
- [Data Extraction Patterns](#data-extraction-patterns)
- [Error Handling](#error-handling)
- [Advanced Patterns](#advanced-patterns)
- [See Also](#see-also)

---

## Overview

Request chaining allows you to:

- Use data from one request in another
- Build authentication flows (login → use token)
- Create resources and use their IDs
- Implement complex workflows

**Key Tools:**

- `lc.env` - Persist data across sessions
- `lc.globals` - Share data within session
- `lc.sendRequest()` - Make requests within scripts

---

## Basic Chaining Patterns

### Store Data for Next Request

**Request 1 - Post-response**: Store user ID

```javascript
// Post-response: Save user ID from response
if (lc.response.status === 200) {
    var user = lc.response.body.json();

    if (user && user.id) {
        lc.globals.set("user_id", user.id);
        lc.env.set("user_id", user.id.toString());
        console.log("Stored user ID: " + user.id);
    }
}
```

**Request 2 - Pre-request**: Use stored ID

```javascript
// Pre-request: Use stored user ID in URL
var userId = lc.globals.get("user_id") || lc.env.get("user_id");

if (userId) {
    lc.request.url = lc.request.url.replace("{userId}", userId);
    console.log("Using user ID: " + userId);
} else {
    console.error("No user ID available - run Get User first");
}
```

### Pass Data via Environment

```javascript
// Request 1 - Post-response: Store multiple values
if (lc.response.status === 200) {
    var data = lc.response.body.json();

    lc.env.set("resource_id", data.id);
    lc.env.set("resource_name", data.name);
    lc.env.set("resource_type", data.type);

    console.log("Stored resource: " + data.name);
}
```

```javascript
// Request 2 - Pre-request: Use stored values
var body = {
    resourceId: lc.env.get("resource_id"),
    resourceName: lc.env.get("resource_name"),
    action: "update"
};

lc.request.body.set(JSON.stringify(body));
```

---

## Authentication Flows

### Login → Use Token

**Login Request - Post-response:**

```javascript
// Store auth tokens from login response
if (lc.response.status === 200) {
    var data = lc.response.body.json();

    if (data.access_token) {
        // Store in both globals (fast) and env (persistent)
        lc.globals.set("access_token", data.access_token);
        lc.env.set("access_token", data.access_token);

        // Store expiry
        if (data.expires_in) {
            var expiresAt = Date.now() + (data.expires_in * 1000);
            lc.globals.set("token_expires", expiresAt);
        }

        // Store user info
        if (data.user) {
            lc.globals.set("current_user", data.user);
            lc.env.set("user_id", data.user.id.toString());
        }

        console.log("Login successful - token stored");
    }
} else {
    console.error("Login failed: " + lc.response.status);
}
```

**Protected Request - Pre-request:**

```javascript
// Add auth token to request
var token = lc.globals.get("access_token") || lc.env.get("access_token");

if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
} else {
    console.error("No token available - please login first");
}
```

### Auto Token Refresh

```javascript
// Pre-request: Check and refresh token if needed
var token = lc.env.get("access_token");
var expiry = lc.globals.get("token_expires") || 0;
var now = Date.now();

// Refresh if expired or expiring soon (within 60 seconds)
if (!token || now >= expiry - 60000) {
    var refreshToken = lc.env.get("refresh_token");

    if (refreshToken) {
        console.log("Refreshing token...");

        lc.sendRequest({
            url: lc.env.get("auth_url") + "/token/refresh",
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ refresh_token: refreshToken })
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
    } else {
        console.error("No refresh token - please login again");
    }
}

// Use the (possibly refreshed) token
token = lc.env.get("access_token");
if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
}
```

---

## CRUD Operations

### Create → Read → Update → Delete

**1. Create Resource - Post-response:**

```javascript
// Store created resource ID
if (lc.response.status === 201) {
    var resource = lc.response.body.json();

    lc.globals.set("test_resource_id", resource.id);
    lc.env.set("test_resource_id", resource.id);

    console.log("Created resource: " + resource.id);
}

lc.test("Resource created", function() {
    lc.expect(lc.response.status).toBe(201);
});
```

**2. Read Resource - Pre-request:**

```javascript
// Use stored ID in URL
var resourceId = lc.globals.get("test_resource_id");
if (resourceId) {
    lc.request.url = lc.request.url.replace("{id}", resourceId);
}
```

**3. Update Resource - Pre-request:**

```javascript
// Use stored ID and add update data
var resourceId = lc.globals.get("test_resource_id");
if (resourceId) {
    lc.request.url = lc.request.url.replace("{id}", resourceId);

    var updateData = {
        name: "Updated Name " + Date.now(),
        updatedAt: lc.variables.isoTimestamp()
    };
    lc.request.body.set(JSON.stringify(updateData));
}
```

**4. Delete Resource - Pre-request:**

```javascript
// Use stored ID for deletion
var resourceId = lc.globals.get("test_resource_id");
if (resourceId) {
    lc.request.url = lc.request.url.replace("{id}", resourceId);
    console.log("Deleting resource: " + resourceId);
}
```

**4. Delete Resource - Post-response:**

```javascript
// Clean up stored references
if (lc.response.status === 204 || lc.response.status === 200) {
    lc.globals.unset("test_resource_id");
    lc.env.unset("test_resource_id");
    console.log("Resource deleted and references cleaned");
}
```

---

## Data Extraction Patterns

### Extract from Array Response

```javascript
// Post-response: Extract specific item from list
if (lc.response.status === 200) {
    var data = lc.response.body.json();

    if (data.items && data.items.length > 0) {
        // Store first item
        var first = data.items[0];
        lc.globals.set("first_item_id", first.id);
        lc.globals.set("first_item", first);

        // Store last item
        var last = data.items[data.items.length - 1];
        lc.globals.set("last_item_id", last.id);

        // Store all IDs for batch operations
        var allIds = data.items.map(function(item) {
            return item.id;
        });
        lc.globals.set("all_item_ids", allIds);

        console.log("Stored " + allIds.length + " item IDs");
    }
}
```

### Extract Nested Data

```javascript
// Post-response: Extract from nested structure
if (lc.response.status === 200) {
    var data = lc.response.body.json();

    // Extract user info
    if (data.user) {
        lc.globals.set("user_id", data.user.id);
        lc.globals.set("user_email", data.user.email);
    }

    // Extract organization info
    if (data.user && data.user.organization) {
        lc.globals.set("org_id", data.user.organization.id);
        lc.globals.set("org_name", data.user.organization.name);
    }

    // Extract permissions
    if (data.permissions) {
        lc.globals.set("user_permissions", data.permissions);
    }
}
```

### Extract Headers

```javascript
// Post-response: Extract important headers
var rateLimitRemaining = lc.response.headers.get("X-RateLimit-Remaining");
var rateLimitReset = lc.response.headers.get("X-RateLimit-Reset");
var requestId = lc.response.headers.get("X-Request-ID");

if (rateLimitRemaining) {
    lc.globals.set("rate_limit_remaining", parseInt(rateLimitRemaining));

    if (parseInt(rateLimitRemaining) < 10) {
        console.warn("Rate limit low: " + rateLimitRemaining + " remaining");
    }
}

if (requestId) {
    lc.globals.set("last_request_id", requestId);
}
```

---

## Error Handling

### Handle Failed Chains

```javascript
// Pre-request: Check prerequisites
var token = lc.env.get("access_token");
var resourceId = lc.globals.get("resource_id");

if (!token) {
    console.error("CHAIN BROKEN: No access token");
    console.error("Action: Run Login request first");
}

if (!resourceId) {
    console.error("CHAIN BROKEN: No resource ID");
    console.error("Action: Run Create Resource request first");
}

if (token && resourceId) {
    lc.request.headers.set("Authorization", "Bearer " + token);
    lc.request.url = lc.request.url.replace("{id}", resourceId);
    console.log("Chain prerequisites met");
}
```

### Retry on Failure

```javascript
// Post-response: Handle and store error for retry
if (lc.response.status >= 400) {
    var retryCount = lc.globals.get("retry_count") || 0;
    retryCount++;
    lc.globals.set("retry_count", retryCount);

    console.error("Request failed (attempt " + retryCount + ")");

    if (lc.response.status === 401) {
        console.error("Auth failed - token may be expired");
        lc.env.unset("access_token");
    } else if (lc.response.status === 429) {
        var retryAfter = lc.response.headers.get("Retry-After");
        console.error("Rate limited. Retry after: " + (retryAfter || "60") + "s");
    }
} else {
    // Reset retry count on success
    lc.globals.set("retry_count", 0);
}
```

---

## Advanced Patterns

### Dynamic Request with sendRequest

```javascript
// Pre-request: Fetch config before main request
lc.sendRequest({
    url: lc.env.get("base_url") + "/config",
    method: "GET",
    headers: {
        "Authorization": "Bearer " + lc.env.get("access_token")
    }
}, function(err, response) {
    if (err || response.status !== 200) {
        console.error("Failed to fetch config");
        return;
    }

    var config = response.body.json();

    // Use config in current request
    if (config.apiVersion) {
        lc.request.headers.set("X-API-Version", config.apiVersion);
    }

    if (config.features && config.features.newEndpoint) {
        lc.request.url = lc.request.url.replace("/v1/", "/v2/");
    }

    lc.globals.set("api_config", config);
    console.log("Config loaded: API v" + config.apiVersion);
});
```

### Create Parent, Then Child

```javascript
// Pre-request: Create folder before uploading file
var folderId = lc.globals.get("upload_folder_id");

if (!folderId) {
    console.log("Creating upload folder...");

    lc.sendRequest({
        url: lc.env.get("base_url") + "/folders",
        method: "POST",
        headers: {
            "Authorization": "Bearer " + lc.env.get("access_token"),
            "Content-Type": "application/json"
        },
        body: JSON.stringify({
            name: "Upload Folder " + lc.variables.timestamp()
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

// Use folder ID in request
folderId = lc.globals.get("upload_folder_id");
if (folderId) {
    var body = lc.request.body.json() || {};
    body.folderId = folderId;
    lc.request.body.set(JSON.stringify(body));
}
```

### Batch Processing

```javascript
// Pre-request: Process items in batch
var itemIds = lc.globals.get("pending_items") || [];
var batchSize = 5;

if (itemIds.length > 0) {
    var batch = itemIds.slice(0, batchSize);
    var remaining = itemIds.slice(batchSize);

    lc.globals.set("current_batch", batch);
    lc.globals.set("pending_items", remaining);

    // Build batch request body
    var batchRequest = {
        items: batch,
        operation: "process"
    };
    lc.request.body.set(JSON.stringify(batchRequest));

    console.log("Processing batch of " + batch.length + " items");
    console.log("Remaining: " + remaining.length + " items");
}
```

### Sequential Workflow State

```javascript
// Track workflow state
var workflowState = lc.globals.get("workflow_state") || "init";

console.log("Current workflow state: " + workflowState);

switch (workflowState) {
    case "init":
        // First request - create resource
        lc.globals.set("workflow_state", "created");
        break;

    case "created":
        // Second request - verify resource
        var resourceId = lc.globals.get("resource_id");
        lc.request.url = lc.request.url.replace("{id}", resourceId);
        lc.globals.set("workflow_state", "verified");
        break;

    case "verified":
        // Third request - activate resource
        var resourceId = lc.globals.get("resource_id");
        lc.request.url = lc.request.url.replace("{id}", resourceId);
        var body = { status: "active" };
        lc.request.body.set(JSON.stringify(body));
        lc.globals.set("workflow_state", "complete");
        break;

    case "complete":
        console.log("Workflow complete!");
        break;
}
```

---

## See Also

- [lc.env](../api/env.md) - Environment variables
- [lc.globals](../api/env.md#lcglobals) - Session variables
- [lc.sendRequest](../api/sendrequest.md) - Request chaining API
- [Authentication Examples](./authentication.md)
- [Testing Examples](./testing.md)

---

*[← Testing Examples](./testing.md) | [API Overview →](../api/overview.md)*
