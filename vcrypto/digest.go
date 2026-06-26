package vcrypto

import (
	"hash"

	cryptoimpl "github.com/imajinyun/knifer-go/internal/crypto"
)

// Digest returns digest bytes computed by newHash.
func Digest(data []byte, newHash func() hash.Hash) []byte { return cryptoimpl.Digest(data, newHash) }

// DigestHex returns the digest computed by newHash in lower-case hex form.
func DigestHex(data []byte, newHash func() hash.Hash) string {
	return cryptoimpl.DigestHex(data, newHash)
}

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
