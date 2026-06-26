package vdb

import dbimpl "github.com/imajinyun/knifer-go/internal/db"

func NewBuilder(opts ...Option) *SQLBuilder { return dbimpl.NewBuilder(opts...) }

func Raw(sql string, args ...any) *SQLBuilder { return dbimpl.Raw(sql, args...) }

func Select(fields ...string) *SQLBuilder { return dbimpl.Select(fields...) }

func Insert(e Entity) *SQLBuilder { return dbimpl.Insert(e) }

func Update(e Entity) *SQLBuilder { return dbimpl.Update(e) }

func Delete(table string) *SQLBuilder { return dbimpl.Delete(table) }

func RemoveOuterOrderBy(sql string) string { return dbimpl.RemoveOuterOrderBy(sql) }

func IsInClause(sql string) bool { return dbimpl.IsInClause(sql) }
