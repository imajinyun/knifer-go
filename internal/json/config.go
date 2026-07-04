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
	// IgnoreNullValue ignores null values during serialization.
	IgnoreNullValue bool
	// IgnoreCase makes keys case-insensitive; only JSONObject uses it and writes keys with their first-seen casing.
	IgnoreCase bool
	// IgnoreError ignores errors on conversion failure.
	IgnoreError bool
	// DateFormat sets the date format as a time.Time layout; empty output uses milliseconds.
	DateFormat string
	// IndentFactor sets the indentation width for pretty output.
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

// NewConfig creates a default config.
func NewConfig() *Config {
	return &Config{IndentFactor: 4}
}

// CreateConfig creates a default JSON config.
func CreateConfig() *Config { return NewConfig() }

// Clone copies the config.
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

func formatUint64(v uint64, cfg *Config) string {
	if v <= maxInt64AsUint64 {
		return configOrDefault(cfg).formatInt(int64(v), 10)
	}
	return strconv.FormatUint(v, 10)
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
