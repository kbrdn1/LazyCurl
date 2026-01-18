# lc.info

Read-only contextual information about the current script execution environment.

**Availability**: Both (pre-request and post-response)

> For the complete API reference, see [Scripting API Reference](../scripting-api-reference.md#lcinfo).

## Table of Contents

- [Overview](#overview)
- [Quick Reference](#quick-reference)
- [Properties](#properties)
- [Examples](#examples)
- [See Also](#see-also)

---

## Overview

The `lc.info` object provides read-only metadata about the current execution context:

- Script type (pre-request or post-response)
- Request and collection names
- Active environment
- Current iteration (for collection runs)

Useful for conditional logic, logging, and environment-specific behavior.

---

## Quick Reference

| Property | Type | Description |
|----------|------|-------------|
| `scriptType` | string | `"pre-request"` or `"post-response"` |
| `requestName` | string \| undefined | Name of the current request |
| `requestId` | string \| undefined | Unique ID of the request |
| `collectionName` | string \| undefined | Name of the collection |
| `environmentName` | string \| undefined | Active environment name |
| `iteration` | number | Current iteration number (default: 1) |

---

## Properties

### scriptType

The type of script currently executing.

**Type**: `string`
**Values**: `"pre-request"` or `"post-response"`

```javascript
if (lc.info.scriptType === "pre-request") {
    console.log("Running before request");
    // Can modify request
} else {
    console.log("Running after response");
    // Can access response
}
```

---

### requestName

The name of the current request as defined in the collection.

**Type**: `string` | `undefined`

```javascript
console.log("Executing: " + lc.info.requestName);
// Output: "Executing: Get User Profile"
```

---

### requestId

The unique identifier of the current request.

**Type**: `string` | `undefined`

```javascript
console.log("Request ID: " + lc.info.requestId);
// Output: "Request ID: req_abc123"
```

---

### collectionName

The name of the collection containing the request.

**Type**: `string` | `undefined`

```javascript
console.log("Collection: " + lc.info.collectionName);
// Output: "Collection: My API"
```

---

### environmentName

The name of the currently active environment.

**Type**: `string` | `undefined`

```javascript
console.log("Environment: " + lc.info.environmentName);
// Output: "Environment: production"
```

---

### iteration

The current iteration number when running collection iterations. Defaults to `1` for single request execution.

**Type**: `number`

```javascript
console.log("Iteration: " + lc.info.iteration);
// Output: "Iteration: 1"
```

---

## Examples

### Conditional Logic by Script Type

```javascript
// Common script that works in both contexts
if (lc.info.scriptType === "pre-request") {
    // Add authentication
    var token = lc.env.get("auth_token");
    if (token) {
        lc.request.headers.set("Authorization", "Bearer " + token);
    }
} else {
    // Validate response
    lc.test("Response successful", function() {
        lc.expect(lc.response.status).toBe(200);
    });
}
```

### Environment-Specific Behavior

```javascript
// Pre-request: Configure based on environment
var env = lc.info.environmentName;

if (env === "production") {
    console.warn("Running against PRODUCTION!");
    // Extra validation for production
    if (!lc.env.has("prod_api_key")) {
        console.error("Missing production API key");
    }
} else if (env === "staging") {
    console.log("Running against staging");
    lc.request.headers.set("X-Debug", "true");
} else {
    console.log("Running against development");
    lc.request.headers.set("X-Debug", "verbose");
}
```

### Iteration-Aware Testing

```javascript
// Post-response: Different expectations per iteration
if (lc.info.iteration === 1) {
    // First iteration creates the resource
    lc.test("Resource created", function() {
        lc.expect(lc.response.status).toBe(201);
    });
} else {
    // Subsequent iterations update or retrieve
    lc.test("Resource exists", function() {
        lc.expect(lc.response.status).toBe(200);
    });
}

console.log("Completed iteration " + lc.info.iteration);
```

### Contextual Logging

```javascript
// Post-response: Detailed logging with context
var context = "[" + lc.info.collectionName + "/" + lc.info.requestName + "]";
var env = lc.info.environmentName || "default";

console.log(context + " Status: " + lc.response.status);
console.log(context + " Time: " + lc.response.time + "ms");
console.log(context + " Environment: " + env);

if (lc.info.iteration > 1) {
    console.log(context + " Iteration: " + lc.info.iteration);
}
```

### Dynamic Test Names

```javascript
// Post-response: Include context in test names
var prefix = lc.info.requestName + " - ";

lc.test(prefix + "returns 200", function() {
    lc.expect(lc.response.status).toBe(200);
});

lc.test(prefix + "responds quickly", function() {
    lc.expect(lc.response.time).toBeLessThan(1000);
});
```

### Skip Logic for Specific Requests

```javascript
// Pre-request: Skip authentication for certain requests
var skipAuth = ["Health Check", "Public Status"];

if (skipAuth.indexOf(lc.info.requestName) === -1) {
    // Add authentication for all other requests
    var token = lc.env.get("auth_token");
    if (token) {
        lc.request.headers.set("Authorization", "Bearer " + token);
    }
} else {
    console.log("Skipping auth for: " + lc.info.requestName);
}
```

### Request Tracking

```javascript
// Pre-request: Log request execution for debugging
console.log("=== Request Execution ===");
console.log("Collection: " + (lc.info.collectionName || "N/A"));
console.log("Request: " + (lc.info.requestName || "N/A"));
console.log("Request ID: " + (lc.info.requestId || "N/A"));
console.log("Environment: " + (lc.info.environmentName || "default"));
console.log("Iteration: " + lc.info.iteration);
console.log("Script Type: " + lc.info.scriptType);
console.log("=========================");
```

### Store Execution Metadata

```javascript
// Post-response: Store execution info for reporting
var execution = {
    collection: lc.info.collectionName,
    request: lc.info.requestName,
    environment: lc.info.environmentName,
    iteration: lc.info.iteration,
    status: lc.response.status,
    time: lc.response.time,
    timestamp: new Date().toISOString()
};

// Store in globals for later analysis
var history = lc.globals.get("execution_history") || [];
history.push(execution);
lc.globals.set("execution_history", history);

console.log("Execution #" + history.length + " recorded");
```

---

## See Also

- [lc.env](./env.md) - Environment variables
- [lc.test](./test.md) - Testing and assertions
- [Request Chaining Examples](../examples/request-chaining.md)

---

*[← lc.variables](./variables.md) | [lc.sendRequest →](./sendrequest.md)*
