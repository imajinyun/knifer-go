package vdb

import (
	"time"

	dbimpl "github.com/imajinyun/go-knifer/internal/db"
)

func NewOptions() Options { return dbimpl.NewOptions() }

func WithDialect(d Dialect) Option { return dbimpl.WithDialect(d) }

func WithWrapper(w Wrapper) Option { return dbimpl.WithWrapper(w) }

func WithMaxOpenConns(n int) Option { return dbimpl.WithMaxOpenConns(n) }

func WithMaxIdleConns(n int) Option { return dbimpl.WithMaxIdleConns(n) }

func WithConnMaxLifetime(d time.Duration) Option { return dbimpl.WithConnMaxLifetime(d) }

func WithConnMaxIdleTime(d time.Duration) Option { return dbimpl.WithConnMaxIdleTime(d) }
