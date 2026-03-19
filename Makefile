SHELL := /bin/bash

.PHONY: dev build test swagger swagger-check sqlc migrate-up migrate-down

ENV_FILE ?= .env.dev
SQLC ?= sqlc
MIGRATE ?= migrate

dev:
	set -a; source $(ENV_FILE); set +a; go run .

build:
	go build ./...

test:
	go test ./...

swagger:
	go generate ./...

swagger-check:
	./scripts/check-swagger.sh

sqlc:
	$(SQLC) generate

migrate-up:
	@if [[ -z "$$DB_DSN" ]]; then echo "DB_DSN is required"; exit 1; fi
	$(MIGRATE) -path db/migrations -database "$$DB_DSN" up

migrate-down:
	@if [[ -z "$$DB_DSN" ]]; then echo "DB_DSN is required"; exit 1; fi
	$(MIGRATE) -path db/migrations -database "$$DB_DSN" down 1
