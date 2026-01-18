# lc.variables

Dynamic variable generation for creating test data, unique identifiers, and random values.

**Availability**: Both (pre-request and post-response)

> For the complete API reference, see [Scripting API Reference](../scripting-api-reference.md#lcvariables).

## Table of Contents

- [Overview](#overview)
- [Quick Reference](#quick-reference)
- [Methods](#methods)
- [Examples](#examples)
- [See Also](#see-also)

---

## Overview

The `lc.variables` API generates dynamic data for:

- Unique identifiers (UUIDs)
- Timestamps in various formats
- Random numbers, strings, and hex values
- Fake user data (names, emails)
- Boolean values

Perfect for creating test data without hardcoding values.

---

## Quick Reference

| Method | Returns | Description |
|--------|---------|-------------|
| `uuid()` | string | UUID v4 |
| `timestamp()` | number | Unix timestamp (seconds) |
| `timestampMs()` | number | Unix timestamp (milliseconds) |
| `isoTimestamp()` | string | ISO 8601 timestamp |
| `randomInt(min?, max?)` | number | Random integer (default 0-100) |
| `randomFloat()` | number | Random float 0-1 |
| `randomString(length?)` | string | Alphanumeric string (default 16) |
| `randomHex(length?)` | string | Hex string (default 16) |
| `randomEmail()` | string | Random email address |
| `randomFirstName()` | string | Random first name |
| `randomLastName()` | string | Random last name |
| `randomBoolean()` | boolean | Random true/false |

---

## Methods

### uuid()

Generates a UUID v4 (universally unique identifier).

**Returns**: `string` - UUID in format `xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx`

```javascript
var id = lc.variables.uuid();
// Returns: "550e8400-e29b-41d4-a716-446655440000"
```

---

### timestamp()

Returns the current Unix timestamp in **seconds**.

**Returns**: `number` - Seconds since Unix epoch

```javascript
var ts = lc.variables.timestamp();
// Returns: 1703145600
```

---

### timestampMs()

Returns the current Unix timestamp in **milliseconds**.

**Returns**: `number` - Milliseconds since Unix epoch

```javascript
var tsMs = lc.variables.timestampMs();
// Returns: 1703145600000
```

---

### isoTimestamp()

Returns the current time as an ISO 8601 formatted string.

**Returns**: `string` - ISO 8601 timestamp

```javascript
var iso = lc.variables.isoTimestamp();
// Returns: "2024-12-21T10:00:00.000Z"
```

---

### randomInt(min?, max?)

Generates a random integer within a range.

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `min` | number | No | 0 | Minimum value (inclusive) |
| `max` | number | No | 100 | Maximum value (inclusive) |

**Returns**: `number` - Random integer

```javascript
var num1 = lc.variables.randomInt();        // 0-100
var num2 = lc.variables.randomInt(1, 10);   // 1-10
var num3 = lc.variables.randomInt(100, 999); // 100-999
```

---

### randomFloat()

Generates a random floating-point number between 0 and 1.

**Returns**: `number` - Random float (0 to 1)

```javascript
var f = lc.variables.randomFloat();
// Returns: 0.7234567891234567
```

---

### randomString(length?)

Generates a random alphanumeric string.

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `length` | number | No | 16 | String length |

**Returns**: `string` - Random alphanumeric string

```javascript
var str1 = lc.variables.randomString();     // 16 characters
var str2 = lc.variables.randomString(8);    // 8 characters
var str3 = lc.variables.randomString(32);   // 32 characters
// Returns: "a1B2c3D4e5F6g7H8"
```

---

### randomHex(length?)

Generates a random hexadecimal string.

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `length` | number | No | 16 | String length |

**Returns**: `string` - Random hex string (lowercase)

```javascript
var hex1 = lc.variables.randomHex();      // 16 characters
var hex2 = lc.variables.randomHex(32);    // 32 characters
// Returns: "a1b2c3d4e5f67890"
```

---

### randomEmail()

Generates a random email address.

**Returns**: `string` - Random email

```javascript
var email = lc.variables.randomEmail();
// Returns: "john.smith.5678@example.com"
```

---

### randomFirstName()

Generates a random first name.

**Returns**: `string` - Random first name

```javascript
var firstName = lc.variables.randomFirstName();
// Returns: "John", "Jane", "Michael", etc.
```

---

### randomLastName()

Generates a random last name.

**Returns**: `string` - Random last name

```javascript
var lastName = lc.variables.randomLastName();
// Returns: "Smith", "Johnson", "Williams", etc.
```

---

### randomBoolean()

Generates a random boolean value.

**Returns**: `boolean` - `true` or `false`

```javascript
var bool = lc.variables.randomBoolean();
// Returns: true or false
```

---

## Examples

### Generate Test User Data

```javascript
// Pre-request: Create complete test user
var user = {
    id: lc.variables.uuid(),
    email: lc.variables.randomEmail(),
    firstName: lc.variables.randomFirstName(),
    lastName: lc.variables.randomLastName(),
    age: lc.variables.randomInt(18, 65),
    isActive: lc.variables.randomBoolean(),
    createdAt: lc.variables.isoTimestamp(),
    apiKey: lc.variables.randomHex(32)
};

lc.request.body.set(JSON.stringify(user));
console.log("Created test user: " + user.email);
```

### Unique Request Identifiers

```javascript
// Pre-request: Add unique identifiers
lc.request.headers.set("X-Request-ID", lc.variables.uuid());
lc.request.headers.set("X-Correlation-ID", lc.variables.randomHex(16));
lc.request.headers.set("X-Timestamp", lc.variables.timestamp().toString());
```

### Dynamic Test Data

```javascript
// Pre-request: Create order with random products
var order = {
    orderId: lc.variables.uuid(),
    customerId: lc.env.get("customer_id"),
    items: [
        {
            productId: lc.variables.randomHex(8),
            quantity: lc.variables.randomInt(1, 5),
            price: parseFloat((lc.variables.randomFloat() * 100).toFixed(2))
        },
        {
            productId: lc.variables.randomHex(8),
            quantity: lc.variables.randomInt(1, 3),
            price: parseFloat((lc.variables.randomFloat() * 50).toFixed(2))
        }
    ],
    createdAt: lc.variables.isoTimestamp()
};

lc.request.body.set(JSON.stringify(order));
```

### Generate Multiple Records

```javascript
// Pre-request: Create batch of test users
var users = [];

for (var i = 0; i < 5; i++) {
    users.push({
        id: lc.variables.uuid(),
        email: lc.variables.randomEmail(),
        name: lc.variables.randomFirstName() + " " + lc.variables.randomLastName(),
        active: lc.variables.randomBoolean()
    });
}

lc.request.body.set(JSON.stringify({ users: users }));
console.log("Created " + users.length + " test users");
```

### Idempotency Key Generation

```javascript
// Pre-request: Generate idempotency key for payment
var idempotencyKey = lc.variables.uuid();
lc.request.headers.set("Idempotency-Key", idempotencyKey);

// Store for potential retry
lc.globals.set("last_idempotency_key", idempotencyKey);
console.log("Idempotency key: " + idempotencyKey);
```

### Randomized Query Parameters

```javascript
// Pre-request: Add random pagination
var page = lc.variables.randomInt(1, 10);
var limit = [10, 20, 50, 100][lc.variables.randomInt(0, 3)];

var url = lc.request.url;
var separator = url.includes("?") ? "&" : "?";
lc.request.url = url + separator + "page=" + page + "&limit=" + limit;

console.log("Requesting page " + page + " with limit " + limit);
```

### Test Data with Timestamps

```javascript
// Pre-request: Create event with timestamps
var event = {
    id: lc.variables.uuid(),
    type: "user_action",
    timestamp: lc.variables.timestampMs(),
    iso_timestamp: lc.variables.isoTimestamp(),
    user: {
        id: lc.variables.randomHex(12),
        session: lc.variables.randomString(24)
    },
    data: {
        action: "click",
        target: "button_" + lc.variables.randomInt(1, 10)
    }
};

lc.request.body.set(JSON.stringify(event));
```

---

## See Also

- [lc.request](./request.md) - HTTP request manipulation
- [lc.crypto](./crypto.md) - Cryptographic functions
- [Testing Examples](../examples/testing.md)

---

*[← lc.base64](./base64.md) | [lc.info →](./info.md)*
