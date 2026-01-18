# Testing Examples

Practical examples for writing tests and assertions in LazyCurl scripts.

## Overview

| Function | Purpose |
|----------|---------|
| `lc.test(name, fn)` | Define a named test case |
| `lc.expect(value)` | Create assertions |
| `lc.expect(value).not.*` | Negate assertions |

Test results appear in the **Tests** tab of the Response panel.

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
```

### Response Time Tests

```javascript
lc.test("Response time under 2 seconds", function() {
    lc.expect(lc.response.time).toBeLessThan(2000);
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
```

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
});

console.log("Page " + data.page + ": " + data.items.length + "/" + data.total + " items");
```

## Error Response Testing

### Validation Error

```javascript
lc.test("Returns 400 for invalid input", function() {
    lc.expect(lc.response.status).toBe(400);
});

var error = lc.response.body.json();

lc.test("Error response has message", function() {
    lc.expect(error).toHaveProperty("message");
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

### Authentication Error

```javascript
lc.test("Returns 401 Unauthorized", function() {
    lc.expect(lc.response.status).toBe(401);
});

// Clear stored tokens on 401
if (lc.response.status === 401) {
    lc.env.unset("access_token");
    console.warn("Authentication failed - token cleared");
}
```

## Performance Testing

```javascript
lc.test("Fast response (< 200ms)", function() {
    lc.expect(lc.response.time).toBeLessThan(200);
});

lc.test("Acceptable response (< 1s)", function() {
    lc.expect(lc.response.time).toBeLessThan(1000);
});

// Track performance
var perfHistory = lc.globals.get("perf_history") || [];
perfHistory.push({
    endpoint: lc.request.url,
    time: lc.response.time,
    timestamp: Date.now()
});

if (perfHistory.length > 100) perfHistory.shift();
lc.globals.set("perf_history", perfHistory);

var avg = perfHistory.reduce(function(sum, e) { return sum + e.time; }, 0) / perfHistory.length;
console.log("Current: " + lc.response.time + "ms | Average: " + Math.round(avg) + "ms");
```

## Authentication Testing

```javascript
lc.test("Login successful", function() {
    lc.expect(lc.response.status).toBe(200);
});

var data = lc.response.body.json();

lc.test("Response contains access token", function() {
    lc.expect(data).toHaveProperty("access_token");
});

lc.test("Token is valid JWT format", function() {
    var parts = data.access_token.split(".");
    lc.expect(parts.length).toBe(3);
});

// Store tokens
if (data && data.access_token) {
    lc.env.set("access_token", data.access_token);
    console.log("Token stored successfully");
}
```

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

### Numeric Ranges

```javascript
var product = lc.response.body.json();

lc.test("Price is positive", function() {
    lc.expect(product.price).toBeGreaterThan(0);
});

lc.test("Rating is within range", function() {
    lc.expect(product.rating).toBeGreaterThanOrEqual(0);
    lc.expect(product.rating).toBeLessThanOrEqual(5);
});
```

## Test Suite Pattern

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
}

// Run all test groups
testStatusCode();
testResponseTime();
testResponseBody();
```

## See Also

- [lc.test](API-Test) - Testing API reference
- [lc.response](API-Response) - Response object
- [Authentication Examples](Examples-Auth)
- [Request Chaining Examples](Examples-Chaining)

---

*[Authentication Examples](Examples-Auth) | [Request Chaining Examples](Examples-Chaining)*
