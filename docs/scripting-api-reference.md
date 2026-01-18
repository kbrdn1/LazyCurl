# LazyCurl JavaScript Scripting API Reference

This document provides a comprehensive reference for all JavaScript APIs available in LazyCurl pre-request and post-response scripts.

## Table of Contents

- [Overview](#overview)
- [lc.request](#lcrequest)
- [lc.response](#lcresponse)
- [lc.env & lc.globals](#lcenv--lcglobals)
- [lc.test & lc.expect](#lctest--lcexpect)
- [lc.cookies](#lccookies)
- [lc.base64](#lcbase64)
- [lc.crypto](#lccrypto)
- [lc.variables](#lcvariables)
- [lc.info](#lcinfo)
- [console & lc.sendRequest](#console--lcsendrequest)

---

## Overview

LazyCurl uses the [Goja](https://github.com/dop251/goja) JavaScript runtime to execute scripts. Scripts have access to the `lc` global object which provides APIs for request/response manipulation, environment variables, testing, and utilities.

### Script Types

| Type              | Execution                    | Available APIs                                                                           |
| ----------------- | ---------------------------- | ---------------------------------------------------------------------------------------- |
| **Pre-request**   | Before HTTP request is sent  | `lc.request` (mutable), `lc.env`, `lc.globals`, `lc.cookies`, `lc.sendRequest`           |
| **Post-response** | After HTTP response received | `lc.request` (read-only), `lc.response`, `lc.env`, `lc.globals`, `lc.test`, `lc.cookies` |

### Quick Example

```javascript
// Pre-request: Add authentication
var token = lc.env.get("auth_token");
if (token) {
  lc.request.setHeader("Authorization", "Bearer " + token);
}

// Post-response: Validate and store data
lc.test("Status is 200", function () {
  lc.expect(lc.response.status).toBe(200);
});

var data = lc.response.json();
if (data.token) {
  lc.env.set("auth_token", data.token);
}
```

---

## lc.request

The `lc.request` object provides access to the current HTTP request data. It allows reading request properties and, in pre-request scripts, modifying the request before it is sent.

### Properties

| Property | Type     | Mutability                    | Description                                    |
| -------- | -------- | ----------------------------- | ---------------------------------------------- |
| `method` | `string` | Read-only                     | The HTTP method (GET, POST, PUT, DELETE, etc.) |
| `url`    | `string` | Read/Write (pre-request only) | The request URL including query string         |

#### method (Read-only)

Returns the HTTP method of the request. This property cannot be modified.

```javascript
// Read the HTTP method
var method = lc.request.method;
console.log("Method: " + method); // "GET", "POST", etc.
```

#### url

Returns or sets the request URL. In pre-request scripts, you can modify this property to change the target URL before the request is sent.

```javascript
// Read the URL
var url = lc.request.url;
console.log("URL: " + url);

// Modify the URL (pre-request only)
lc.request.url = "https://api.example.com/v2/users";

// Add query parameters dynamically
lc.request.url = lc.request.url + "?timestamp=" + Date.now();
```

### lc.request.headers

Object providing methods for header manipulation.

#### headers.get(name)

Returns the value of a header by name. Header lookup is case-insensitive.

```javascript
var contentType = lc.request.headers.get("Content-Type");
var auth = lc.request.headers.get("authorization"); // Case-insensitive
```

**Parameters:**

- `name` (string): The header name to retrieve

**Returns:** `string` | `undefined` - The header value, or `undefined` if not found

#### headers.set(name, value)

Sets or updates a header value. Available only in pre-request scripts.

```javascript
// Set a new header
lc.request.headers.set("X-Custom-Header", "custom-value");

// Update an existing header
lc.request.headers.set("Content-Type", "application/json");

// Add authentication dynamically
var token = lc.env.get("auth_token");
lc.request.headers.set("Authorization", "Bearer " + token);
```

**Parameters:**

- `name` (string): The header name
- `value` (string): The header value

#### headers.remove(name)

Removes a header from the request. Available only in pre-request scripts. Header lookup is case-insensitive.

```javascript
// Remove a header
lc.request.headers.remove("X-Deprecated-Header");
```

**Parameters:**

- `name` (string): The header name to remove

#### headers.all()

Returns a copy of all headers as a key-value object.

```javascript
var allHeaders = lc.request.headers.all();
console.log(JSON.stringify(allHeaders));
// { "Content-Type": "application/json", "Authorization": "Bearer ..." }

// Iterate over headers
var headers = lc.request.headers.all();
for (var name in headers) {
  console.log(name + ": " + headers[name]);
}
```

**Returns:** `object` - An object containing all headers as key-value pairs

### lc.request.body

Object providing methods for request body access and manipulation.

#### body.raw()

Returns the raw request body as a string.

```javascript
var rawBody = lc.request.body.raw();
console.log("Body length: " + rawBody.length);
```

**Returns:** `string` - The raw body content

#### body.json()

Parses the request body as JSON and returns the resulting object. Returns `null` if parsing fails or body is empty.

```javascript
var data = lc.request.body.json();
if (data) {
  console.log("User ID: " + data.userId);
  console.log("Name: " + data.name);
}
```

**Returns:** `object` | `null` - Parsed JSON object, or `null` if invalid/empty

#### body.set(content)

Sets the request body content. Available only in pre-request scripts.

```javascript
// Set body as string
lc.request.body.set('{"name": "John", "email": "john@example.com"}');

// Build body programmatically
var payload = {
  timestamp: Date.now(),
  requestId: Math.random().toString(36).substring(7),
  data: lc.request.body.json(),
};
lc.request.body.set(JSON.stringify(payload));
```

**Parameters:**

- `content` (string): The new body content

### lc.request.params

Object providing read-only access to URL query parameters.

#### params.get(name)

Returns the first value of a query parameter.

```javascript
// URL: https://api.example.com/users?page=1&limit=10
var page = lc.request.params.get("page"); // "1"
var limit = lc.request.params.get("limit"); // "10"
var missing = lc.request.params.get("sort"); // undefined
```

**Parameters:**

- `name` (string): The parameter name

**Returns:** `string` | `undefined` - The parameter value, or `undefined` if not found

#### params.getAll(name)

Returns all values for a query parameter (for parameters with multiple values).

```javascript
// URL: https://api.example.com/search?tag=javascript&tag=golang&tag=rust
var tags = lc.request.params.getAll("tag");
// ["javascript", "golang", "rust"]

tags.forEach(function (tag) {
  console.log("Tag: " + tag);
});
```

**Parameters:**

- `name` (string): The parameter name

**Returns:** `string[]` - Array of all values for the parameter (empty array if not found)

#### params.has(name)

Checks if a query parameter exists.

```javascript
// URL: https://api.example.com/users?active=true
if (lc.request.params.has("active")) {
  console.log("Filtering by active status");
}
```

**Parameters:**

- `name` (string): The parameter name to check

**Returns:** `boolean` - `true` if the parameter exists, `false` otherwise

#### params.keys()

Returns an array of all query parameter names.

```javascript
// URL: https://api.example.com/search?q=test&page=1&limit=20
var keys = lc.request.params.keys();
// ["q", "page", "limit"]
```

**Returns:** `string[]` - Array of parameter names

#### params.all()

Returns all query parameters as a key-value object. For parameters with multiple values, only the first value is included.

```javascript
// URL: https://api.example.com/users?page=1&limit=10&sort=name
var params = lc.request.params.all();
// { "page": "1", "limit": "10", "sort": "name" }

console.log("Page: " + params.page);
```

**Returns:** `object` - Object containing all parameters as key-value pairs

### Mutability Notes

| Context                  | Properties               | Behavior                                            |
| ------------------------ | ------------------------ | --------------------------------------------------- |
| **Pre-request script**   | `url`, `headers`, `body` | Fully mutable - changes affect the outgoing request |
| **Post-response script** | All properties           | Read-only - reflects the request as it was sent     |

---

## lc.response

The `lc.response` object provides read-only access to HTTP response data in post-response scripts. This object is immutable and only available after the HTTP request has completed.

> **Availability**: Post-response scripts only. Not available in pre-request scripts.

### Properties

| Property     | Type     | Description                                                           |
| ------------ | -------- | --------------------------------------------------------------------- |
| `status`     | `number` | HTTP status code (e.g., `200`, `404`, `500`)                          |
| `statusText` | `string` | Full status text including code (e.g., `"200 OK"`, `"404 Not Found"`) |
| `time`       | `number` | Response time in milliseconds                                         |

```javascript
// Check status code
if (lc.response.status === 200) {
  lc.console.log("Request successful");
}

// Log response timing
lc.console.log("Response received in " + lc.response.time + "ms");

// Use full status text for logging
lc.console.log("Status: " + lc.response.statusText);
```

### Methods

#### lc.response.headers.get(name)

Returns the value of a response header. Header name lookup is case-insensitive.

**Parameters:**

- `name` (string): The header name to retrieve
  **Returns:** `string | undefined` - The header value, or `undefined` if not found ```javascript
  // Get content type
  var contentType = lc.response.headers.get("Content-Type");
  lc.console.log("Content-Type: " + contentType);

// Case-insensitive lookup var auth = lc.response.headers.get("x-auth-token");
var authAlt = lc.response.headers.get("X-Auth-Token"); // Same result

````
#### lc.response.headers.all() Returns a copy of all response headers as a key-value object. **Returns:** `object` - Object containing all headers ```javascript // Get all headers
var headers = lc.response.headers.all();

// Iterate over headers
for (var key in headers) {
  lc.console.log(key + ": " + headers[key]);
}
````

#### lc.response.body.raw()

Returns the raw response body as a string.

**Returns:** `string` - The response body content

```javascript
// Get raw body
var rawBody = lc.response.body.raw();
lc.console.log("Body length: " + rawBody.length + " characters");
```

#### lc.response.body.json()

Parses the response body as JSON and returns the resulting object. Returns `null` if the body is empty or cannot be parsed as valid JSON.

**Returns:** `object | array | null` - Parsed JSON data, or `null` on parse failure

```javascript
// Parse JSON response
var data = lc.response.body.json();

if (data !== null) {
  lc.console.log("User ID: " + data.id);
  lc.console.log("Username: " + data.username);
} else {
  lc.console.error("Failed to parse response as JSON");
}
```

### Complete Example

```javascript
// Comprehensive post-response script
lc.console.log("Status: " + lc.response.statusText);
lc.console.log("Time: " + lc.response.time + "ms");

if (lc.response.status >= 200 && lc.response.status < 300) {
  var data = lc.response.body.json();

  if (data && data.accessToken) {
    lc.env.set("access_token", data.accessToken);
    lc.console.log("Token saved");
  }
} else if (lc.response.status === 401) {
  lc.console.error("Unauthorized - check credentials");
}

lc.test("Authentication successful", function () {
  lc.expect(lc.response.status).toBe(200);
});
```

---

## lc.env & lc.globals

LazyCurl provides two distinct mechanisms for managing variables in scripts: **environment variables** (`lc.env`) for project-level configuration and **global variables** (`lc.globals`) for cross-request data sharing within a session.

### lc.env - Environment Variables

Environment variables are tied to your active environment file and persist to disk. Changes made via `lc.env.set()` are saved to the environment file after script execution.

| Method     | Signature                 | Description                                                                        |
| ---------- | ------------------------- | ---------------------------------------------------------------------------------- |
| `get`      | `lc.env.get(name)`        | Retrieves the value of an environment variable. Returns empty string if not found. |
| `set`      | `lc.env.set(name, value)` | Sets an environment variable. Value is persisted to the environment file.          |
| `unset`    | `lc.env.unset(name)`      | Removes an environment variable from the environment.                              |
| `has`      | `lc.env.has(name)`        | Returns `true` if the variable exists, `false` otherwise.                          |
| `toObject` | `lc.env.toObject()`       | Returns all environment variables as a JavaScript object.                          |

```javascript
// Get a variable
var baseUrl = lc.env.get("base_url");

// Set a variable (persisted to environment file)
lc.env.set("last_run", new Date().toISOString());

// Check if variable exists before using
if (lc.env.has("api_key")) {
  lc.request.setHeader("X-API-Key", lc.env.get("api_key"));
}

// Remove a variable
lc.env.unset("temp_token");

// Get all variables as an object
var allVars = lc.env.toObject();
```

### lc.globals - Session Global Variables

Global variables exist only in memory during the current LazyCurl session. They persist across multiple request executions but are lost when you close the application. Unlike `lc.env`, globals can store any JavaScript value (objects, arrays, numbers, booleans) not just strings.

| Method     | Signature                     | Description                                               |
| ---------- | ----------------------------- | --------------------------------------------------------- |
| `get`      | `lc.globals.get(name)`        | Retrieves a global variable. Returns `null` if not found. |
| `set`      | `lc.globals.set(name, value)` | Sets a global variable. Accepts any JavaScript value.     |
| `unset`    | `lc.globals.unset(name)`      | Removes a global variable.                                |
| `has`      | `lc.globals.has(name)`        | Returns `true` if the variable exists, `false` otherwise. |
| `clear`    | `lc.globals.clear()`          | Removes all global variables.                             |
| `toObject` | `lc.globals.toObject()`       | Returns all global variables as a JavaScript object.      |

```javascript
// Store complex data structures
lc.globals.set("user", {
  id: 123,
  name: "John Doe",
  roles: ["admin", "user"],
});

// Retrieve and use stored data
var user = lc.globals.get("user");
if (user) {
  lc.console.log("User ID: " + user.id);
}

// Store counters or state
var count = lc.globals.get("request_count") || 0;
lc.globals.set("request_count", count + 1);

// Clear all globals
lc.globals.clear();
```

### Comparison: lc.env vs lc.globals

| Feature         | lc.env                             | lc.globals                         |
| --------------- | ---------------------------------- | ---------------------------------- |
| **Persistence** | Saved to environment file on disk  | In-memory only (lost on app close) |
| **Scope**       | Tied to active environment         | Session-wide, all requests         |
| **Value Types** | Strings only                       | Any JavaScript value               |
| **Use Case**    | Configuration, API keys, base URLs | Request chaining, temporary state  |

### Request Chaining Example

```javascript
// Request 1 - Post-response: Store token
var data = lc.response.json();
if (data && data.access_token) {
  lc.globals.set("access_token", data.access_token);
  lc.env.set("access_token", data.access_token);
}

// Request 2 - Pre-request: Use token
var token = lc.globals.get("access_token") || lc.env.get("access_token");
if (token) {
  lc.request.setHeader("Authorization", "Bearer " + token);
}
```

---

## lc.test & lc.expect

The scripting API provides a Jest-like testing framework with `lc.test()` for organizing tests and `lc.expect()` for fluent assertions.

### lc.test(name, fn)

Defines a named test case. The test passes if the function executes without throwing an error.

```javascript
lc.test("Response status is OK", function () {
  lc.expect(lc.response.status).toBe(200);
});

lc.test("User data is valid", function () {
  var data = lc.response.body.json();
  lc.expect(data).toHaveProperty("id");
  lc.expect(data.name).toBeDefined();
});
```

### lc.expect(value)

Creates an expectation object for fluent assertions. Returns a chainable object with matcher methods.

### Matchers

| Matcher                     | Description                         |
| --------------------------- | ----------------------------------- |
| `toBe(expected)`            | Strict equality comparison          |
| `toEqual(expected)`         | Deep equality comparison            |
| `toBeTruthy()`              | Asserts value is truthy             |
| `toBeFalsy()`               | Asserts value is falsy              |
| `toContain(substring)`      | Asserts string contains substring   |
| `toHaveProperty(name)`      | Asserts object has property         |
| `toMatch(pattern)`          | Asserts value matches regex         |
| `toBeNull()`                | Asserts value is null               |
| `toBeUndefined()`           | Asserts value is undefined          |
| `toBeDefined()`             | Asserts value is not null/undefined |
| `toHaveLength(n)`           | Asserts length of string/array      |
| `toBeGreaterThan(n)`        | Asserts number > n                  |
| `toBeLessThan(n)`           | Asserts number < n                  |
| `toBeGreaterThanOrEqual(n)` | Asserts number >= n                 |
| `toBeLessThanOrEqual(n)`    | Asserts number <= n                 |

```javascript
lc.expect(200).toBe(200); // Pass
lc.expect({ a: 1 }).toEqual({ a: 1 }); // Pass
lc.expect("hello").toContain("ell"); // Pass
lc.expect(data).toHaveProperty("id"); // Pass
lc.expect("user@example.com").toMatch("@"); // Pass
lc.expect([1, 2, 3]).toHaveLength(3); // Pass
lc.expect(lc.response.time).toBeLessThan(1000); // Pass
```

### Negation with .not

All matchers can be negated using the `.not` chain:

```javascript
lc.expect(200).not.toBe(404);
lc.expect("hello").not.toContain("goodbye");
lc.expect(data).not.toHaveProperty("deleted");
lc.expect(null).not.toBeDefined();
```

### Complete Example

```javascript
lc.test("API returns valid user", function () {
  lc.expect(lc.response.status).toBe(200);

  var user = lc.response.body.json();
  lc.expect(user).not.toBeNull();
  lc.expect(user).toHaveProperty("id");
  lc.expect(user.email).toMatch("@");
  lc.expect(user.roles).toHaveLength(2);
});

lc.test("Response time acceptable", function () {
  lc.expect(lc.response.time).toBeLessThan(2000);
});
```

---

## lc.cookies

Cookie management API for handling HTTP cookies in pre-request and post-response scripts.

### Methods

| Method     | Signature                               | Description                         |
| ---------- | --------------------------------------- | ----------------------------------- |
| `get`      | `lc.cookies.get(name)`                  | Get cookie value by name            |
| `getAll`   | `lc.cookies.getAll()`                   | Get all cookies as array of objects |
| `set`      | `lc.cookies.set(name, value, options?)` | Set a cookie                        |
| `delete`   | `lc.cookies.delete(name)`               | Remove a cookie                     |
| `clear`    | `lc.cookies.clear()`                    | Remove all cookies                  |
| `has`      | `lc.cookies.has(name)`                  | Check if cookie exists              |
| `toHeader` | `lc.cookies.toHeader()`                 | Generate Cookie header string       |

### set() Options

| Property   | Type    | Description                       |
| ---------- | ------- | --------------------------------- |
| `domain`   | string  | Domain scope for the cookie       |
| `path`     | string  | Path scope for the cookie         |
| `secure`   | boolean | Cookie sent only over HTTPS       |
| `httpOnly` | boolean | Cookie inaccessible to JavaScript |
| `expires`  | string  | Expiration date (RFC1123 format)  |

```javascript
// Get a cookie
var sessionId = lc.cookies.get("session_id");

// Set a cookie with options
lc.cookies.set("auth_token", "abc123xyz", {
  domain: "api.example.com",
  path: "/",
  secure: true,
  httpOnly: true,
});

// Check and use cookie
if (lc.cookies.has("csrf_token")) {
  lc.request.setHeader("X-CSRF-Token", lc.cookies.get("csrf_token"));
}

// Generate Cookie header
lc.request.setHeader("Cookie", lc.cookies.toHeader());

// Clear all cookies
lc.cookies.clear();
```

### CSRF Token Workflow

```javascript
// Post-response: Capture CSRF token
if (lc.cookies.has("csrf_token")) {
  lc.env.set("csrf_token", lc.cookies.get("csrf_token"));
}

// Pre-request: Use CSRF token
if (lc.cookies.has("csrf_token")) {
  lc.request.setHeader("X-CSRF-Token", lc.cookies.get("csrf_token"));
  lc.request.setHeader("Cookie", lc.cookies.toHeader());
}
```

---

## lc.base64

Base64 encoding and decoding utilities. Global `btoa()` and `atob()` functions are also available for browser-style compatibility.

### Methods

| Method                      | Description                         |
| --------------------------- | ----------------------------------- |
| `lc.base64.encode(data)`    | Encode string to Base64             |
| `lc.base64.decode(encoded)` | Decode Base64 string                |
| `btoa(data)`                | Global function, same as `encode()` |
| `atob(encoded)`             | Global function, same as `decode()` |

```javascript
// Encode/decode
var encoded = lc.base64.encode("Hello, World!");
// Returns: "SGVsbG8sIFdvcmxkIQ=="

var decoded = lc.base64.decode("SGVsbG8sIFdvcmxkIQ==");
// Returns: "Hello, World!"

// Browser-style functions
var encoded = btoa("username:password");
var decoded = atob("dXNlcm5hbWU6cGFzc3dvcmQ=");

// Basic Auth header
var credentials = btoa("username:password");
lc.request.setHeader("Authorization", "Basic " + credentials);
```

### Edge Cases

| Scenario       | Behavior                            |
| -------------- | ----------------------------------- |
| No argument    | Returns empty string `""`           |
| Empty string   | Returns empty string                |
| Invalid Base64 | Returns empty string (no exception) |

---

## lc.crypto

Cryptographic hash functions and HMAC operations. All functions return **lowercase hex-encoded strings**.

### Hash Functions

| Function                 | Output Length | Description         |
| ------------------------ | ------------- | ------------------- |
| `lc.crypto.md5(data)`    | 32 chars      | MD5 hash (legacy)   |
| `lc.crypto.sha1(data)`   | 40 chars      | SHA-1 hash (legacy) |
| `lc.crypto.sha256(data)` | 64 chars      | SHA-256 hash        |
| `lc.crypto.sha512(data)` | 128 chars     | SHA-512 hash        |

```javascript
var hash = lc.crypto.sha256("hello world");
// Returns: "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
```

### HMAC Functions

| Function                             | Output Length | Description           |
| ------------------------------------ | ------------- | --------------------- |
| `lc.crypto.hmacSha1(data, secret)`   | 40 chars      | HMAC-SHA1 (OAuth 1.0) |
| `lc.crypto.hmacSha256(data, secret)` | 64 chars      | HMAC-SHA256           |
| `lc.crypto.hmacSha512(data, secret)` | 128 chars     | HMAC-SHA512           |

```javascript
var signature = lc.crypto.hmacSha256("message", "secret");
// Returns: "6e9ef29b75fffc5b7abae527d58fdadb2fe42e7219011976917343065f58ed4a"
```

### Use Cases

```javascript
// API Request Signing
var timestamp = Math.floor(Date.now() / 1000).toString();
var body = lc.request.body || "";
var signature = lc.crypto.hmacSha256(
  timestamp + body,
  lc.env.get("api_secret"),
);

lc.request.setHeader("X-Timestamp", timestamp);
lc.request.setHeader("X-Signature", signature);

// Webhook Signature Verification
var payload = lc.response.body;
var expected =
  "sha256=" + lc.crypto.hmacSha256(payload, lc.env.get("webhook_secret"));
var received = lc.response.getHeader("X-Hub-Signature-256");

lc.test("Webhook signature valid", function () {
  lc.expect(received).toBe(expected);
});
```

---

## lc.variables

Dynamic variable generation for creating test data.

### Functions

| Function                | Returns | Description                      |
| ----------------------- | ------- | -------------------------------- |
| `uuid()`                | string  | UUID v4                          |
| `timestamp()`           | number  | Unix timestamp (seconds)         |
| `timestampMs()`         | number  | Unix timestamp (milliseconds)    |
| `isoTimestamp()`        | string  | ISO 8601 timestamp               |
| `randomInt(min?, max?)` | number  | Random integer (default 0-100)   |
| `randomFloat()`         | number  | Random float 0-1                 |
| `randomString(length?)` | string  | Alphanumeric string (default 16) |
| `randomHex(length?)`    | string  | Hex string (default 16)          |
| `randomEmail()`         | string  | Random email address             |
| `randomFirstName()`     | string  | Random first name                |
| `randomLastName()`      | string  | Random last name                 |
| `randomBoolean()`       | boolean | Random true/false                |

```javascript
// Generate test data
var user = {
  id: lc.variables.uuid(),
  email: lc.variables.randomEmail(),
  firstName: lc.variables.randomFirstName(),
  lastName: lc.variables.randomLastName(),
  isActive: lc.variables.randomBoolean(),
  createdAt: lc.variables.isoTimestamp(),
  score: lc.variables.randomInt(0, 100),
  apiKey: lc.variables.randomHex(32),
};

lc.request.body = JSON.stringify(user);
lc.request.setHeader("X-Request-ID", lc.variables.uuid());
```

---

## lc.info

Read-only contextual information about the current script execution.

### Properties

| Property          | Type                | Description                          |
| ----------------- | ------------------- | ------------------------------------ |
| `scriptType`      | string              | `"pre-request"` or `"post-response"` |
| `requestName`     | string \| undefined | Name of the request                  |
| `requestId`       | string \| undefined | ID of the request                    |
| `collectionName`  | string \| undefined | Name of the collection               |
| `environmentName` | string \| undefined | Active environment name              |
| `iteration`       | number              | Current iteration (default: 1)       |

```javascript
// Conditional logic based on script type
if (lc.info.scriptType === "pre-request") {
  lc.request.setHeader("Authorization", "Bearer " + lc.env.get("token"));
}

// Environment-specific behavior
if (lc.info.environmentName === "production") {
  lc.console.warn("Running against production!");
}

// Iteration-aware testing
if (lc.info.iteration === 1) {
  lc.test.assertStatus(201); // First iteration creates
} else {
  lc.test.assertStatus(200); // Subsequent iterations update
}

// Logging with context
lc.console.log(
  "[" +
    lc.info.collectionName +
    "/" +
    lc.info.requestName +
    "] Status: " +
    lc.response.status,
);
```

---

## console & lc.sendRequest

### Console API

Standard JavaScript-style logging functions. All output is captured and displayed in the Console tab.

| Method                   | Description            |
| ------------------------ | ---------------------- |
| `console.log(args...)`   | General logging        |
| `console.info(args...)`  | Informational messages |
| `console.warn(args...)`  | Warning messages       |
| `console.error(args...)` | Error messages         |
| `console.debug(args...)` | Debug messages         |

```javascript
console.log("Request sent successfully");
console.info("Processing", 42, "items");
console.warn("Rate limit approaching:", 95, "%");
console.error("Failed to parse response");
console.log({ name: "John", age: 30 }); // Objects are JSON-formatted
```

### lc.sendRequest

Enables request chaining by sending HTTP requests from within scripts.

#### Syntax

```javascript
lc.sendRequest(options, callback);
```

#### Options Object

| Property  | Type   | Required | Description                                       |
| --------- | ------ | -------- | ------------------------------------------------- |
| `url`     | string | Yes      | Target URL (supports `{{variable}}` substitution) |
| `method`  | string | No       | HTTP method (default: `"GET"`)                    |
| `headers` | object | No       | Key-value pairs for headers                       |
| `body`    | any    | No       | Request body (objects are JSON-stringified)       |

#### Callback Response

| Property      | Type     | Description        |
| ------------- | -------- | ------------------ |
| `status`      | number   | HTTP status code   |
| `statusText`  | string   | HTTP status text   |
| `time`        | number   | Response time (ms) |
| `headers`     | object   | Response headers   |
| `body.raw`    | string   | Raw response body  |
| `body.json()` | function | Parse body as JSON |

```javascript
// Basic request
lc.sendRequest(
  {
    url: "https://api.example.com/users",
    method: "GET",
  },
  function (err, response) {
    if (err) {
      console.error("Request failed:", err);
      return;
    }
    console.log("Status:", response.status);
    console.log("Users:", response.body.json());
  },
);

// POST with body
lc.sendRequest(
  {
    url: "{{base_url}}/auth/login",
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username: "admin", password: "secret" }),
  },
  function (err, response) {
    if (err) return;

    var data = response.body.json();
    if (data && data.token) {
      lc.env.set("auth_token", data.token);
    }
  },
);
```

### OAuth2 Token Refresh Example

```javascript
// Pre-request: Auto-refresh expired token
var tokenExpiry = lc.globals.get("token_expiry");
var now = Date.now();

if (!tokenExpiry || now >= tokenExpiry) {
  console.info("Token expired, refreshing...");

  lc.sendRequest(
    {
      url: "{{auth_url}}/oauth/token",
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body: "grant_type=client_credentials&client_id={{client_id}}&client_secret={{client_secret}}",
    },
    function (err, response) {
      if (err || response.status !== 200) return;

      var data = response.body.json();
      lc.env.set("access_token", data.access_token);
      lc.globals.set("token_expiry", now + data.expires_in * 1000 - 60000);
    },
  );
}

lc.request.setHeader("Authorization", "Bearer " + lc.env.get("access_token"));
```

---

## Summary

| API                     | Purpose                           | Availability                                     |
| ----------------------- | --------------------------------- | ------------------------------------------------ |
| `lc.request`            | HTTP request manipulation         | Pre-request (mutable), Post-response (read-only) |
| `lc.response`           | HTTP response access              | Post-response only                               |
| `lc.env`                | Environment variables (persisted) | Both                                             |
| `lc.globals`            | Session variables (in-memory)     | Both                                             |
| `lc.test` / `lc.expect` | Testing and assertions            | Both                                             |
| `lc.cookies`            | Cookie management                 | Both                                             |
| `lc.base64`             | Base64 encoding/decoding          | Both                                             |
| `lc.crypto`             | Cryptographic functions           | Both                                             |
| `lc.variables`          | Dynamic data generation           | Both                                             |
| `lc.info`               | Execution context info            | Both                                             |
| `console`               | Logging                           | Both                                             |
| `lc.sendRequest`        | Request chaining                  | Both                                             |
