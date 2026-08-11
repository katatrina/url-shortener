DATABASE_URL ?= postgres://root:secret@localhost:5432/url_shortener?sslmode=disable
MIGRATIONS_DIR = ./migrations

.PHONY: geoip

build:
	go build ./...

run:
	go run ./cmd/server

lint:
	golangci-lint run

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

db2dbml:
	db2dbml postgres "$(DATABASE_URL)" -o docs/schema.dbml

GEOIP_DB = geoip/dbip-country-lite.mmdb
GEOIP_MONTH ?= 2026-08

geoip:
	mkdir -p geoip
	curl -fsSL "https://download.db-ip.com/free/dbip-country-lite-$(GEOIP_MONTH).mmdb.gz" | gunzip > $(GEOIP_DB)

compose:
	docker compose up -d

compose-down:
	docker compose down