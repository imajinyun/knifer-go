package vdb

import (
	"context"
	"database/sql"
	"time"

	dbimpl "github.com/imajinyun/go-knifer/internal/db"
)

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

func NormalizeDialect(name string) Dialect     { return dbimpl.NormalizeDialect(name) }
func NewWrapper(prefix, suffix string) Wrapper { return dbimpl.NewWrapper(prefix, suffix) }
func WrapperForDialect(d Dialect) Wrapper      { return dbimpl.WrapperForDialect(d) }
func NewEntity(table string) Entity            { return dbimpl.NewEntity(table) }
func EntityFromMap(table string, values map[string]any) Entity {
	return dbimpl.EntityFromMap(table, values)
}
func Asc(field string) Order                         { return dbimpl.Asc(field) }
func Desc(field string) Order                        { return dbimpl.Desc(field) }
func NewPage(number, size int, orders ...Order) Page { return dbimpl.NewPage(number, size, orders...) }
func NewPageResult[T any](page Page, total int64, items []T) PageResult[T] {
	return dbimpl.NewPageResult(page, total, items)
}
func NewQuery(tables ...string) Query                { return dbimpl.NewQuery(tables...) }
func Eq(field string, value any) Condition           { return dbimpl.Eq(field, value) }
func Ne(field string, value any) Condition           { return dbimpl.Ne(field, value) }
func Gt(field string, value any) Condition           { return dbimpl.Gt(field, value) }
func Gte(field string, value any) Condition          { return dbimpl.Gte(field, value) }
func Lt(field string, value any) Condition           { return dbimpl.Lt(field, value) }
func Lte(field string, value any) Condition          { return dbimpl.Lte(field, value) }
func Like(field string, value any) Condition         { return dbimpl.Like(field, value) }
func In(field string, values ...any) Condition       { return dbimpl.In(field, values...) }
func Between(field string, begin, end any) Condition { return dbimpl.Between(field, begin, end) }
func IsNull(field string) Condition                  { return dbimpl.IsNull(field) }
func IsNotNull(field string) Condition               { return dbimpl.IsNotNull(field) }
func OrWith(c Condition) Condition                   { return dbimpl.OrWith(c) }
func AndGroup(conds ...Condition) Condition          { return dbimpl.AndGroup(conds...) }
func OrGroup(conds ...Condition) Condition           { return dbimpl.OrGroup(conds...) }
func ConditionsFromEntity(e Entity) []Condition      { return dbimpl.ConditionsFromEntity(e) }
func BuildConditions(conds ...Condition) (string, []any, error) {
	return dbimpl.BuildConditions(conds...)
}
func BuildLikeValue(value any, mode string) string { return dbimpl.BuildLikeValue(value, mode) }
func ParseNamed(query string, args map[string]any, dialect Dialect) (NamedSQL, error) {
	return dbimpl.ParseNamed(query, args, dialect)
}
func NewBuilder(opts ...Option) *SQLBuilder      { return dbimpl.NewBuilder(opts...) }
func Raw(sql string, args ...any) *SQLBuilder    { return dbimpl.Raw(sql, args...) }
func Select(fields ...string) *SQLBuilder        { return dbimpl.Select(fields...) }
func Insert(e Entity) *SQLBuilder                { return dbimpl.Insert(e) }
func Update(e Entity) *SQLBuilder                { return dbimpl.Update(e) }
func Delete(table string) *SQLBuilder            { return dbimpl.Delete(table) }
func RemoveOuterOrderBy(sql string) string       { return dbimpl.RemoveOuterOrderBy(sql) }
func IsInClause(sql string) bool                 { return dbimpl.IsInClause(sql) }
func NewOptions() Options                        { return dbimpl.NewOptions() }
func WithDialect(d Dialect) Option               { return dbimpl.WithDialect(d) }
func WithWrapper(w Wrapper) Option               { return dbimpl.WithWrapper(w) }
func WithMaxOpenConns(n int) Option              { return dbimpl.WithMaxOpenConns(n) }
func WithMaxIdleConns(n int) Option              { return dbimpl.WithMaxIdleConns(n) }
func WithConnMaxLifetime(d time.Duration) Option { return dbimpl.WithConnMaxLifetime(d) }
func WithConnMaxIdleTime(d time.Duration) Option { return dbimpl.WithConnMaxIdleTime(d) }
func Open(driverName, dataSourceName string, opts ...Option) (*DB, error) {
	return dbimpl.Open(driverName, dataSourceName, opts...)
}
func Use(sqlDB *sql.DB, opts ...Option) *DB        { return dbimpl.Use(sqlDB, opts...) }
func ScanRows(rows *sql.Rows) ([]Entity, error)    { return dbimpl.ScanRows(rows) }
func ScanOne(rows *sql.Rows) (Entity, bool, error) { return dbimpl.ScanOne(rows) }
func AssignEntity(entity Entity, dst any) error    { return dbimpl.AssignEntity(entity, dst) }

func Exec(ctx context.Context, db *DB, query string, args ...any) (sql.Result, error) {
	return db.Exec(ctx, query, args...)
}
