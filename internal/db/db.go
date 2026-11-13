package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"scaffold-api/internal/config"

	turso "turso.tech/database/tursogo"
)

// Open opens the embedded Turso (SQLite-compatible) database.
// The DSN is a file path, or ":memory:" for an ephemeral database.
func Open(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	dsn := strings.TrimSpace(cfg.DBPath)
	if dsn == "" {
		return nil, fmt.Errorf("db path is required")
	}

	if dsn != ":memory:" {
		if dir := filepath.Dir(dsn); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("create database directory: %w", err)
			}
		}
	}

	connector, err := turso.NewConnector(dsn, turso.WithBusyTimeout(5000))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db := sql.OpenDB(connector)
	// SQLite (and Turso) is a single-writer database. Keeping a single
	// connection avoids spurious "database is busy" errors and keeps
	// transaction ordering deterministic.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
