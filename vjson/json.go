package vjson

import jsonx "github.com/imajinyun/go-knifer/internal/json"

// Object is an ordered JSON object.
type Object = jsonx.JSONObject

// Array is an ordered JSON array.
type Array = jsonx.JSONArray

// Config controls JSON serialization behavior.
type Config = jsonx.Config

// EncodeOption customizes JSON serialization helpers.
type EncodeOption = jsonx.EncodeOption

// FormatOption customizes raw JSON string formatting.
type FormatOption = jsonx.FormatOption

// ParseOption customizes JSON parsing helpers.
type ParseOption = jsonx.ParseOption

// Error is the JSON module error type.
type Error = jsonx.JSONError

// Null is the JSON null singleton value.
var Null = jsonx.Null

// NewObject creates an empty ordered JSON object.
func NewObject() *Object { return jsonx.NewJSONObject() }

// NewObjectWithConfig creates a JSON object with cfg.
func NewObjectWithConfig(cfg *Config) *Object {
	return jsonx.NewJSONObjectWithConfig(cfg)
}

// NewArray creates an empty ordered JSON array.
func NewArray() *Array { return jsonx.NewJSONArray() }

// NewArrayWithConfig creates a JSON array with cfg.
func NewArrayWithConfig(cfg *Config) *Array {
	return jsonx.NewJSONArrayWithConfig(cfg)
}

// NewConfig creates a default JSON config.
func NewConfig() *Config { return jsonx.NewConfig() }

// WithConfig sets the JSON config used by serialization helpers.
func WithConfig(cfg *Config) EncodeOption { return jsonx.WithConfig(cfg) }

// WithIndent sets the indentation width. Use 0 for compact output.
func WithIndent(indent int) EncodeOption { return jsonx.WithIndent(indent) }

// WithIgnoreNullValue controls whether null values are ignored during serialization.
func WithIgnoreNullValue(ignore bool) EncodeOption { return jsonx.WithIgnoreNullValue(ignore) }

// WithDateFormat sets the time layout used for time.Time values.
func WithDateFormat(layout string) EncodeOption { return jsonx.WithDateFormat(layout) }

// WithFormatIndent sets the indentation string used by FormatWithOptions.
func WithFormatIndent(indent string) FormatOption { return jsonx.WithFormatIndent(indent) }

// WithFormatIndentWidth sets indentation to n spaces.
func WithFormatIndentWidth(n int) FormatOption { return jsonx.WithFormatIndentWidth(n) }

// WithFormatSpaceAfterKey controls whether a space is written after ':'.
func WithFormatSpaceAfterKey(space bool) FormatOption { return jsonx.WithFormatSpaceAfterKey(space) }

// WithParseConfig sets the JSON config used by parsing helpers.
func WithParseConfig(cfg *Config) ParseOption { return jsonx.WithParseConfig(cfg) }

// IsNull reports whether v is nil or JSON null.
func IsNull(v any) bool { return jsonx.IsNull(v) }

// Parse automatically detects and parses JSON.
func Parse(src any) (any, error) { return jsonx.Parse(src) }

// ParseWithOptions automatically detects and parses JSON with options.
func ParseWithOptions(src any, opts ...ParseOption) (any, error) {
	return jsonx.ParseWithOptions(src, opts...)
}

// ParseObj parses src as a JSON object.
func ParseObj(src any) (*Object, error) { return jsonx.ParseObj(src) }

// ParseObjWithOptions parses src as a JSON object with options.
func ParseObjWithOptions(src any, opts ...ParseOption) (*Object, error) {
	return jsonx.ParseObjWithOptions(src, opts...)
}

// ParseArray parses src as a JSON array.
func ParseArray(src any) (*Array, error) { return jsonx.ParseArray(src) }

// ParseArrayWithOptions parses src as a JSON array with options.
func ParseArrayWithOptions(src any, opts ...ParseOption) (*Array, error) {
	return jsonx.ParseArrayWithOptions(src, opts...)
}

// ToStr serializes v to compact JSON.
func ToStr(v any, opts ...EncodeOption) (string, error) { return jsonx.ToJSONStr(v, opts...) }

// ToPrettyStr serializes v to pretty JSON with 4-space indentation.
func ToPrettyStr(v any, opts ...EncodeOption) (string, error) {
	return jsonx.ToJSONPrettyStr(v, opts...)
}

// ToStrIndent serializes v to pretty JSON with custom indentation.
func ToStrIndent(v any, indent int, opts ...EncodeOption) (string, error) {
	return jsonx.ToJSONStrIndent(v, indent, opts...)
}

// ToStrWithConfig serializes v using cfg.
func ToStrWithConfig(v any, cfg *Config) (string, error) { return jsonx.ToJSONStrWithConfig(v, cfg) }

// ToPrettyStrWithConfig serializes v using cfg and cfg.IndentFactor.
func ToPrettyStrWithConfig(v any, cfg *Config) (string, error) {
	return jsonx.ToJSONPrettyStrWithConfig(v, cfg)
}

// Format formats raw JSON string.
func Format(raw string) string { return jsonx.FormatJSONStr(raw) }

// FormatWithOptions formats raw JSON string with custom formatting options.
func FormatWithOptions(raw string, opts ...FormatOption) string {
	return jsonx.FormatJSONStrWithOptions(raw, opts...)
}

// IsJSON reports whether s is valid JSON.
func IsJSON(s string) bool { return jsonx.IsJSON(s) }

// IsObj reports whether s is a JSON object.
func IsObj(s string) bool { return jsonx.IsJSONObj(s) }

// IsArray reports whether s is a JSON array.
func IsArray(s string) bool { return jsonx.IsJSONArray(s) }

// GetByPath gets a value by path expression.
func GetByPath(root any, path string) any { return jsonx.GetByPath(root, path) }

// GetByPathOr gets a value by path expression with a default.
func GetByPathOr(root any, path string, def any) any {
	return jsonx.GetByPathOr(root, path, def)
}

// PutByPath writes a value by path expression.
func PutByPath(root any, path string, value any) error {
	return jsonx.PutByPath(root, path, value)
}

// Quote adds JSON double quotes and escapes s.
func Quote(s string) string { return jsonx.Quote(s) }

// ToBean deserializes JSON to dst, which must be a pointer.
func ToBean(src any, dst any) error { return jsonx.ToBean(src, dst) }

// ToList deserializes a JSON array to dst, which must point to a slice.
func ToList(src any, dst any) error { return jsonx.ToList(src, dst) }

// XMLToJSON 将 XML 字符串解析为 JSONObject。
func XMLToJSON(xmlStr string) (*Object, error) { return jsonx.XMLToJSON(xmlStr) }

// ToXML serializes a JSON value to XML string. When rootTag is empty, keys are concatenated directly.
func ToXML(root any, rootTag string) (string, error) {
	return jsonx.JSONToXML(root, rootTag)
}
