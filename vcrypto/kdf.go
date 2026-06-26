package vcrypto

import (
	"hash"

	cryptoimpl "github.com/imajinyun/knifer-go/internal/crypto"
)

// PBKDF2 derives a key from password and salt using PBKDF2.
func PBKDF2(password, salt []byte, iterations, keyLen int, fn func() hash.Hash) ([]byte, error) {
	return cryptoimpl.PBKDF2(password, salt, iterations, keyLen, fn)
}

// PBKDF2SHA256 derives a key using PBKDF2-HMAC-SHA256.
func PBKDF2SHA256(password, salt []byte, iterations, keyLen int) ([]byte, error) {
	return cryptoimpl.PBKDF2SHA256(password, salt, iterations, keyLen)
}
