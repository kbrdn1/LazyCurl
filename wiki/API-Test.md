# lc.test & lc.expect

Jest-like testing framework for validating API responses with fluent assertions.

**Availability**: Both (pre-request and post-response)

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

## lc.test

### test(name, fn)

Defines a named test case. The test **passes** if the function executes without throwing.

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

## lc.expect

### expect(value)

Creates an expectation object for fluent assertions.

```javascript
var status = lc.response.status;
lc.expect(status).toBe(200);

var data = lc.response.body.json();
lc.expect(data).not.toBeNull();
```

## Matchers

### toBe(expected)

Strict equality using `===`.

```javascript
lc.expect(lc.response.status).toBe(200);
lc.expect("hello").toBe("hello");
```

### toEqual(expected)

Deep equality for objects and arrays.

```javascript
lc.expect({ a: 1 }).toEqual({ a: 1 });
lc.expect([1, 2, 3]).toEqual([1, 2, 3]);
```

### toBeTruthy() / toBeFalsy()

```javascript
lc.expect(true).toBeTruthy();
lc.expect("text").toBeTruthy();
lc.expect(null).toBeFalsy();
lc.expect(0).toBeFalsy();
```

### toContain(substring)

```javascript
lc.expect("hello world").toContain("world");
lc.expect("user@example.com").toContain("@");
```

### toHaveProperty(name)

```javascript
var user = { id: 1, name: "John" };
lc.expect(user).toHaveProperty("id");
lc.expect(user).toHaveProperty("name");
```

### toMatch(pattern)

```javascript
lc.expect("user@example.com").toMatch("@");
lc.expect("hello123").toMatch("\\d+");
```

### toBeNull() / toBeUndefined() / toBeDefined()

```javascript
lc.expect(null).toBeNull();
lc.expect(obj.missing).toBeUndefined();
lc.expect(data).toBeDefined();
```

### toHaveLength(n)

```javascript
lc.expect([1, 2, 3]).toHaveLength(3);
lc.expect("hello").toHaveLength(5);
```

### Numeric Comparisons

```javascript
lc.expect(10).toBeGreaterThan(5);
lc.expect(5).toBeLessThan(10);
lc.expect(lc.response.status).toBeGreaterThanOrEqual(200);
lc.expect(lc.response.status).toBeLessThanOrEqual(299);
```

## Negation with .not

All matchers can be negated:

```javascript
lc.expect(200).not.toBe(404);
lc.expect("hello").not.toContain("goodbye");
lc.expect(data).not.toBeNull();
lc.expect(data).not.toHaveProperty("deleted");
```

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
    lc.expect(user.email).toMatch("@");
});
```

### List/Pagination Response

```javascript
var data = lc.response.body.json();

lc.test("Response is a paginated list", function() {
    lc.expect(data).toHaveProperty("items");
    lc.expect(data).toHaveProperty("total");
    lc.expect(data).toHaveProperty("page");
});

lc.test("Items array is not empty", function() {
    lc.expect(data.items.length).toBeGreaterThan(0);
});
```

### Error Response Testing

```javascript
lc.test("Returns 400 for invalid input", function() {
    lc.expect(lc.response.status).toBe(400);
});

lc.test("Error response has message", function() {
    var error = lc.response.body.json();
    lc.expect(error).toHaveProperty("message");
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
});

// Store tokens for subsequent requests
var data = lc.response.body.json();
if (data && data.access_token) {
    lc.env.set("access_token", data.access_token);
}
```

## See Also

- [lc.response](API-Response) - HTTP response access
- [Testing Examples](Examples-Testing)
- [Authentication Examples](Examples-Auth)

---

*[lc.env](API-Env) | [lc.cookies](API-Cookies)*
