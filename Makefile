.PHONY: server compose compose-down network build run

# --- Dev hằng ngày: chạy thẳng trên máy ---
server:
	go run ./cmd/server

# --- Hạ tầng local ---
network:
	docker network inspect url-shortener-network >/dev/null 2>&1 || docker network create url-shortener-network

compose: network
	docker compose up -d

compose-down:
	docker compose down

# --- Deploy ---
build:
	docker build -t url-shortener .

run:
	docker run -p 8080:8080 \
		--network url-shortener-network \
		-e APP_ENV=development \
		-e LOG_LEVEL=info \
		-e DATABASE_URL="postgres://root:secret@postgres:5432/url_shortener?sslmode=disable" \
		url-shortener