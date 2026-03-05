# dreon-auth

A Go-based **Auth Service** for authentication, sessions, token management, **RBAC** (roles & permissions), and **Zanzibar-style relation tuples** for relationship-based authorization.

## 🎯 Overview

dreon-auth provides user management, JWT (RS256) issue/verify, session handling, **Google OAuth2 login**, **role-based access control** (roles + permissions, project-scoped), and **relation tuples** in the style of [Google Zanzibar](https://research.google/pubs/pub48190/) for fine-grained, relationship-based access checks.

## 🏗️ Architecture

### Tech Stack

- **Go 1.25** – Backend auth service
- **PostgreSQL** – Users, sessions, projects, roles, relation tuples
- **Redis** – Session store, cache, OAuth state
- **JWT (RS256)** – Asymmetric token signing and verification
- **OAuth2 (Google)** – Sign-in with Google
- **gRPC** – Internal API for relation tuples and permission checks (AuthInternalService)
- **Docker** – Containerization and orchestration

### High-Level Design

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────────────────────────┐
│                  dreon-auth (API)                        │
│                   Go + Fx + Echo                         │
│  • Auth: login, register, refresh, logout, Google OAuth  │
│  • Users, Projects, Roles, Permissions                   │
│  • Relation tuples: grant, revoke, check, expand         │
│  • gRPC: GrantRelationTuple, CheckPermission, GetUserPermissions │
└─────────────────────────────────────────────────────────┘
       │              │
       ▼              ▼
┌──────────────┐   ┌─────────────┐
│  PostgreSQL   │   │    Redis    │
│  Users        │   │  Sessions   │
│  Sessions     │   │  OAuth state│
│  Projects     │   │  Cache      │
│  Roles        │   └─────────────┘
│  Relation     │
│  tuples       │
└──────────────┘
```

## 🚀 Features

- ✅ **Auth** – Email/password login & register, JWT access/refresh, logout
- ✅ **Google OAuth2** – Sign-in with Google (redirect flow, session-from-state)
- ✅ **JWT RS256** – Asymmetric keys, configurable via env
- ✅ **Sessions** – Session model and storage (PostgreSQL + Redis)
- ✅ **Users** – User CRUD, multi-auth (email, Google; extensible to Facebook, Apple)
- ✅ **Projects** – Project CRUD (multi-tenant scope)
- ✅ **RBAC** – Roles with permissions, system roles (`admin`, `editor`, `user`), project roles, assign/remove roles to users
- ✅ **Permissions** – Registry from config file (`PERMISSIONS_FILE`), list permissions, user permission checks
- ✅ **Relation tuples (Zanzibar-style)** – Grant/revoke/check/expand relations (`object#relation@subject`), bulk grant/revoke, optional expiry
- ✅ **REST API** – Echo, validation, error handling
- ✅ **gRPC API** – Internal `AuthInternalService` for relation tuples and user permissions (server + generated client)
- ✅ **Docker** – docker-compose for local dev

## 📡 API Overview

Base URL: `http://localhost:8080/api/v1`

| Area        | Path           | Description |
|------------|----------------|-------------|
| **Auth**   | `/auth`        | Login, register, refresh-token, logout, Google OAuth callback, session-from-state, session (JWT) |
| **Users**  | `/users`      | List, get, create, update, delete users |
| **Projects** | `/projects` | List, get, create, update, delete projects (super-admin) |
| **Roles**  | `/roles`      | CRUD roles, assign/remove role to user, get user permissions |
| **Permissions** | `/permissions` | List permission registry |
| **Relations** | `/relations` | Grant, revoke, check, list, expand, bulk-grant, bulk-revoke (JWT required) |

**gRPC** (internal): default `localhost:9090` — see [Using gRPC](#-using-grpc) below.

### Auth Endpoints (no JWT unless noted)

- `POST /auth/login` – Login (email or `authType: "GOOGLE"` with `redirectUrl` for OAuth start)
- `POST /auth/register` – Register with email/password
- `POST /auth/refresh-token` – Exchange refresh token for new tokens
- `POST /auth/logout` – Invalidate refresh token
- `GET /auth/google/callback` – Google OAuth callback (redirect; exchanges code, stores user, redirects to frontend with `?refreshState=...`)
- `POST /auth/session-from-state` – Exchange `refreshState` for session tokens (after Google OAuth or other providers)
- `GET /auth/session` – Get current session (requires JWT)

## 📦 Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.25+ (for local development)
- Make (optional)

### Quick Start with Docker

```bash
git clone https://github.com/hiamthach108/dreon-auth.git
cd dreon-auth
docker-compose up -d
```

- API: http://localhost:8080  
- gRPC: localhost:9090 (set `GRPC_PORT` in env if needed)  
- PostgreSQL: localhost:5432  
- Redis: localhost:6379  

### Local Development

1. **Dependencies**
   ```bash
   go mod download
   ```

2. **Environment**
   ```bash
   cp .env.example .env
   # Edit .env (DB, Redis, JWT, Google OAuth if needed)
   ```

3. **JWT keys (RS256)**  
   Generate and configure as in the [JWT setup](#jwt-setup) section below.

4. **Start DB & Redis**
   ```bash
   docker-compose up -d postgres redis
   ```

5. **Run app**
   ```bash
   go run cmd/main.go
   # or: make run
   ```

### JWT Setup

The app uses **RS256** JWT. Generate keys and set them in `.env`.

```bash
mkdir -p keys
openssl genrsa -out keys/private.pem 2048
openssl rsa -pubout -in keys/private.pem -out keys/public.pem
```

In `.env` use either **file paths** or **inline PEM**:

- **File paths (recommended for local):**
  ```env
  JWT_PRIVATE_KEY=keys/private.pem
  JWT_PUBLIC_KEY=keys/public.pem
  JWT_ACCESS_TOKEN_EXPIRES_IN=3600
  JWT_REFRESH_TOKEN_EXPIRES_IN=86400
  ```

- **Inline PEM:** set `JWT_PRIVATE_KEY` and `JWT_PUBLIC_KEY` to the full PEM content (e.g. for Docker/CI).

**Google OAuth (optional):** set `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, and in Google Cloud Console set redirect URI to `http://<HTTP_HOST>:<HTTP_PORT>/api/v1/auth/google/callback`.

---

## 🔐 Setting Up RBAC

RBAC is built around **projects**, **roles**, and **permissions**. Roles can be **system** (projectId = `"system"`) or **project-scoped**. Permissions are defined in a config file and attached to roles; users get permissions by being assigned roles (system or per project).

### 1. Permissions config

Define permissions in a JSON file (e.g. `config/permissions.json`) and set `PERMISSIONS_FILE` in `.env`:

```json
[
  {"name": "User View", "code": "users.view"},
  {"name": "User Create", "code": "users.create"},
  {"name": "User Update", "code": "users.update"},
  {"name": "User Delete", "code": "users.delete"},
  {"name": "Role View", "code": "roles.view"},
  {"name": "Role Create", "code": "roles.create"},
  {"name": "Role Update", "code": "roles.update"},
  {"name": "Role Delete", "code": "roles.delete"},
  {"name": "Role Assign", "code": "roles.assign"},
  {"name": "Role Revoke", "code": "roles.revoke"},
  {"name": "Project View", "code": "projects.view"},
  {"name": "Project Create", "code": "projects.create"},
  {"name": "Project Update", "code": "projects.update"},
  {"name": "Project Delete", "code": "projects.delete"}
]
```

### 2. Create system roles (super-admin only)

System roles are shared across the platform. Only a **super-admin** (logged in with `authType: "SUPER_ADMIN"`) can create/update/delete system roles and assign them. Typical codes: `admin`, `editor`, `user`.

```bash
# Use a super-admin JWT
export JWT="<super_admin_access_token>"

# Admin role – full permissions
curl -s -X POST http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "admin",
    "name": "Administrator",
    "description": "Full access",
    "projectId": "system",
    "permissions": ["users.view","users.create","users.update","users.delete","roles.view","roles.create","roles.update","roles.delete","roles.assign","roles.revoke","projects.view","projects.create","projects.update","projects.delete"]
  }'

# Editor role
curl -s -X POST http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "editor",
    "name": "Editor",
    "description": "Edit content, manage roles",
    "projectId": "system",
    "permissions": ["users.view","users.update","roles.view","roles.assign","projects.view","projects.update"]
  }'

# User role – read-only
curl -s -X POST http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "user",
    "name": "User",
    "description": "Basic read access",
    "projectId": "system",
    "permissions": ["users.view","roles.view","projects.view"]
  }'
```

### 3. Create a project (super-admin)

```bash
curl -s -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{"code": "my-app", "name": "My Application", "description": "Optional"}'
```

### 4. Create project-scoped roles (optional)

Roles can be scoped to a project by passing `projectId` = project UUID (from step 3):

```bash
# Get project ID from list
curl -s "http://localhost:8080/api/v1/projects?page=1&pageSize=10" -H "Authorization: Bearer $JWT"

# Create role for project
curl -s -X POST http://localhost:8080/api/v1/roles \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "project-editor",
    "name": "Project Editor",
    "description": "Edit this project",
    "projectId": "<project-uuid>",
    "permissions": ["users.view","users.update","projects.view","projects.update"]
  }'
```

### 5. Assign roles to users (super-admin for system roles)

Use the role IDs returned from create/list.

```bash
# Assign system role "user" to a user
curl -s -X POST http://localhost:8080/api/v1/roles/assign \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "<user-uuid>",
    "roleId": "<role-uuid>",
    "projectId": null
  }'

# Assign project role (projectId = project UUID)
curl -s -X POST http://localhost:8080/api/v1/roles/assign \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "<user-uuid>",
    "roleId": "<project-role-uuid>",
    "projectId": "<project-uuid>"
  }'
```

### 6. Check user permissions

```bash
curl -s "http://localhost:8080/api/v1/roles/user/<user-uuid>/permissions" \
  -H "Authorization: Bearer $JWT"
```

Response is a map of permission keys (e.g. `users.view`, `projects.view`) to `true` for the permissions the user has (from all assigned roles, including project-scoped).

---

## 🔗 Relation Tuples (Zanzibar-style)

Relation tuples model **relationships** between objects and subjects, used for “can subject X do relation R on object O?”. Format:

```
<namespace>:<objectId>#<relation>@<subjectNamespace>:<subjectObjectId>[#<subjectRelation>]
```

Examples: `document:readme#viewer@user:alice`, `project:proj-1#member@team:eng#member`.

All relation endpoints are under `/api/v1/relations` and require JWT.

### Grant a relation

```bash
curl -s -X POST http://localhost:8080/api/v1/relations/grant \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "document",
    "objectId": "readme",
    "relation": "viewer",
    "subjectNamespace": "user",
    "subjectObjectId": "alice-uuid"
  }'
```

Optional: `expiresAt` (ISO8601) for temporary access; `subjectRelation` for usersets (e.g. team#member).

### Check a relation (authorization check)

```bash
curl -s -X POST http://localhost:8080/api/v1/relations/check \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "document",
    "objectId": "readme",
    "relation": "viewer",
    "subjectNamespace": "user",
    "subjectObjectId": "alice-uuid"
  }'
# -> {"code":200,"message":"success","data":{"allowed":true,"reason":""}}
```

### Revoke a relation

```bash
curl -s -X POST http://localhost:8080/api/v1/relations/revoke \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "document",
    "objectId": "readme",
    "relation": "viewer",
    "subjectNamespace": "user",
    "subjectObjectId": "alice-uuid"
  }'
```

### List relations (with filters)

```bash
curl -s "http://localhost:8080/api/v1/relations/list?namespace=document&objectId=readme&relation=viewer&page=1&pageSize=10" \
  -H "Authorization: Bearer $JWT"
```

### Expand: list subjects with a relation on an object

```bash
curl -s -X POST http://localhost:8080/api/v1/relations/expand \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "document",
    "objectId": "readme",
    "relation": "viewer"
  }'
# -> {"data":{"subjects":[{"namespace":"user","objectId":"alice-uuid"}],"count":1}}
```

### Team-based access (usersets)

Grant a **team** a relation on an object, and add users as **members** of the team. Then “project:proj-1#contributor” can include “team:eng#member” so all members of `team:eng` get contributor.

```bash
# User bob is member of team engineering
curl -s -X POST http://localhost:8080/api/v1/relations/grant \
  -H "Authorization: Bearer $JWT" \
  -H "Content-Type: application/json" \
  -d '{
    "namespace": "team",
    "objectId": "engineering",
    "relation": "member",
    "subjectNamespace": "user",
    "subjectObjectId": "bob-uuid"
  }'

# Team engineering has contributor on project proj-001 (subjectRelation = member)
curl -s -X POST http://localhost:8080/api/v1/relations/grant \
  -H "Authorization: Bearer $JWT" \
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

### Bulk grant / bulk revoke

- `POST /api/v1/relations/bulk-grant` – body: `{"relations": [ { ... }, ... ]}`  
- `POST /api/v1/relations/bulk-revoke` – same shape for revoke

For more detail and examples, see [docs/RELATION_TUPLES_API.md](docs/RELATION_TUPLES_API.md).

---

## 🔌 Using gRPC

The service exposes an **internal gRPC API** (`AuthInternalService`) for relation-tuple management and permission checks, intended for server-to-server use (e.g. other backend services in your system).

### Configuration

- **Port:** Set `GRPC_PORT` in `.env` (default: `9090`). The server listens on `HTTP_HOST:GRPC_PORT`.
- **Proto & codegen:** Protos live in `presentation/grpc/proto/`. Generate Go code with:
  ```bash
  make buf-gen
  # or: cd presentation/grpc && buf dep update && buf generate
  ```

### RPCs

| RPC | Description |
|-----|-------------|
| **GrantRelationTuple** | Create a Zanzibar-style relation tuple (same semantics as `POST /relations/grant`). |
| **CheckPermission** | Check whether a subject has a relation on an object (same as `POST /relations/check`). |
| **GetUserPermissions** | Get the permission map for a user (same as `GET /roles/user/:userId/permissions`). |

Request/response types mirror the REST aggregates (namespace, object_id, relation, subject_*, optional expires_at, etc.). See `presentation/grpc/proto/auth_internal.proto`.

### Using the internal client (Go)

Other Go services can call this API using the **internal gRPC client**:

```go
import (
    "context"
    clientgrpc "github.com/hiamthach108/dreon-auth/internal/client/grpc"
    authinternal "github.com/hiamthach108/dreon-auth/presentation/grpc/gen/proto"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// From target address
conn, _ := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
defer conn.Close()

client := clientgrpc.NewAuthInternalClientFromConn(conn)
defer client.Close()

// Grant a relation
resp, err := client.Client().GrantRelationTuple(ctx, &authinternal.GrantRelationTupleRequest{
    Namespace:        "document",
    ObjectId:         "readme",
    Relation:         "viewer",
    SubjectNamespace: "user",
    SubjectObjectId:  "alice-uuid",
})

// Check permission
check, err := client.Client().CheckPermission(ctx, &authinternal.CheckPermissionRequest{
    Namespace: "document", ObjectId: "readme", Relation: "viewer",
    SubjectNamespace: "user", SubjectObjectId: "alice-uuid",
})
// check.Allowed

// User permissions
perms, err := client.Client().GetUserPermissions(ctx, &authinternal.GetUserPermissionsRequest{UserId: "user-uuid"})
// perms.Permissions map[string]bool
```

For app-config–based dial (e.g. with Fx), use `clientgrpc.NewAuthInternalClientFromAppConfig(ctx, cfg, nil)` so the target is derived from `HTTP_HOST` and `GRPC_PORT`.

### Testing with grpcurl (optional)

```bash
# List services
grpcurl -plaintext localhost:9090 list

# Describe AuthInternalService
grpcurl -plaintext localhost:9090 describe dreon.auth.internal.v1.AuthInternalService

# Example: CheckPermission
grpcurl -plaintext -d '{
  "namespace": "document",
  "object_id": "readme",
  "relation": "viewer",
  "subject_namespace": "user",
  "subject_object_id": "alice-uuid"
}' localhost:9090 dreon.auth.internal.v1.AuthInternalService/CheckPermission
```

---

## 📊 Auth Flows

### Email login

```
POST /auth/login { "authType": "EMAIL", "email": "...", "password": "..." }
  -> accessToken, refreshToken, expires
```

### Google OAuth

1. **Start:** `POST /auth/login` with `{ "authType": "GOOGLE", "redirectUrl": "https://yourapp.com/callback" }`  
   → response: `refreshState`, `redirectUrl` (Google auth URL).  
2. **Redirect** user to `redirectUrl`.  
3. User signs in with Google; Google redirects to **GET** `.../auth/google/callback?code=...&state=<refreshState>`.  
4. Backend exchanges code, stores user data keyed by `state`, redirects browser to `redirectUrl?refreshState=<refreshState>`.  
5. **Session:** frontend calls `POST /auth/session-from-state` with `{ "refreshState": "..." }` → access + refresh tokens.

### Token refresh

```
POST /auth/refresh-token { "refreshToken": "..." } -> new accessToken, refreshToken
```

---

## 🛠️ Project Structure

```
dreon-auth/
├── cmd/main.go
├── config/                 # App config, permissions path
├── internal/
│   ├── dto/               # Request/response DTOs
│   ├── errorx/            # Error codes and messages
│   ├── model/             # Domain models (User, Session, Role, RelationTuple, …)
│   ├── repository/        # Data access
│   ├── service/           # Business logic (auth, user, project, role, relation)
│   └── shared/            # Constants, permission registry, helpers
├── pkg/                    # Cache, JWT, logger, …
├── presentation/
│   ├── http/               # Echo handlers, middleware
│   └── grpc/               # gRPC server, proto, generated code (AuthInternalService)
├── internal/client/grpc/   # Internal gRPC client for AuthInternalService
├── docs/                   # RELATION_TUPLES_API.md, etc.
├── docker-compose.yml
├── Dockerfile
└── Makefile
```

### Testing

```bash
make test
```

---

## 📄 License

MIT License.

## 👤 Author

**hiamthach108** – [GitHub](https://github.com/hiamthach108)

## Acknowledgments

- [Fx](https://uber-go.github.io/fx/) dependency injection, [Echo](https://echo.labstack.com/) HTTP
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt) RS256, [golang.org/x/oauth2](https://pkg.go.dev/golang.org/x/oauth2) Google OAuth
- GORM, Redis, PostgreSQL
- Authorization model inspired by [Google Zanzibar](https://research.google/pubs/pub48190/)
