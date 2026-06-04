package vcrypto

import cryptoimpl "github.com/imajinyun/go-knifer/internal/crypto"

// SignParams joins params by sorted key and returns the digest hex using digestHex.
func SignParams(params map[string]any, digestHex func([]byte) string, separator, keyValueSeparator string, ignoreNil bool, otherParams ...string) string {
	return cryptoimpl.SignParams(params, digestHex, separator, keyValueSeparator, ignoreNil, otherParams...)
}

// SignParamsMD5 signs sorted params with MD5.
func SignParamsMD5(params map[string]any, otherParams ...string) string {
	return cryptoimpl.SignParamsMD5(params, otherParams...)
}

// SignParamsSHA1 signs sorted params with SHA1.
func SignParamsSHA1(params map[string]any, otherParams ...string) string {
	return cryptoimpl.SignParamsSHA1(params, otherParams...)
}

// SignParamsSHA256 signs sorted params with SHA256.
func SignParamsSHA256(params map[string]any, otherParams ...string) string {
	return cryptoimpl.SignParamsSHA256(params, otherParams...)
}
