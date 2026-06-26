package vcrypto

import (
	"io"

	cryptoimpl "github.com/imajinyun/knifer-go/internal/crypto"
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

// GenAESKey returns a random AES key. Valid sizes are 16, 24, or 32 bytes.
func GenAESKey(size int) ([]byte, error) { return cryptoimpl.GenAESKey(size) }

// GenAESKeyWithOptions returns a random AES key using custom random options.
func GenAESKeyWithOptions(size int, opts ...RandomOption) ([]byte, error) {
	return cryptoimpl.GenAESKeyWithOptions(size, opts...)
}
