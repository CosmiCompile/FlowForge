package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type Config struct {
	// For SQLite use e.g. file:flowforge.db?_pragma=busy_timeout(5000)
	DSN string
}

type DB struct {
	SQL *sql.DB
}

func Open(ctx context.Context, cfg Config) (*DB, error) {
	if cfg.DSN == "" {
		cfg.DSN = "file:flowforge.db?_pragma=busy_timeout(5000)"
	}
	sqlDB, err := sql.Open("sqlite", cfg.DSN)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1) // SQLite: keep simple for MVP
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("db ping: %w", err)
	}
	return &DB{SQL: sqlDB}, nil
}

func (d *DB) Close() error {
	if d == nil || d.SQL == nil {
		return nil
	}
	return d.SQL.Close()
}
