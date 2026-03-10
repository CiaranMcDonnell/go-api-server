# go-api-server

A reusable Go backend server with built-in authentication, audit logging, and user management. Built with [Gin](https://github.com/gin-gonic/gin), PostgreSQL, and clean architecture principles.

## Features

- JWT authentication with HTTP-only cookie sessions
- User registration with struct-tag validation and Argon2id password hashing
- CRUD resource management with ownership enforcement
- Automatic audit trail via middleware
- Database migrations with golang-migrate
- Rate limiting on auth endpoints
- Prometheus metrics and Grafana dashboards
- Structured logging with `log/slog`
- Docker ready with multi-stage builds

## Getting Started

1. Clone the repository and copy the example env file:
   ```bash
   git clone https://github.com/ciaranmcdonnell/go-api-server.git
   cd go-api-server
   cp app.env.example app.env
   ```

2. Update `app.env` with your settings (especially `JWT_SECRET`).

3. Start with Docker Compose:
   ```bash
   docker compose up --build
   ```

The server will be available at `http://localhost:8080`.

## API

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/v1/auth/register` | No | Register a new user |
| `POST` | `/api/v1/auth/login` | No | Login and receive auth cookie |
| `POST` | `/api/v1/auth/logout` | Yes | Clear auth cookie |
| `GET` | `/api/v1/auth/me` | Yes | Get current user info |
| `POST` | `/api/v1/items` | Yes | Create an item |
| `GET` | `/api/v1/items` | Yes | List items |
| `GET` | `/api/v1/items/:id` | Yes | Get an item |
| `PUT` | `/api/v1/items/:id` | Yes | Update an item |
| `DELETE` | `/api/v1/items/:id` | Yes | Delete an item |
| `GET` | `/health` | No | Health check |

## Configuration

All configuration is via environment variables (or an `app.env` file):

| Variable | Default | Description |
|---|---|---|
| `DB_SOURCE` | — | PostgreSQL connection string |
| `JWT_SECRET` | — | Secret key for signing JWTs |
| `SERVER_ADDRESS` | `0.0.0.0:8080` | Server listen address |
| `ENVIRONMENT` | `development` | `development` or `production` |
| `JWT_EXPIRATION_HOURS` | `8` | JWT token lifetime |
| `CORS_ORIGINS` | `http://localhost:3000` | Allowed CORS origins |
| `DB_MAX_CONNS` | `25` | Max database connections |
| `DB_MIN_CONNS` | `5` | Min database connections |
| `RUN_MIGRATIONS` | `false` | Run migrations on startup |

## License

MIT License. See [LICENSE](LICENSE) for details.
