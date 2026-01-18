# lc.test & lc.expect

Jest-like testing framework for validating API responses with fluent assertions.

**Availability**: Both (pre-request and post-response)

> For the complete API reference, see [Scripting API Reference](../scripting-api-reference.md#lctest--lcexpect).

## Table of Contents

- [Overview](#overview)
- [Quick Reference](#quick-reference)
- [lc.test](#lctest)
- [lc.expect](#lcexpect)
- [Matchers](#matchers)
- [Negation with .not](#negation-with-not)
- [Examples](#examples)
- [See Also](#see-also)

---

## Overview

LazyCurl provides a Jest-like testing API:

- **`lc.test(name, fn)`** - Define named test cases
- **`lc.expect(value)`** - Create fluent assertions

Test results are displayed in the **Tests** tab of the Response panel.

---

## Quick Reference

### Test Definition

| Function | Description |
|----------|-------------|
| `lc.test(name, fn)` | Define a named test case |
| `lc.expect(value)` | Create an expectation for assertions |

### Matchers

| Matcher | Description |
|---------|-------------|
| `.toBe(expected)` | Strict equality (`===`) |
| `.toEqual(expected)` | Deep equality |
| `.toBeTruthy()` | Value is truthy |
| `.toBeFalsy()` | Value is falsy |
| `.toContain(substring)` | String contains substring |
| `.toHaveProperty(name)` | Object has property |
| `.toMatch(pattern)` | Value matches regex/string |
| `.toBeNull()` | Value is `null` |
| `.toBeUndefined()` | Value is `undefined` |
| `.toBeDefined()` | Value is not null/undefined |
| `.toHaveLength(n)` | Array/string has length n |
| `.toBeGreaterThan(n)` | Number > n |
| `.toBeLessThan(n)` | Number < n |
| `.toBeGreaterThanOrEqual(n)` | Number >= n |
| `.toBeLessThanOrEqual(n)` | Number <= n |

---

## lc.test

### test(name, fn)

Defines a named test case. The test **passes** if the function executes without throwing an error.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Descriptive name for the test |
| `fn` | function | Yes | Test function containing assertions |

**Returns**: `void`

```javascript
lc.test("Response status is OK", function() {
    lc.expect(lc.response.status).toBe(200);
});

lc.test("User data is valid", function() {
    var data = lc.response.body.json();
    lc.expect(data).toHaveProperty("id");
    lc.expect(data).toHaveProperty("name");
    lc.expect(data.email).toMatch("@");
});
```

---

## lc.expect

### expect(value)

Creates an expectation object for fluent assertions.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `value` | any | Yes | The value to test |

**Returns**: `Expectation` - Chainable expectation object

```javascript
var status = lc.response.status;
lc.expect(status).toBe(200);

var data = lc.response.body.json();
lc.expect(data).not.toBeNull();
```

---

## Matchers

### toBe(expected)

Strict equality comparison using `===`.

```javascript
lc.expect(lc.response.status).toBe(200);
lc.expect(true).toBe(true);
lc.expect("hello").toBe("hello");
```

---

### toEqual(expected)

Deep equality comparison for objects and arrays.

```javascript
lc.expect({ a: 1 }).toEqual({ a: 1 });
lc.expect([1, 2, 3]).toEqual([1, 2, 3]);
```

---

### toBeTruthy()

Asserts the value is truthy (not `false`, `0`, `""`, `null`, `undefined`, `NaN`).

```javascript
lc.expect(true).toBeTruthy();
lc.expect(1).toBeTruthy();
lc.expect("text").toBeTruthy();
lc.expect({}).toBeTruthy();
```

---

### toBeFalsy()

Asserts the value is falsy.

```javascript
lc.expect(false).toBeFalsy();
lc.expect(0).toBeFalsy();
lc.expect("").toBeFalsy();
lc.expect(null).toBeFalsy();
```

---

### toContain(substring)

Asserts a string contains the given substring.

```javascript
lc.expect("hello world").toContain("world");
lc.expect("user@example.com").toContain("@");
```

---

### toHaveProperty(name)

Asserts an object has the specified property.

```javascript
var user = { id: 1, name: "John" };
lc.expect(user).toHaveProperty("id");
lc.expect(user).toHaveProperty("name");
```

---

### toMatch(pattern)

Asserts the value matches a regex pattern or contains a string.

```javascript
lc.expect("user@example.com").toMatch("@");
lc.expect("hello123").toMatch("\\d+");  // Regex pattern
```

---

### toBeNull()

Asserts the value is exactly `null`.

```javascript
lc.expect(null).toBeNull();
```

---

### toBeUndefined()

Asserts the value is exactly `undefined`.

```javascript
var obj = { a: 1 };
lc.expect(obj.b).toBeUndefined();
```

---

### toBeDefined()

Asserts the value is not `null` or `undefined`.

```javascript
var data = lc.response.body.json();
lc.expect(data).toBeDefined();
lc.expect(data.id).toBeDefined();
```

---

### toHaveLength(n)

Asserts an array or string has the specified length.

```javascript
lc.expect([1, 2, 3]).toHaveLength(3);
lc.expect("hello").toHaveLength(5);
```

---

### toBeGreaterThan(n)

Asserts a number is greater than the expected value.

```javascript
lc.expect(10).toBeGreaterThan(5);
lc.expect(lc.response.status).toBeGreaterThan(199);
```

---

### toBeLessThan(n)

Asserts a number is less than the expected value.

```javascript
lc.expect(5).toBeLessThan(10);
lc.expect(lc.response.time).toBeLessThan(2000);
```

---

### toBeGreaterThanOrEqual(n)

Asserts a number is greater than or equal to the expected value.

```javascript
lc.expect(lc.response.status).toBeGreaterThanOrEqual(200);
```

---

### toBeLessThanOrEqual(n)

Asserts a number is less than or equal to the expected value.

```javascript
lc.expect(lc.response.status).toBeLessThanOrEqual(299);
```

---

## Negation with .not

All matchers can be negated using the `.not` chain:

```javascript
lc.expect(200).not.toBe(404);
lc.expect("hello").not.toContain("goodbye");
lc.expect(data).not.toBeNull();
lc.expect(data).not.toHaveProperty("deleted");
lc.expect(null).not.toBeDefined();
lc.expect(lc.response.status).not.toBeGreaterThan(399);
```

---

## Examples

### Basic Response Validation

```javascript
lc.test("Status is 200 OK", function() {
    lc.expect(lc.response.status).toBe(200);
});

lc.test("Response time under 2 seconds", function() {
    lc.expect(lc.response.time).toBeLessThan(2000);
});

lc.test("Content-Type is JSON", function() {
    var contentType = lc.response.headers.get("Content-Type");
    lc.expect(contentType).toContain("application/json");
});
```

### User API Response

```javascript
lc.test("API returns valid user", function() {
    lc.expect(lc.response.status).toBe(200);

    var user = lc.response.body.json();
    lc.expect(user).not.toBeNull();
    lc.expect(user).toHaveProperty("id");
    lc.expect(user).toHaveProperty("name");
    lc.expect(user).toHaveProperty("email");
    lc.expect(user.email).toMatch("@");
});

lc.test("User has valid roles", function() {
    var user = lc.response.body.json();
    lc.expect(user.roles).toBeDefined();
    lc.expect(user.roles.length).toBeGreaterThan(0);
});
```

### List/Pagination Response

```javascript
var data = lc.response.body.json();

lc.test("Response is a paginated list", function() {
    lc.expect(data).toHaveProperty("items");
    lc.expect(data).toHaveProperty("total");
    lc.expect(data).toHaveProperty("page");
    lc.expect(data).toHaveProperty("pageSize");
});

lc.test("Items array is not empty", function() {
    lc.expect(data.items.length).toBeGreaterThan(0);
});

lc.test("Each item has required fields", function() {
    var firstItem = data.items[0];
    lc.expect(firstItem).toHaveProperty("id");
    lc.expect(firstItem).toHaveProperty("name");
});

// Log summary
console.log("Retrieved " + data.items.length + " of " + data.total + " items");
```

### Error Response Testing

```javascript
// Test for expected error response (e.g., validation error)
lc.test("Returns 400 for invalid input", function() {
    lc.expect(lc.response.status).toBe(400);
});

lc.test("Error response has message", function() {
    var error = lc.response.body.json();
    lc.expect(error).toHaveProperty("message");
    lc.expect(error.message).toBeDefined();
});

lc.test("Error response has code", function() {
    var error = lc.response.body.json();
    lc.expect(error).toHaveProperty("code");
    lc.expect(error.code).toBe("VALIDATION_ERROR");
});
```

### Authentication Flow

```javascript
lc.test("Login successful", function() {
    lc.expect(lc.response.status).toBe(200);
});

lc.test("Response contains tokens", function() {
    var data = lc.response.body.json();
    lc.expect(data).toHaveProperty("access_token");
    lc.expect(data).toHaveProperty("refresh_token");
    lc.expect(data).toHaveProperty("expires_in");
});

lc.test("Token is valid format", function() {
    var data = lc.response.body.json();
    // JWT tokens have 3 parts separated by dots
    var parts = data.access_token.split(".");
    lc.expect(parts.length).toBe(3);
});

// Store tokens for subsequent requests
var data = lc.response.body.json();
if (data && data.access_token) {
    lc.env.set("access_token", data.access_token);
    lc.env.set("refresh_token", data.refresh_token);
    console.log("Tokens saved to environment");
}
```

### Performance Testing

```javascript
lc.test("Response time is fast", function() {
    lc.expect(lc.response.time).toBeLessThan(500);
});

lc.test("Response time is acceptable", function() {
    lc.expect(lc.response.time).toBeLessThan(2000);
});

// Performance categories
var time = lc.response.time;
if (time < 200) {
    console.log("Performance: Excellent (" + time + "ms)");
} else if (time < 500) {
    console.log("Performance: Good (" + time + "ms)");
} else if (time < 1000) {
    console.log("Performance: Acceptable (" + time + "ms)");
} else {
    console.warn("Performance: Slow (" + time + "ms)");
}
```

---

## See Also

- [lc.response](./response.md) - HTTP response access
- [Testing Examples](../examples/testing.md)
- [Authentication Examples](../examples/authentication.md)

---

*[← lc.env](./env.md) | [lc.cookies →](./cookies.md)*
