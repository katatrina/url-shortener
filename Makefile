.PHONY: migrate-up migrate-down server mockgen infra infra-down lint service-test

include .env

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" -verbose up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" -verbose down

migrate-down-1:
	migrate -path migrations -database "$(DATABASE_URL)" -verbose down 1

server:
	go run ./cmd/api

infra:
	docker compose up -d

infra-down:
	docker compose down

lint:
	golangci-lint run

service-test:
	go test -v ./internal/service