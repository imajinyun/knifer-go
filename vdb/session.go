package vdb

import (
	"context"
	"database/sql"

	dbimpl "github.com/imajinyun/go-knifer/internal/db"
)

func Open(driverName, dataSourceName string, opts ...Option) (*DB, error) {
	return dbimpl.Open(driverName, dataSourceName, opts...)
}

func Use(sqlDB *sql.DB, opts ...Option) *DB { return dbimpl.Use(sqlDB, opts...) }

func Exec(ctx context.Context, db *DB, query string, args ...any) (sql.Result, error) {
	return db.Exec(ctx, query, args...)
}
