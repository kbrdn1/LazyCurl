# console

Standard JavaScript-style logging functions. Output is displayed in the Console tab.

**Availability**: Both (pre-request and post-response)

## Quick Reference

| Method | Description |
|--------|-------------|
| `log(args...)` | General logging |
| `info(args...)` | Informational messages |
| `warn(args...)` | Warning messages |
| `error(args...)` | Error messages |
| `debug(args...)` | Debug messages |

## Methods

### log(args...)

General-purpose logging.

```javascript
console.log("Request sent successfully");
console.log("User:", userName);
console.log("Count:", 42, "items");
```

### info(args...)

Informational messages.

```javascript
console.info("Processing started");
console.info("Found", items.length, "items");
```

### warn(args...)

Warning messages for non-critical issues.

```javascript
console.warn("Rate limit approaching");
console.warn("Deprecated API version detected");
console.warn("Missing optional field:", fieldName);
```

### error(args...)

Error messages for failures.

```javascript
console.error("Request failed");
console.error("Invalid response format:", response.status);
console.error("Missing required field: api_key");
```

### debug(args...)

Debug messages for detailed information.

```javascript
console.debug("Request headers:", headers);
console.debug("Parsed body:", JSON.stringify(body, null, 2));
```

## Output Formatting

### Multiple Arguments

All methods accept multiple arguments:

```javascript
console.log("User:", userName, "- ID:", userId);
// Output: "User: John - ID: 123"
```

### Objects

Objects are JSON-formatted:

```javascript
console.log({ name: "John", age: 30 });
// Output: {"name":"John","age":30}

// Pretty-printing
console.log(JSON.stringify(data, null, 2));
```

## Examples

### Basic Logging

```javascript
// Pre-request: Log request info
console.log("Sending request to:", lc.request.url);
console.log("Method:", lc.request.method);
console.log("Headers:", lc.request.headers.all());
```

### Response Logging

```javascript
// Post-response: Log response details
console.log("=== Response ===");
console.log("Status:", lc.response.statusText);
console.log("Time:", lc.response.time, "ms");
console.log("Content-Type:", lc.response.headers.get("Content-Type"));
```

### Conditional Logging

```javascript
// Post-response: Log based on status
if (lc.response.status === 200) {
    console.info("Request successful");
} else if (lc.response.status >= 400 && lc.response.status < 500) {
    console.warn("Client error:", lc.response.status);
} else if (lc.response.status >= 500) {
    console.error("Server error:", lc.response.status);
}
```

### Error Handling with Logging

```javascript
// Post-response: Detailed error logging
if (lc.response.status >= 400) {
    console.error("=== Error Details ===");
    console.error("Status:", lc.response.statusText);
    console.error("Endpoint:", lc.request.url);

    var errorBody = lc.response.body.json();
    if (errorBody) {
        console.error("Error Code:", errorBody.code);
        console.error("Error Message:", errorBody.message);
        if (errorBody.details) {
            console.error("Details:", JSON.stringify(errorBody.details, null, 2));
        }
    }
    console.error("=====================");
}
```

### Progress Tracking

```javascript
// Pre-request: Log progress in collection runs
console.info("=== Request", lc.info.iteration, "of collection ===");
console.info("Collection:", lc.info.collectionName);
console.info("Request:", lc.info.requestName);
console.info("Environment:", lc.info.environmentName);
```

### Performance Logging

```javascript
// Post-response: Performance analysis
var time = lc.response.time;

if (time < 200) {
    console.log("Performance: Excellent (" + time + "ms)");
} else if (time < 500) {
    console.info("Performance: Good (" + time + "ms)");
} else if (time < 1000) {
    console.warn("Performance: Acceptable (" + time + "ms)");
} else {
    console.error("Performance: SLOW (" + time + "ms)");
}
```

### Variable Inspection

```javascript
// Pre-request: Inspect environment variables
console.log("=== Environment Variables ===");
var vars = lc.env.toObject();
for (var key in vars) {
    // Mask sensitive values
    var value = vars[key];
    if (key.toLowerCase().includes("secret") || key.toLowerCase().includes("password")) {
        value = "****";
    }
    console.log(key + ":", value);
}
```

## See Also

- [lc.test](API-Test) - Testing and assertions
- [lc.info](API-Info) - Execution context
- [Testing Examples](Examples-Testing)

---

*[lc.sendRequest](API-SendRequest) | [API Overview](API-Overview)*
