package vcrypto

import cryptoimpl "github.com/imajinyun/go-knifer/internal/crypto"

// RandomBytes returns n cryptographically secure random bytes.
func RandomBytes(n int) ([]byte, error) { return cryptoimpl.RandomBytes(n) }

// GenerateAESKey returns a random AES key. Valid sizes are 16, 24, or 32 bytes.
func GenerateAESKey(size int) ([]byte, error) { return cryptoimpl.GenerateAESKey(size) }
