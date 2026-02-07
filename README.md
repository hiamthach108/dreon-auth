# dreon-auth

A Go-based **Auth Service** for authentication, sessions, and token management. Built with clean architecture and designed to evolve toward **Google Zanzibar**-style relationship-based authorization.

## 🎯 Overview

dreon-auth is a dedicated authentication service that provides user management, JWT (RS256) issue/verify, and session handling. It is built with Go and Fx for dependency injection, PostgreSQL for persistence, and Redis for caching/sessions. The architecture and data model are prepared for a future migration to **Zanzibar**-inspired permission checks (relationship-based access control, multi-tenant relation tuples).

## 🏗️ Architecture

### Tech Stack

- **Go 1.25** – Backend auth service
- **PostgreSQL** – Users, tenants, sessions, and (future) relation tuples
- **Redis** – Session store and cache
- **JWT (RS256)** – Asymmetric token signing and verification
- **Docker** – Containerization and orchestration

### Architecture Design

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│     dreon-auth (Auth Service)      │
│         Go + Fx DI + Echo           │
│  • Login / Register / Token        │
│  • Session management              │
│  • (Future) Zanzibar Check/Expand  │
└─────────────────────────────────────┘
       │              │
       ▼              ▼
┌──────────┐   ┌─────────────┐
│PostgreSQL│   │   Redis     │
│          │   │             │
│- Users   │   │- Sessions   │
│- Tenants │   │- Cache     │
│- Sessions│   │             │
│- (Future)│   └─────────────┘
│  Tuples  │
└──────────┘
```

### Component Responsibilities

#### PostgreSQL
- **Users** – Identity and credentials
- **Tenants** – Multi-tenant isolation (per client/org)
- **Sessions** – Persistent session metadata
- **Future (Zanzibar)** – Relation tuples, namespace definitions

#### Redis
- Session storage and invalidation
- Cache for tokens / hot data
- Optional: high-throughput permission cache (when Zanzibar is in place)

#### JWT (RS256)
- Issue access/refresh tokens (private key)
- Verify tokens in APIs (public key)
- Configurable issuer, audience, and expiry

## 🚀 Features

- ✅ **Auth Service** – User CRUD, login, token issue/verify
- ✅ **JWT RS256** – Asymmetric keys, configurable via env/file
- ✅ **Multi-tenant** – Tenant-scoped users and (future) permissions
- ✅ **Sessions** – Session model and storage
- ✅ **RESTful API** – HTTP + Echo
- ✅ **Clean architecture** – Interfaces, DTOs, repositories, services
- ✅ **Docker-ready** – docker-compose for local dev
- 🔜 **Zanzibar-style** – Relationship-based authorization (planned)

## 📦 Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.25+ (for local development)
- Make (optional, for using Makefile commands)

### Quick Start with Docker

1. Clone the repository:
```bash
git clone https://github.com/hiamthach108/dreon-auth.git
cd dreon-auth
```

2. Start all services:
```bash
docker-compose up -d
```

3. Access the services:
- dreon-auth API: http://localhost:8080
- PostgreSQL: localhost:5432
- Redis: localhost:6379

### Local Development

1. Install dependencies:
```bash
go mod download
```

2. Copy and configure environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. **Set up JWT key pairs** (required for auth):

   The app uses **RS256** (asymmetric) JWT. You must provide a private key (to sign tokens) and a public key (to verify them).

   **Generate keys with OpenSSL:**

   ```bash
   # Create a directory for keys (optional; add keys/ to .gitignore)
   mkdir -p keys

   # Generate 2048-bit RSA private key
   openssl genrsa -out keys/private.pem 2048

   # Derive public key from private key
   openssl rsa -pubout -in keys/private.pem -out keys/public.pem
   ```

   **Configure in `.env`** — use either **file paths** or **inline PEM**:

   - **Option A – File paths** (recommended for local dev):

     ```env
     JWT_PRIVATE_KEY=keys/private.pem
     JWT_PUBLIC_KEY=keys/public.pem
     JWT_ACCESS_TOKEN_EXPIRES_IN=3600
     JWT_REFRESH_TOKEN_EXPIRES_IN=86400
     ```

   - **Option B – Inline PEM** (e.g. for Docker/CI): set `JWT_PRIVATE_KEY` and `JWT_PUBLIC_KEY` to the full PEM content (including `-----BEGIN ... -----` and newlines). The app treats values starting with `-----BEGIN` as raw PEM.

   **Security notes:**

   - Keep `private.pem` only on the service that **issues** tokens; never commit it.
   - Only the **public** key is needed on services that only **verify** tokens.
   - For production, use at least 2048-bit RSA; 4096-bit is stronger: `openssl genrsa -out keys/private.pem 4096`.

4. Start dependencies (PostgreSQL, Redis):
```bash
docker-compose up -d postgres redis
```

5. Run the application:
```bash
go run cmd/main.go
```

Or use Make:
```bash
make run
```

## 📊 Data Flow

### Auth / Token Flow
```
1. Client sends login or token request
   ↓
2. Validate credentials / refresh token
   ↓
3. Load user (and tenant) from PostgreSQL
   ↓
4. Issue JWT (RS256) with private key
   ↓
5. Optionally create/update session (Redis + PostgreSQL)
   ↓
6. Return access (and refresh) token to client
```

### Token Verification Flow (downstream services)
```
1. Request includes Bearer token
   ↓
2. Verify JWT with public key (signature + expiry)
   ↓
3. Extract user/tenant from claims
   ↓
4. Proceed with request (or future: Zanzibar Check)
```

## 🔜 Roadmap: Google Zanzibar

Authorization is planned to align with the Google Zanzibar model:

- **Relationship-based access control (ReBAC)** – Permissions as relations (e.g. `document:doc-1#viewer@user:alice`).
- **Multi-tenant relation tuples** – All tuples scoped by `tenant_id`; models (`relation_tuples`, `namespace_definitions`) are already in place.
- **Check / Expand APIs** – “Can user X do Y on Z?” and “List subjects with relation R on resource O” with consistent, scalable evaluation.

The codebase is structured so that relation tuples and namespaces can be wired into services and HTTP/gRPC endpoints when you implement the Zanzibar engine.

## 🛠️ Development

### Project Structure

```
dreon-auth/
├── cmd/                    # Application entrypoints
│   └── main.go
├── config/                 # Configuration management
├── internal/               # Private application code
│   ├── dto/               # Data Transfer Objects
│   ├── errorx/            # Custom error handling
│   ├── model/             # Domain models
│   ├── repository/        # Data access layer
│   ├── service/           # Business logic
│   └── shared/            # Shared utilities
├── pkg/                   # Public reusable packages
│   ├── cache/             # Cache abstraction
│   ├── database/          # Database clients
│   ├── jwt/               # JWT utilities
│   ├── kafka/             # Kafka integration
│   └── logger/            # Logging utilities
├── presentation/          # Presentation layer
│   ├── grpc/              # gRPC handlers
│   ├── http/              # HTTP handlers & middleware
│   └── socket/            # WebSocket handlers
├── script/                # Build and deployment scripts
├── docker-compose.yml     # Docker orchestration
├── Dockerfile             # Application container
└── Makefile              # Build automation
```

### Testing

Run tests:
```bash
make test
```
## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License.

## 👤 Author

**hiamthach108**
- GitHub: [@hiamthach108](https://github.com/hiamthach108)

## Acknowledgments

- Built with [Fx](https://uber-go.github.io/fx/) dependency injection and [Echo](https://echo.labstack.com/) for HTTP
- JWT handling via [golang-jwt/jwt](https://github.com/golang-jwt/jwt) with RS256
- Uses GORM for database operations
- Future authorization design inspired by Google Zanzibar