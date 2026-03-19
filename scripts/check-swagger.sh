#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "${ROOT_DIR}"

GENERATED_FILES=(
  "docs/docs.go"
  "docs/swagger.json"
  "docs/swagger.yaml"
)

if ! command -v swag >/dev/null 2>&1; then
  echo "swag is required but was not found in PATH." >&2
  echo "Install it with: go install github.com/swaggo/swag/cmd/swag@v1.16.4" >&2
  exit 1
fi

echo "Regenerating Swagger artifacts..."
go generate ./...

echo "Checking generated Swagger artifacts..."
if ! git diff --exit-code -- "${GENERATED_FILES[@]}"; then
  echo >&2
  echo "Swagger artifacts are out of date." >&2
  echo "Run 'go generate ./...' and commit the updated files:" >&2
  printf '  - %s\n' "${GENERATED_FILES[@]}" >&2
  exit 1
fi

echo "Swagger artifacts are up to date."
