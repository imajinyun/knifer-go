package crypto

import (
	"bytes"
	"fmt"
	"sort"
)

// SignParams joins params by sorted key and returns the digest hex using digestHex.
func SignParams(params map[string]any, digestHex func([]byte) string, separator, keyValueSeparator string, ignoreNil bool, otherParams ...string) string {
	keys := make([]string, 0, len(params))
	for key, value := range params {
		if ignoreNil && value == nil {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys)+len(otherParams))
	for _, key := range keys {
		value := params[key]
		parts = append(parts, key+keyValueSeparator+fmt.Sprint(value))
	}
	parts = append(parts, otherParams...)
	return digestHex([]byte(stringsJoin(parts, separator)))
}

// SignParamsMD5 signs sorted params with MD5.
func SignParamsMD5(params map[string]any, otherParams ...string) string {
	return SignParams(params, MD5Hex, "", "", true, otherParams...)
}

// SignParamsSHA1 signs sorted params with SHA1.
func SignParamsSHA1(params map[string]any, otherParams ...string) string {
	return SignParams(params, SHA1Hex, "", "", true, otherParams...)
}

// SignParamsSHA256 signs sorted params with SHA256.
func SignParamsSHA256(params map[string]any, otherParams ...string) string {
	return SignParams(params, SHA256Hex, "", "", true, otherParams...)
}

func stringsJoin(parts []string, separator string) string {
	if len(parts) == 0 {
		return ""
	}
	if separator == "" {
		var b bytes.Buffer
		for _, part := range parts {
			b.WriteString(part)
		}
		return b.String()
	}
	var b bytes.Buffer
	for i, part := range parts {
		if i > 0 {
			b.WriteString(separator)
		}
		b.WriteString(part)
	}
	return b.String()
}
