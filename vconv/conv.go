package vconv

import convimpl "github.com/imajinyun/knifer-go/internal/conv"

// ErrInvalidConversion reports that a value cannot be converted to the requested scalar type.
var ErrInvalidConversion = convimpl.ErrInvalidConversion

// Option customizes conversion helpers per call.
type Option = convimpl.Option

func WithBoolParser(parser func(string) (bool, error)) Option {
	return convimpl.WithBoolParser(parser)
}

func WithParseIntFunc(parser func(string, int, int) (int64, error)) Option {
	return convimpl.WithParseIntFunc(parser)
}

func WithParseFloatFunc(parser func(string, int) (float64, error)) Option {
	return convimpl.WithParseFloatFunc(parser)
}

func WithFormatBoolFunc(formatter func(bool) string) Option {
	return convimpl.WithFormatBoolFunc(formatter)
}

func WithFormatFloatFunc(formatter func(float64, byte, int, int) string) Option {
	return convimpl.WithFormatFloatFunc(formatter)
}

func ToString(v any) string { return convimpl.ToString(v) }
func ToStringWithOptions(v any, opts ...Option) string {
	return convimpl.ToStringWithOptions(v, opts...)
}
func ToStringDefault(v any, def string) string { return convimpl.ToStringDefault(v, def) }
func ToStringDefaultWithOptions(v any, def string, opts ...Option) string {
	return convimpl.ToStringDefaultWithOptions(v, def, opts...)
}
func ToInt(v any) int                            { return convimpl.ToInt(v) }
func ToIntWithOptions(v any, opts ...Option) int { return convimpl.ToIntWithOptions(v, opts...) }
func ToIntDefault(v any, def int) int            { return convimpl.ToIntDefault(v, def) }
func ToIntDefaultWithOptions(v any, def int, opts ...Option) int {
	return convimpl.ToIntDefaultWithOptions(v, def, opts...)
}
func ToIntE(v any) (int, error) { return convimpl.ToIntE(v) }
func ToIntEWithOptions(v any, opts ...Option) (int, error) {
	return convimpl.ToIntEWithOptions(v, opts...)
}
func ToInt64(v any) int64 { return convimpl.ToInt64(v) }
func ToInt64WithOptions(v any, opts ...Option) int64 {
	return convimpl.ToInt64WithOptions(v, opts...)
}
func ToInt64Default(v any, def int64) int64 { return convimpl.ToInt64Default(v, def) }
func ToInt64DefaultWithOptions(v any, def int64, opts ...Option) int64 {
	return convimpl.ToInt64DefaultWithOptions(v, def, opts...)
}
func ToInt64E(v any) (int64, error) { return convimpl.ToInt64E(v) }
func ToInt64EWithOptions(v any, opts ...Option) (int64, error) {
	return convimpl.ToInt64EWithOptions(v, opts...)
}
func ToFloat64(v any) float64 { return convimpl.ToFloat64(v) }
func ToFloat64WithOptions(v any, opts ...Option) float64 {
	return convimpl.ToFloat64WithOptions(v, opts...)
}
func ToFloat64Default(v any, def float64) float64 { return convimpl.ToFloat64Default(v, def) }
func ToFloat64DefaultWithOptions(v any, def float64, opts ...Option) float64 {
	return convimpl.ToFloat64DefaultWithOptions(v, def, opts...)
}
func ToFloat64E(v any) (float64, error) { return convimpl.ToFloat64E(v) }
func ToFloat64EWithOptions(v any, opts ...Option) (float64, error) {
	return convimpl.ToFloat64EWithOptions(v, opts...)
}
func ToBool(v any) bool                            { return convimpl.ToBool(v) }
func ToBoolWithOptions(v any, opts ...Option) bool { return convimpl.ToBoolWithOptions(v, opts...) }
func ToBoolDefault(v any, def bool) bool           { return convimpl.ToBoolDefault(v, def) }
func ToBoolDefaultWithOptions(v any, def bool, opts ...Option) bool {
	return convimpl.ToBoolDefaultWithOptions(v, def, opts...)
}
func ToBoolE(v any) (bool, error) { return convimpl.ToBoolE(v) }
func ToBoolEWithOptions(v any, opts ...Option) (bool, error) {
	return convimpl.ToBoolEWithOptions(v, opts...)
}
func ToBytes(v any) []byte { return convimpl.ToBytes(v) }
func ToBytesWithOptions(v any, opts ...Option) []byte {
	return convimpl.ToBytesWithOptions(v, opts...)
}
