package crypto

import knifer "github.com/imajinyun/knifer-go"

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
	// ErrInvalidSM2Signature indicates an invalid SM2 signature.
	ErrInvalidSM2Signature error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "invalid sm2 signature"}
	// ErrInvalidOTP indicates an invalid HOTP/TOTP input.
	ErrInvalidOTP error = &sentinel{code: knifer.ErrCodeInvalidInput, msg: "invalid otp"}
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

// ValidateSM4Key reports whether key is a valid SM4 key length.
func ValidateSM4Key(key []byte) error {
	if len(key) == 16 {
		return nil
	}
	return knifer.WrapError(knifer.ErrCodeInvalidInput, "sm4 key must be 16 bytes", ErrInvalidKey)
}

// ValidateSM4IV reports whether iv has the required block size for SM4 CBC helpers.
func ValidateSM4IV(iv []byte) error {
	if len(iv) == 16 {
		return nil
	}
	return knifer.WrapError(knifer.ErrCodeInvalidInput, "sm4 iv must be 16 bytes", ErrInvalidIV)
}
