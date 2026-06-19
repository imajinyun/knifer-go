package vbean

import beanimpl "github.com/imajinyun/go-knifer/internal/bean"

// Option customizes bean mapping behavior.
type Option = beanimpl.Option

// Options controls struct/map property mapping.
type Options = beanimpl.Options

// Error is the code-aware error type returned by bean helpers.
type Error = beanimpl.BeanError

// Result reports which source properties were consumed, skipped, or left unused.
type Result = beanimpl.Result

// NewOptions returns default mapping options.
func NewOptions() Options { return beanimpl.NewOptions() }

// WithTagNames sets tag names used to resolve field names and aliases.
func WithTagNames(names ...string) Option { return beanimpl.WithTagNames(names...) }

// WithWeaklyTyped controls whether weak type conversion is enabled.
func WithWeaklyTyped(enable bool) Option { return beanimpl.WithWeaklyTyped(enable) }

// WithCaseInsensitive controls case-insensitive name matching.
func WithCaseInsensitive(enable bool) Option { return beanimpl.WithCaseInsensitive(enable) }

// WithIgnoreEmpty skips empty source values.
func WithIgnoreEmpty(enable bool) Option { return beanimpl.WithIgnoreEmpty(enable) }

// WithIgnoreZero skips zero source values.
func WithIgnoreZero(enable bool) Option { return beanimpl.WithIgnoreZero(enable) }

// WithStrictUnused reports unmatched source keys or fields as errors after assignment.
func WithStrictUnused(enable bool) Option { return beanimpl.WithStrictUnused(enable) }

// WithBoolParser sets the parser used during weak string-to-bool conversion.
func WithBoolParser(parser func(string) (bool, error)) Option {
	return beanimpl.WithBoolParser(parser)
}

// WithIntParser sets the parser used during weak string-to-signed-integer conversion.
func WithIntParser(parser func(string, int, int) (int64, error)) Option {
	return beanimpl.WithIntParser(parser)
}

// WithUintParser sets the parser used during weak string-to-unsigned-integer conversion.
func WithUintParser(parser func(string, int, int) (uint64, error)) Option {
	return beanimpl.WithUintParser(parser)
}

// WithFloatParser sets the parser used during weak string-to-floating-point conversion.
func WithFloatParser(parser func(string, int) (float64, error)) Option {
	return beanimpl.WithFloatParser(parser)
}

// ToMap converts a struct or map to map[string]any using field tags and aliases.
func ToMap(src any, opts ...Option) (map[string]any, error) { return beanimpl.ToMap(src, opts...) }

// FillMap copies properties from src into dst.
func FillMap(src any, dst map[string]any, opts ...Option) error {
	return beanimpl.FillMap(src, dst, opts...)
}

// ToStruct copies properties from src into dst, which must be a pointer to struct.
func ToStruct(src any, dst any, opts ...Option) error { return beanimpl.ToStruct(src, dst, opts...) }

// Decode converts matching properties from src into dst using the configured weak conversion rules.
func Decode(src any, dst any, opts ...Option) error { return beanimpl.Decode(src, dst, opts...) }

// DecodeResult converts matching properties from src into dst and reports mapping metadata.
func DecodeResult(src any, dst any, opts ...Option) (Result, error) {
	return beanimpl.DecodeResult(src, dst, opts...)
}

// Merge copies one or more sources into dst from left to right.
func Merge(dst any, sources ...any) error { return beanimpl.Merge(dst, sources...) }

// MergeResult copies one or more sources into dst and reports aggregate mapping metadata.
func MergeResult(dst any, sources ...any) (Result, error) {
	return beanimpl.MergeResult(dst, sources...)
}

// MergeWithOptions copies sources into dst from left to right using options.
func MergeWithOptions(dst any, sources []any, opts ...Option) error {
	return beanimpl.MergeWithOptions(dst, sources, opts...)
}

// MergeResultWithOptions copies sources into dst from left to right using options and reports metadata.
func MergeResultWithOptions(dst any, sources []any, opts ...Option) (Result, error) {
	return beanimpl.MergeResultWithOptions(dst, sources, opts...)
}

// Copy is an alias of CopyProperties.
func Copy(src any, dst any, opts ...Option) error { return beanimpl.Copy(src, dst, opts...) }

// CopyProperties copies matching properties between struct/map values.
func CopyProperties(src any, dst any, opts ...Option) error {
	return beanimpl.CopyProperties(src, dst, opts...)
}
