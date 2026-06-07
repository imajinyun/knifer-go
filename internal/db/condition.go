package db

import (
	"fmt"
	"reflect"
	"strings"
)

// LogicalOperator links conditions.
type LogicalOperator string

const (
	And LogicalOperator = "AND"
	Or  LogicalOperator = "OR"
)

// Condition represents a SQL predicate.
type Condition struct {
	Field  string
	Op     string
	Value  any
	Second any
	Link   LogicalOperator
	Group  []Condition
}

// Eq creates an equality condition.
func Eq(field string, value any) Condition {
	return Condition{Field: field, Op: "=", Value: value, Link: And}
}

// Ne creates a not-equal condition.
func Ne(field string, value any) Condition {
	return Condition{Field: field, Op: "<>", Value: value, Link: And}
}

// Gt creates a greater-than condition.
func Gt(field string, value any) Condition {
	return Condition{Field: field, Op: ">", Value: value, Link: And}
}

// Gte creates a greater-than-or-equal condition.
func Gte(field string, value any) Condition {
	return Condition{Field: field, Op: ">=", Value: value, Link: And}
}

// Lt creates a less-than condition.
func Lt(field string, value any) Condition {
	return Condition{Field: field, Op: "<", Value: value, Link: And}
}

// Lte creates a less-than-or-equal condition.
func Lte(field string, value any) Condition {
	return Condition{Field: field, Op: "<=", Value: value, Link: And}
}

// Like creates a LIKE condition.
func Like(field string, value any) Condition {
	return Condition{Field: field, Op: "LIKE", Value: value, Link: And}
}

// In creates an IN condition.
func In(field string, values ...any) Condition {
	return Condition{Field: field, Op: "IN", Value: values, Link: And}
}

// Between creates a BETWEEN condition.
func Between(field string, begin, end any) Condition {
	return Condition{Field: field, Op: "BETWEEN", Value: begin, Second: end, Link: And}
}

// IsNull creates an IS NULL condition.
func IsNull(field string) Condition { return Condition{Field: field, Op: "IS NULL", Link: And} }

// IsNotNull creates an IS NOT NULL condition.
func IsNotNull(field string) Condition { return Condition{Field: field, Op: "IS NOT NULL", Link: And} }

// OrWith marks c as linked by OR.
func OrWith(c Condition) Condition { c.Link = Or; return c }

// AndGroup groups conditions with AND.
func AndGroup(conds ...Condition) Condition { return Condition{Link: And, Group: conds} }

// OrGroup groups conditions with OR.
func OrGroup(conds ...Condition) Condition { return Condition{Link: Or, Group: conds} }

func buildConditions(conds []Condition, d Dialect, w Wrapper, start int) (string, []any, int, error) {
	parts := make([]string, 0, len(conds))
	params := make([]any, 0, len(conds))
	pos := start
	for i, cond := range conds {
		sqlPart, values, next, err := buildCondition(cond, d, w, pos)
		if err != nil {
			return "", nil, pos, err
		}
		if sqlPart == "" {
			continue
		}
		link := strings.ToUpper(string(cond.Link))
		if link != string(Or) {
			link = string(And)
		}
		if i > 0 && len(parts) > 0 {
			parts = append(parts, link)
		}
		parts = append(parts, sqlPart)
		params = append(params, values...)
		pos = next
	}
	return strings.Join(parts, " "), params, pos, nil
}

func buildCondition(cond Condition, d Dialect, w Wrapper, start int) (string, []any, int, error) {
	if len(cond.Group) > 0 {
		part, params, next, err := buildConditions(cond.Group, d, w, start)
		if err != nil || part == "" {
			return part, params, next, err
		}
		return "(" + part + ")", params, next, nil
	}
	if strings.TrimSpace(cond.Field) == "" {
		return "", nil, start, nil
	}
	if err := validateIdentifier(cond.Field, "condition field"); err != nil {
		return "", nil, start, err
	}
	op := strings.ToUpper(strings.TrimSpace(cond.Op))
	if op == "" {
		op = "="
	}
	field := w.Wrap(cond.Field)
	switch op {
	case "IS NULL", "IS NOT NULL":
		return field + " " + op, nil, start, nil
	case "BETWEEN":
		return field + " BETWEEN " + d.placeholder(start) + " AND " + d.placeholder(start+1), []any{cond.Value, cond.Second}, start + 2, nil
	case "IN", "NOT IN":
		values := flattenValues(cond.Value)
		if len(values) == 0 {
			if op == "IN" {
				return "1 = 0", nil, start, nil
			}
			return "1 = 1", nil, start, nil
		}
		ph := make([]string, len(values))
		for i := range values {
			ph[i] = d.placeholder(start + i)
		}
		return field + " " + op + " (" + strings.Join(ph, ", ") + ")", values, start + len(values), nil
	default:
		return field + " " + op + " " + d.placeholder(start), []any{cond.Value}, start + 1, nil
	}
}

func flattenValues(v any) []any {
	if values, ok := v.([]any); ok {
		return values
	}
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil
	}
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return []any{v}
	}
	out := make([]any, 0, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		out = append(out, rv.Index(i).Interface())
	}
	return out
}

// ConditionsFromEntity creates equality conditions from entity values.
func ConditionsFromEntity(e Entity) []Condition {
	keys := e.sortedKeys()
	conds := make([]Condition, 0, len(keys))
	for _, key := range keys {
		conds = append(conds, Eq(key, e.Values[key]))
	}
	return conds
}

// BuildConditions builds a WHERE fragment without the WHERE keyword.
func BuildConditions(conds ...Condition) (string, []any, error) {
	part, params, _, err := buildConditions(conds, DialectQuestion, Wrapper{}, 1)
	return part, params, err
}

// BuildLikeValue returns a LIKE value according to mode: contains, prefix, suffix, or exact.
func BuildLikeValue(value any, mode string) string {
	s := fmt.Sprint(value)
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "prefix", "start", "left":
		return s + "%"
	case "suffix", "end", "right":
		return "%" + s
	case "exact", "none":
		return s
	default:
		return "%" + s + "%"
	}
}
