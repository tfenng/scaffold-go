package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	migrations "scaffold-api/db/migrations"
)

// schemaMigrationsDDL bootstraps the bookkeeping table used by the embedded
// migrator. It is created programmatically so sqlc does not generate a model
// for it from the schema directory.
const schemaMigrationsDDL = `CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
)`

// MigrateUp applies every pending *.up.sql migration in lexical order.
func MigrateUp(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schemaMigrationsDDL); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied, err := appliedVersions(ctx, db)
	if err != nil {
		return err
	}

	names, err := migrationFiles(".up.sql")
	if err != nil {
		return err
	}

	for _, name := range names {
		version := strings.TrimSuffix(name, ".up.sql")
		if applied[version] {
			continue
		}

		body, err := fs.ReadFile(migrations.FS, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		if err := applyMigration(ctx, db, version, string(body)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}

	return nil
}

// MigrateDown reverts the most recently applied migration.
func MigrateDown(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schemaMigrationsDDL); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	versions, err := appliedVersionList(ctx, db)
	if err != nil {
		return err
	}
	if len(versions) == 0 {
		return nil
	}

	version := versions[len(versions)-1]
	name := version + ".down.sql"

	body, err := fs.ReadFile(migrations.FS, name)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", name, err)
	}

	if err := revertMigration(ctx, db, version, string(body)); err != nil {
		return fmt.Errorf("revert migration %s: %w", name, err)
	}

	return nil
}

func applyMigration(ctx context.Context, db *sql.DB, version, body string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, body); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
		return err
	}

	return tx.Commit()
}

func revertMigration(ctx context.Context, db *sql.DB, version, body string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, body); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version = ?`, version); err != nil {
		return err
	}

	return tx.Commit()
}

func appliedVersions(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	versions, err := appliedVersionList(ctx, db)
	if err != nil {
		return nil, err
	}

	applied := make(map[string]bool, len(versions))
	for _, version := range versions {
		applied[version] = true
	}
	return applied, nil
}

func appliedVersionList(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return versions, nil
}

func migrationFiles(suffix string) ([]string, error) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), suffix) {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	return names, nil
}
