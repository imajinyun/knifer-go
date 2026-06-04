package vcrypto

import (
	knifer "github.com/imajinyun/go-knifer"
	cryptoimpl "github.com/imajinyun/go-knifer/internal/crypto"
)

// ErrInvalidKey indicates an invalid cryptographic key.
var ErrInvalidKey = cryptoimpl.ErrInvalidKey

// ErrInvalidIV indicates an invalid initialization vector.
var ErrInvalidIV = cryptoimpl.ErrInvalidIV

// ErrInvalidCipherText indicates invalid encrypted data.
var ErrInvalidCipherText = cryptoimpl.ErrInvalidCipherText

// ValidateAESKey reports whether key is a valid AES key length (16, 24, or 32
// bytes). On failure it returns a *knifer.Error classified as
// knifer.ErrCodeInvalidInput while preserving ErrInvalidKey on the chain, so
// callers may match either errors.Is(err, knifer.ErrCodeInvalidInput) or
// errors.Is(err, vcrypto.ErrInvalidKey).
func ValidateAESKey(key []byte) error {
	switch len(key) {
	case 16, 24, 32:
		return nil
	default:
		return knifer.WrapError(knifer.ErrCodeInvalidInput, "aes key must be 16, 24, or 32 bytes", ErrInvalidKey)
	}
}

// ValidateAESIV reports whether iv has the required block size for AES CBC/CFB/OFB/CTR helpers.
func ValidateAESIV(iv []byte) error {
	if len(iv) == 16 {
		return nil
	}
	return knifer.WrapError(knifer.ErrCodeInvalidInput, "aes iv must be 16 bytes", ErrInvalidIV)
}

// ValidateAESGCMNonce reports whether nonce has the default nonce size used by AES-GCM helpers.
func ValidateAESGCMNonce(nonce []byte) error {
	if len(nonce) == 12 {
		return nil
	}
	return knifer.WrapError(knifer.ErrCodeInvalidInput, "aes-gcm nonce must be 12 bytes", ErrInvalidIV)
}
