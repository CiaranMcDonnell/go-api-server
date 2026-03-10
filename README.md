# go-api-server

A reusable Go backend server with built-in authentication, audit logging, and user management. Built with [Gin](https://github.com/gin-gonic/gin), PostgreSQL, and clean architecture principles.

## Features

- **JWT Authentication** — Login, logout, and session management via HTTP-only cookies
- **User Management** — Registration with input validation and bcrypt password hashing
- **Audit Trail** — Automatic request logging via middleware with generic entity tracking
- **Database Migrations** — Automatic schema management with golang-migrate
- **Clean Architecture** — Handlers, services, and repositories with clear separation of concerns
- **Structured Logging** — JSON-friendly logging with `log/slog`
- **Request Tracing** — Auto-generated `X-Request-ID` headers
- **Docker Ready** — Multi-stage build with non-root user, docker-compose with PostgreSQL

## Project Structure

```
.
├── api/v1/             # HTTP routing and route groups
├── cmd/server/         # Application entrypoint
├── internal/
│   ├── core/
│   │   ├── audit/      # Audit trail (domain, service, repository, middleware)
│   │   ├── auth/       # Authentication (handlers, middleware, service)
│   │   ├── common/     # Shared registries, health check, request ID middleware
│   │   └── user/       # User registration and queries
│   └── database/       # Connection pool, migrations, SQL schemas
├── models/             # Shared data models
├── pkg/utils/          # Config, constants, validation, password hashing
├── Dockerfile
└── docker-compose.yml
```

## Getting Started

### Prerequisites

- Docker and Docker Compose
- (Optional) Go 1.23+ for local development

### Quick Start

1. Clone the repository:
   ```bash
   git clone https://github.com/ciaranmcdonnell/go-api-server.git
   cd go-api-server
   ```

2. Copy the example environment file:
   ```bash
   cp app.env.example app.env
   ```

3. Update `app.env` with your settings (especially `JWT_SECRET`).

4. Start with Docker Compose:
   ```bash
   docker compose up --build
   ```

   To run database migrations on startup:
   ```bash
   RUN_MIGRATIONS=true docker compose up --build
   ```

The server will be available at `http://localhost:8080`.

### Configuration

All configuration is via environment variables (or an `app.env` file):

| Variable | Default | Description |
|---|---|---|
| `DB_SOURCE` | — | PostgreSQL connection string |
| `JWT_SECRET` | — | **Required.** Secret key for signing JWTs (min 32 chars recommended) |
| `SERVER_ADDRESS` | `0.0.0.0:8080` | Server listen address |
| `ENVIRONMENT` | `development` | `development` or `production` |
| `JWT_EXPIRATION_HOURS` | `8` | JWT token lifetime |
| `CORS_ORIGINS` | `http://localhost:3000` | Allowed CORS origins (comma-separated) |
| `DB_MAX_CONNS` | `25` | Max database connections |
| `DB_MIN_CONNS` | `5` | Min database connections |
| `RUN_MIGRATIONS` | `false` | Run database migrations on startup |

## API Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/v1/auth/register` | No | Register a new user |
| `POST` | `/api/v1/auth/login` | No | Login and receive auth cookie |
| `POST` | `/api/v1/auth/logout` | No | Clear auth cookie |
| `GET` | `/api/v1/auth/me` | Yes | Get current user info |
| `GET` | `/api/v1/health` | No | Health check |

## License

MIT License. See [LICENSE](LICENSE) for details.
