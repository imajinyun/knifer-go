package json

import (
	"bytes"
	"encoding/json"
	"reflect"
	"time"
)

// wrap 将任意 Go 值转为 JSON 兼容值（基础类型 / *JSONObject / *JSONArray / Null）。
func wrap(v any, cfg *Config) any {
	if cfg == nil {
		cfg = NewConfig()
	}
	if v == nil {
		return Null
	}
	switch x := v.(type) {
	case jsonNull, *jsonNull:
		return Null
	case *JSONObject, *JSONArray:
		return x
	case string, bool:
		return x
	case int:
		return int64(x)
	case int8:
		return int64(x)
	case int16:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	case uint:
		return int64(x)
	case uint8:
		return int64(x)
	case uint16:
		return int64(x)
	case uint32:
		return int64(x)
	case uint64:
		return int64(x)
	case float32:
		return float64(x)
	case float64:
		return x
	case json.Number:
		if i, err := x.Int64(); err == nil {
			return i
		}
		if f, err := x.Float64(); err == nil {
			return f
		}
		return x.String()
	case time.Time:
		if cfg.DateFormat == "" {
			return x.UnixMilli()
		}
		return x.Format(cfg.DateFormat)
	case []byte:
		// []byte 视作字符串。
		return string(x)
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Interface:
		if rv.IsNil() {
			return Null
		}
		return wrap(rv.Elem().Interface(), cfg)
	case reflect.Map:
		obj := NewJSONObjectWithConfig(cfg)
		// 仅支持字符串 key
		iter := rv.MapRange()
		for iter.Next() {
			k := cfg.sprint(iter.Key().Interface())
			obj.Set(k, wrap(iter.Value().Interface(), cfg))
		}
		return obj
	case reflect.Slice, reflect.Array:
		arr := NewJSONArrayWithConfig(cfg)
		for i := 0; i < rv.Len(); i++ {
			arr.Add(wrap(rv.Index(i).Interface(), cfg))
		}
		return arr
	case reflect.Struct:
		// 通过 encoding/json 反序列化为通用结构后再 wrap，确保 tag 生效。
		marshal := json.Marshal
		if cfg != nil && cfg.MarshalFunc != nil {
			marshal = cfg.MarshalFunc
		}
		b, err := marshal(v)
		if err != nil {
			return cfg.sprint(v)
		}
		if cfg != nil && cfg.UnmarshalFunc != nil {
			var raw any
			if err := cfg.UnmarshalFunc(b, &raw); err != nil {
				return cfg.sprint(v)
			}
			return wrap(raw, cfg)
		}
		dec := newDecoderWithConfig(bytes.NewReader(b), cfg)
		if dec == nil {
			return cfg.sprint(v)
		}
		var raw any
		if err := dec.Decode(&raw); err != nil {
			return cfg.sprint(v)
		}
		return wrap(raw, cfg)
	}
	return cfg.sprint(v)
}

// toString 把任意 JSON 值转换为字符串。
func toString(v any, def string, cfg *Config) string {
	cfg = configOrDefault(cfg)
	if IsNull(v) {
		return def
	}
	switch x := v.(type) {
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case int64:
		return cfg.formatInt(x, 10)
	case float64:
		return cfg.formatFloat(x, 'f', -1, 64)
	case *JSONObject:
		return x.String()
	case *JSONArray:
		return x.String()
	}
	return cfg.sprint(v)
}

// toInt64 转 int64，失败时返回 def。
func toInt64(v any, def int64, cfg *Config) int64 {
	cfg = configOrDefault(cfg)
	if IsNull(v) {
		return def
	}
	switch x := v.(type) {
	case int64:
		return x
	case float64:
		return int64(x)
	case bool:
		if x {
			return 1
		}
		return 0
	case string:
		n, err := cfg.parseInt(x, 10, 64)
		if err == nil {
			return n
		}
		f, err := cfg.parseFloat(x, 64)
		if err == nil {
			return int64(f)
		}
	}
	return def
}

// toFloat64 转 float64，失败时返回 def。
func toFloat64(v any, def float64, cfg *Config) float64 {
	cfg = configOrDefault(cfg)
	if IsNull(v) {
		return def
	}
	switch x := v.(type) {
	case float64:
		return x
	case int64:
		return float64(x)
	case bool:
		if x {
			return 1
		}
		return 0
	case string:
		f, err := cfg.parseFloat(x, 64)
		if err == nil {
			return f
		}
	}
	return def
}

// toBool 转 bool。
func toBool(v any, def bool, cfg *Config) bool {
	cfg = configOrDefault(cfg)
	if IsNull(v) {
		return def
	}
	switch x := v.(type) {
	case bool:
		return x
	case int64:
		return x != 0
	case float64:
		return x != 0
	case string:
		if b, err := cfg.parseBool(x); err == nil {
			return b
		}
	}
	return def
}
