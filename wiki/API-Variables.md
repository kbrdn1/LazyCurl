# lc.variables

Generate dynamic values like UUIDs, timestamps, and random data.

**Availability**: Both (pre-request and post-response)

## Quick Reference

| Method | Description | Example Output |
|--------|-------------|----------------|
| `uuid()` | UUID v4 | `550e8400-e29b-41d4-a716-446655440000` |
| `timestamp()` | Unix timestamp (seconds) | `1699123456` |
| `isoTimestamp()` | ISO 8601 timestamp | `2024-11-04T15:30:45.123Z` |
| `randomInt(min, max)` | Random integer | `42` |
| `randomFloat(min, max)` | Random float | `3.14159` |
| `randomString(length)` | Random alphanumeric | `aB3xK9mN` |
| `randomHex(length)` | Random hex string | `a1b2c3d4` |
| `randomEmail()` | Random email | `user_abc123@example.com` |
| `randomName()` | Random name | `John Smith` |
| `randomBoolean()` | Random true/false | `true` |
| `randomItem(array)` | Random array item | `"item2"` |
| `randomDate(start, end)` | Random date | `2024-06-15T12:00:00.000Z` |

## Methods

### uuid()

Generates a UUID v4.

```javascript
var id = lc.variables.uuid();
// "550e8400-e29b-41d4-a716-446655440000"
```

### timestamp()

Returns current Unix timestamp in seconds.

```javascript
var ts = lc.variables.timestamp();
// 1699123456
```

### isoTimestamp()

Returns current timestamp in ISO 8601 format.

```javascript
var ts = lc.variables.isoTimestamp();
// "2024-11-04T15:30:45.123Z"
```

### randomInt(min, max)

Generates random integer in range [min, max].

```javascript
var num = lc.variables.randomInt(1, 100);
// 42
```

### randomFloat(min, max)

Generates random float in range [min, max].

```javascript
var num = lc.variables.randomFloat(0, 1);
// 0.7234
```

### randomString(length)

Generates random alphanumeric string.

```javascript
var str = lc.variables.randomString(8);
// "aB3xK9mN"
```

### randomHex(length)

Generates random hexadecimal string.

```javascript
var hex = lc.variables.randomHex(16);
// "a1b2c3d4e5f67890"
```

### randomEmail()

Generates random email address.

```javascript
var email = lc.variables.randomEmail();
// "user_abc123@example.com"
```

### randomName()

Generates random full name.

```javascript
var name = lc.variables.randomName();
// "John Smith"
```

### randomBoolean()

Generates random boolean.

```javascript
var bool = lc.variables.randomBoolean();
// true or false
```

### randomItem(array)

Selects random item from array.

```javascript
var color = lc.variables.randomItem(["red", "green", "blue"]);
// "green"
```

### randomDate(start, end)

Generates random date between start and end.

```javascript
var start = new Date("2024-01-01");
var end = new Date("2024-12-31");
var date = lc.variables.randomDate(start, end);
// "2024-06-15T12:00:00.000Z"
```

## Examples

### Generate Test User Data

```javascript
// Pre-request: Create random test user
var testUser = {
    id: lc.variables.uuid(),
    name: lc.variables.randomName(),
    email: lc.variables.randomEmail(),
    age: lc.variables.randomInt(18, 65),
    active: lc.variables.randomBoolean(),
    createdAt: lc.variables.isoTimestamp()
};

lc.request.body.set(JSON.stringify(testUser));
console.log("Created test user: " + testUser.name);
```

### Add Request Metadata

```javascript
// Pre-request: Add metadata to requests
var body = lc.request.body.json() || {};

body.metadata = {
    requestId: lc.variables.uuid(),
    timestamp: lc.variables.isoTimestamp(),
    client: "LazyCurl"
};

lc.request.body.set(JSON.stringify(body));
lc.request.headers.set("X-Request-ID", body.metadata.requestId);
```

### Unique Test Data

```javascript
// Pre-request: Ensure unique test data
var uniqueSuffix = lc.variables.randomString(6);

var body = lc.request.body.json() || {};
body.email = "test_" + uniqueSuffix + "@example.com";
body.username = "user_" + uniqueSuffix;

lc.request.body.set(JSON.stringify(body));
lc.globals.set("test_suffix", uniqueSuffix);
```

### Random Test Scenarios

```javascript
// Pre-request: Randomize test scenario
var statuses = ["pending", "active", "completed", "cancelled"];
var priorities = ["low", "medium", "high", "critical"];

var body = {
    status: lc.variables.randomItem(statuses),
    priority: lc.variables.randomItem(priorities),
    score: lc.variables.randomFloat(0, 100).toFixed(2),
    dueDate: lc.variables.randomDate(new Date(), new Date("2025-12-31"))
};

lc.request.body.set(JSON.stringify(body));
console.log("Testing with status: " + body.status);
```

## See Also

- [lc.request](API-Request) - HTTP request manipulation
- [lc.crypto](API-Crypto) - Cryptographic functions
- [Request Chaining Examples](Examples-Chaining)

---

*[lc.base64](API-Base64) | [lc.info](API-Info)*
