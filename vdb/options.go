package vdb

import (
	"time"

	dbimpl "github.com/imajinyun/go-knifer/internal/db"
)

// NewOptions returns default database helper options.
func NewOptions() Options { return dbimpl.NewOptions() }

// WithDialect sets SQL dialect behavior and selects its default identifier wrapper when no wrapper is set.
func WithDialect(d Dialect) Option { return dbimpl.WithDialect(d) }

// WithWrapper sets identifier wrapper behavior and overrides the dialect default wrapper.
func WithWrapper(w Wrapper) Option { return dbimpl.WithWrapper(w) }

// WithMaxOpenConns sets database/sql max open connections when opening or wrapping a DB.
func WithMaxOpenConns(n int) Option { return dbimpl.WithMaxOpenConns(n) }

// WithMaxIdleConns sets database/sql max idle connections when opening or wrapping a DB.
func WithMaxIdleConns(n int) Option { return dbimpl.WithMaxIdleConns(n) }

// WithConnMaxLifetime sets database/sql max connection lifetime when opening or wrapping a DB.
func WithConnMaxLifetime(d time.Duration) Option { return dbimpl.WithConnMaxLifetime(d) }

// WithConnMaxIdleTime sets database/sql max idle time when opening or wrapping a DB.
func WithConnMaxIdleTime(d time.Duration) Option { return dbimpl.WithConnMaxIdleTime(d) }
