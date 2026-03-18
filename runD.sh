#!/bin/zsh

docker run --rm -it --name scaffold-api-dev --add-host=host.docker.internal:host-gateway  -p 8080:8080  -e APP_DB_DSN='postgres://tony:tony@host.docker.internal:5432/tony?sslmode=disable'   scaffold-api:dev
