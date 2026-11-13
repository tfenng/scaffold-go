SHELL := /bin/bash

.PHONY: dev dev-web build web release test swagger swagger-check sqlc migrate-up migrate-down

ENV_FILE ?= .env.dev
SQLC ?= sqlc
WEB_DIR ?= scaffold-web
DIST_DIR ?= internal/webui/dist
BIN ?= bin/scaffold-api

# Run the API locally. The web UI is expected to run separately via `make dev-web`.
dev:
	set -a; source $(ENV_FILE); set +a; go run .

# Run the Next.js dev server against the local API.
dev-web:
	cd $(WEB_DIR) && pnpm dev

# Compile check. Works with the placeholder dist; does not bundle a fresh UI.
build:
	go build ./...

# Build the static frontend and copy it into the embed directory.
web:
	cd $(WEB_DIR) && pnpm install --frozen-lockfile && pnpm build
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)
	cp -R $(WEB_DIR)/out/. $(DIST_DIR)/
	touch $(DIST_DIR)/.gitkeep

# Produce the single self-contained binary (API + embedded web UI).
release: web
	mkdir -p $(dir $(BIN))
	go build -trimpath -o $(BIN) .

test:
	go test ./...

swagger:
	go generate ./...

swagger-check:
	./scripts/check-swagger.sh

sqlc:
	$(SQLC) generate

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down
