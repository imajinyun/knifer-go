package vdb

import dbimpl "github.com/imajinyun/go-knifer/internal/db"

func NormalizeDialect(name string) Dialect { return dbimpl.NormalizeDialect(name) }

func NewWrapper(prefix, suffix string) Wrapper { return dbimpl.NewWrapper(prefix, suffix) }

func WrapperForDialect(d Dialect) Wrapper { return dbimpl.WrapperForDialect(d) }
