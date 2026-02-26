# Relation Tuples API

## Overview

This API manages **Zanzibar-style relation tuples** for fine-grained access control. Relation tuples represent relationships between objects and subjects in the format:

```
<object>#<relation>@<subject>
```

**Important:** This is separate from RBAC permissions. When you implement RBAC in the future, you'll have:
- **Relation Tuples** (this API) - For Zanzibar-style authorization checks
- **RBAC Permissions** (future) - For role-based permission management

## Why "Relations" instead of "Permissions"?

The naming uses "relation" terminology to:
1. **Avoid confusion** with future RBAC permission system
2. **Accurately describe** what it does - manages relation tuples
3. **Follow Zanzibar** terminology (Google's authorization system)
4. **Be explicit** about the underlying model

## API Endpoints

All endpoints require JWT authentication and are prefixed with `/api/v1/relations`.

### 1. Grant Relation

**POST** `/api/v1/relations/grant`

Grants a relation tuple (e.g., user:alice can view document:readme).

**Request Body:**
```json
{
  "namespace": "document",
  "objectId": "readme",
  "relation": "viewer",
  "subjectNamespace": "user",
  "subjectObjectId": "alice",
  "subjectRelation": "",
  "expiresAt": "2026-12-31T23:59:59Z"
}
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "uuid",
    "namespace": "document",
    "objectId": "readme",
    "relation": "viewer",
    "subjectNamespace": "user",
    "subjectObjectId": "alice",
    "isActive": true,
    "createdAt": "2026-02-25T10:00:00Z",
    "updatedAt": "2026-02-25T10:00:00Z"
  }
}
```

### 2. Revoke Relation

**POST** `/api/v1/relations/revoke`

Revokes a relation tuple.

**Request Body:**
```json
{
  "namespace": "document",
  "objectId": "readme",
  "relation": "viewer",
  "subjectNamespace": "user",
  "subjectObjectId": "alice"
}
```

### 3. Check Relation

**POST** `/api/v1/relations/check`

Checks if a relation exists (authorization check).

**Request Body:**
```json
{
  "namespace": "document",
  "objectId": "readme",
  "relation": "viewer",
  "subjectNamespace": "user",
  "subjectObjectId": "alice"
}
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "allowed": true,
    "reason": ""
  }
}
```

### 4. List Relations

**GET** `/api/v1/relations/list`

Lists relation tuples with optional filters.

**Query Parameters:**
- `namespace` - Filter by object namespace
- `objectId` - Filter by object ID
- `relation` - Filter by relation type
- `subjectNamespace` - Filter by subject namespace
- `subjectObjectId` - Filter by subject object ID
- `page` - Page number (default: 1)
- `pageSize` - Items per page (default: 10, max: 100)

### 5. Expand Relation

**POST** `/api/v1/relations/expand`

Gets all subjects with a specific relation on an object.

**Request Body:**
```json
{
  "namespace": "document",
  "objectId": "readme",
  "relation": "viewer"
}
```

**Response:**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "subjects": [
      {
        "namespace": "user",
        "objectId": "alice"
      },
      {
        "namespace": "user",
        "objectId": "bob"
      }
    ],
    "count": 2
  }
}
```

### 6. Bulk Grant Relations

**POST** `/api/v1/relations/bulk-grant`

Grants multiple relation tuples in one request.

**Request Body:**
```json
{
  "relations": [
    {
      "namespace": "document",
      "objectId": "doc1",
      "relation": "viewer",
      "subjectNamespace": "user",
      "subjectObjectId": "alice"
    },
    {
      "namespace": "document",
      "objectId": "doc2",
      "relation": "viewer",
      "subjectNamespace": "user",
      "subjectObjectId": "alice"
    }
  ]
}
```

### 7. Bulk Revoke Relations

**POST** `/api/v1/relations/bulk-revoke`

Revokes multiple relation tuples.

### 8. Cleanup Expired Relations

**DELETE** `/api/v1/relations/cleanup`

Removes expired relation tuples (maintenance endpoint).

## Common Use Cases

### Document Access Control

```bash
# Grant Alice view access
curl -X POST http://localhost:8080/api/v1/relations/grant \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "document",
    "objectId": "doc-123",
    "relation": "viewer",
    "subjectNamespace": "user",
    "subjectObjectId": "alice-uuid"
  }'

# Check if Alice can view
curl -X POST http://localhost:8080/api/v1/relations/check \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "document",
    "objectId": "doc-123",
    "relation": "viewer",
    "subjectNamespace": "user",
    "subjectObjectId": "alice-uuid"
  }'
```

### Team-based Access (Usersets)

```bash
# Add user to team
curl -X POST http://localhost:8080/api/v1/relations/grant \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "team",
    "objectId": "engineering",
    "relation": "member",
    "subjectNamespace": "user",
    "subjectObjectId": "bob-uuid"
  }'

# Grant team access to project
curl -X POST http://localhost:8080/api/v1/relations/grant \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "project",
    "objectId": "proj-001",
    "relation": "contributor",
    "subjectNamespace": "team",
    "subjectObjectId": "engineering",
    "subjectRelation": "member"
  }'
```

## Integration in Your Code

### Check Relation Before Action

```go
func (s *DocumentService) UpdateDocument(ctx context.Context, docID string, req UpdateDocReq) error {
    userID := ctx.Value("user_id").(string)
    
    // Check if user has editor relation
    resp, err := s.relationSvc.CheckRelation(ctx, dto.CheckRelationReq{
        Namespace:        "document",
        ObjectID:         docID,
        Relation:         "editor",
        SubjectNamespace: "user",
        SubjectObjectID:  userID,
    })
    
    if err != nil || !resp.Allowed {
        return errorx.New(errorx.ErrPermissionDenied, "You cannot edit this document")
    }
    
    return s.docRepo.Update(ctx, docID, req)
}
```

### Grant Relation on Resource Creation

```go
func (s *DocumentService) CreateDocument(ctx context.Context, req CreateDocReq) (*Document, error) {
    doc, err := s.docRepo.Create(ctx, &Document{
        Title:   req.Title,
        Content: req.Content,
    })
    if err != nil {
        return nil, err
    }
    
    // Grant owner relation to creator
    _, err = s.relationSvc.GrantRelation(ctx, dto.GrantRelationReq{
        Namespace:        "document",
        ObjectID:         doc.ID,
        Relation:         "owner",
        SubjectNamespace: "user",
        SubjectObjectID:  req.CreatorID,
    })
    if err != nil {
        s.docRepo.Delete(ctx, doc.ID)
        return nil, err
    }
    
    return doc, nil
}
```

## Future RBAC Integration

When you implement RBAC, you'll have two systems working together:

### Relation Tuples (Current System)
- **Purpose:** Fine-grained authorization checks
- **Use for:** Document access, project membership, folder permissions
- **Example:** "Does user:alice have viewer relation on document:readme?"

### RBAC Permissions (Future System)
- **Purpose:** Role and permission management
- **Use for:** Role definitions, permission assignments, role hierarchies
- **Example:** "Does role:admin have permission:users.delete?"

### How They Work Together

```go
// RBAC: Check if user's role has permission
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
    // Check RBAC permission
    hasPermission := s.rbacSvc.HasPermission(ctx, currentUser.RoleID, "users.delete")
    if !hasPermission {
        return errorx.New(errorx.ErrForbidden, "Your role doesn't have this permission")
    }
    
    // Relation: Check if user has specific relation to this resource
    canDelete, _ := s.relationSvc.CheckRelation(ctx, dto.CheckRelationReq{
        Namespace:        "user",
        ObjectID:         userID,
        Relation:         "admin",
        SubjectNamespace: "user",
        SubjectObjectID:  currentUser.ID,
    })
    
    if !canDelete {
        return errorx.New(errorx.ErrForbidden, "You don't have admin relation to this user")
    }
    
    return s.userRepo.Delete(ctx, userID)
}
```

## Summary

- **Service:** `IRelationSvc` / `RelationSvc`
- **Handler:** `RelationHandler`
- **DTOs:** `GrantRelationReq`, `CheckRelationReq`, `RelationTupleResp`, etc.
- **Endpoints:** `/api/v1/relations/*`
- **Purpose:** Manage Zanzibar-style relation tuples for authorization
- **Future:** Will coexist with RBAC permission system without naming conflicts

This naming clearly separates concerns and makes the codebase ready for future RBAC implementation! 🎯
