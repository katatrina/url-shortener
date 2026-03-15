.PHONY: migrate-up migrate-down server

migrate-up:
	migrate -path migrations -database "postgres://root:secret@localhost:5432/url_shortener?sslmode=disable" -verbose up

migrate-down:
	migrate -path migrations -database "postgres://root:secret@localhost:5432/url_shortener?sslmode=disable" -verbose down

server:
	go run ./cmd/api