package db

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// Dialect controls placeholder and pagination syntax.
type Dialect string

const (
	DialectUnknown    Dialect = ""
	DialectQuestion   Dialect = "question"
	DialectMySQL      Dialect = "mysql"
	DialectSQLite     Dialect = "sqlite"
	DialectPostgres   Dialect = "postgres"
	DialectSQLServer  Dialect = "sqlserver"
	DialectOracle     Dialect = "oracle"
	DialectClickHouse Dialect = "clickhouse"
)

// NormalizeDialect maps common driver names to a Dialect.
func NormalizeDialect(name string) Dialect {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "mysql", "mariadb":
		return DialectMySQL
	case "sqlite", "sqlite3", "moderncsqlite":
		return DialectSQLite
	case "postgres", "postgresql", "pgx":
		return DialectPostgres
	case "sqlserver", "mssql":
		return DialectSQLServer
	case "oracle", "godror", "oci8":
		return DialectOracle
	case "clickhouse":
		return DialectClickHouse
	default:
		return DialectQuestion
	}
}

func (d Dialect) placeholder(pos int) string {
	if pos < 1 {
		pos = 1
	}
	switch d {
	case DialectPostgres:
		return fmt.Sprintf("$%d", pos)
	case DialectSQLServer:
		return fmt.Sprintf("@p%d", pos)
	case DialectOracle:
		return fmt.Sprintf(":%d", pos)
	default:
		return "?"
	}
}

// Wrapper wraps SQL identifiers.
type Wrapper struct {
	Prefix string
	Suffix string
}

// NewWrapper returns an identifier wrapper. When suffix is empty, prefix is used
// for both sides.
func NewWrapper(prefix, suffix string) Wrapper {
	if suffix == "" {
		suffix = prefix
	}
	return Wrapper{Prefix: prefix, Suffix: suffix}
}

// WrapperForDialect returns the conventional identifier wrapper for dialect.
func WrapperForDialect(d Dialect) Wrapper {
	switch d {
	case DialectMySQL, DialectSQLite:
		return NewWrapper("`", "`")
	case DialectPostgres, DialectOracle, DialectClickHouse:
		return NewWrapper("\"", "\"")
	case DialectSQLServer:
		return NewWrapper("[", "]")
	default:
		return Wrapper{}
	}
}

// Wrap wraps a SQL identifier unless it already looks like an expression.
func (w Wrapper) Wrap(name string) string {
	name = strings.TrimSpace(name)
	if name == "" || name == "*" || w.Prefix == "" {
		return name
	}
	upper := strings.ToUpper(name)
	if strings.ContainsAny(name, " ()\t\n\r") || strings.Contains(upper, " AS ") || strings.Contains(name, ".*") {
		return name
	}
	parts := strings.Split(name, ".")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "*" || strings.HasPrefix(part, w.Prefix) && strings.HasSuffix(part, w.Suffix) {
			parts[i] = part
			continue
		}
		parts[i] = w.Prefix + part + w.Suffix
	}
	return strings.Join(parts, ".")
}

// IsSafeIdentifier reports whether name is a plain SQL identifier path.
// It accepts identifiers like "users", "schema.users", "users.*", and already
// wrapped single parts. It deliberately rejects whitespace, comments,
// delimiters, and expressions; callers that need expressions should use raw SQL
// APIs with trusted constants.
func IsSafeIdentifier(name string) bool {
	name = strings.TrimSpace(name)
	if name == "" || strings.Contains(name, "..") {
		return false
	}
	if strings.IndexFunc(name, unicode.IsSpace) >= 0 {
		return false
	}
	if name == "*" {
		return true
	}
	for _, part := range strings.Split(name, ".") {
		part = strings.TrimSpace(part)
		if part == "" {
			return false
		}
		if part == "*" {
			continue
		}
		if isWrappedIdentifierPart(part) {
			part = part[1 : len(part)-1]
		}
		if !isBareIdentifierPart(part) {
			return false
		}
	}
	return true
}

func isWrappedIdentifierPart(part string) bool {
	if len(part) < 2 {
		return false
	}
	return part[0] == '`' && part[len(part)-1] == '`' ||
		part[0] == '"' && part[len(part)-1] == '"' ||
		part[0] == '[' && part[len(part)-1] == ']'
}

func isBareIdentifierPart(part string) bool {
	if part == "" {
		return false
	}
	for i, r := range part {
		if i == 0 {
			if r != '_' && !unicode.IsLetter(r) {
				return false
			}
			continue
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func validateIdentifier(name, context string) error {
	if !IsSafeIdentifier(name) {
		return invalidInputf("db: unsafe SQL identifier for %s: %q", context, name)
	}
	return nil
}

func validateIdentifierList(values []string, context string, allowStar bool) error {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if !allowStar && value == "*" {
			return invalidInputf("db: wildcard SQL identifier is not allowed for %s", context)
		}
		if err := validateIdentifier(value, context); err != nil {
			return err
		}
	}
	return nil
}

// Unwrap removes wrapper characters from an identifier.
func (w Wrapper) Unwrap(name string) string {
	name = strings.TrimSpace(name)
	if w.Prefix == "" || w.Suffix == "" {
		return name
	}
	parts := strings.Split(name, ".")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.TrimPrefix(part, w.Prefix)
		part = strings.TrimSuffix(part, w.Suffix)
		parts[i] = part
	}
	return strings.Join(parts, ".")
}

// Entity represents a table-bound record or query condition map.
type Entity struct {
	Table  string
	Values map[string]any
	Fields []string
}

// NewEntity creates an Entity for table.
func NewEntity(table string) Entity {
	return Entity{Table: table, Values: map[string]any{}}
}

// EntityFromMap creates an Entity from values.
func EntityFromMap(table string, values map[string]any) Entity {
	out := NewEntity(table)
	for k, v := range values {
		out.Values[k] = v
	}
	return out
}

// Set sets a field value.
func (e Entity) Set(field string, value any) Entity {
	if e.Values == nil {
		e.Values = map[string]any{}
	}
	e.Values[field] = value
	return e
}

// SetIfNotNil sets a field when value is not nil.
func (e Entity) SetIfNotNil(field string, value any) Entity {
	if value != nil {
		e = e.Set(field, value)
	}
	return e
}

// Select limits selected fields.
func (e Entity) Select(fields ...string) Entity {
	e.Fields = slices.Clone(fields)
	return e
}

// Filter keeps only selected fields in Values.
func (e Entity) Filter(fields ...string) Entity {
	keep := map[string]struct{}{}
	for _, field := range fields {
		keep[field] = struct{}{}
	}
	for field := range e.Values {
		if _, ok := keep[field]; !ok {
			delete(e.Values, field)
		}
	}
	return e
}

// Remove removes fields from Values.
func (e Entity) Remove(fields ...string) Entity {
	for _, field := range fields {
		delete(e.Values, field)
	}
	return e
}

func (e Entity) sortedKeys() []string {
	keys := make([]string, 0, len(e.Values))
	for key := range e.Values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

// Direction is an ORDER BY direction.
type Direction string

const (
	AscDirection  Direction = "ASC"
	DescDirection Direction = "DESC"
)

// Order represents one ORDER BY field.
type Order struct {
	Field     string
	Direction Direction
}

// Asc returns ascending order.
func Asc(field string) Order { return Order{Field: field, Direction: AscDirection} }

// Desc returns descending order.
func Desc(field string) Order { return Order{Field: field, Direction: DescDirection} }

func (o Order) build(w Wrapper) string {
	dir := strings.ToUpper(strings.TrimSpace(string(o.Direction)))
	if dir != string(DescDirection) {
		dir = string(AscDirection)
	}
	return w.Wrap(o.Field) + " " + dir
}

// Page describes a page request. Number is one-based.
type Page struct {
	Number int
	Size   int
	Orders []Order
}

// NewPage creates a page request.
func NewPage(number, size int, orders ...Order) Page {
	if number < 1 {
		number = 1
	}
	if size < 1 {
		size = 20
	}
	return Page{Number: number, Size: size, Orders: slices.Clone(orders)}
}

// Offset returns the zero-based row offset.
func (p Page) Offset() int {
	p = NewPage(p.Number, p.Size, p.Orders...)
	return (p.Number - 1) * p.Size
}

// Limit returns the page size.
func (p Page) Limit() int { return NewPage(p.Number, p.Size, p.Orders...).Size }

// PageResult contains paged records and counters.
type PageResult[T any] struct {
	Page      int
	PageSize  int
	Total     int64
	TotalPage int
	Items     []T
}

// NewPageResult creates a paged result.
func NewPageResult[T any](page Page, total int64, items []T) PageResult[T] {
	page = NewPage(page.Number, page.Size, page.Orders...)
	totalPage := 0
	if total > 0 {
		totalPage = int((total + int64(page.Size) - 1) / int64(page.Size))
	}
	return PageResult[T]{Page: page.Number, PageSize: page.Size, Total: total, TotalPage: totalPage, Items: items}
}

// IsFirst reports whether this is the first page.
func (p PageResult[T]) IsFirst() bool { return p.Page <= 1 }

// IsLast reports whether this is the last page.
func (p PageResult[T]) IsLast() bool { return p.TotalPage == 0 || p.Page >= p.TotalPage }

// Query describes a SELECT query.
type Query struct {
	Fields     []string
	Tables     []string
	Conditions []Condition
	Page       *Page
	Orders     []Order
}

// NewQuery creates a query for tables.
func NewQuery(tables ...string) Query { return Query{Tables: slices.Clone(tables)} }

// Select sets selected fields.
func (q Query) Select(fields ...string) Query { q.Fields = slices.Clone(fields); return q }

// Where appends conditions.
func (q Query) Where(conds ...Condition) Query {
	q.Conditions = append(q.Conditions, conds...)
	return q
}

// WithPage sets page.
func (q Query) WithPage(page Page) Query { q.Page = &page; return q }

// OrderBy appends orders.
func (q Query) OrderBy(orders ...Order) Query { q.Orders = append(q.Orders, orders...); return q }

// FirstTable returns the first table name.
func (q Query) FirstTable() string {
	if len(q.Tables) == 0 {
		return ""
	}
	return q.Tables[0]
}
