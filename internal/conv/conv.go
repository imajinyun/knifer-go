// Package conv provides permissive type conversion helpers.
package conv

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// This file provides permissive conversion helpers aligned with the utility toolkit-core Convert.
// Failed conversions return zero values or caller-provided defaults instead of panicking.

// ToString converts any value to a string; nil becomes an empty string.
func ToString(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	case fmt.Stringer:
		return x.String()
	case error:
		return x.Error()
	case bool:
		return strconv.FormatBool(x)
	case float32:
		return strconv.FormatFloat(float64(x), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	}
	return fmt.Sprint(v)
}

// ToStringDefault converts a value to a string and returns def when the value is nil.
func ToStringDefault(v any, def string) string {
	if v == nil {
		return def
	}
	return ToString(v)
}

// ToInt converts a value to int and returns 0 on failure.
func ToInt(v any) int { return ToIntDefault(v, 0) }

// ToIntDefault converts a value to int and returns def on failure.
func ToIntDefault(v any, def int) int {
	i, ok := toInt64(v)
	if !ok {
		return def
	}
	return int(i)
}

// ToInt64 converts a value to int64 and returns 0 on failure.
func ToInt64(v any) int64 { return ToInt64Default(v, 0) }

// ToInt64Default converts a value to int64 and returns def on failure.
func ToInt64Default(v any, def int64) int64 {
	i, ok := toInt64(v)
	if !ok {
		return def
	}
	return i
}

// ToFloat64 converts a value to float64 and returns 0 on failure.
func ToFloat64(v any) float64 { return ToFloat64Default(v, 0) }

// ToFloat64Default converts a value to float64 and returns def on failure.
func ToFloat64Default(v any, def float64) float64 {
	f, ok := toFloat64(v)
	if !ok {
		return def
	}
	return f
}

// ToBool converts a value to bool and returns false on failure.
func ToBool(v any) bool { return ToBoolDefault(v, false) }

// ToBoolDefault converts a value to bool and returns def on failure.
func ToBoolDefault(v any, def bool) bool {
	if v == nil {
		return def
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		s := strings.ToLower(strings.TrimSpace(x))
		switch s {
		case "true", "yes", "y", "ok", "1", "on":
			return true
		case "false", "no", "n", "0", "off":
			return false
		}
		return def
	}
	if i, ok := toInt64(v); ok {
		return i != 0
	}
	return def
}

// ToBytes converts a value to bytes; strings are converted directly and other values use ToString.
func ToBytes(v any) []byte {
	switch x := v.(type) {
	case nil:
		return nil
	case []byte:
		return x
	case string:
		return []byte(x)
	}
	return []byte(ToString(v))
}

func toInt64(v any) (int64, bool) {
	if v == nil {
		return 0, false
	}
	switch x := v.(type) {
	case int:
		return int64(x), true
	case int8:
		return int64(x), true
	case int16:
		return int64(x), true
	case int32:
		return int64(x), true
	case int64:
		return x, true
	case uint:
		return int64(x), true
	case uint8:
		return int64(x), true
	case uint16:
		return int64(x), true
	case uint32:
		return int64(x), true
	case uint64:
		return int64(x), true
	case float32:
		return int64(x), true
	case float64:
		return int64(x), true
	case bool:
		if x {
			return 1, true
		}
		return 0, true
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0, false
		}
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i, true
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return int64(f), true
		}
		return 0, false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return int64(rv.Float()), true
	}
	return 0, false
}

func toFloat64(v any) (float64, bool) {
	if v == nil {
		return 0, false
	}
	switch x := v.(type) {
	case float32:
		return float64(x), true
	case float64:
		return x, true
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return f, true
		}
		return 0, false
	}
	if i, ok := toInt64(v); ok {
		return float64(i), true
	}
	return 0, false
}
