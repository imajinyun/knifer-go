package crypto

import knifer "github.com/imajinyun/go-knifer"

type sentinel struct {
	code knifer.ErrCode
	msg  string
}

func (e *sentinel) Error() string { return e.msg }

func (e *sentinel) ErrorCode() knifer.ErrCode { return e.code }

func (e *sentinel) Is(target error) bool {
	if e == target {
		return true
	}
	code, ok := target.(knifer.ErrCode)
	return ok && e.code == code
}

var (
	// ErrInvalidKey indicates an invalid cryptographic key.
	ErrInvalidKey error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "invalid key"}
	// ErrInvalidIV indicates an invalid initialization vector.
	ErrInvalidIV error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "invalid iv"}
	// ErrInvalidCipherText indicates invalid encrypted data.
	ErrInvalidCipherText error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "invalid cipher text"}
)

// ValidateAESKey reports whether key is a valid AES key length.
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
