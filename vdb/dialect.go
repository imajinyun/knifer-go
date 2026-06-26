package vdb

import dbimpl "github.com/imajinyun/knifer-go/internal/db"

// NormalizeDialect maps common driver names to a SQL dialect.
func NormalizeDialect(name string) Dialect { return dbimpl.NormalizeDialect(name) }

// NewWrapper returns an identifier wrapper. When suffix is empty, prefix is used on both sides.
func NewWrapper(prefix, suffix string) Wrapper { return dbimpl.NewWrapper(prefix, suffix) }

// WrapperForDialect returns the conventional identifier wrapper for dialect.
func WrapperForDialect(d Dialect) Wrapper { return dbimpl.WrapperForDialect(d) }

// IsSafeIdentifier reports whether name is a plain SQL identifier path.
func IsSafeIdentifier(name string) bool { return dbimpl.IsSafeIdentifier(name) }
