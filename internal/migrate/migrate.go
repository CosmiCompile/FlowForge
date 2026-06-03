package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/CosmiCompile/FlowForge/internal/migrations"
)

func ApplyAll(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("nil db")
	}
	// Ensure migration table exists first.
	if _, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version INTEGER PRIMARY KEY,
	applied_at TEXT NOT NULL
);
`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range migrations.All {
		applied, err := isApplied(ctx, db, m.Version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		if _, err := db.ExecContext(ctx, m.UpSQL); err != nil {
			return fmt.Errorf("apply migration %d: %w", m.Version, err)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, m.Version, time.Now().UTC().Format(time.RFC3339)); err != nil {
			return fmt.Errorf("record migration %d: %w", m.Version, err)
		}
	}
	return nil
}

func isApplied(ctx context.Context, db *sql.DB, version int) (bool, error) {
	row := db.QueryRowContext(ctx, `SELECT 1 FROM schema_migrations WHERE version = ?`, version)
	var one int
	switch err := row.Scan(&one); {
	case err == nil:
		return true, nil
	case err == sql.ErrNoRows:
		return false, nil
	default:
		return false, fmt.Errorf("check migration %d: %w", version, err)
	}
}
