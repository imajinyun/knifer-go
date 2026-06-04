package vdb

import dbimpl "github.com/imajinyun/go-knifer/internal/db"

func ParseNamed(query string, args map[string]any, dialect Dialect) (NamedSQL, error) {
	return dbimpl.ParseNamed(query, args, dialect)
}
