package db

import (
	"fmt"
	"strings"
)

// SQLBuilder builds SQL and parameter lists.
type SQLBuilder struct {
	dialect Dialect
	wrapper Wrapper
	verb    string
	fields  []string
	tables  []string
	joins   []string
	sets    []string
	conds   []Condition
	groups  []string
	having  string
	orders  []Order
	page    *Page
	params  []any
	raw     []string
}

// NewBuilder creates a SQL builder.
func NewBuilder(opts ...Option) *SQLBuilder {
	cfg := applyOptions(opts...)
	return &SQLBuilder{dialect: cfg.Dialect, wrapper: cfg.Wrapper}
}

// Raw creates a builder from a raw SQL fragment.
func Raw(sql string, args ...any) *SQLBuilder {
	b := NewBuilder()
	b.raw = append(b.raw, sql)
	b.params = append(b.params, args...)
	return b
}

// Select starts a SELECT builder.
func Select(fields ...string) *SQLBuilder { return NewBuilder().Select(fields...) }

// Insert builds an INSERT statement for entity.
func Insert(e Entity) *SQLBuilder { return NewBuilder().Insert(e) }

// Update builds an UPDATE statement for entity.
func Update(e Entity) *SQLBuilder { return NewBuilder().Update(e) }

// Delete builds a DELETE statement for table.
func Delete(table string) *SQLBuilder { return NewBuilder().Delete(table) }

// Select starts or updates SELECT fields.
func (b *SQLBuilder) Select(fields ...string) *SQLBuilder {
	b.verb = "SELECT"
	b.fields = append(b.fields, fields...)
	return b
}

// From sets FROM tables.
func (b *SQLBuilder) From(tables ...string) *SQLBuilder {
	b.tables = append(b.tables, tables...)
	return b
}

// Where appends WHERE conditions.
func (b *SQLBuilder) Where(conds ...Condition) *SQLBuilder {
	b.conds = append(b.conds, conds...)
	return b
}

// GroupBy appends GROUP BY fields.
func (b *SQLBuilder) GroupBy(fields ...string) *SQLBuilder {
	b.groups = append(b.groups, fields...)
	return b
}

// Having sets HAVING expression.
func (b *SQLBuilder) Having(expr string) *SQLBuilder { b.having = expr; return b }

// OrderBy appends ORDER BY fields.
func (b *SQLBuilder) OrderBy(orders ...Order) *SQLBuilder {
	b.orders = append(b.orders, orders...)
	return b
}

// Join appends a raw JOIN fragment.
func (b *SQLBuilder) Join(join string) *SQLBuilder {
	b.joins = append(b.joins, strings.TrimSpace(join))
	return b
}

// Page applies pagination.
func (b *SQLBuilder) Page(page Page) *SQLBuilder { b.page = &page; return b }

// Append appends raw SQL.
func (b *SQLBuilder) Append(sql string, args ...any) *SQLBuilder {
	b.raw = append(b.raw, sql)
	b.params = append(b.params, args...)
	return b
}

// Insert starts an INSERT statement.
func (b *SQLBuilder) Insert(e Entity) *SQLBuilder {
	b.verb = "INSERT"
	b.tables = []string{e.Table}
	keys := e.sortedKeys()
	b.fields = keys
	b.params = b.params[:0]
	for _, key := range keys {
		b.params = append(b.params, e.Values[key])
	}
	return b
}

// Update starts an UPDATE statement.
func (b *SQLBuilder) Update(e Entity) *SQLBuilder {
	b.verb = "UPDATE"
	b.tables = []string{e.Table}
	b.sets = b.sets[:0]
	b.params = b.params[:0]
	keys := e.sortedKeys()
	for _, key := range keys {
		b.sets = append(b.sets, key)
		b.params = append(b.params, e.Values[key])
	}
	return b
}

// Delete starts a DELETE statement.
func (b *SQLBuilder) Delete(table string) *SQLBuilder {
	b.verb = "DELETE"
	b.tables = []string{table}
	return b
}

// Query builds from Query.
func (b *SQLBuilder) Query(q Query) *SQLBuilder {
	b.Select(q.Fields...).From(q.Tables...).Where(q.Conditions...).OrderBy(q.Orders...)
	if q.Page != nil {
		b.Page(*q.Page)
	}
	return b
}

// SQL returns built SQL and params.
func (b *SQLBuilder) SQL() (string, []any, error) {
	if len(b.raw) > 0 && b.verb == "" {
		return strings.Join(b.raw, " "), append([]any(nil), b.params...), nil
	}
	switch b.verb {
	case "SELECT":
		return b.selectSQL()
	case "INSERT":
		return b.insertSQL()
	case "UPDATE":
		return b.updateSQL()
	case "DELETE":
		return b.deleteSQL()
	default:
		return "", nil, fmt.Errorf("db: SQL verb is not set")
	}
}

func (b *SQLBuilder) selectSQL() (string, []any, error) {
	if len(b.tables) == 0 {
		return "", nil, fmt.Errorf("db: SELECT requires table")
	}
	fields := b.fields
	if len(fields) == 0 {
		fields = []string{"*"}
	}
	parts := []string{"SELECT", wrapList(fields, b.wrapper), "FROM", wrapList(b.tables, b.wrapper)}
	parts = append(parts, b.joins...)
	params := append([]any(nil), b.params...)
	if len(b.conds) > 0 {
		where, values, _, err := buildConditions(b.conds, b.dialect, b.wrapper, len(params)+1)
		if err != nil {
			return "", nil, err
		}
		if where != "" {
			parts = append(parts, "WHERE", where)
			params = append(params, values...)
		}
	}
	if len(b.groups) > 0 {
		parts = append(parts, "GROUP BY", wrapList(b.groups, b.wrapper))
	}
	if b.having != "" {
		parts = append(parts, "HAVING", b.having)
	}
	if len(b.orders) > 0 {
		parts = append(parts, "ORDER BY", buildOrders(b.orders, b.wrapper))
	}
	if b.page != nil {
		parts = append(parts, b.paginationSQL(*b.page))
	}
	return strings.Join(parts, " "), params, nil
}

func (b *SQLBuilder) insertSQL() (string, []any, error) {
	if len(b.tables) == 0 || b.tables[0] == "" || len(b.fields) == 0 {
		return "", nil, fmt.Errorf("db: INSERT requires table and values")
	}
	ph := make([]string, len(b.fields))
	for i := range ph {
		ph[i] = b.dialect.placeholder(i + 1)
	}
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", b.wrapper.Wrap(b.tables[0]), wrapList(b.fields, b.wrapper), strings.Join(ph, ", "))
	return sql, append([]any(nil), b.params...), nil
}

func (b *SQLBuilder) updateSQL() (string, []any, error) {
	if len(b.tables) == 0 || b.tables[0] == "" || len(b.sets) == 0 {
		return "", nil, fmt.Errorf("db: UPDATE requires table and values")
	}
	sets := make([]string, len(b.sets))
	for i, field := range b.sets {
		sets[i] = b.wrapper.Wrap(field) + " = " + b.dialect.placeholder(i+1)
	}
	parts := []string{"UPDATE", b.wrapper.Wrap(b.tables[0]), "SET", strings.Join(sets, ", ")}
	params := append([]any(nil), b.params...)
	if len(b.conds) > 0 {
		where, values, _, err := buildConditions(b.conds, b.dialect, b.wrapper, len(params)+1)
		if err != nil {
			return "", nil, err
		}
		if where != "" {
			parts = append(parts, "WHERE", where)
			params = append(params, values...)
		}
	}
	return strings.Join(parts, " "), params, nil
}

func (b *SQLBuilder) deleteSQL() (string, []any, error) {
	if len(b.tables) == 0 || b.tables[0] == "" {
		return "", nil, fmt.Errorf("db: DELETE requires table")
	}
	parts := []string{"DELETE FROM", b.wrapper.Wrap(b.tables[0])}
	params := append([]any(nil), b.params...)
	if len(b.conds) > 0 {
		where, values, _, err := buildConditions(b.conds, b.dialect, b.wrapper, len(params)+1)
		if err != nil {
			return "", nil, err
		}
		if where != "" {
			parts = append(parts, "WHERE", where)
			params = append(params, values...)
		}
	}
	return strings.Join(parts, " "), params, nil
}

func (b *SQLBuilder) paginationSQL(page Page) string {
	page = NewPage(page.Number, page.Size, page.Orders...)
	switch b.dialect {
	case DialectSQLServer:
		return fmt.Sprintf("OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", page.Offset(), page.Limit())
	default:
		return fmt.Sprintf("LIMIT %d OFFSET %d", page.Limit(), page.Offset())
	}
}

func wrapList(values []string, w Wrapper) string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, w.Wrap(value))
		}
	}
	return strings.Join(out, ", ")
}

func buildOrders(orders []Order, w Wrapper) string {
	out := make([]string, 0, len(orders))
	for _, order := range orders {
		if strings.TrimSpace(order.Field) != "" {
			out = append(out, order.build(w))
		}
	}
	return strings.Join(out, ", ")
}

// RemoveOuterOrderBy removes the last top-level ORDER BY clause from a SELECT statement.
func RemoveOuterOrderBy(sql string) string {
	upper := strings.ToUpper(sql)
	depth := 0
	inSingle := false
	inDouble := false
	last := -1
	for i := 0; i < len(sql); i++ {
		switch sql[i] {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '(':
			if !inSingle && !inDouble {
				depth++
			}
		case ')':
			if !inSingle && !inDouble && depth > 0 {
				depth--
			}
		}
		if depth == 0 && !inSingle && !inDouble && strings.HasPrefix(upper[i:], "ORDER BY") {
			last = i
		}
	}
	if last < 0 {
		return sql
	}
	return strings.TrimSpace(sql[:last])
}

// IsInClause reports whether sql contains a top-level IN clause token.
func IsInClause(sql string) bool { return strings.Contains(strings.ToUpper(sql), " IN ") }
