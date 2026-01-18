# Getting Started

This guide will help you start using LazyCurl's JavaScript scripting API.

## Installation

### macOS (Homebrew)

```bash
brew install lazycurl
```

### From Source

```bash
git clone https://github.com/kbrdn1/LazyCurl.git
cd LazyCurl
make build
./bin/lazycurl
```

### Binary Releases

Download the latest release from [GitHub Releases](https://github.com/kbrdn1/LazyCurl/releases).

## Your First Script

LazyCurl scripts run in two phases:

1. **Pre-request**: Modify the request before it's sent
2. **Post-response**: Process the response and run tests

### Writing a Pre-request Script

1. Open a request in LazyCurl
2. Press `Tab` to navigate to the **Pre-request** tab
3. Write your script:

```javascript
// Add timestamp header
lc.request.headers.set("X-Timestamp", Date.now().toString());

// Add authentication from environment
var token = lc.env.get("auth_token");
if (token) {
    lc.request.headers.set("Authorization", "Bearer " + token);
}

console.log("Request prepared");
```

### Writing a Post-response Script

1. Navigate to the **Post-response** tab
2. Write your script:

```javascript
// Log response info
console.log("Status:", lc.response.status);
console.log("Time:", lc.response.time + "ms");

// Run tests
lc.test("Status is 200", function() {
    lc.expect(lc.response.status).toBe(200);
});

lc.test("Response is JSON", function() {
    var contentType = lc.response.headers.get("Content-Type");
    lc.expect(contentType).toContain("application/json");
});

// Extract and store data
var data = lc.response.body.json();
if (data && data.access_token) {
    lc.env.set("auth_token", data.access_token);
    console.log("Token saved!");
}
```

### Running Your Script

1. Press `Ctrl+S` to send the request
2. The pre-request script runs first
3. The HTTP request is sent
4. The post-response script runs
5. Check the **Tests** tab for assertion results
6. Check the **Console** tab for log output

## Script Types Comparison

| Feature | Pre-request | Post-response |
|---------|-------------|---------------|
| Modify URL | Yes | No |
| Modify Headers | Yes | No |
| Modify Body | Yes | No |
| Read Response | No | Yes |
| Run Tests | No | Yes |
| Set Variables | Yes | Yes |
| Console Output | Yes | Yes |

## Next Steps

- [API Overview](API-Overview) - Learn about all available objects
- [Authentication Examples](Examples-Auth) - Common auth patterns
- [Testing Examples](Examples-Testing) - Writing effective tests

---

*[Home](Home) | [API Overview](API-Overview)*
