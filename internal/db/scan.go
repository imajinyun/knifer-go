package db

import (
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	refimpl "github.com/imajinyun/knifer-go/internal/ref"
)

// ScanRows scans all rows into Entity values.
func ScanRows(rows *sql.Rows) ([]Entity, error) {
	if rows == nil {
		return nil, invalidInputf("db: rows is nil")
	}
	defer func() { _ = rows.Close() }()
	cols, err := rows.Columns()
	if err != nil {
		return nil, wrapInternal("db: read columns", err)
	}
	out := make([]Entity, 0)
	for rows.Next() {
		entity, err := scanCurrentRow(rows, cols)
		if err != nil {
			return nil, err
		}
		out = append(out, entity)
	}
	if err := rows.Err(); err != nil {
		return nil, wrapInternal("db: iterate rows", err)
	}
	return out, nil
}

// ScanOne scans the first row into Entity.
func ScanOne(rows *sql.Rows) (Entity, bool, error) {
	if rows == nil {
		return Entity{}, false, invalidInputf("db: rows is nil")
	}
	defer func() { _ = rows.Close() }()
	cols, err := rows.Columns()
	if err != nil {
		return Entity{}, false, wrapInternal("db: read columns", err)
	}
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return Entity{}, false, wrapInternal("db: iterate rows", err)
		}
		return Entity{}, false, nil
	}
	entity, err := scanCurrentRow(rows, cols)
	if err != nil {
		return Entity{}, false, err
	}
	return entity, true, nil
}

func scanCurrentRow(rows *sql.Rows, cols []string) (Entity, error) {
	values := make([]any, len(cols))
	dest := make([]any, len(cols))
	for i := range values {
		dest[i] = &values[i]
	}
	if err := rows.Scan(dest...); err != nil {
		return Entity{}, wrapInternal("db: scan row", err)
	}
	entity := NewEntity("")
	for i, col := range cols {
		entity.Values[col] = normalizeDBValue(values[i])
	}
	return entity, nil
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
		return invalidInputf("db: dst is nil")
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
		return invalidInputf("db: dst must be a non-nil pointer")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return invalidInputf("db: dst must point to struct")
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
			return wrapInternal(fmt.Sprintf("db: set field %s", key), err)
		}
	}
	return nil
}

func setValue(dst reflect.Value, value any) error {
	if dst.CanAddr() {
		if scanner, ok := dst.Addr().Interface().(sql.Scanner); ok {
			return scanner.Scan(value)
		}
	}
	if value == nil {
		dst.Set(reflect.Zero(dst.Type()))
		return nil
	}
	if dst.Kind() == reflect.Pointer {
		elem := reflect.New(dst.Type().Elem())
		if err := setValue(elem.Elem(), value); err != nil {
			return err
		}
		dst.Set(elem)
		return nil
	}
	src := reflect.ValueOf(value)
	if src.Type().AssignableTo(dst.Type()) {
		dst.Set(src)
		return nil
	}
	if src.Type().ConvertibleTo(dst.Type()) {
		if err := rejectFractionalNumericAssignment(src, dst.Type()); err != nil {
			return err
		}
		converted, err := refimpl.SafeConvert(src, dst.Type())
		if err != nil {
			return fmt.Errorf("cannot assign %T value %v to %s without overflow: %w", value, value, dst.Type(), err)
		}
		dst.Set(converted)
		return nil
	}
	if err := setScalarFromText(dst, value); err == nil {
		return nil
	}
	if dst.Kind() == reflect.String {
		dst.SetString(fmt.Sprint(value))
		return nil
	}
	return fmt.Errorf("cannot assign %T to %s", value, dst.Type())
}

func rejectFractionalNumericAssignment(src reflect.Value, dstType reflect.Type) error {
	switch src.Kind() {
	case reflect.Float32, reflect.Float64:
	default:
		return nil
	}
	switch dstType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
	default:
		return nil
	}
	value := src.Float()
	if math.Trunc(value) != value {
		return fmt.Errorf("cannot assign fractional float %v to %s", value, dstType)
	}
	return nil
}

func setScalarFromText(dst reflect.Value, value any) error {
	var text string
	switch v := value.(type) {
	case string:
		text = strings.TrimSpace(v)
	case []byte:
		text = strings.TrimSpace(string(v))
	default:
		return fmt.Errorf("not text")
	}
	if text == "" {
		return fmt.Errorf("empty text")
	}
	switch dst.Kind() {
	case reflect.Bool:
		parsed, err := strconv.ParseBool(text)
		if err != nil {
			return err
		}
		dst.SetBool(parsed)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(text, 10, typeBits(dst.Type()))
		if err != nil {
			return err
		}
		dst.SetInt(parsed)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		parsed, err := strconv.ParseUint(text, 10, typeBits(dst.Type()))
		if err != nil {
			return err
		}
		dst.SetUint(parsed)
		return nil
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(text, typeBits(dst.Type()))
		if err != nil {
			return err
		}
		dst.SetFloat(parsed)
		return nil
	default:
		return fmt.Errorf("not scalar")
	}
}

func typeBits(t reflect.Type) int {
	bits := t.Bits()
	if bits == 0 {
		return nativeIntBits()
	}
	return bits
}

func nativeIntBits() int {
	return 32 << (^uint(0) >> 63)
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
