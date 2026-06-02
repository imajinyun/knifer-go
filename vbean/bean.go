package vbean

import beanimpl "github.com/imajinyun/go-knifer/internal/bean"

// Option customizes bean mapping behavior.
type Option = beanimpl.Option

// Options controls struct/map property mapping.
type Options = beanimpl.Options

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

// ToMap converts a struct or map to map[string]any using field tags and aliases.
func ToMap(src any, opts ...Option) (map[string]any, error) { return beanimpl.ToMap(src, opts...) }

// FillMap copies properties from src into dst.
func FillMap(src any, dst map[string]any, opts ...Option) error {
	return beanimpl.FillMap(src, dst, opts...)
}

// ToStruct copies properties from src into dst, which must be a pointer to struct.
func ToStruct(src any, dst any, opts ...Option) error { return beanimpl.ToStruct(src, dst, opts...) }

// Copy is an alias of CopyProperties.
func Copy(src any, dst any, opts ...Option) error { return beanimpl.Copy(src, dst, opts...) }

// CopyProperties copies matching properties between struct/map values.
func CopyProperties(src any, dst any, opts ...Option) error {
	return beanimpl.CopyProperties(src, dst, opts...)
}
