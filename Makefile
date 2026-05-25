.PHONY: dev test build lint sqlc migrate-up migrate-down migrate-new

GOOSE_DRIVER   ?= postgres
MIGRATIONS_DIR  = internal/db/migrations

# Read DATABASE_URL from .env (strip surrounding quotes) or from the environment.
# This means `make migrate-up` works without sourcing .env first.
_ENV_DB_URL := $(shell sed -n "s/^DATABASE_URL=['\"]\\?\\(.*[^'\"]\\)['\"]\\?$$/\\1/p" .env 2>/dev/null | head -1)
GOOSE_DBSTRING ?= $(if $(DATABASE_URL),$(DATABASE_URL),$(_ENV_DB_URL))

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
