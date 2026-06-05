package vdb

import (
	"context"
	"database/sql"

	dbimpl "github.com/imajinyun/go-knifer/internal/db"
)

// Open opens a database using database/sql and applies pool/dialect options.
func Open(driverName, dataSourceName string, opts ...Option) (*DB, error) {
	return dbimpl.Open(driverName, dataSourceName, opts...)
}

// Use wraps an existing *sql.DB and applies pool/dialect options.
func Use(sqlDB *sql.DB, opts ...Option) *DB { return dbimpl.Use(sqlDB, opts...) }

// Exec executes SQL against db.
func Exec(ctx context.Context, db *DB, query string, args ...any) (sql.Result, error) {
	return db.Exec(ctx, query, args...)
}
