.PHONY: migrate-up migrate-down server mockgen

migrate-up:
	migrate -path migrations -database "postgres://root:secret@localhost:5432/url_shortener?sslmode=disable" -verbose up

migrate-down:
	migrate -path migrations -database "postgres://root:secret@localhost:5432/url_shortener?sslmode=disable" -verbose down

migrate-down-1:
	migrate -path migrations -database "postgres://root:secret@localhost:5432/url_shortener?sslmode=disable" -verbose down 1

server:
	go run ./cmd/api
