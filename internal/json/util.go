package json

import (
	"encoding/json"
	"io"
	"strings"
)

type encodeConfig struct {
	cfg    *Config
	indent int
}

type parseConfig struct {
	cfg            *Config
	unmarshalFunc  func([]byte, any) error
	decoderFactory func(io.Reader) *json.Decoder
}

type validConfig struct {
	valid func([]byte) bool
}

// BeanOption customizes ToBeanWithOptions and ToListWithOptions.
type BeanOption func(*beanConfig)

type beanConfig struct {
	cfg           *Config
	unmarshalFunc func([]byte, any) error
}

// EncodeOption customizes JSON serialization helpers.
type EncodeOption func(*encodeConfig)

// ParseOption customizes JSON parsing helpers.
type ParseOption func(*parseConfig)

// ValidOption customizes JSON validity helpers.
type ValidOption func(*validConfig)

// WithJSONValidFunc sets the validator used by IsJSONWithOptions.
func WithJSONValidFunc(valid func([]byte) bool) ValidOption {
	return func(c *validConfig) {
		if valid != nil {
			c.valid = valid
		}
	}
}

func defaultEncodeConfig(indent int) encodeConfig {
	return encodeConfig{cfg: NewConfig(), indent: indent}
}

// WithConfig sets the JSON config used by serialization helpers.
func WithConfig(cfg *Config) EncodeOption {
	return func(c *encodeConfig) {
		if cfg != nil {
			c.cfg = cfg
		}
	}
}

// WithIndent sets the indentation width. Use 0 for compact output.
func WithIndent(indent int) EncodeOption { return func(c *encodeConfig) { c.indent = indent } }

// WithIgnoreNullValue controls whether null values are ignored during serialization.
func WithIgnoreNullValue(ignore bool) EncodeOption {
	return func(c *encodeConfig) {
		c.cfg = c.cfg.Clone()
		c.cfg.IgnoreNullValue = ignore
	}
}

// WithDateFormat sets the time layout used for time.Time values.
func WithDateFormat(layout string) EncodeOption {
	return func(c *encodeConfig) {
		c.cfg = c.cfg.Clone()
		c.cfg.DateFormat = layout
	}
}

// WithMarshalFunc sets the marshal provider used when wrapping structs for serialization.
func WithMarshalFunc(marshal func(any) ([]byte, error)) EncodeOption {
	return func(c *encodeConfig) {
		if marshal != nil {
			c.cfg = c.cfg.Clone()
			c.cfg.MarshalFunc = marshal
		}
	}
}

// WithUnmarshalFunc sets the unmarshal provider stored in the JSON config.
func WithUnmarshalFunc(unmarshal func([]byte, any) error) EncodeOption {
	return func(c *encodeConfig) {
		if unmarshal != nil {
			c.cfg = c.cfg.Clone()
			c.cfg.UnmarshalFunc = unmarshal
		}
	}
}

// WithParseConfig sets the JSON config used by parsing helpers.
func WithParseConfig(cfg *Config) ParseOption {
	return func(c *parseConfig) {
		if cfg != nil {
			c.cfg = cfg
		}
	}
}

// WithParseUnmarshalFunc sets a per-call unmarshal provider for parsing helpers.
func WithParseUnmarshalFunc(unmarshal func([]byte, any) error) ParseOption {
	return func(c *parseConfig) {
		if unmarshal != nil {
			c.unmarshalFunc = unmarshal
		}
	}
}

// WithParseDecoderFactory sets a per-call decoder factory for token parsing helpers.
func WithParseDecoderFactory(factory func(io.Reader) *json.Decoder) ParseOption {
	return func(c *parseConfig) {
		if factory != nil {
			c.decoderFactory = factory
		}
	}
}

// WithDecoderFactory sets the decoder factory stored in the JSON config.
func WithDecoderFactory(factory func(io.Reader) *json.Decoder) EncodeOption {
	return func(c *encodeConfig) {
		if factory != nil {
			c.cfg = c.cfg.Clone()
			c.cfg.DecoderFactory = factory
		}
	}
}

// WithSprintFunc sets the fallback scalar formatter stored in the JSON config.
func WithSprintFunc(sprint func(any) string) EncodeOption {
	return func(c *encodeConfig) {
		if sprint != nil {
			c.cfg = c.cfg.Clone()
			c.cfg.SprintFunc = sprint
		}
	}
}

// WithParseIntFunc sets the integer parser stored in the JSON config.
func WithParseIntFunc(parse func(string, int, int) (int64, error)) EncodeOption {
	return func(c *encodeConfig) {
		if parse != nil {
			c.cfg = c.cfg.Clone()
			c.cfg.ParseIntFunc = parse
		}
	}
}

// WithParseFloatFunc sets the float parser stored in the JSON config.
func WithParseFloatFunc(parse func(string, int) (float64, error)) EncodeOption {
	return func(c *encodeConfig) {
		if parse != nil {
			c.cfg = c.cfg.Clone()
			c.cfg.ParseFloatFunc = parse
		}
	}
}

// WithParseBoolFunc sets the bool parser stored in the JSON config.
func WithParseBoolFunc(parse func(string) (bool, error)) EncodeOption {
	return func(c *encodeConfig) {
		if parse != nil {
			c.cfg = c.cfg.Clone()
			c.cfg.ParseBoolFunc = parse
		}
	}
}

// WithFormatIntFunc sets the integer formatter stored in the JSON config.
func WithFormatIntFunc(format func(int64, int) string) EncodeOption {
	return func(c *encodeConfig) {
		if format != nil {
			c.cfg = c.cfg.Clone()
			c.cfg.FormatIntFunc = format
		}
	}
}

// WithFormatFloatFunc sets the float formatter stored in the JSON config.
func WithFormatFloatFunc(format func(float64, byte, int, int) string) EncodeOption {
	return func(c *encodeConfig) {
		if format != nil {
			c.cfg = c.cfg.Clone()
			c.cfg.FormatFloatFunc = format
		}
	}
}

// WithBeanConfig sets the JSON config used by bean conversion helpers.
func WithBeanConfig(cfg *Config) BeanOption {
	return func(c *beanConfig) {
		if cfg != nil {
			c.cfg = cfg
		}
	}
}

// WithBeanUnmarshalFunc sets a per-call unmarshal provider for bean conversion helpers.
func WithBeanUnmarshalFunc(unmarshal func([]byte, any) error) BeanOption {
	return func(c *beanConfig) {
		if unmarshal != nil {
			c.unmarshalFunc = unmarshal
		}
	}
}

func applyEncodeOptions(defaultIndent int, opts []EncodeOption) encodeConfig {
	cfg := defaultEncodeConfig(defaultIndent)
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

func applyParseOptions(opts []ParseOption) parseConfig {
	cfg := parseConfig{cfg: NewConfig()}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

func applyBeanOptions(opts []BeanOption) beanConfig {
	cfg := beanConfig{cfg: NewConfig()}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return cfg
}

func applyValidOptions(opts []ValidOption) validConfig {
	cfg := validConfig{valid: json.Valid}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.valid == nil {
		cfg.valid = json.Valid
	}
	return cfg
}

// Parse 自动判断 JSON 类型：对象/数组/基础值。
func Parse(src any) (any, error) { return ParseWithConfig(src, nil) }

// ParseWithOptions automatically detects and parses JSON with options.
func ParseWithOptions(src any, opts ...ParseOption) (any, error) {
	cfg := applyParseOptions(opts)
	if cfg.unmarshalFunc != nil {
		cfg.cfg = cfg.cfg.Clone()
		cfg.cfg.UnmarshalFunc = cfg.unmarshalFunc
	}
	if cfg.decoderFactory != nil {
		cfg.cfg = cfg.cfg.Clone()
		cfg.cfg.DecoderFactory = cfg.decoderFactory
	}
	return ParseWithConfig(src, cfg.cfg)
}

// ParseWithConfig 解析并使用配置。
func ParseWithConfig(src any, cfg *Config) (any, error) {
	switch x := src.(type) {
	case nil:
		return Null, nil
	case []byte:
		return parseBytesWithConfig(x, cfg)
	case string:
		return parseBytesWithConfig([]byte(x), cfg)
	case *JSONObject, *JSONArray:
		return x, nil
	}
	// 复杂类型：先 wrap 再返回
	return wrap(src, configOrDefault(cfg)), nil
}

// ParseObj 强制解析为 JSONObject。
func ParseObj(src any) (*JSONObject, error) { return ParseObjWithConfig(src, nil) }

// ParseObjWithOptions parses src as a JSON object with options.
func ParseObjWithOptions(src any, opts ...ParseOption) (*JSONObject, error) {
	v, err := ParseWithOptions(src, opts...)
	if err != nil {
		return nil, err
	}
	if obj, ok := v.(*JSONObject); ok {
		return obj, nil
	}
	return nil, NewJSONError("expect json object, got %T", v)
}

// ParseObjWithConfig 解析为 JSONObject。
func ParseObjWithConfig(src any, cfg *Config) (*JSONObject, error) {
	v, err := ParseWithConfig(src, cfg)
	if err != nil {
		return nil, err
	}
	if obj, ok := v.(*JSONObject); ok {
		return obj, nil
	}
	return nil, NewJSONError("expect json object, got %T", v)
}

// ParseArray 强制解析为 JSONArray。
func ParseArray(src any) (*JSONArray, error) { return ParseArrayWithConfig(src, nil) }

// ParseArrayWithOptions parses src as a JSON array with options.
func ParseArrayWithOptions(src any, opts ...ParseOption) (*JSONArray, error) {
	v, err := ParseWithOptions(src, opts...)
	if err != nil {
		return nil, err
	}
	if arr, ok := v.(*JSONArray); ok {
		return arr, nil
	}
	return nil, NewJSONError("expect json array, got %T", v)
}

// ParseArrayWithConfig 解析为 JSONArray。
func ParseArrayWithConfig(src any, cfg *Config) (*JSONArray, error) {
	v, err := ParseWithConfig(src, cfg)
	if err != nil {
		return nil, err
	}
	if arr, ok := v.(*JSONArray); ok {
		return arr, nil
	}
	return nil, NewJSONError("expect json array, got %T", v)
}

// ToJSONStr 紧凑序列化。
func ToJSONStr(v any, opts ...EncodeOption) (string, error) {
	cfg := applyEncodeOptions(0, opts)
	w := wrap(v, cfg.cfg)
	return writeValueWithConfig(w, cfg.indent, cfg.cfg)
}

// ToJSONPrettyStr 4 空格缩进序列化。
func ToJSONPrettyStr(v any, opts ...EncodeOption) (string, error) {
	cfg := applyEncodeOptions(4, opts)
	w := wrap(v, cfg.cfg)
	return writeValueWithConfig(w, cfg.indent, cfg.cfg)
}

// ToJSONStrIndent 自定义缩进序列化。
func ToJSONStrIndent(v any, indent int, opts ...EncodeOption) (string, error) {
	cfg := applyEncodeOptions(indent, opts)
	w := wrap(v, cfg.cfg)
	return writeValueWithConfig(w, cfg.indent, cfg.cfg)
}

// ToJSONStrWithConfig serializes v using cfg.
func ToJSONStrWithConfig(v any, cfg *Config) (string, error) { return ToJSONStr(v, WithConfig(cfg)) }

// ToJSONPrettyStrWithConfig serializes v using cfg and cfg.IndentFactor.
func ToJSONPrettyStrWithConfig(v any, cfg *Config) (string, error) {
	cfg = configOrDefault(cfg)
	return ToJSONPrettyStr(v, WithConfig(cfg), WithIndent(cfg.IndentFactor))
}

// IsJSON 检查字符串是否合法 JSON。
func IsJSON(s string) bool {
	return IsJSONWithOptions(s)
}

// IsJSONWithOptions checks whether a string is valid JSON with options.
func IsJSONWithOptions(s string, opts ...ValidOption) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	return applyValidOptions(opts).valid([]byte(s))
}

// IsJSONObj 检查字符串是否是 JSON 对象。
func IsJSONObj(s string) bool {
	return IsJSONObjWithOptions(s)
}

// IsJSONObjWithOptions checks whether a string is a JSON object with options.
func IsJSONObjWithOptions(s string, opts ...ValidOption) bool {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") || !strings.HasSuffix(s, "}") {
		return false
	}
	return IsJSONWithOptions(s, opts...)
}

// IsJSONArray 检查字符串是否是 JSON 数组。
func IsJSONArray(s string) bool {
	return IsJSONArrayWithOptions(s)
}

// IsJSONArrayWithOptions checks whether a string is a JSON array with options.
func IsJSONArrayWithOptions(s string, opts ...ValidOption) bool {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
		return false
	}
	return IsJSONWithOptions(s, opts...)
}

// GetByPath 顶层路径查询。
func GetByPath(root any, path string) any { return getByPath(root, path) }

// GetByPathOr 顶层路径查询，缺省回退。
func GetByPathOr(root any, path string, def any) any {
	if v := getByPath(root, path); v != nil && !IsNull(v) {
		return v
	}
	return def
}

// PutByPath 顶层路径写入。
func PutByPath(root any, path string, value any) error { return putByPath(root, path, value) }

// Quote 在 JSON 字符串两侧加引号并进行必要转义。
func Quote(s string) string {
	var sb strings.Builder
	writeQuoted(&sb, s)
	return sb.String()
}

// configOrDefault 返回非空配置。
func configOrDefault(cfg *Config) *Config {
	if cfg == nil {
		return NewConfig()
	}
	return cfg
}
