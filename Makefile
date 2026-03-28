.PHONY: migrate-up migrate-down server mockgen

include .env

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" -verbose up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" -verbose down

migrate-down-1:
	migrate -path migrations -database "$(DATABASE_URL)" -verbose down 1

server:
	go run ./cmd/api
