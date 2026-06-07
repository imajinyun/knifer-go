package json

import (
	"bytes"
)

// ToBean 将 JSON 值转换到给定 dst（应为指针）。
// 当传入 *JSONObject/*JSONArray/string/[]byte/map/slice 等时，会先序列化为
// JSON 字节，再交给 encoding/json 反序列化到 dst。
func ToBean(src any, dst any) error {
	return ToBeanWithOptions(src, dst)
}

// ToBeanWithOptions converts a JSON value to dst using per-call options.
func ToBeanWithOptions(src any, dst any, opts ...BeanOption) error {
	if dst == nil {
		return NewJSONError("dst is nil")
	}
	cfg := applyBeanOptions(opts)
	var data []byte
	switch x := src.(type) {
	case []byte:
		data = x
	case string:
		data = []byte(x)
	default:
		w := wrap(src, cfg.cfg)
		s, err := writeValueWithConfig(w, 0, cfg.cfg)
		if err != nil {
			return err
		}
		data = []byte(s)
	}
	if cfg.unmarshalFunc != nil {
		if err := cfg.unmarshalFunc(data, dst); err != nil {
			return WrapJSONError(err, "to bean failed")
		}
		return nil
	}
	if cfg.cfg != nil && cfg.cfg.UnmarshalFunc != nil {
		if err := cfg.cfg.UnmarshalFunc(data, dst); err != nil {
			return WrapJSONError(err, "to bean failed")
		}
		return nil
	}
	dec := newDecoderWithConfig(bytes.NewReader(data), cfg.cfg)
	if dec == nil {
		return NewJSONError("to bean failed: decoder factory returned nil")
	}
	if err := dec.Decode(dst); err != nil {
		return WrapJSONError(err, "to bean failed")
	}
	return nil
}

// ToList 将 JSON 数组转换到 dst（必须是指向 slice 的指针）。
func ToList(src any, dst any) error { return ToListWithOptions(src, dst) }

// ToListWithOptions converts a JSON array to dst using per-call options.
func ToListWithOptions(src any, dst any, opts ...BeanOption) error {
	return ToBeanWithOptions(src, dst, opts...)
}
