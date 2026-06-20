package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"hash"
)

// HMACBytes returns HMAC digest bytes using the given hash function.
// When fn is nil, HMACBytes uses SHA-256 instead of panicking.
func HMACBytes(fn func() hash.Hash, key, data []byte) []byte {
	if fn == nil {
		fn = sha256.New
	}
	h := hmac.New(fn, key)
	_, _ = h.Write(data)
	return h.Sum(nil)
}

// HMACHex returns HMAC digest in lower-case hex form using the given hash function.
// When fn is nil, HMACHex uses SHA-256 instead of panicking.
func HMACHex(fn func() hash.Hash, key, data []byte) string {
	if fn == nil {
		fn = sha256.New
	}
	h := hmac.New(fn, key)
	_, _ = h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// HMACSHA256Hex returns HMAC-SHA256 in lower-case hex form.
func HMACSHA256Hex(key, data []byte) string { return HMACHex(sha256.New, key, data) }

// HMACSHA384Hex returns HMAC-SHA384 in lower-case hex form.
func HMACSHA384Hex(key, data []byte) string { return HMACHex(sha512.New384, key, data) }

// HMACSHA512Hex returns HMAC-SHA512 in lower-case hex form.
func HMACSHA512Hex(key, data []byte) string { return HMACHex(sha512.New, key, data) }

// HMACEqual compares two MAC values in constant time.
func HMACEqual(a, b []byte) bool { return hmac.Equal(a, b) }

// ConstantTimeEqual compares two byte slices in constant time when lengths match.
func ConstantTimeEqual(a, b []byte) bool {
	return len(a) == len(b) && subtle.ConstantTimeCompare(a, b) == 1
}
