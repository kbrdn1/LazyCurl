# Environments Guide

Complete guide to managing environments and variables in LazyCurl.

## Table of Contents

- [Overview](#overview)
- [Environment Structure](#environment-structure)
- [Variable System](#variable-system)
- [Managing Environments](#managing-environments)
- [Managing Variables](#managing-variables)
- [Variable Substitution](#variable-substitution)
- [System Variables](#system-variables)
- [File Format Reference](#file-format-reference)

---

## Overview

Environments allow you to define variables that can be used across your requests. This enables you to easily switch between different configurations (development, staging, production) without modifying your requests.

### Key Features

- **Multiple Environments**: Maintain separate configs for dev/staging/prod
- **Variable Substitution**: Use `{{variable}}` syntax anywhere
- **Secret Variables**: Hide sensitive values in the UI
- **Active/Inactive Toggle**: Enable or disable variables without deleting
- **System Variables**: Built-in dynamic values

---

## Environment Structure

### Directory Layout

```
.lazycurl/
└── environments/
    ├── development.json
    ├── staging.json
    └── production.json
```

### Visual Representation

In LazyCurl, environments appear as an expandable tree:

```
▼ Development ✓ (active)
  ○ base_url: http://localhost:3000
  ● token: ••••••••
  ○ user_id: 123
▷ Staging
▷ Production
```

**Legend:**

- `✓` = Active environment
- `●` = Secret variable (value hidden)
- `○` = Regular variable
- Dimmed text = Inactive variable

---

## Variable System

### Variable Types

| Type | Display | Description |
|------|---------|-------------|
| **Regular** | `○ name: value` | Normal visible variable |
| **Secret** | `● name: ••••••••` | Value hidden in UI |
| **Inactive** | Dimmed text | Defined but not used |

### Variable States

Each variable has three properties:

```json
{
  "value": "the-actual-value",
  "secret": false,
  "active": true
}
```

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `value` | string | `""` | The variable's value |
| `secret` | boolean | `false` | Hide value in UI |
| `active` | boolean | `true` | Use in substitution |

---

## Managing Environments

### Creating an Environment

1. Switch to Environments tab (press `2`)
2. Press `N` to create a new environment
3. Enter the environment name
4. Press `Enter` to confirm

### Selecting an Environment

1. Navigate to the environment
2. Press `S` or `Enter` to select
3. The environment shows `✓` and becomes active

**Only one environment can be active at a time.**

### Duplicating an Environment

1. Select the environment
2. Press `D` to duplicate
3. A copy is created with "_copy" suffix

### Deleting an Environment

1. Select the environment
2. Press `d` to delete
3. Confirm the deletion

### Renaming an Environment

1. Select the environment
2. Press `R` to rename
3. Enter the new name
4. Press `Enter` to confirm

---

## Managing Variables

### Creating a Variable

1. Expand an environment (press `l`)
2. Press `n` to create a new variable
3. Enter:
   - **Name**: Variable identifier (e.g., `base_url`)
   - **Value**: Variable value (e.g., `https://api.example.com`)
4. Press `Enter` to confirm

### Editing a Variable

1. Navigate to the variable with `j`/`k`
2. Press `c` or `i` to edit
3. Modify the value
4. Press `Enter` to save or `Esc` to cancel

### Duplicating a Variable

1. Select the variable
2. Press `D` to duplicate
3. A copy is created with "_copy" suffix

### Deleting a Variable

1. Select the variable
2. Press `d` to delete
3. Confirm the deletion

### Renaming a Variable

1. Select the variable
2. Press `R` to rename
3. Enter the new name
4. Press `Enter` to confirm

### Toggling Active State

| Key | Action |
|-----|--------|
| `a` | Toggle active/inactive |
| `A` | Toggle active/inactive (same) |

Inactive variables are not used in substitution.

### Toggling Secret State

| Key | Action |
|-----|--------|
| `s` | Toggle secret/visible |

Secret variables have their values hidden in the UI but are still used in requests.

---

## Variable Substitution

### Syntax

Variables use double curly braces:

```
{{variable_name}}
```

### Where Variables Work

Variables can be used in:

| Location | Example |
|----------|---------|
| URL | `{{base_url}}/api/users` |
| Headers | `Bearer {{token}}` |
| Body | `{"user": "{{user_name}}"}` |
| Query Parameters | `?id={{user_id}}` |

### Substitution Process

1. Select an environment
2. Create requests with `{{variable}}` placeholders
3. When sending, variables are replaced with actual values

**Example:**

Environment:

```json
{
  "base_url": { "value": "https://api.example.com", "active": true },
  "token": { "value": "abc123", "active": true }
}
```

Request URL:

```
{{base_url}}/users
```

Sent as:

```
https://api.example.com/users
```

### Inactive Variables

Inactive variables are **not** substituted:

```
{{inactive_var}} → {{inactive_var}} (unchanged)
```

---

## System Variables

LazyCurl provides built-in system variables that generate dynamic values.

### Available System Variables

| Variable | Description | Example Output |
|----------|-------------|----------------|
| `{{$timestamp}}` | Current Unix timestamp | `1699876543` |
| `{{$datetime}}` | Current RFC3339 datetime | `2024-11-13T10:15:30Z` |
| `{{$date}}` | Current date (YYYY-MM-DD) | `2024-11-13` |
| `{{$time}}` | Current time (HH:MM:SS) | `10:15:30` |
| `{{$uuid}}` | Random UUID v4 | `a1b2c3d4-e5f6-...` |
| `{{$randomInt}}` | Random integer (0-999999) | `427891` |
| `{{$random}}` | Random 10-char string | `aB3cD7eF9g` |

### Usage Examples

**Request ID header:**

```
X-Request-ID: {{$uuid}}
```

**Timestamp in body:**

```json
{
  "created_at": "{{$datetime}}",
  "request_id": "{{$uuid}}"
}
```

**Random test data:**

```json
{
  "email": "user_{{$randomInt}}@test.com",
  "code": "{{$random}}"
}
```

---

## File Format Reference

### Complete Environment Schema

```json
{
  "name": "Environment Name",
  "description": "Optional description",
  "variables": {
    "variable_name": {
      "value": "variable_value",
      "secret": false,
      "active": true
    }
  }
}
```

### Field Descriptions

#### EnvironmentFile

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Environment display name |
| `description` | string | No | Environment description |
| `variables` | object | Yes | Map of variable names to configs |

#### EnvironmentVariable

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `value` | string | Yes | - | The variable's value |
| `secret` | boolean | No | `false` | Hide value in UI |
| `active` | boolean | No | `true` | Use in substitution |

---

## Examples

### Development Environment

```json
{
  "name": "Development",
  "description": "Local development environment",
  "variables": {
    "base_url": {
      "value": "http://localhost:3000",
      "active": true
    },
    "api_version": {
      "value": "v1",
      "active": true
    },
    "token": {
      "value": "dev-token-12345",
      "secret": true,
      "active": true
    },
    "debug": {
      "value": "true",
      "active": true
    }
  }
}
```

### Staging Environment

```json
{
  "name": "Staging",
  "description": "Staging server for testing",
  "variables": {
    "base_url": {
      "value": "https://staging.api.example.com",
      "active": true
    },
    "api_version": {
      "value": "v1",
      "active": true
    },
    "token": {
      "value": "staging-token-67890",
      "secret": true,
      "active": true
    },
    "debug": {
      "value": "false",
      "active": true
    }
  }
}
```

### Production Environment

```json
{
  "name": "Production",
  "description": "Production API - Use with caution!",
  "variables": {
    "base_url": {
      "value": "https://api.example.com",
      "active": true
    },
    "api_version": {
      "value": "v1",
      "active": true
    },
    "token": {
      "value": "prod-secret-token",
      "secret": true,
      "active": true
    },
    "debug": {
      "value": "false",
      "active": false
    },
    "admin_override": {
      "value": "disabled",
      "active": false
    }
  }
}
```

---

## Keybindings Reference

### Environment-Level Actions

| Key | Action |
|-----|--------|
| `N` | Create new environment |
| `S` / `Enter` | Select/activate environment |
| `d` | Delete environment |
| `D` | Duplicate environment |
| `R` | Rename environment |
| `h` | Collapse environment |
| `l` | Expand environment |

### Variable-Level Actions

| Key | Action |
|-----|--------|
| `n` | Create new variable |
| `c` / `i` | Edit variable value |
| `d` | Delete variable |
| `D` | Duplicate variable |
| `R` | Rename variable |
| `a` / `A` | Toggle active/inactive |
| `s` | Toggle secret/visible |

### Navigation

| Key | Action |
|-----|--------|
| `j` / `k` | Move up/down |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `/` | Search environments/variables |

---

## Best Practices

### 1. Use Consistent Naming

```
✅ base_url, api_token, user_id
❌ baseUrl, BaseURL, base-url (inconsistent)
```

### 2. Mark Sensitive Values as Secret

Always mark these as secret:

- API tokens and keys
- Passwords
- Private URLs
- Personal identifiable information

### 3. Use Descriptive Environment Names

```
✅ Development, Staging, Production
✅ Local, QA, UAT, Prod
❌ Env1, Test, New
```

### 4. Keep Environments in Sync

Maintain the same variable names across environments:

```
Development: base_url = http://localhost:3000
Staging:     base_url = https://staging.api.com
Production:  base_url = https://api.com
```

### 5. Version Control Carefully

```bash
# .gitignore - Protect secrets
.lazycurl/environments/production.json
.lazycurl/environments/*-local.json

# Keep templates
!.lazycurl/environments/example.json
```

### 6. Use System Variables for Dynamic Data

Instead of hardcoding:

```json
{"timestamp": "1699876543"}
```

Use:

```json
{"timestamp": "{{$timestamp}}"}
```

### 7. Deactivate Instead of Delete

Use inactive state for:

- Temporarily disabling features
- Keeping reference values
- A/B testing configurations

---

## Troubleshooting

### Variable Not Replaced

1. **Check environment is selected**: Look for `✓` indicator
2. **Check variable is active**: Inactive variables are dimmed
3. **Check syntax**: Use `{{name}}` not `{name}` or `{{ name }}`
4. **Check spelling**: Variable names are case-sensitive

### Secret Value Shown

Secret values are only hidden in the UI. They are:

- Stored in plain text in JSON files
- Sent in requests as normal values
- Visible if you open the JSON file

### Environment Not Loading

1. Check file is valid JSON: `cat file.json | jq .`
2. Check file permissions
3. Check file is in `.lazycurl/environments/`
4. Restart LazyCurl

### Legacy Format Migration

LazyCurl supports legacy format (simple string values):

```json
// Legacy format (still works)
{
  "variables": {
    "key": "value"
  }
}

// New format (recommended)
{
  "variables": {
    "key": {
      "value": "value",
      "secret": false,
      "active": true
    }
  }
}
```

Legacy format is automatically converted when saving.
