// Package conv provides permissive type conversion helpers.
package conv

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	knifer "github.com/imajinyun/knifer-go"
)

// ErrInvalidConversion reports that a value cannot be converted to the requested scalar type.
var ErrInvalidConversion = errors.New("conv: invalid conversion")

type config struct {
	parseBool   func(string) (bool, error)
	parseInt    func(string, int, int) (int64, error)
	parseFloat  func(string, int) (float64, error)
	formatBool  func(bool) string
	formatFloat func(float64, byte, int, int) string
}

// Option customizes conversion helpers per call.
type Option func(*config)

// WithBoolParser sets the parser used for string-to-bool conversion.
func WithBoolParser(parser func(string) (bool, error)) Option {
	return func(c *config) {
		if parser != nil {
			c.parseBool = parser
		}
	}
}

// WithParseIntFunc sets the parser used for string-to-integer conversion.
func WithParseIntFunc(parser func(string, int, int) (int64, error)) Option {
	return func(c *config) {
		if parser != nil {
			c.parseInt = parser
		}
	}
}

// WithParseFloatFunc sets the parser used for string-to-float conversion.
func WithParseFloatFunc(parser func(string, int) (float64, error)) Option {
	return func(c *config) {
		if parser != nil {
			c.parseFloat = parser
		}
	}
}

// WithFormatBoolFunc sets the formatter used for bool-to-string conversion.
func WithFormatBoolFunc(formatter func(bool) string) Option {
	return func(c *config) {
		if formatter != nil {
			c.formatBool = formatter
		}
	}
}

// WithFormatFloatFunc sets the formatter used for float-to-string conversion.
func WithFormatFloatFunc(formatter func(float64, byte, int, int) string) Option {
	return func(c *config) {
		if formatter != nil {
			c.formatFloat = formatter
		}
	}
}

func applyOptions(opts []Option) config {
	cfg := config{
		parseBool:   defaultBoolParser,
		parseInt:    strconv.ParseInt,
		parseFloat:  strconv.ParseFloat,
		formatBool:  strconv.FormatBool,
		formatFloat: strconv.FormatFloat,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.parseBool == nil {
		cfg.parseBool = defaultBoolParser
	}
	if cfg.parseInt == nil {
		cfg.parseInt = strconv.ParseInt
	}
	if cfg.parseFloat == nil {
		cfg.parseFloat = strconv.ParseFloat
	}
	if cfg.formatBool == nil {
		cfg.formatBool = strconv.FormatBool
	}
	if cfg.formatFloat == nil {
		cfg.formatFloat = strconv.FormatFloat
	}
	return cfg
}

// This file provides permissive conversion helpers aligned with the utility toolkit-core Convert.
// Failed conversions return zero values or caller-provided defaults instead of panicking.

// ToString converts any value to a string; nil becomes an empty string.
func ToString(v any) string {
	return ToStringWithOptions(v)
}

// ToStringWithOptions converts any value to a string using per-call options.
func ToStringWithOptions(v any, opts ...Option) string {
	cfg := applyOptions(opts)
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
		return cfg.formatBool(x)
	case float32:
		return cfg.formatFloat(float64(x), 'f', -1, 32)
	case float64:
		return cfg.formatFloat(x, 'f', -1, 64)
	}
	return fmt.Sprint(v)
}

// ToStringDefault converts a value to a string and returns def when the value is nil.
func ToStringDefault(v any, def string) string {
	return ToStringDefaultWithOptions(v, def)
}

// ToStringDefaultWithOptions converts a value to a string using per-call options and returns def when nil.
func ToStringDefaultWithOptions(v any, def string, opts ...Option) string {
	if v == nil {
		return def
	}
	return ToStringWithOptions(v, opts...)
}

// ToInt converts a value to int and returns 0 on failure.
func ToInt(v any) int { return ToIntWithOptions(v) }

// ToIntWithOptions converts a value to int using per-call options and returns 0 on failure.
func ToIntWithOptions(v any, opts ...Option) int { return ToIntDefaultWithOptions(v, 0, opts...) }

// ToIntE converts a value to int and returns an error on failure.
func ToIntE(v any) (int, error) { return ToIntEWithOptions(v) }

// ToIntEWithOptions converts a value to int using per-call options and returns an error on failure.
func ToIntEWithOptions(v any, opts ...Option) (int, error) {
	cfg := applyOptions(opts)
	i, ok := toInt64Strict(v, cfg)
	if !ok || i < int64(math.MinInt) || i > int64(math.MaxInt) {
		return 0, invalidConversionError("int")
	}
	return int(i), nil
}

// ToIntDefault converts a value to int and returns def on failure.
func ToIntDefault(v any, def int) int {
	return ToIntDefaultWithOptions(v, def)
}

// ToIntDefaultWithOptions converts a value to int using per-call options and returns def on failure.
func ToIntDefaultWithOptions(v any, def int, opts ...Option) int {
	cfg := applyOptions(opts)
	i, ok := toInt64(v, cfg)
	if !ok {
		return def
	}
	return int(i)
}

// ToInt64 converts a value to int64 and returns 0 on failure.
func ToInt64(v any) int64 { return ToInt64WithOptions(v) }

// ToInt64WithOptions converts a value to int64 using per-call options and returns 0 on failure.
func ToInt64WithOptions(v any, opts ...Option) int64 { return ToInt64DefaultWithOptions(v, 0, opts...) }

// ToInt64E converts a value to int64 and returns an error on failure.
func ToInt64E(v any) (int64, error) { return ToInt64EWithOptions(v) }

// ToInt64EWithOptions converts a value to int64 using per-call options and returns an error on failure.
func ToInt64EWithOptions(v any, opts ...Option) (int64, error) {
	cfg := applyOptions(opts)
	i, ok := toInt64Strict(v, cfg)
	if !ok {
		return 0, invalidConversionError("int64")
	}
	return i, nil
}

// ToInt64Default converts a value to int64 and returns def on failure.
func ToInt64Default(v any, def int64) int64 {
	return ToInt64DefaultWithOptions(v, def)
}

// ToInt64DefaultWithOptions converts a value to int64 using per-call options and returns def on failure.
func ToInt64DefaultWithOptions(v any, def int64, opts ...Option) int64 {
	cfg := applyOptions(opts)
	i, ok := toInt64(v, cfg)
	if !ok {
		return def
	}
	return i
}

// ToFloat64 converts a value to float64 and returns 0 on failure.
func ToFloat64(v any) float64 { return ToFloat64WithOptions(v) }

// ToFloat64WithOptions converts a value to float64 using per-call options and returns 0 on failure.
func ToFloat64WithOptions(v any, opts ...Option) float64 {
	return ToFloat64DefaultWithOptions(v, 0, opts...)
}

// ToFloat64E converts a value to float64 and returns an error on failure.
func ToFloat64E(v any) (float64, error) { return ToFloat64EWithOptions(v) }

// ToFloat64EWithOptions converts a value to float64 using per-call options and returns an error on failure.
func ToFloat64EWithOptions(v any, opts ...Option) (float64, error) {
	cfg := applyOptions(opts)
	f, ok := toFloat64(v, cfg)
	if !ok {
		return 0, invalidConversionError("float64")
	}
	return f, nil
}

// ToFloat64Default converts a value to float64 and returns def on failure.
func ToFloat64Default(v any, def float64) float64 {
	return ToFloat64DefaultWithOptions(v, def)
}

// ToFloat64DefaultWithOptions converts a value to float64 using per-call options and returns def on failure.
func ToFloat64DefaultWithOptions(v any, def float64, opts ...Option) float64 {
	cfg := applyOptions(opts)
	f, ok := toFloat64(v, cfg)
	if !ok {
		return def
	}
	return f
}

// ToBool converts a value to bool and returns false on failure.
func ToBool(v any) bool { return ToBoolWithOptions(v) }

// ToBoolWithOptions converts a value to bool using per-call options and returns false on failure.
func ToBoolWithOptions(v any, opts ...Option) bool {
	return ToBoolDefaultWithOptions(v, false, opts...)
}

// ToBoolE converts a value to bool and returns an error on failure.
func ToBoolE(v any) (bool, error) { return ToBoolEWithOptions(v) }

// ToBoolEWithOptions converts a value to bool using per-call options and returns an error on failure.
func ToBoolEWithOptions(v any, opts ...Option) (bool, error) {
	cfg := applyOptions(opts)
	if v == nil {
		return false, invalidConversionError("bool")
	}
	switch x := v.(type) {
	case bool:
		return x, nil
	case string:
		b, err := cfg.parseBool(x)
		if err != nil {
			return false, invalidConversionError("bool")
		}
		return b, nil
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool(), nil
	case reflect.String:
		b, err := cfg.parseBool(rv.String())
		if err != nil {
			return false, invalidConversionError("bool")
		}
		return b, nil
	}
	if i, ok := toInt64(v, cfg); ok {
		return i != 0, nil
	}
	return false, invalidConversionError("bool")
}

// ToBoolDefault converts a value to bool and returns def on failure.
func ToBoolDefault(v any, def bool) bool {
	return ToBoolDefaultWithOptions(v, def)
}

// ToBoolDefaultWithOptions converts a value to bool using per-call options and returns def on failure.
func ToBoolDefaultWithOptions(v any, def bool, opts ...Option) bool {
	cfg := applyOptions(opts)
	if v == nil {
		return def
	}
	switch x := v.(type) {
	case bool:
		return x
	case string:
		b, err := cfg.parseBool(x)
		if err != nil {
			return def
		}
		return b
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		return rv.Bool()
	case reflect.String:
		b, err := cfg.parseBool(rv.String())
		if err != nil {
			return def
		}
		return b
	}
	if i, ok := toInt64(v, cfg); ok {
		return i != 0
	}
	return def
}

// ToBytes converts a value to bytes; strings are converted directly and other values use ToString.
func ToBytes(v any) []byte {
	return ToBytesWithOptions(v)
}

// ToBytesWithOptions converts a value to bytes using per-call options.
func ToBytesWithOptions(v any, opts ...Option) []byte {
	switch x := v.(type) {
	case nil:
		return nil
	case []byte:
		return x
	case string:
		return []byte(x)
	}
	return []byte(ToStringWithOptions(v, opts...))
}

func toInt64(v any, cfg config) (int64, bool) {
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
		return parseStringToInt64(x, cfg)
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(rv.Uint()), true
	case reflect.Float32, reflect.Float64:
		return int64(rv.Float()), true
	case reflect.Bool:
		if rv.Bool() {
			return 1, true
		}
		return 0, true
	case reflect.String:
		return parseStringToInt64(rv.String(), cfg)
	}
	return 0, false
}

func toInt64Strict(v any, cfg config) (int64, bool) {
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
		if uint64(x) > math.MaxInt64 {
			return 0, false
		}
		return int64(x), true
	case uint8:
		return int64(x), true
	case uint16:
		return int64(x), true
	case uint32:
		return int64(x), true
	case uint64:
		if x > math.MaxInt64 {
			return 0, false
		}
		return int64(x), true
	case float32:
		return float64ToInt64Strict(float64(x))
	case float64:
		return float64ToInt64Strict(x)
	case bool:
		if x {
			return 1, true
		}
		return 0, true
	case string:
		return parseStringToInt64Strict(x, cfg)
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u := rv.Uint()
		if u > math.MaxInt64 {
			return 0, false
		}
		return int64(u), true
	case reflect.Float32, reflect.Float64:
		return float64ToInt64Strict(rv.Float())
	case reflect.Bool:
		if rv.Bool() {
			return 1, true
		}
		return 0, true
	case reflect.String:
		return parseStringToInt64Strict(rv.String(), cfg)
	}
	return 0, false
}

func parseStringToInt64(s string, cfg config) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	if i, err := cfg.parseInt(s, 10, 64); err == nil {
		return i, true
	}
	if f, err := cfg.parseFloat(s, 64); err == nil {
		return int64(f), true
	}
	return 0, false
}

func parseStringToInt64Strict(s string, cfg config) (int64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	if i, err := cfg.parseInt(s, 10, 64); err == nil {
		return i, true
	}
	if f, err := cfg.parseFloat(s, 64); err == nil {
		return float64ToInt64Strict(f)
	}
	return 0, false
}

func float64ToInt64Strict(f float64) (int64, bool) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, false
	}
	if f < float64(math.MinInt64) || f >= -float64(math.MinInt64) {
		return 0, false
	}
	return int64(f), true
}

func toFloat64(v any, cfg config) (float64, bool) {
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
		f, err := cfg.parseFloat(s, 64)
		if err == nil {
			return f, true
		}
		return 0, false
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	case reflect.String:
		s := strings.TrimSpace(rv.String())
		if s == "" {
			return 0, false
		}
		f, err := cfg.parseFloat(s, 64)
		if err == nil {
			return f, true
		}
		return 0, false
	}
	if i, ok := toInt64(v, cfg); ok {
		return float64(i), true
	}
	return 0, false
}

func defaultBoolParser(s string) (bool, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "true", "yes", "y", "ok", "1", "on":
		return true, nil
	case "false", "no", "n", "0", "off":
		return false, nil
	default:
		return false, fmt.Errorf("cannot parse bool %q", s)
	}
}

func invalidConversionError(target string) error {
	return knifer.WrapError(
		knifer.ErrCodeInvalidInput,
		"conv: invalid conversion to "+target,
		ErrInvalidConversion,
	)
}
