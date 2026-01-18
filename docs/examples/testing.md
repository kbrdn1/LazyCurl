# Testing Examples

Practical examples for writing tests and assertions in LazyCurl scripts.

## Table of Contents

- [Overview](#overview)
- [Basic Response Validation](#basic-response-validation)
- [JSON Response Testing](#json-response-testing)
- [Error Response Testing](#error-response-testing)
- [Performance Testing](#performance-testing)
- [Authentication Testing](#authentication-testing)
- [Data Validation Patterns](#data-validation-patterns)
- [Common Patterns](#common-patterns)
- [See Also](#see-also)

---

## Overview

LazyCurl provides Jest-like testing APIs:

| Function | Purpose |
|----------|---------|
| `lc.test(name, fn)` | Define a named test case |
| `lc.expect(value)` | Create assertions |
| `lc.expect(value).not.*` | Negate assertions |

Test results appear in the **Tests** tab of the Response panel.

---

## Basic Response Validation

### Status Code Tests

```javascript
// Test specific status code
lc.test("Returns 200 OK", function() {
    lc.expect(lc.response.status).toBe(200);
});

// Test status code range
lc.test("Returns success status", function() {
    lc.expect(lc.response.status).toBeGreaterThanOrEqual(200);
    lc.expect(lc.response.status).toBeLessThan(300);
});

// Test for specific error
lc.test("Returns 404 for missing resource", function() {
    lc.expect(lc.response.status).toBe(404);
});
```

### Response Time Tests

```javascript
// Basic performance check
lc.test("Response time under 2 seconds", function() {
    lc.expect(lc.response.time).toBeLessThan(2000);
});

// Strict performance requirement
lc.test("Response time under 500ms", function() {
    lc.expect(lc.response.time).toBeLessThan(500);
});

// Performance tiers
var time = lc.response.time;
if (time < 200) {
    console.log("Performance: Excellent");
} else if (time < 500) {
    console.log("Performance: Good");
} else {
    console.warn("Performance: Needs improvement");
}
```

### Content-Type Tests

```javascript
lc.test("Returns JSON content", function() {
    var contentType = lc.response.headers.get("Content-Type");
    lc.expect(contentType).toContain("application/json");
});

lc.test("Returns correct content type", function() {
    var contentType = lc.response.headers.get("Content-Type");
    lc.expect(contentType).toBeDefined();
    lc.expect(contentType).not.toContain("text/html");
});
```

---

## JSON Response Testing

### Basic Structure

```javascript
var data = lc.response.body.json();

lc.test("Response is valid JSON", function() {
    lc.expect(data).not.toBeNull();
});

lc.test("Response has required fields", function() {
    lc.expect(data).toHaveProperty("id");
    lc.expect(data).toHaveProperty("name");
    lc.expect(data).toHaveProperty("email");
});
```

### Nested Objects

```javascript
var data = lc.response.body.json();

lc.test("User has address", function() {
    lc.expect(data).toHaveProperty("address");
    lc.expect(data.address).toHaveProperty("city");
    lc.expect(data.address).toHaveProperty("country");
});

lc.test("User has contact info", function() {
    lc.expect(data.contact).toBeDefined();
    lc.expect(data.contact.email).toMatch("@");
});
```

### Array Responses

```javascript
var data = lc.response.body.json();

lc.test("Response is an array", function() {
    lc.expect(Array.isArray(data.items)).toBe(true);
});

lc.test("Array is not empty", function() {
    lc.expect(data.items.length).toBeGreaterThan(0);
});

lc.test("Array items have required fields", function() {
    var firstItem = data.items[0];
    lc.expect(firstItem).toHaveProperty("id");
    lc.expect(firstItem).toHaveProperty("name");
});

lc.test("All items have IDs", function() {
    var allHaveIds = data.items.every(function(item) {
        return item.id !== undefined;
    });
    lc.expect(allHaveIds).toBe(true);
});
```

### Pagination Response

```javascript
var data = lc.response.body.json();

lc.test("Response includes pagination info", function() {
    lc.expect(data).toHaveProperty("items");
    lc.expect(data).toHaveProperty("total");
    lc.expect(data).toHaveProperty("page");
    lc.expect(data).toHaveProperty("pageSize");
});

lc.test("Pagination values are valid", function() {
    lc.expect(data.page).toBeGreaterThanOrEqual(1);
    lc.expect(data.pageSize).toBeGreaterThan(0);
    lc.expect(data.total).toBeGreaterThanOrEqual(0);
});

lc.test("Items count matches page size", function() {
    var itemCount = data.items.length;
    lc.expect(itemCount).toBeLessThanOrEqual(data.pageSize);
});

console.log("Page " + data.page + ": " + data.items.length + "/" + data.total + " items");
```

---

## Error Response Testing

### Validation Error

```javascript
// Test 400 Bad Request response
lc.test("Returns 400 for invalid input", function() {
    lc.expect(lc.response.status).toBe(400);
});

var error = lc.response.body.json();

lc.test("Error response has message", function() {
    lc.expect(error).toHaveProperty("message");
    lc.expect(error.message).toBeDefined();
});

lc.test("Error includes validation details", function() {
    lc.expect(error).toHaveProperty("errors");
    lc.expect(Array.isArray(error.errors)).toBe(true);
});

// Log validation errors
if (error && error.errors) {
    error.errors.forEach(function(err) {
        console.log("Validation error: " + err.field + " - " + err.message);
    });
}
```

### Not Found Error

```javascript
lc.test("Returns 404 Not Found", function() {
    lc.expect(lc.response.status).toBe(404);
});

lc.test("Error message mentions resource", function() {
    var error = lc.response.body.json();
    lc.expect(error.message).toContain("not found");
});
```

### Authentication Error

```javascript
lc.test("Returns 401 Unauthorized", function() {
    lc.expect(lc.response.status).toBe(401);
});

lc.test("Error indicates auth required", function() {
    var error = lc.response.body.json();
    lc.expect(error.code).toBe("UNAUTHORIZED");
});

// Clear stored tokens on 401
if (lc.response.status === 401) {
    lc.env.unset("access_token");
    console.warn("Authentication failed - token cleared");
}
```

---

## Performance Testing

### Response Time Thresholds

```javascript
lc.test("Fast response (< 200ms)", function() {
    lc.expect(lc.response.time).toBeLessThan(200);
});

lc.test("Acceptable response (< 1s)", function() {
    lc.expect(lc.response.time).toBeLessThan(1000);
});

lc.test("Within SLA (< 2s)", function() {
    lc.expect(lc.response.time).toBeLessThan(2000);
});
```

### Performance Categories

```javascript
var time = lc.response.time;

lc.test("Response time is acceptable", function() {
    lc.expect(time).toBeLessThan(2000);
});

// Categorize performance
if (time < 100) {
    console.log("Excellent: " + time + "ms");
} else if (time < 300) {
    console.log("Good: " + time + "ms");
} else if (time < 1000) {
    console.log("Acceptable: " + time + "ms");
} else {
    console.warn("Slow: " + time + "ms - needs optimization");
}
```

### Track Performance Over Time

```javascript
// Store performance data
var perfHistory = lc.globals.get("perf_history") || [];

perfHistory.push({
    endpoint: lc.request.url,
    time: lc.response.time,
    status: lc.response.status,
    timestamp: Date.now()
});

// Keep last 100 entries
if (perfHistory.length > 100) {
    perfHistory.shift();
}

lc.globals.set("perf_history", perfHistory);

// Calculate average
var total = perfHistory.reduce(function(sum, entry) {
    return sum + entry.time;
}, 0);
var avg = Math.round(total / perfHistory.length);

console.log("Current: " + lc.response.time + "ms | Average: " + avg + "ms");
```

---

## Authentication Testing

### Login Response

```javascript
lc.test("Login successful", function() {
    lc.expect(lc.response.status).toBe(200);
});

var data = lc.response.body.json();

lc.test("Response contains access token", function() {
    lc.expect(data).toHaveProperty("access_token");
    lc.expect(data.access_token).toBeDefined();
});

lc.test("Response contains token expiry", function() {
    lc.expect(data).toHaveProperty("expires_in");
    lc.expect(data.expires_in).toBeGreaterThan(0);
});

lc.test("Token is valid JWT format", function() {
    var parts = data.access_token.split(".");
    lc.expect(parts.length).toBe(3);
});

// Store tokens for subsequent requests
if (data && data.access_token) {
    lc.env.set("access_token", data.access_token);
    console.log("Token stored successfully");
}
```

### Protected Endpoint Access

```javascript
lc.test("Protected endpoint returns data", function() {
    lc.expect(lc.response.status).toBe(200);
});

lc.test("Response is not unauthorized", function() {
    lc.expect(lc.response.status).not.toBe(401);
    lc.expect(lc.response.status).not.toBe(403);
});
```

---

## Data Validation Patterns

### Email Format

```javascript
lc.test("User email is valid format", function() {
    var user = lc.response.body.json();
    lc.expect(user.email).toMatch("@");
    lc.expect(user.email).toContain(".");
});
```

### UUID Format

```javascript
lc.test("ID is valid UUID", function() {
    var data = lc.response.body.json();
    var uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
    lc.expect(data.id).toMatch(uuidRegex.source);
});
```

### Date Format

```javascript
lc.test("Created date is valid ISO format", function() {
    var data = lc.response.body.json();
    var date = new Date(data.createdAt);
    lc.expect(isNaN(date.getTime())).toBe(false);
});

lc.test("Date is recent", function() {
    var data = lc.response.body.json();
    var created = new Date(data.createdAt);
    var now = new Date();
    var hourAgo = new Date(now.getTime() - 60 * 60 * 1000);

    lc.expect(created.getTime()).toBeGreaterThan(hourAgo.getTime());
});
```

### Numeric Ranges

```javascript
var product = lc.response.body.json();

lc.test("Price is positive", function() {
    lc.expect(product.price).toBeGreaterThan(0);
});

lc.test("Quantity is non-negative", function() {
    lc.expect(product.quantity).toBeGreaterThanOrEqual(0);
});

lc.test("Rating is within range", function() {
    lc.expect(product.rating).toBeGreaterThanOrEqual(0);
    lc.expect(product.rating).toBeLessThanOrEqual(5);
});
```

---

## Common Patterns

### Conditional Tests

```javascript
var data = lc.response.body.json();

// Only run test if condition is met
if (data && data.items && data.items.length > 0) {
    lc.test("First item has required fields", function() {
        var first = data.items[0];
        lc.expect(first).toHaveProperty("id");
        lc.expect(first).toHaveProperty("name");
    });
}
```

### Test Suite Pattern

```javascript
// Organize tests into logical groups
function testStatusCode() {
    lc.test("Status is 200", function() {
        lc.expect(lc.response.status).toBe(200);
    });
}

function testResponseTime() {
    lc.test("Response under 2s", function() {
        lc.expect(lc.response.time).toBeLessThan(2000);
    });
}

function testResponseBody() {
    var data = lc.response.body.json();

    lc.test("Body is valid JSON", function() {
        lc.expect(data).not.toBeNull();
    });

    lc.test("Body has required fields", function() {
        lc.expect(data).toHaveProperty("id");
    });
}

// Run all test groups
testStatusCode();
testResponseTime();
testResponseBody();
```

### Reusable Assertions

```javascript
// Helper function for common validations
function assertUserFields(user, testPrefix) {
    testPrefix = testPrefix || "";

    lc.test(testPrefix + "User has ID", function() {
        lc.expect(user).toHaveProperty("id");
    });

    lc.test(testPrefix + "User has valid email", function() {
        lc.expect(user.email).toMatch("@");
    });

    lc.test(testPrefix + "User has name", function() {
        lc.expect(user.name).toBeDefined();
        lc.expect(user.name.length).toBeGreaterThan(0);
    });
}

// Use the helper
var data = lc.response.body.json();
if (data && data.user) {
    assertUserFields(data.user, "Response: ");
}
```

### Log and Test Together

```javascript
// Combined logging and testing
var data = lc.response.body.json();

console.log("=== Test Results ===");
console.log("Status: " + lc.response.status);
console.log("Time: " + lc.response.time + "ms");

if (data) {
    console.log("Items: " + (data.items ? data.items.length : 0));
    console.log("Total: " + (data.total || "N/A"));
}

lc.test("Status is 200", function() {
    lc.expect(lc.response.status).toBe(200);
});

lc.test("Has items", function() {
    lc.expect(data.items.length).toBeGreaterThan(0);
});

console.log("====================");
```

---

## See Also

- [lc.test](../api/test.md) - Testing API reference
- [lc.response](../api/response.md) - Response object
- [Authentication Examples](./authentication.md)
- [Request Chaining Examples](./request-chaining.md)

---

*[← Authentication Examples](./authentication.md) | [Request Chaining Examples →](./request-chaining.md)*
