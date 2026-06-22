.PHONY: run test build up down tidy

run:
	go run ./cmd/server

build:
	go build ./...

test:
	go test ./...

tidy:
	go mod tidy

up:
	docker compose up -d

down:
	docker compose down
