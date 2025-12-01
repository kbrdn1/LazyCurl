# Collections Guide

Complete guide to managing request collections in LazyCurl.

## Table of Contents

- [Overview](#overview)
- [Collection Structure](#collection-structure)
- [Creating Collections](#creating-collections)
- [Managing Requests](#managing-requests)
- [Organizing with Folders](#organizing-with-folders)
- [Request Configuration](#request-configuration)
- [Collection Operations](#collection-operations)
- [File Format Reference](#file-format-reference)

---

## Overview

Collections are the primary way to organize your API requests in LazyCurl. They are stored as JSON files in your workspace's `.lazycurl/collections/` directory.

### Key Features

- **Hierarchical Organization**: Nest requests in folders
- **File-Based Storage**: Version control friendly JSON files
- **Variable Support**: Use `{{variable}}` syntax in URLs, headers, and body
- **Portable**: Share collections via git or file transfer

---

## Collection Structure

### Directory Layout

```
.lazycurl/
└── collections/
    ├── api.json              # Main API collection
    ├── authentication.json   # Auth-related requests
    └── admin.json            # Admin endpoints
```

### Visual Representation

In LazyCurl, collections appear as a tree:

```
▼ My API
  ▼ Users
    GET  /users
    POST /users
    GET  /users/:id
  ▼ Products
    GET  /products
    POST /products
  GET /health
```

---

## Creating Collections

### From the UI

1. Navigate to the Collections panel (press `h` until selected)
2. Press `N` to create a new folder (collection root)
3. Enter the collection name
4. The collection file is automatically created

### Manually

Create a JSON file in `.lazycurl/collections/`:

```json
{
  "name": "My API",
  "description": "REST API for my application",
  "folders": [],
  "requests": []
}
```

---

## Managing Requests

### Creating a Request

1. Select a collection or folder in the tree
2. Press `n` to create a new request
3. Fill in the dialog:
   - **Name**: Descriptive request name
   - **Method**: HTTP method (GET, POST, etc.)
   - **URL**: Endpoint URL (supports variables)
4. Press `Enter` to confirm

### Editing a Request

1. Navigate to the request with `j`/`k`
2. Press `c` or `i` to edit
3. Modify the request in the Request panel
4. Changes are auto-saved

### Duplicating a Request

1. Select the request
2. Press `D` to duplicate
3. A copy is created with "_copy" suffix

### Deleting a Request

1. Select the request
2. Press `d` to delete
3. Confirm the deletion

### Renaming a Request

1. Select the request
2. Press `R` to rename
3. Enter the new name
4. Press `Enter` to confirm

---

## Organizing with Folders

### Creating Folders

1. Select parent location (collection or folder)
2. Press `N` to create a new folder
3. Enter folder name
4. Press `Enter`

### Folder Navigation

| Key | Action |
|-----|--------|
| `j` / `k` | Move up/down in tree |
| `h` | Collapse folder / Go to parent |
| `l` | Expand folder / Enter folder |
| `Enter` | Open selected request |
| `Space` | Toggle folder expand/collapse |

### Nested Folders

Folders can be nested to any depth:

```
▼ My API
  ▼ v1
    ▼ Users
      ▼ Admin
        GET /v1/admin/users
      GET /v1/users
    ▼ Products
      GET /v1/products
  ▼ v2
    ▼ Users
      GET /v2/users
```

---

## Request Configuration

### Supported HTTP Methods

| Method | Color | Description |
|--------|-------|-------------|
| **GET** | Green | Retrieve resource |
| **POST** | Peach | Create resource |
| **PUT** | Blue | Replace resource |
| **PATCH** | Mauve | Partial update |
| **DELETE** | Red | Remove resource |
| **HEAD** | Green | Headers only |
| **OPTIONS** | Yellow | Check capabilities |

### URL Configuration

URLs support variable substitution:

```
{{base_url}}/api/{{api_version}}/users/{{user_id}}
```

**Examples:**
- `https://api.example.com/users`
- `{{base_url}}/users`
- `{{base_url}}/users/{{user_id}}`

### Headers Configuration

Headers are key-value pairs:

```json
{
  "headers": {
    "Authorization": "Bearer {{token}}",
    "Content-Type": "application/json",
    "Accept": "application/json",
    "X-Custom-Header": "{{custom_value}}"
  }
}
```

### Body Configuration

Request body supports multiple formats:

#### JSON Body
```json
{
  "body": {
    "name": "{{user_name}}",
    "email": "{{user_email}}",
    "role": "user"
  }
}
```

#### String Body
```json
{
  "body": "raw text content"
}
```

#### Form Data
```json
{
  "body": {
    "field1": "value1",
    "field2": "value2"
  }
}
```

---

## Collection Operations

### Keybindings Reference

| Key | Action | Context |
|-----|--------|---------|
| `n` | New request | Any |
| `N` | New folder | Any |
| `c` / `i` | Edit request | On request |
| `R` | Rename | On any item |
| `d` | Delete | On any item |
| `D` | Duplicate | On any item |
| `y` | Yank (copy) | On any item |
| `p` | Paste | Any |
| `/` | Search | Any |
| `Enter` | Open request | On request |
| `Space` | Toggle expand | On folder |

### Search

1. Press `/` to open search
2. Type your search query
3. Matching items are highlighted
4. Press `n` for next match, `N` for previous
5. Press `Enter` to open selected
6. Press `Esc` to clear search

### Clipboard Operations

Copy and paste requests between collections:

1. Select a request
2. Press `y` to yank (copy)
3. Navigate to destination
4. Press `p` to paste

---

## File Format Reference

### Complete Collection Schema

```json
{
  "name": "Collection Name",
  "description": "Optional description",
  "folders": [
    {
      "name": "Folder Name",
      "description": "Optional folder description",
      "folders": [],
      "requests": [
        {
          "id": "req_unique_id",
          "name": "Request Name",
          "description": "Optional request description",
          "method": "GET",
          "url": "{{base_url}}/endpoint",
          "headers": {
            "Header-Name": "Header-Value"
          },
          "body": null,
          "tests": [
            {
              "name": "Status is 200",
              "assert": "response.status === 200"
            }
          ]
        }
      ]
    }
  ],
  "requests": []
}
```

### Field Descriptions

#### CollectionFile

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Collection display name |
| `description` | string | No | Collection description |
| `folders` | Folder[] | No | Nested folders |
| `requests` | Request[] | No | Root-level requests |

#### Folder

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Folder display name |
| `description` | string | No | Folder description |
| `folders` | Folder[] | No | Nested subfolders |
| `requests` | Request[] | No | Folder's requests |

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier (e.g., `req_abc123`) |
| `name` | string | Yes | Request display name |
| `description` | string | No | Request description |
| `method` | string | Yes | HTTP method (GET, POST, etc.) |
| `url` | string | Yes | Request URL (supports variables) |
| `headers` | object | No | Key-value header pairs |
| `body` | any | No | Request body (JSON, string, or null) |
| `tests` | Test[] | No | Test assertions |

#### Test

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Test name |
| `assert` | string | Yes | JavaScript assertion expression |

---

## Examples

### REST API Collection

```json
{
  "name": "User API",
  "description": "User management endpoints",
  "folders": [
    {
      "name": "Authentication",
      "requests": [
        {
          "id": "req_login",
          "name": "Login",
          "method": "POST",
          "url": "{{base_url}}/auth/login",
          "headers": {
            "Content-Type": "application/json"
          },
          "body": {
            "email": "{{user_email}}",
            "password": "{{user_password}}"
          }
        },
        {
          "id": "req_logout",
          "name": "Logout",
          "method": "POST",
          "url": "{{base_url}}/auth/logout",
          "headers": {
            "Authorization": "Bearer {{token}}"
          }
        }
      ]
    },
    {
      "name": "Users",
      "requests": [
        {
          "id": "req_list_users",
          "name": "List Users",
          "method": "GET",
          "url": "{{base_url}}/users",
          "headers": {
            "Authorization": "Bearer {{token}}"
          }
        },
        {
          "id": "req_create_user",
          "name": "Create User",
          "method": "POST",
          "url": "{{base_url}}/users",
          "headers": {
            "Authorization": "Bearer {{token}}",
            "Content-Type": "application/json"
          },
          "body": {
            "name": "{{new_user_name}}",
            "email": "{{new_user_email}}",
            "role": "user"
          }
        },
        {
          "id": "req_get_user",
          "name": "Get User",
          "method": "GET",
          "url": "{{base_url}}/users/{{user_id}}",
          "headers": {
            "Authorization": "Bearer {{token}}"
          }
        },
        {
          "id": "req_delete_user",
          "name": "Delete User",
          "method": "DELETE",
          "url": "{{base_url}}/users/{{user_id}}",
          "headers": {
            "Authorization": "Bearer {{token}}"
          }
        }
      ]
    }
  ]
}
```

### GraphQL Collection

```json
{
  "name": "GraphQL API",
  "folders": [
    {
      "name": "Queries",
      "requests": [
        {
          "id": "req_get_users",
          "name": "Get Users",
          "method": "POST",
          "url": "{{graphql_url}}",
          "headers": {
            "Content-Type": "application/json",
            "Authorization": "Bearer {{token}}"
          },
          "body": {
            "query": "query { users { id name email } }"
          }
        }
      ]
    },
    {
      "name": "Mutations",
      "requests": [
        {
          "id": "req_create_user",
          "name": "Create User",
          "method": "POST",
          "url": "{{graphql_url}}",
          "headers": {
            "Content-Type": "application/json",
            "Authorization": "Bearer {{token}}"
          },
          "body": {
            "query": "mutation CreateUser($input: UserInput!) { createUser(input: $input) { id name } }",
            "variables": {
              "input": {
                "name": "{{user_name}}",
                "email": "{{user_email}}"
              }
            }
          }
        }
      ]
    }
  ]
}
```

---

## Best Practices

### 1. Use Descriptive Names

```
✅ "Get User by ID"
✅ "Create New Product"
❌ "Request 1"
❌ "Test"
```

### 2. Organize by Domain

```
▼ User Service
  ▼ Users
  ▼ Roles
  ▼ Permissions
▼ Product Service
  ▼ Products
  ▼ Categories
```

### 3. Use Variables

Always use variables for:
- Base URLs (`{{base_url}}`)
- Authentication tokens (`{{token}}`)
- Dynamic IDs (`{{user_id}}`)

### 4. Document with Descriptions

Add descriptions to help team members:

```json
{
  "name": "Delete User",
  "description": "Permanently deletes a user. Requires admin role.",
  "method": "DELETE"
}
```

### 5. Version Control

Commit collections to git for:
- History tracking
- Team collaboration
- Backup and restore
