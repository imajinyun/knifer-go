package vcrypto

import cryptoimpl "github.com/imajinyun/knifer-go/internal/crypto"

// ErrInvalidKey indicates an invalid cryptographic key.
var ErrInvalidKey = cryptoimpl.ErrInvalidKey

// ErrInvalidIV indicates an invalid initialization vector.
var ErrInvalidIV = cryptoimpl.ErrInvalidIV

// ErrInvalidCipherText indicates invalid encrypted data.
var ErrInvalidCipherText = cryptoimpl.ErrInvalidCipherText

// ErrInvalidSM2Signature indicates an invalid SM2 signature.
var ErrInvalidSM2Signature = cryptoimpl.ErrInvalidSM2Signature

// ErrInvalidOTP indicates an invalid HOTP/TOTP input.
var ErrInvalidOTP = cryptoimpl.ErrInvalidOTP

// ValidateAESKey reports whether key is a valid AES key length (16, 24, or 32
// bytes). On failure it returns a *knifer.Error classified as
// knifer.ErrCodeInvalidInput while preserving ErrInvalidKey on the chain, so
// callers may match either errors.Is(err, knifer.ErrCodeInvalidInput) or
// errors.Is(err, vcrypto.ErrInvalidKey).
func ValidateAESKey(key []byte) error {
	return cryptoimpl.ValidateAESKey(key)
}

// ValidateAESIV reports whether iv has the required block size for AES CBC/CFB/OFB/CTR helpers.
func ValidateAESIV(iv []byte) error {
	return cryptoimpl.ValidateAESIV(iv)
}

// ValidateAESGCMNonce reports whether nonce has the default nonce size used by AES-GCM helpers.
func ValidateAESGCMNonce(nonce []byte) error {
	return cryptoimpl.ValidateAESGCMNonce(nonce)
}

// ValidateSM4Key reports whether key is a valid SM4 key length.
func ValidateSM4Key(key []byte) error {
	return cryptoimpl.ValidateSM4Key(key)
}

// ValidateSM4IV reports whether iv has the required block size for SM4 CBC helpers.
func ValidateSM4IV(iv []byte) error {
	return cryptoimpl.ValidateSM4IV(iv)
}
