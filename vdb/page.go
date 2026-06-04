package vdb

import dbimpl "github.com/imajinyun/go-knifer/internal/db"

func Asc(field string) Order { return dbimpl.Asc(field) }

func Desc(field string) Order { return dbimpl.Desc(field) }

func NewPage(number, size int, orders ...Order) Page { return dbimpl.NewPage(number, size, orders...) }

func NewPageResult[T any](page Page, total int64, items []T) PageResult[T] {
	return dbimpl.NewPageResult(page, total, items)
}
