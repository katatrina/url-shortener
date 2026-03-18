# URL Shortener

A URL shortening service that converts long URLs into short, shareable links and tracks click analytics. Built with Go, designed as a monolith with a clear path to microservices.

## Features

- **Shorten URLs** — paste a long URL, get a short link
- **Anonymous or authenticated** — anyone can shorten, logged-in users can manage their links
- **Custom aliases** — choose your own short code (e.g., `/myrepo`)
- **Click tracking** — every redirect is counted
- **Link expiry** — URLs can have an optional expiration date
- **Soft delete** — deleted URLs are recoverable

## Tech Stack

| Component      | Technology          |
|----------------|---------------------|
| Language        | Go 1.25             |
| HTTP Framework  | Gin                 |
| Database        | PostgreSQL 16 (pgx) |
| Auth            | JWT (HS256)         |
| Config          | Viper               |
| Infrastructure  | Docker Compose      |

## Project Structure

```
url-shortener/
├── cmd/api/main.go            # Application entrypoint
├── internal/
│   ├── config/                # Environment config loading
│   ├── handler/               # HTTP handlers (request/response)
│   ├── middleware/             # Auth middleware (strict + optional)
│   ├── service/               # Business logic
│   ├── repository/            # Database queries
│   ├── model/                 # Domain models and errors
│   ├── token/                 # JWT creation and verification
│   ├── request/               # Validation, normalization, pagination
│   ├── response/              # Standardized API response format
│   └── shortcode/             # Short code generation (crypto/rand + base62)
├── migrations/                # PostgreSQL migrations
├── docker-compose.yml
├── Makefile
└── go.mod
```

## Getting Started

### Prerequisites

- Go 1.25+
- Docker and Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI

### Setup

1. Clone the repository:

```bash
git clone https://github.com/katatrina/url-shortener.git
cd url-shortener
```

2. Start the database:

```bash
docker compose up -d
```

3. Run migrations:

```bash
make migrate-up
```

4. Create your `.env` file:

```bash
cp .env.example .env
```

5. Start the server:

```bash
make server
```

The server starts on `http://localhost:8080` by default.

## API Endpoints

### Public

| Method | Path                    | Description                                        |
|--------|-------------------------|----------------------------------------------------|
| GET    | `/:code`                | Redirect to original URL (302 Found)               |
| POST   | `/api/v1/shorten`       | Create short URL (works with or without auth)      |
| POST   | `/api/v1/auth/register` | Register a new account                             |
| POST   | `/api/v1/auth/login`    | Login and receive JWT                              |

### Protected (require JWT)

| Method | Path                     | Description                  |
|--------|--------------------------|------------------------------|
| GET    | `/api/v1/me/urls`        | List your URLs (paginated)   |
| GET    | `/api/v1/me/urls/:code`  | Get URL details              |
| DELETE | `/api/v1/me/urls/:code`  | Soft delete a URL            |

### Example: Shorten a URL

```bash
# Anonymous
curl -X POST http://localhost:8083/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"originalUrl": "https://github.com/katatrina/url-shortener"}'

# With custom alias (authenticated)
curl -X POST http://localhost:8083/api/v1/shorten \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"originalUrl": "https://github.com/katatrina/url-shortener", "customAlias": "myrepo"}'
```

### Example: Register and Login

```bash
# Register
curl -X POST http://localhost:8083/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "displayName": "Alice", "password": "securepass123"}'

# Login
curl -X POST http://localhost:8083/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "alice@example.com", "password": "securepass123"}'
```

## API Response Format

All endpoints return a consistent JSON structure:

```json
{
  "success": true,
  "code": "OK",
  "message": "URL shortened successfully",
  "data": {
    "shortCode": "aB3kX9m",
    "shortUrl": "http://localhost:8080/aB3kX9m",
    "originalUrl": "https://github.com/katatrina/url-shortener",
    "clickCount": 0,
    "createdAt": 1718900000
  },
  "meta": {
    "requestId": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": 1718900000
  }
}
```

## Architecture

```
Handler (HTTP) → Service (Business Logic) → Repository (Database)
```

Dependencies flow inward only. The service layer defines interfaces for its dependencies, enabling unit testing with mocks at every layer.

### Key Design Decisions

- **302 redirect instead of 301** — 301 causes browsers to cache the redirect permanently, making click tracking impossible. 302 ensures every click hits the server.
- **`crypto/rand` for short codes** — `math/rand` is predictable if you know the seed. `crypto/rand` reads from `/dev/urandom`, making codes unguessable.
- **Soft delete** — URLs are marked as deleted (`deleted_at` timestamp) rather than removed, allowing recovery and maintaining referential integrity.
- **Optional auth on shorten** — the `POST /shorten` endpoint uses optional auth middleware. No token = anonymous URL. Valid token = URL linked to account. Invalid token = 401 rejection.

## Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run a specific package
go test ./internal/service/ -v
```

### Generating Mocks

Mocks are generated using [gomock](https://github.com/uber-go/mock):

```bash
go install go.uber.org/mock/mockgen@latest
go generate ./internal/mock/...
```

## Configuration

| Variable       | Description                        | Example                                              |
|----------------|------------------------------------|------------------------------------------------------|
| `SERVER_PORT`  | Port the server listens on         | `8083`                                               |
| `DATABASE_URL` | PostgreSQL connection string       | `postgres://root:secret@localhost:5432/url_shortener` |
| `BASE_URL`     | Public base URL for short links    | `http://localhost:8080`                               |
| `JWT_SECRET`   | Secret key for signing JWT (≥32B)  | `PgK13YiT0Upo...`                                   |
| `JWT_EXPIRY`   | Token expiration duration          | `24h`                                                |

## Roadmap

- [x] **Phase 1** — Core MVP: shorten, redirect, auth, CRUD, unit tests
- [ ] **Phase 2** — Redis caching + rate limiting
- [ ] **Phase 3** — Async analytics pipeline (goroutines + channels)
- [ ] **Phase 4** — Observability + CI/CD (Prometheus, Grafana, GitHub Actions)
- [ ] **Phase 5** — Microservices migration
- [ ] **Phase 6** — Frontend + public deployment

## License

This project is for learning purposes.