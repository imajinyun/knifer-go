package vcrypto

import cryptoimpl "github.com/imajinyun/go-knifer/internal/crypto"

// MD5Hex returns the MD5 digest of s in lower-case hex form.
func MD5Hex(s string) string { return cryptoimpl.MD5Hex([]byte(s)) }

// MD5HexBytes returns the MD5 digest of data in lower-case hex form.
func MD5HexBytes(data []byte) string { return cryptoimpl.MD5Hex(data) }

// MD5 returns the MD5 digest bytes of data.
func MD5(data []byte) []byte { return cryptoimpl.MD5(data) }

// MD5Hex16 returns the middle 16 characters of the MD5 hex digest.
func MD5Hex16(data []byte) string { return cryptoimpl.MD5Hex16(data) }

// MD5HexTo16 returns the middle 16 characters of a 32-character MD5 hex digest.
func MD5HexTo16(md5Hex string) string { return cryptoimpl.MD5HexTo16(md5Hex) }

// SHA1Hex returns the SHA1 digest of s in lower-case hex form.
func SHA1Hex(s string) string { return cryptoimpl.SHA1Hex([]byte(s)) }

// SHA1 returns the SHA1 digest bytes of data.
func SHA1(data []byte) []byte { return cryptoimpl.SHA1(data) }

// SHA1HexBytes returns the SHA1 digest of data in lower-case hex form.
func SHA1HexBytes(data []byte) string { return cryptoimpl.SHA1Hex(data) }

// SHA224 returns the SHA224 digest bytes of data.
func SHA224(data []byte) []byte { return cryptoimpl.SHA224(data) }

// SHA224Hex returns the SHA224 digest of data in lower-case hex form.
func SHA224Hex(data []byte) string { return cryptoimpl.SHA224Hex(data) }

// SHA256Hex returns the SHA256 digest of s in lower-case hex form.
func SHA256Hex(s string) string { return cryptoimpl.SHA256Hex([]byte(s)) }

// SHA256 returns the SHA256 digest bytes of data.
func SHA256(data []byte) []byte { return cryptoimpl.SHA256(data) }

// SHA256HexBytes returns the SHA256 digest of data in lower-case hex form.
func SHA256HexBytes(data []byte) string { return cryptoimpl.SHA256Hex(data) }

// SHA384 returns the SHA384 digest bytes of data.
func SHA384(data []byte) []byte { return cryptoimpl.SHA384(data) }

// SHA384Hex returns the SHA384 digest of data in lower-case hex form.
func SHA384Hex(data []byte) string { return cryptoimpl.SHA384Hex(data) }

// SHA512Hex returns the SHA512 digest of s in lower-case hex form.
func SHA512Hex(s string) string { return cryptoimpl.SHA512Hex([]byte(s)) }

// SHA512 returns the SHA512 digest bytes of data.
func SHA512(data []byte) []byte { return cryptoimpl.SHA512(data) }

// SHA512HexBytes returns the SHA512 digest of data in lower-case hex form.
func SHA512HexBytes(data []byte) string { return cryptoimpl.SHA512Hex(data) }
