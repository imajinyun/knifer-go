package json

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
)

// parseBytes parses JSON bytes into *JSONObject, *JSONArray, or a primitive value.
func parseBytes(b []byte) (any, error) {
	return parseBytesWithConfig(b, nil)
}

// parseBytesWithConfig parses JSON with config.
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
	dec := newDecoderWithConfig(bytes.NewReader(b), cfg)
	if dec == nil {
		return nil, NewJSONError("json: decoder factory returned nil")
	}
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

func newDecoderWithConfig(r io.Reader, cfg *Config) *json.Decoder {
	if cfg != nil && cfg.DecoderFactory != nil {
		return cfg.DecoderFactory(r)
	}
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return dec
}

// parseValue parses recursively from the current token.
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
			// Consume '}'.
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
		// Prefer exact integer representations before falling back to float64.
		if i, err := t.Int64(); err == nil {
			return i, nil
		}
		if u, err := strconv.ParseUint(t.String(), 10, 64); err == nil {
			return u, nil
		}
		f, err := t.Float64()
		if err != nil {
			return nil, WrapJSONError(err, "json: invalid number %q", t.String())
		}
		return f, nil
	}
	return nil, NewJSONError("json: unexpected token %v", tok)
}
