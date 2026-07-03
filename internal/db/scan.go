package db

import (
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"strings"
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
		if !canConvertWithoutOverflow(src, dst.Type()) {
			return fmt.Errorf("cannot assign %T value %v to %s without overflow", value, value, dst.Type())
		}
		dst.Set(src.Convert(dst.Type()))
		return nil
	}
	if dst.Kind() == reflect.String {
		dst.SetString(fmt.Sprint(value))
		return nil
	}
	return fmt.Errorf("cannot assign %T to %s", value, dst.Type())
}

func canConvertWithoutOverflow(src reflect.Value, dstType reflect.Type) bool {
	switch src.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value := src.Int()
		switch dstType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			min, max := signedIntegerBounds(typeBits(dstType))
			return value >= min && value <= max
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if value < 0 {
				return false
			}
			return uint64(value) <= unsignedIntegerMax(typeBits(dstType))
		case reflect.Float32:
			return math.Abs(float64(value)) <= math.MaxFloat32
		case reflect.Float64:
			return true
		default:
			return true
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		value := src.Uint()
		switch dstType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			_, max := signedIntegerBounds(typeBits(dstType))
			if max < 0 {
				return false
			}
			return value <= uint64(max)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return value <= unsignedIntegerMax(typeBits(dstType))
		case reflect.Float32:
			return true
		case reflect.Float64:
			return true
		default:
			return true
		}
	case reflect.Float32, reflect.Float64:
		value := src.Float()
		switch dstType.Kind() {
		case reflect.Float32:
			return math.Abs(value) <= math.MaxFloat32
		case reflect.Float64:
			return true
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if math.Trunc(value) != value {
				return false
			}
			min, max := signedFloatBounds(typeBits(dstType))
			return value >= min && value <= max
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			if math.Trunc(value) != value || value < 0 {
				return false
			}
			return value <= unsignedFloatMax(typeBits(dstType))
		default:
			return true
		}
	default:
		return true
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

func signedIntegerBounds(bits int) (int64, int64) {
	if bits >= 64 {
		const maxInt64 = int64(1<<63 - 1)
		return -maxInt64 - 1, maxInt64
	}
	max := int64(1)<<(bits-1) - 1
	return -max - 1, max
}

func unsignedIntegerMax(bits int) uint64 {
	if bits >= 64 {
		return ^uint64(0)
	}
	return uint64(1)<<bits - 1
}

func signedFloatBounds(bits int) (float64, float64) {
	limit := math.Ldexp(1, bits-1)
	return -limit, math.Nextafter(limit, 0)
}

func unsignedFloatMax(bits int) float64 {
	return math.Nextafter(math.Ldexp(1, bits), 0)
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
