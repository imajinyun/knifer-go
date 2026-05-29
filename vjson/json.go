package vjson

import jsonx "github.com/imajinyun/go-knifer/internal/json"

// Object is an ordered JSON object.
type Object = jsonx.JSONObject

// Array is an ordered JSON array.
type Array = jsonx.JSONArray

// Config controls JSON serialization behavior.
type Config = jsonx.Config

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

// IsNull reports whether v is nil or JSON null.
func IsNull(v any) bool { return jsonx.IsNull(v) }

// Parse automatically detects and parses JSON.
func Parse(src any) (any, error) { return jsonx.Parse(src) }

// ParseObj parses src as a JSON object.
func ParseObj(src any) (*Object, error) { return jsonx.ParseObj(src) }

// ParseArray parses src as a JSON array.
func ParseArray(src any) (*Array, error) { return jsonx.ParseArray(src) }

// ToStr serializes v to compact JSON.
func ToStr(v any) (string, error) { return jsonx.ToJSONStr(v) }

// ToPrettyStr serializes v to pretty JSON with 4-space indentation.
func ToPrettyStr(v any) (string, error) { return jsonx.ToJSONPrettyStr(v) }

// ToStrIndent serializes v to pretty JSON with custom indentation.
func ToStrIndent(v any, indent int) (string, error) {
	return jsonx.ToJSONStrIndent(v, indent)
}

// Format formats raw JSON string.
func Format(raw string) string { return jsonx.FormatJSONStr(raw) }

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
