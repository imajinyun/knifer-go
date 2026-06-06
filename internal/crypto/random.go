package crypto

import (
	"crypto/rand"
	"io"
)

type randomConfig struct {
	reader io.Reader
}

// RandomOption customizes random byte generation helpers.
type RandomOption func(*randomConfig)

// WithRandomReader sets the entropy source used by random byte helpers.
func WithRandomReader(reader io.Reader) RandomOption {
	return func(c *randomConfig) {
		if reader != nil {
			c.reader = reader
		}
	}
}

func applyRandomOptions(opts []RandomOption) randomConfig {
	cfg := randomConfig{reader: rand.Reader}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.reader == nil {
		cfg.reader = rand.Reader
	}
	return cfg
}

// RandomBytes returns n cryptographically secure random bytes.
func RandomBytes(n int) ([]byte, error) {
	return RandomBytesWithOptions(n)
}

// RandomBytesWithOptions returns n random bytes using custom random options.
func RandomBytesWithOptions(n int, opts ...RandomOption) ([]byte, error) {
	if n < 0 {
		return nil, ErrInvalidKey
	}
	cfg := applyRandomOptions(opts)
	b := make([]byte, n)
	_, err := io.ReadFull(cfg.reader, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateAESKey returns a random AES key. Valid sizes are 16, 24, or 32 bytes.
func GenerateAESKey(size int) ([]byte, error) {
	return GenerateAESKeyWithOptions(size)
}

// GenerateAESKeyWithOptions returns a random AES key using custom random options.
func GenerateAESKeyWithOptions(size int, opts ...RandomOption) ([]byte, error) {
	if size != 16 && size != 24 && size != 32 {
		return nil, ErrInvalidKey
	}
	return RandomBytesWithOptions(size, opts...)
}
