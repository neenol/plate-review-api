.PHONY: dev test build lint sqlc migrate-up migrate-down migrate-new

GOOSE_DRIVER ?= postgres
GOOSE_DBSTRING ?= $(DATABASE_URL)
MIGRATIONS_DIR = internal/db/migrations

dev:
	air

test:
	go test ./...

build:
	go build -o bin/server ./cmd/server

lint:
	go vet ./...

sqlc:
	cd internal/db && sqlc generate

migrate-up:
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" down

migrate-new:
	goose -dir $(MIGRATIONS_DIR) create $(name) sql
