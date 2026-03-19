#!/usr/bin/env bash

set -euo pipefail

IMAGE_NAME="${IMAGE_NAME:-scaffold-api}"
IMAGE_TAG="${1:-${IMAGE_TAG:-dev}}"
APP_ENV="${APP_ENV:-dev}"
ENV_FILE="${ENV_FILE:-.env.${APP_ENV}}"

set -a
source "./${ENV_FILE}"
set +a

DOCKER_DB_DSN="${DB_DSN:?DB_DSN is required in ${ENV_FILE}}"
DOCKER_DB_DSN="${DOCKER_DB_DSN/@localhost:/@host.docker.internal:}"
DOCKER_DB_DSN="${DOCKER_DB_DSN/@127.0.0.1:/@host.docker.internal:}"

docker run --rm -it \
  --name scaffold-api-dev \
  --add-host=host.docker.internal:host-gateway \
  -p 8080:8080 \
  -e APP_ENV="${APP_ENV}" \
  -e HTTP_HOST="${HTTP_HOST:-0.0.0.0}" \
  -e HTTP_PORT="${HTTP_PORT:-8080}" \
  -e DB_DSN="${DOCKER_DB_DSN}" \
  -e LOG_LEVEL="${LOG_LEVEL:-info}" \
  -e CORS_ALLOW_ORIGINS="${CORS_ALLOW_ORIGINS:-http://localhost:3000,http://127.0.0.1:3000}" \
  "${IMAGE_NAME}:${IMAGE_TAG}"
