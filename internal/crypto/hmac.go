package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"hash"
)

// HMACBytes returns HMAC digest bytes using the given hash function.
func HMACBytes(fn func() hash.Hash, key, data []byte) []byte {
	h := hmac.New(fn, key)
	_, _ = h.Write(data)
	return h.Sum(nil)
}

// HMACHex returns HMAC digest in lower-case hex form using the given hash function.
func HMACHex(fn func() hash.Hash, key, data []byte) string {
	h := hmac.New(fn, key)
	_, _ = h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// HMACMD5Hex returns HMAC-MD5 in lower-case hex form.
func HMACMD5Hex(key, data []byte) string { return HMACHex(md5.New, key, data) }

// HMACSHA1Hex returns HMAC-SHA1 in lower-case hex form.
func HMACSHA1Hex(key, data []byte) string { return HMACHex(sha1.New, key, data) }

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
