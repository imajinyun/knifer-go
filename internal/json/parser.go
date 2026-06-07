package json

import (
	"bytes"
	"encoding/json"
	"io"
)

// parseBytes 把 JSON 字节解析为 *JSONObject 或 *JSONArray 或基础值。
func parseBytes(b []byte) (any, error) {
	return parseBytesWithConfig(b, nil)
}

// parseBytesWithConfig 使用配置解析 JSON。
func parseBytesWithConfig(b []byte, cfg *Config) (any, error) {
	if cfg == nil {
		cfg = NewConfig()
	}
	if cfg.UnmarshalFunc != nil {
		var raw any
		if err := cfg.UnmarshalFunc(b, &raw); err != nil {
			return nil, WrapJSONError(err, "json: parse failed")
		}
		return wrap(raw, cfg), nil
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return nil, WrapJSONError(err, "json: parse failed")
	}
	v, err := parseValue(dec, tok, cfg)
	if err != nil {
		return nil, err
	}
	// Check whether there is any non-whitespace token after the root value.
	if _, err := dec.Token(); err == nil {
		return nil, NewJSONError("json: unexpected trailing content")
	} else if err != io.EOF {
		return nil, WrapJSONError(err, "json: read trailing content failed")
	}
	return v, nil
}

// parseValue 根据当前 token 递归解析。
func parseValue(dec *json.Decoder, tok json.Token, cfg *Config) (any, error) {
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			obj := NewJSONObjectWithConfig(cfg)
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return nil, WrapJSONError(err, "json: read key failed")
				}
				key, ok := kt.(string)
				if !ok {
					return nil, NewJSONError("json: expect string key, got %T", kt)
				}
				vt, err := dec.Token()
				if err != nil {
					return nil, WrapJSONError(err, "json: read value failed")
				}
				v, err := parseValue(dec, vt, cfg)
				if err != nil {
					return nil, err
				}
				obj.Set(key, v)
			}
			// 消费 '}'
			if _, err := dec.Token(); err != nil {
				return nil, WrapJSONError(err, "json: missing '}'")
			}
			return obj, nil
		case '[':
			arr := NewJSONArrayWithConfig(cfg)
			for dec.More() {
				vt, err := dec.Token()
				if err != nil {
					return nil, WrapJSONError(err, "json: read element failed")
				}
				v, err := parseValue(dec, vt, cfg)
				if err != nil {
					return nil, err
				}
				arr.Add(v)
			}
			if _, err := dec.Token(); err != nil {
				return nil, WrapJSONError(err, "json: missing ']'")
			}
			return arr, nil
		}
	case nil:
		return Null, nil
	case bool:
		return t, nil
	case string:
		return t, nil
	case json.Number:
		// 优先 int64，失败回退 float64
		if i, err := t.Int64(); err == nil {
			return i, nil
		}
		f, err := t.Float64()
		if err != nil {
			return nil, WrapJSONError(err, "json: invalid number %q", t.String())
		}
		return f, nil
	}
	return nil, NewJSONError("json: unexpected token %v", tok)
}
