package vcrypto

import cryptoimpl "github.com/imajinyun/knifer-go/internal/crypto"

// SignParams joins params by sorted key and returns the digest hex using digestHex.
func SignParams(params map[string]any, digestHex func([]byte) string, separator, keyValueSeparator string, ignoreNil bool, otherParams ...string) string {
	return cryptoimpl.SignParams(params, digestHex, separator, keyValueSeparator, ignoreNil, otherParams...)
}

// SignParamsSHA256 signs sorted params with SHA256.
func SignParamsSHA256(params map[string]any, otherParams ...string) string {
	return cryptoimpl.SignParamsSHA256(params, otherParams...)
}
