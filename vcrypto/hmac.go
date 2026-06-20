package vcrypto

import (
	"hash"

	cryptoimpl "github.com/imajinyun/go-knifer/internal/crypto"
)

// HMACHex returns HMAC digest in lower-case hex form using the given hash function.
// When fn is nil, HMACHex uses SHA-256 instead of panicking.
func HMACHex(fn func() hash.Hash, key, data []byte) string { return cryptoimpl.HMACHex(fn, key, data) }

// HMACBytes returns HMAC digest bytes using the given hash function.
// When fn is nil, HMACBytes uses SHA-256 instead of panicking.
func HMACBytes(fn func() hash.Hash, key, data []byte) []byte {
	return cryptoimpl.HMACBytes(fn, key, data)
}

// HMACSHA256Hex returns HMAC-SHA256 in lower-case hex form.
func HMACSHA256Hex(key, data []byte) string { return cryptoimpl.HMACSHA256Hex(key, data) }

// HMACSHA512Hex returns HMAC-SHA512 in lower-case hex form.
func HMACSHA512Hex(key, data []byte) string { return cryptoimpl.HMACSHA512Hex(key, data) }

// HMACSHA384Hex returns HMAC-SHA384 in lower-case hex form.
func HMACSHA384Hex(key, data []byte) string { return cryptoimpl.HMACSHA384Hex(key, data) }

// HMACEqual compares two MAC values in constant time.
func HMACEqual(a, b []byte) bool { return cryptoimpl.HMACEqual(a, b) }

// ConstantTimeEqual compares two byte slices in constant time when lengths match.
func ConstantTimeEqual(a, b []byte) bool { return cryptoimpl.ConstantTimeEqual(a, b) }
