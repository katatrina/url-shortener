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

migrate-down-1:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down 1

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" status -v

# --- Schema docs ---
db2dbml:
	db2dbml postgres "$(DATABASE_URL)" -o docs/schema.dbml

# --- GeoIP (DB-IP Country Lite, CC BY 4.0, cập nhật hàng tháng) ---
GEOIP_DB = geoip/dbip-country-lite.mmdb
GEOIP_MONTH = $(shell date +%Y-%m)

geoip:
	mkdir -p geoip
	curl -fsSL "https://download.db-ip.com/free/dbip-country-lite-$(GEOIP_MONTH).mmdb.gz" | gunzip > $(GEOIP_DB)

# --- Hạ tầng local ---
network:
	docker network inspect url-shortener-network >/dev/null 2>&1 || docker network create url-shortener-network

compose: network
	docker compose up -d

compose-down:
	docker compose down