# lc.info

Read-only access to the current execution context.

**Availability**: Both (pre-request and post-response)

## Quick Reference

| Property | Type | Description |
|----------|------|-------------|
| `requestName` | string | Name of the current request |
| `collectionName` | string | Name of the current collection |
| `environmentName` | string | Name of the active environment |
| `iteration` | number | Current iteration number (collection runs) |
| `iterationCount` | number | Total iterations planned |
| `requestId` | string | Unique identifier for this request |

## Properties

### requestName

The name of the current request as defined in the collection.

```javascript
console.log("Running request: " + lc.info.requestName);
// "Running request: Get User"
```

### collectionName

The name of the current collection.

```javascript
console.log("Collection: " + lc.info.collectionName);
// "Collection: User API"
```

### environmentName

The name of the active environment.

```javascript
console.log("Environment: " + lc.info.environmentName);
// "Environment: Production"

// Conditional logic based on environment
if (lc.info.environmentName === "production") {
    console.warn("Running against PRODUCTION!");
}
```

### iteration

Current iteration number when running collections (1-indexed).

```javascript
console.log("Iteration " + lc.info.iteration + " of " + lc.info.iterationCount);
// "Iteration 3 of 10"
```

### iterationCount

Total number of iterations planned for collection run.

```javascript
var progress = (lc.info.iteration / lc.info.iterationCount * 100).toFixed(0);
console.log("Progress: " + progress + "%");
```

### requestId

Unique identifier for this request execution.

```javascript
console.log("Request ID: " + lc.info.requestId);
lc.request.headers.set("X-Request-ID", lc.info.requestId);
```

## Examples

### Structured Logging

```javascript
// Pre-request: Log execution context
console.log("=== Request Context ===");
console.log("Collection: " + lc.info.collectionName);
console.log("Request: " + lc.info.requestName);
console.log("Environment: " + lc.info.environmentName);
console.log("Iteration: " + lc.info.iteration + "/" + lc.info.iterationCount);
console.log("=======================");
```

### Environment-Aware Behavior

```javascript
// Pre-request: Different behavior per environment
var env = lc.info.environmentName || "development";

switch (env) {
    case "production":
        console.warn("PRODUCTION environment");
        // Skip destructive operations in production
        if (lc.request.method === "DELETE") {
            console.error("DELETE operations disabled in production");
        }
        break;

    case "staging":
        lc.request.headers.set("X-Debug", "true");
        break;

    default:
        lc.request.headers.set("X-Debug", "verbose");
}
```

### Progress Tracking

```javascript
// Pre-request: Show collection run progress
if (lc.info.iterationCount > 1) {
    var current = lc.info.iteration;
    var total = lc.info.iterationCount;
    var percent = Math.round((current / total) * 100);

    console.log("[" + current + "/" + total + "] " + percent + "% - " + lc.info.requestName);
}
```

### Request Correlation

```javascript
// Pre-request: Add correlation headers
lc.request.headers.set("X-Request-ID", lc.info.requestId);
lc.request.headers.set("X-Request-Name", lc.info.requestName);
lc.request.headers.set("X-Collection", lc.info.collectionName);
```

### Conditional Test Execution

```javascript
// Post-response: Skip some tests in certain environments
lc.test("Status is 200", function() {
    lc.expect(lc.response.status).toBe(200);
});

// Only run performance tests in staging/production
if (lc.info.environmentName !== "development") {
    lc.test("Response time under 500ms", function() {
        lc.expect(lc.response.time).toBeLessThan(500);
    });
}
```

## See Also

- [lc.env](API-Env) - Environment variables
- [lc.test](API-Test) - Testing and assertions
- [Testing Examples](Examples-Testing)

---

*[lc.variables](API-Variables) | [lc.sendRequest](API-SendRequest)*
