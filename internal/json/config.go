package json

import (
	stdjson "encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Config controls JSON serialization behavior.
type Config struct {
	// IgnoreNullValue 序列化时忽略 null。
	IgnoreNullValue bool
	// IgnoreCase 键不区分大小写（仅在 JSONObject 上生效，写入时按首次出现的大小写存储）。
	IgnoreCase bool
	// IgnoreError 在转换失败时忽略错误。
	IgnoreError bool
	// DateFormat 日期格式（time.Time 的 layout），为空时输出毫秒数。
	DateFormat string
	// IndentFactor pretty 输出时缩进字符数。
	IndentFactor int
	// MarshalFunc serializes arbitrary Go values when struct tags must be honored. nil means encoding/json.Marshal.
	MarshalFunc func(any) ([]byte, error)
	// UnmarshalFunc deserializes JSON bytes for bean conversion and struct wrapping. nil means encoding/json with UseNumber.
	UnmarshalFunc func([]byte, any) error
	// DecoderFactory creates JSON decoders for token parsing. nil means encoding/json.NewDecoder with UseNumber.
	DecoderFactory func(io.Reader) *stdjson.Decoder
	// SprintFunc formats fallback scalar values. nil means fmt.Sprint.
	SprintFunc func(any) string
	// ParseIntFunc parses integer strings. nil means strconv.ParseInt.
	ParseIntFunc func(string, int, int) (int64, error)
	// ParseFloatFunc parses floating-point strings. nil means strconv.ParseFloat.
	ParseFloatFunc func(string, int) (float64, error)
	// ParseBoolFunc parses boolean strings. nil means the package-compatible bool parser.
	ParseBoolFunc func(string) (bool, error)
	// FormatIntFunc formats integer values. nil means strconv.FormatInt.
	FormatIntFunc func(int64, int) string
	// FormatFloatFunc formats floating-point values. nil means strconv.FormatFloat.
	FormatFloatFunc func(float64, byte, int, int) string
}

// NewConfig 创建一个默认配置。
func NewConfig() *Config {
	return &Config{IndentFactor: 4}
}

// CreateConfig creates a default JSON config.
func CreateConfig() *Config { return NewConfig() }

// Clone 拷贝配置。
func (c *Config) Clone() *Config {
	if c == nil {
		return NewConfig()
	}
	cp := *c
	return &cp
}

func (c *Config) sprint(v any) string {
	if c != nil && c.SprintFunc != nil {
		return c.SprintFunc(v)
	}
	return fmt.Sprint(v)
}

func (c *Config) parseInt(s string, base, bitSize int) (int64, error) {
	if c != nil && c.ParseIntFunc != nil {
		return c.ParseIntFunc(s, base, bitSize)
	}
	return strconv.ParseInt(s, base, bitSize)
}

func (c *Config) parseFloat(s string, bitSize int) (float64, error) {
	if c != nil && c.ParseFloatFunc != nil {
		return c.ParseFloatFunc(s, bitSize)
	}
	return strconv.ParseFloat(s, bitSize)
}

func (c *Config) parseBool(s string) (bool, error) {
	if c != nil && c.ParseBoolFunc != nil {
		return c.ParseBoolFunc(s)
	}
	return defaultParseBool(s)
}

func (c *Config) formatInt(v int64, base int) string {
	if c != nil && c.FormatIntFunc != nil {
		return c.FormatIntFunc(v, base)
	}
	return strconv.FormatInt(v, base)
}

func (c *Config) formatFloat(v float64, fmtByte byte, prec, bitSize int) string {
	if c != nil && c.FormatFloatFunc != nil {
		return c.FormatFloatFunc(v, fmtByte, prec, bitSize)
	}
	return strconv.FormatFloat(v, fmtByte, prec, bitSize)
}

func defaultParseBool(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no", "":
		return false, nil
	default:
		return false, strconv.ErrSyntax
	}
}
