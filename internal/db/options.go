package db

import (
	"database/sql"
	"time"
)

// Option customizes database helpers.
type Option func(*Options)

// Options contains database helper settings.
type Options struct {
	Dialect         Dialect
	Wrapper         Wrapper
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// NewOptions returns default settings.
func NewOptions() Options {
	return Options{Dialect: DialectQuestion}
}

// WithDialect sets SQL dialect behavior.
func WithDialect(d Dialect) Option {
	return func(o *Options) {
		o.Dialect = d
		if o.Wrapper == (Wrapper{}) {
			o.Wrapper = WrapperForDialect(d)
		}
	}
}

// WithWrapper sets identifier wrapper behavior.
func WithWrapper(w Wrapper) Option { return func(o *Options) { o.Wrapper = w } }

// WithMaxOpenConns sets database/sql max open connections.
func WithMaxOpenConns(n int) Option { return func(o *Options) { o.MaxOpenConns = n } }

// WithMaxIdleConns sets database/sql max idle connections.
func WithMaxIdleConns(n int) Option { return func(o *Options) { o.MaxIdleConns = n } }

// WithConnMaxLifetime sets database/sql max connection lifetime.
func WithConnMaxLifetime(d time.Duration) Option { return func(o *Options) { o.ConnMaxLifetime = d } }

// WithConnMaxIdleTime sets database/sql max idle time.
func WithConnMaxIdleTime(d time.Duration) Option { return func(o *Options) { o.ConnMaxIdleTime = d } }

func applyOptions(opts ...Option) Options {
	cfg := NewOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

func applyPoolOptions(sqlDB *sql.DB, cfg Options) {
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}
}
