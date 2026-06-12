DATABASE_URL ?= postgres://root:secret@localhost:5432/url_shortener?sslmode=disable
MIGRATIONS_DIR = ./migrations

# --- Dev hằng ngày: chạy thẳng trên máy ---
server:
	go run ./cmd/server

lint:
	golangci-lint run

# --- Migration ---
migrate-create:
	goose -dir $(MIGRATIONS_DIR) -s create $(name) sql

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status -v

# --- Hạ tầng local ---
network:
	docker network inspect url-shortener-network >/dev/null 2>&1 || docker network create url-shortener-network

compose: network
	docker compose up -d

compose-down:
	docker compose down

# --- Test container ---
build:
	docker build -t url-shortener .

run:
	docker run -p 8080:8080 \
		--network url-shortener-network \
		-e APP_ENV=development \
		-e LOG_LEVEL=info \
		-e DATABASE_URL="postgres://root:secret@postgres:5432/url_shortener?sslmode=disable" \
		url-shortener