# dreon-auth

A high-performance, real-time leaderboard service designed to handle millions of ranking operations with event-driven architecture.

## 🎯 Overview

dreon-auth is a scalable leaderboard service that supports different types of leaderboards with real-time updates, historical tracking, and analytics capabilities. Built with Go and leveraging Redis for blazing-fast ranking operations.

## 🏗️ Architecture

### Tech Stack

- **Go 1.25** - High-performance backend service
- **PostgreSQL** - Transactional database and source of truth
- **Redis** - Real-time leaderboard engine with Sorted Sets and Streams
- **ClickHouse** - Time-series data and analytics storage
- **Keycloak** - Authentication and authorization
- **Docker** - Containerization and orchestration

### Architecture Design

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────┐
│        dreon-auth API Server        │
│         (Go with Fx DI)             │
└─────────────────────────────────────┘
       │         │            │
       ▼         ▼            ▼
┌──────────┐ ┌─────────┐ ┌────────────┐
│PostgreSQL│ │Redis    │ │ ClickHouse │
│          │ │         │ │            │
│- Users   │ │- ZSET   │ │- Snapshots │
│- Config  │ │- Stream │ │- History   │
│- Metadata│ │         │ │- Analytics │
└──────────┘ └─────────┘ └────────────┘
```

### Component Responsibilities

#### PostgreSQL - Transactional Database
- Store persistent data (users, leaderboard configurations, metadata)
- Handle transactional operations
- Source of truth for critical business data
- Ensure data consistency and integrity

#### Redis - Real-time Leaderboard Engine
- **Sorted Sets (ZSET)**: Core leaderboard functionality with O(log N) ranking
- **Redis Streams**: Event-driven architecture for real-time updates
- Ultra-fast reads/writes for live rankings
- Handle millions of score updates per second
- Real-time notifications and pub/sub

#### ClickHouse - Analytics & Historical Data
- Store historical snapshots of leaderboard states
- Time-series data for trend analysis
- Track score changes and user activity over time
- Fast aggregation queries for reports and dashboards
- Analytics for business intelligence

#### Keycloak - Authentication & Authorization
- User authentication and session management
- Role-based access control (RBAC)
- API security and token validation
- Multi-tenancy support

## 🚀 Features

- ✅ Real-time leaderboard updates with sub-millisecond latency
- ✅ Multiple leaderboard types (global, time-based, group-based)
- ✅ Event-driven architecture with Redis Streams
- ✅ Historical tracking and analytics
- ✅ Scalable and distributed design
- ✅ RESTful API with HTTP
- ✅ Secure authentication and authorization
- ✅ Docker-ready with docker-compose

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
- Keycloak Admin: http://localhost:8000 (admin/admin)
- PostgreSQL: localhost:5432
- Redis: localhost:6379
- ClickHouse HTTP: localhost:8123

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

3. Start dependencies (PostgreSQL, Redis, ClickHouse):
```bash
docker-compose up -d postgres redis clickhouse
```

4. Run the application:
```bash
go run cmd/main.go
```

Or use Make:
```bash
make run
```

## 📊 Data Flow

### Score Update Flow
```
1. Client submits score update
   ↓
2. Validate & authenticate request
   ↓
3. Update Redis Sorted Set (instant ranking)
   ↓
4. Publish event to Redis Stream
   ↓
5. Persist to PostgreSQL (transactional)
   ↓
6. Store snapshot in ClickHouse (async)
   ↓
7. Return updated rank to client
```

### Leaderboard Query Flow
```
1. Client requests leaderboard
   ↓
2. Check Redis cache
   ↓
3. If miss → Query PostgreSQL
   ↓
4. Return rankings with pagination
```

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

## 🙏 Acknowledgments

- Built with [Fx](https://uber-go.github.io/fx/) dependency injection
- Powered by Redis Sorted Sets for optimal leaderboard performance
- Uses GORM for database operations