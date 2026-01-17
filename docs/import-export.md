# Import & Export

LazyCurl supports importing from and exporting to multiple formats for interoperability with other tools.

## Overview

| Format | Import | Export | TUI Shortcut | CLI Command |
|--------|--------|--------|--------------|-------------|
| **cURL** | ✅ | ✅ | `Ctrl+I` / `Ctrl+E` | - |
| **OpenAPI 3.x** | ✅ | ❌ | `Ctrl+O` | `lazycurl import openapi` |
| **Postman** | ✅ | ✅ | `Ctrl+P` | `lazycurl import postman` |

---

## cURL Import/Export

### Import cURL Command (`Ctrl+I`)

Paste a cURL command to create a new request in LazyCurl.

**Supported Features:**

- HTTP methods (`-X`, `--request`)
- Headers (`-H`, `--header`)
- Request body (`-d`, `--data`, `--data-raw`)
- Basic authentication (`-u`, `--user`)
- Multiline commands (backslash `\` or backtick `` ` `` continuation)

**Variable Conversion:**

Shell variables are automatically converted to LazyCurl syntax:

| Shell Syntax | LazyCurl Syntax |
|--------------|-----------------|
| `$VAR` | `{{VAR}}` |
| `${VAR}` | `{{VAR}}` |

**Example:**

```bash
curl -X POST https://api.example.com/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "John", "email": "john@example.com"}'
```

Becomes a POST request with:

- URL: `https://api.example.com/users`
- Headers: `Content-Type`, `Authorization` (with `{{TOKEN}}`)
- Body: JSON object

### Export as cURL (`Ctrl+E`)

Copy the current request as a cURL command to clipboard.

**Output Format:**

```bash
curl -X POST 'https://api.example.com/users' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer your-token' \
  -d '{"name": "John"}'
```

**Included Elements:**

- HTTP method
- URL with resolved variables
- All headers
- Request body (properly escaped)
- Authentication headers

---

## OpenAPI Import

Import OpenAPI 3.x specifications to create collections automatically.

### Supported Versions

| Version | Support |
|---------|---------|
| OpenAPI 3.0.x | ✅ Full |
| OpenAPI 3.1.x | ✅ Full |
| Swagger 2.0 | ❌ (upgrade guidance provided) |

### TUI Import (`Ctrl+O`)

1. Press `Ctrl+O` to open import modal
2. Enter path to OpenAPI file (JSON or YAML)
3. Preview shows endpoints count and tags
4. Confirm to create collection

### CLI Import

```bash
# Basic import
lazycurl import openapi api.yaml

# Custom collection name
lazycurl import openapi spec.json --name "My API"

# Specify output location
lazycurl import openapi spec.yaml --output ./collections/api.json

# Preview without saving
lazycurl import openapi spec.yaml --dry-run

# JSON output for scripting
lazycurl import openapi spec.yaml --json
```

### Conversion Details

**Organization:**

- Operations grouped by tags into folders
- Untagged operations go to "Untagged" folder
- Operation summary becomes request name

**Parameters:**

| OpenAPI | LazyCurl |
|---------|----------|
| Path parameters | URL with `{param}` syntax |
| Query parameters | Request params with examples |
| Header parameters | Request headers |
| Request body | JSON body with schema examples |

**Authentication:**

Security schemes are automatically extracted:

| OpenAPI Security | LazyCurl Auth |
|------------------|---------------|
| `type: http, scheme: bearer` | Bearer Token |
| `type: http, scheme: basic` | Basic Auth |
| `type: apiKey, in: header` | API Key (Header) |
| `type: apiKey, in: query` | API Key (Query) |

**Example:**

```yaml
# OpenAPI spec
paths:
  /users:
    get:
      tags: [Users]
      summary: List all users
      security:
        - bearerAuth: []
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
            example: 10

components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
```

Creates:

- Folder: "Users"
- Request: "List all users"
- URL: `/users?limit=10`
- Auth: Bearer Token

### Handling Complex Schemas

**$ref Resolution:**

References are automatically resolved via libopenapi.

**Circular References:**

Handled with depth limit (max 5 levels) to prevent infinite loops.

**Example Generation:**

Schema examples are generated from:

1. `example` field (if present)
2. `default` field
3. Type-based placeholders (`"string"`, `0`, `true`, etc.)

---

## Postman Import/Export

Import Postman collections and environments, or export LazyCurl collections to Postman format.

### Supported Formats

| Format | Support |
|--------|---------|
| Postman Collection v2.1 | ✅ Full |
| Postman Collection v2.0 | ✅ Full |
| Postman Environment | ✅ Full |

### TUI Import (`Ctrl+P`)

1. Press `Ctrl+P` to open import modal
2. Enter path to Postman export file
3. File type auto-detected (collection vs environment)
4. Confirm to import

### CLI Import

```bash
# Import collection
lazycurl import postman collection.json

# Import environment (auto-detected)
lazycurl import postman environment.json

# Auto-detect format (postman or openapi)
lazycurl import collection.json

# Preview without saving
lazycurl import postman collection.json --dry-run

# JSON output for scripting
lazycurl import postman collection.json --json
```

### Collection Conversion

**Structure Mapping:**

| Postman | LazyCurl |
|---------|----------|
| Collection | Collection |
| Folder | Folder |
| Request | Request |
| Variables | Variables (same `{{var}}` syntax) |

**Request Conversion:**

| Postman Field | LazyCurl Field |
|---------------|----------------|
| `request.method` | `method` |
| `request.url` | `url` |
| `request.header` | `headers` |
| `request.body.raw` | `body` |
| `request.auth` | `auth` |

**Authentication Mapping:**

| Postman Auth | LazyCurl Auth |
|--------------|---------------|
| `bearer` | Bearer Token |
| `basic` | Basic Auth |
| `apikey` | API Key |

### Environment Conversion

Postman environments map directly to LazyCurl environments:

```json
// Postman Environment
{
  "name": "Development",
  "values": [
    {"key": "base_url", "value": "http://localhost:3000", "enabled": true},
    {"key": "token", "value": "secret", "enabled": true}
  ]
}
```

Becomes:

```json
// LazyCurl Environment
{
  "name": "Development",
  "variables": {
    "base_url": {"value": "http://localhost:3000", "active": true},
    "token": {"value": "secret", "active": true}
  }
}
```

### Export to Postman

Export LazyCurl collections to Postman format for sharing with team members using Postman.

**Via TUI:**

1. Select collection in Collections panel
2. Use export command (coming soon)

**Exported Format:**

- Postman Collection v2.1 format
- Compatible with Postman import

---

## Best Practices

### Importing Large Collections

For large OpenAPI specs or Postman collections:

1. Use `--dry-run` first to preview
2. Review the collection structure
3. Import to a dedicated workspace

### Variable Compatibility

- **cURL to LazyCurl**: Shell variables (`$VAR`) auto-convert to `{{VAR}}`
- **Postman to LazyCurl**: Variables use same syntax, direct mapping
- **OpenAPI to LazyCurl**: Server variables become environment variables

### Team Collaboration

1. **Share OpenAPI specs** - Source of truth for API definition
2. **Export to Postman** - For team members preferring Postman
3. **Commit collections to Git** - Version control your API tests

---

## Troubleshooting

### OpenAPI Import Errors

| Error | Solution |
|-------|----------|
| "Unsupported version: 2.0" | Upgrade to OpenAPI 3.x |
| "Failed to parse" | Validate YAML/JSON syntax |
| "Missing required field" | Check OpenAPI spec validity |

### Postman Import Errors

| Error | Solution |
|-------|----------|
| "Unrecognized format" | Ensure file is Postman export |
| "Version not supported" | Use Collection v2.0 or v2.1 |

### cURL Parse Errors

| Error | Solution |
|-------|----------|
| "Invalid URL" | Check URL format |
| "Missing method" | Add `-X METHOD` flag |

---

## See Also

- [CLI Reference](cli.md) - Full CLI documentation
- [Collections](collections.md) - Managing collections
- [Environments](environments.md) - Variable management
