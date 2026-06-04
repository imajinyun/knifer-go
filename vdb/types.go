package vdb

import dbimpl "github.com/imajinyun/go-knifer/internal/db"

type (
	Dialect           = dbimpl.Dialect
	Wrapper           = dbimpl.Wrapper
	Entity            = dbimpl.Entity
	Direction         = dbimpl.Direction
	Order             = dbimpl.Order
	Page              = dbimpl.Page
	PageResult[T any] = dbimpl.PageResult[T]
	Query             = dbimpl.Query
	LogicalOperator   = dbimpl.LogicalOperator
	Condition         = dbimpl.Condition
	NamedSQL          = dbimpl.NamedSQL
	SQLBuilder        = dbimpl.SQLBuilder
	Options           = dbimpl.Options
	Option            = dbimpl.Option
	DB                = dbimpl.DB
	DBError           = dbimpl.DBError
	Session           = dbimpl.Session
	Column            = dbimpl.Column
	Table             = dbimpl.Table
)

const (
	DialectUnknown    = dbimpl.DialectUnknown
	DialectQuestion   = dbimpl.DialectQuestion
	DialectMySQL      = dbimpl.DialectMySQL
	DialectSQLite     = dbimpl.DialectSQLite
	DialectPostgres   = dbimpl.DialectPostgres
	DialectSQLServer  = dbimpl.DialectSQLServer
	DialectOracle     = dbimpl.DialectOracle
	DialectClickHouse = dbimpl.DialectClickHouse

	AscDirection  = dbimpl.AscDirection
	DescDirection = dbimpl.DescDirection

	And = dbimpl.And
	Or  = dbimpl.Or
)
