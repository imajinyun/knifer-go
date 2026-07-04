package json

import (
	"strings"
	"unicode/utf8"
)

// writeValue writes a value as a JSON string and pretty-formats it when indent > 0.
func writeValue(v any, indent int) (string, error) {
	return writeValueWithConfig(v, indent, nil)
}

func writeValueWithConfig(v any, indent int, cfg *Config) (string, error) {
	var sb strings.Builder
	if err := writeAny(&sb, v, indent, 0, configOrDefault(cfg)); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func writeAny(sb *strings.Builder, v any, indent, depth int, cfg *Config) error {
	if IsNull(v) {
		sb.WriteString("null")
		return nil
	}
	switch x := v.(type) {
	case *JSONObject:
		return writeObject(sb, x, indent, depth)
	case *JSONArray:
		return writeArray(sb, x, indent, depth)
	case string:
		writeQuoted(sb, x)
	case bool:
		if x {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case int64:
		sb.WriteString(cfg.formatInt(x, 10))
	case uint64:
		sb.WriteString(formatUint64(x, cfg))
	case float64:
		// Match the utility toolkit by using the shortest form where possible.
		s := cfg.formatFloat(x, 'f', -1, 64)
		sb.WriteString(s)
	default:
		// Generic fallback: wrap into standard types before writing.
		w := wrap(v, cfg)
		if _, ok := w.(string); ok {
			writeQuoted(sb, w.(string))
			return nil
		}
		// Prevent recursion.
		switch w.(type) {
		case *JSONObject, *JSONArray, bool, int64, uint64, float64, string, jsonNull:
			return writeAny(sb, w, indent, depth, cfg)
		}
		return NewJSONError("unsupported JSON value type %T", v)
	}
	return nil
}

func writeObject(sb *strings.Builder, o *JSONObject, indent, depth int) error {
	if o == nil || len(o.keys) == 0 {
		sb.WriteString("{}")
		return nil
	}
	sb.WriteByte('{')
	first := true
	for _, k := range o.keys {
		v := o.values[k]
		if !first {
			sb.WriteByte(',')
		}
		first = false
		if indent > 0 {
			sb.WriteByte('\n')
			writeIndent(sb, indent, depth+1)
		}
		writeQuoted(sb, k)
		sb.WriteByte(':')
		if indent > 0 {
			sb.WriteByte(' ')
		}
		if err := writeAny(sb, v, indent, depth+1, o.cfg); err != nil {
			return err
		}
	}
	if indent > 0 {
		sb.WriteByte('\n')
		writeIndent(sb, indent, depth)
	}
	sb.WriteByte('}')
	return nil
}

func writeArray(sb *strings.Builder, a *JSONArray, indent, depth int) error {
	if a == nil || len(a.values) == 0 {
		sb.WriteString("[]")
		return nil
	}
	sb.WriteByte('[')
	for i, v := range a.values {
		if i > 0 {
			sb.WriteByte(',')
		}
		if indent > 0 {
			sb.WriteByte('\n')
			writeIndent(sb, indent, depth+1)
		}
		if err := writeAny(sb, v, indent, depth+1, a.cfg); err != nil {
			return err
		}
	}
	if indent > 0 {
		sb.WriteByte('\n')
		writeIndent(sb, indent, depth)
	}
	sb.WriteByte(']')
	return nil
}

func writeIndent(sb *strings.Builder, indent, depth int) {
	for i := 0; i < indent*depth; i++ {
		sb.WriteByte(' ')
	}
}

// writeQuoted writes a quoted and escaped string.
func writeQuoted(sb *strings.Builder, s string) {
	sb.WriteByte('"')
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c == '"':
			sb.WriteString("\\\"")
			i++
		case c == '\\':
			sb.WriteString("\\\\")
			i++
		case c == '\n':
			sb.WriteString("\\n")
			i++
		case c == '\r':
			sb.WriteString("\\r")
			i++
		case c == '\t':
			sb.WriteString("\\t")
			i++
		case c == '\b':
			sb.WriteString("\\b")
			i++
		case c == '\f':
			sb.WriteString("\\f")
			i++
		case c < 0x20:
			sb.WriteString("\\u")
			hex := "0123456789abcdef"
			sb.WriteByte('0')
			sb.WriteByte('0')
			sb.WriteByte(hex[c>>4])
			sb.WriteByte(hex[c&0xF])
			i++
		default:
			r, size := utf8.DecodeRuneInString(s[i:])
			sb.WriteRune(r)
			i += size
		}
	}
	sb.WriteByte('"')
}
