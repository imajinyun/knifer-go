package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// ScanRows scans all rows into Entity values.
func ScanRows(rows *sql.Rows) ([]Entity, error) {
	if rows == nil {
		return nil, fmt.Errorf("db: rows is nil")
	}
	defer func() { _ = rows.Close() }()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	out := make([]Entity, 0)
	for rows.Next() {
		values := make([]any, len(cols))
		dest := make([]any, len(cols))
		for i := range values {
			dest[i] = &values[i]
		}
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		entity := NewEntity("")
		for i, col := range cols {
			entity.Values[col] = normalizeDBValue(values[i])
		}
		out = append(out, entity)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ScanOne scans the first row into Entity.
func ScanOne(rows *sql.Rows) (Entity, bool, error) {
	items, err := ScanRows(rows)
	if err != nil {
		return Entity{}, false, err
	}
	if len(items) == 0 {
		return Entity{}, false, nil
	}
	return items[0], true, nil
}

func normalizeDBValue(v any) any {
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return v
}

// AssignEntity copies entity values into dst struct or map.
func AssignEntity(entity Entity, dst any) error {
	if dst == nil {
		return fmt.Errorf("db: dst is nil")
	}
	if m, ok := dst.(*map[string]any); ok {
		if *m == nil {
			*m = map[string]any{}
		}
		for k, v := range entity.Values {
			(*m)[k] = v
		}
		return nil
	}
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("db: dst must be a non-nil pointer")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("db: dst must point to struct")
	}
	rt := rv.Type()
	index := map[string]reflect.Value{}
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.PkgPath != "" {
			continue
		}
		names := []string{field.Name, strings.ToLower(field.Name), toSnake(field.Name)}
		for _, tag := range []string{"db", "bean", "json"} {
			if value := field.Tag.Get(tag); value != "" {
				name := strings.Split(value, ",")[0]
				if name != "" && name != "-" {
					names = append(names, name)
				}
			}
		}
		for _, name := range names {
			index[strings.ToLower(name)] = rv.Field(i)
		}
	}
	for key, value := range entity.Values {
		field, ok := index[strings.ToLower(key)]
		if !ok || !field.CanSet() {
			continue
		}
		if err := setValue(field, value); err != nil {
			return fmt.Errorf("db: set field %s: %w", key, err)
		}
	}
	return nil
}

func setValue(dst reflect.Value, value any) error {
	if value == nil {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}
	src := reflect.ValueOf(value)
	if src.Type().AssignableTo(dst.Type()) {
		dst.Set(src)
		return nil
	}
	if src.Type().ConvertibleTo(dst.Type()) {
		dst.Set(src.Convert(dst.Type()))
		return nil
	}
	if dst.Kind() == reflect.String {
		dst.SetString(fmt.Sprint(value))
		return nil
	}
	return fmt.Errorf("cannot assign %T to %s", value, dst.Type())
}

func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}
