package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
)

// MD5 returns the MD5 digest bytes of data.
func MD5(data []byte) []byte {
	sum := md5.Sum(data)
	return sum[:]
}

// MD5Hex returns the MD5 digest of data in lower-case hex form.
func MD5Hex(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

// MD5Hex16 returns the middle 16 characters of the MD5 hex digest.
func MD5Hex16(data []byte) string { return MD5HexTo16(MD5Hex(data)) }

// MD5HexTo16 returns the middle 16 characters of a 32-character MD5 hex digest.
func MD5HexTo16(md5Hex string) string {
	if len(md5Hex) < 24 {
		return ""
	}
	return md5Hex[8:24]
}

// SHA1 returns the SHA1 digest bytes of data.
func SHA1(data []byte) []byte {
	sum := sha1.Sum(data)
	return sum[:]
}

// SHA1Hex returns the SHA1 digest of data in lower-case hex form.
func SHA1Hex(data []byte) string {
	sum := sha1.Sum(data)
	return hex.EncodeToString(sum[:])
}

// SHA224 returns the SHA224 digest bytes of data.
func SHA224(data []byte) []byte {
	sum := sha256.Sum224(data)
	return sum[:]
}

// SHA224Hex returns the SHA224 digest of data in lower-case hex form.
func SHA224Hex(data []byte) string { return hex.EncodeToString(SHA224(data)) }

// SHA256 returns the SHA256 digest bytes of data.
func SHA256(data []byte) []byte {
	sum := sha256.Sum256(data)
	return sum[:]
}

// SHA256Hex returns the SHA256 digest of data in lower-case hex form.
func SHA256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// SHA384 returns the SHA384 digest bytes of data.
func SHA384(data []byte) []byte {
	sum := sha512.Sum384(data)
	return sum[:]
}

// SHA384Hex returns the SHA384 digest of data in lower-case hex form.
func SHA384Hex(data []byte) string { return hex.EncodeToString(SHA384(data)) }

// SHA512 returns the SHA512 digest bytes of data.
func SHA512(data []byte) []byte {
	sum := sha512.Sum512(data)
	return sum[:]
}

// SHA512Hex returns the SHA512 digest of data in lower-case hex form.
func SHA512Hex(data []byte) string {
	sum := sha512.Sum512(data)
	return hex.EncodeToString(sum[:])
}
