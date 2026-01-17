# CLI Reference

LazyCurl provides command-line interface for automation and scripting.

## Synopsis

```bash
lazycurl [command] [options]
```

## Commands

### Interactive Mode (Default)

```bash
lazycurl
```

Launches the TUI application in the current directory.

**Options:**

| Flag | Description |
|------|-------------|
| `--help`, `-h` | Show help |
| `--version`, `-v` | Show version |

### Import Command

Import collections from external formats.

```bash
lazycurl import <format> <file> [options]
```

#### Import OpenAPI

```bash
lazycurl import openapi <file> [options]
```

**Arguments:**

| Argument | Description |
|----------|-------------|
| `file` | Path to OpenAPI spec (JSON or YAML) |

**Options:**

| Flag | Description |
|------|-------------|
| `--name`, `-n` | Custom collection name |
| `--output`, `-o` | Output file path |
| `--dry-run` | Preview without saving |
| `--json` | Output as JSON (for scripting) |

**Examples:**

```bash
# Basic import
lazycurl import openapi api.yaml

# Custom name
lazycurl import openapi spec.json --name "My API"

# Preview only
lazycurl import openapi spec.yaml --dry-run

# JSON output for scripting
lazycurl import openapi spec.yaml --json
```

**Output (normal):**

```
Imported OpenAPI specification:
  Name: My API
  Endpoints: 15
  Tags: Users, Products, Orders
  Saved to: .lazycurl/collections/my-api.json
```

**Output (JSON):**

```json
{
  "success": true,
  "collection_name": "My API",
  "endpoints_count": 15,
  "tags": ["Users", "Products", "Orders"],
  "output_path": ".lazycurl/collections/my-api.json"
}
```

#### Import Postman

```bash
lazycurl import postman <file> [options]
```

**Arguments:**

| Argument | Description |
|----------|-------------|
| `file` | Path to Postman export (JSON) |

**Options:**

| Flag | Description |
|------|-------------|
| `--name`, `-n` | Custom collection name |
| `--output`, `-o` | Output file path |
| `--dry-run` | Preview without saving |
| `--json` | Output as JSON (for scripting) |

**Examples:**

```bash
# Import collection
lazycurl import postman collection.json

# Import environment (auto-detected)
lazycurl import postman environment.json

# Preview import
lazycurl import postman collection.json --dry-run
```

**Output:**

```
Imported Postman collection:
  Name: My Collection
  Requests: 25
  Folders: 5
  Saved to: .lazycurl/collections/my-collection.json
```

Or for environments:

```
Imported Postman environment:
  Name: Development
  Variables: 8
  Saved to: .lazycurl/environments/development.json
```

#### Auto-Detection

```bash
lazycurl import <file> [options]
```

Without specifying format, LazyCurl auto-detects:

- OpenAPI specs (by `openapi` or `swagger` field)
- Postman collections (by `info._postman_id` field)
- Postman environments (by `_postman_variable_scope` field)

```bash
# Auto-detect format
lazycurl import api.yaml
lazycurl import collection.json
```

## Exit Codes

| Code | Description |
|------|-------------|
| `0` | Success |
| `1` | General error |
| `2` | Invalid arguments |
| `3` | File not found |
| `4` | Parse error |
| `5` | Validation error |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `LAZYCURL_CONFIG` | Custom global config path |
| `VISUAL` | Preferred external editor |
| `EDITOR` | Fallback external editor |
| `NO_COLOR` | Disable colored output |

## Scripting Examples

### CI/CD Integration

```bash
#!/bin/bash
# Import OpenAPI spec in CI

lazycurl import openapi api.yaml --json > import-result.json

if [ $? -eq 0 ]; then
  echo "Import successful"
  cat import-result.json | jq '.endpoints_count'
else
  echo "Import failed"
  exit 1
fi
```

### Batch Import

```bash
#!/bin/bash
# Import all OpenAPI specs in a directory

for spec in specs/*.yaml; do
  name=$(basename "$spec" .yaml)
  lazycurl import openapi "$spec" --name "$name" --json
done
```

### Format Validation

```bash
#!/bin/bash
# Validate OpenAPI spec without saving

if lazycurl import openapi api.yaml --dry-run > /dev/null 2>&1; then
  echo "Valid OpenAPI spec"
else
  echo "Invalid OpenAPI spec"
fi
```

## Future Commands

The following commands are planned for future releases:

### Run Collection (Planned)

```bash
lazycurl run <collection> [options]
```

Run all requests in a collection sequentially.

### Export (Planned)

```bash
lazycurl export <format> <collection> [options]
```

Export collection to external format (Postman, OpenAPI).

### Workspace Commands (Planned)

```bash
lazycurl workspace list
lazycurl workspace create <name>
lazycurl workspace switch <name>
```

Manage workspaces from CLI.

---

## See Also

- [Import/Export](import-export.md) - Detailed import/export guide
- [Configuration](configuration.md) - Configuration options
- [Getting Started](getting-started.md) - First steps with LazyCurl
