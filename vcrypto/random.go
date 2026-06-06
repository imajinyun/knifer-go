package vcrypto

import (
	"io"

	cryptoimpl "github.com/imajinyun/go-knifer/internal/crypto"
)

// RandomOption customizes random byte generation helpers.
type RandomOption = cryptoimpl.RandomOption

// WithRandomReader sets the entropy source used by random byte helpers.
func WithRandomReader(reader io.Reader) RandomOption { return cryptoimpl.WithRandomReader(reader) }

// RandomBytes returns n cryptographically secure random bytes.
func RandomBytes(n int) ([]byte, error) { return cryptoimpl.RandomBytes(n) }

// RandomBytesWithOptions returns n random bytes using custom random options.
func RandomBytesWithOptions(n int, opts ...RandomOption) ([]byte, error) {
	return cryptoimpl.RandomBytesWithOptions(n, opts...)
}

// GenerateAESKey returns a random AES key. Valid sizes are 16, 24, or 32 bytes.
func GenerateAESKey(size int) ([]byte, error) { return cryptoimpl.GenerateAESKey(size) }

// GenerateAESKeyWithOptions returns a random AES key using custom random options.
func GenerateAESKeyWithOptions(size int, opts ...RandomOption) ([]byte, error) {
	return cryptoimpl.GenerateAESKeyWithOptions(size, opts...)
}
