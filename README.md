# URL Shortener

A URL shortening service built with Go — shorten links, track clicks, and view analytics. Supports anonymous and authenticated usage with custom aliases, link expiry, and async analytics.

## Tech Stack

Go · Gin · PostgreSQL · Redis · JWT · Docker Compose

## Getting Started

```bash
git clone https://github.com/katatrina/url-shortener.git
cd url-shortener
cp .env.example .env
docker compose up -d
make migrate-up
make server
```

The server starts on `http://localhost:8080` by default. See `.env.example` for all configuration options.

## Architecture

```
Handler (HTTP) → Service (Business Logic) → Repository (Database)
```

Dependencies flow inward. The service layer defines interfaces for its dependencies, enabling unit testing with mocks.

## Roadmap

- [x] Core MVP: shorten, redirect, auth, CRUD, unit tests
- [x] Redis caching + rate limiting
- [x] Async analytics pipeline
- [x] Structured logging (log/slog)
- [ ] Microservices migration
- [ ] Frontend + public deployment

## License

This project is for learning purposes.