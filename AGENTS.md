# Repository Guidelines

## Project Structure & Module Organization
`main.go` is the entrypoint that wires config, logging, database access, services, and HTTP routes. Keep application code under `internal/`: `internal/http` for handlers and middleware, `internal/service` for business rules, `internal/db` for store and database setup, `internal/config` for env-driven config, and `internal/errs` for shared error types. SQL sources live in `db/query/*.sql`, migrations in `db/migrations/`, and generated artifacts in `internal/db/query/` and `docs/`. Treat generated files as derived output: regenerate them instead of editing by hand.

## Build, Test, and Development Commands
Use the `Makefile` as the default interface:

- `make dev` loads `.env.dev` and runs the API locally.
- `make build` verifies the project compiles with `go build ./...`.
- `make test` runs all Go tests with `go test ./...`.
- `make sqlc` regenerates query code from `sqlc.yaml`.
- `make swagger` regenerates Swagger docs via `go generate ./...`.
- `make swagger-check` regenerates docs and fails if committed Swagger files are stale.
- `make migrate-up` / `make migrate-down` apply migrations using `DB_DSN`.

For container work, use `./build.sh` and `./run.sh`.

## Coding Style & Naming Conventions
Follow standard Go style: run `gofmt` on every change, keep imports organized, and prefer small packages with clear responsibilities. Use `CamelCase` for exported identifiers and `camelCase` for unexported names. Keep HTTP DTOs and validation tags close to handlers or services, and place SQL changes in `db/query/users.sql` before regenerating `sqlc` output.

## Testing Guidelines
Tests live next to the code they cover as `*_test.go`. The repo uses Go’s `testing` package with `testify/require`; current tests also use `t.Parallel()` where safe. Prefer focused unit tests around handler and service behavior, especially validation, error mapping, and partial update flows. Run `make test` before opening a PR.

## Commit & Pull Request Guidelines
Recent commits use short, imperative subjects such as `Refine user update API semantics` and `Add explicit local CORS support`. Keep subjects concise and action-oriented. PRs should explain the behavior change, note any required env or migration updates, and list the verification commands you ran. If API shapes or comments change, include regenerated Swagger files in the same PR.
