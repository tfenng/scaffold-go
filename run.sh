#!/bin/zsh
APP_DB_DSN='postgres://tony:tony@localhost:5432/tony?sslmode=disable' go run . serve
