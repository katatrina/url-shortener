.PHONY: server compose

compose:
	docker compose up -d

server:
	go run ./cmd/server/main.go